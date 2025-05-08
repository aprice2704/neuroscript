// NeuroScript Version: 0.3.0
// File version: 0.1.10 // Correct NewApp, RegisterAIWorkerTools calls, procArgsConfig type. Remove unused llmClient var.
// filename: cmd/ng/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/aprice2704/neuroscript/pkg/core"    // Import logging for type usage
	"github.com/aprice2704/neuroscript/pkg/neurogo" // neurogo package for App, Config, constants
	"github.com/aprice2704/neuroscript/pkg/neurogo/tui"
	// "github.com/aprice2704/neuroscript/pkg/toolsets" // Not used currently, but might be for RegisterExtendedTools
)

func main() {
	// --- Configuration Setup (Flag Definitions) ---
	logFile := flag.String("log-file", "", "Path to log file (optional, defaults to stderr)")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	sandboxDir := flag.String("sandbox", ".", "Root directory for secure file operations and ai_wm persistence")

	apiKey := flag.String("api-key", os.Getenv("GEMINI_API_KEY"), "Gemini API Key (env: GEMINI_API_KEY or NEUROSCRIPT_API_KEY)")
	if *apiKey == "" {
		*apiKey = os.Getenv("NEUROSCRIPT_API_KEY") // Fallback to NEUROSCRIPT_API_KEY
	}
	apiHost := flag.String("api-host", "", "Optional API Host/Endpoint override")
	insecure := flag.Bool("insecure", false, "Disable security checks (Use with extreme caution!)") // Used for Config.Insecure
	modelName := flag.String("model", neurogo.DefaultModelName, "Default generative model name for LLM interactions")
	startupScript := flag.String("script", "", "Path to a NeuroScript (.ns) file to execute on startup")
	tuiMode := flag.Bool("tui", true, "Enable Terminal User Interface (TUI) mode") // Default to TUI true
	replMode := flag.Bool("repl", false, "Enable basic REPL mode (if TUI is false and no script is run)")

	libPathsConfig := neurogo.NewStringSliceFlag()
	flag.Var(libPathsConfig, "lib-path", "Path to a NeuroScript library directory (can be specified multiple times)")

	aiServiceAllowCfg := neurogo.NewStringSliceFlag()
	flag.Var(aiServiceAllowCfg, "ai-allow", "Tool/service name to allow for AI (can be specified multiple times, e.g., 'FileSystem.ReadFile')")

	aiServiceDenyCfg := neurogo.NewStringSliceFlag()
	flag.Var(aiServiceDenyCfg, "ai-deny", "Tool/service name to deny for AI (can be specified multiple times, overrides allows)")

	targetArg := flag.String("target", "main", "Target procedure for the script")
	// CORRECTED: Use NewStringSliceFlag for procArgsConfig as well
	procArgsConfig := neurogo.NewStringSliceFlag()
	flag.Var(procArgsConfig, "arg", "Argument for the script process/procedure (can be specified multiple times)")

	flag.Parse()

	// --- Logger Initialization ---
	logger, err := initializeLogger(*logLevel, *logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	logger.Info("Logger initialized", "level", *logLevel, "file", *logFile)

	// --- Application Context ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Trap SIGINT and SIGTERM to trigger context cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Received signal, shutting down...", "signal", sig.String())
		cancel()
	}()

	// --- Sandbox Directory Resolution ---
	absSandboxDir, err := filepath.Abs(*sandboxDir)
	if err != nil {
		logger.Error("Failed to resolve absolute path for sandbox directory", "path", *sandboxDir, "error", err)
		fmt.Fprintf(os.Stderr, "Error resolving sandbox directory '%s': %v\n", *sandboxDir, err)
		os.Exit(1)
	}
	logger.Info("Sandbox directory resolved", "path", absSandboxDir)

	// --- NeuroGo App Configuration & Initialization ---
	appConfig := &neurogo.Config{
		APIKey:        *apiKey,
		APIHost:       *apiHost,
		ModelName:     *modelName,
		StartupScript: *startupScript,
		SandboxDir:    absSandboxDir,
		Insecure:      *insecure,
		LibPaths:      libPathsConfig.Value,
		TargetArg:     *targetArg,
		ProcArgs:      procArgsConfig.Value, // Access the .Value field
	}

	// CORRECTED: NewApp only takes the logger now.
	app := neurogo.NewApp(logger)
	app.Config = appConfig // Assign config after creation

	// --- Initialize Core Components (LLM, Interpreter, AIWM) ---
	// REMOVED: Unused declaration: var llmClient core.LLMClient
	var interpreter *core.Interpreter
	var aiWm *core.AIWorkerManager

	// Pass app (which contains config) and logger to initializeCoreComponents
	// initializeCoreComponents now returns the components it creates/initializes.
	// We capture them here, although they are also set within the app instance by initializeCoreComponents.
	_, interpreter, aiWm, err = initializeCoreComponents(app, logger, absSandboxDir) // Capture interpreter and aiWm
	if err != nil {
		logger.Error("Failed to initialize core components", "error", err)
		fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Core components (LLM, Interpreter, AIWM) initialized successfully.")

	// --- Register Tools ---
	// Core tools are registered within NewInterpreter.
	// AI Worker tools are registered if AIWM is available.
	if aiWm != nil {
		// CORRECTED: RegisterAIWorkerTools only takes the interpreter.
		if err := core.RegisterAIWorkerTools(interpreter); err != nil {
			logger.Error("Failed to register AI Worker tools", "error", err)
			fmt.Fprintf(os.Stderr, "Warning: Failed to register AI Worker tools: %v\n", err)
		} else {
			logger.Info("AI Worker tools registered.")
		}
	} else {
		logger.Warn("AI Worker Manager not initialized, skipping AI Worker tool registration.")
	}

	// --- Execute Startup Script (if provided) ---
	scriptExecuted := false
	if app.Config.StartupScript != "" {
		logger.Info("Executing startup script", "script", app.Config.StartupScript)
		if err := app.ExecuteScriptFile(ctx, app.Config.StartupScript); err != nil {
			logger.Error("Error executing startup script", "script", app.Config.StartupScript, "error", err)
			fmt.Fprintf(os.Stderr, "Error executing startup script '%s': %v\n", app.Config.StartupScript, err)
			os.Exit(1)
		} else {
			logger.Info("Startup script finished successfully.")
			scriptExecuted = true
		}
	} else {
		logger.Info("No startup script specified.")
	}

	// --- Start Primary Interaction (TUI or REPL or Exit) ---
	if *tuiMode {
		logger.Info("Starting interactive TUI...")
		if err := tui.Start(app); err != nil {
			logger.Error("TUI Error", "error", err)
			fmt.Fprintf(os.Stderr, "TUI Error: %v\n", err)
			os.Exit(1)
		}
		logger.Info("TUI finished.")
	} else if scriptExecuted {
		logger.Info("Startup script executed and no interactive UI specified. Exiting.")
	} else if *replMode {
		logger.Info("TUI disabled, no script run. Starting basic REPL...")
		runRepl(ctx, app)
		logger.Info("Basic REPL finished.")
	} else {
		logger.Info("No TUI, no script, no REPL. Nothing to do. Exiting.")
		if flag.NArg() > 0 {
			scriptPath := flag.Arg(0)
			logger.Info("Found positional argument, attempting to execute as script", "script", scriptPath)
			app.Config.StartupScript = scriptPath
			if err := app.ExecuteScriptFile(ctx, scriptPath); err != nil {
				logger.Error("Error executing script from positional argument", "script", scriptPath, "error", err)
				fmt.Fprintf(os.Stderr, "Error executing script '%s': %v\n", scriptPath, err)
				os.Exit(1)
			}
			logger.Info("Script from positional argument finished successfully.")
		} else {
			fmt.Println("No action specified. Use -script <file>, -tui, or provide a script file as an argument.")
			flag.Usage()
		}
	}

	logger.Info("NeuroScript application finished.")
	if closer, ok := logger.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing logger: %v\n", err)
		}
	}
}
