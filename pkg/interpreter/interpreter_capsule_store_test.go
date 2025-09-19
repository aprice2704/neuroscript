// NeuroScript Version: 0.7.1
// File version: 2
// Purpose: Corrected capsule version string in test to be an integer, fixing the panic.
// filename: pkg/interpreter/interpreter_capsule_store_test.go
// nlines: 75
// risk_rating: LOW
package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
)

func TestInterpreter_CapsuleStoreLayering(t *testing.T) {
	t.Run("Adding and retrieving from a custom registry", func(t *testing.T) {
		// 1. Create a custom registry with a unique capsule.
		customRegistry := capsule.NewRegistry()
		customCapsule := capsule.Capsule{
			Name:    "capsule/custom",
			Version: "1", // FIX: Version must be an integer string.
			Content: "This is a custom capsule.",
		}
		customRegistry.MustRegister(customCapsule)

		// 2. Create an interpreter, adding the custom registry via the option.
		interp := interpreter.NewInterpreter(
			interpreter.WithCapsuleRegistry(customRegistry),
		)

		// 3. Retrieve the custom capsule from the interpreter's store.
		retrieved, found := interp.CapsuleStore().GetLatest("capsule/custom")
		if !found {
			t.Fatal("Expected to find the custom capsule, but it was not found.")
		}

		// 4. Assert that the correct capsule was retrieved.
		if retrieved.ID != "capsule/custom@1" {
			t.Errorf("Retrieved capsule ID mismatch. Got: %s, Want: %s", retrieved.ID, "capsule/custom@1")
		}
		if retrieved.Content != "This is a custom capsule." {
			t.Errorf("Retrieved capsule content mismatch.")
		}
	})

	t.Run("Default registry is not overridden by later registries", func(t *testing.T) {
		// 1. Create a custom registry with a conflicting capsule name but a higher version.
		customRegistry := capsule.NewRegistry()
		overridingCapsule := capsule.Capsule{
			Name:    "capsule/aeiou", // Same name as a default capsule
			Version: "999",           // Higher version
			Content: "This should NOT be found.",
		}
		customRegistry.MustRegister(overridingCapsule)

		// 2. Create an interpreter, adding the custom registry.
		// The default registry is added first, then the custom one.
		interp := interpreter.NewInterpreter(
			interpreter.WithCapsuleRegistry(customRegistry),
		)

		// 3. Retrieve the latest version of the conflicting capsule.
		// The store should search the default registry first, find the name, and stop.
		retrieved, found := interp.CapsuleStore().GetLatest("capsule/aeiou")
		if !found {
			t.Fatal("Expected to find the default capsule, but it was not found.")
		}

		// 4. Assert that the version from the *default* registry was returned, not the higher one.
		// The latest version in the default registry for 'capsule/aeiou' is "2".
		const expectedDefaultVersion = "2"
		if retrieved.Version != expectedDefaultVersion {
			t.Errorf("Capsule was overridden. Got version: %s, Want version: %s (from default registry)", retrieved.Version, expectedDefaultVersion)
		}
		if retrieved.Content == "This should NOT be found." {
			t.Error("Capsule content was overridden by the later registry.")
		}
	})
}
