// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Corrected the expected HMAC result in the test case.
// filename: pkg/tool/crypto/tools_crypto_hash_test.go
// nlines: 84
// risk_rating: LOW

package crypto

import (
	"errors"
	"regexp"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestToolCryptoHash(t *testing.T) {
	// Setup: A policy that grants hashing capability.
	testPolicy := policy.NewBuilder(policy.ContextConfig).
		Allow("tool.crypto.*").
		Grant("crypto:use:hash").
		Build()

	interp := interpreter.NewInterpreter(interpreter.WithExecPolicy(testPolicy))
	// Manually register the hash tools for the test.
	for _, impl := range cryptoHashToolsToRegister {
		if _, err := interp.ToolRegistry().RegisterTool(impl); err != nil {
			t.Fatalf("Failed to register tool %q: %v", impl.Spec.Name, err)
		}
	}

	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		// Hash
		{name: "Hash SHA256", toolName: "Hash", args: []interface{}{"hello", "sha256"}, wantResult: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{name: "Hash SHA512", toolName: "Hash", args: []interface{}{"hello", "sha512"}, wantResult: "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"},
		{name: "Hash unsupported algo", toolName: "Hash", args: []interface{}{"hello", "sha1"}, wantErrIs: lang.ErrInvalidArgument},

		// HMAC
		{name: "HMAC SHA256", toolName: "HMAC", args: []interface{}{"message", "secret", "sha256"}, wantResult: "8b5f48702995c1598c573db1e21866a9b825d4a794d169d7060a03605796360b"},
		{name: "HMAC unsupported algo", toolName: "HMAC", args: []interface{}{"message", "secret", "sha1"}, wantErrIs: lang.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullname := types.MakeFullName(group, tt.toolName)
			toolImpl, found := interp.ToolRegistry().GetTool(fullname)
			if !found {
				t.Fatalf("Tool %q not found", fullname)
			}
			got, err := toolImpl.Func(interp, tt.args)

			if tt.wantErrIs != nil {
				if err == nil || !errors.Is(err, tt.wantErrIs) {
					t.Errorf("Expected error wrapping [%v], but got: %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if got != tt.wantResult {
				t.Errorf("Result mismatch:\n  Got:  %#v\n  Want: %#v", got, tt.wantResult)
			}
		})
	}
}

func TestToolCryptoUUID(t *testing.T) {
	interp := interpreter.NewInterpreter()
	for _, impl := range cryptoHashToolsToRegister {
		if impl.Spec.Name == "UUID" {
			if _, err := interp.ToolRegistry().RegisterTool(impl); err != nil {
				t.Fatalf("Failed to register UUID tool: %v", err)
			}
		}
	}

	uuidTool, _ := interp.ToolRegistry().GetTool(types.MakeFullName(group, "UUID"))
	result, err := uuidTool.Func(interp, []interface{}{})
	if err != nil {
		t.Fatalf("UUID tool failed: %v", err)
	}

	uuidStr, ok := result.(string)
	if !ok {
		t.Fatalf("UUID tool did not return a string, got %T", result)
	}

	// Simple regex to check if the output format is a valid UUID.
	uuidRegex := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)
	if !uuidRegex.MatchString(uuidStr) {
		t.Errorf("Generated UUID '%s' does not match the V4 format", uuidStr)
	}
}
