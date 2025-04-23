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
)

// App encapsulates the application's state and configuration.
type App struct {
	Config   *Config
	InfoLog  *log.Logger
	WarnLog  *log.Logger // Keep the field
	ErrorLog *log.Logger
	DebugLog *log.Logger
	LLMLog   *log.Logger

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
	if app.InfoLog == nil {
		app.InfoLog = log.New(io.Discard, "INFO: ", log.LstdFlags)
	}
	if app.WarnLog == nil {
		app.WarnLog = log.New(io.Discard, "WARN: ", log.LstdFlags|log.Lshortfile)
	} // Ensure WarnLog is also non-nil
	if app.ErrorLog == nil {
		app.ErrorLog = log.New(io.Discard, "ERROR: ", log.LstdFlags|log.Lshortfile)
	}
	if app.DebugLog == nil {
		app.DebugLog = log.New(io.Discard, "DEBUG: ", log.LstdFlags|log.Lshortfile)
	}
	if app.LLMLog == nil {
		app.LLMLog = log.New(io.Discard, "LLM:   ", log.LstdFlags|log.Lshortfile)
	}

	// Register standard and agent-specific tools
	// Standard tools first
	if err := core.RegisterCoreTools(interpreter.ToolRegistry()); err != nil {
		app.GetErrorLogger().Printf("CRITICAL: Failed to register standard tools: %v", err)
		// Potentially panic or os.Exit(1) here if standard tools are essential
	}
	// Agent-specific tools (lives in this package)
	if err := RegisterAgentTools(interpreter.ToolRegistry()); err != nil {
		app.GetErrorLogger().Printf("CRITICAL: Failed to register agent tools: %v", err)
		// Potentially panic or os.Exit(1)
	}

	return app
}

// --- ADDED BACK: GetWarnLogger ---
// Ensure non-nil WarnLog and return it.
func (a *App) GetWarnLogger() *log.Logger {
	if a.WarnLog == nil {
		a.WarnLog = log.New(io.Discard, "WARN-FALLBACK: ", log.LstdFlags|log.Lshortfile)
	}
	return a.WarnLog
}

// --- Logger Getters Removed (defined in app_interface.go) ---
// func (a *App) GetInfoLogger() *log.Logger { ... }
// func (a *App) GetErrorLogger() *log.Logger { ... }
// func (a *App) GetDebugLogger() *log.Logger { ... }

// --- GetLLMClient Removed (defined in app_interface.go) ---
// func (a *App) GetLLMClient() *core.LLMClient { ... }

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
		a.GetErrorLogger().Printf("CRITICAL: Logging initialization failed: %v", err)
		return fmt.Errorf("logging initialization failed: %w", err)
	}

	// Use loggers safely now (use getters defined in app_interface.go or here)
	infoLog := a.GetInfoLogger()
	errLog := a.GetErrorLogger()
	debugLog := a.GetDebugLogger()
	warnLog := a.GetWarnLogger() // Now defined again

	// Ensure config is not nil before proceeding
	if a.Config == nil {
		errMsg := "internal error: App.Config is nil at start of Run"
		errLog.Println(errMsg) // Use getter now that it's safe
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
	debugLog.Printf("Run: initLLMClient finished. Error: %v, Client Available: %v, Needs LLM: %v", initErr, llmClientAvailable, needsLLM)

	if initErr != nil {
		errLog.Printf("LLM Client initialization failed: %v", initErr)
		if needsLLM {
			return fmt.Errorf("LLM client required but initialization failed: %w", initErr)
		} else {
			warnLog.Printf("Continuing without LLM client despite initialization error: %v", initErr)
		}
	} else if needsLLM && !llmClientAvailable {
		return errors.New("LLM client required but is unavailable after initialization")
	}

	// 3. Update interpreter with the (potentially nil) LLMClient
	// --- REMOVED direct assignment to interpreter fields ---
	// a.interpreter.LLMClient = a.llmClient // REMOVED - Interpreter handles its client via constructor

	// 4. Select Mode Based on Config
	infoLog.Println("Run: Dispatching based on mode...")

	// Mode precedence was handled in main.go which set the boolean flags.
	if a.Config.RunCleanAPIMode {
		infoLog.Println("--- Running in Clean API Mode ---")
		if !llmClientAvailable {
			return errors.New("LLM Client (for File API) required for clean-api mode but is unavailable")
		}
		return a.runCleanAPIMode(ctx) // Assumes exists in app_helpers.go or similar
	} else if a.Config.RunSyncMode {
		infoLog.Println("--- Running in Sync Mode ---")
		if !llmClientAvailable {
			return errors.New("LLM Client (for File API) required for sync mode but is unavailable")
		}
		return a.runSyncMode(ctx) // Assumes exists in app_sync.go
	} else if a.Config.RunScriptMode {
		infoLog.Println("--- Running in Script Mode ---")
		return a.runScriptMode(ctx) // Assumes exists in app_script.go
	} else if a.Config.RunTuiMode {
		infoLog.Println("--- Running in TUI Mode ---")
		return a.runTuiMode(ctx) // Assumes exists in app_tui.go
	} else if a.Config.RunAgentMode {
		infoLog.Println("--- Running in Agent Mode ---")
		if !llmClientAvailable {
			return errors.New("LLM Client required for agent mode but is unavailable")
		}
		return a.runAgentMode(ctx) // Assumes exists in app_agent.go
	} else {
		errMsg := "internal error: no execution mode determined"
		errLog.Println(errMsg)
		return errors.New(errMsg)
	}
}

// NOTE: Ensure methods like runCleanAPIMode, runSyncMode, initLogging, initLLMClient etc.
// are defined either here or (more likely) in their respective app_*.go files (app_helpers, app_sync, etc).
