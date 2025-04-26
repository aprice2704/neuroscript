// filename: pkg/neurogo/app.go
// UPDATED: Add back GetWarnLogger, remove direct assignment to interpreter.LLMClient
package neurogo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	// Needed for error joining fallback
	"github.com/aprice2704/neuroscript/pkg/core" // Import core
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// App encapsulates the application's state and configuration.
type App struct {
	Config *Config
	Logger interfaces.Logger
	LLMLog interfaces.Logger

	// Core components needed by the app
	interpreter *core.Interpreter // Keep interpreter internal
	llmClient   *core.LLMClient   // Keep LLM client internal
	agentCtx    *AgentContext     // Added: Hold the agent context if in agent mode
}

// NewApp creates a new App instance with default loggers and registers agent tools.
func NewApp() *App {
	// Initialize core logger first for components that need it
	tmpLogger := log.New(os.Stderr, "INIT: ", log.LstdFlags|log.Lshortfile) // Temporary logger

	// Create components (LLM client might be nil initially)
	var llmClient *core.LLMClient                            // Initialize as nil, will be created in initLLMClient
	interpreter := core.NewInterpreter(tmpLogger, llmClient) // Pass nil client for now

	app := &App{
		Config:   NewConfig(),
		InfoLog:  log.New(os.Stdout, "INFO:  ", log.LstdFlags),
		WarnLog:  log.New(os.Stderr, "WARN:  ", log.LstdFlags|log.Lshortfile),
		ErrorLog: log.New(os.Stderr, "ERROR: ", log.LstdFlags|log.Lshortfile),
		DebugLog: log.New(io.Discard, "DEBUG: ", log.LstdFlags|log.Lshortfile), // Start discarded
		LLMLog:   log.New(io.Discard, "LLM:   ", log.LstdFlags|log.Lshortfile), // Start discarded

		// Assign core components
		interpreter: interpreter,
		llmClient:   llmClient, // Still nil here
		// agentCtx will be created when agent mode starts
	}

	// Ensure essential loggers are non-nil right away
	if app.Logger == nil {
		panic("App needs a valid logger")
	}
	if app.LLMLog == nil {
		panic("App needs a valid LLM logger")
	}

	// Register standard and agent-specific tools
	// Standard tools first
	if err := core.RegisterCoreTools(interpreter.ToolRegistry()); err != nil {
		app.GetLogger().Debug("CRITICAL: Failed to register standard tools: %v", err)
		// Potentially panic or os.Exit(1) here if standard tools are essential
	}
	// Agent-specific tools (lives in this package)
	if err := RegisterAgentTools(interpreter.ToolRegistry()); err != nil {
		app.GetLogger().Debug("CRITICAL: Failed to register agent tools: %v", err)
		// Potentially panic or os.Exit(1)
	}

	return app
}

// GetInterpreter returns the application's NeuroScript interpreter.
func (a *App) GetInterpreter() *core.Interpreter {
	return a.interpreter
}

// Run executes the appropriate application mode based on parsed flags.
// Assumes Config field is already populated by main.go
func (a *App) Run(ctx context.Context) error {
	// 1. Initialize logging based on Config
	if err := a.initLogging(); err != nil { // Call initLogging from app_helpers.go
		// Log to initial stderr logger if possible
		a.GetLogger().Debug("CRITICAL: Logging initialization failed: %v", err)
		return fmt.Errorf("logging initialization failed: %w", err)
	}

	// Ensure config is not nil before proceeding
	if a.Config == nil {
		errMsg := "internal error: App.Config is nil at start of Run"
		a.Logger.Info(errMsg) // Use getter now that it's safe
		return errors.New(errMsg)
	}

	// 2. Initialize LLM Client if needed for the selected mode
	needsLLM := false
	if a.Config.RunAgentMode || a.Config.RunTuiMode || a.Config.RunSyncMode || a.Config.RunCleanAPIMode {
		needsLLM = true
	} else if a.Config.RunScriptMode && a.Config.EnableLLM {
		needsLLM = true
	}

	initErr := a.initLLMClient(ctx) // Attempt initialization (defined in app_helpers.go)
	llmClientAvailable := a.llmClient != nil && a.llmClient.Client() != nil
	a.Logger.Debug("Run: initLLMClient finished. Error: %v, Client Available: %v, Needs LLM: %v", initErr, llmClientAvailable, needsLLM)

	if initErr != nil {
		a.Logger.Error("LLM Client initialization failed: %v", initErr)
		if needsLLM {
			return fmt.Errorf("LLM client required but initialization failed: %w", initErr)
		} else {
			a.Logger.Warn("Continuing without LLM client despite initialization error: %v", initErr)
		}
	} else if needsLLM && !llmClientAvailable {
		return errors.New("LLM client required but is unavailable after initialization")
	}

	// 3. Update interpreter with the (potentially nil) LLMClient
	// --- REMOVED direct assignment to interpreter fields ---
	// a.interpreter.LLMClient = a.llmClient // REMOVED - Interpreter handles its client via constructor

	// 4. Select Mode Based on Config
	a.Logger.Info("Run: Dispatching based on mode...")

	// Mode precedence was handled in main.go which set the boolean flags.
	if a.Config.RunCleanAPIMode {
		a.Logger.Info("--- Running in Clean API Mode ---")
		if !llmClientAvailable {
			return errors.New("LLM Client (for File API) required for clean-api mode but is unavailable")
		}
		return a.runCleanAPIMode(ctx) // Assumes exists in app_helpers.go or similar
	} else if a.Config.RunSyncMode {
		a.Logger.Info("--- Running in Sync Mode ---")
		if !llmClientAvailable {
			return errors.New("LLM Client (for File API) required for sync mode but is unavailable")
		}
		return a.runSyncMode(ctx) // Assumes exists in app_sync.go
	} else if a.Config.RunScriptMode {
		a.Logger.Info("--- Running in Script Mode ---")
		return a.runScriptMode(ctx) // Assumes exists in app_script.go
	} else if a.Config.RunTuiMode {
		a.Logger.Info("--- Running in TUI Mode ---")
		return a.runTuiMode(ctx) // Assumes exists in app_tui.go
	} else if a.Config.RunAgentMode {
		a.Logger.Info("--- Running in Agent Mode ---")
		if !llmClientAvailable {
			return errors.New("LLM Client required for agent mode but is unavailable")
		}
		return a.runAgentMode(ctx) // Assumes exists in app_agent.go
	} else {
		errMsg := "internal error: no execution mode determined"
		a.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
}

// NOTE: Ensure methods like runCleanAPIMode, runSyncMode, initLogging, initLLMClient etc.
// are defined either here or (more likely) in their respective app_*.go files (app_helpers, app_sync, etc).
