// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Decrypts secrets based on their encoding type.
// filename: pkg/api/secret/secret.go
// nlines: 36
// risk_rating: HIGH

package secret

import (
	"bytes"
	"errors"
	"io"

	"filippo.io/age"
)

// Ref represents a reference to a secret.
type Ref struct {
	Path string // "prod/db/main"
	Enc  string // "none"|"age"|"sealedbox"
	Raw  []byte // encoded payload
}

// Decode decrypts the secret based on its encoding type.
func Decode(ref Ref, privKey []byte) (string, error) {
	switch ref.Enc {
	case "none":
		return string(ref.Raw), nil
	case "age":
		identity, err := age.ParseX25519Identity(string(privKey))
		if err != nil {
			return "", err
		}

		r, err := age.Decrypt(bytes.NewReader(ref.Raw), identity)
		if err != nil {
			return "", err
		}
		decrypted, err := io.ReadAll(r)
		if err != nil {
			return "", err
		}
		return string(decrypted), nil
	case "sealedbox":
		return "", errors.New("sealedbox encryption is not yet supported")
	default:
		return "", errors.New("unsupported secret encoding")
	}
}
