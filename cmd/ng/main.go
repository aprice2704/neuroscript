// NeuroScript Version: 0.3.1
// File version: 0.2.1
// Purpose: Removed AI Worker Manager related code after its functionality was excised from the project.
// filename: cmd/ng/main.go
// nlines: 240
// risk_rating: HIGH
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

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/neurogo"
)

// Version information, injected at build time via -ldflags.
var (
	AppVersion string
)

func main() {

	// --- Configuration Setup (Flag Definitions) ---
	versionFlag := flag.Bool("version", false, "Print version information in JSON format and exit")
	logFile := flag.String("log-file", "", "Path to log file (optional, defaults to stderr)")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	sandboxDir := flag.String("sandbox", ".", "Root directory for secure file operations")

	apiKey := flag.String("api-key", os.Getenv("GEMINI_API_KEY"), "Gemini API Key (env: GEMINI_API_KEY or NEUROSCRIPT_API_KEY)")
	if *apiKey == "" {
		*apiKey = os.Getenv("NEUROSCRIPT_API_KEY")
	}
	apiHost := flag.String("api-host", "", "Optional API Host/Endpoint override")
	insecure := flag.Bool("insecure", false, "Disable security checks (Use with extreme caution!)")
	modelName := flag.String("model", neurogo.DefaultModelName, "Default generative model name for LLM interactions")
	startupScriptPath := flag.String("script", "", "Path to a NeuroScript (.ns) file to execute")

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
	if *versionFlag {
		appVersion := AppVersion
		if appVersion == "" {
			appVersion = "dev"
		}
		grammarVersion := lang.GrammarVersion
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
	}

	logger.Debug("Preparing to create LLM client...")
	tempAppForLLMCreation := &neurogo.App{Config: appConfig, Log: logger}
	llmClient, err := tempAppForLLMCreation.CreateLLMClient()
	if err != nil {
		logger.Error("Failed to create LLM client", "error", err)
		fmt.Fprintf(os.Stderr, "LLM client creation error: %v\n", err)
		os.Exit(1)
	}
	if llmClient == nil {
		logger.Error("LLM client is nil after creation without error.")
		fmt.Fprintf(os.Stderr, "LLM client is nil after creation.\n")
		os.Exit(1)
	}
	logger.Debug("LLM client created successfully.")

	// --- NeuroGo App Initialization (with LLM Client) ---
	app, err := neurogo.NewApp(appConfig, logger, llmClient)
	if err != nil {
		logger.Error("Failed to create NeuroGo App", "error", err)
		fmt.Fprintf(os.Stderr, "NeuroGo App creation error: %v\n", err)
		os.Exit(1)
	}
	logger.Debug("NeuroGo App instance created.")

	// --- Initialize Core Components (Interpreter) ---
	if err := app.InitializeCoreComponents(); err != nil {
		logger.Error("Failed to initialize core components", "error", err)
		fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
		os.Exit(1)
	}
	logger.Debug("Core components (Interpreter) initialized successfully.")
	fmt.Printf("---> STARTING NG with %d tools <---\n", app.GetInterpreter().NTools())

	// --- Determine Mode of Operation ---
	scriptToRunNonTUI := app.Config.StartupScript
	if scriptToRunNonTUI == "" && flag.NArg() > 0 && !*tuiMode {
		scriptToRunNonTUI = flag.Arg(0)
		logger.Debug("Using types.Positional argument as script for non-TUI execution", "script", scriptToRunNonTUI)
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

			// --- NEW PROTOCOL IMPLEMENTATION ---
			interpreter := app.GetInterpreter()

			// 1. Read file content via interpreter tool
			filepathArg, wrapErr := lang.Wrap(scriptToRunNonTUI)
			if wrapErr != nil {
				err := fmt.Errorf("internal error wrapping script path: %w", wrapErr)
				logger.Error(err.Error())
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			toolArgs := map[string]lang.Value{"filepath": filepathArg}
			contentValue, toolErr := interpreter.ExecuteTool("TOOL.ReadFile", toolArgs)
			if toolErr != nil {
				err := fmt.Errorf("error reading script '%s': %w", scriptToRunNonTUI, toolErr)
				logger.Error(err.Error())
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
			scriptContent, ok := lang.Unwrap(contentValue).(string)
			if !ok {
				err := fmt.Errorf("internal error: TOOL.ReadFile did not return a string")
				logger.Error(err.Error())
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			// 2. Load the script from its content string
			if _, loadErr := app.LoadScriptString(ctx, scriptContent); loadErr != nil {
				err := fmt.Errorf("error loading script '%s': %w", scriptToRunNonTUI, loadErr)
				logger.Error(err.Error())
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			// 3. Run the entrypoint (TargetArg)
			if app.Config.TargetArg != "" {
				logger.Info("Executing entrypoint.", "procedure", app.Config.TargetArg, "args", app.Config.ProcArgs)
				if _, runErr := app.RunProcedure(ctx, app.Config.TargetArg, app.Config.ProcArgs); runErr != nil {
					err := fmt.Errorf("error executing entrypoint '%s': %w", app.Config.TargetArg, runErr)
					logger.Error(err.Error())
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}
			}
			// --- END NEW PROTOCOL IMPLEMENTATION ---

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
