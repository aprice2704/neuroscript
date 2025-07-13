// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Corrected to use the public API functions instead of internal packages.
// filename: pkg/api/e2e_test.go
// nlines: 70
// risk_rating: HIGH

package api

import (
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/analysis"
	"github.com/aprice2704/neuroscript/pkg/interp"
	"github.com/aprice2704/neuroscript/pkg/sign"
)

func TestEndToEndSmoke(t *testing.T) {
	// 1. Read a test script
	script := `command
		emit "hello world"
	endcommand`

	// 2. Parse it to a Tree using the public API
	// CORRECTED: Use api.Parse, not the internal parser.
	tree, err := Parse([]byte(script), ParseSkipComments)
	if err != nil {
		t.Fatalf("E2E Test: api.Parse failed unexpectedly: %v", err)
	}

	// 3. Canonicalize and Sign it using the public API
	publicKey, privateKey, _ := ed25519.GenerateKey(nil)
	// CORRECTED: Use api.Canonicalise (no package prefix needed).
	blob, sum, err := Canonicalise(tree)
	if err != nil {
		t.Fatalf("E2E Test: api.Canonicalise failed: %v", err)
	}

	// The sign package is an internal detail for now.
	signedAST, err := sign.Sign(privateKey, blob, sum)
	if err != nil {
		t.Fatalf("E2E Test: Sign failed: %v", err)
	}

	// 4. Verify the signature
	verifiedTree, err := sign.Verify(publicKey, signedAST)
	if err != nil {
		t.Fatalf("E2E Test: Verify failed: %v", err)
	}
	if verifiedTree == nil {
		t.Fatal("E2E Test: Verified tree is nil")
	}

	// 5. Vet the AST using analysis passes
	// This uses the api/analysis sub-package, which is correct.
	diags := analysis.Vet(verifiedTree)
	if len(diags) > 0 {
		t.Fatalf("E2E Test: Vetting failed with %d diagnostics: %v", len(diags), diags)
	}

	// 6. Execute the command
	// This uses the internal interpreter, which is expected for this integration test.
	cfg := interp.Config{}
	result, err := interp.ExecCommand(context.Background(), verifiedTree, cfg)
	if err != nil {
		t.Fatalf("E2E Test: ExecCommand failed: %v", err)
	}

	// 7. Check for the expected output
	expectedOutput := "execution successful (shim)"
	if result.Output != expectedOutput {
		t.Errorf("E2E Test: Expected output '%s', but got '%s'", expectedOutput, result.Output)
	}
}
