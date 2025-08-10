// NeuroScript Version: 0.5.0
// File version: 4
// Purpose: Main entry point. Handles CLI flag parsing and version check, then delegates to Run().
// filename: cmd/ng/main.go
// nlines: 83
// risk_rating: MEDIUM
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/neurogo"
)

// Version information, injected at build time via -ldflags.
var (
	AppVersion string
)

// main is the entry point. It defines and parses flags, handles the version
// command, then populates a config struct and passes it to the main Run function.
func main() {
	// --- Flag Definitions ---
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
		jsonOutput, _ := json.MarshalIndent(versionInfo, "", "  ")
		fmt.Println(string(jsonOutput))
		os.Exit(0)
	}

	// --- Populate Config from Flags ---
	cfg := CliConfig{
		LogFile:          *logFile,
		LogLevel:         *logLevel,
		SandboxDir:       *sandboxDir,
		APIKey:           *apiKey,
		APIHost:          *apiHost,
		Insecure:         *insecure,
		ModelName:        *modelName,
		StartupScript:    *startupScriptPath,
		TuiMode:          *tuiMode,
		ReplMode:         *replMode,
		LibPaths:         libPathsConfig.Value,
		TargetArg:        *targetArg,
		ProcArgs:         procArgsConfig.Value,
		PositionalScript: flag.Arg(0), // First positional argument
	}

	// --- Delegate to the main runner ---
	os.Exit(Run(cfg))
}
