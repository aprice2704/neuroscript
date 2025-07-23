// NeuroScript Version: 0.6.0
// File version: 9
// Purpose: Adds high-level functions for loading from a unit and running specific procedures with Go-native arguments.
// filename: pkg/api/exec.go
// nlines: 88
// risk_rating: HIGH

package api

import (
	"context"
	"fmt"
	"io"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// ExecInNewInterpreter provides a stateless, one-shot execution. It will parse,
// load, and ONLY execute top-level 'command' blocks. It does NOT automatically
// run any function, including 'main'.
func ExecInNewInterpreter(ctx context.Context, src string, opts ...interpreter.InterpreterOption) (Value, error) {
	tree, err := Parse([]byte(src), ParseSkipComments)
	if err != nil {
		return nil, fmt.Errorf("parsing failed in ExecInNewInterpreter: %w", err)
	}

	interp := New(opts...)
	return ExecWithInterpreter(ctx, interp, tree)
}

// ExecWithInterpreter executes any top-level 'command' blocks from a given Tree
// using a persistent interpreter. It does NOT automatically run any function.
// Its primary role in a library context is to load the program.
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

	// 1. Load the program. This registers all function definitions.
	if err := interp.Load(program); err != nil {
		return nil, fmt.Errorf("failed to load program into interpreter: %w", err)
	}

	// 2. ONLY execute top-level command blocks.
	// If there are no command blocks, this is a no-op, which is the
	// correct behavior for loading a library of functions.
	finalValue, err := interp.ExecuteCommands()
	if err != nil {
		return nil, err
	}

	return finalValue, nil
}

// ExecScript is a convenience helper that wraps ExecInNewInterpreter.
func ExecScript(ctx context.Context, script string, stdout io.Writer) (Value, error) {
	opts := []interpreter.InterpreterOption{}
	if stdout != nil {
		opts = append(opts, interpreter.WithStdout(stdout))
	}
	return ExecInNewInterpreter(ctx, script, opts...)
}

// LoadFromUnit loads the definitions from a verified LoadedUnit into an interpreter instance.
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
	return interp.Load(program)
}

// RunProcedure executes a named procedure with the given Go-native arguments.
func RunProcedure(ctx context.Context, interp *Interpreter, name string, args ...any) (Value, error) {
	if interp == nil {
		return nil, fmt.Errorf("RunProcedure requires a non-nil interpreter")
	}
	wrappedArgs := make([]lang.Value, len(args))
	for i, arg := range args {
		var err error
		// FIX: Corrected function call from NewValueFromGo to Wrap
		wrappedArgs[i], err = lang.Wrap(arg)
		if err != nil {
			return nil, fmt.Errorf("error converting argument %d for procedure '%s': %w", i, name, err)
		}
	}
	return interp.Run(name, wrappedArgs...)
}
