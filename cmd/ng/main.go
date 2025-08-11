// NeuroScript Version: 0.5.0
// File version: 9
// Purpose: Main entry point. Correctly handles build-time version injection.
// filename: cmd/ng/main.go
// nlines: 77
// risk_rating: LOW
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/neurogo"
)

// AppVersion is injected at build time by the Makefile using -ldflags.
var AppVersion string

// main is the entry point. It defines and parses flags, handles the version
// command, then populates a config struct and passes it to the main Run function.
func main() {
	// --- Flag Definitions ---
	versionFlag := flag.Bool("version", false, "Print version information in JSON format and exit")
	insecure := flag.Bool("insecure", false, "Disable security checks (Use with extreme caution!)")
	startupScriptPath := flag.String("script", "", "Path to a NeuroScript (.ns) file to execute")
	tuiMode := flag.Bool("tui", false, "Enable Terminal User Interface (TUI) mode")
	replMode := flag.Bool("repl", false, "Enable basic REPL mode (if TUI is false and no script is run)")
	targetArg := flag.String("target", "main", "Target procedure for the script")
	procArgsConfig := neurogo.NewStringSliceFlag() // This helper is still convenient
	flag.Var(procArgsConfig, "arg", "Argument for the script process/procedure (can be specified multiple times)")

	// --- Privileged Configuration Flags ---
	trustedConfig := flag.String("trusted-config", "", "Path to a trusted startup script that runs with elevated privileges")
	trustedConfigTarget := flag.String("trusted-config-target", "main", "Target procedure for the trusted-config script")
	trustedTargetArgs := neurogo.NewStringSliceFlag()
	flag.Var(trustedTargetArgs, "trusted-target-arg", "Argument for the trusted-config script's target procedure (can be specified multiple times)")

	flag.Parse()

	// --- Handle Version Flag ---
	if *versionFlag {
		// Use the build-time AppVersion if available, otherwise fall back to the api package version.
		displayVersion := AppVersion
		if displayVersion == "" {
			displayVersion = api.ProgramVersion
		}

		versionInfo := struct {
			AppVersion     string `json:"app_version"`
			GrammarVersion string `json:"grammar_version"`
		}{
			AppVersion:     displayVersion,
			GrammarVersion: api.GrammarVersion,
		}
		jsonOutput, _ := json.MarshalIndent(versionInfo, "", "  ")
		fmt.Println(string(jsonOutput))
		os.Exit(0)
	}

	// --- Populate Config from Flags ---
	cfg := CliConfig{
		Insecure:            *insecure,
		StartupScript:       *startupScriptPath,
		TuiMode:             *tuiMode,
		ReplMode:            *replMode,
		TargetArg:           *targetArg,
		ProcArgs:            procArgsConfig.Value,
		PositionalScript:    flag.Arg(0),
		TrustedConfig:       *trustedConfig,
		TrustedConfigTarget: *trustedConfigTarget,
		TrustedTargetArgs:   trustedTargetArgs.Value,
	}

	// --- Delegate to the main runner ---
	os.Exit(Run(cfg))
}
