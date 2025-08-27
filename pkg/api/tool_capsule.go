package api

import (
	"context"

	"github.com/aprice2704/neuroscript/pkg/capsule"
)

// CapsuleTool is the public surface you can expose to NS as tool.capsule
type CapsuleTool interface {
	// List reports presence (by ID). If names is empty, list all known IDs.
	List(ctx context.Context, names []string) map[string]bool
	// Read returns (content, version, sha256, mime). ok=false if not found.
	Read(ctx context.Context, id string) (content, version, sha256, mime string, ok bool)
}

type capsuleTool struct{}

func NewCapsuleTool() CapsuleTool { return &capsuleTool{} }

func (c *capsuleTool) List(ctx context.Context, names []string) map[string]bool {
	out := map[string]bool{}
	if len(names) == 0 {
		for _, cap := range capsule.List() {
			out[cap.ID] = true
		}
		return out
	}
	for _, name := range names {
		_, ok := capsule.Get(name)
		out[name] = ok
	}
	return out
}

func (c *capsuleTool) Read(ctx context.Context, id string) (string, string, string, string, bool) {
	if cap, ok := capsule.Get(id); ok {
		return cap.Content, cap.Version, cap.SHA256, cap.MIME, true
	}
	return "", "", "", "", false
}
