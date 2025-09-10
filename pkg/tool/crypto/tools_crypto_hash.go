// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Implements hashing, HMAC, and UUID generation tools. Corrected HMAC panic.
// filename: pkg/tool/crypto/tools_crypto_hash.go
// nlines: 79
// risk_rating: HIGH

package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/google/uuid"
)

func getHash(algo string) (hash.Hash, error) {
	switch strings.ToLower(algo) {
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	case "md5":
		return md5.New(), nil
	default:
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("unsupported hash algorithm: %s", algo), lang.ErrInvalidArgument)
	}
}

func toolHash(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Hash: expected 2 arguments (input_string, algorithm)", lang.ErrArgumentMismatch)
	}
	input, ok1 := args[0].(string)
	algo, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "Hash: arguments must be strings", lang.ErrArgumentMismatch)
	}

	h, err := getHash(algo)
	if err != nil {
		return nil, err
	}
	h.Write([]byte(input))
	result := hex.EncodeToString(h.Sum(nil))
	interpreter.GetLogger().Debug("Tool: Hash", "algorithm", algo)
	return result, nil
}

func toolHMAC(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "HMAC: expected 3 arguments (input_string, secret_key, algorithm)", lang.ErrArgumentMismatch)
	}
	input, ok1 := args[0].(string)
	secret, ok2 := args[1].(string)
	algo, ok3 := args[2].(string)
	if !ok1 || !ok2 || !ok3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "HMAC: arguments must be strings", lang.ErrArgumentMismatch)
	}

	var hf func() hash.Hash
	switch strings.ToLower(algo) {
	case "sha256":
		hf = sha256.New
	case "sha512":
		hf = sha512.New
	case "md5":
		hf = md5.New
	default:
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("unsupported hash algorithm: %s", algo), lang.ErrInvalidArgument)
	}

	h := hmac.New(hf, []byte(secret))
	h.Write([]byte(input))
	result := hex.EncodeToString(h.Sum(nil))
	interpreter.GetLogger().Debug("Tool: HMAC", "algorithm", algo)
	return result, nil
}

func toolUUID(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "UUID: expected 0 arguments", lang.ErrArgumentMismatch)
	}
	id := uuid.New().String()
	interpreter.GetLogger().Debug("Tool: UUID", "generated_id", id)
	return id, nil
}
