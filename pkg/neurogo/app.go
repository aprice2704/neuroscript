// filename: pkg/neurogo/app.go
package neurogo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync" // Added for Mutex

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/toolsets"

	"github.com/aprice2704/neuroscript/pkg/neurodata/models" // Keep the import for type signature
	// Import nspatch package (even if NewHandler is gone)
)

// App orchestrates the NeuroScript agent, TUI, and script execution modes.
type App struct {
	Config *Config
	Log    logging.Logger // Use the Logger interface

	// LLM Interaction
	llmClient core.LLMClient

	// Interpreter and State
	interpreter *core.Interpreter
	mu          sync.Mutex // Protect access to shared state if needed

	// TUI components (if TUI is enabled)
	// tui *tea.Program // Assuming bubbletea is used

	// Patch Handling
	patchHandler PatchHandler // Use the interface defined in app_interface.go

	// Loaded Schema (optional, depends on application needs)
	loadedSchema *models.Schema // Keep the field for future use
}

// NewApp creates a new NeuroGo application instance.
func NewApp(logger logging.Logger) *App {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Critical Warning: NewApp called with nil logger. Using fallback stderr logging.\n")
	} else {
		logger.Debug("Creating new App instance.")
	}
	return &App{
		Log: logger,
	}
}

// Run starts the application based on the configuration.
func (app *App) Run(ctx context.Context) error {
	if app.Config == nil {
		errMsg := "application config is nil"
		if app.Log != nil {
			app.Log.Error(errMsg)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
		}
		return fmt.Errorf("%s", errMsg)
	}
	if app.Log == nil {
		fmt.Fprintf(os.Stderr, "Error: application logger is nil\n")
		return fmt.Errorf("application logger is nil")
	}

	app.Log.Info("Starting NeuroGo application...")
	app.Log.Debug("Configuration loaded.", "config", fmt.Sprintf("%+v", app.Config))

	// --- Load Schema (Placeholder) ---
	schemaPath := app.Config.SchemaPath // Use field added to Config
	if schemaPath != "" {
		var loadErr error
		app.loadedSchema, loadErr = app.loadSchema(schemaPath)
		if loadErr != nil {
			app.Log.Warn("Schema loading attempted but failed (using placeholder).", "path", schemaPath, "error", loadErr)
		} else if app.loadedSchema != nil {
			app.Log.Info("Schema loaded successfully (using placeholder).", "path", schemaPath, "name", app.loadedSchema.Name, "version", app.loadedSchema.Version)
		} else {
			app.Log.Info("Schema loading skipped (using placeholder).", "path", schemaPath)
		}
	} else {
		app.Log.Info("No schema path configured, skipping schema load.")
	}
	// --- End Schema Load ---

	// --- Initialize LLM Client ---
	var llmErr error
	app.llmClient, llmErr = app.createLLMClient()
	if llmErr != nil {
		// Log the specific error before returning a generic one might be useful
		app.Log.Error("LLM Client initialization failed", "error", llmErr)
		return fmt.Errorf("failed to initialize LLM client: %w", llmErr)
	}
	app.Log.Info("LLM Client initialized.")

	// --- Initialize Interpreter ---
	app.interpreter, _ = core.NewInterpreter(app.Log, app.llmClient, "", nil)
	if app.Config.SandboxDir != "" {
		app.interpreter.SetSandboxDir(app.Config.SandboxDir)
		app.Log.Info("Interpreter sandbox directory configured.", "path", app.Config.SandboxDir)
	} else {
		app.Log.Warn("No sandbox directory configured, interpreter using default.")
	}
	app.Log.Info("Interpreter initialized.")

	app.Log.Info("Registering extended toolsets...")                           // Log before calling
	err := toolsets.RegisterExtendedTools(app.GetInterpreter().ToolRegistry()) // <-- Capture the error
	if err != nil {
		// Log the error from toolsets registration
		app.Log.Error("Failed to register one or more extended toolsets", "error", err)
		// --- Decide how critical this is ---
		// Option 1: Log and continue (if tools are optional)
		// Option 2: Return the error to stop app startup (if tools are essential)
		return fmt.Errorf("failed during extended tool registration: %w", err)
		// --- End Decision ---
	}
	app.Log.Info("Extended toolsets registered successfully.") // Log success

	// --- Initialize Patch Handler ---
	if app.interpreter == nil {
		return fmt.Errorf("cannot initialize patch handler: interpreter is nil")
	}
	fileAPI := app.interpreter.FileAPI()
	if fileAPI == nil {
		// This check might be redundant if NewInterpreter guarantees a non-nil FileAPI
		app.Log.Warn("Interpreter FileAPI is nil, patch handler may not function correctly.")
		// return fmt.Errorf("cannot initialize patch handler: interpreter FileAPI is nil")
	}
	// >>> FIX: nspatch.NewHandler is undefined. Using nil as placeholder. <<<
	// app.patchHandler = nspatch.NewHandler(fileAPI, app.Log) // Original error line
	app.patchHandler = nil // Placeholder - Patching will not work!
	if app.patchHandler == nil {
		app.Log.Warn("Patch Handler initialization skipped (nspatch.NewHandler undefined). Patching functionality disabled.")
	} else {
		app.Log.Info("Patch Handler initialized.")
	}

	// --- Mode Dispatch ---
	app.Log.Debug("Dispatching based on run mode flags.")
	switch {
	case app.Config.RunScriptMode:
		app.Log.Info("Running in Script Mode.")
		return app.runScriptMode(ctx)
	case app.Config.RunTuiMode: // <<< FIX: Correct case for RunTuiMode
		app.Log.Info("Running in TUI Mode.")
		return app.runTuiMode(ctx) // <<< FIX: Correct case for runTuiMode
	case app.Config.RunSyncMode:
		app.Log.Info("Running in Sync Mode.")
		return app.runSyncMode(ctx)
	// Add other modes as needed
	// case app.Config.RunCleanAPIMode:
	// 	...
	default:
		if !app.Config.EnableLLM {
			return fmt.Errorf("cannot run in default (Agent) mode: LLM must be enabled (check --enable-llm or config), and no other mode was specified")
		}
		app.Log.Info("Running in Agent Mode (default).")
		return app.runAgentMode(ctx)
	}
}

// GetInterpreter returns the application's core interpreter instance.
func (app *App) GetInterpreter() *core.Interpreter {
	if app.interpreter == nil {
		app.Log.Error("GetInterpreter called before interpreter was initialized!")
	}
	return app.interpreter
}

// loadSchema loads the NeuroData schema from the specified path.
// *** PLACEHOLDER IMPLEMENTATION ***
func (app *App) loadSchema(schemaPath string) (*models.Schema, error) {
	if schemaPath == "" {
		app.Log.Info("loadSchema called with empty path, skipping.")
		return nil, nil
	}
	app.Log.Warn("Schema loading requested, but using placeholder implementation.", "path", schemaPath)
	return nil, nil
}

// findProjectRoot searches upwards from the current directory for a marker file.
func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			fmt.Fprintf(os.Stderr, "Warning: Project root marker (.git) not found upwards from %s. Defaulting to CWD.\n", cwd)
			return cwd, nil
		}
		dir = parent
	}
}

// createLLMClient creates the LLM client based on config.
func (app *App) createLLMClient() (core.LLMClient, error) {
	if !app.Config.EnableLLM {
		app.Log.Info("LLM is disabled by config, creating NoOpLLMClient.")
		return core.NewNoOpLLMClient(app.Log), nil
	}

	app.Log.Info("LLM is enabled, creating real LLMClient.")
	apiKey := app.Config.APIKey
	apiHost := app.Config.APIHost // <<< FIX: Use added field
	modelID := app.Config.ModelID // <<< FIX: Use added field

	if apiKey == "" {
		apiKey = os.Getenv("NEUROSCRIPT_API_KEY") // Standardized env var name
		if apiKey == "" {
			app.Log.Error("LLM is enabled, but API key is missing in config and environment variable (NEUROSCRIPT_API_KEY).")
			return nil, fmt.Errorf("LLM API key is required but not found")
		}
		app.Log.Info("Using LLM API key from environment variable.")
	} else {
		app.Log.Debug("Using LLM API key from configuration.")
	}

	llmClient := core.NewLLMClient(apiKey, apiHost, app.Log, true)
	if llmClient == nil {
		app.Log.Error("NewLLMClient returned nil unexpectedly.")
		return nil, fmt.Errorf("failed to create LLM client instance (NewLLMClient returned nil)")
	}

	// Here you might want to explicitly set the model on the client if the factory doesn't handle it
	// For example: if hasattr(llmClient, 'SetModel'): llmClient.SetModel(modelID)

	app.Log.Info("Real LLMClient created.", "host", apiHost, "model", modelID)
	return llmClient, nil
}

// --- Methods implementing AppAccess interface for TUI ---

func (a *App) GetModelName() string {
	return a.Config.ModelID // <<< FIX: Use added field
}

func (a *App) GetSyncDir() string {
	return a.Config.SyncDir
}

func (a *App) GetSandboxDir() string {
	return a.Config.SandboxDir
}

func (a *App) GetSyncFilter() string {
	return a.Config.SyncFilter
}

func (a *App) GetSyncIgnoreGitignore() bool {
	return a.Config.SyncIgnoreGitignore
}

func (a *App) GetLogger() logging.Logger {
	if a.Log == nil {
		fmt.Fprintf(os.Stderr, "Warning: GetLogger called when App.Log is nil. Returning nil.\n")
		return nil
	}
	return a.Log
}

func (a *App) GetLLMClient() core.LLMClient {
	if a.llmClient == nil {
		if a.Log != nil {
			a.Log.Warn("GetLLMClient called when App.llmClient is nil.")
		} else {
			fmt.Fprintf(os.Stderr, "Warning: GetLLMClient called when App.llmClient is nil.\n")
		}
		return nil
	}
	return a.llmClient
}
