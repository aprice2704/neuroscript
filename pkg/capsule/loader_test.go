// NeuroScript Version: 0.7.1
// File version: 4
// Purpose: Tests that the multi-format capsule loader correctly registers content from both .md and .ns files.
// filename: pkg/capsule/loader_test.go
// nlines: 45
// risk_rating: LOW
package capsule_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
)

func TestCapsuleLoader_MultiFormat(t *testing.T) {
	// The loader runs on init, populating the default registry.
	reg := capsule.DefaultRegistry()

	t.Run("Loads Markdown Capsule", func(t *testing.T) {
		const expectedName = "capsule/aeiou"
		c, ok := reg.GetLatest(expectedName)

		if !ok {
			t.Fatalf("reg.GetLatest(%q) failed: markdown capsule was not loaded", expectedName)
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
	})

	t.Run("Loads NeuroScript Capsule", func(t *testing.T) {
		const expectedName = "capsule/ns-example"
		c, ok := reg.Get(expectedName, "1") // Get specific version

		if !ok {
			t.Fatalf("reg.Get(%q, '1') failed: neuroscript capsule was not loaded", expectedName)
		}
		if c.Name != expectedName {
			t.Errorf("capsule Name mismatch: got %q, want %q", c.Name, expectedName)
		}
		if c.Priority != 50 {
			t.Errorf("capsule Priority mismatch: got %d, want 50", c.Priority)
		}
		if !strings.Contains(c.Content, `emit "hello from an .ns capsule"`) {
			t.Error("neuroscript capsule content seems incorrect")
		}
	})
}
