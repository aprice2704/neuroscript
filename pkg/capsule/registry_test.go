// pkg/capsule/registry_test.go
package capsule_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/capsule"
)

func TestRegisterComputesSHAWhenEmpty(t *testing.T) {
	id := "capsule/sha-demo/1"
	content := "hello, capsule"

	// Register without SHA to ensure the registry computes it.
	if err := capsule.Register(capsule.Capsule{
		ID:      id,
		Version: "1",
		MIME:    "text/markdown; charset=utf-8",
		Content: content,
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	c, ok := capsule.Get(id)
	if !ok {
		t.Fatalf("Get(%q) not found", id)
	}
	sum := sha256.Sum256([]byte(content))
	want := hex.EncodeToString(sum[:])
	if c.SHA256 != want {
		t.Fatalf("SHA mismatch: got %s, want %s", c.SHA256, want)
	}
	if c.Size != len(content) {
		t.Fatalf("Size mismatch: got %d, want %d", c.Size, len(content))
	}
}

func TestMustRegisterPanicsOnBadID(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("MustRegister should panic on invalid ID")
		}
	}()
	capsule.MustRegister(capsule.Capsule{
		ID:      "Capsule/BadUpper/1", // invalid: uppercase "C"
		Version: "1",
		MIME:    "text/plain",
		Content: "x",
	})
}

func TestListOrderingByPriorityThenID(t *testing.T) {
	// Same priority, order by ID
	a := capsule.Capsule{ID: "capsule/sorta/1", Version: "1", MIME: "text/plain", Content: "A", Priority: 20}
	b := capsule.Capsule{ID: "capsule/sortb/1", Version: "1", MIME: "text/plain", Content: "B", Priority: 20}
	// Lower priority sorts first
	lo := capsule.Capsule{ID: "capsule/low/1", Version: "1", MIME: "text/plain", Content: "L", Priority: 10}

	for _, c := range []capsule.Capsule{a, b, lo} {
		if err := capsule.Register(c); err != nil {
			t.Fatalf("Register %s: %v", c.ID, err)
		}
	}

	list := capsule.List()
	// Extract only our test IDs in their observed order.
	var got []string
	for _, c := range list {
		if c.ID == a.ID || c.ID == b.ID || c.ID == lo.ID {
			got = append(got, c.ID)
		}
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 test capsules in List(), got %d", len(got))
	}
	// Expect low-priority first, then a then b (alphabetical by ID).
	want := []string{lo.ID, a.ID, b.ID}
	if !equalStrings(got, want) {
		t.Fatalf("order mismatch: got %v, want %v", got, want)
	}
}

func TestRegisterReplacesAndUpdatesSHA(t *testing.T) {
	id := "capsule/replace/1"
	if err := capsule.Register(capsule.Capsule{ID: id, Version: "1", MIME: "text/plain", Content: "old"}); err != nil {
		t.Fatalf("Register old: %v", err)
	}
	oldC, _ := capsule.Get(id)

	if err := capsule.Register(capsule.Capsule{ID: id, Version: "1", MIME: "text/plain", Content: "new"}); err != nil {
		t.Fatalf("Register new: %v", err)
	}
	newC, _ := capsule.Get(id)

	if oldC.SHA256 == newC.SHA256 {
		t.Fatalf("SHA should change when content changes")
	}
	if newC.Content != "new" {
		t.Fatalf("content not replaced: %q", newC.Content)
	}
}

func TestValidateIDCases(t *testing.T) {
	cases := []struct {
		id    string
		valid bool
	}{
		{"capsule/aeiou/1", true},
		{"capsule/foo-bar_9/42", true},
		{"Capsule/bad/1", false}, // uppercase not allowed
		{"capsule/Bad/1", false}, // uppercase in name
		{"capsule/space bad/1", false},
		{"capsule/missingver", false},
		{"capsule/name/one", false}, // version must be integer
		{"capsule//1", false},
	}
	for _, tc := range cases {
		err := capsule.ValidateID(tc.id)
		if tc.valid && err != nil {
			t.Errorf("ValidateID(%q) unexpected error: %v", tc.id, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("ValidateID(%q) expected error, got nil", tc.id)
		}
	}
}

func TestToolListAndRead(t *testing.T) {
	const id = "capsule/aeiou/1" // registered via embed init()

	tool := api.NewCapsuleTool()
	ctx := context.Background()

	// List all
	all := tool.List(ctx, nil)
	if len(all) == 0 || !all[id] {
		t.Fatalf("tool.List(nil) did not include %q", id)
	}

	// List specific
	specific := tool.List(ctx, []string{id, "capsule/not-found/1"})
	if !specific[id] || specific["capsule/not-found/1"] {
		t.Fatalf("tool.List specific mismatch: %v", specific)
	}

	// Read ok
	content, ver, sha, mime, ok := tool.Read(ctx, id)
	if !ok {
		t.Fatalf("tool.Read(%q) ok=false", id)
	}
	if content == "" || ver == "" || sha == "" || mime == "" {
		t.Fatalf("tool.Read returned empty fields")
	}

	// Read not found
	if _, _, _, _, ok := tool.Read(ctx, "capsule/does-not-exist/1"); ok {
		t.Fatalf("tool.Read(nonexistent) ok=true, want false")
	}
}

// Helper
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aa := append([]string(nil), a...)
	bb := append([]string(nil), b...)
	sort.Strings(aa)
	sort.Strings(bb)
	for i := range aa {
		if aa[i] != bb[i] {
			return false
		}
	}
	return true
}
