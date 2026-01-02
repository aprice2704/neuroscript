// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 11
// :: description: Logic for capsule tools, including registry operations and metadata parsing.
// :: latestChange: Updated parseCapsuleFunc to use ErrInvalidCapsuleData and map format names to full MIME types.
// :: filename: pkg/tool/capsule/tools.go
// :: serialization: go

package capsule

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/metadata"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- Tool Functions ---

// capsuleRuntime defines the interface we expect from the runtime
// for all capsule store operations (read and write).
type capsuleRuntime interface {
	CapsuleStore() *capsule.Store
	GetExecPolicy() *policy.ExecPolicy
}

func getCapsuleStore(rt tool.Runtime) (*capsule.Store, error) {
	interp, ok := rt.(capsuleRuntime)
	if !ok {
		return nil, fmt.Errorf("internal error: runtime does not provide a CapsuleStore")
	}
	store := interp.CapsuleStore()
	if store == nil {
		return nil, fmt.Errorf("internal error: runtime returned a nil CapsuleStore")
	}
	return store, nil
}

func listCapsulesFunc(rt tool.Runtime, args []interface{}) (interface{}, error) {
	store, err := getCapsuleStore(rt)
	if err != nil {
		return nil, err
	}
	allCapsules := store.List()
	ids := make([]string, len(allCapsules))
	for i, c := range allCapsules {
		ids[i] = c.ID
	}
	return ids, nil
}

func readCapsuleFunc(rt tool.Runtime, args []interface{}) (interface{}, error) {
	id, ok := args[0].(string)
	if !ok {
		return nil, lang.ErrInvalidArgument
	}

	store, err := getCapsuleStore(rt)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(id, "@")
	if len(parts) != 2 {
		return lang.NewErrorValue("invalid_argument", fmt.Sprintf("invalid capsule ID format; expected <name>@<version>, got %s", id), nil), nil
	}
	name, version := parts[0], parts[1]

	c, found := store.Get(name, version)
	if !found {
		return lang.NewErrorValue("not_found", fmt.Sprintf("capsule '%s' not found", id), nil), nil
	}

	return capsuleToMap(c), nil
}

func getLatestCapsuleFunc(rt tool.Runtime, args []interface{}) (interface{}, error) {
	name, ok := args[0].(string)
	if !ok {
		return nil, lang.ErrInvalidArgument
	}

	store, err := getCapsuleStore(rt)
	if err != nil {
		return nil, err
	}

	c, found := store.GetLatest(name)
	if !found {
		return lang.NewErrorValue("not_found", fmt.Sprintf("latest capsule for '%s' not found", name), nil), nil
	}

	return capsuleToMap(c), nil
}

func addCapsuleFunc(rt tool.Runtime, args []interface{}) (interface{}, error) {
	capsuleContent, ok := args[0].(string)
	if !ok {
		return nil, lang.ErrInvalidArgument
	}

	runtimeWithPolicy, ok := rt.(capsuleRuntime)
	if !ok {
		return nil, fmt.Errorf("internal error: runtime does not implement capsuleRuntime interface")
	}
	execPolicy := runtimeWithPolicy.GetExecPolicy()
	if execPolicy == nil {
		return nil, fmt.Errorf("internal error: runtime returned nil ExecPolicy")
	}

	toolMeta := policy.ToolMeta{
		Name:          "tool.capsule.Add",
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "capsule", Verbs: []string{"write"}, Scopes: []string{"*"}},
		},
	}

	if err := execPolicy.CanCall(toolMeta); err != nil {
		return nil, err
	}

	store, err := getCapsuleStore(rt)
	if err != nil {
		return nil, err
	}

	meta, contentBody, _, err := metadata.ParseWithAutoDetect(strings.NewReader(capsuleContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse capsule content: %w", err)
	}

	extractor := metadata.NewExtractor(meta)
	if err := extractor.CheckRequired("id", "version", "description", "serialization"); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCapsuleData, err)
	}

	newCap := capsule.Capsule{
		Name:        extractor.MustGet("id"),
		Version:     extractor.MustGet("version"),
		Description: extractor.MustGet("description"),
		Content:     string(bytes.TrimSpace(contentBody)),
	}

	if err := store.Register(newCap); err != nil {
		return nil, fmt.Errorf("failed to register new capsule: %w", err)
	}

	return map[string]interface{}{
		"id":            newCap.Name,
		"version":       newCap.Version,
		"description":   newCap.Description,
		"serialization": extractor.MustGet("serialization"),
	}, nil
}

func parseCapsuleFunc(rt tool.Runtime, args []interface{}) (interface{}, error) {
	content, ok := args[0].(string)
	if !ok {
		return nil, lang.ErrInvalidArgument
	}

	meta, _, detectedFormat, err := metadata.ParseWithAutoDetect(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	extractor := metadata.NewExtractor(meta)
	if err := extractor.CheckRequired("id", "version", "description", "serialization"); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCapsuleData, err)
	}

	return map[string]interface{}{
		"handle":      extractor.MustGet("id"),
		"version":     extractor.MustGet("version"),
		"description": extractor.MustGet("description"),
		"mime":        extractor.GetOr("mime", toMimeType(detectedFormat)),
	}, nil
}

func toMimeType(format string) string {
	switch format {
	case "md":
		return "text/markdown"
	case "ns":
		return "application/x-neuroscript"
	default:
		return format
	}
}

func capsuleToMap(c capsule.Capsule) map[string]interface{} {
	return map[string]interface{}{
		"id":          c.ID,
		"name":        c.Name,
		"version":     c.Version,
		"description": c.Description,
		"mime":        c.MIME,
		"content":     c.Content,
		"sha256":      c.SHA256,
		"size":        c.Size,
	}
}
