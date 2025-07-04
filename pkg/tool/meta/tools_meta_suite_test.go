// filename: pkg/tool/meta/tools_meta_suite_test.go
package meta

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- Local Test Setup ---

// newMetaTestInterpreter sets up an interpreter instance specifically for meta tool testing.
// It registers the meta tools so they can be executed.
func newMetaTestInterpreter(t *testing.T) (*interpreter.Interpreter, error) {
	t.Helper()

	interp := interpreter.NewInterpreter(
		interpreter.WithLogger(logging.NewTestLogger(t)),
	)

	// Manually register the meta tools for this test suite.
	// This ensures our tests are isolated and only depend on what's explicitly registered.
	for _, toolImpl := range metaToolsToRegister {
		if err := interp.RegisterTool(toolImpl); err != nil {
			return nil, fmt.Errorf("failed to register tool '%s': %w", toolImpl.Spec.Name, err)
		}
	}

	// Register a few other dummy tools to test filtering.
	dummySpec := tool.ToolSpec{Name: "FS.Read", Description: "Dummy FS tool."}
	dummyFunc := func(rt tool.Runtime, args []interface{}) (interface{}, error) { return "dummy fs read", nil }
	interp.RegisterTool(tool.ToolImplementation{Spec: dummySpec, Func: dummyFunc})

	return interp, nil
}
