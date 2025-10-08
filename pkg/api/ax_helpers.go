// NeuroScript Version: 0.7.4
// File version: 1
// Purpose: Provides public helper functions for interacting with the ax API, abstracting away internal types.
// filename: pkg/api/ax_helpers.go
// nlines: 55
// risk_rating: MEDIUM

package api

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ax"
)

// AXBootLoad parses a boot script, loads its definitions into a new Config runner,
// and executes its top-level commands to configure the shared environment.
func AXBootLoad(ctx context.Context, fac ax.RunnerFactory, src []byte) error {
	cfg, err := fac.NewRunner(ctx, ax.RunnerConfig, ax.RunnerOpts{})
	if err != nil {
		return fmt.Errorf("ax boot: failed to create new config runner: %w", err)
	}

	tree, err := Parse(src, ParseSkipComments)
	if err != nil {
		return fmt.Errorf("ax boot: failed to parse source: %w", err)
	}

	// Safely unwrap the runner to its internal implementation within the api package.
	r, ok := cfg.(*axRunner)
	if !ok {
		return fmt.Errorf("ax boot: unsupported runner implementation")
	}

	// Use the non-destructive RunScriptWithInterpreter to load and execute.
	if _, err := RunScriptWithInterpreter(ctx, r.itp, tree); err != nil {
		return fmt.Errorf("ax boot: failed to run script: %w", err)
	}

	return nil
}

// AXRunScript parses, loads, and runs a procedure from a script on a User runner.
// It is a convenience wrapper that handles value wrapping and unwrapping.
func AXRunScript(ctx context.Context, run ax.Runner, src []byte, entry string, args ...any) (any, error) {
	tree, err := Parse(src, ParseSkipComments)
	if err != nil {
		return nil, fmt.Errorf("ax run: failed to parse source: %w", err)
	}

	r, ok := run.(*axRunner)
	if !ok {
		return nil, fmt.Errorf("ax run: unsupported runner implementation")
	}

	// Use AppendScript to add definitions without executing commands automatically.
	if err := r.itp.AppendScript(tree); err != nil {
		return nil, fmt.Errorf("ax run: failed to load script definitions: %w", err)
	}

	// Wrap args to engine values, run the specified procedure, and unwrap the result.
	val, err := RunProcedure(ctx, r.itp, entry, args...)
	if err != nil {
		return nil, err // The error from RunProcedure is already descriptive.
	}
	return Unwrap(val)
}

// AXInterpreter retrieves the internal interpreter from an ax.Runner.
// This should be used sparingly, primarily for testing and diagnostics.
func AXInterpreter(r ax.Runner) (*Interpreter, bool) {
	runner, ok := r.(*axRunner)
	if !ok {
		return nil, false
	}
	return runner.itp, true
}
