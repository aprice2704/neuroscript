// NeuroScript Version: 0.7.4
// File version: 3
// Purpose: Adds a test to ensure a custom tool.Runtime is propagated to clones.
// filename: pkg/interpreter/interpreter_clone_test.go
// nlines: 115
// risk_rating: LOW
package interpreter_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// mockRuntimeWithIdentity is a simple mock to verify propagation.
type mockRuntimeWithIdentity struct {
	*interpreter.Interpreter // Embed interpreter to satisfy the interface
	identity                 string
}

func TestInterpreter_Clone_CapsuleStore(t *testing.T) {
	// 1. Create a custom registry and add it to a parent interpreter.
	customRegistry := capsule.NewRegistry()
	customCapsule := capsule.Capsule{
		Name:    "capsule/clone-test",
		Version: "1",
		Content: "Content for clone test",
	}
	customRegistry.MustRegister(customCapsule)

	parent, err := interpreter.NewTestInterpreter(t, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to create parent interpreter: %v", err)
	}
	// Add the custom registry to the parent's store
	parent.CapsuleStore().Add(customRegistry)

	// 2. Clone the parent interpreter.
	clone := parent.Clone()

	// 3. Check if the clone has access to the custom capsule store.
	if clone.CapsuleStore() == nil {
		t.Fatal("Cloned interpreter has a nil capsuleStore.")
	}

	// 4. Try to retrieve the custom capsule through the clone.
	retrieved, found := clone.CapsuleStore().GetLatest("capsule/clone-test")
	if !found {
		t.Fatal("Custom capsule not found in cloned interpreter's store.")
	}

	// 5. Verify the retrieved capsule is correct.
	if retrieved.ID != "capsule/clone-test@1" {
		t.Errorf("Expected capsule ID 'capsule/clone-test@1', but got '%s'", retrieved.ID)
	}
	if retrieved.Content != "Content for clone test" {
		t.Errorf("Retrieved capsule content mismatch.")
	}
}

func TestInterpreter_Clone_CustomFuncs(t *testing.T) {
	parent, err := interpreter.NewTestInterpreter(t, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to create parent interpreter: %v", err)
	}

	var emitCaptured bool
	var whisperCaptured bool

	// Set custom functions on the PARENT
	parent.SetEmitFunc(func(v lang.Value) {
		emitCaptured = true
	})
	parent.SetWhisperFunc(func(h, d lang.Value) {
		whisperCaptured = true
	})

	// Create the clone
	clone := parent.Clone()

	// Execute a script in the CLONE that uses emit and whisper
	script := `
	func main() means
		emit "hello"
		whisper "self", "data"
	endfunc
	`
	_, execErr := clone.ExecuteScriptString("main", script, nil)
	if execErr != nil {
		t.Fatalf("Script execution in clone failed: %v", execErr)
	}

	// Assert that the custom functions from the PARENT were called
	if !emitCaptured {
		t.Error("customEmitFunc was not propagated to the clone")
	}
	if !whisperCaptured {
		t.Error("customWhisperFunc was not propagated to the clone")
	}
}

func TestInterpreter_Clone_CustomRuntime(t *testing.T) {
	parent, err := interpreter.NewTestInterpreter(t, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create parent interpreter: %v", err)
	}

	// 1. Create and set a custom runtime on the parent.
	customRT := &mockRuntimeWithIdentity{
		Interpreter: parent,
		identity:    "custom_runtime_context",
	}
	parent.SetRuntime(customRT)

	// 2. Register a probe tool that checks the runtime's identity.
	probeTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{Name: "probe", Group: "test"},
		Func: func(rt tool.Runtime, args []any) (any, error) {
			if mock, ok := rt.(*mockRuntimeWithIdentity); ok {
				if mock.identity == "custom_runtime_context" {
					return "success", nil
				}
			}
			return nil, errors.New("runtime was not the expected custom one")
		},
	}
	_, _ = parent.ToolRegistry().RegisterTool(probeTool)

	// 3. Execute a script that calls the probe from within a procedure (which forces a clone).
	script := `func main() means 
	must tool.test.probe() 
	endfunc`
	_, execErr := parent.ExecuteScriptString("main", script, nil)

	// 4. Assert that the execution succeeded, meaning the clone received the custom runtime.
	if execErr != nil {
		t.Fatalf("Script execution failed, indicating custom runtime was not propagated to clone: %v", execErr)
	}
}
