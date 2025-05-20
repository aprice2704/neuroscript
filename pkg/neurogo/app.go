// NeuroScript Version: 0.3.0
// File version: 0.1.13 // Modified App struct for multiple chat sessions
// Added chat session management methods and fields to App struct.
// filename: pkg/neurogo/app.go
// nlines: 350 // Approximate
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
	mu           sync.RWMutex // General mutex for app-level fields like interpreter, llmClient
	appCtx       context.Context
	cancelFunc   context.CancelFunc

	// --- tview specific fields ---
	tui            *tviewAppPointers // Holds references to tview UI components
	originalStdout io.Writer         // To store interpreter's original stdout
	// --- End tview specific fields ---

	// --- Chat Session Management Fields ---
	chatMu              sync.Mutex              // Mutex specifically for chat-related fields below
	chatSessions        map[string]*ChatSession // Stores all active chat sessions, keyed by a unique session ID
	activeChatSessionID string                  // ID of the currently focused chat session in the TUI
	nextChatIDSuffix    int                     // Counter to help generate unique display names or IDs
	// --- End Chat Session Management Fields ---
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

// Run starts the application based on the configured mode.
func (a *App) Run() error {
	a.Log.Info("NeuroScript application starting...", "version", "0.4.0") // Version can be updated

	if err := a.initializeCoreComponents(); err != nil {
		a.Log.Error("Failed to initialize core components", "error", err)
		return fmt.Errorf("core component initialization failed: %w", err)
	}

	if a.Config.SyncDir != "" {
		a.Log.Info("Sync mode selected.", "sync_dir", a.Config.SyncDir)
		return a.runSyncMode(a.appCtx)
	}
	if a.Config.StartupScript != "" && !a.Config.TuiMode { // Ensure TUI mode doesn't run script then TUI
		a.Log.Info("Script execution mode selected.", "script", a.Config.StartupScript)
		err := a.ExecuteScriptFile(a.appCtx, a.Config.StartupScript)
		if err != nil {
			a.Log.Error("Script execution failed", "script", a.Config.StartupScript, "error", err)
		}
		return err // Exit after script if not also in TUI mode
	}

	// If TuiMode is true, or no other mode is specified, default to TUI.
	a.Log.Info("Defaulting to TUI mode.")
	return a.runTuiMode(a.appCtx)
}

// runTuiMode starts the application in TUI mode.
func (a *App) runTuiMode(ctx context.Context) error {
	a.Log.Info("Starting TUI mode (tview)...")

	initialTuiScript := ""
	if a.Config.StartupScript != "" && a.Config.TuiMode {
		// If TUI mode is active and a startup script is specified, pass it to the TUI
		initialTuiScript = a.Config.StartupScript
		a.Log.Info("TUI mode will execute initial script", "script", initialTuiScript)
	}

	if interpreter := a.GetInterpreter(); interpreter != nil {
		a.originalStdout = interpreter.Stdout()
	} else {
		a.Log.Warn("Interpreter is nil before starting TUI mode, cannot save original stdout.")
	}

	err := StartTviewTUI(a, initialTuiScript) // StartTviewTUI handles App's tui field

	if interpreter := a.GetInterpreter(); interpreter != nil && a.originalStdout != nil {
		interpreter.SetStdout(a.originalStdout)
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
		// This should ideally not happen if NewApp ensures logger is set.
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
	a.mu.Lock() // Lock for setting llmClient and interpreter
	defer a.mu.Unlock()

	if a.llmClient == nil {
		var errLLM error
		// Assuming CreateLLMClient is a method on App or a helper that uses App.Config
		// This part needs to be defined or use a.Config directly.
		// For now, let's assume a.Config.APIKey and a.Config.Provider are used.
		// This is a placeholder for actual LLM client creation logic.
		// The actual creation is in app_init.go's CreateLLMClient method which NewApp calls.
		// This path in initializeCoreComponents might be for a scenario where it wasn't pre-created.
		// For now, ensure it gets a client if nil.
		if a.Config != nil { // Check if config exists to create a client
			createdClient, err := a.CreateLLMClient() // CreateLLMClient is on *App
			if err != nil {
				a.Log.Error("Failed to create LLM client during core component init", "error", err)
				// Fallback to NoOp if creation fails
				a.llmClient = adapters.NewNoOpLLMClient()
				a.Log.Info("Using default NoOpLLMClient after creation failure in core component init.")
			} else {
				a.llmClient = createdClient
				a.Log.Info("LLM Client created successfully during core component init.")
			}
		} else {
			a.llmClient = adapters.NewNoOpLLMClient()
			a.Log.Info("Using default NoOpLLMClient during core component init (config was nil).", "llm_error", errLLM)
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
		a.Config.SandboxDir = sandboxDir // Update config if it was auto-generated
	}

	if _, err := os.Stat(sandboxDir); os.IsNotExist(err) {
		if errMkdir := os.MkdirAll(sandboxDir, 0755); errMkdir != nil {
			return fmt.Errorf("failed to create specified sandbox directory %s: %w", sandboxDir, errMkdir)
		}
		a.Log.Info("Created sandbox directory", "path", sandboxDir)
	} else {
		a.Log.Info("Using existing sandbox directory", "path", sandboxDir)
	}

	// Ensure patchHandler is initialized if needed by other components
	// a.patchHandler = NewPatchHandler(a.Log) // Example initialization

	interpLLMClient := a.llmClient
	if interpLLMClient == nil {
		a.Log.Warn("LLMClient is nil for interpreter, using NoOpLLMClient for interpreter.")
		interpLLMClient = adapters.NewNoOpLLMClient()
	}

	// Ensure initialDefinitionsContent and initialPerformanceContent are handled correctly
	// For now, assuming nil/empty strings as per NewAIWorkerManager signature for file-based loading.
	var errInterp error
	// Initialize interpreter with nil for libPaths initially, can be set from config later if needed
	a.interpreter, errInterp = core.NewInterpreter(a.Log, interpLLMClient, sandboxDir, nil, a.Config.LibPaths)

	if errInterp != nil {
		return fmt.Errorf("failed to create interpreter: %w", errInterp)
	}
	if errRegister := core.RegisterCoreTools(a.interpreter); errRegister != nil {
		// Log as warning, not fatal error, to allow basic operation
		a.Log.Warn("Failed to register all core tools", "error", errRegister)
	}

	// AgentContext initialization
	a.agentCtx = NewAgentContext(a.Log)
	if a.agentCtx != nil {
		a.agentCtx.SetSandboxDir(sandboxDir)
		// These might come from config or defaults
		// a.agentCtx.SetAllowlistPath(a.Config.AllowlistFile)
		// a.agentCtx.SetModelName(a.Config.ModelName)
	}

	// Set sandbox directory on interpreter after it's created
	if errSetSandbox := a.interpreter.SetSandboxDir(sandboxDir); errSetSandbox != nil {
		// This might be a more critical error if sandbox is essential
		a.Log.Error("Failed to set sandbox dir on interpreter post init", "error", errSetSandbox)
	}

	aiWmLLMClient := a.llmClient
	if aiWmLLMClient == nil {
		aiWmLLMClient = adapters.NewNoOpLLMClient() // Fallback for AIWM
	}

	// Let NewAIWorkerManager handle loading from default paths or content
	aiWm, errManager := core.NewAIWorkerManager(a.Log, sandboxDir, aiWmLLMClient, "", "") // Empty strings for initial content
	if errManager != nil {
		a.Log.Error("Failed to initialize AIWorkerManager", "error", errManager)
		// Depending on requirements, this could be a fatal error.
	} else {
		// Attempt to load definitions from the default file path after manager creation
		if errLoadDefs := aiWm.LoadWorkerDefinitionsFromFile(); errLoadDefs != nil {
			a.Log.Warn("Could not load AI worker definitions from file", "path", aiWm.FullPathForDefinitions(), "error", errLoadDefs)
		}
		// Attempt to load performance data from the default file path
		if errLoadPerf := aiWm.LoadRetiredInstancePerformanceDataFromFile(); errLoadPerf != nil {
			a.Log.Warn("Could not load AI worker performance data from file", "path", aiWm.FullPathForPerformanceData(), "error", errLoadPerf)
		}
	}

	if a.interpreter != nil && aiWm != nil {
		a.interpreter.SetAIWorkerManager(aiWm)
	} else if a.interpreter == nil {
		a.Log.Error("Interpreter not initialized, cannot set AIWorkerManager.")
	}

	// Initialize chat session map
	a.chatSessions = make(map[string]*ChatSession)

	a.Log.Info("Core components initialization attempt finished.")
	return nil
}

// HandleSystemCommand (existing method, ensure no conflicts)
func (a *App) HandleSystemCommand(command string) {
	// ... (existing implementation)
	// Consider if system commands should interact with chat (e.g., //endchat)
	if a.Log == nil {
		fmt.Printf("App.Log is nil, cannot log system command: %s\n", command)
		return
	}
	a.Log.Info("System command received by App (from TUI)", "command", command)
}

// ExecuteScriptLine (existing method, ensure no conflicts)
func (a *App) ExecuteScriptLine(ctx context.Context, line string) {
	// ... (existing implementation)
	if a.Log == nil {
		fmt.Printf("App.Log is nil, cannot log script line: %s\n", line)
		return
	}
	a.Log.Info("Script line received by App (from TUI)", "line", line)
}

// CreateLLMClient function (as it appeared in app_init.go, now part of app.go for self-containment, can be called by NewApp or initializeCoreComponents)
// This function creates an LLM client based on application configuration.
func (app *App) CreateLLMClient() (core.LLMClient, error) {
	if app.Config == nil {
		return nil, fmt.Errorf("cannot create LLM client: app config is nil")
	}

	apiKey := app.Config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("NEUROSCRIPT_API_KEY") // Standardized env var name
		if apiKey == "" {
			if app.Log != nil { // Check logger
				app.Log.Debug("API key is missing in config and environment variable (NEUROSCRIPT_API_KEY). Creating NoOpLLMClient.")
			}
			return adapters.NewNoOpLLMClient(), nil
		}
		if app.Log != nil {
			app.Log.Debug("Using LLM API key from environment variable NEUROSCRIPT_API_KEY.")
		}
	} else {
		if app.Log != nil {
			app.Log.Debug("Using LLM API key from configuration.")
		}
	}

	if app.Log != nil {
		app.Log.Debug("Creating real LLMClient.")
	}
	apiHost := app.Config.APIHost
	modelName := app.Config.ModelName // This is ModelID for NewLLMClient

	loggerToUse := app.Log
	if loggerToUse == nil { // Should be set by NewApp
		loggerToUse = adapters.NewNoOpLogger() // Safety fallback
	}

	llmClient := core.NewLLMClient(apiKey, apiHost, modelName, loggerToUse, !app.Config.Insecure)

	if llmClient == nil {
		if app.Log != nil {
			app.Log.Error("core.NewLLMClient returned nil unexpectedly. This indicates a critical LLM client creation failure.")
		}
		return adapters.NewNoOpLLMClient(), fmt.Errorf("core.NewLLMClient returned nil for real client creation")
	}
	if app.Log != nil {
		app.Log.Debug("Real LLMClient created.", "host", apiHost, "model", modelName)
	}
	return llmClient, nil
}
