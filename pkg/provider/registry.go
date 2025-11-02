// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Replaced non-existent lang.ErrNil with lang.ErrInvalidArgument.
// filename: pkg/provider/registry.go
// nlines: 119
// risk_rating: MEDIUM

package provider

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

// ---------- registry ----------

// Registry holds the concrete AIProvider implementations, mapped by name.
// This is populated by the host application at startup.
type Registry struct {
	mu sync.RWMutex
	m  map[string]AIProvider // key: lower(name)
}

// NewRegistry creates a new, empty provider registry.
func NewRegistry() *Registry {
	return &Registry{m: make(map[string]AIProvider)}
}

// ---------- reader view ----------

type readerView struct {
	s *Registry
}

// NewReader creates a new read-only view of the provider registry.
func NewReader(s *Registry) interfaces.ProviderRegistryReader {
	return &readerView{s: s}
}

func (v *readerView) List() []string {
	v.s.mu.RLock()
	defer v.s.mu.RUnlock()
	out := make([]string, 0, len(v.s.m))
	for name := range v.s.m {
		out = append(out, name)
	}
	return out
}

func (v *readerView) Get(name string) (any, bool) {
	key := strings.ToLower(name)
	v.s.mu.RLock()
	defer v.s.mu.RUnlock()
	p, ok := v.s.m[key]
	return p, ok
}

// ---------- admin view (policy-gated) ----------

type adminView struct {
	s   *Registry
	pol *policy.ExecPolicy
}

// NewAdmin creates a new admin view of the provider registry,
// gated by the provided execution policy.
func NewAdmin(s *Registry, pol *policy.ExecPolicy) interfaces.ProviderRegistryAdmin {
	return &adminView{s: s, pol: pol}
}

func (v *adminView) List() []string { return NewReader(v.s).List() }

func (v *adminView) Get(name string) (any, bool) {
	return NewReader(v.s).Get(name)
}

// Register adds a new AIProvider implementation to the registry.
// This is intended to be called by the host application during configuration.
func (v *adminView) Register(name string, p any) error {
	if err := v.ensureConfigContext(); err != nil {
		return err
	}
	if p == nil {
		return lang.ErrInvalidArgument // FIX: Use correct sentinel error
	}

	providerImpl, ok := p.(AIProvider)
	if !ok {
		return fmt.Errorf("invalid type for Register: expected provider.AIProvider, got %T", p)
	}

	key := strings.ToLower(name)

	v.s.mu.Lock()
	defer v.s.mu.Unlock()

	if _, exists := v.s.m[key]; exists {
		return lang.ErrDuplicateKey
	}

	v.s.m[key] = providerImpl
	return nil
}

func (v *adminView) Delete(name string) bool {
	if err := v.ensureConfigContext(); err != nil {
		return false
	}
	key := strings.ToLower(name)

	v.s.mu.Lock()
	defer v.s.mu.Unlock()
	if _, ok := v.s.m[key]; !ok {
		return false
	}
	delete(v.s.m, key)
	return true
}

// ---------- helpers ----------

func (v *adminView) ensureConfigContext() error {
	if v.pol == nil || v.pol.Context != policy.ContextConfig {
		return policy.ErrTrust
	}
	return nil
}
