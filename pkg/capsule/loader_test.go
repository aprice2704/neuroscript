// NeuroScript Version: 0.3.0
// File version: 1
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
	const expectedID = "capsule/aeiou/1"
	c, ok := capsule.Get(expectedID)

	if !ok {
		t.Fatalf("capsule.Get(%q) failed: capsule was not loaded and registered", expectedID)
	}

	if c.ID != expectedID {
		t.Errorf("capsule ID mismatch: got %q, want %q", c.ID, expectedID)
	}
	if c.Version != "1" {
		t.Errorf("capsule Version mismatch: got %q, want %q", c.Version, "1")
	}
	if c.Content == "" {
		t.Error("capsule Content is empty after loading")
	}
}
