package capsule_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/capsule"
)

func TestAEIOUCapsuleRegistered(t *testing.T) {
	const id = "capsule/aeiou/1"

	// Exists in registry (external package reference)
	c, ok := capsule.Get(id)
	if !ok {
		t.Fatalf("capsule %q not registered", id)
	}
	if c.Content == "" {
		t.Fatalf("capsule %q has empty content", id)
	}
	if c.MIME == "" {
		t.Fatalf("capsule %q MIME not set", id)
	}
	if c.Version != "1" {
		t.Fatalf("capsule %q version = %q, want %q", id, c.Version, "1")
	}

	// SHA should match content
	sum := sha256.Sum256([]byte(c.Content))
	want := hex.EncodeToString(sum[:])
	if c.SHA256 != want {
		t.Fatalf("capsule %q SHA mismatch: got %s, want %s", id, c.SHA256, want)
	}

	// List should include it
	found := false
	for _, x := range capsule.List() {
		if x.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("capsule %q not found in List()", id)
	}

	// Tool surface should see it and round-trip the same metadata.
	tool := api.NewCapsuleTool()
	ctx := context.Background()

	listed := tool.List(ctx, []string{id})
	if !listed[id] {
		t.Fatalf("tool.List did not report %q", id)
	}

	content, ver, sha, mime, ok := tool.Read(ctx, id)
	if !ok {
		t.Fatalf("tool.Read(%q) returned ok=false", id)
	}
	if ver != c.Version || sha != c.SHA256 || mime != c.MIME || content != c.Content {
		t.Fatalf("tool.Read metadata/content mismatch")
	}
}

func TestValidateID(t *testing.T) {
	cases := []struct {
		id    string
		valid bool
	}{
		{"capsule/aeiou/1", true},
		{"capsule/foo-bar_9/42", true},
		{"Capsule/Bad/1", false},       // wrong prefix case
		{"capsule/missingver", false},  // missing version
		{"capsule//1", false},          // empty name
		{"capsule/Name/one", false},    // version not integer
		{"capsule/space bad/1", false}, // invalid chars
	}
	for _, tc := range cases {
		err := capsule.ValidateID(tc.id)
		if tc.valid && err != nil {
			t.Fatalf("ValidateID(%q) unexpected error: %v", tc.id, err)
		}
		if !tc.valid && err == nil {
			t.Fatalf("ValidateID(%q) expected error, got nil", tc.id)
		}
	}
}
