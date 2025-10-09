// NeuroScript Version: 0.7.4
// File version: 23
// Purpose: FIX: RunProcedure now manually clones the interpreter facade before execution, ensuring the hostRuntime is correctly propagated to the sandboxed clone where tools are run.
// filename: pkg/api/exec.go
// nlines: 145
// risk_rating: HIGH

package api

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
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

// defensiveSetTurnContext prevents a non-nil but empty context (like context.Background())
// from overwriting a valid, inherited AEIOU context.
func defensiveSetTurnContext(interp *Interpreter, ctx context.Context) {
	if ctx == nil {
		return
	}
	// Only set the context if the new context actually contains an AEIOU session ID.
	if sid, ok := ctx.Value(aeiou.SessionIDKey).(string); ok && sid != "" {
		interp.SetTurnContext(ctx)
	}
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

	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		return nil, fmt.Errorf("failed to load program into interpreter: %w", err)
	}

	defensiveSetTurnContext(interp, ctx)

	finalValue, err := interp.ExecuteCommands()
	if err != nil {
		return nil, err
	}

	return finalValue, nil
}

// RunScriptWithInterpreter loads and executes a script using a persistent
// interpreter in a non-destructive way.
func RunScriptWithInterpreter(ctx context.Context, interp *Interpreter, tree *Tree) (Value, error) {
	if interp == nil || interp.internal == nil {
		return nil, fmt.Errorf("RunScriptWithInterpreter requires a non-nil interpreter")
	}
	if tree == nil || tree.Root == nil {
		return nil, fmt.Errorf("cannot execute a nil tree")
	}

	program, ok := tree.Root.(*ast.Program)
	if !ok {
		return nil, fmt.Errorf("internal error: tree root is not a runnable *ast.Program, but %T", tree.Root)
	}

	if err := interp.AppendScript(&interfaces.Tree{Root: program}); err != nil {
		return nil, fmt.Errorf("failed to append program to interpreter: %w", err)
	}

	defensiveSetTurnContext(interp, ctx)

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

	// --- CORE BUG FIX ---
	// The internal interpreter always runs procedures in a sandboxed clone.
	// To ensure the API-level 'hostRuntime' wrapper is preserved, we must
	// manually clone the facade here. This calls api.Interpreter.Clone(),
	// which correctly re-wraps the new internal clone's runtime.
	fmt.Fprintf(os.Stderr, "[DEBUG RunProcedure] Cloning facade before execution. Parent runtime is %T\n", interp.runtime)
	execClone := interp.Clone()

	if ctx != nil {
		sid, _ := ctx.Value(aeiou.SessionIDKey).(string)
		turn, _ := ctx.Value(aeiou.TurnIndexKey).(int)
		fmt.Fprintf(os.Stderr, "[DEBUG RunProcedure] Setting context on CLONED interpreter with SID: %q, Turn: %d\n", sid, turn)
		defensiveSetTurnContext(execClone, ctx)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG RunProcedure] Executing on cloned facade. Clone runtime is %T\n", execClone.runtime)

	wrappedArgs := make([]lang.Value, len(args))
	for i, arg := range args {
		var err error
		wrappedArgs[i], err = lang.Wrap(arg)
		if err != nil {
			return nil, fmt.Errorf("error converting argument %d for procedure '%s': %w", i, name, err)
		}
	}
	// Execute on the clone, which now has the correct runtime.
	return execClone.Run(name, wrappedArgs...)
}
