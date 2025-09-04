// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Tests that the AEIOU capsule is correctly registered and accessible.
// filename: pkg/capsule/aeiou_test.go
// nlines: 60
// risk_rating: LOW
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
	const name = "capsule/aeiou"

	// Get the latest version from the registry
	c, ok := capsule.GetLatest(name)
	if !ok {
		t.Fatalf("latest capsule for %q not registered", name)
	}
	if c.Content == "" {
		t.Fatalf("capsule %q has empty content", name)
	}
	if c.MIME == "" {
		t.Fatalf("capsule %q MIME not set", name)
	}
	if c.Version == "" {
		t.Fatalf("capsule %q version not set", name)
	}

	// SHA should match content
	sum := sha256.Sum256([]byte(c.Content))
	want := hex.EncodeToString(sum[:])
	if c.SHA256 != want {
		t.Fatalf("capsule %q SHA mismatch: got %s, want %s", name, c.SHA256, want)
	}

	// List should include it
	found := false
	for _, x := range capsule.List() {
		if x.Name == name {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("capsule %q not found in List()", name)
	}

	// Tool surface should see it and round-trip the same metadata.
	tool := api.NewCapsuleTool()
	ctx := context.Background()

	// Use the fully qualified ID for the tool tests
	listed := tool.List(ctx, []string{c.ID})
	if !listed[c.ID] {
		t.Fatalf("tool.List did not report %q", c.ID)
	}

	content, ver, sha, mime, ok := tool.Read(ctx, c.ID)
	if !ok {
		t.Fatalf("tool.Read(%q) returned ok=false", c.ID)
	}
	if ver != c.Version || sha != c.SHA256 || mime != c.MIME || content != c.Content {
		t.Fatalf("tool.Read metadata/content mismatch for %q", c.ID)
	}
}
