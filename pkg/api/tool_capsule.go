// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Updates the capsule tool to use the version-aware registry, parsing 'name@version' IDs.
// filename: pkg/api/tool_capsule.go
// nlines: 65
// risk_rating: LOW

package api

import (
	"context"
	"errors"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/capsule"
)

// CapsuleTool is the public surface you can expose to NS as tool.capsule
type CapsuleTool interface {
	// List reports presence by fully-qualified ID (name@version). If ids is empty, list all known IDs.
	List(ctx context.Context, ids []string) map[string]bool
	// Read returns (content, version, sha256, mime) for a fully-qualified ID. ok=false if not found.
	Read(ctx context.Context, id string) (content, version, sha256, mime string, ok bool)
}

type capsuleTool struct{}

func NewCapsuleTool() CapsuleTool { return &capsuleTool{} }

func (c *capsuleTool) List(ctx context.Context, ids []string) map[string]bool {
	out := map[string]bool{}
	if len(ids) == 0 {
		// List all known capsule IDs of all versions.
		for _, cap := range capsule.List() {
			out[cap.ID] = true
		}
		return out
	}
	// Check for the presence of specific IDs.
	for _, id := range ids {
		name, version, err := parseID(id)
		if err != nil {
			out[id] = false
			continue
		}
		_, ok := capsule.Get(name, version)
		out[id] = ok
	}
	return out
}

func (c *capsuleTool) Read(ctx context.Context, id string) (string, string, string, string, bool) {
	name, version, err := parseID(id)
	if err != nil {
		return "", "", "", "", false
	}
	if cap, ok := capsule.Get(name, version); ok {
		return cap.Content, cap.Version, cap.SHA256, cap.MIME, true
	}
	return "", "", "", "", false
}

// parseID splits a fully-qualified capsule ID into its name and version parts.
func parseID(id string) (name, version string, err error) {
	parts := strings.Split(id, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("invalid capsule ID format; expected name@version")
	}
	return parts[0], parts[1], nil
}
