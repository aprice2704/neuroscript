// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Fixed test logic to correctly check for nil, sentinel, and generic errors.
// filename: pkg/provider/registry_test.go
// nlines: 115
// risk_rating: LOW

package provider_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/provider/test"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	s := provider.NewRegistry()
	admin := provider.NewAdmin(s, &policy.ExecPolicy{Context: policy.ContextConfig})
	reader := provider.NewReader(s)

	// 1. Test empty
	if _, ok := reader.Get("test"); ok {
		t.Fatal("Expected Get to fail on empty registry, but it succeeded")
	}
	if len(reader.List()) != 0 {
		t.Fatal("Expected empty list, but got items")
	}

	// 2. Register
	testProvider := test.New()
	if err := admin.Register("TestProvider", testProvider); err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	// 3. Get (case-insensitive)
	p, ok := reader.Get("testprovider")
	if !ok {
		t.Fatal("Failed to get provider after registration")
	}
	if p != testProvider {
		t.Fatal("Got provider does not match registered provider")
	}

	// 4. List
	list := reader.List()
	if len(list) != 1 || list[0] != "testprovider" {
		t.Fatalf("List() returned incorrect data: %v", list)
	}
}

func TestAdmin_Register(t *testing.T) {
	configPolicy := &policy.ExecPolicy{Context: policy.ContextConfig}
	invalidPolicy := &policy.ExecPolicy{Context: policy.ContextNormal}
	testProvider := test.New()

	testCases := []struct {
		name      string
		adminPol  *policy.ExecPolicy
		setupFunc func(s *provider.Registry)
		provName  string
		provider  any
		wantErrIs error
	}{
		{"Success", configPolicy, nil, "p1", testProvider, nil},
		{"Fail - Duplicate", configPolicy, func(s *provider.Registry) {
			provider.NewAdmin(s, configPolicy).Register("p1", testProvider)
		}, "P1", testProvider, lang.ErrDuplicateKey},
		{"Fail - Nil Provider", configPolicy, nil, "p-nil", nil, lang.ErrInvalidArgument}, // FIX: Use correct sentinel
		{"Fail - Wrong Type", configPolicy, nil, "p-wrong", "not a provider", errors.New("invalid type")},
		{"Fail - Wrong Policy", invalidPolicy, nil, "p-policy", testProvider, policy.ErrTrust},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := provider.NewRegistry()
			if tc.setupFunc != nil {
				tc.setupFunc(s)
			}
			admin := provider.NewAdmin(s, tc.adminPol)
			err := admin.Register(tc.provName, tc.provider)

			// --- FIX: Replaced broken logic with a clear, correct check ---
			if tc.wantErrIs == nil {
				// We wanted success
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			} else {
				// We wanted an error
				if err == nil {
					t.Errorf("Expected error wrapping [%v], but got nil", tc.wantErrIs)
				} else if errors.Is(err, tc.wantErrIs) {
					// This is the success path (e.g., lang.ErrDuplicateKey)
					// Do nothing, test passed.
				} else if tc.wantErrIs.Error() == "invalid type" && strings.Contains(err.Error(), "invalid type") {
					// This is the special case for the "Wrong Type" test.
					// Do nothing, test passed.
				} else {
					// The error didn't match.
					t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantErrIs, err)
				}
			}
			// --- END FIX ---
		})
	}
}

func TestAdmin_Delete(t *testing.T) {
	configPolicy := &policy.ExecPolicy{Context: policy.ContextConfig}
	invalidPolicy := &policy.ExecPolicy{Context: policy.ContextNormal}
	s := provider.NewRegistry()
	admin := provider.NewAdmin(s, configPolicy)
	reader := provider.NewReader(s)
	testProvider := test.New()

	admin.Register("p1", testProvider)
	admin.Register("p2", testProvider)

	// Fail - Wrong Policy
	if adminDel := provider.NewAdmin(s, invalidPolicy); adminDel.Delete("p1") {
		t.Fatal("Delete succeeded with wrong policy context")
	}
	if _, ok := reader.Get("p1"); !ok {
		t.Fatal("Provider was deleted with wrong policy")
	}

	// Fail - Not Found
	if admin.Delete("p3") {
		t.Fatal("Delete returned true for non-existent provider")
	}

	// Success
	if !admin.Delete("p1") {
		t.Fatal("Delete returned false for existing provider")
	}
	if _, ok := reader.Get("p1"); ok {
		t.Fatal("Provider was not deleted")
	}
	if len(reader.List()) != 1 {
		t.Fatal("Provider list is incorrect after delete")
	}
}
