// NeuroScript Version: 0.3.0
// File version: 0.1.11
// Corrected logic in runTuiMode by removing reference to non-existent Config.RunScriptAndExit.
// Updated App struct and runTuiMode for tview integration.
// filename: pkg/neurogo/app.go
// nlines: 275 // Approximate
// risk_rating: MEDIUM
package neurogo

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// App orchestrates the main application logic.
type App struct {
	Config       *Config
	Log          logging.Logger
	interpreter  *core.Interpreter
	llmClient    core.LLMClient
	agentCtx     *AgentContext
	patchHandler *PatchHandler
	mu           sync.RWMutex
	appCtx       context.Context
	cancelFunc   context.CancelFunc

	// --- tview specific fields ---
	tui            *tviewAppPointers // Holds references to tview UI components
	originalStdout io.Writer         // To store interpreter's original stdout
	// --- End tview specific fields ---
}

// NewApp creates a new application instance.
func NewApp(cfg *Config, logger logging.Logger, llmClient core.LLMClient) (*App, error) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Warning: NewApp received a nil logger. Defaulting to NoOpLogger.\n")
		logger = adapters.NewNoOpLogger()
	}
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil for NewApp")
	}

	mainCtx, cancel := context.WithCancel(context.Background())

	app := &App{
		Config:     cfg,
		Log:        logger,
		llmClient:  llmClient,
		appCtx:     mainCtx,
		cancelFunc: cancel,
	}
	return app, nil
}

// SetInterpreter allows setting the interpreter after App creation.
func (a *App) SetInterpreter(interp *core.Interpreter) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.interpreter = interp
}

// SetAIWorkerManager sets the AI Worker Manager on the interpreter.
func (a *App) SetAIWorkerManager(wm *core.AIWorkerManager) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.interpreter != nil {
		a.interpreter.SetAIWorkerManager(wm)
	} else {
		a.Log.Warn("Attempted to set AIWorkerManager, but interpreter is nil.")
	}
}

// AIWorkerManager returns the application's AIWorkerManager instance.
// It provides read-only access from other parts of the neurogo package.
func (a *App) AIWorkerManager() *core.AIWorkerManager {
	return a.interpreter.AIWorkerManager()
}

// GetInterpreter safely retrieves the interpreter.
func (a *App) GetInterpreter() *core.Interpreter {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.interpreter
}

// GetAIWorkerManager safely retrieves the AIWorkerManager from the interpreter.
func (a *App) GetAIWorkerManager() *core.AIWorkerManager {
	a.mu.RLock()
	interpreter := a.interpreter
	a.mu.RUnlock()

	if interpreter == nil {
		a.Log.Warn("GetAIWorkerManager called when interpreter is nil.")
		return nil
	}
	return interpreter.AIWorkerManager()
}

// Run starts the application based on the configured mode.
func (a *App) Run() error {
	a.Log.Info("NeuroScript application starting...", "version", "0.4.0")

	if err := a.initializeCoreComponents(); err != nil {
		a.Log.Error("Failed to initialize core components", "error", err)
		return fmt.Errorf("core component initialization failed: %w", err)
	}

	if a.Config.SyncDir != "" {
		a.Log.Info("Sync mode selected.", "sync_dir", a.Config.SyncDir)
		return a.runSyncMode(a.appCtx)
	}
	if a.Config.StartupScript != "" {
		a.Log.Info("Script execution mode selected.", "script", a.Config.StartupScript)
		err := a.ExecuteScriptFile(a.appCtx, a.Config.StartupScript)
		if err != nil {
			a.Log.Error("Script execution failed", "script", a.Config.StartupScript, "error", err)
		}
		return err
	}

	a.Log.Info("No specific script or sync directory provided, defaulting to TUI mode.")
	return a.runTuiMode(a.appCtx)
}

// runTuiMode starts the application in TUI mode.
func (a *App) runTuiMode(ctx context.Context) error {
	a.Log.Info("Starting TUI mode (tview)...")

	// Based on App.Run() logic, if StartupScript was set, it would have run and exited.
	// So, for TUI mode reached this way, there's no initial script from a.Config.StartupScript.
	// If a TUI-specific startup script is desired, a new Config field would be needed.
	initialTuiScript := ""

	if a.interpreter != nil {
		a.originalStdout = a.interpreter.Stdout()
	}

	// Call the tview TUI starter function (defined in pkg/neurogo/tview_tui.go)
	err := StartTviewTUI(a, initialTuiScript)

	if a.interpreter != nil && a.originalStdout != nil {
		a.interpreter.SetStdout(a.originalStdout)
		a.Log.Info("Restored interpreter's original stdout after TUI exit.")
	}

	if err != nil {
		a.Log.Error("TUI mode exited with error", "error", err)
		return err
	}
	a.Log.Info("TUI mode exited gracefully.")
	return nil
}

func (a *App) GetLogger() logging.Logger {
	if a.Log == nil {
		return adapters.NewNoOpLogger()
	}
	return a.Log
}

func (a *App) GetLLMClient() core.LLMClient {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.llmClient
}

func (a *App) Context() context.Context {
	return a.appCtx
}

func (a *App) initializeCoreComponents() error {
	a.Log.Debug("Initializing core components...")

	if a.llmClient == nil {
		var errLLM error
		a.llmClient, errLLM = a.CreateLLMClient()
		if errLLM != nil {
			a.Log.Error("Could not initialize LLM Client during core component setup", "error", errLLM)
			a.Log.Warn("LLMClient creation failed. Using new NoOpLLMClient as fallback.")
			a.llmClient = adapters.NewNoOpLLMClient()
		}
	}

	sandboxDir := a.Config.SandboxDir
	if sandboxDir == "" {
		var err error
		homeDir, errHome := os.UserHomeDir()
		if errHome != nil {
			sandboxDir, err = os.MkdirTemp("", "neuroscript_sandbox_")
			if err != nil {
				return fmt.Errorf("failed to create temporary sandbox directory: %w", err)
			}
			a.Log.Info("Created temporary sandbox (home dir failed)", "path", sandboxDir)
		} else {
			sandboxDir = filepath.Join(homeDir, ".neuroscript", "sandbox")
			a.Log.Info("Using default sandbox location", "path", sandboxDir)
		}
		a.Config.SandboxDir = sandboxDir
	}

	if _, err := os.Stat(sandboxDir); os.IsNotExist(err) {
		if errMkdir := os.MkdirAll(sandboxDir, 0755); errMkdir != nil {
			return fmt.Errorf("failed to create specified sandbox directory %s: %w", sandboxDir, errMkdir)
		}
		a.Log.Info("Created sandbox directory", "path", sandboxDir)
	} else {
		a.Log.Info("Using existing sandbox directory", "path", sandboxDir)
	}

	if a.patchHandler == nil {
		a.Log.Warn("PatchHandler is nil after NewApp. Patch functionality may be disabled.")
	}

	interpLLMClient := a.llmClient
	if interpLLMClient == nil {
		a.Log.Warn("LLMClient is nil for interpreter, using NoOpLLMClient for interpreter.")
		interpLLMClient = adapters.NewNoOpLLMClient()
	}

	var errInterp error
	a.interpreter, errInterp = core.NewInterpreter(a.Log, interpLLMClient, sandboxDir, nil, nil)
	if errInterp != nil {
		return fmt.Errorf("failed to create interpreter: %w", errInterp)
	}
	if errRegister := core.RegisterCoreTools(a.interpreter); errRegister != nil {
		a.Log.Warn("Failed to register all core tools", "error", errRegister)
	}

	a.agentCtx = NewAgentContext(a.Log)
	if a.agentCtx != nil {
		a.agentCtx.SetSandboxDir(sandboxDir)
		a.agentCtx.SetAllowlistPath(a.Config.AllowlistFile)
		a.agentCtx.SetModelName(a.Config.ModelName)
	}
	if errSetSandbox := a.interpreter.SetSandboxDir(sandboxDir); errSetSandbox != nil {
		a.Log.Error("Failed to set sandbox dir on interpreter post init", "error", errSetSandbox)
	}

	aiWmLLMClient := a.llmClient
	if aiWmLLMClient == nil {
		aiWmLLMClient = adapters.NewNoOpLLMClient()
	}
	workerDefsPath := filepath.Join(sandboxDir, "ai_worker_definitions.json")
	perfDataPath := filepath.Join(sandboxDir, "ai_worker_performance.jsonl")

	aiWm, errManager := core.NewAIWorkerManager(a.Log, sandboxDir, aiWmLLMClient, workerDefsPath, perfDataPath)
	if errManager != nil {
		a.Log.Error("Failed to initialize AIWorkerManager", "error", errManager)
	} else {
		if loadErr := aiWm.LoadWorkerDefinitionsFromFile(); loadErr != nil {
			a.Log.Warn("Could not load AI worker definitions", "path", workerDefsPath, "error", loadErr)
		}
	}

	if a.interpreter != nil && aiWm != nil {
		a.interpreter.SetAIWorkerManager(aiWm)
	} else if a.interpreter == nil {
		a.Log.Error("Interpreter not initialized, cannot set AIWorkerManager.")
	}

	a.Log.Info("Core components initialization attempt finished.")
	return nil
}

var _ WMStatusViewDataProvider = (*App)(nil) // This interface might change/be removed with tview
// filename: pkg/neurogo/app.go
// Add these methods to your *App type

// HandleSystemCommand processes system-level commands (e.g., //chat, //run) from the TUI.
func (a *App) HandleSystemCommand(command string) {
	if a.Log == nil { // Safety check
		fmt.Printf("App.Log is nil, cannot log system command: %s\n", command)
		return
	}
	a.Log.Info("System command received by App (from TUI)", "command", command)

	// TODO: Implement actual parsing and handling of system commands from tui_design.md
	//       (e.g., //chat <num>, //run <script>, //sync)

	// For now, just EMIT it back to local output (Pane A) for visual feedback
	// if a.interpreter != nil && a.interpreter.Stdout() != nil {
	// 	// Ensure the write is thread-safe if called from a different goroutine
	// 	// The tviewWriter already handles QueueUpdateDraw.
	// 	fmt.Fprintf(a.interpreter.Stdout(), "[App echoed System Command]: %s\n", command)
	// } else if a.tui != nil && a.tui.localOutputView != nil && a.tui.tviewApp != nil {
	// 	// Fallback if interpreter or its stdout is not set up, write directly to TUI Pane A
	// 	a.tui.tviewApp.QueueUpdateDraw(func() {
	// 		fmt.Fprintf(a.tui.localOutputView, "[App echoed System Command]: %s\n", command)
	// 	})
	// }
}

// ExecuteScriptLine executes a single line of NeuroScript from the TUI.
// This is a placeholder; actual single-line execution might be complex or not directly supported
// by the current interpreter design without wrapping it as a temporary script.
func (a *App) ExecuteScriptLine(ctx context.Context, line string) {
	if a.Log == nil { // Safety check
		fmt.Printf("App.Log is nil, cannot log script line: %s\n", line)
		return
	}
	a.Log.Info("Script line received by App (from TUI)", "line", line)

	// TODO: Implement robust single-line script execution if this feature is desired.
	//       This might involve parsing the line, determining if it's a valid standalone statement
	//       or expression, and then using appropriate methods on a.interpreter.
	//       For example, core.Interpreter might need an ExecuteAdHoc(ctx, line) method.

	// For now, just EMIT it back to local output (Pane A) for visual feedback
	if a.interpreter != nil && a.interpreter.Stdout() != nil {
		fmt.Fprintf(a.interpreter.Stdout(), "[App echoed Script Line]: %s\n", line)
		// Example conceptual execution (your interpreter API may vary):
		// results, execErr := a.interpreter.ExecuteAdHocStatement(ctx, line)
		// if execErr != nil {
		// 	fmt.Fprintf(a.interpreter.Stdout(), "Error executing line: %v\n", execErr)
		// } else if results != nil {
		//	// Format and print results if any
		//  fmt.Fprintf(a.interpreter.Stdout(), "Result: %v\n", results)
		// }
	} else if a.tui != nil && a.tui.localOutputView != nil && a.tui.tviewApp != nil {
		// Fallback if interpreter or its stdout is not set up
		// 	a.tui.tviewApp.QueueUpdateDraw(func() {
		// 		fmt.Fprintf(a.tui.localOutputView, "[App echoed Script Line]: %s\n", line)
		// 	})
	}
}
