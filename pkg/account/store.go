// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Updated config parser to use snake_case keys as per the new standard.
// filename: pkg/account/store.go
// nlines: 135
// risk_rating: HIGH

package account

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

// ---------- store ----------

type Store struct {
	mu sync.RWMutex
	m  map[string]Account // key: lower(name)
}

func NewStore() *Store {
	return &Store{m: make(map[string]Account)}
}

// ---------- reader view ----------

type readerView struct {
	s *Store
}

func NewReader(s *Store) interfaces.AccountReader {
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
	acc, ok := v.s.m[key]
	return acc, ok
}

// ---------- admin view (policy-gated) ----------

type adminView struct {
	s   *Store
	pol *policy.ExecPolicy
}

func NewAdmin(s *Store, pol *policy.ExecPolicy) interfaces.AccountAdmin {
	return &adminView{s: s, pol: pol}
}

func (v *adminView) Register(name string, cfg map[string]any) error {
	if err := v.ensureConfigContext(); err != nil {
		return err
	}
	key := strings.ToLower(name)

	v.s.mu.Lock()
	defer v.s.mu.Unlock()

	if _, exists := v.s.m[key]; exists {
		return lang.ErrDuplicateKey
	}

	acc, err := accountFromCfg(name, cfg)
	if err != nil {
		return err
	}

	v.s.m[key] = acc
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

func accountFromCfg(name string, cfg map[string]any) (Account, error) {
	acc := Account{Name: name}
	var ok bool

	if acc.Kind, ok = cfg["kind"].(string); !ok || acc.Kind == "" {
		return Account{}, fmt.Errorf("'kind' is a required string field: %w", ErrInvalidConfiguration)
	}
	if acc.Provider, ok = cfg["provider"].(string); !ok || acc.Provider == "" {
		return Account{}, fmt.Errorf("'provider' is a required string field: %w", ErrInvalidConfiguration)
	}
	if acc.APIKey, ok = cfg["api_key"].(string); !ok || acc.APIKey == "" {
		return Account{}, fmt.Errorf("'api_key' is a required string field: %w", ErrInvalidConfiguration)
	}

	if val, present := cfg["org_id"]; present {
		acc.OrgID, _ = val.(string)
	}
	if val, present := cfg["project_id"]; present {
		acc.ProjectID, _ = val.(string)
	}
	if val, present := cfg["notes"]; present {
		acc.Notes, _ = val.(string)
	}

	return acc, nil
}
