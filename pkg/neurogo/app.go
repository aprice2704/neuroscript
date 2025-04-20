// filename: pkg/neurogo/app.go
package neurogo

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

// App encapsulates the NeuroScript application logic (interpreter runner or agent).
type App struct {
	Config   Config // Configuration loaded from flags
	InfoLog  *log.Logger
	DebugLog *log.Logger
	ErrorLog *log.Logger
	// +++ ADDED: Dedicated logger for LLM interactions +++
	LLMLog *log.Logger
}

// NewApp creates a new application instance with default logging to discard.
// Loggers are properly initialized after flags are parsed via initLoggers.
func NewApp() *App {
	// Initialize Config with zero values or defaults if necessary
	// cfg := Config{ ... default values ... }

	return &App{
		// Config: cfg, // Initialize Config if needed
		InfoLog:  log.New(io.Discard, "", 0),
		DebugLog: log.New(io.Discard, "", 0),
		ErrorLog: log.New(io.Discard, "", 0),
		// +++ ADDED: Initialize LLMLog to discard by default +++
		LLMLog: log.New(io.Discard, "", 0),
	}
}

// initLoggers sets up the App's loggers based on config.
// Should be called after ParseFlags.
func (a *App) initLoggers() {
	// Standard loggers
	a.InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	a.ErrorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// DebugLog enabled by -debug-ast OR -debug-interpreter
	debugOutput := io.Discard
	enableDebug := a.Config.DebugAST || a.Config.DebugInterpreter
	if enableDebug {
		debugOutput = os.Stderr
		// Use Fprintf to avoid logger prefix for this initial message
		fmt.Fprintf(os.Stderr, "--- Interpreter/AST Debug Logging Enabled ---\n")
	}
	a.DebugLog = log.New(debugOutput, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	// +++ ADDED: Initialize LLMLog based on -debug-llm flag +++
	llmDebugOutput := io.Discard
	if a.Config.DebugLLM {
		llmDebugOutput = os.Stderr // LLM debug usually goes to stderr
		fmt.Fprintf(os.Stderr, "--- LLM Debug Logging Enabled ---\n")
	}
	// Use a distinct prefix for LLM debug messages
	a.LLMLog = log.New(llmDebugOutput, "DEBUG-LLM: ", log.Ldate|log.Ltime|log.Lshortfile)
	// --- End Added Code ---

	// Optionally log the final logger states
	a.DebugLog.Println("DebugLog initialized.")
	a.LLMLog.Println("LLMLog initialized.") // This only prints if -debug-llm is set
}

// Run determines the mode based on parsed flags and executes the appropriate logic.
func (a *App) Run(ctx context.Context) error {
	// Initialize loggers *after* flags are parsed but *before* running modes
	a.initLoggers()

	// Log configuration details (optional, consider privacy of API key)
	a.InfoLog.Printf("Mode: %s", map[bool]string{true: "Agent", false: "Script"}[a.Config.AgentMode])
	a.InfoLog.Printf("Model: %s", a.Config.ModelName)
	// Be careful logging sensitive info like API keys
	// a.DebugLog.Printf("Config: %+v", a.Config)

	if a.Config.AgentMode {
		return a.runAgentMode(ctx)
	}
	return a.runScriptMode(ctx)
}
