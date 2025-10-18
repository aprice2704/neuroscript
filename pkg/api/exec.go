// NeuroScript Version: 0.8.0
// File version: 28
// Purpose: Extracted pre-flight tool check into a separate CheckScriptTools function. Removed automatic checks from Load/Exec.
// filename: pkg/api/exec.go
// nlines: 139
// risk_rating: HIGH

package api

import (
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
)

// CheckScriptTools verifies that all tools required by the script exist in the interpreter's registry.
// This function should be called by the host before LoadFromUnit or ExecWithInterpreter
// if pre-flight validation is desired.
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
		// Or handle gracefully if non-program roots are possible but cannot contain tools
		return fmt.Errorf("internal error: tree root is not a checkable *ast.Program, but %T", tree.Root)
	}

	requiredTools := analysis.FindRequiredTools(&interfaces.Tree{Root: program}) // Pass interfaces.Tree
	if len(requiredTools) == 0 {
		return nil // No tools required
	}

	availableTools := make(map[string]struct{})
	for _, impl := range registry.ListTools() {
		availableTools[string(impl.FullName)] = struct{}{}
	}

	var missingTools []string
	for toolName := range requiredTools {
		if _, found := availableTools[toolName]; !found {
			missingTools = append(missingTools, toolName)
		}
	}

	if len(missingTools) > 0 {
		errMsg := fmt.Sprintf("script requires tools that are not registered: [%s]", strings.Join(missingTools, ", "))
		// Log this loudly as well
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
	interp.SetTurnContext(ctx) // Set ephemeral context

	// Host *could* call CheckScriptTools(tree, interp) here if desired.

	return ExecWithInterpreter(ctx, interp, tree)
}

// ExecWithInterpreter executes top-level 'command' blocks using a persistent interpreter.
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

	// --- PRE-FLIGHT TOOL CHECK REMOVED ---
	// Host should call CheckScriptTools(tree, interp) before this if needed.

	concreteInterp.SetTurnContext(ctx) // Set ephemeral context

	if err := concreteInterp.Load(&interfaces.Tree{Root: program}); err != nil {
		return nil, fmt.Errorf("failed to load program into interpreter: %w", err)
	}

	finalValue, err := concreteInterp.Execute(tree)
	if err != nil {
		return nil, err // Propagate runtime errors (including tool not found during actual call)
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
	// Note: ExecInNewInterpreter is used, so the check isn't automatic there either.
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

	// --- PRE-FLIGHT TOOL CHECK REMOVED ---
	// Host should call CheckScriptTools(unit.Tree, interp) before this if needed.

	return interp.Load(&interfaces.Tree{Root: program}) // Load definitions
}

// RunProcedure executes a named procedure with Go-native arguments.
func RunProcedure(ctx context.Context, interp *Interpreter, name string, args ...any) (Value, error) {
	if interp == nil {
		return nil, fmt.Errorf("RunProcedure requires a non-nil interpreter")
	}

	interp.SetTurnContext(ctx) // Set ephemeral context

	wrappedArgs := make([]lang.Value, len(args))
	for i, arg := range args {
		var err error
		wrappedArgs[i], err = lang.Wrap(arg)
		if err != nil {
			return nil, fmt.Errorf("error converting argument %d for procedure '%s': %w", i, name, err)
		}
	}
	// Tool check is not performed here; assumed to be done during loading.
	return interp.Run(name, wrappedArgs...)
}
