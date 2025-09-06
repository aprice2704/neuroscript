// NeuroScript Version: 0.7.1
// File version: 3
// Purpose: Tests that the capsule loader correctly registers embedded content into the default registry.
// filename: pkg/capsule/loader_test.go
// nlines: 25
// risk_rating: LOW
package capsule_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
)

func TestCapsuleLoader(t *testing.T) {
	const expectedName = "capsule/aeiou"

	// The loader runs on init, populating the default registry.
	reg := capsule.DefaultRegistry()
	c, ok := reg.GetLatest(expectedName)

	if !ok {
		t.Fatalf("reg.GetLatest(%q) failed: capsule was not loaded and registered", expectedName)
	}

	if c.Name != expectedName {
		t.Errorf("capsule Name mismatch: got %q, want %q", c.Name, expectedName)
	}

	// Version can change, so we just check it's not empty
	if c.Version == "" {
		t.Errorf("capsule Version is empty after loading")
	}

	if c.Content == "" {
		t.Error("capsule Content is empty after loading")
	}
}
