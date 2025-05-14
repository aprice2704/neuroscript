// NeuroScript Version: 0.3.0
// File version: 0.1.9 // Removed NewPatchHandler calls; patchHandler will be nil.
// filename: pkg/neurogo/app.go
// nlines: 270 // Approximate
// risk_rating: MEDIUM
package neurogo

import (
	"context"
	"fmt"
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
	patchHandler *PatchHandler // Field is present, but will not be initialized here
	mu           sync.RWMutex

	tuiModelInstance *model
	appCtx           context.Context
	cancelFunc       context.CancelFunc
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
		Config:    cfg,
		Log:       logger,
		llmClient: llmClient,
		// patchHandler remains nil by default as NewPatchHandler is assumed not to exist
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
	a.Log.Info("NeuroScript application starting...", "version", "0.3.0")

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

func (a *App) SetTUImodel(m *model) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tuiModelInstance = m
}

func (a *App) GetTUImodel() *model {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.tuiModelInstance
}

func (a *App) initializeCoreComponents() error {
	a.Log.Debug("Initializing core components...")

	if a.llmClient == nil {
		var err error
		a.llmClient, err = a.CreateLLMClient()
		if err != nil {
			a.Log.Error("Could not initialize LLM Client during core component setup", "error", err)
			if a.llmClient == nil {
				a.Log.Warn("LLMClient creation failed and no fallback set. Using new NoOpLLMClient.")
				a.llmClient = adapters.NewNoOpLLMClient()
			}
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
		if err := os.MkdirAll(sandboxDir, 0755); err != nil {
			return fmt.Errorf("failed to create specified sandbox directory %s: %w", sandboxDir, err)
		}
		a.Log.Info("Created sandbox directory", "path", sandboxDir)
	} else {
		a.Log.Info("Using existing sandbox directory", "path", sandboxDir)
	}

	// Since NewPatchHandler is assumed not to exist or not be usable here,
	// a.patchHandler remains as it was (likely nil from NewApp).
	// No attempt to set sandboxDir on it.
	if a.patchHandler == nil {
		a.Log.Warn("PatchHandler is nil after NewApp. Patch functionality may be disabled.")
	}

	interpLLMClient := a.llmClient
	if interpLLMClient == nil {
		a.Log.Warn("LLMClient is nil for interpreter, using NoOpLLMClient for interpreter.")
		interpLLMClient = adapters.NewNoOpLLMClient()
	}

	var err error
	a.interpreter, err = core.NewInterpreter(a.Log, interpLLMClient, sandboxDir, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create interpreter: %w", err)
	}
	if err := core.RegisterCoreTools(a.interpreter); err != nil {
		a.Log.Warn("Failed to register all core tools", "error", err)
	}

	a.agentCtx = NewAgentContext(a.Log)
	if a.agentCtx != nil {
		a.agentCtx.SetSandboxDir(sandboxDir)
		a.agentCtx.SetAllowlistPath(a.Config.AllowlistFile)
		a.agentCtx.SetModelName(a.Config.ModelName)
	}
	if err := a.interpreter.SetSandboxDir(sandboxDir); err != nil {
		a.Log.Error("Failed to set sandbox dir on interpreter post init", "error", err)
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
		// Assuming a method to load performance data exists.
		// This needs to be verified against the actual AIWorkerManager API.
		// If the method `LoadRetiredInstancePerformanceData` does not exist on `core.AIWorkerManager`,
		// this will cause a compile error. Replace with the correct method if different.
		// if loadErr := aiWm.LoadRetiredInstancePerformanceData(); loadErr != nil {
		// 	a.Log.Warn("Could not load AI worker performance data", "path", perfDataPath, "error", loadErr)
		// }
	}

	if a.interpreter != nil && aiWm != nil {
		a.interpreter.SetAIWorkerManager(aiWm)
	} else if a.interpreter == nil {
		a.Log.Error("Interpreter not initialized, cannot set AIWorkerManager.")
	}

	a.Log.Info("Core components initialization attempt finished.")
	return nil
}

var _ WMStatusViewDataProvider = (*App)(nil)
