// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Wraps the 'any' result from CapsuleProvider with lang.Wrap().
// filename: pkg/tool/capsule/tools.go
// nlines: 247
// risk_rating: HIGH
package capsule

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/metadata"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- Tool Functions ---

// capsuleRuntime defines the interface we expect from the runtime
// for capsule store read operations (the fallback).
type capsuleRuntime interface {
	CapsuleStore() *capsule.Store
}

// capsuleAdminRuntime defines the interface for write operations (the fallback),
// allowing access to a mutable registry.
type capsuleAdminRuntime interface {
	CapsuleRegistryForAdmin() *capsule.Registry
}

// capsuleProviderRuntime defines the interface we expect from the runtime
// if it has a host-provided capsule provider.
type capsuleProviderRuntime interface {
	tool.Runtime // Embed base runtime
	CapsuleProvider() interfaces.CapsuleProvider
	GetTurnContext() context.Context
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
	// 1. Check for host-provided capsule service
	if providerRuntime, ok := rt.(capsuleProviderRuntime); ok {
		if provider := providerRuntime.CapsuleProvider(); provider != nil {
			nativeResult, err := provider.List(providerRuntime.GetTurnContext())
			if err != nil {
				return nil, err
			}
			// Wrap the primitive 'any' result into a lang.Value
			return lang.Wrap(nativeResult)
		}
	}

	// 2. Fallback to internal registry logic
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

	// 1. Check for host-provided capsule service
	if providerRuntime, ok := rt.(capsuleProviderRuntime); ok {
		if provider := providerRuntime.CapsuleProvider(); provider != nil {
			nativeResult, err := provider.Read(providerRuntime.GetTurnContext(), id)
			if err != nil {
				return nil, err
			}
			// Wrap the primitive 'any' result into a lang.Value
			return lang.Wrap(nativeResult)
		}
	}

	// 2. Fallback to internal registry logic
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

	// 1. Check for host-provided capsule service
	if providerRuntime, ok := rt.(capsuleProviderRuntime); ok {
		if provider := providerRuntime.CapsuleProvider(); provider != nil {
			nativeResult, err := provider.GetLatest(providerRuntime.GetTurnContext(), name)
			if err != nil {
				return nil, err
			}
			// Wrap the primitive 'any' result into a lang.Value
			return lang.Wrap(nativeResult)
		}
	}

	// 2. Fallback to internal registry logic
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

	// 1. Check for host-provided capsule service
	if providerRuntime, ok := rt.(capsuleProviderRuntime); ok {
		if provider := providerRuntime.CapsuleProvider(); provider != nil {
			nativeResult, err := provider.Add(providerRuntime.GetTurnContext(), capsuleContent)
			if err != nil {
				return nil, err
			}
			// Wrap the primitive 'any' result into a lang.Value
			return lang.Wrap(nativeResult)
		}
	}

	// 2. Fallback to internal registry logic
	//	fmt.Fprintf(os.Stderr, "\n[CAPSULE DEBUG] --- tool.capsule.Add called (fallback) ---\n")
	reg, err := getCapsuleRegistryForAdmin(rt)
	if err != nil {
		return nil, err
	}

	meta, contentBody, _, err := metadata.ParseWithAutoDetect(strings.NewReader(capsuleContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse capsule content: %w", err)
	}

	extractor := metadata.NewExtractor(meta)
	if err := extractor.CheckRequired("id", "version"); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCapsuleData, err)
	}

	newCap := capsule.Capsule{
		Name:    extractor.MustGet("id"),
		Version: extractor.MustGet("version"),
		Content: string(bytes.TrimSpace(contentBody)),
	}

	if err := reg.Register(newCap); err != nil {
		return nil, fmt.Errorf("failed to register new capsule: %w", err)
	}

	return map[string]interface{}{
		"id":          newCap.Name,
		"version":     newCap.Version,
		"description": extractor.GetOr("description", ""),
	}, nil
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
