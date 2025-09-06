// NeuroScript Version: 0.7.1
// File version: 6
// Purpose: Corrected the test to use a mock interpreter to properly initialize the capsule tool.
// filename: pkg/capsule/registry_test.go
// nlines: 140
// risk_rating: MEDIUM
package capsule_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
)

func TestRegistry_ComputesSHAWhenEmpty(t *testing.T) {
	name := "capsule/sha-demo"
	content := "hello, capsule"
	reg := capsule.NewRegistry()

	if err := reg.Register(capsule.Capsule{
		Name:    name,
		Version: "1",
		MIME:    "text/markdown; charset=utf-8",
		Content: content,
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	c, ok := reg.Get(name, "1")
	if !ok {
		t.Fatalf("Get(%q, '1') not found", name)
	}
	sum := sha256.Sum256([]byte(content))
	want := hex.EncodeToString(sum[:])
	if c.SHA256 != want {
		t.Fatalf("SHA mismatch: got %s, want %s", c.SHA256, want)
	}
	if c.Size != len(content) {
		t.Fatalf("Size mismatch: got %d, want %d", c.Size, len(content))
	}
	if c.ID != "capsule/sha-demo@1" {
		t.Errorf("Expected ID to be 'capsule/sha-demo@1', got %s", c.ID)
	}
}

func TestRegistry_MustRegisterPanicsOnBadName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("MustRegister should panic on invalid name")
		}
	}()
	reg := capsule.NewRegistry()
	reg.MustRegister(capsule.Capsule{
		Name:    "Capsule/BadUpper", // invalid: uppercase "C"
		Version: "1",
		MIME:    "text/plain",
		Content: "x",
	})
}

func TestStore_ListOrderingByPriorityThenID(t *testing.T) {
	reg1 := capsule.NewRegistry()
	reg2 := capsule.NewRegistry()

	a := capsule.Capsule{Name: "capsule/sorta", Version: "1", Content: "A", Priority: 20}
	b := capsule.Capsule{Name: "capsule/sortb", Version: "1", Content: "B", Priority: 20}
	lo := capsule.Capsule{Name: "capsule/low", Version: "1", Content: "L", Priority: 10}

	reg1.MustRegister(a)
	reg2.MustRegister(b)
	reg1.MustRegister(lo)

	store := capsule.NewStore(reg1, reg2)
	list := store.List()

	if len(list) != 3 {
		t.Fatalf("expected 3 capsules in List(), got %d", len(list))
	}

	// Correct sorted order: lo (10), then a (20), then b (20)
	if list[0].Name != "capsule/low" {
		t.Errorf("Expected first item to be 'capsule/low', got %s", list[0].Name)
	}
	if list[1].Name != "capsule/sorta" {
		t.Errorf("Expected second item to be 'capsule/sorta', got %s", list[1].Name)
	}
	if list[2].Name != "capsule/sortb" {
		t.Errorf("Expected third item to be 'capsule/sortb', got %s", list[2].Name)
	}
}

func TestStore_GetLatest(t *testing.T) {
	reg1 := capsule.NewRegistry()
	reg2 := capsule.NewRegistry()

	// Integers in reg1
	name := "capsule/version-test"
	reg1.MustRegister(capsule.Capsule{Name: name, Version: "1", Content: "v1"})
	reg1.MustRegister(capsule.Capsule{Name: name, Version: "10", Content: "v10"})
	reg1.MustRegister(capsule.Capsule{Name: name, Version: "2", Content: "v2"})
	// Shadowed by reg1
	reg2.MustRegister(capsule.Capsule{Name: name, Version: "99", Content: "v99"})

	// Semver in reg2
	semverName := "capsule/semver-test"
	reg2.MustRegister(capsule.Capsule{Name: semverName, Version: "1.0.0"})
	reg2.MustRegister(capsule.Capsule{Name: semverName, Version: "1.1.0"})
	reg2.MustRegister(capsule.Capsule{Name: semverName, Version: "0.9.0"})

	store := capsule.NewStore(reg1, reg2)

	// Test case 1: GetLatest finds latest in the first registry and stops.
	latest, ok := store.GetLatest(name)
	if !ok {
		t.Fatalf("GetLatest(%q) failed", name)
	}
	if latest.Version != "10" {
		t.Errorf("GetLatest version mismatch: got %s, want 10 (should ignore v99 in reg2)", latest.Version)
	}

	// Test case 2: GetLatest finds latest in a subsequent registry.
	latestSemver, ok := store.GetLatest(semverName)
	if !ok {
		t.Fatalf("GetLatest(%q) failed", semverName)
	}
	if latestSemver.Version != "1.1.0" {
		t.Errorf("GetLatest semver mismatch: got %s, want 1.1.0", latestSemver.Version)
	}

	// Test case 3: Not found
	_, ok = store.GetLatest("capsule/not-found")
	if ok {
		t.Error("GetLatest found a capsule that does not exist")
	}
}

func TestValidateNameCases(t *testing.T) {
	cases := []struct {
		name  string
		valid bool
	}{
		{"capsule/aeiou", true},
		{"capsule/foo-bar_9", true},
		{"Capsule/bad", false},       // uppercase not allowed
		{"capsule/Bad", false},       // uppercase in name
		{"capsule/space bad", false}, // space
		{"capsule/missingver", true},
		{"capsule/", false}, // empty name
		{"foo/bar", false},  // wrong prefix
	}
	for _, tc := range cases {
		err := capsule.ValidateName(tc.name)
		if tc.valid && err != nil {
			t.Errorf("ValidateName(%q) unexpected error: %v", tc.name, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("ValidateName(%q) expected error, got nil", tc.name)
		}
	}
}
