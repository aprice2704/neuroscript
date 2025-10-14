// NeuroScript Version: 0.8.0
// File version: 25
// Purpose: Corrected a compiler error by changing the call to lang.Unwrap to handle its single return value.
// filename: pkg/testutil/testing_helpers.go

package testutil

import (
	"fmt"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/types"
)

var dummyPos = &types.Position{Line: 1, Column: 1, File: "test"}

// --- Generic AST Creation Helpers (Exported) ---

func NewTestStringLiteral(val string) *ast.StringLiteralNode {
	return &ast.StringLiteralNode{Value: val}
}

func NewTestNumberLiteral(val float64) *ast.NumberLiteralNode {
	return &ast.NumberLiteralNode{Value: val}
}

func NewTestBooleanLiteral(val bool) *ast.BooleanLiteralNode {
	return &ast.BooleanLiteralNode{Value: val}
}

func NewVariableNode(name string) *ast.VariableNode {
	return &ast.VariableNode{Name: name}
}

// --- Test Execution Helpers (Exported) ---

// NewTestInterpreter creates a fully configured interpreter instance for testing,
// using the new public API. It initializes it with a HostContext, a sandbox
// directory, and a map of initial global variables.
func NewTestInterpreter(t *testing.T, initialVars map[string]lang.Value) (*api.Interpreter, error) {
	t.Helper()
	testLogger := logging.NewTestLogger(t)

	// Build the mandatory HostContext.
	hc, err := api.NewHostContextBuilder().
		WithLogger(testLogger).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithStdin(os.Stdin).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build HostContext: %w", err)
	}

	// Convert lang.Value to interface{} for WithGlobals.
	globals := make(map[string]interface{})
	if initialVars != nil {
		for k, v := range initialVars {
			// This unwrap is safe for test setup with known types.
			// FIX: lang.Unwrap now returns a single value.
			unwrapped := lang.Unwrap(v)
			globals[k] = unwrapped
		}
	}

	// Instantiate the interpreter with all options.
	interp := api.New(
		api.WithHostContext(hc),
		NewTestSandbox(t), // Creates and cleans up a temp sandbox dir.
		api.WithGlobals(globals),
	)

	return interp, nil
}
