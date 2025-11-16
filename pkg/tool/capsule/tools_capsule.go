// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Wraps the 'any' result from CapsuleProvider with lang.Wrap().
// Latest change: Added explicit policy check to addCapsuleFunc to enforce security.
// filename: pkg/tool/capsule/tools.go
// nlines: 194
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
	// --- ADDED: We must be able to get the policy ---
	GetExecPolicy() *policy.ExecPolicy
}

// capsuleAdminRuntime -- REMOVED.
// capsuleProviderRuntime -- REMOVED.

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

// getCapsuleRegistryForAdmin -- REMOVED.

func listCapsulesFunc(rt tool.Runtime, args []interface{}) (interface{}, error) {
	// 1. Check for host-provided capsule service -- REMOVED.

	// 2. Use internal registry logic
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

	// 1. Check for host-provided capsule service -- REMOVED.

	// 2. Use internal registry logic
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

	// 1. Check for host-provided capsule service -- REMOVED.

	// 2. Use internal registry logic
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

	// --- THE FIX: Enforce policy AT THE TOOL LEVEL ---
	// This ensures the tool is secure even if called directly (as in tests).
	runtimeWithPolicy, ok := rt.(capsuleRuntime)
	if !ok {
		return nil, fmt.Errorf("internal error: runtime does not implement capsuleRuntime interface")
	}
	execPolicy := runtimeWithPolicy.GetExecPolicy()
	if execPolicy == nil {
		return nil, fmt.Errorf("internal error: runtime returned nil ExecPolicy")
	}

	// Manually define the tool's metadata for the policy check
	// (This must match tooldefs_capsule.go)
	toolMeta := policy.ToolMeta{
		Name:          "tool.capsule.Add",
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "capsule", Verbs: []string{"write"}, Scopes: []string{"*"}},
		},
	}

	if err := execPolicy.CanCall(toolMeta); err != nil {
		return nil, err // This will be policy.ErrCapability
	}
	// --- END FIX ---

	// 1. Check for host-provided capsule service -- REMOVED.

	// 2. Fallback to internal registry logic
	store, err := getCapsuleStore(rt)
	if err != nil {
		return nil, err
	}

	meta, contentBody, _, err := metadata.ParseWithAutoDetect(strings.NewReader(capsuleContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse capsule content: %w", err)
	}

	extractor := metadata.NewExtractor(meta)
	// --- THE FIX: Enforce all required fields from the spec ---
	if err := extractor.CheckRequired("id", "version", "description", "serialization"); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCapsuleData, err)
	}
	// --- END FIX ---

	newCap := capsule.Capsule{
		Name:    extractor.MustGet("id"),
		Version: extractor.MustGet("version"),
		// --- THE FIX: Set the description field for Register ---
		Description: extractor.MustGet("description"),
		// --- END FIX ---
		Content: string(bytes.TrimSpace(contentBody)),
	}

	// --- THE FIX: Call Register on the *store* ---
	if err := store.Register(newCap); err != nil {
		return nil, fmt.Errorf("failed to register new capsule: %w", err)
	}
	// --- END FIX ---

	// --- THE FIX: Return the parsed metadata map, not capsuleToMap() ---
	// This matches what the test expects and confirms what was parsed.
	return map[string]interface{}{
		"id":            newCap.Name, // 'id' in metadata maps to 'Name' in struct
		"version":       newCap.Version,
		"description":   newCap.Description,
		"serialization": extractor.MustGet("serialization"),
	}, nil
	// --- END FIX ---
}

func capsuleToMap(c capsule.Capsule) map[string]interface{} {
	return map[string]interface{}{
		"id":          c.ID, // This is "name@version"
		"name":        c.Name,
		"version":     c.Version,
		"description": c.Description,
		"mime":        c.MIME,
		"content":     c.Content,
		"sha256":      c.SHA256,
		"size":        c.Size,
	}
}
