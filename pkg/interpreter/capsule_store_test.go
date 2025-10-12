// NeuroScript Version: 0.7.1
// File version: 3
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
// filename: pkg/interpreter/interpreter_capsule_store_test.go
// nlines: 80
// risk_rating: LOW
package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
)

func TestInterpreter_CapsuleStoreLayering(t *testing.T) {
	t.Run("Adding and retrieving from a custom registry", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'Adding and retrieving from a custom registry' test.")
		h := NewTestHarness(t)
		customRegistry := capsule.NewRegistry()
		customCapsule := capsule.Capsule{
			Name:    "capsule/custom",
			Version: "1",
			Content: "This is a custom capsule.",
		}
		customRegistry.MustRegister(customCapsule)
		t.Logf("[DEBUG] Turn 2: Custom registry created.")

		// Create a new interpreter with the custom registry, but reuse the harness's HostContext.
		interp := interpreter.NewInterpreter(
			interpreter.WithHostContext(h.HostContext),
			interpreter.WithCapsuleRegistry(customRegistry),
		)
		t.Logf("[DEBUG] Turn 3: New interpreter created with custom capsule registry.")

		retrieved, found := interp.CapsuleStore().GetLatest("capsule/custom")
		if !found {
			t.Fatal("Expected to find the custom capsule, but it was not found.")
		}
		t.Logf("[DEBUG] Turn 4: Custom capsule retrieved.")

		if retrieved.ID != "capsule/custom@1" {
			t.Errorf("Retrieved capsule ID mismatch. Got: %s, Want: %s", retrieved.ID, "capsule/custom@1")
		}
		if retrieved.Content != "This is a custom capsule." {
			t.Errorf("Retrieved capsule content mismatch.")
		}
		t.Logf("[DEBUG] Turn 5: Assertions passed.")
	})

	t.Run("Default registry is not overridden by later registries", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'Default registry is not overridden' test.")
		h := NewTestHarness(t)
		customRegistry := capsule.NewRegistry()
		overridingCapsule := capsule.Capsule{
			Name:    "capsule/aeiou",
			Version: "999",
			Content: "This should NOT be found.",
		}
		customRegistry.MustRegister(overridingCapsule)
		t.Logf("[DEBUG] Turn 2: Custom registry with overriding capsule created.")

		interp := interpreter.NewInterpreter(
			interpreter.WithHostContext(h.HostContext),
			interpreter.WithCapsuleRegistry(customRegistry),
		)
		t.Logf("[DEBUG] Turn 3: New interpreter created.")

		retrieved, found := interp.CapsuleStore().GetLatest("capsule/aeiou")
		if !found {
			t.Fatal("Expected to find the default capsule, but it was not found.")
		}
		t.Logf("[DEBUG] Turn 4: Default capsule retrieved.")

		const expectedDefaultVersion = "2"
		if retrieved.Version != expectedDefaultVersion {
			t.Errorf("Capsule was overridden. Got version: %s, Want version: %s", retrieved.Version, expectedDefaultVersion)
		}
		if retrieved.Content == "This should NOT be found." {
			t.Error("Capsule content was overridden by the later registry.")
		}
		t.Logf("[DEBUG] Turn 5: Assertions passed.")
	})
}
