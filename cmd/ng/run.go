// NeuroScript Version: 0.5.0
// File version: 6
// Purpose: Corrects the SetLevel method signature to match the interfaces.Logger interface.
// filename: cmd/ng/run.go
// nlines: 290
// risk_rating: HIGH
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/neurogo"
	"golang.org/x/exp/slog"
)

// SlogAdapter wraps an slog.Logger to satisfy the internal interfaces.Logger interface
// by adding the required formatted logging methods and dynamic level setting.
type SlogAdapter struct {
	*slog.Logger
	level *slog.LevelVar
}

// SetLevel changes the logging level dynamically. It accepts the interface's
// LogLevel type and converts it to the concrete slog.Level.
func (a *SlogAdapter) SetLevel(level interfaces.LogLevel) {
	a.level.Set(slog.Level(level))
}

// Debugf implements the formatted debug log.
func (a *SlogAdapter) Debugf(format string, args ...any) {
	a.Debug(fmt.Sprintf(format, args...))
}

// Infof implements the formatted info log.
func (a *SlogAdapter) Infof(format string, args ...any) {
	a.Info(fmt.Sprintf(format, args...))
}

// Warnf implements the formatted warn log.
func (a *SlogAdapter) Warnf(format string, args ...any) {
	a.Warn(fmt.Sprintf(format, args...))
}

// Errorf implements the formatted error log.
func (a *SlogAdapter) Errorf(format string, args ...any) {
	a.Error(fmt.Sprintf(format, args...))
}

// CliConfig holds all configuration passed from the command line flags.
type CliConfig struct {
	LogFile          string
	LogLevel         string
	SandboxDir       string
	APIKey           string
	APIHost          string
	Insecure         bool
	ModelName        string
	StartupScript    string
	TuiMode          bool
	ReplMode         bool
	LibPaths         []string
	TargetArg        string
	ProcArgs         []string
	PositionalScript string
}

// Run executes the main application logic based on the provided configuration and returns an exit code.
func Run(cfg CliConfig) int {
	// --- Logger Initialization ---
	logger, closer, err := InitializeLogger(cfg.LogLevel, cfg.LogFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		return 1
	}
	if closer != nil {
		defer func() {
			if err := closer(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing logger: %v\n", err)
			}
		}()
	}
	logger.Debug("Logger initialized", "level", cfg.LogLevel, "file", cfg.LogFile)

	// --- Application Context ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Debugf("Received signal, shutting down... (signal: %s)", sig.String())
		cancel()
	}()

	// --- Sandbox Directory Resolution ---
	absSandboxDir, err := filepath.Abs(cfg.SandboxDir)
	if err != nil {
		logger.Errorf("Failed to resolve absolute path for sandbox directory '%s': %v", cfg.SandboxDir, err)
		fmt.Fprintf(os.Stderr, "Error resolving sandbox directory '%s': %v\n", cfg.SandboxDir, err)
		return 1
	}
	logger.Debugf("Sandbox directory resolved to '%s'", absSandboxDir)

	// --- NeuroGo App Configuration & LLM Client Creation FIRST ---
	appConfig := &neurogo.Config{
		APIKey:        cfg.APIKey,
		APIHost:       cfg.APIHost,
		ModelName:     cfg.ModelName,
		StartupScript: cfg.StartupScript,
		SandboxDir:    absSandboxDir,
		Insecure:      cfg.Insecure,
		LibPaths:      cfg.LibPaths,
		TargetArg:     cfg.TargetArg,
		ProcArgs:      cfg.ProcArgs,
	}

	logger.Debug("Preparing to create LLM client...")
	tempAppForLLMCreation := &neurogo.App{Config: appConfig, Log: logger}
	llmClient, err := tempAppForLLMCreation.CreateLLMClient()
	if err != nil {
		logger.Errorf("Failed to create LLM client: %v", err)
		fmt.Fprintf(os.Stderr, "LLM client creation error: %v\n", err)
		return 1
	}
	if llmClient == nil {
		logger.Errorf("LLM client is nil after creation without error.")
		fmt.Fprintf(os.Stderr, "LLM client is nil after creation.\n")
		return 1
	}
	logger.Debug("LLM client created successfully.")

	// --- NeuroGo App Initialization (with LLM Client) ---
	app, err := neurogo.NewApp(appConfig, logger, llmClient)
	if err != nil {
		logger.Errorf("Failed to create NeuroGo App: %v", err)
		fmt.Fprintf(os.Stderr, "NeuroGo App creation error: %v\n", err)
		return 1
	}
	logger.Debug("NeuroGo App instance created.")

	// --- Initialize Core Components (Interpreter) ---
	if err := app.InitializeCoreComponents(); err != nil {
		logger.Errorf("Failed to initialize core components: %v", err)
		fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
		return 1
	}
	logger.Debugf("---> STARTING NG with %d tools <---\n", app.GetInterpreter().NTools())

	// --- Determine Mode of Operation ---
	scriptToRunNonTUI := cfg.StartupScript
	if scriptToRunNonTUI == "" && cfg.PositionalScript != "" && !cfg.TuiMode {
		scriptToRunNonTUI = cfg.PositionalScript
		logger.Debugf("Using positional argument as script for non-TUI execution: %s", scriptToRunNonTUI)
	}

	if cfg.TuiMode {
		logger.Debug("TUI mode requested. Starting TUI...")
		if err := neurogo.StartTviewTUI(app, cfg.StartupScript); err != nil {
			logger.Errorf("TUI Error: %v", err)
			fmt.Fprintf(os.Stderr, "TUI Error: %v\n", err)
			return 1
		}
		logger.Debug("TUI finished.")
	} else {
		// --- Non-TUI Mode ---
		scriptExecutedInNonTUI := false
		if scriptToRunNonTUI != "" {
			logger.Debugf("Executing script (non-TUI mode): %s", scriptToRunNonTUI)
			if err := executeScript(ctx, app, scriptToRunNonTUI); err != nil {
				logger.Errorf("Script execution failed for '%s': %v", scriptToRunNonTUI, err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return 1
			}
			logger.Debug("Script finished successfully (non-TUI mode).")
			scriptExecutedInNonTUI = true
		}

		if !scriptExecutedInNonTUI {
			if cfg.ReplMode {
				logger.Debug("No script run or TUI disabled. Starting basic REPL...")
				RunRepl(ctx, app)
				logger.Debug("Basic REPL finished.")
			} else if cfg.StartupScript == "" && cfg.PositionalScript == "" {
				logger.Debug("No TUI, no script, no REPL. Nothing to do. Exiting.")
				fmt.Println("No action specified. Use -script <file>, -tui, or -repl, or provide a script file as an argument.")
			}
		}
	}

	logger.Debug("NeuroScript application finished.")
	return 0
}

func executeScript(ctx context.Context, app *neurogo.App, scriptPath string) error {
	interpreter := app.GetInterpreter()
	filepathArg, wrapErr := lang.Wrap(scriptPath)
	if wrapErr != nil {
		return fmt.Errorf("internal error wrapping script path: %w", wrapErr)
	}
	toolArgs := map[string]lang.Value{"filepath": filepathArg}
	contentValue, toolErr := interpreter.ExecuteTool("TOOL.ReadFile", toolArgs)
	if toolErr != nil {
		return fmt.Errorf("error reading script '%s': %w", scriptPath, toolErr)
	}
	scriptContent, ok := lang.Unwrap(contentValue).(string)
	if !ok {
		return fmt.Errorf("internal error: TOOL.ReadFile did not return a string")
	}
	if _, loadErr := app.LoadScriptString(ctx, scriptContent); loadErr != nil {
		return fmt.Errorf("error loading script '%s': %w", scriptPath, loadErr)
	}
	if app.Config.TargetArg != "" {
		app.Log.Debugf("Executing entrypoint: %s with args: %v", app.Config.TargetArg, app.Config.ProcArgs)
		if _, runErr := app.RunProcedure(ctx, app.Config.TargetArg, app.Config.ProcArgs); runErr != nil {
			return fmt.Errorf("error executing entrypoint '%s': %w", app.Config.TargetArg, runErr)
		}
	}
	return nil
}

// InitializeLogger sets up the logger based on command-line flags.
// It returns an adapter that satisfies the neurogo Logger interface and a closer function.
func InitializeLogger(level, filePath string) (adapter *SlogAdapter, closer func() error, err error) {
	var levelVar slog.LevelVar
	var initialLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		initialLevel = slog.LevelDebug
	case "info":
		initialLevel = slog.LevelInfo
	case "warn":
		initialLevel = slog.LevelWarn
	case "error":
		initialLevel = slog.LevelError
	default:
		return nil, nil, fmt.Errorf("invalid log level: %s", level)
	}
	levelVar.Set(initialLevel)

	var output io.Writer = os.Stderr
	if filePath != "" {
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file %s: %w", filePath, err)
		}
		output = f
		closer = f.Close
	}

	handler := slog.NewTextHandler(output, &slog.HandlerOptions{Level: &levelVar})
	slogger := slog.New(handler)
	adapter = &SlogAdapter{Logger: slogger, level: &levelVar}

	return adapter, closer, nil
}

// RunRepl starts a basic read-eval-print loop.
func RunRepl(ctx context.Context, app *neurogo.App) {
	// Implementation would go here
	fmt.Println("Basic REPL not yet implemented.")
}
