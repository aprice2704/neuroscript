// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines the secret decoding logic stub.
// filename: pkg/secret/decoder.go
// nlines: 30
// risk_rating: HIGH

package secret

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

// ErrSecretUnsupported is returned when the encryption type of a secret is not supported.
var ErrSecretUnsupported = errors.New("unsupported secret encryption type")

// Decode decrypts the value of a secret reference based on its encryption type.
// For now, it only supports unencrypted ("none") secrets.
// The `priv` argument is a placeholder for the private key needed for decryption.
func Decode(ref *ast.SecretRef, priv []byte) (string, error) {
	if ref == nil {
		return "", fmt.Errorf("cannot decode a nil secret reference")
	}

	switch ref.Enc {
	case "none":
		// For "none", the raw value is the secret itself.
		return string(ref.Raw), nil
	case "age", "sealedbox":
		// These are placeholders for future implementation.
		return "", fmt.Errorf("%w: %s", ErrSecretUnsupported, ref.Enc)
	default:
		return "", fmt.Errorf("%w: '%s'", ErrSecretUnsupported, ref.Enc)
	}
}
