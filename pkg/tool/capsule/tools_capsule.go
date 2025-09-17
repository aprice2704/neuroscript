// NeuroScript Version: 0.7.2
// File version: 7
// Purpose: Implements the Go functions for the capsule toolset, now returning sentinel errors for invalid input.
// filename: pkg/tool/capsule/tools.go
// nlines: 148
// risk_rating: HIGH
package capsule

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// capsuleRuntime defines the interface we expect from the runtime
// for capsule store read operations.
type capsuleRuntime interface {
	CapsuleStore() *capsule.Store
}

// capsuleAdminRuntime defines the interface for write operations,
// allowing access to a mutable registry.
type capsuleAdminRuntime interface {
	CapsuleRegistryForAdmin() *capsule.Registry
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

func getCapsuleRegistryForAdmin(rt tool.Runtime) (*capsule.Registry, error) {
	interp, ok := rt.(capsuleAdminRuntime)
	if !ok {
		return nil, ErrAdminRegistryNotAvailable
	}
	reg := interp.CapsuleRegistryForAdmin()
	if reg == nil {
		return nil, ErrAdminRegistryNotAvailable
	}
	return reg, nil
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
	store, err := getCapsuleStore(rt)
	if err != nil {
		return nil, err
	}
	id, ok := args[0].(string)
	if !ok {
		return nil, lang.ErrInvalidArgument
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
	store, err := getCapsuleStore(rt)
	if err != nil {
		return nil, err
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, lang.ErrInvalidArgument
	}

	c, found := store.GetLatest(name)
	if !found {
		return lang.NewErrorValue("not_found", fmt.Sprintf("latest capsule for '%s' not found", name), nil), nil
	}

	return capsuleToMap(c), nil
}

func addCapsuleFunc(rt tool.Runtime, args []interface{}) (interface{}, error) {
	reg, err := getCapsuleRegistryForAdmin(rt)
	if err != nil {
		return nil, err
	}
	data, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, ErrInvalidCapsuleData
	}

	getString := func(key string) string {
		if val, ok := data[key].(string); ok {
			return val
		}
		return ""
	}

	newCap := capsule.Capsule{
		Name:    getString("name"),
		Version: getString("version"),
		Content: getString("content"),
	}

	if newCap.Name == "" || newCap.Version == "" {
		return nil, ErrInvalidCapsuleData
	}

	if err := reg.Register(newCap); err != nil {
		// Propagate the specific validation error from the registry.
		return nil, fmt.Errorf("failed to register new capsule: %w", err)
	}

	return true, nil
}

func capsuleToMap(c capsule.Capsule) map[string]interface{} {
	return map[string]interface{}{
		"id":      c.ID,
		"name":    c.Name,
		"version": c.Version,
		"mime":    c.MIME,
		"content": c.Content,
		"sha256":  c.SHA256,
		"size":    c.Size,
	}
}
