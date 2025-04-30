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

	"github.com/aprice2704/neuroscript/pkg/nspatch" // Import nspatch
)

// App orchestrates the NeuroScript agent, TUI, and script execution modes.
type App struct {
	Config *Config
	Log    logging.Logger // Use the Logger interface

	// LLM Interaction
	// Use the LLMClient interface from pkg/interfaces
	llmClient core.LLMClient

	// Interpreter and State
	interpreter *core.Interpreter
	// agentState  *AgentState // Consider defining AgentState struct if complex
	mu sync.Mutex // Protect access to shared state if needed

	// TUI components (if TUI is enabled)
	// tui *tea.Program // Assuming bubbletea is used

	// Patch Handling
	patchHandler PatchHandler // Use the interface defined in app_interface.go

	// Other components like vector store, file syncer etc.
	// vectorStore VectorStoreInterface
	// fileSyncer  FileSyncerInterface
}

// NewApp creates a new NeuroGo application instance.
// It requires a logger. The LLM client is configured later based on Config.
func NewApp(logger logging.Logger) *App {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Warning: NewApp called with nil logger\n")
		// Potentially create a default logger here if absolutely necessary
		// logger = someDefaultLogger()
	}
	return &App{
		Log: logger,
		// Config and other fields initialized later or in Run()
	}
}

// Run starts the application based on the configuration.
func (app *App) Run(ctx context.Context) error {
	if app.Config == nil {
		return fmt.Errorf("application config is nil")
	}
	if app.Log == nil {
		// This should ideally not happen if NewApp ensures logger is non-nil
		return fmt.Errorf("application logger is nil")
	}

	app.Log.Info("Starting NeuroGo application...")
	app.Log.Debug("Configuration:", "config", fmt.Sprintf("%+v", app.Config)) // Log config details

	// --- Initialize LLM Client ---
	// Moved initialization here from NewApp to use Config
	var err error
	app.llmClient, err = app.createLLMClient() // Use helper to create client
	if err != nil {
		return fmt.Errorf("failed to initialize LLM client: %w", err)
	}
	if app.llmClient == nil && app.Config.EnableLLM {
		// If LLM is enabled but client creation failed silently (shouldn't happen with error check)
		return fmt.Errorf("LLM client is nil despite LLM being enabled")
	}
	app.Log.Info("LLM Client initialized.")

	// --- Initialize Interpreter ---
	// Pass the potentially NoOp LLM client to the interpreter
	app.interpreter = core.NewInterpreter(app.Log, app.llmClient)
	app.Log.Info("Interpreter initialized.")

	// --- Initialize Patch Handler ---
	// Default patch handler using the interpreter's file API
	// Ensure interpreter is initialized first
	app.patchHandler = nspatch.NewHandler(app.interpreter.FileAPI, app.Log)
	app.Log.Info("Patch Handler initialized.")

	// --- Mode Dispatch ---
	switch {
	case app.Config.RunScriptMode:
		app.Log.Info("Running in Script Mode.")
		return app.runScriptMode(ctx)
	case app.Config.RunTUIMode:
		app.Log.Info("Running in TUI Mode.")
		return app.runTUIMode(ctx)
	case app.Config.RunSyncMode:
		app.Log.Info("Running in Sync Mode.")
		return app.runSyncMode(ctx)
	default:
		// Default to Agent mode if no specific mode is set? Or return error?
		// Let's assume Agent mode is the default interactive mode if TUI is off.
		if !app.Config.EnableLLM {
			return fmt.Errorf("cannot run in default Agent mode: LLM must be enabled (check --enable-llm or config)")
		}
		app.Log.Info("Running in Agent Mode (default).")
		return app.runAgentMode(ctx)
	}
}

// loadSchema loads the NeuroData schema from the specified path.
// Placeholder implementation.
func (app *App) loadSchema(schemaPath string) (*schema.Schema, error) {
	app.Log.Info("Loading schema...", "path", schemaPath)
	// TODO: Implement actual schema loading logic from neurodata/schema package
	if schemaPath == "" {
		app.Log.Warn("Schema path is empty, using default or no schema.")
		return nil, nil // Or return a default schema
	}
	// _, err := os.Stat(schemaPath)
	// if os.IsNotExist(err) {
	// 	 return nil, fmt.Errorf("schema file not found at %s", schemaPath)
	// } else if err != nil {
	// 	 return nil, fmt.Errorf("error checking schema file %s: %w", schemaPath, err)
	// }
	// Actual parsing logic needed here
	return nil, fmt.Errorf("schema loading not yet implemented")
}

// findProjectRoot searches upwards from the current directory for a marker file.
// Placeholder implementation.
func findProjectRoot() (string, error) {
	// TODO: Implement project root finding logic (e.g., look for .neuroproject, .git)
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	dir := cwd
	// Simplified: Assume cwd is project root for now
	// Add logic to search upwards for a marker like ".git" or a specific project file
	for {
		// Check for marker file (e.g., .git directory)
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir, nil // Found project root
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory without finding marker
			return cwd, fmt.Errorf("project root marker (.git) not found upwards from %s", cwd) // Or return cwd as default?
		}
		dir = parent
	}
	// return cwd, nil // Default to current dir if no marker found? Needs decision.
}

// Helper to create the LLM client based on config
func (app *App) createLLMClient() (core.LLMClient, error) {
	if !app.Config.EnableLLM {
		app.Log.Info("LLM is disabled, using NoOpLLMClient.")
		// We need an adapter for NoOpLLMClient, assuming it's in pkg/adapters
		// Ensure adapters package is created and imported if needed
		// return adapters.NewNoOpLLMClient(), nil // Assuming NewNoOpLLMClient exists
		// TEMPORARY: Return nil until NoOpLLMClient adapter is confirmed/created
		// return nil, fmt.Errorf("NoOpLLMClient adapter not available")
		// If core defines a NoOp client directly:
		return core.NewNoOpLLMClient(app.Log), nil // Use core's NoOp client
	}

	app.Log.Info("LLM is enabled, creating real LLMClient.")
	// Use configuration values
	apiKey := app.Config.APIKey
	apiHost := app.Config.APIHost
	modelID := app.Config.ModelID // Assuming ModelID is added to Config

	if apiKey == "" {
		// Attempt to get from environment variable if not in config
		apiKey = os.Getenv("NEUROSCRIPT_API_KEY") // Example env var name
		if apiKey == "" {
			return nil, fmt.Errorf("LLM API key is required but not found in config or environment variables")
		}
		app.Log.Info("Using LLM API key from environment variable.")
	}

	// Use the core LLM client factory function
	// Assuming NewLLMClient handles different providers based on host/config later
	llmClient := core.NewLLMClient(apiKey, apiHost, app.Log, true) // Pass true for enabled
	if llmClient == nil {
		// NewLLMClient should ideally return an error, but handle nil just in case
		return nil, fmt.Errorf("failed to create LLM client instance (NewLLMClient returned nil)")
	}

	app.Log.Info("Real LLMClient created.", "host", apiHost, "model", modelID)
	// We might need to configure the specific model on the client here if NewLLMClient doesn't handle it
	// e.g., if llmClient has a SetModel(modelID) method

	return llmClient, nil
}
