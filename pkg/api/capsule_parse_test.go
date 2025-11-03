// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Corrects tests to use a valid capsule file and adds '::serialization:' to failure cases.
// filename: pkg/api/capsule_parse_test.go
// nlines: 104
// risk_rating: LOW

package api_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// bootstrapAgenticContent is the full content of the valid capsule file
// provided in 'pkg/api/test_fixtures/bootstrap_agentic.md'.
const bootstrapAgenticContent = `
# AEIOU v3 — Agentic Bootstrap Capsule (v7-draft)

You run inside the host’s NeuroScript (ns) interpreter. You will always receive a single AEIOU v3 envelope. Your job is to return that envelope with only the ACTIONS section filled by exactly one ` + "`command … endcommand`" + ` block.

The host controls the turn loop based on the 'max_turns' setting. You signal when you are done.

---

## Part 1 — Hard Contract (must be followed literally)
... (omitted for brevity) ...
---

## Part 3 — Minimal Examples
... (omitted for brevity) ...
---

## Metadata

::schema: instructions  
::serialization: md  
::id: capsule/bootstrap_agentic  
::version: 7
::fileVersion: 1  
::filename: pkg/api/test_fixtures/bootstrap_agentic.md
::author: NeuroScript Docs Team  
::modified: 2025-10-16  
::description: Hard-contract AEIOU v3 bootstrap capsule for multi-turn agents. The Go host controls the loop; the agent signals 'done' with a simple emit.
`

func TestAPI_ParseCapsule_Success(t *testing.T) {
	// Use the full, valid content of the provided .md capsule
	cap, err := api.ParseCapsule([]byte(bootstrapAgenticContent))
	if err != nil {
		t.Fatalf("api.ParseCapsule failed unexpectedly on valid file: %v", err)
	}

	if cap.Name != "capsule/bootstrap_agentic" {
		t.Errorf("Name mismatch: got %q, want 'capsule/bootstrap_agentic'", cap.Name)
	}
	if cap.Version != "7" {
		t.Errorf("Version mismatch: got %q, want '7'", cap.Version)
	}
	if cap.ID != "capsule/bootstrap_agentic@7" {
		t.Errorf("ID mismatch: got %q, want 'capsule/bootstrap_agentic@7'", cap.ID)
	}
	if cap.Description != "Hard-contract AEIOU v3 bootstrap capsule for multi-turn agents. The Go host controls the loop; the agent signals 'done' with a simple emit." {
		t.Errorf("Description mismatch: got %q", cap.Description)
	}
	if !strings.HasPrefix(cap.Content, "# AEIOU v3") {
		t.Errorf("Content seems incorrect, does not start with expected title: %q", cap.Content)
	}
	if cap.SHA256 == "" {
		t.Error("SHA256 was not calculated")
	}
}

func TestAPI_ParseCapsule_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "Missing version",
			content: "::serialization: ns\n:: id: capsule/foo\n:: description: bar",
			wantErr: "missing required metadata",
		},
		{
			name:    "Invalid name format",
			content: "::serialization: ns\n:: id: capsule/with.dot\n:: version: 1\n:: description: bar",
			wantErr: "invalid capsule '::id'",
		},
		{
			name:    "Invalid version format",
			content: "::serialization: ns\n:: id: capsule/foo\n:: version: v1.0\n:: description: bar",
			wantErr: "must be an integer",
		},
		{
			name:    "No metadata (missing serialization)",
			content: "Just content",
			wantErr: "failed to parse capsule metadata",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := api.ParseCapsule([]byte(tc.content))
			if err == nil {
				t.Fatalf("Expected an error containing %q, but got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Error mismatch:\n  Got: %v\n  Want: %v", err.Error(), tc.wantErr)
			}
		})
	}
}
