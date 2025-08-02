// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Provides internal helper functions for creating interpreters in isolated test environments.
// filename: pkg/testing/testhelpers/helpers.go
// nlines: 28
// risk_rating: LOW

package testhelpers

import (
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// NewInterpreterForToolTest creates a bare-bones interpreter with a specific set
// of tools registered for isolated testing. It does NOT register the standard library,
// which prevents import cycles from tool tests back to the api package.
func NewInterpreterForToolTest(tools ...tool.ToolImplementation) *interpreter.Interpreter {
	// Create an option to register only the specified tools for this test.
	withTestTools := func(i *interpreter.Interpreter) {
		registry := i.ToolRegistry()
		for _, t := range tools {
			if _, err := registry.RegisterTool(t); err != nil {
				// Panicking is acceptable in a test helper if setup fails.
				panic(err)
			}
		}
	}

	// Create an interpreter that skips the standard library and uses only our test tools.
	return interpreter.NewInterpreter(interpreter.WithoutStandardTools(), withTestTools)
}
