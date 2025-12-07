// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 34
// :: description: Fixes LoadOrExecute to use AppendScript for commands, preserving interpreter state.
// :: latestChange: LoadOrExecute now uses AppendScript+Execute for commands to prevent wiping definitions.
// :: filename: pkg/api/exec.go
// :: serialization: go

package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/api/analysis"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// CheckScriptTools verifies that all tools required by the script exist in the interpreter's registry.
func CheckScriptTools(tree *Tree, interp Runtime) error {
	if tree == nil || tree.Root == nil {
		return fmt.Errorf("cannot check tools on a nil tree")
	}
	if interp == nil {
		return fmt.Errorf("cannot check tools with a nil interpreter")
	}
	registry := interp.ToolRegistry()
	if registry == nil {
		return fmt.Errorf("interpreter does not have a tool registry")
	}

	program, ok := tree.Root.(*ast.Program)
	if !ok {
		return fmt.Errorf("internal error: tree root is not a checkable *ast.Program, but %T", tree.Root)
	}

	requiredTools := analysis.FindRequiredTools(&interfaces.Tree{Root: program})
	if len(requiredTools) == 0 {
		return nil
	}

	var missingTools []string
	for rawToolName := range requiredTools {
		if _, found := registry.GetTool(types.FullName(rawToolName)); !found {
			missingTools = append(missingTools, "tool."+rawToolName)
		}
	}

	if len(missingTools) > 0 {
		errMsg := fmt.Sprintf("script requires tools that are not registered: [%s]", strings.Join(missingTools, ", "))
		if interp.GetLogger() != nil {
			interp.GetLogger().Error("Tool check failed", "missing", missingTools)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", errMsg)
		}
		return lang.NewRuntimeError(lang.ErrorCodeToolNotFound, errMsg, lang.ErrToolNotFound)
	}

	return nil
}

// ExecInNewInterpreter provides a stateless, one-shot execution.
func ExecInNewInterpreter(ctx context.Context, src string, opts ...interpreter.InterpreterOption) (Value, error) {
	tree, err := Parse([]byte(src), ParseSkipComments)
	if err != nil {
		return nil, fmt.Errorf("parsing failed in ExecInNewInterpreter: %w", err)
	}

	interp := New(opts...)
	interp.SetTurnContext(ctx)

	return ExecWithInterpreter(ctx, interp, tree)
}

// ExecWithInterpreter executes top-level 'command' blocks using a persistent interpreter.
// NOTE: This function calls Load(), which replaces the interpreter's current program state.
// For accumulating state (REPL/Session), use AppendScript instead.
func ExecWithInterpreter(ctx context.Context, interp Runtime, tree *Tree) (Value, error) {
	concreteInterp, ok := interp.(*Interpreter)
	if !ok {
		return nil, fmt.Errorf("ExecWithInterpreter requires an *api.Interpreter, but got %T", interp)
	}
	if concreteInterp == nil {
		return nil, fmt.Errorf("ExecWithInterpreter requires a non-nil interpreter")
	}
	if tree == nil || tree.Root == nil {
		return nil, fmt.Errorf("cannot execute a nil tree")
	}
	program, ok := tree.Root.(*ast.Program)
	if !ok {
		return nil, fmt.Errorf("internal error: tree root is not a runnable *ast.Program, but %T", tree.Root)
	}

	concreteInterp.SetTurnContext(ctx)

	if err := concreteInterp.Load(&interfaces.Tree{Root: program}); err != nil {
		return nil, fmt.Errorf("failed to load program into interpreter: %w", err)
	}

	finalValue, err := concreteInterp.Execute(tree)
	if err != nil {
		return nil, err
	}

	return finalValue, nil
}

// ExecScript is a convenience helper that wraps ExecInNewInterpreter.
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
		return nil, fmt.Errorf("ExecScript failed to build minimal HostContext: %w", err)
	}
	opts := []interpreter.InterpreterOption{
		WithHostContext(hc),
	}
	return ExecInNewInterpreter(ctx, script, opts...)
}

// LoadFromUnit loads definitions from a verified LoadedUnit.
func LoadFromUnit(interp *Interpreter, unit *LoadedUnit) error {
	if interp == nil {
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
	interp.SetTurnContext(ctx)
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

// LoadOrExecute inspects the AST and routes execution based on the content:
//   - Mixed Definitions & Commands -> Returns error (ErrMixedScript).
//   - Definitions only -> Calls interp.AppendScript (Persistent load).
//   - Commands only -> Calls interp.AppendScript + Execute (Transient execution over Persistent state).
//   - Empty -> Returns nil, nil.
func LoadOrExecute(ctx context.Context, interp *Interpreter, tree *Tree) (Value, error) {
	if interp == nil {
		return nil, fmt.Errorf("LoadOrExecute requires a non-nil interpreter")
	}
	if tree == nil || tree.Root == nil {
		return nil, nil // Treat as empty
	}
	program, ok := tree.Root.(*ast.Program)
	if !ok {
		return nil, fmt.Errorf("internal error: tree root is not a runnable *ast.Program, but %T", tree.Root)
	}

	hasCmds := HasCommandBlock(program)
	hasDefs := HasDefinitions(program)

	// Rule: Mixed? -> Error ("Cannot mix").
	if hasCmds && hasDefs {
		return nil, fmt.Errorf("mixed script detected: scripts must contain either definitions OR commands, not both")
	}

	// Rule: Definitions? -> AppendScript (Persistent).
	if hasDefs {
		if err := interp.AppendScript(tree); err != nil {
			return nil, fmt.Errorf("failed to append definitions: %w", err)
		}
		return nil, nil
	}

	// Rule: Commands? -> Append then Execute (Transient action, Persistent state).
	if hasCmds {
		// We must NOT use ExecWithInterpreter here because it calls Load(),
		// which resets the interpreter's state (wiping previously loaded definitions).
		// Instead, we Append the commands (to register them in the current session)
		// and then Execute them specifically.

		interp.SetTurnContext(ctx)

		// 1. Append (to merge into current state without wiping)
		if err := interp.AppendScript(tree); err != nil {
			return nil, fmt.Errorf("failed to append command script: %w", err)
		}

		// 2. Execute just this tree
		return interp.Execute(tree)
	}

	// Neither (e.g., empty or comments only)
	return nil, nil
}

// ExecuteSandboxedAST runs a pre-parsed AST in a sandboxed fork
// of the provided interpreter. It captures and returns all
// 'emit' and 'whisper' I/O from the execution.
// This is the core callback for the AEIOU v2+ host loop.
func ExecuteSandboxedAST(
	interp *Interpreter,
	tree *interfaces.Tree,
	turnCtx context.Context,
) (
	emits []string,
	whispers map[string]lang.Value,
	execErr error,
) {
	if interp == nil {
		execErr = fmt.Errorf("ExecuteSandboxedAST: interpreter cannot be nil")
		return
	}
	if tree == nil || tree.Root == nil {
		execErr = fmt.Errorf("ExecuteSandboxedAST: tree cannot be nil")
		return
	}
	program, ok := tree.Root.(*ast.Program)
	if !ok {
		execErr = fmt.Errorf("ExecuteSandboxedAST: tree root is not a runnable *ast.Program, but %T", tree.Root)
		return
	}

	// 1. Get the internal "root" interpreter.
	rootInterp, ok := interp.Unwrap().(*interpreter.Interpreter)
	if !ok {
		execErr = fmt.Errorf("ExecuteSandboxedAST: could not unwrap internal interpreter")
		return
	}

	// 2. Create a sandboxed fork using the new exported method.
	// This fixes the 'undefined: fork' and 'undefined: ForkOptions' errors.
	execInterp, err := rootInterp.ForkSandboxed()
	if err != nil {
		execErr = fmt.Errorf("ExecuteSandboxedAST: failed to fork interpreter: %w", err)
		return
	}

	// 3. Set the ephemeral turn context.
	execInterp.SetTurnContext(turnCtx)

	// 4. Set up I/O capture for emits and whispers.
	emits = []string{}
	whispers = make(map[string]lang.Value)
	var emitBuf bytes.Buffer

	// Get a mutable copy of the fork's host context.
	turnHostContext := *execInterp.HostContext()
	turnHostContext.EmitFunc = func(v lang.Value) {
		fmt.Fprintln(&emitBuf, lang.Unwrap(v))
	}
	turnHostContext.WhisperFunc = func(handle, data lang.Value) {
		whispers[handle.String()] = data
	}
	execInterp.SetHostContext(&turnHostContext)

	// 5. Execute the AST's 'command' blocks.
	_, execErr = execInterp.Execute(program)

	// 6. Process captured emits.
	rawEmits := strings.TrimSpace(emitBuf.String())
	if rawEmits != "" {
		emits = strings.Split(rawEmits, "\n")
	}

	// 7. Return captured data and any execution error.
	return emits, whispers, execErr
}
