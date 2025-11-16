// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Tests for the public API wrappers for capsule management.
// Latest change: Replaced stale NewAdminCapsuleRegistry with NewCapsuleRegistry.
// filename: pkg/api/capsule_test.go
// nlines: 70
// risk_rating: LOW

package api_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// TestAPI_DefaultCapsuleStore verifies that the default store can be
// accessed via the public API and that it contains the expected embedded capsules.
func TestAPI_DefaultCapsuleStore(t *testing.T) {
	// --- THE FIX: Use the new DefaultCapsuleStore ---
	store := api.DefaultCapsuleStore()
	if store == nil {
		t.Fatal("api.DefaultCapsuleStore() returned nil")
	}
	// --- END FIX ---

	// Check for a known-good capsule that is loaded from the
	// embedded content/ directory.
	const expectedName = "capsule/aeiou"
	c, ok := store.GetLatest(expectedName)
	if !ok {
		t.Fatalf("Default store failed to find capsule %q", expectedName)
	}

	if c.Name != expectedName {
		t.Errorf("Capsule name mismatch: got %q, want %q", c.Name, expectedName)
	}
	if c.Content == "" {
		t.Error("Capsule content is empty")
	}
	// --- THE FIX: The loader guarantees a description ---
	if c.Description == "" {
		t.Error("Capsule description is empty, but loader should require it")
	}
	// --- END FIX ---
}

// TestAPI_NewCapsuleStore verifies that a layered store can be created
// and queried via the public API.
func TestAPI_NewCapsuleStore(t *testing.T) {
	// 1. Create two separate registries (using the admin constructor)
	// --- THE FIX: Use the correct, exported function ---
	reg1 := api.NewCapsuleRegistry()
	reg2 := api.NewCapsuleRegistry()
	// --- END FIX ---

	name := "capsule/store-test"
	// --- THE FIX: Added mandatory Description field ---
	reg1.MustRegister(api.Capsule{Name: name, Version: "1", Content: "v1", Description: "v1"})
	reg1.MustRegister(api.Capsule{Name: name, Version: "10", Content: "v10", Description: "v10"})
	reg2.MustRegister(api.Capsule{Name: name, Version: "99", Content: "v99", Description: "v99"})
	// --- END FIX ---

	// 2. Create a store with reg1 having priority
	store := api.NewCapsuleStore(reg1, reg2)

	// 3. GetLatest should find v10 from reg1 and ignore reg2
	latest, ok := store.GetLatest(name)
	if !ok {
		t.Fatalf("store.GetLatest(%q) failed", name)
	}
	if latest.Version != "10" {
		t.Errorf("GetLatest version mismatch: got %s, want 10 (should ignore v99 in reg2)", latest.Version)
	}

	// 4. Get should find v99 from reg2
	specific, ok := store.Get(name, "99")
	if !ok {
		t.Fatalf("store.Get(%q, '99') failed", name)
	}
	if specific.Content != "v99" {
		t.Errorf("Get content mismatch: got %q, want 'v99'", specific.Content)
	}
}
