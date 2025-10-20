// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Corrected test setup to initialize the interpreter with a valid HostContext, fixing a panic.
// filename: pkg/tool/crypto/tools_crypto_security_test.go
// nlines: 101
// risk_rating: HIGH

package crypto

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/golang-jwt/jwt/v5"
)

func TestToolJWTSecurity(t *testing.T) {
	// Setup: A permissive policy for testing tool logic directly.
	testPolicy := policy.NewBuilder(policy.ContextConfig).Allow("tool.crypto.*").Grant("crypto:sign:jwt").Build()
	hostCtx := &interpreter.HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: &bytes.Buffer{},
		Stdin:  &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	interp := interpreter.NewInterpreter(
		interpreter.WithExecPolicy(testPolicy),
		interpreter.WithHostContext(hostCtx),
	)
	// for _, impl := range cryptoToolsToRegister {
	// 	if _, err := interp.ToolRegistry().RegisterTool(impl); err != nil {
	// 		t.Fatalf("Failed to register tool %q: %v", impl.Spec.Name, err)
	// 	}
	// }

	secret := "a-different-secret-for-security-tests"
	signTool, _ := interp.ToolRegistry().GetTool(types.MakeFullName(group, "SignJWT"))
	verifyTool, _ := interp.ToolRegistry().GetTool(types.MakeFullName(group, "VerifyJWT"))
	decodeTool, _ := interp.ToolRegistry().GetTool(types.MakeFullName(group, "DecodeJWT"))

	t.Run("Verify_Expired_Token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub": "user1",
			"exp": time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		}
		res, err := signTool.Func(interp, []interface{}{map[string]interface{}(claims), secret, "HS256"})
		if err != nil {
			t.Fatalf("Failed to sign expired token: %v", err)
		}
		expiredToken := res.(string)

		_, err = verifyTool.Func(interp, []interface{}{expiredToken, secret})
		if !errors.Is(err, lang.ErrInvalidArgument) {
			t.Errorf("Expected ErrInvalidArgument for expired token, got %v", err)
		}
	})

	t.Run("Verify_Not_Yet_Valid_Token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub": "user2",
			"nbf": time.Now().Add(1 * time.Hour).Unix(), // Not valid for 1 hour
		}
		res, err := signTool.Func(interp, []interface{}{map[string]interface{}(claims), secret, "HS256"})
		if err != nil {
			t.Fatalf("Failed to sign NBF token: %v", err)
		}
		nbfToken := res.(string)

		_, err = verifyTool.Func(interp, []interface{}{nbfToken, secret})
		if !errors.Is(err, lang.ErrInvalidArgument) {
			t.Errorf("Expected ErrInvalidArgument for not-yet-valid token, got %v", err)
		}
	})

	t.Run("Malformed_Tokens", func(t *testing.T) {
		testCases := []struct {
			name  string
			token string
		}{
			{"Not a JWT", "hello.world.again"},
			{"Invalid Base64 Header", "aGV?.eyJzdWIiOiIxMjMifQ.SIGNATURE"},
			{"Invalid JSON Payload", "eyJhbGciOiJIUzI1NiJ9.ey-123.SIGNATURE"},
		}
		for _, tc := range testCases {
			_, err := verifyTool.Func(interp, []interface{}{tc.token, secret})
			if !errors.Is(err, lang.ErrInvalidArgument) {
				t.Errorf("[%s] Expected ErrInvalidArgument for VerifyJWT with malformed token, got %v", tc.name, err)
			}
			_, err = decodeTool.Func(interp, []interface{}{tc.token})
			if !errors.Is(err, lang.ErrInvalidArgument) {
				t.Errorf("[%s] Expected ErrInvalidArgument for DecodeJWT with malformed token, got %v", tc.name, err)
			}
		}
	})

	t.Run("Sign_Unsupported_Algorithm", func(t *testing.T) {
		claims := map[string]interface{}{"sub": "user3"}
		_, err := signTool.Func(interp, []interface{}{claims, secret, "HS257"}) // Invalid algorithm
		if !errors.Is(err, lang.ErrInvalidArgument) {
			t.Errorf("Expected ErrInvalidArgument for unsupported algorithm, got %v", err)
		}
	})
}
