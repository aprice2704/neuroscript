// filename: pkg/neurogo/app.go
package neurogo

import (
	"context"
	"fmt" // Keep fmt for initLoggers
	"io"
	"log"
	"os"
	// Don't need other imports here directly
)

// App encapsulates the NeuroScript application logic (interpreter runner or agent).
type App struct {
	Config   Config // Configuration loaded from flags
	InfoLog  *log.Logger
	DebugLog *log.Logger
	ErrorLog *log.Logger
}

// NewApp creates a new application instance with default logging to discard.
// Loggers are properly initialized after flags are parsed via initLoggers.
func NewApp() *App {
	return &App{
		InfoLog:  log.New(io.Discard, "", 0),
		DebugLog: log.New(io.Discard, "", 0),
		ErrorLog: log.New(io.Discard, "", 0),
	}
}

// initLoggers sets up the App's loggers based on config.
// Should be called after ParseFlags.
func (a *App) initLoggers() {
	enableDebug := a.Config.DebugAST || a.Config.DebugInterpreter
	a.InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	a.ErrorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugOutput := io.Discard
	if enableDebug {
		debugOutput = os.Stderr
		fmt.Fprintf(os.Stderr, "--- Debug Logging Enabled ---\n") // Use Fprintf
	}
	a.DebugLog = log.New(debugOutput, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Run determines the mode based on parsed flags and executes the appropriate logic.
func (a *App) Run(ctx context.Context) error {
	// Initialize loggers *after* flags are parsed but *before* running modes
	a.initLoggers()

	if a.Config.AgentMode {
		// Calls the method defined in app_agent.go
		return a.runAgentMode(ctx)
	}
	// Calls the method defined in app_script.go
	return a.runScriptMode(ctx)
}
