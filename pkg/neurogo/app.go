// filename: pkg/neurogo/app.go
package neurogo

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// App encapsulates the NeuroScript application logic.
type App struct {
	Config    Config // Configuration loaded from flags
	InfoLog   *log.Logger
	DebugLog  *log.Logger
	ErrorLog  *log.Logger
	LLMLog    *log.Logger
	llmClient *core.LLMClient
}

// NewApp creates a new application instance.
func NewApp() *App {
	return &App{
		InfoLog:  log.New(io.Discard, "", 0),
		DebugLog: log.New(io.Discard, "", 0),
		ErrorLog: log.New(io.Discard, "", 0),
		LLMLog:   log.New(io.Discard, "", 0),
	}
}

// initLoggers sets up loggers based on config.
func (a *App) initLoggers() {
	a.InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	a.ErrorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugOutput := io.Discard
	if a.Config.DebugAST || a.Config.DebugInterpreter {
		debugOutput = os.Stderr
		fmt.Fprintf(os.Stderr, "--- Interpreter/AST Debug Logging Enabled ---\n")
	}
	a.DebugLog = log.New(debugOutput, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	llmDebugOutput := io.Discard
	if a.Config.DebugLLM {
		llmDebugOutput = os.Stderr
		fmt.Fprintf(os.Stderr, "--- LLM Debug Logging Enabled ---\n")
	}
	a.LLMLog = log.New(llmDebugOutput, "DEBUG-LLM: ", log.Ldate|log.Ltime|log.Lshortfile)
	a.DebugLog.Println("DebugLog initialized.")
	a.LLMLog.Println("LLMLog initialized.")
}

// Run determines the mode based on parsed flags and executes.
func (a *App) Run(ctx context.Context) error {
	a.initLoggers() // Initialize loggers first

	// Initialize LLM Client if needed (Agent or Sync mode)
	if a.Config.AgentMode || a.Config.SyncMode {
		a.InfoLog.Println("Initializing LLM Client for Agent/Sync mode...")
		a.llmClient = core.NewLLMClient(
			a.Config.APIKey,
			a.Config.ModelName,
			a.LLMLog,
			a.Config.DebugLLM,
		)
		// Check if client initialization failed (NewLLMClient handles empty key logging)
		if a.llmClient == nil || a.llmClient.Client() == nil {
			// NewLLMClient logs the error, just return a generic failure here
			return fmt.Errorf("failed to initialize LLM Client, check API key and logs")
		}
	} else {
		a.InfoLog.Println("LLM Client not needed for Script mode.")
		// Even if not needed, pass a minimal client to NewInterpreter
		// Or NewInterpreter could handle a nil client? Let's pass minimal.
		a.llmClient = core.NewLLMClient("", "", a.LLMLog, false) // Minimal client for script mode interpreter
	}

	// --- Select Mode ---
	if a.Config.SyncMode {
		a.InfoLog.Printf("--- Running in Sync Mode ---")
		return a.runSyncMode(ctx) // Call new sync mode function
	} else if a.Config.AgentMode {
		a.InfoLog.Printf("--- Running in Agent Mode ---")
		return a.runAgentMode(ctx)
	} else {
		a.InfoLog.Printf("--- Running in Script Mode ---")
		return a.runScriptMode(ctx)
	}
}
