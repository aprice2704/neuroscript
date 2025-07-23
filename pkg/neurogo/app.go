// NeuroScript Version: 0.3.1
// File version: 0.2.5
// Purpose: Updated to use the new WithSandboxDir option for interpreter creation.
// filename: pkg/neurogo/app.go
package neurogo

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llm"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// App orchestrates the main application logic.
type App struct {
	Config       *Config
	Log          interfaces.Logger
	interpreter  *interpreter.Interpreter
	llmClient    interfaces.LLMClient
	agentCtx     *AgentContext
	patchHandler *PatchHandler
	mu           sync.RWMutex
	appCtx       context.Context
	cancelFunc   context.CancelFunc

	tui            *tviewAppPointers
	originalStdout io.Writer

	chatSessions        map[string]*ChatSession
	activeChatSessionID string
	nextChatIDSuffix    int
}

// GetInterpreter allows safe access to the interpreter.
func (a *App) GetInterpreter() *interpreter.Interpreter {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.interpreter
}

// SetInterpreter allows setting the interpreter after App creation.
func (a *App) SetInterpreter(interp *interpreter.Interpreter) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.interpreter = interp
}

// SetStdout sets the standard output writer for the underlying interpreter.
func (a *App) SetStdout(writer io.Writer) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.interpreter != nil {
		a.interpreter.SetStdout(writer)
	} else {
		a.Log.Warn("Attempted to set stdout, but interpreter is nil.")
	}
}

// Stdout gets the standard output writer from the underlying interpreter.
func (a *App) Stdout() io.Writer {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.interpreter != nil {
		return a.interpreter.Stdout()
	}
	a.Log.Warn("Attempted to get stdout, but interpreter is nil. Returning os.Stdout.")
	return os.Stdout
}

// SetStderr sets the standard error writer for the underlying interpreter.
func (a *App) SetStderr(writer io.Writer) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.interpreter != nil {
		a.interpreter.SetStderr(writer)
	} else {
		a.Log.Warn("Attempted to set stderr, but interpreter is nil.")
	}
}

// Run starts the application based on the configured mode.
func (a *App) Run() error {
	a.Log.Info("NeuroScript application starting...", "version", "0.4.0")

	if err := a.InitializeCoreComponents(); err != nil {
		a.Log.Error("Failed to initialize core components", "error", err)
		return fmt.Errorf("core component initialization failed: %w", err)
	}

	if a.Config.SyncDir != "" {
		a.Log.Info("Sync mode selected.", "sync_dir", a.Config.SyncDir)
		return a.runSyncMode(a.appCtx)
	}
	if a.Config.StartupScript != "" && !a.Config.TuiMode {
		a.Log.Info("Script execution mode selected.", "script", a.Config.StartupScript)

		filepathArg, err := lang.Wrap(a.Config.StartupScript)
		if err != nil {
			return fmt.Errorf("internal error wrapping startup script path: %w", err)
		}
		toolArgs := map[string]lang.Value{"filepath": filepathArg}
		contentValue, err := a.interpreter.ExecuteTool("FS.Read", toolArgs)
		if err != nil {
			a.Log.Error("Failed to read startup script", "script", a.Config.StartupScript, "error", err)
			return fmt.Errorf("failed to read startup script '%s': %w", a.Config.StartupScript, err)
		}
		scriptContent, ok := lang.Unwrap(contentValue).(string)
		if !ok {
			return fmt.Errorf("internal error: FS.Read did not return a string for startup script")
		}

		if _, loadErr := a.LoadScriptString(a.appCtx, scriptContent); loadErr != nil {
			a.Log.Error("Failed to load startup script", "script", a.Config.StartupScript, "error", loadErr)
			return fmt.Errorf("failed to load startup script '%s': %w", a.Config.StartupScript, loadErr)
		}
		a.Log.Info("Startup script loaded successfully", "script", a.Config.StartupScript)

		if a.Config.TargetArg != "" {
			a.Log.Info("Executing entrypoint from config.", "procedure", a.Config.TargetArg, "args", a.Config.ProcArgs)
			_, runErr := a.RunProcedure(a.appCtx, a.Config.TargetArg, a.Config.ProcArgs)
			if runErr != nil {
				a.Log.Error("Entrypoint execution failed", "procedure", a.Config.TargetArg, "error", runErr)
				return fmt.Errorf("entrypoint '%s' failed: %w", a.Config.TargetArg, runErr)
			}
			a.Log.Info("Entrypoint executed successfully.", "procedure", a.Config.TargetArg)
		}

		return nil
	}

	a.Log.Info("Defaulting to TUI mode.")
	return a.runTuiMode(a.appCtx)
}

// runTuiMode starts the application in TUI mode.
func (a *App) runTuiMode(ctx context.Context) error {
	a.Log.Info("Starting TUI mode (tview)...")

	initialTuiScript := ""
	if a.Config.StartupScript != "" && a.Config.TuiMode {
		initialTuiScript = a.Config.StartupScript
		a.Log.Info("TUI mode will execute initial script", "script", initialTuiScript)
	}

	a.originalStdout = a.Stdout()
	if a.originalStdout == nil {
		a.Log.Warn("Original stdout is nil before starting TUI mode.")
	}

	err := StartTviewTUI(a, initialTuiScript)

	if a.originalStdout != nil {
		a.SetStdout(a.originalStdout)
		a.Log.Info("Restored original stdout after TUI exit.")
	}

	if err != nil {
		a.Log.Error("TUI mode exited with error", "error", err)
		return err
	}
	a.Log.Info("TUI mode exited gracefully.")
	return nil
}

func (a *App) GetLogger() interfaces.Logger {
	if a.Log == nil {
		return logging.NewNoOpLogger()
	}
	return a.Log
}

func (a *App) GetLLMClient() interfaces.LLMClient {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.llmClient
}

func (a *App) Context() context.Context {
	return a.appCtx
}

func (a *App) InitializeCoreComponents() error {
	a.Log.Debug("Initializing core components...")
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.llmClient == nil {
		if a.Config != nil {
			createdClient, err := a.CreateLLMClient()
			if err != nil {
				a.Log.Error("Failed to create LLM client during core component init", "error", err)
				a.llmClient = adapters.NewNoOpLLMClient()
				a.Log.Info("Using default NoOpLLMClient after creation failure in core component init.")
			} else {
				a.llmClient = createdClient
				a.Log.Info("LLM Client created successfully during core component init.")
			}
		} else {
			a.llmClient = adapters.NewNoOpLLMClient()
			a.Log.Info("Using default NoOpLLMClient during core component init (config was nil).")
		}
	}

	sandboxDir := a.Config.SandboxDir
	if sandboxDir == "" {
		homeDir, errHome := os.UserHomeDir()
		if errHome != nil {
			var err error
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
		if errMkdir := os.MkdirAll(sandboxDir, 0755); errMkdir != nil {
			return fmt.Errorf("failed to create specified sandbox directory %s: %w", sandboxDir, errMkdir)
		}
		a.Log.Info("Created sandbox directory", "path", sandboxDir)
	} else {
		a.Log.Info("Using existing sandbox directory", "path", sandboxDir)
	}

	interpLLMClient := a.llmClient
	if interpLLMClient == nil {
		a.Log.Warn("LLMClient is nil for interpreter, using NoOpLLMClient for interpreter.")
		interpLLMClient = adapters.NewNoOpLLMClient()
	}

	// Correctly pass the sandbox directory as an option during creation.
	a.interpreter = interpreter.NewInterpreter(
		interpreter.WithLogger(a.Log),
		interpreter.WithLLMClient(interpLLMClient),
		interpreter.WithSandboxDir(sandboxDir),
	)

	if errRegister := tool.RegisterCoreTools(a.interpreter.ToolRegistry()); errRegister != nil {
		a.Log.Warn("Failed to register all core tools", "error", errRegister)
	}

	a.agentCtx = NewAgentContext(a.Log)
	if a.agentCtx != nil {
		a.agentCtx.SetSandboxDir(sandboxDir)
	}

	a.chatSessions = make(map[string]*ChatSession)

	a.Log.Info("Core components initialization attempt finished.")
	return nil
}

func (a *App) HandleSystemCommand(command string) {
	if a.Log == nil {
		fmt.Printf("App.Log is nil, cannot log system command: %s\n", command)
		return
	}
	a.Log.Info("System command received by App (from TUI)", "command", command)
}

func (a *App) ExecuteScriptLine(ctx context.Context, line string) {
	if a.Log == nil {
		fmt.Printf("App.Log is nil, cannot log script line: %s\n", line)
		return
	}
	a.Log.Info("Script line received by App (from TUI)", "line", line)
}

func (app *App) CreateLLMClient() (interfaces.LLMClient, error) {
	if app.Config == nil {
		return nil, fmt.Errorf("cannot create LLM client: app config is nil")
	}

	apiKey := app.Config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("NEUROSCRIPT_API_KEY")
		if apiKey == "" {
			if app.Log != nil {
				app.Log.Debug("API key is missing in config and environment variable (NEUROSCRIPT_API_KEY). Creating NoOpLLMClient.")
			}
			return adapters.NewNoOpLLMClient(), nil
		}
		if app.Log != nil {
			app.Log.Debug("Using LLM API key from environment variable NEUROSCRIPT_API_KEY.")
		}
	} else {
		if app.Log != nil {
			app.Log.Debug("Using LLM API key from configuration.")
		}
	}

	if app.Log != nil {
		app.Log.Debug("Creating real LLMClient.")
	}
	apiHost := app.Config.APIHost
	modelName := app.Config.ModelName

	loggerToUse := app.Log
	if loggerToUse == nil {
		loggerToUse = logging.NewNoOpLogger()
	}

	llmClient, err := llm.NewLLMClient(apiKey, modelName, loggerToUse)
	if err != nil || llmClient == nil {
		if app.Log != nil {
			app.Log.Error("NewLLMClient returned an error or is nil.", "error", err)
		}
		return adapters.NewNoOpLLMClient(), fmt.Errorf("NewLLMClient failed for real client creation")
	}
	if app.Log != nil {
		app.Log.Debug("Real LLMClient created.", "host", apiHost, "model", modelName)
	}
	return llmClient, nil
}
