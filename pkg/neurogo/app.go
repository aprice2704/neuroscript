// filename: pkg/neurogo/app.go
package neurogo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	// Needed for Run method's error check
	"github.com/aprice2704/neuroscript/pkg/core"
)

// App encapsulates the application's state and configuration.
type App struct {
	Config   *Config
	InfoLog  *log.Logger // Needed by interface
	WarnLog  *log.Logger
	ErrorLog *log.Logger // Needed by interface
	DebugLog *log.Logger // Needed by interface
	LLMLog   *log.Logger
	Insecure *bool // Keep if still used by Config

	llmClient *core.LLMClient // Unexported, managed internally
}

// NewApp creates a new App instance with default loggers.
func NewApp() *App {
	app := &App{
		Config:   NewConfig(), // Assuming NewConfig initializes defaults
		InfoLog:  log.New(os.Stdout, "INFO:  ", log.LstdFlags),
		WarnLog:  log.New(os.Stderr, "WARN:  ", log.LstdFlags|log.Lshortfile),
		ErrorLog: log.New(os.Stderr, "ERROR: ", log.LstdFlags|log.Lshortfile),
		DebugLog: log.New(io.Discard, "DEBUG: ", log.LstdFlags|log.Lshortfile),
		LLMLog:   log.New(io.Discard, "LLM:   ", log.LstdFlags|log.Lshortfile),
	}
	// Ensure essential loggers are non-nil
	if app.InfoLog == nil {
		app.InfoLog = log.New(io.Discard, "INFO: ", log.LstdFlags)
	}
	if app.WarnLog == nil {
		app.WarnLog = log.New(io.Discard, "WARN: ", log.LstdFlags|log.Lshortfile)
	}
	if app.ErrorLog == nil {
		app.ErrorLog = log.New(io.Discard, "ERROR: ", log.LstdFlags|log.Lshortfile)
	}
	if app.DebugLog == nil {
		app.DebugLog = log.New(io.Discard, "DEBUG: ", log.LstdFlags|log.Lshortfile)
	}
	if app.LLMLog == nil {
		app.LLMLog = log.New(io.Discard, "LLM:   ", log.LstdFlags|log.Lshortfile)
	}
	return app
}

// Run executes the appropriate application mode based on parsed flags.
func (a *App) Run(ctx context.Context) error {
	// 1. Initialize logging first
	if err := a.initLogging(); err != nil {
		log.Printf("CRITICAL: Logging initialization failed: %v", err) // Use standard log
		return fmt.Errorf("logging initialization failed: %w", err)
	}

	// Use loggers safely now (use getters to ensure non-nil)
	infoLog := a.GetInfoLogger()
	errLog := a.GetErrorLogger()
	debugLog := a.GetDebugLogger()
	warnLog := a.WarnLog // Use field directly, check for nil if using

	// 2. Initialize LLM Client
	debugLog.Println("Run: Calling initLLMClient...")
	initErr := a.initLLMClient(ctx)
	llmClientAvailable := a.GetLLMClient() != nil && a.GetLLMClient().Client() != nil // Check status *after* init attempt
	debugLog.Printf("Run: initLLMClient finished. Error: %v, Client Available: %v", initErr, llmClientAvailable)

	// Check if init error is fatal *for the selected mode*
	fatalInitErr := false
	if initErr != nil {
		errLog.Printf("LLM Client initialization failed: %v", initErr)
		// Determine if the selected mode strictly requires the LLM client
		strictNeed := false
		scriptNeed := false
		if a.Config != nil {
			strictNeed = a.Config.RunTuiMode || a.Config.RunAgentMode || a.Config.RunSyncMode || a.Config.CleanAPI
			scriptNeed = a.Config.ScriptFile != "" && a.Config.EnableLLM
		}
		if strictNeed || scriptNeed {
			fatalInitErr = true // If mode needs LLM, init error is fatal
		} else if warnLog != nil {
			// Otherwise, just warn
			warnLog.Printf("Continuing despite LLM init error as it might not be required for this mode: %v", initErr)
		}
	}
	// If initialization failed and was required, return error now
	if fatalInitErr {
		return fmt.Errorf("LLM client initialization failed: %w", initErr)
	}

	// 3. Select Mode Based on Config
	infoLog.Println("Run: Dispatching based on mode...")

	// Ensure config is not nil before accessing flags
	if a.Config == nil {
		errMsg := "internal error: App.Config is nil before mode dispatch"
		errLog.Println(errMsg)
		return errors.New(errMsg)
	}

	// Mode precedence: clean-api > sync > script > tui > agent (default handled in Config)
	if a.Config.CleanAPI {
		infoLog.Println("--- Running in Clean API Mode ---")
		if !llmClientAvailable { // Check status after init attempt
			err := errors.New("LLM Client required for clean-api mode but is unavailable")
			errLog.Println(err.Error())
			return err
		}
		return a.runCleanAPIMode(ctx) // Assumes exists in app_helpers.go
	} else if a.Config.RunSyncMode {
		infoLog.Println("--- Running in Sync Mode ---")
		if !llmClientAvailable { // Check status after init attempt
			err := errors.New("LLM Client required for sync mode but is unavailable")
			errLog.Println(err.Error())
			return err
		}
		return a.runSyncMode(ctx) // Assumes exists in app_sync.go
	} else if a.Config.ScriptFile != "" {
		infoLog.Println("--- Running in Script Mode ---")
		// Script mode might proceed even if llmClient is nil
		return a.runScriptMode(ctx) // Assumes exists in app_script.go
	} else if a.Config.RunTuiMode {
		infoLog.Println("--- Running in TUI Mode ---")
		// TUI mode initialization succeeded if we got here (fatalInitErr was false)
		// The TUI /sync command itself needs the client, which it checks via the interface.
		return a.runTuiMode(ctx) // Assumes exists in app_tui.go
	} else if a.Config.RunAgentMode { // Check if this should be the default
		infoLog.Println("--- Running in Agent Mode ---")
		if !a.Config.EnableLLM {
			err := errors.New("agent mode requires LLM but it was disabled with -enable-llm=false")
			errLog.Println(err.Error())
			return err
		}
		if !llmClientAvailable { // Check status after init attempt
			err := errors.New("LLM Client required for agent mode but is unavailable")
			errLog.Println(err.Error())
			return err
		}
		return a.runAgentMode(ctx) // Assumes exists in app_agent.go
	} else {
		// This case should only be reached if config parsing doesn't set a default mode
		errMsg := "internal error: no execution mode specified or determined by flags"
		errLog.Println(errMsg)
		return errors.New(errMsg)
	}
}

// NOTE: Ensure methods like runCleanAPIMode, runSyncMode, etc. are defined
// either here (if not moved) or in their respective app_*.go files.
// Remove any duplicate definitions from this file if they exist in app_helpers.go.
