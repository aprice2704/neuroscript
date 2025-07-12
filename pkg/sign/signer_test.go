// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Expanded tests to cover the full Sign and Verify round-trip.
// filename: pkg/sign/signer_test.go
// nlines: 100+
// risk_rating: HIGH

package sign

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/canon"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestSignAndVerify(t *testing.T) {
	// 1. Generate a new, random key pair for this test.
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate ed25519 key pair: %v", err)
	}

	// 2. Create a sample AST to sign.
	script := `func main() means
		emit "hello, signature"
	endfunc`
	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	antlrTree, _ := parserAPI.Parse(script)
	builder := parser.NewASTBuilder(logging.NewNoOpLogger())
	program, _, _ := builder.Build(antlrTree)
	tree := &ast.Tree{Root: program}

	// 3. Canonicalize the AST to get the data to sign.
	blob, sum, err := canon.Canonicalise(tree)
	if err != nil {
		t.Fatalf("Failed to canonicalize AST: %v", err)
	}

	// 4. Sign the canonical data.
	signedAST, err := Sign(privateKey, blob, sum)
	if err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}
	if signedAST == nil {
		t.Fatal("Sign() returned a nil SignedAST")
	}
	if len(signedAST.Sig) == 0 {
		t.Fatal("Sign() produced an empty signature")
	}

	// 5. Perform Verification Tests
	t.Run("successful verification", func(t *testing.T) {
		// This test now uses the actual Verify function.
		// We expect to get a non-nil tree and no error.
		// Note: The Verify function currently returns a placeholder tree.
		verifiedTree, err := Verify(publicKey, signedAST)
		if err != nil {
			t.Errorf("Verification of a valid signature failed: %v", err)
		}
		if verifiedTree == nil {
			t.Error("Verification of a valid signature returned a nil tree")
		}
	})

	t.Run("verification fails with tampered signature", func(t *testing.T) {
		tamperedSignedAST := &SignedAST{
			Blob: signedAST.Blob,
			Sum:  signedAST.Sum,
			Sig:  make([]byte, len(signedAST.Sig)),
		}
		copy(tamperedSignedAST.Sig, signedAST.Sig)
		tamperedSignedAST.Sig[0] ^= 0xff // Flip the first byte

		_, err := Verify(publicKey, tamperedSignedAST)
		if err == nil {
			t.Error("Verification of a tampered signature succeeded, but should have failed")
		}
		if !errors.Is(err, fmt.Errorf("signature verification failed")) {
			// This check is a bit brittle, a sentinel error would be better.
			if err.Error() != "signature verification failed" {
				t.Errorf("Expected signature verification failed error, got: %v", err)
			}
		}
	})

	t.Run("verification fails with tampered message", func(t *testing.T) {
		tamperedBlob := make([]byte, len(signedAST.Blob))
		copy(tamperedBlob, signedAST.Blob)
		tamperedBlob[0] ^= 0xff // Flip the first byte

		tamperedSignedAST := &SignedAST{
			Blob: tamperedBlob,
			Sum:  signedAST.Sum, // The original sum, which no longer matches the tampered blob
			Sig:  signedAST.Sig,
		}

		_, err := Verify(publicKey, tamperedSignedAST)
		if err == nil {
			t.Error("Verification of a tampered message succeeded, but should have failed")
		}
	})

	t.Run("verification fails with wrong public key", func(t *testing.T) {
		wrongPublicKey, _, _ := ed25519.GenerateKey(nil)
		_, err := Verify(wrongPublicKey, signedAST)
		if err == nil {
			t.Error("Verification with wrong public key succeeded, but should have failed")
		}
	})
}
