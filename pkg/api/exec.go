// NeuroScript Version: 0.7.2
// File version: 14
// Purpose: Corrected Load calls to wrap the *ast.Program in an *interfaces.Tree, conforming to the dependency-injected API.
// filename: pkg/api/exec.go
// nlines: 98
// risk_rating: HIGH

package api

import (
	"context"
	"fmt"
	"io"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
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

	// 1. Load the program by wrapping the ast.Program in an interfaces.Tree.
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		return nil, fmt.Errorf("failed to load program into interpreter: %w", err)
	}

	// --- DEBUG ---
	// This is the final check. Does the *internal* interpreter have the admin
	// registry right before we tell it to execute the command?
	if interp.internal.CapsuleRegistryForAdmin() != nil {
		fmt.Println("[DEBUG] ExecWithInterpreter: Admin registry is PRESENT on internal interpreter before ExecuteCommands.")
	} else {
		fmt.Println("[DEBUG] ExecWithInterpreter: Admin registry is NIL on internal interpreter before ExecuteCommands.")
	}
	// --- END DEBUG ---

	// 2. Execute top-level command blocks.
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
