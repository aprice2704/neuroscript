// NeuroScript Version: 0.7.0
// File version: 8
// Purpose: Updated the config parser to look for 'AccountName' to align with test data.
// filename: pkg/agentmodel/agentmodel_store.go
// nlines: 327
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
	pol *policy.ExecPolicy
}

func NewAgentModelAdmin(s *AgentModelStore, pol *policy.ExecPolicy) interfaces.AgentModelAdmin {
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

// RegisterFromModel provides a type-safe way to register a pre-constructed AgentModel.
func (v *agentModelAdminView) RegisterFromModel(model types.AgentModel) error {
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
	if v.pol == nil || v.pol.Context != policy.ContextConfig {
		return policy.ErrTrust
	}
	return nil
}

func modelFromCfg(name types.AgentModelName, cfg map[string]any, base *types.AgentModel) (types.AgentModel, error) {
	var err error
	var out types.AgentModel
	if base != nil {
		out = *base
	} else {
		out = types.AgentModel{Name: name}
	}

	// Top-level fields
	if out.Provider, err = getString(cfg, "provider", out.Provider); err != nil {
		return out, err
	}
	if out.Model, err = getString(cfg, "model", out.Model); err != nil {
		return out, err
	}
	if out.AccountName, err = getString(cfg, "AccountName", out.AccountName); err != nil {
		return out, err
	}
	if out.BaseURL, err = getString(cfg, "base_url", out.BaseURL); err != nil {
		return out, err
	}
	if out.BudgetCurrency, err = getString(cfg, "budget_currency", out.BudgetCurrency); err != nil {
		return out, err
	}
	if out.Notes, err = getString(cfg, "notes", out.Notes); err != nil {
		return out, err
	}
	if out.Disabled, err = getBool(cfg, "disabled", out.Disabled); err != nil {
		return out, err
	}
	if out.ContextKTok, err = getInt(cfg, "context_ktok", out.ContextKTok); err != nil {
		return out, err
	}
	if out.MaxTurns, err = getInt(cfg, "max_turns", out.MaxTurns); err != nil {
		return out, err
	}
	if out.MaxRetries, err = getInt(cfg, "max_retries", out.MaxRetries); err != nil {
		return out, err
	}

	// GenerationConfig
	gcfg := &out.Generation
	if gcfg.Temperature, err = getFloat64(cfg, "temperature", gcfg.Temperature); err != nil {
		return out, err
	}
	if gcfg.TopP, err = getFloat64(cfg, "top_p", gcfg.TopP); err != nil {
		return out, err
	}
	if gcfg.TopK, err = getInt(cfg, "top_k", gcfg.TopK); err != nil {
		return out, err
	}
	if gcfg.MaxOutputTokens, err = getInt(cfg, "max_output_tokens", gcfg.MaxOutputTokens); err != nil {
		return out, err
	}
	if gcfg.StopSequences, err = getStringSlice(cfg, "stop_sequences", gcfg.StopSequences); err != nil {
		return out, err
	}
	if gcfg.PresencePenalty, err = getFloat64(cfg, "presence_penalty", gcfg.PresencePenalty); err != nil {
		return out, err
	}
	if gcfg.FrequencyPenalty, err = getFloat64(cfg, "frequency_penalty", gcfg.FrequencyPenalty); err != nil {
		return out, err
	}
	if gcfg.RepetitionPenalty, err = getFloat64(cfg, "repetition_penalty", gcfg.RepetitionPenalty); err != nil {
		return out, err
	}
	if gcfg.Seed, err = getInt64Ptr(cfg, "seed", gcfg.Seed); err != nil {
		return out, err
	}
	if gcfg.LogProbs, err = getBool(cfg, "log_probs", gcfg.LogProbs); err != nil {
		return out, err
	}
	if s, err := getString(cfg, "response_format", string(gcfg.ResponseFormat)); err != nil {
		return out, err
	} else {
		gcfg.ResponseFormat = types.ResponseFormat(s)
	}

	// ToolConfig
	tcfg := &out.Tools
	if tcfg.ToolLoopPermitted, err = getBool(cfg, "tool_loop_permitted", tcfg.ToolLoopPermitted); err != nil {
		return out, err
	}
	if tcfg.AutoLoopEnabled, err = getBool(cfg, "auto_loop_enabled", tcfg.AutoLoopEnabled); err != nil {
		return out, err
	}
	if s, err := getString(cfg, "tool_choice", string(tcfg.ToolChoice)); err != nil {
		return out, err
	} else {
		tcfg.ToolChoice = types.ToolChoice(s)
	}

	// SafetyConfig
	scfg := &out.Safety
	if scfg.SafePrompt, err = getBool(cfg, "safe_prompt", scfg.SafePrompt); err != nil {
		return out, err
	}
	if scfg.Settings, err = getStringMap(cfg, "safety_settings", scfg.Settings); err != nil {
		return out, err
	}

	// --- Handle Deprecated Fields for backward compatibility ---
	out.Temperature = out.Generation.Temperature
	out.ToolLoopPermitted = out.Tools.ToolLoopPermitted
	out.AutoLoopEnabled = out.Tools.AutoLoopEnabled
	// ---

	if out.Provider == "" || out.Model == "" {
		return types.AgentModel{}, fmt.Errorf("'provider' and 'model' are required")
	}
	return out, nil
}

func typeError(k string, want string, got interface{}) error {
	return fmt.Errorf("field '%s' has wrong type: expected %s, got %T", k, want, got)
}

func getString(m map[string]any, k string, defaultVal string) (string, error) {
	v, ok := m[k]
	if !ok {
		return defaultVal, nil
	}
	if s, isString := v.(string); isString {
		return s, nil
	}
	return "", typeError(k, "string", v)
}

func getBool(m map[string]any, k string, defaultVal bool) (bool, error) {
	v, ok := m[k]
	if !ok {
		return defaultVal, nil
	}
	if b, isBool := v.(bool); isBool {
		return b, nil
	}
	return false, typeError(k, "bool", v)
}

func getInt(m map[string]any, k string, defaultVal int) (int, error) {
	v, ok := m[k]
	if !ok {
		return defaultVal, nil
	}
	if f, isFloat := v.(float64); isFloat { // ns numbers are float64
		return int(f), nil
	}
	if i, isInt := v.(int); isInt {
		return i, nil
	}
	return 0, typeError(k, "number", v)
}

func getFloat64(m map[string]any, k string, defaultVal float64) (float64, error) {
	v, ok := m[k]
	if !ok {
		return defaultVal, nil
	}
	if f, isFloat := v.(float64); isFloat {
		return f, nil
	}
	if i, isInt := v.(int); isInt { // Coerce int to float64
		return float64(i), nil
	}
	return 0, typeError(k, "number", v)
}

func getStringSlice(m map[string]any, k string, defaultVal []string) ([]string, error) {
	v, ok := m[k]
	if !ok {
		return defaultVal, nil
	}
	if slice, isSlice := v.([]interface{}); isSlice {
		out := make([]string, len(slice))
		for i, item := range slice {
			if s, isString := item.(string); isString {
				out[i] = s
			} else {
				return nil, typeError(fmt.Sprintf("%s[%d]", k, i), "string", item)
			}
		}
		return out, nil
	}
	return nil, typeError(k, "slice", v)
}

func getInt64Ptr(m map[string]any, k string, defaultVal *int64) (*int64, error) {
	v, ok := m[k]
	if !ok {
		return defaultVal, nil
	}
	if v == nil {
		return nil, nil
	}
	if f, isFloat := v.(float64); isFloat {
		val := int64(f)
		return &val, nil
	}
	if i, isInt := v.(int); isInt {
		val := int64(i)
		return &val, nil
	}
	return nil, typeError(k, "number or nil", v)
}

func getStringMap(m map[string]any, k string, defaultVal map[string]string) (map[string]string, error) {
	v, ok := m[k]
	if !ok {
		return defaultVal, nil
	}
	if m, isMap := v.(map[string]interface{}); isMap {
		out := make(map[string]string)
		for key, val := range m {
			if s, isString := val.(string); isString {
				out[key] = s
			} else {
				return nil, typeError(fmt.Sprintf("%s['%s']", k, key), "string", val)
			}
		}
		return out, nil
	}
	return nil, typeError(k, "map", v)
}
