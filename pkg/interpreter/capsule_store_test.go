// NeuroScript Version: 0.7.1
// File version: 5
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
// Latest change: Updated to use WithCapsuleStore and prove that overriding built-ins now works.
// filename: pkg/interpreter/interpreter_capsule_store_test.go
// nlines: 87
// risk_rating: LOW
package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
)

func TestInterpreter_CapsuleStoreLayering(t *testing.T) {
	t.Run("Adding and retrieving from a custom store", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'Adding and retrieving from a custom store' test.")
		h := NewTestHarness(t)
		customRegistry := capsule.NewRegistry()
		customCapsule := capsule.Capsule{
			Name:        "capsule/custom",
			Version:     "1",
			Content:     "This is a custom capsule.",
			Description: "A custom test capsule.",
		}
		customRegistry.MustRegister(customCapsule)
		t.Logf("[DEBUG] Turn 2: Custom registry created.")

		// --- THE FIX ---
		// Create a new store containing only the custom registry
		customStore := capsule.NewStore(customRegistry)

		// Create a new interpreter with the custom store.
		interp := interpreter.NewInterpreter(
			interpreter.WithHostContext(h.HostContext),
			interpreter.WithCapsuleStore(customStore), // Use the new option
		)
		// --- END FIX ---
		t.Logf("[DEBUG] Turn 3: New interpreter created with custom capsule store.")

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

	t.Run("Injected store with custom layering overrides built-in registry", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'Injected store overrides built-in' test.")
		h := NewTestHarness(t)
		customRegistry := capsule.NewRegistry()
		overridingCapsule := capsule.Capsule{
			Name:        "capsule/aeiou", // This name collides with a built-in
			Version:     "999",
			Content:     "This should BE found.",
			Description: "An overriding capsule.",
		}
		customRegistry.MustRegister(overridingCapsule)
		t.Logf("[DEBUG] Turn 2: Custom registry with overriding capsule created.")

		// --- THE FIX ---
		// Create a new store, layering the custom registry *FIRST*
		// This ensures it is searched first and wins the lookup.
		layeredStore := capsule.NewStore(customRegistry, capsule.BuiltInRegistry())

		interp := interpreter.NewInterpreter(
			interpreter.WithHostContext(h.HostContext),
			interpreter.WithCapsuleStore(layeredStore), // Use the new option
		)
		// --- END FIX ---
		t.Logf("[DEBUG] Turn 3: New interpreter created with custom-layered store.")

		retrieved, found := interp.CapsuleStore().GetLatest("capsule/aeiou")
		if !found {
			t.Fatal("Expected to find the capsule, but it was not found.")
		}
		t.Logf("[DEBUG] Turn 4: 'capsule/aeiou' retrieved.")

		// --- ASSERTIONS REVERSED ---
		const expectedOverriddenVersion = "999"
		if retrieved.Version != expectedOverriddenVersion {
			t.Errorf("Capsule was NOT overridden. Got version: %s, Want version: %s", retrieved.Version, expectedOverriddenVersion)
		}
		if retrieved.Content != "This should BE found." {
			t.Error("Capsule content was not overridden by the later registry.")
		}
		t.Logf("[DEBUG] Turn 5: Assertions passed. Overriding store works.")
		// --- END REVERSAL ---
	})
}
