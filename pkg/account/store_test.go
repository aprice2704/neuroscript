// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Updated test configurations to use snake_case keys.
// filename: pkg/account/store_test.go
// nlines: 191
// risk_rating: LOW

package account_test

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

func newValidConfig() map[string]interface{} {
	return map[string]interface{}{
		"kind":       "llm",
		"provider":   "test-provider",
		"api_key":    "key-12345",
		"org_id":     "org-abc",
		"project_id": "proj-xyz",
		"notes":      "A test account",
	}
}

func TestStoreAndReader(t *testing.T) {
	s := account.NewStore()
	if s == nil {
		t.Fatal("NewStore() returned nil")
	}

	reader := account.NewReader(s)
	admin := account.NewAdmin(s, &policy.ExecPolicy{Context: policy.ContextConfig})

	// 1. Test empty state
	if len(reader.List()) != 0 {
		t.Errorf("Expected empty list on new store, got %d items", len(reader.List()))
	}
	if _, found := reader.Get("any"); found {
		t.Error("Expected Get to fail on empty store, but it found an item")
	}

	// 2. Populate the store
	if err := admin.Register("TestAcc1", newValidConfig()); err != nil {
		t.Fatalf("Failed to register TestAcc1: %v", err)
	}
	cfg2 := newValidConfig()
	cfg2["provider"] = "other"
	if err := admin.Register("TestAcc2", cfg2); err != nil {
		t.Fatalf("Failed to register TestAcc2: %v", err)
	}

	// 3. Test List
	names := reader.List()
	sort.Strings(names)
	expectedNames := []string{"testacc1", "testacc2"}
	if !reflect.DeepEqual(names, expectedNames) {
		t.Errorf("List() mismatch.\nGot:    %v\nWanted: %v", names, expectedNames)
	}

	// 4. Test Get
	t.Run("Get", func(t *testing.T) {
		// Case-insensitive check
		accAny, found := reader.Get("tEsTaCc1")
		if !found {
			t.Fatal("Expected to find 'tEsTaCc1', but didn't")
		}
		acc, ok := accAny.(account.Account)
		if !ok {
			t.Fatalf("Expected Get to return account.Account, got %T", accAny)
		}
		if acc.Provider != "test-provider" {
			t.Errorf("Expected provider 'test-provider', got %q", acc.Provider)
		}

		_, notFound := reader.Get("nonexistent")
		if notFound {
			t.Error("Expected not to find 'nonexistent', but did")
		}
	})
}

func TestAdminView_Register(t *testing.T) {
	configPolicy := &policy.ExecPolicy{Context: policy.ContextConfig}
	// Use a different, valid context for the negative test case.
	invalidContextPolicy := &policy.ExecPolicy{Context: "invalid-context"}

	testCases := []struct {
		name      string
		adminPol  *policy.ExecPolicy
		setupFunc func(t *testing.T, s *account.Store)
		accName   string
		accCfg    map[string]interface{}
		wantErrIs error
	}{
		{
			name:      "Success",
			adminPol:  configPolicy,
			accName:   "new-acc",
			accCfg:    newValidConfig(),
			wantErrIs: nil,
		},
		{
			name:     "Fail - Duplicate Key",
			adminPol: configPolicy,
			setupFunc: func(t *testing.T, s *account.Store) {
				admin := account.NewAdmin(s, configPolicy)
				if err := admin.Register("new-acc", newValidConfig()); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			},
			accName:   "NEW-ACC", // case-insensitive duplicate
			accCfg:    newValidConfig(),
			wantErrIs: lang.ErrDuplicateKey,
		},
		{
			name:      "Fail - Invalid Config (missing kind)",
			adminPol:  configPolicy,
			accName:   "bad-acc",
			accCfg:    map[string]interface{}{"provider": "p", "api_key": "k"},
			wantErrIs: account.ErrInvalidConfiguration,
		},
		{
			name:      "Fail - Wrong Policy Context",
			adminPol:  invalidContextPolicy,
			accName:   "wrong-context-acc",
			accCfg:    newValidConfig(),
			wantErrIs: policy.ErrTrust,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := account.NewStore()
			admin := account.NewAdmin(s, tc.adminPol)

			if tc.setupFunc != nil {
				tc.setupFunc(t, s)
			}

			err := admin.Register(tc.accName, tc.accCfg)

			if !errors.Is(err, tc.wantErrIs) {
				t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantErrIs, err)
			}
		})
	}
}

func TestAdminView_Delete(t *testing.T) {
	configPolicy := &policy.ExecPolicy{Context: policy.ContextConfig}
	s := account.NewStore()
	admin := account.NewAdmin(s, configPolicy)
	reader := account.NewReader(s)

	if err := admin.Register("acc-to-delete", newValidConfig()); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Test deleting non-existent item
	if admin.Delete("nonexistent") {
		t.Error("Expected Delete to return false for non-existent item, but it returned true")
	}

	// Test deleting existing item (case-insensitively)
	if !admin.Delete("ACC-TO-DELETE") {
		t.Error("Expected Delete to return true for existing item, but it returned false")
	}

	// Verify it's gone
	if _, found := reader.Get("acc-to-delete"); found {
		t.Error("Expected account to be deleted, but it was still found")
	}

	// Test wrong policy context
	invalidContextPolicy := &policy.ExecPolicy{Context: "invalid-context"}
	runtimeAdmin := account.NewAdmin(s, invalidContextPolicy)
	if err := admin.Register("another-acc", newValidConfig()); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	if runtimeAdmin.Delete("another-acc") {
		t.Error("Expected Delete under wrong policy context to return false, but it was true")
	}
	if _, found := reader.Get("another-acc"); !found {
		t.Error("Account was deleted under wrong policy context")
	}
}
