// NeuroScript Version: 0.7.4
// File version: 1
// Purpose: Tests the SetRuntime method and custom runtime context injection.
// filename: pkg/interpreter/interpreter_runtime_test.go
// nlines: 80
// risk_rating: LOW

package interpreter_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// mockRuntime is a custom runtime for testing context injection.
// It embeds the base interpreter to satisfy the interface while adding a custom field.
type mockRuntime struct {
	*interpreter.Interpreter
	customField string
}

// probeTool is a tool designed to check if it's running within our custom runtime context.
var probeTool = tool.ToolImplementation{
	Spec: tool.ToolSpec{
		Name:  "probeRuntime",
		Group: "test",
	},
	Func: func(rt tool.Runtime, args []any) (any, error) {
		// Try to type-assert the provided runtime to our custom mock type.
		if mock, ok := rt.(*mockRuntime); ok {
			// If successful, check for the custom field's value.
			if mock.customField == "success" {
				return "probe_ok", nil
			}
			return nil, errors.New("custom field had wrong value")
		}
		// If the type assertion fails, it means the wrong runtime was passed.
		return nil, errors.New("runtime was not the expected mockRuntime type")
	},
}

// TestInterpreter_SetRuntime verifies that a host can inject a custom runtime
// context and that tools executed by the interpreter receive it correctly.
func TestInterpreter_SetRuntime(t *testing.T) {
	// 1. Create a standard interpreter. By default, it is its own runtime.
	interp, err := interpreter.NewTestInterpreter(t, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}

	// 2. Create a custom runtime wrapper.
	// This simulates a host application (like zadeh) adding its own context.
	customRuntime := &mockRuntime{
		Interpreter: interp,
		customField: "success",
	}

	// 3. Use the new API to inject the custom runtime into the interpreter.
	interp.SetRuntime(customRuntime)

	// 4. Register the probe tool that will verify the runtime context.
	if _, err := interp.ToolRegistry().RegisterTool(probeTool); err != nil {
		t.Fatalf("Failed to register probe tool: %v", err)
	}

	// 5. Execute a script that calls the probe tool.
	// The interpreter's internal logic should now pass `customRuntime` to the tool, not `interp`.
	script := `func main(returns result) means
		set result = tool.test.probeRuntime()
		return result
	endfunc
	`
	result, execErr := interp.ExecuteScriptString("main", script, nil)

	// 6. Assert that the script executed successfully.
	// A failure here means the tool's function received the wrong runtime and returned an error.
	if execErr != nil {
		t.Fatalf("Script execution failed, indicating the custom runtime was not passed to the tool: %v", execErr)
	}

	// 7. Assert that the return value is correct, confirming the tool ran as expected.
	if result.String() != "probe_ok" {
		t.Errorf("Expected result 'probe_ok', but got '%s'", result.String())
	}
}
