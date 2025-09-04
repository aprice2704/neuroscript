// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Tests that the capsule loader correctly registers embedded content.
// filename: pkg/capsule/loader_test.go
// nlines: 21
// risk_rating: LOW
package capsule_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
)

func TestCapsuleLoader(t *testing.T) {
	const expectedName = "capsule/aeiou"
	c, ok := capsule.GetLatest(expectedName)

	if !ok {
		t.Fatalf("capsule.GetLatest(%q) failed: capsule was not loaded and registered", expectedName)
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
