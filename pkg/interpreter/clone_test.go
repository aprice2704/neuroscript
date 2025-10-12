// NeuroScript Version: 0.7.1
// File version: 3
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
// filename: pkg/interpreter/interpreter_clone_test.go
// nlines: 80
// risk_rating: LOW
package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestInterpreter_Clone_CapsuleStore(t *testing.T) {
	t.Logf("[DEBUG] Turn 1: Starting TestInterpreter_Clone_CapsuleStore.")
	h := NewTestHarness(t)
	parent := h.Interpreter

	customRegistry := capsule.NewRegistry()
	customCapsule := capsule.Capsule{
		Name:    "capsule/clone-test",
		Version: "1",
		Content: "Content for clone test",
	}
	customRegistry.MustRegister(customCapsule)
	parent.CapsuleStore().Add(customRegistry)
	t.Logf("[DEBUG] Turn 2: Parent interpreter configured with custom capsule registry.")

	clone := parent.Clone()
	t.Logf("[DEBUG] Turn 3: Interpreter cloned.")

	if clone.CapsuleStore() == nil {
		t.Fatal("Cloned interpreter has a nil capsuleStore.")
	}

	retrieved, found := clone.CapsuleStore().GetLatest("capsule/clone-test")
	if !found {
		t.Fatal("Custom capsule not found in cloned interpreter's store.")
	}
	t.Logf("[DEBUG] Turn 4: Custom capsule retrieved from clone.")

	if retrieved.ID != "capsule/clone-test@1" {
		t.Errorf("Expected capsule ID 'capsule/clone-test@1', but got '%s'", retrieved.ID)
	}
	if retrieved.Content != "Content for clone test" {
		t.Errorf("Retrieved capsule content mismatch.")
	}
	t.Logf("[DEBUG] Turn 5: Test completed successfully.")
}

func TestInterpreter_Clone_CustomFuncs(t *testing.T) {
	t.Logf("[DEBUG] Turn 1: Starting TestInterpreter_Clone_CustomFuncs.")
	h := NewTestHarness(t)
	parent := h.Interpreter
	clone := parent.Clone() // Clone before setting funcs to test propagation

	var emitCaptured bool
	var whisperCaptured bool

	h.HostContext.EmitFunc = func(v lang.Value) {
		emitCaptured = true
	}
	h.HostContext.WhisperFunc = func(h, d lang.Value) {
		whisperCaptured = true
	}
	t.Logf("[DEBUG] Turn 2: Custom I/O funcs set on HostContext.")

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
	t.Logf("[DEBUG] Turn 3: Script executed in clone.")

	if !emitCaptured {
		t.Error("customEmitFunc was not propagated to the clone")
	}
	if !whisperCaptured {
		t.Error("customWhisperFunc was not propagated to the clone")
	}
	t.Logf("[DEBUG] Turn 4: Test completed successfully.")
}
