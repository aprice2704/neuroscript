// filename: cmd/neurogo/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/neurogo"
	"github.com/aprice2704/neuroscript/pkg/neurogo/tui"
)

func main() {
	// --- Configuration Setup ---
	// Logging (Keep these)
	logFile := flag.String("log-file", "", "Path to log file (optional, defaults to stderr)")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")

	// Modes (Keep these)
	scriptFile := flag.String("script", "", "Path to a NeuroScript file to execute (script mode)")
	agentMode := flag.Bool("agent", false, "Run in interactive agent mode (uses -startup-script)")
	syncMode := flag.Bool("sync", false, "Run in sync-only mode") // Sync logic might also move to AgentContext/startup later
	cleanAPI := flag.Bool("clean-api", false, "Delete all files from the File API (use with caution!)")
	tuiMode := flag.Bool("tui", false, "Run in interactive TUI mode (experimental)") // TUI might also use startup script

	// Agent Startup Script (NEW)
	startupScript := flag.String("startup-script", "agent_startup.ns", "Path to NeuroScript file for agent initialization")

	// Essential Config (Keep/Review)
	apiKey := flag.String("api-key", os.Getenv("GEMINI_API_KEY"), "Gemini API Key (env: GEMINI_API_KEY)")
	insecure := flag.Bool("insecure", false, "Disable security checks (Use with extreme caution!)") // Keep for now, review necessity

	// --- REMOVED FLAGS ---
	// modelName := flag.String("model", "gemini-1.5-flash", ...) -> Handled by startup script
	// sandboxDir := flag.String("sandbox", ".", ...) -> Handled by startup script
	// allowlistFile := flag.String("allowlist", "", ...) -> Handled by startup script
	// initialAttachments := flag.String("attach", "", ...) -> Handled by startup script (e.g., TOOL.AgentPinFile)
	// syncDir := flag.String("sync-dir", ".", ...) -> Handled by startup script (or kept for -sync mode?)
	// syncFilter := flag.String("sync-filter", "", ...) -> Handled by startup script (or kept for -sync mode?)
	// syncIgnoreGitignore := flag.Bool("sync-ignore-gitignore", false, ...) -> Handled by startup script (or kept for -sync mode?)

	// TODO: Revisit sync-related flags. Do they only apply to -sync mode, or should agent's TOOL.Sync also use them?
	// If only for -sync mode, keep them. If used by agent tools, remove them and handle via startup script.
	// For now, let's assume they are primarily for the standalone -sync mode and keep them minimally.
	syncDir := flag.String("sync-dir", ".", "Directory for sync operations (-sync mode)")
	syncFilter := flag.String("sync-filter", "", "Glob pattern for sync (-sync mode)")
	syncIgnoreGitignore := flag.Bool("sync-ignore-gitignore", false, "Ignore .gitignore during sync (-sync mode)")

	// TODO: Decide on flag parsing library. The `flag` package is basic.
	// Consider `pflag` or `cobra` if more complex flag interactions are needed.
	// The simple `flag` package doesn't handle default mode logic as elegantly.
	flag.Parse()

	// --- Determine Mode ---
	// Basic mode precedence check. This logic could be improved.
	// Config parsing in config.go handled this better, consider moving back.
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

	if *cleanAPI { // Highest precedence
		runScript, runAgent, runSync, runTui = false, false, false, false
	} else if *syncMode {
		runScript, runAgent, runCleanAPI, runTui = false, false, false, false
	} else if *scriptFile != "" {
		runAgent, runSync, runCleanAPI, runTui = false, false, false, false
	} else if *tuiMode {
		runAgent, runSync, runCleanAPI, runScript = false, false, false, false
	}
	// If agent flag wasn't explicitly set BUT no other mode was, default to agent mode
	if !runScript && !runAgent && !runSync && !runCleanAPI && !runTui {
		runAgent = true
		fmt.Fprintln(os.Stderr, "Defaulting to interactive agent mode.")
	}
	// Simple check for multiple modes after defaulting
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
		fmt.Fprintln(os.Stderr, "Error: Only one execution mode (-script, -agent, -sync, -clean-api, -tui) can be specified.")
		flag.Usage() // Consider defining a better Usage message
		os.Exit(1)
	}

	// --- Logger Initialization ---
	if *logFile != "" {
		logOutput, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening log file %s: %v\n", *logFile, err)
			os.Exit(1)
		}
	}

	// --- App Initialization ---
	app := neurogo.NewApp() // Constructor remains the same

	// Configure app.Config struct with REMAINING flags
	// The core configuration (sandbox, model, etc.) will be set
	// via the startup script and AgentContext for agent/tui modes.
	// We pass the paths/flags needed to *find* and run those scripts.
	app.Config.APIKey = *apiKey               // Still needed globally for client init
	app.Config.Insecure = *insecure           // Security flag
	app.Config.StartupScript = *startupScript // NEW: Pass startup script path

	// Set mode flags in config
	app.Config.RunAgentMode = runAgent
	app.Config.RunScriptMode = runScript
	app.Config.RunSyncMode = runSync
	app.Config.RunCleanAPIMode = runCleanAPI
	app.Config.RunTuiMode = runTui

	// Pass script-specific args if in script mode
	if runScript {
		app.Config.ScriptFile = *scriptFile
		// TODO: Re-add -L, -target, -arg flags if needed for script mode
		// app.Config.LibPaths = ...
		// app.Config.TargetArg = ...
		// app.Config.ProcArgs = flag.Args() // Or handle args more explicitly
	}

	// Pass sync-specific args if in sync mode
	if runSync {
		app.Config.SyncDir = *syncDir
		app.Config.SyncFilter = *syncFilter
		app.Config.SyncIgnoreGitignore = *syncIgnoreGitignore
	}

	// --- Application Execution ---
	// TUI mode might eventually leverage AgentContext/startup script too,
	// but keep its separate invocation for now.
	if runTui {
		app.Logger.Info("Starting in TUI mode...")
		if err := tui.Start(app); err != nil {
			app.Logger.Error("TUI Error: %v", err)
			fmt.Fprintf(os.Stderr, "TUI Error: %v\n", err)
			os.Exit(1)
		}
		app.Logger.Debug("TUI finished.")
		os.Exit(0)
	}

	// --- Run selected mode via app.Run ---
	ctx := context.Background()
	if err := app.Run(ctx); err != nil {
		// app.Run handles logging the error detail internally
		fmt.Fprintf(os.Stderr, "Error: %v\n", err) // Keep simple stderr message
		os.Exit(1)
	}

	app.Logger.Info("NeuroGo finished successfully.")
}
