// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides common test helper functions for the shape toolset tests.
// filename: pkg/tool/shape/tools_shape_helpers_test.go
// nlines: 55
// risk_rating: LOW

package shape_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/google/go-cmp/cmp"
)

// testShapeToolHelper sets up a fresh interpreter with the shape tools registered.
func testShapeToolHelper(t *testing.T, testName string, testFunc func(t *testing.T, interp tool.Runtime)) {
	t.Run(testName, func(t *testing.T) {
		interp := interpreter.NewInterpreter(interpreter.WithLogger(logging.NewTestLogger(t)))
		if err := tool.RegisterGlobalToolsets(interp.ToolRegistry()); err != nil {
			t.Fatalf("Failed to register extended tools: %v", err)
		}
		testFunc(t, interp)
	})
}

// runTool executes a tool from the 'shape' group.
func runTool(t *testing.T, interp tool.Runtime, toolName types.ToolName, args ...interface{}) (interface{}, error) {
	t.Helper()
	fullName := types.MakeFullName("shape", string(toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullName)
	if !found {
		t.Fatalf("Tool %q not found in registry", fullName)
	}
	return toolImpl.Func(interp, args)
}

// assertResult checks for expected errors and result values.
func assertResult(t *testing.T, result interface{}, err error, expectedResult interface{}, expectedErr error) {
	t.Helper()
	if expectedErr != nil {
		if err == nil {
			t.Fatalf("expected error '%v', but got nil", expectedErr)
		}
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected error to wrap '%v', but got: %v", expectedErr, err)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff := cmp.Diff(expectedResult, result); diff != "" {
		t.Errorf("result does not match expected (-want +got):\n%s", diff)
	}
}
