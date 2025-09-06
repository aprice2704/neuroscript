// NeuroScript Version: 0.7.1
// File version: 1
// Purpose: Implements the Go functions for the capsule toolset.
// filename: pkg/tool/capsule/tools_capsule.go
// nlines: 85
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
// for capsule store operations.
type capsuleRuntime interface {
	CapsuleStore() *capsule.Store
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

func toolListCapsules(rt tool.Runtime, args []interface{}) (interface{}, error) {
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

func toolReadCapsule(rt tool.Runtime, args []interface{}) (interface{}, error) {
	store, err := getCapsuleStore(rt)
	if err != nil {
		return nil, err
	}
	id, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument 'id' must be a string")
	}

	parts := strings.Split(id, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid capsule ID format; expected <name>@<version>, got %s", id)
	}
	name, version := parts[0], parts[1]

	c, found := store.Get(name, version)
	if !found {
		return &lang.NilValue{}, nil
	}

	return capsuleToMap(c), nil
}

func toolGetLatestCapsule(rt tool.Runtime, args []interface{}) (interface{}, error) {
	store, err := getCapsuleStore(rt)
	if err != nil {
		return nil, err
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument 'name' must be a string")
	}

	c, found := store.GetLatest(name)
	if !found {
		return &lang.NilValue{}, nil
	}

	return capsuleToMap(c), nil
}

func capsuleToMap(c capsule.Capsule) map[string]interface{} {
	return map[string]interface{}{
		"id":       c.ID,
		"name":     c.Name,
		"version":  c.Version,
		"mime":     c.MIME,
		"content":  c.Content,
		"sha256":   c.SHA256,
		"size":     c.Size,
		"priority": c.Priority,
	}
}
