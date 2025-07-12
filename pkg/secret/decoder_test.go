// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Provides tests for the secret decoding logic.
// filename: pkg/secret/decoder_test.go
// nlines: 60
// risk_rating: HIGH

package secret

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestDecode(t *testing.T) {
	t.Run("successful decode for 'none' encryption", func(t *testing.T) {
		secretText := "my-secret-value"
		ref := &ast.SecretRef{
			Enc: "none",
			Raw: []byte(secretText),
		}

		decoded, err := Decode(ref, nil)
		if err != nil {
			t.Fatalf("Decode() with 'none' encryption failed unexpectedly: %v", err)
		}
		if decoded != secretText {
			t.Errorf("Expected decoded secret to be '%s', but got '%s'", secretText, decoded)
		}
	})

	t.Run("unsupported encryption type 'age'", func(t *testing.T) {
		ref := &ast.SecretRef{
			Enc: "age",
			Raw: []byte("some-encrypted-data"),
		}

		_, err := Decode(ref, nil)
		if err == nil {
			t.Fatal("Expected an error for unsupported encryption type 'age', but got nil")
		}
		if !errors.Is(err, ErrSecretUnsupported) {
			t.Errorf("Expected error to be ErrSecretUnsupported, but got: %v", err)
		}
	})

	t.Run("unsupported encryption type 'sealedbox'", func(t *testing.T) {
		ref := &ast.SecretRef{
			Enc: "sealedbox",
			Raw: []byte("some-other-encrypted-data"),
		}

		_, err := Decode(ref, nil)
		if err == nil {
			t.Fatal("Expected an error for unsupported encryption type 'sealedbox', but got nil")
		}
		if !errors.Is(err, ErrSecretUnsupported) {
			t.Errorf("Expected error to be ErrSecretUnsupported, but got: %v", err)
		}
	})

	t.Run("unknown encryption type", func(t *testing.T) {
		ref := &ast.SecretRef{
			Enc: "unknown-future-encryption",
			Raw: []byte("data"),
		}
		_, err := Decode(ref, nil)
		if err == nil {
			t.Fatal("Expected an error for unknown encryption type, but got nil")
		}
		if !errors.Is(err, ErrSecretUnsupported) {
			t.Errorf("Expected error to be ErrSecretUnsupported, but got: %v", err)
		}
	})

	t.Run("nil secret reference", func(t *testing.T) {
		_, err := Decode(nil, nil)
		if err == nil {
			t.Fatal("Expected an error for nil secret reference, but got nil")
		}
	})
}
