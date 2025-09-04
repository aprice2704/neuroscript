// NeuroScript Version: 0.7.0
// File version: 19
// Purpose: Fixes issue where emit was not producing output by using SetEmitFunc.
// filename: cmd/ng/run.go
// nlines: 200
// risk_rating: HIGH
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aprice2704/neuroscript/pkg/api"
	// The provider is no longer directly instantiated here.
)

// CliConfig holds all configuration passed from the command line flags.
type CliConfig struct {
	Insecure            bool
	StartupScript       string
	TuiMode             bool
	ReplMode            bool
	TargetArg           string
	ProcArgs            []string
	PositionalScript    string
	TrustedConfig       string
	TrustedConfigTarget string
	TrustedTargetArgs   []string
}

// Run executes the main application logic based on the provided configuration and returns an exit code.
func Run(cfg CliConfig) int {
	// --- Application Context ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived signal, shutting down...")
		cancel()
	}()

	var interp *api.Interpreter

	// --- Execute Trusted Config Script ---
	if cfg.TrustedConfig != "" {
		fmt.Printf("Executing trusted config script: %s\n", cfg.TrustedConfig)
		requiredGrants := []api.Capability{
			{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
			{Resource: "model", Verbs: []string{"use"}, Scopes: []string{"*"}},
			{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"*"}},
			{Resource: "net", Verbs: []string{"read", "write"}, Scopes: []string{"*"}},
		}
		var allowedTools []string
		// Trusted interpreter created via NewConfigInterpreter will have default providers.
		trustedInterp := api.NewConfigInterpreter(
			allowedTools,
			requiredGrants,
			// api.WithStdout(os.Stdout), // This was not working as expected.
			api.WithStderr(os.Stderr),
		)

		// FIX: Explicitly set a handler for the 'emit' command to ensure output.
		trustedInterp.SetEmitFunc(func(v api.Value) {
			val, err := api.Unwrap(v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[emit error] %v\n", err)
				return
			}
			fmt.Println(val)
		})

		fmt.Println("Interpreter created with elevated privileges.")

		args := stringSliceToAnySlice(cfg.TrustedTargetArgs)
		err := executeScript(ctx, trustedInterp, cfg.TrustedConfig, cfg.TrustedConfigTarget, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error in trusted config script '%s': %v\n", cfg.TrustedConfig, err)
			return 1
		}
		fmt.Println("Trusted config script finished successfully.")
	}

	// --- Determine Mode of Operation ---
	scriptToRunNonTUI := cfg.StartupScript
	if scriptToRunNonTUI == "" && cfg.PositionalScript != "" && !cfg.TuiMode {
		scriptToRunNonTUI = cfg.PositionalScript
	}

	// --- Create Standard Interpreter for subsequent operations ---
	if scriptToRunNonTUI != "" || cfg.TuiMode || cfg.ReplMode {
		// api.New() now handles registration of default providers automatically.
		interp = api.New(
			api.WithStderr(os.Stderr),
		)
		// Also apply the emit fix to the standard interpreter.
		interp.SetEmitFunc(func(v api.Value) {
			val, err := api.Unwrap(v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[emit error] %v\n", err)
				return
			}
			fmt.Println(val)
		})

		fmt.Println("Interpreter created with standard privileges.")
	}

	// TUI Mode (Conceptual)
	if cfg.TuiMode {
		fmt.Println("TUI mode requested. (Note: TUI needs to be updated to use the new api.Interpreter)")
		return 0
	}

	// Non-TUI Script Execution
	if scriptToRunNonTUI != "" {
		fmt.Printf("Executing script: %s\n", scriptToRunNonTUI)
		args := stringSliceToAnySlice(cfg.ProcArgs)
		err := executeScript(ctx, interp, scriptToRunNonTUI, cfg.TargetArg, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Script execution failed for '%s': %v\n", scriptToRunNonTUI, err)
			return 1
		}
		fmt.Println("Script finished successfully.")
		return 0
	}

	// REPL Mode (Conceptual)
	if cfg.ReplMode {
		fmt.Println("Starting basic REPL... (Note: REPL needs to be updated for the new api.Interpreter)")
		return 0
	}

	if cfg.TrustedConfig == "" {
		fmt.Println("No action specified. Use -trusted-config, -script <file>, -tui, or -repl.")
	}

	fmt.Println("NeuroScript application finished.")
	return 0
}

func executeScript(ctx context.Context, interp *api.Interpreter, scriptPath string, target string, args []any) error {
	scriptBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("could not read script file '%s': %w", scriptPath, err)
	}

	tree, err := api.Parse(scriptBytes, api.ParseSkipComments)
	if err != nil {
		return fmt.Errorf("failed to parse script '%s': %w", scriptPath, err)
	}

	if _, err := api.ExecWithInterpreter(ctx, interp, tree); err != nil {
		return fmt.Errorf("failed to load script '%s' into interpreter: %w", scriptPath, err)
	}

	if target != "" {
		fmt.Printf("Running procedure '%s'...\n", target)
		if _, err := api.RunProcedure(ctx, interp, target, args...); err != nil {
			return fmt.Errorf("error executing procedure '%s': %w", target, err)
		}
	}

	return nil
}

func stringSliceToAnySlice(ss []string) []any {
	if ss == nil {
		return nil
	}
	as := make([]any, len(ss))
	for i, s := range ss {
		as[i] = s
	}
	return as
}
