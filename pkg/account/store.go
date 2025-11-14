// NeuroScript Version: 0.7.0
// File version: 7
// Purpose: Updated adminView to embed readerView, satisfying the updated AccountAdmin interface.
// Latest change: Embed *readerView in adminView and update NewAdmin.
// filename: pkg/account/store.go
// nlines: 164
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
	*readerView // FIX: Embed readerView to satisfy AccountReader part of the interface
	pol         *policy.ExecPolicy
}

func NewAdmin(s *Store, pol *policy.ExecPolicy) interfaces.AccountAdmin {
	// FIX: Construct the embedded readerView
	return &adminView{
		readerView: &readerView{s: s},
		pol:        pol,
	}
}

func (v *adminView) Register(name string, cfg map[string]any) error {
	if err := v.ensureConfigContext(); err != nil {
		return err
	}
	key := strings.ToLower(name)

	// v.s is promoted from the embedded readerView
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

// RegisterFromAccount implements the interface method, accepting 'any'
// to break the import cycle.
func (v *adminView) RegisterFromAccount(acc any) error {
	if err := v.ensureConfigContext(); err != nil {
		return err
	}

	accStruct, ok := acc.(Account)
	if !ok {
		return fmt.Errorf("invalid type for RegisterFromAccount: expected account.Account, got %T", acc)
	}

	key := strings.ToLower(accStruct.Name)
	if key == "" {
		return fmt.Errorf("account name cannot be empty: %w", ErrInvalidConfiguration)
	}

	// v.s is promoted from the embedded readerView
	v.s.mu.Lock()
	defer v.s.mu.Unlock()

	if _, exists := v.s.m[key]; exists {
		return lang.ErrDuplicateKey
	}

	v.s.m[key] = accStruct
	return nil
}

func (v *adminView) Delete(name string) bool {
	if err := v.ensureConfigContext(); err != nil {
		return false
	}
	key := strings.ToLower(name)

	// v.s is promoted from the embedded readerView
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
