// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements codec and compression tools for the strtools package.
// filename: pkg/tool/strtools/tools_string_codecs.go
// nlines: 121
// risk_rating: MEDIUM

package strtools

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func toolStringToBase64(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "ToBase64: expected 1 argument (input_string)", lang.ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ToBase64: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(inputStr))
	interpreter.GetLogger().Debug("Tool: ToBase64", "input_len", len(inputStr), "output_len", len(encoded))
	return encoded, nil
}

func toolStringFromBase64(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "FromBase64: expected 1 argument (encoded_string)", lang.ErrArgumentMismatch)
	}
	encodedStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("FromBase64: encoded_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	decoded, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("FromBase64: invalid base64 string: %v", err), lang.ErrInvalidArgument)
	}
	interpreter.GetLogger().Debug("Tool: FromBase64", "input_len", len(encodedStr), "output_len", len(decoded))
	return string(decoded), nil
}

func toolStringToHex(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "ToHex: expected 1 argument (input_string)", lang.ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ToHex: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	encoded := hex.EncodeToString([]byte(inputStr))
	interpreter.GetLogger().Debug("Tool: ToHex", "input_len", len(inputStr), "output_len", len(encoded))
	return encoded, nil
}

func toolStringFromHex(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "FromHex: expected 1 argument (encoded_string)", lang.ErrArgumentMismatch)
	}
	encodedStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("FromHex: encoded_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	decoded, err := hex.DecodeString(encodedStr)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("FromHex: invalid hex string: %v", err), lang.ErrInvalidArgument)
	}
	interpreter.GetLogger().Debug("Tool: FromHex", "input_len", len(encodedStr), "output_len", len(decoded))
	return string(decoded), nil
}

func toolStringCompress(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Compress: expected 1 argument (input_string)", lang.ErrArgumentMismatch)
	}
	inputStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Compress: input_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(inputStr)); err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("Compress: failed to write gzip data: %v", err), lang.ErrInternal)
	}
	if err := gz.Close(); err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("Compress: failed to close gzip writer: %v", err), lang.ErrInternal)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	interpreter.GetLogger().Debug("Tool: Compress", "input_len", len(inputStr), "compressed_b64_len", len(encoded))
	return encoded, nil
}

func toolStringDecompress(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Decompress: expected 1 argument (base64_encoded_string)", lang.ErrArgumentMismatch)
	}
	encodedStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Decompress: base64_encoded_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}

	decoded, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("Decompress: invalid base64 string: %v", err), lang.ErrInvalidArgument)
	}

	gz, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("Decompress: failed to create gzip reader: %v", err), lang.ErrInvalidArgument)
	}
	defer gz.Close()

	decompressed, err := io.ReadAll(gz)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("Decompress: failed to read decompressed data: %v", err), lang.ErrInvalidArgument)
	}

	interpreter.GetLogger().Debug("Tool: Decompress", "input_len", len(encodedStr), "decompressed_len", len(decompressed))
	return string(decompressed), nil
}
