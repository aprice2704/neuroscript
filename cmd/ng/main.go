package main

import (
	"context" // Added context import
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	// Added imports for signal and syscall for graceful shutdown context
	"os/signal"
	"syscall" // Added syscall import

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/neurogo"
	"github.com/aprice2704/neuroscript/pkg/neurogo/tui"
)

// initializeLogger sets up the slog logger based on configuration strings.
func initializeLogger(levelStr string, filePath string) (logging.Logger, error) {
	var level slog.Level
	// Parse the log level string
	switch levelStr {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid log level: %q", levelStr)
	}

	var output io.Writer = os.Stderr // Default to stderr

	// If a log file path is provided, open or create the file
	if filePath != "" {
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// Return error, cannot proceed with file logging
			return nil, fmt.Errorf("failed to open log file %q: %w", filePath, err)
		}
		// Note: Closing the file will be the responsibility of the caller or program exit.
		output = file
	}

	// Create the slog handler (TextHandler as requested)
	handlerOpts := &slog.HandlerOptions{
		Level: level,
		// AddSource: true, // Optionally add source file/line numbers
	}
	handler := slog.NewTextHandler(output, handlerOpts)

	// Create the slog logger
	slogLogger := slog.New(handler)

	// Create the adapter
	appLogger, err := adapters.NewSlogAdapter(slogLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger adapter: %w", err)
	}

	// Return the adapter which satisfies the logging.Logger interface
	return appLogger, nil
}

func main() {
	// --- Configuration Setup ---
	logFile := flag.String("log-file", "", "Path to log file (optional, defaults to stderr)")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")

	apiKey := flag.String("api-key", os.Getenv("GEMINI_API_KEY"), "Gemini API Key (env: GEMINI_API_KEY)")
	insecure := flag.Bool("insecure", false, "Disable security checks (Use with extreme caution!)")
	startupScript := flag.String("startup-script", "agent_startup.ns", "Path to NeuroScript file for agent initialization")

	syncDir := flag.String("sync-dir", ".", "Directory for sync operations (-sync mode)")
	syncFilter := flag.String("sync-filter", "", "Glob pattern for sync (-sync mode)")
	syncIgnoreGitignore := flag.Bool("sync-ignore-gitignore", false, "Ignore .gitignore during sync (-sync mode)")

	scriptFile := flag.String("script", "", "Path to a NeuroScript file to execute (script mode)")
	agentMode := flag.Bool("agent", false, "Run in interactive agent mode (uses -startup-script)")
	syncMode := flag.Bool("sync", false, "Run in sync-only mode")
	cleanAPI := flag.Bool("clean-api", false, "Delete all files from the File API (use with caution!)")
	tuiMode := flag.Bool("tui", false, "Run in interactive TUI mode (experimental)")

	flag.Parse()

	// --- Logger Initialization ---
	logger, err := initializeLogger(*logLevel, *logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	logger.Info("Logger initialized", "level", *logLevel, "file", *logFile) // Log initialization success

	// --- Determine Mode ---
	modeCount := 0
	if *scriptFile != "" {
		modeCount++
	}
	if *agentMode {
		modeCount++
	}
	if *syncMode {
		modeCount++
	}
	if *cleanAPI {
		modeCount++
	}
	if *tuiMode {
		modeCount++
	}

	runScript := *scriptFile != ""
	runAgent := *agentMode
	runSync := *syncMode
	runCleanAPI := *cleanAPI
	runTui := *tuiMode

	// Mode precedence
	if *cleanAPI {
		runScript, runAgent, runSync, runTui = false, false, false, false
	} else if *syncMode {
		runScript, runAgent, runCleanAPI, runTui = false, false, false, false
	} else if *scriptFile != "" {
		runAgent, runSync, runCleanAPI, runTui = false, false, false, false
	} else if *tuiMode {
		runAgent, runSync, runCleanAPI, runScript = false, false, false, false
	}
	// Default mode
	if !runScript && !runAgent && !runSync && !runCleanAPI && !runTui {
		runAgent = true
		logger.Info("Defaulting to interactive agent mode.")
	}
	// Final check for multiple modes
	finalModeCount := 0
	if runScript {
		finalModeCount++
	}
	if runAgent {
		finalModeCount++
	}
	if runSync {
		finalModeCount++
	}
	if runCleanAPI {
		finalModeCount++
	}
	if runTui {
		finalModeCount++
	}

	if finalModeCount > 1 {
		logger.Error("Multiple execution modes specified", "script", runScript, "agent", runAgent, "sync", runSync, "cleanAPI", runCleanAPI, "tui", runTui)
		fmt.Fprintln(os.Stderr, "Error: Only one execution mode (-script, -agent, -sync, -clean-api, -tui) can be specified.")
		flag.Usage()
		os.Exit(1)
	}

	// --- Create App and Populate Config ---
	app := neurogo.NewApp(logger) // Should log "Creating new App instance."

	// Check if app or app.Config is nil AFTER NewApp returns (defensive)
	if app == nil {
		logger.Error("neurogo.NewApp returned nil App instance")
		os.Exit(1)
	}
	if app.Config == nil {
		// Initialize config if NewApp doesn't guarantee it (it should)
		logger.Warn("neurogo.NewApp returned App with nil Config, initializing.")
		app.Config = neurogo.NewConfig() // Use constructor if available
	}

	app.Config.APIKey = *apiKey
	app.Config.Insecure = *insecure
	app.Config.StartupScript = *startupScript
	app.Config.RunAgentMode = runAgent
	app.Config.RunScriptMode = runScript
	app.Config.RunSyncMode = runSync
	app.Config.RunCleanAPIMode = runCleanAPI
	app.Config.RunTuiMode = runTui
	if runScript {
		app.Config.ScriptFile = *scriptFile
	}
	if runSync {
		app.Config.SyncDir = *syncDir
		app.Config.SyncFilter = *syncFilter
		app.Config.SyncIgnoreGitignore = *syncIgnoreGitignore
	}
	// Note: No Validate() call shown here, consistent with neurogo.Config definition

	// --- Application Execution ---
	if runTui {
		logger.Info("Starting in TUI mode...")
		if err := tui.Start(app); err != nil {
			logger.Error("TUI Error", "error", err)
			fmt.Fprintf(os.Stderr, "TUI Error: %v\n", err)
			os.Exit(1)
		}
		logger.Debug("TUI finished.")
		os.Exit(0)
	}

	// Setup context for non-TUI modes
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// --- ADDED DEBUG LOG ---
	logger.Debug("Configuration populated. About to call app.Run.")
	// --- END ADDED DEBUG LOG ---

	// --- Run selected mode via app.Run --- (Approx line 173 in original trace?)
	if err := app.Run(ctx); err != nil {
		// Error logging happens within app.Run using the injected logger
		fmt.Fprintf(os.Stderr, "Error: %v\n", err) // Simple final message
		os.Exit(1)
	}

	logger.Info("NeuroGo finished successfully.")
	os.Exit(0) // Explicit success exit
}
