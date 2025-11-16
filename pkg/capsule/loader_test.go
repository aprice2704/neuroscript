// NeuroScript Version: 0.7.2
// File version: 5
// Purpose: Tests that the multi-format capsule loader correctly registers content from both .md and .ns files.
// Latest change: Updated to use DefaultStore() and check for Description field.
// filename: pkg/capsule/loader_test.go
// nlines: 51
package capsule_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
)

func TestCapsuleLoader_MultiFormat(t *testing.T) {
	// The loader runs on init, populating the BuiltInRegistry.
	// We test against the DefaultStore, which consumes that registry.
	store := capsule.DefaultStore()

	t.Run("Loads Markdown Capsule", func(t *testing.T) {
		const expectedName = "capsule/aeiou"
		c, ok := store.GetLatest(expectedName)

		if !ok {
			t.Fatalf("store.GetLatest(%q) failed: markdown capsule was not loaded", expectedName)
		}
		if c.Name != expectedName {
			t.Errorf("capsule Name mismatch: got %q, want %q", c.Name, expectedName)
		}
		if c.Version == "" {
			t.Errorf("capsule Version is empty after loading")
		}
		if c.Content == "" {
			t.Error("capsule Content is empty after loading")
		}
		if c.Description == "" {
			t.Error("capsule Description is empty after loading")
		}
	})

	t.Run("Loads NeuroScript Capsule", func(t *testing.T) {
		const expectedName = "capsule/ns-example"
		c, ok := store.Get(expectedName, "1") // Get specific version

		if !ok {
			t.Fatalf("store.Get(%q, '1') failed: neuroscript capsule was not loaded", expectedName)
		}
		if c.Name != expectedName {
			t.Errorf("capsule Name mismatch: got %q, want %q", c.Name, expectedName)
		}
		if c.Priority != 50 {
			t.Errorf("capsule Priority mismatch: got %d, want 50", c.Priority)
		}
		if c.Description == "" {
			t.Error("capsule Description is empty after loading")
		}
		if !strings.Contains(c.Content, `emit "hello from an .ns capsule"`) {
			t.Error("neuroscript capsule content seems incorrect")
		}
	})
}
