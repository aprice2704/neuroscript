// NeuroScript Version: 0.6.0
// File version: 3.3.0
// Purpose: Replaced undefined error constant with an available one from the lang package to fix compilation.
// filename: pkg/runtime/agentmodel_store.go
// nlines: 190
// risk_rating: HIGH

package runtime

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
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

type agentModelReaderView struct {
	s *AgentModelStore
}

func NewAgentModelReader(s *AgentModelStore) interfaces.AgentModelReader {
	return &agentModelReaderView{s: s}
}

func (v *agentModelReaderView) List() []types.AgentModelName {
	v.s.mu.RLock()
	defer v.s.mu.RUnlock()
	out := make([]types.AgentModelName, 0, len(v.s.m))
	for _, model := range v.s.m {
		out = append(out, model.Name)
	}
	return out
}

func (v *agentModelReaderView) Get(name types.AgentModelName) (any, bool) {
	key := strings.ToLower(string(name))
	v.s.mu.RLock()
	defer v.s.mu.RUnlock()
	model, ok := v.s.m[key]
	return model, ok
}

// ---------- admin view (policy-gated) ----------

type agentModelAdminView struct {
	s   *AgentModelStore
	pol *ExecPolicy
}

func NewAgentModelAdmin(s *AgentModelStore, pol *ExecPolicy) interfaces.AgentModelAdmin {
	return &agentModelAdminView{s: s, pol: pol}
}

func (v *agentModelAdminView) List() []types.AgentModelName { return NewAgentModelReader(v.s).List() }

func (v *agentModelAdminView) Get(name types.AgentModelName) (any, bool) {
	return NewAgentModelReader(v.s).Get(name)
}

func (v *agentModelAdminView) Register(name types.AgentModelName, cfg map[string]any) error {
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

func (v *agentModelAdminView) Update(name types.AgentModelName, updates map[string]any) error {
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

func (v *agentModelAdminView) Delete(name types.AgentModelName) bool {
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

func (v *agentModelAdminView) ensureConfigContext() error {
	if v.pol == nil || v.pol.Context != ContextConfig {
		return ErrTrust
	}
	return nil
}

func modelFromCfg(name types.AgentModelName, cfg map[string]any, base *types.AgentModel) (types.AgentModel, error) {
	var out types.AgentModel
	if base != nil {
		out = *base
	} else {
		out = types.AgentModel{Name: name}
	}

	if s, ok := getString(cfg, "provider"); ok {
		out.Provider = s
	}
	if s, ok := getString(cfg, "model"); ok {
		out.Model = s
	}
	if s, ok := getString(cfg, "api_key_ref"); ok {
		out.SecretRef = s
	}
	if s, ok := getString(cfg, "base_url"); ok {
		out.BaseURL = s
	}
	if s, ok := getString(cfg, "budget_currency"); ok {
		out.BudgetCurrency = s
	}
	if s, ok := getString(cfg, "notes"); ok {
		out.Notes = s
	}
	if b, ok := getBool(cfg, "disabled"); ok {
		out.Disabled = b
	}

	if out.Provider == "" || out.Model == "" {
		return types.AgentModel{}, fmt.Errorf("'provider' and 'model' are required")
	}
	return out, nil
}

func getString(m map[string]any, k string) (string, bool) {
	if v, ok := m[k].(string); ok {
		return v, true
	}
	return "", false
}

func getBool(m map[string]any, k string) (bool, bool) {
	if v, ok := m[k].(bool); ok {
		return v, true
	}
	return false, false
}
