// NeuroScript Version: 0.7.0
// File version: 14
// Purpose: Adds a public, thread-safe Exists(string) bool method to AgentModelStore.
// filename: pkg/agentmodel/agentmodel_store.go
// nlines: 159
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

// --- ADDED: Public Exists method ---

// Exists checks if an agent model with the given name is registered.
// This is thread-safe and can be called by other Go services.
func (s *AgentModelStore) Exists(name string) bool {
	key := strings.ToLower(name)
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.m[key]
	return ok
}

// --- END ADDED ---

// ---------- reader view ----------

type agentModelReaderView struct {
	s *AgentModelStore
}

func NewAgentModelReader(s *AgentModelStore) interfaces.AgentModelReader {
	return &agentModelReaderView{s: s}
}

func (v *agentModelReaderView) List() []string {
	v.s.mu.RLock()
	defer v.s.mu.RUnlock()
	out := make([]string, 0, len(v.s.m))
	for _, model := range v.s.m {
		out = append(out, string(model.Name)) // Convert types.AgentModelName to string
	}
	return out
}

func (v *agentModelReaderView) Get(name string) (any, bool) {
	key := strings.ToLower(name)
	v.s.mu.RLock()
	defer v.s.mu.RUnlock()
	model, ok := v.s.m[key]
	return model, ok
}

// ---------- admin view (policy-gated) ----------

type agentModelAdminView struct {
	s   *AgentModelStore
	pol *policy.ExecPolicy
}

func NewAgentModelAdmin(s *AgentModelStore, pol *policy.ExecPolicy) interfaces.AgentModelAdmin {
	return &agentModelAdminView{s: s, pol: pol}
}

func (v *agentModelAdminView) List() []string { return NewAgentModelReader(v.s).List() }

func (v *agentModelAdminView) Get(name string) (any, bool) {
	return NewAgentModelReader(v.s).Get(name)
}

func (v *agentModelAdminView) Register(name string, cfg map[string]any) error {
	if err := v.ensureConfigContext(); err != nil {
		return err
	}
	key := strings.ToLower(name)
	modelName := types.AgentModelName(name) // Convert string to types.AgentModelName

	v.s.mu.Lock()
	defer v.s.mu.Unlock()

	if _, exists := v.s.m[key]; exists {
		return lang.ErrDuplicateKey
	}

	model, err := modelFromCfg(modelName, cfg, nil)
	if err != nil {
		return err
	}

	v.s.m[key] = model
	return nil
}

// RegisterFromModel (any) implements the interface.
func (v *agentModelAdminView) RegisterFromModel(model any) error {
	modelStruct, ok := model.(types.AgentModel)
	if !ok {
		return fmt.Errorf("invalid type for RegisterFromModel: expected types.AgentModel, got %T", model)
	}
	return v.registerFromModelInternal(modelStruct) // Call the type-safe internal method
}

// registerFromModelInternal provides a type-safe way for the host to register a pre-constructed AgentModel.
func (v *agentModelAdminView) registerFromModelInternal(model types.AgentModel) error {
	if err := v.ensureConfigContext(); err != nil {
		return err
	}
	key := strings.ToLower(string(model.Name))

	v.s.mu.Lock()
	defer v.s.mu.Unlock()

	if _, exists := v.s.m[key]; exists {
		return lang.ErrDuplicateKey
	}

	v.s.m[key] = model
	return nil
}

func (v *agentModelAdminView) Update(name string, updates map[string]any) error {
	if err := v.ensureConfigContext(); err != nil {
		return err
	}
	key := strings.ToLower(name)
	modelName := types.AgentModelName(name) // Convert string to types.AgentModelName

	v.s.mu.Lock()
	defer v.s.mu.Unlock()

	cur, ok := v.s.m[key]
	if !ok {
		return lang.ErrNotFound
	}

	model, err := modelFromCfg(modelName, updates, &cur)
	if err != nil {
		return err
	}

	v.s.m[key] = model
	return nil
}

func (v *agentModelAdminView) Delete(name string) bool {
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

func (v *agentModelAdminView) ensureConfigContext() error {
	if v.pol == nil || v.pol.Context != policy.ContextConfig {
		return policy.ErrTrust
	}
	return nil
}

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
	// Only overwrite nested fields if the deprecated top-level field was explicitly in the input map.
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
