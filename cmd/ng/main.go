// NeuroScript Version: 0.3.0
// File version: 0.1.15
// filename: cmd/ng/main.go
// nlines: 229
// risk_rating: HIGH
// Changes:
// - Added a --version flag to print application and grammar versions in JSON format.
// - Declared main.AppVersion to be injected via ldflags.
// - Added logic to handle the --version flag immediately after parsing.
// - Added "encoding/json" to imports.
// - Corrected app.InitLLMClient to app.CreateLLMClient.
// - Fixed LLMClient argument for InitializeCoreComponents (using app.InitLLMClient).
// - TUI flag defaults to false.
// - TUI only starts if -tui is explicitly true.
// - If -script and -tui, script path is passed to TUI for delayed execution.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/neurogo"
)

// Version information, injected at build time via -ldflags.
// Example: go build -ldflags="-X main.AppVersion=1.2.3"
var (
	AppVersion string
)

func main() {
	// --- Configuration Setup (Flag Definitions) ---
	versionFlag := flag.Bool("version", false, "Print version information in JSON format and exit")
	logFile := flag.String("log-file", "", "Path to log file (optional, defaults to stderr)")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	sandboxDir := flag.String("sandbox", ".", "Root directory for secure file operations and ai_wm persistence")

	apiKey := flag.String("api-key", os.Getenv("GEMINI_API_KEY"), "Gemini API Key (env: GEMINI_API_KEY or NEUROSCRIPT_API_KEY)")
	if *apiKey == "" {
		*apiKey = os.Getenv("NEUROSCRIPT_API_KEY")
	}
	apiHost := flag.String("api-host", "", "Optional API Host/Endpoint override")
	insecure := flag.Bool("insecure", false, "Disable security checks (Use with extreme caution!)")
	modelName := flag.String("model", neurogo.DefaultModelName, "Default generative model name for LLM interactions")
	startupScriptPath := flag.String("script", "", "Path to a NeuroScript (.ns) file to execute")

	// TUI flag now defaults to false
	tuiMode := flag.Bool("tui", false, "Enable Terminal User Interface (TUI) mode")
	replMode := flag.Bool("repl", false, "Enable basic REPL mode (if TUI is false and no script is run)")

	libPathsConfig := neurogo.NewStringSliceFlag()
	flag.Var(libPathsConfig, "lib-path", "Path to a NeuroScript library directory (can be specified multiple times)")

	aiServiceAllowCfg := neurogo.NewStringSliceFlag()
	flag.Var(aiServiceAllowCfg, "ai-allow", "Tool/service name to allow for AI (can be specified multiple times, e.g., 'FileSystem.ReadFile')")

	aiServiceDenyCfg := neurogo.NewStringSliceFlag()
	flag.Var(aiServiceDenyCfg, "ai-deny", "Tool/service name to deny for AI (can be specified multiple times, overrides allows)")

	targetArg := flag.String("target", "main", "Target procedure for the script")
	procArgsConfig := neurogo.NewStringSliceFlag()

	flag.Var(procArgsConfig, "arg", "Argument for the script process/procedure (can be specified multiple times)")

	flag.Parse()

	// --- Handle Version Flag ---
	// If the --version flag is provided, print version info and exit immediately.
	if *versionFlag {
		// Provide default values if not injected during build.
		appVersion := AppVersion
		if appVersion == "" {
			appVersion = "dev"
		}
		grammarVersion := core.GrammarVersion
		if grammarVersion == "" {
			grammarVersion = "unknown"
		}

		versionInfo := struct {
			AppVersion     string `json:"app_version"`
			GrammarVersion string `json:"grammar_version"`
		}{
			AppVersion:     appVersion,
			GrammarVersion: grammarVersion,
		}

		jsonOutput, err := json.MarshalIndent(versionInfo, "", "  ")
		if err != nil {
			// Logger is not initialized yet, so print directly to stderr.
			fmt.Fprintf(os.Stderr, "Error creating version JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonOutput))
		os.Exit(0)
	}

	// --- Logger Initialization ---
	logger, err := initializeLogger(*logLevel, *logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	logger.Debug("Logger initialized", "level", *logLevel, "file", *logFile)

	// --- Application Context ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Debug("Received signal, shutting down...", "signal", sig.String())
		cancel()
	}()

	// --- Sandbox Directory Resolution ---
	absSandboxDir, err := filepath.Abs(*sandboxDir)
	if err != nil {
		logger.Error("Failed to resolve absolute path for sandbox directory", "path", *sandboxDir, "error", err)
		fmt.Fprintf(os.Stderr, "Error resolving sandbox directory '%s': %v\n", *sandboxDir, err)
		os.Exit(1)
	}
	logger.Debug("Sandbox directory resolved", "path", absSandboxDir)
	// --- NeuroGo App Configuration & LLM Client Creation FIRST ---
	appConfig := &neurogo.Config{
		APIKey:        *apiKey,
		APIHost:       *apiHost,
		ModelName:     *modelName,
		StartupScript: *startupScriptPath,
		SandboxDir:    absSandboxDir,
		Insecure:      *insecure,
		LibPaths:      libPathsConfig.Value,
		TargetArg:     *targetArg,
		ProcArgs:      procArgsConfig.Value,
		// Note: SyncDir, SyncFilter, SyncIgnoreGitignore, AllowlistFile, SchemaPath are not set here from flags
		// They might be intended to be set by scripts or TUI later if needed, or flags are missing.
	}

	// Create a temporary App shell or pass nil if CreateLLMClient doesn't strictly need a full App instance,
	// OR, if CreateLLMClient is a static/package-level function.
	// HOWEVER, app.CreateLLMClient() is a METHOD on *App.
	// This implies App needs to exist at least partially to call CreateLLMClient.
	// But NewApp needs the llmClient. This is the chicken/egg.

	// Let's look at what App.CreateLLMClient() *actually* needs from `app`:
	// It needs `app.Config` and `app.Log`.
	// It does NOT set `app.llmClient`.

	// Solution:
	// 1. Create a preliminary App instance just for its Config and Log, to call CreateLLMClient.
	// 2. Create the LLM Client using this preliminary app instance.
	// 3. Create the *final* App instance by passing this LLM client to NewApp.

	logger.Debug("Preparing to create LLM client...")
	// Temporary app instance to facilitate CreateLLMClient call, as it needs app.Config and app.Log
	tempAppForLLMCreation := &neurogo.App{Config: appConfig, Log: logger}
	llmClient, err := tempAppForLLMCreation.CreateLLMClient()
	if err != nil {
		logger.Error("Failed to create LLM client", "error", err)
		fmt.Fprintf(os.Stderr, "LLM client creation error: %v\n", err)
		os.Exit(1)
	}
	if llmClient == nil { // Should be caught by CreateLLMClient's internal error handling, but good to check.
		logger.Error("LLM client is nil after creation without error.")
		fmt.Fprintf(os.Stderr, "LLM client is nil after creation.\n")
		os.Exit(1)
	}
	logger.Debug("LLM client created successfully.")

	// --- NeuroGo App Initialization (with LLM Client) ---
	app, err := neurogo.NewApp(appConfig, logger, llmClient) // Now passing llmClient
	if err != nil {                                          // NewApp now returns an error
		logger.Error("Failed to create NeuroGo App", "error", err)
		fmt.Fprintf(os.Stderr, "NeuroGo App creation error: %v\n", err)
		os.Exit(1)
	}
	// app.Config = appConfig; // This line is redundant if NewApp assigns cfg to app.Config, which it does.

	logger.Debug("NeuroGo App instance created.")

	// --- Initialize Core Components (Interpreter, AIWM) ---
	// The 'llmClient' passed to InitializeCoreComponents is the one already created and given to NewApp.
	// InitializeCoreComponents will then use this to set up Interpreter and AIWM.
	var interpreter *core.Interpreter // Keep var declarations if they are used later for tool registration
	var aiWm *core.AIWorkerManager    // Keep var declarations

	interpreter, aiWm, err = InitializeCoreComponents(app, logger, llmClient) // llmClient here is the one created above
	if err != nil {
		logger.Error("Failed to initialize core components", "error", err)
		fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
		os.Exit(1)
	}
	logger.Debug("Core components (Interpreter, AIWM) initialized successfully.")
	// --- Register Tools ---
	if aiWm != nil {
		if err := core.RegisterAIWorkerTools(interpreter); err != nil {
			logger.Error("Failed to register AI Worker tools", "error", err)
			fmt.Fprintf(os.Stderr, "Warning: Failed to register AI Worker tools: %v\n", err)
		} else {
			logger.Debug("AI Worker tools registered.")
		}
	} else {
		logger.Warn("AI Worker Manager not initialized, skipping AI Worker tool registration.")
	}

	// --- Determine Mode of Operation ---
	scriptToRunNonTUI := app.Config.StartupScript
	if scriptToRunNonTUI == "" && flag.NArg() > 0 && !*tuiMode {
		scriptToRunNonTUI = flag.Arg(0)
		logger.Debug("Using positional argument as script for non-TUI execution", "script", scriptToRunNonTUI)
	}

	if *tuiMode {
		logger.Debug("TUI mode requested. Starting TUI...")
		if err := neurogo.StartTviewTUI(app, app.Config.StartupScript); err != nil {
			logger.Error("TUI Error", "error", err)
			fmt.Fprintf(os.Stderr, "TUI Error: %v\n", err)
			os.Exit(1)
		}
		logger.Debug("TUI finished.")
	} else {
		// --- Non-TUI Mode ---
		scriptExecutedInNonTUI := false
		if scriptToRunNonTUI != "" {
			logger.Debug("Executing script (non-TUI mode)", "script", scriptToRunNonTUI)
			originalConfigScript := app.Config.StartupScript
			if app.Config.StartupScript != scriptToRunNonTUI && scriptToRunNonTUI == flag.Arg(0) && flag.NArg() > 0 {
				app.Config.StartupScript = scriptToRunNonTUI
			}

			if err := app.ExecuteScriptFile(ctx, app.Config.StartupScript); err != nil {
				logger.Error("Error executing script", "script", app.Config.StartupScript, "error", err)
				fmt.Fprintf(os.Stderr, "Error executing script '%s': %v\n", app.Config.StartupScript, err)
				if app.Config.StartupScript != originalConfigScript {
					app.Config.StartupScript = originalConfigScript
				}
				os.Exit(1)
			}

			if app.Config.StartupScript != originalConfigScript {
				app.Config.StartupScript = originalConfigScript
			}
			logger.Debug("Script finished successfully (non-TUI mode).")
			scriptExecutedInNonTUI = true
		}

		if scriptExecutedInNonTUI {
			logger.Debug("Script executed (non-TUI mode). Exiting.")
		} else if *replMode {
			logger.Debug("No script run or TUI disabled. Starting basic REPL...")
			runRepl(ctx, app)
			logger.Debug("Basic REPL finished.")
		} else {
			if app.Config.StartupScript == "" && flag.NArg() == 0 && !*replMode {
				logger.Debug("No TUI, no script, no REPL. Nothing to do. Exiting.")
				fmt.Println("No action specified. Use -script <file>, -tui, or -repl, or provide a script file as an argument.")
				flag.Usage()
			} else {
				logger.Debug("Non-TUI mode: No further action. Exiting.")
			}
		}
	}

	logger.Debug("NeuroScript application finished.")
	if closer, ok := logger.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing logger: %v\n", err)
		}
	}
}
