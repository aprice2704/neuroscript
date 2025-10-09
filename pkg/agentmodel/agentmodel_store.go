// NeuroScript Version: 0.8.0
// File version: 17
// Purpose: FIX: Restored NewAgentModelAdmin and NewAgentModelReader facades to fix test compilation.
// filename: pkg/agentmodel/agentmodel_store.go
// nlines: 142
// risk_rating: HIGH

package agentmodel

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/mitchellh/mapstructure"
)

// ---------- store ----------

type AgentModelStore struct {
	mu sync.RWMutex
	m  map[string]types.AgentModel // key: lower(name)
}

func NewAgentModelStore() *AgentModelStore {
	return &AgentModelStore{m: make(map[string]types.AgentModel)}
}

// ---------- reader view ----------

type readerView struct {
	s *AgentModelStore
}

func NewReader(s *AgentModelStore) interfaces.AgentModelReader {
	return &readerView{s: s}
}

func (v *readerView) List() []types.AgentModelName {
	v.s.mu.RLock()
	defer v.s.mu.RUnlock()
	out := make([]types.AgentModelName, 0, len(v.s.m))
	for _, model := range v.s.m {
		out = append(out, model.Name)
	}
	return out
}

func (v *readerView) Get(name types.AgentModelName) (any, bool) {
	key := strings.ToLower(string(name))
	v.s.mu.RLock()
	defer v.s.mu.RUnlock()
	model, ok := v.s.m[key]
	return model, ok
}

// ---------- admin view (policy-gated) ----------

type adminView struct {
	s   *AgentModelStore
	pol *interfaces.ExecPolicy
}

func NewAdmin(s *AgentModelStore, pol *interfaces.ExecPolicy) interfaces.AgentModelAdmin {
	return &adminView{s: s, pol: pol}
}

func (v *adminView) List() []types.AgentModelName { return NewReader(v.s).List() }

func (v *adminView) Get(name types.AgentModelName) (any, bool) {
	return NewReader(v.s).Get(name)
}

func (v *adminView) Register(name types.AgentModelName, cfg map[string]any) error {
	if err := v.ensureConfigContext(); err != nil {
		return err
	}
	key := strings.ToLower(string(name))

	v.s.mu.Lock()
	defer v.s.mu.Unlock()

	if _, exists := v.s.m[key]; exists {
		return lang.ErrDuplicateKey
	}

	model, err := modelFromCfg(name, cfg, nil)
	if err != nil {
		return err
	}

	v.s.m[key] = model
	return nil
}

func (v *adminView) Update(name types.AgentModelName, updates map[string]any) error {
	if err := v.ensureConfigContext(); err != nil {
		return err
	}
	key := strings.ToLower(string(name))

	v.s.mu.Lock()
	defer v.s.mu.Unlock()

	cur, ok := v.s.m[key]
	if !ok {
		return lang.ErrNotFound
	}

	model, err := modelFromCfg(name, updates, &cur)
	if err != nil {
		return err
	}

	v.s.m[key] = model
	return nil
}

func (v *adminView) Delete(name types.AgentModelName) bool {
	if err := v.ensureConfigContext(); err != nil {
		return false
	}
	key := strings.ToLower(string(name))

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

// --- FIX: Add back constructor facades to keep tests working ---
func NewAgentModelReader(s *AgentModelStore) interfaces.AgentModelReader {
	return NewReader(s)
}

func NewAgentModelAdmin(s *AgentModelStore, pol *interfaces.ExecPolicy) interfaces.AgentModelAdmin {
	return NewAdmin(s, pol)
}

// --- END FIX ---

func modelFromCfg(name types.AgentModelName, cfg map[string]any, base *types.AgentModel) (types.AgentModel, error) {
	out := types.AgentModel{Name: name} // Always start fresh
	if base != nil {
		out = *base // If updating, start with the existing model
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &out,
		WeaklyTypedInput: true, // Allows, e.g., float64 -> int
		TagName:          "mapstructure",
	})
	if err != nil {
		return types.AgentModel{}, fmt.Errorf("internal error: failed to create decoder: %w", err)
	}

	if err := decoder.Decode(cfg); err != nil {
		return types.AgentModel{}, fmt.Errorf("failed to decode agentmodel config: %w", err)
	}

	// --- Handle Deprecated Fields for backward compatibility ---
	if val, ok := cfg["temperature"]; ok {
		if f, isF := val.(float64); isF {
			out.Generation.Temperature = f
		}
	}
	if val, ok := cfg["tool_loop_permitted"]; ok {
		if b, isB := val.(bool); isB {
			out.Tools.ToolLoopPermitted = b
		}
	}
	if val, ok := cfg["auto_loop_enabled"]; ok {
		if b, isB := val.(bool); isB {
			out.Tools.AutoLoopEnabled = b
		}
	}
	// ---

	if out.Provider == "" || out.Model == "" {
		return types.AgentModel{}, fmt.Errorf("'provider' and 'model' are required")
	}

	out.Name = name // Ensure the name is always set correctly
	return out, nil
}
