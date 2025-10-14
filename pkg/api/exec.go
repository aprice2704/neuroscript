// NeuroScript Version: 0.8.0
// File version: 17
// Purpose: Adds debug output to trace the ExecPolicy within execution entry points.
// filename: pkg/api/exec.go
// nlines: 120
// risk_rating: HIGH

package api

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// ExecInNewInterpreter provides a stateless, one-shot execution.
func ExecInNewInterpreter(ctx context.Context, src string, opts ...interpreter.InterpreterOption) (Value, error) {
	tree, err := Parse([]byte(src), ParseSkipComments)
	if err != nil {
		return nil, fmt.Errorf("parsing failed in ExecInNewInterpreter: %w", err)
	}

	interp := New(opts...)
	return ExecWithInterpreter(ctx, interp, tree)
}

// ExecWithInterpreter executes top-level 'command' blocks using a persistent interpreter.
func ExecWithInterpreter(ctx context.Context, interp *Interpreter, tree *Tree) (Value, error) {
	if interp == nil || interp.internal == nil {
		return nil, fmt.Errorf("ExecWithInterpreter requires a non-nil interpreter")
	}
	if tree == nil || tree.Root == nil {
		return nil, fmt.Errorf("cannot execute a nil tree")
	}

	program, ok := tree.Root.(*ast.Program)
	if !ok {
		return nil, fmt.Errorf("internal error: tree root is not a runnable *ast.Program, but %T", tree.Root)
	}

	// --- DEBUG ---
	if interp.internal.ExecPolicy != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] ExecWithInterpreter: BEFORE Load, ExecPolicy is PRESENT. Context: %v, Allow count: %d\n", interp.internal.ExecPolicy.Context, len(interp.internal.ExecPolicy.Allow))
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG] ExecWithInterpreter: BEFORE Load, ExecPolicy is NIL.\n")
	}
	// --- END DEBUG ---

	// 1. Load the program by wrapping the ast.Program in an interfaces.Tree.
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		return nil, fmt.Errorf("failed to load program into interpreter: %w", err)
	}

	// --- DEBUG ---
	if interp.internal.ExecPolicy != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] ExecWithInterpreter: AFTER Load, ExecPolicy is PRESENT. Context: %v, Allow count: %d\n", interp.internal.ExecPolicy.Context, len(interp.internal.ExecPolicy.Allow))
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG] ExecWithInterpreter: AFTER Load, ExecPolicy is NIL.\n")
	}
	// --- END DEBUG ---

	// 2. Execute top-level command blocks.
	finalValue, err := interp.Execute(tree)
	if err != nil {
		return nil, err
	}

	return finalValue, nil
}

// ExecScript is a convenience helper that wraps ExecInNewInterpreter.
// It builds a minimal HostContext to fulfill the API contract.
func ExecScript(ctx context.Context, script string, stdout io.Writer) (Value, error) {
	if stdout == nil {
		stdout = io.Discard
	}

	hc, err := NewHostContextBuilder().
		WithLogger(logging.NewNoOpLogger()).
		WithStdout(stdout).
		WithStdin(os.Stdin).
		WithStderr(io.Discard).
		Build()
	if err != nil {
		// This should not happen with the provided values, but check for safety.
		return nil, fmt.Errorf("ExecScript failed to build minimal HostContext: %w", err)
	}

	opts := []interpreter.InterpreterOption{
		WithHostContext(hc),
	}
	return ExecInNewInterpreter(ctx, script, opts...)
}

// LoadFromUnit loads definitions from a verified LoadedUnit.
func LoadFromUnit(interp *Interpreter, unit *LoadedUnit) error {
	if interp == nil || interp.internal == nil {
		return fmt.Errorf("LoadFromUnit requires a non-nil interpreter")
	}
	if unit == nil || unit.Tree == nil || unit.Tree.Root == nil {
		return fmt.Errorf("cannot load from a nil unit or tree")
	}
	program, ok := unit.Tree.Root.(*ast.Program)
	if !ok {
		return fmt.Errorf("internal error: loaded unit root is not a runnable *ast.Program, but %T", unit.Tree.Root)
	}
	return interp.Load(&interfaces.Tree{Root: program})
}

// RunProcedure executes a named procedure with Go-native arguments.
func RunProcedure(ctx context.Context, interp *Interpreter, name string, args ...any) (Value, error) {
	if interp == nil {
		return nil, fmt.Errorf("RunProcedure requires a non-nil interpreter")
	}
	// --- DEBUG ---
	if interp.internal.ExecPolicy != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] RunProcedure: BEFORE Run, ExecPolicy is PRESENT. Context: %v, Allow count: %d\n", interp.internal.ExecPolicy.Context, len(interp.internal.ExecPolicy.Allow))
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG] RunProcedure: BEFORE Run, ExecPolicy is NIL.\n")
	}
	// --- END DEBUG ---
	wrappedArgs := make([]lang.Value, len(args))
	for i, arg := range args {
		var err error
		wrappedArgs[i], err = lang.Wrap(arg)
		if err != nil {
			return nil, fmt.Errorf("error converting argument %d for procedure '%s': %w", i, name, err)
		}
	}
	return interp.Run(name, wrappedArgs...)
}
