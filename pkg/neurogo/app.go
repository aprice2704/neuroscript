// filename: pkg/neurogo/app.go
package neurogo

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aprice2704/neuroscript/pkg/core" // Import core
)

// App encapsulates the NeuroScript application logic (interpreter runner or agent).
type App struct {
	Config   Config // Configuration loaded from flags
	InfoLog  *log.Logger
	DebugLog *log.Logger
	ErrorLog *log.Logger
	LLMLog   *log.Logger
	// +++ ADDED: Field for shared LLM client +++
	llmClient *core.LLMClient
}

// NewApp creates a new application instance with default logging to discard.
// Loggers are properly initialized after flags are parsed via initLoggers.
func NewApp() *App {
	return &App{
		InfoLog:  log.New(io.Discard, "", 0),
		DebugLog: log.New(io.Discard, "", 0),
		ErrorLog: log.New(io.Discard, "", 0),
		LLMLog:   log.New(io.Discard, "", 0),
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
		fmt.Fprintf(os.Stderr, "--- Interpreter/AST Debug Logging Enabled ---\n")
	}
	a.DebugLog = log.New(debugOutput, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialize LLMLog based on -debug-llm flag
	llmDebugOutput := io.Discard
	if a.Config.DebugLLM {
		llmDebugOutput = os.Stderr
		fmt.Fprintf(os.Stderr, "--- LLM Debug Logging Enabled ---\n")
	}
	a.LLMLog = log.New(llmDebugOutput, "DEBUG-LLM: ", log.Ldate|log.Ltime|log.Lshortfile)

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

	// +++ ADDED: Initialize the shared LLMClient +++
	// Do this after loggers are set up, as NewLLMClient uses a.LLMLog
	a.llmClient = core.NewLLMClient(
		a.Config.APIKey,    // Pass API key from config (or it checks env var)
		a.Config.ModelName, // Pass model name from config
		a.LLMLog,           // Pass the dedicated LLM logger
		a.Config.DebugLLM,  // Pass the debug flag from config
	)
	// Check if client initialization failed critically (e.g., invalid API key structure, though NewLLMClient handles empty key)
	// Note: NewLLMClient logs errors internally but returns a (potentially non-functional) client.
	// A check like if a.llmClient.Client() == nil might be useful if needed later.
	// --- END ADDED ---

	if a.Config.AgentMode {
		return a.runAgentMode(ctx)
	}
	return a.runScriptMode(ctx)
}
