// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Tests for JWT cryptographic tools. Corrected tool registration loop.
// filename: pkg/tool/crypto/tools_crypto_test.go
// nlines: 104
// risk_rating: MEDIUM

package crypto

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestToolJWT(t *testing.T) {
	// Setup: Create a policy that allows the crypto tools to run and grants signing capability.
	testPolicy := policy.NewBuilder(policy.ContextConfig).
		Allow("tool.crypto.*").
		Grant("crypto:sign:jwt").
		Build()

	interp := interpreter.NewInterpreter(interpreter.WithExecPolicy(testPolicy))
	for _, impl := range cryptoToolsToRegister {
		if _, err := interp.ToolRegistry().RegisterTool(impl); err != nil {
			t.Fatalf("Failed to register tool %q: %v", impl.Spec.Name, err)
		}
	}

	// --- Test Case Data ---
	claims := map[string]interface{}{"sub": "12345", "nbf": float64(time.Now().Unix())}
	secret := "my-very-secret-key"
	algo := "HS256"

	// 1. Test SignJWT
	signFullName := types.MakeFullName(group, "SignJWT")
	signTool, _ := interp.ToolRegistry().GetTool(signFullName)
	signedResult, err := signTool.Func(interp, []interface{}{claims, secret, algo})
	if err != nil {
		t.Fatalf("SignJWT failed: %v", err)
	}
	signedToken, ok := signedResult.(string)
	if !ok {
		t.Fatalf("SignJWT did not return a string")
	}

	// 2. Test VerifyJWT (Success)
	verifyFullName := types.MakeFullName(group, "VerifyJWT")
	verifyTool, _ := interp.ToolRegistry().GetTool(verifyFullName)
	verifiedResult, err := verifyTool.Func(interp, []interface{}{signedToken, secret})
	if err != nil {
		t.Fatalf("VerifyJWT failed: %v", err)
	}
	verifiedClaims, ok := verifiedResult.(map[string]interface{})
	if !ok {
		t.Fatalf("VerifyJWT did not return a map")
	}
	if verifiedClaims["sub"] != "12345" {
		t.Errorf("Verified claims mismatch: got %v, want %v", verifiedClaims["sub"], "12345")
	}

	// 3. Test VerifyJWT (Failure - wrong secret)
	_, err = verifyTool.Func(interp, []interface{}{signedToken, "wrong-secret"})
	if !errors.Is(err, lang.ErrInvalidArgument) {
		t.Errorf("VerifyJWT should fail with ErrInvalidArgument for wrong secret, but got %v", err)
	}

	// 4. Test DecodeJWT
	decodeFullName := types.MakeFullName(group, "DecodeJWT")
	decodeTool, _ := interp.ToolRegistry().GetTool(decodeFullName)
	decodedResult, err := decodeTool.Func(interp, []interface{}{signedToken})
	if err != nil {
		t.Fatalf("DecodeJWT failed: %v", err)
	}
	decodedClaims, ok := decodedResult.(map[string]interface{})
	if !ok {
		t.Fatalf("DecodeJWT did not return a map")
	}
	// reflect.DeepEqual is needed because of the float64 type from JSON/map decoding
	if !reflect.DeepEqual(decodedClaims, claims) {
		t.Errorf("Decoded claims mismatch:\nGot:  %#v\nWant: %#v", decodedClaims, claims)
	}

	// 5. Test SignJWT without capability
	restrictedPolicy := policy.NewBuilder(policy.ContextConfig).
		Allow("tool.crypto.SignJWT").
		Build() // No Grant this time

	toolMeta := policy.ToolMeta{Name: string(signFullName), RequiredCaps: signTool.RequiredCaps}
	err = restrictedPolicy.CanCall(toolMeta)
	if !errors.Is(err, policy.ErrCapability) {
		t.Errorf("Calling SignJWT without capability should result in ErrCapability, but got %v", err)
	}
}
