// filename: cmd/neurogo/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"io" // Needed for io.Discard
	"log"
	"os"
	"strings"

	// tea "github.com/charmbracelet/bubbletea" // Not needed directly in main
	// "github.com/google/generative-ai-go/genai" // Not needed directly in main
	// "google.golang.org/api/option" // Not needed directly in main

	// "github.com/aprice2704/neuroscript/pkg/core" // Only needed if using core types/consts directly
	"github.com/aprice2704/neuroscript/pkg/neurogo"
	"github.com/aprice2704/neuroscript/pkg/neurogo/tui" // Import the new TUI package
)

func main() {
	// --- Configuration Setup ---
	// Logging
	logFile := flag.String("log-file", "", "Path to log file (optional, defaults to stderr)")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	debugLogFile := flag.String("debug-log-file", "", "Path to debug log file (optional)")
	llmDebugLogFile := flag.String("llm-debug-log-file", "", "Path to LLM debug log file (optional)")

	// Modes
	scriptFile := flag.String("script", "", "Path to a NeuroScript file to execute")
	agentMode := flag.Bool("agent", false, "Run in interactive agent mode")
	syncMode := flag.Bool("sync", false, "Run in sync-only mode (uses -sync-dir, -sync-filter)")
	cleanAPI := flag.Bool("clean-api", false, "Delete all files from the File API (use with caution!)")

	// Common Config
	apiKey := flag.String("api-key", os.Getenv("GEMINI_API_KEY"), "Gemini API Key (env: GEMINI_API_KEY)")
	modelName := flag.String("model", "gemini-1.5-flash", "Gemini model name")
	sandboxDir := flag.String("sandbox", ".", "Sandbox directory for restricted mode") // Use literal default
	insecure := flag.Bool("insecure", false, "Disable security checks (Use with extreme caution!)")
	allowlistFile := flag.String("allowlist", "", "Path to tool allowlist file (agent mode only)")
	initialAttachments := flag.String("attach", "", "Comma-separated list of initial files to attach (agent mode only)")
	syncDir := flag.String("sync-dir", ".", "Directory to use for sync operations (default: ., used by /sync)") // Use literal default
	syncFilter := flag.String("sync-filter", "", "Optional glob pattern to filter files during sync")
	syncIgnoreGitignore := flag.Bool("sync-ignore-gitignore", false, "Ignore .gitignore files during sync (default: false)")
	tuiMode := flag.Bool("tui", false, "Run in interactive TUI mode (experimental)")

	flag.Parse()

	// --- Logger Initialization ---
	var logOutput io.Writer = os.Stderr
	var err error
	if *logFile != "" {
		logOutput, err = os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening log file %s: %v\n", *logFile, err)
			os.Exit(1)
		}
		// Note: Cannot defer close on os.Stderr. If logFile is used, it won't be closed until exit.
		// For long-running processes, might need more sophisticated log rotation/handling.
	}

	flags := log.Ldate | log.Ltime | log.Lshortfile
	infoLog := log.New(logOutput, "INFO:  ", flags)
	warnLog := log.New(logOutput, "WARN:  ", flags) // Create WarnLog
	errorLog := log.New(logOutput, "ERROR: ", flags)
	debugLog := log.New(io.Discard, "DEBUG: ", flags)
	llmLog := log.New(io.Discard, "LLM:   ", flags)

	switch strings.ToLower(*logLevel) {
	case "debug":
		debugLog.SetOutput(logOutput)
		llmLog.SetOutput(logOutput)
		fallthrough
	case "info":
		// Info enabled by default via infoLog
		fallthrough
	case "warn":
		// Warn enabled by default via warnLog
		fallthrough
	case "error":
		// Error enabled by default via errorLog
	default:
		infoLog.Printf("Invalid log level '%s', defaulting to info", *logLevel)
	}

	// --- App Initialization ---
	app := neurogo.NewApp() // Call constructor

	// Assign loggers to the App instance
	app.InfoLog = infoLog
	// *** NOTE: Ensure 'WarnLog *log.Logger' field exists in App struct in pkg/neurogo/app.go ***
	app.WarnLog = warnLog
	app.ErrorLog = errorLog
	app.DebugLog = debugLog
	app.LLMLog = llmLog

	// Process initial attachments string into a slice
	var initialAttachPaths []string
	if *initialAttachments != "" {
		initialAttachPaths = strings.Split(*initialAttachments, ",")
		for i := range initialAttachPaths {
			initialAttachPaths[i] = strings.TrimSpace(initialAttachPaths[i])
		}
	}

	// Configure app.Config struct
	app.Config.APIKey = *apiKey
	app.Config.ModelName = *modelName
	app.Config.SandboxDir = *sandboxDir
	app.Config.ScriptFile = *scriptFile
	app.Config.RunAgentMode = *agentMode
	app.Config.RunSyncMode = *syncMode
	// *** NOTE: Assuming 'Insecure bool' field exists in Config struct in pkg/neurogo/config.go ***
	app.Config.Insecure = *insecure // This line assumes the field exists
	app.Config.InitialAttachments = initialAttachPaths
	app.Config.SyncDir = *syncDir
	app.Config.SyncFilter = *syncFilter
	app.Config.SyncIgnoreGitignore = *syncIgnoreGitignore
	app.Config.AllowlistFile = *allowlistFile
	app.Config.CleanAPI = *cleanAPI
	app.Config.DebugLogFile = *debugLogFile
	app.Config.LLMDebugLogFile = *llmDebugLogFile

	// --- Application Execution ---
	if *tuiMode {
		// --- TUI Mode ---
		app.DebugLog.Println("Starting in TUI mode...")
		// Manually initialize logging/LLM client before starting TUI
		// Assumes InitLoggingAndLLMClient exists in app.go
		if err := app.InitLoggingAndLLMClient(context.Background()); err != nil {
			app.ErrorLog.Printf("TUI Mode: Failed to initialize prerequisites: %v", err)
			// Decide if TUI can run without LLM - logging error but continuing for now
		}

		if err := tui.Start(app); err != nil { // Call the TUI start function
			app.ErrorLog.Fatalf("TUI Error: %v", err)
			fmt.Fprintf(os.Stderr, "TUI Error: %v\n", err)
			os.Exit(1)
		}
		app.DebugLog.Println("TUI finished.")
		os.Exit(0)
	}

	// --- Original CLI Modes ---
	ctx := context.Background()
	if err := app.Run(ctx); err != nil {
		// app.Run logs fatal error internally
		fmt.Fprintf(os.Stderr, "Error: %v\n", err) // Keep simple stderr message
		os.Exit(1)
	}

	app.InfoLog.Println("NeuroGo finished successfully.")
}

// resolvePath function removed as it's no longer needed
