// NeuroScript Version: 0.8.0
// File version: 5
// Purpose: Implements extra string/codec tools. Updated ToJsonString to support pretty_print, prefix, and indent. Added ParseJsonString.
// filename: pkg/tool/strtools/tools_string_extra.go
// nlines: 168
// risk_rating: MEDIUM

package strtools

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolBytesFromBase64 converts a base64 encoded string (representing bytes) to a UTF-8 string.
func toolBytesFromBase64(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "BytesFromBase64: expected 1 argument (base64_string)", lang.ErrArgumentMismatch)
	}
	base64Data, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("BytesFromBase64: base64_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}

	byteData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("BytesFromBase64: invalid base64 input: %v", err), lang.ErrInvalidArgument)
	}

	if !utf8.Valid(byteData) {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, "BytesFromBase64: byte data is not valid UTF-8", lang.ErrInvalidArgument)
	}

	interpreter.GetLogger().Debug("Tool: BytesFromBase64", "input_len", len(base64Data), "output_len", len(byteData))
	return string(byteData), nil
}

// toolBytesToBase64 converts a string to a base64 encoded string (representing bytes).
func toolBytesToBase64(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "BytesToBase64: expected 1 argument (string_data)", lang.ErrArgumentMismatch)
	}
	stringData, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("BytesToBase64: string_data argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}

	byteData := []byte(stringData)
	base64Data := base64.StdEncoding.EncodeToString(byteData)

	interpreter.GetLogger().Debug("Tool: BytesToBase64", "input_len", len(stringData), "output_len", len(base64Data))
	return base64Data, nil
}

// toolParseFromJsonBase64 parses a JSON object from a base64 encoded string (representing bytes).
func toolParseFromJsonBase64(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "ParseFromJsonBase64: expected 1 argument (base64_string)", lang.ErrArgumentMismatch)
	}
	base64Data, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ParseFromJsonBase64: base64_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}

	byteData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("ParseFromJsonBase64: invalid base64 input: %v", err), lang.ErrInvalidArgument)
	}

	var parsedValue interface{}
	// Important: Unmarshal into interface{} to handle both maps and lists dynamically.
	if err := json.Unmarshal(byteData, &parsedValue); err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("ParseFromJsonBase64: invalid JSON data: %v", err), lang.ErrInvalidArgument)
	}

	// lang.Wrap should handle the conversion to NeuroScript maps/lists if necessary.
	interpreter.GetLogger().Debug("Tool: ParseFromJsonBase64", "input_len", len(base64Data))
	return parsedValue, nil
}

// toolParseJsonString parses a JSON object from a plain string.
func toolParseJsonString(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "ParseJsonString: expected 1 argument (json_string)", lang.ErrArgumentMismatch)
	}
	jsonString, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ParseJsonString: json_string argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
	}

	var parsedValue interface{}
	// Important: Unmarshal into interface{} to handle both maps and lists dynamically.
	if err := json.Unmarshal([]byte(jsonString), &parsedValue); err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("ParseJsonString: invalid JSON data: %v", err), lang.ErrInvalidArgument)
	}

	interpreter.GetLogger().Debug("Tool: ParseJsonString", "input_len", len(jsonString))
	return parsedValue, nil
}

// toolToJsonString converts a NeuroScript map/list (passed as interface{}) to a JSON string.
func toolToJsonString(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 4 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "ToJsonString: expected 1 to 4 arguments (value, [pretty_print], [prefix], [indent])", lang.ErrArgumentMismatch)
	}
	value := args[0] // Value is already unwrapped interface{}

	// --- Argument Coercion ---
	prettyPrint := false
	if len(args) > 1 && args[1] != nil {
		var ok bool
		prettyPrint, ok = args[1].(bool)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ToJsonString: pretty_print argument must be a boolean, got %T", args[1]), lang.ErrArgumentMismatch)
		}
	}

	prefix := ""
	if len(args) > 2 && args[2] != nil {
		var ok bool
		prefix, ok = args[2].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ToJsonString: prefix argument must be a string, got %T", args[2]), lang.ErrArgumentMismatch)
		}
	}

	indent := "  " // Default indent
	if len(args) > 3 && args[3] != nil {
		var ok bool
		indent, ok = args[3].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ToJsonString: indent argument must be a string, got %T", args[3]), lang.ErrArgumentMismatch)
		}
	}
	// --- End Argument Coercion ---

	// Ensure the input is actually a map or slice, as expected by the spec.
	switch value.(type) {
	case map[string]interface{}, []interface{}:
		// Okay, proceed
	default:
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ToJsonString: expected a map or list argument, got %T", value), lang.ErrArgumentMismatch)
	}

	var jsonData []byte
	var err error

	if prettyPrint {
		jsonData, err = json.MarshalIndent(value, prefix, indent)
	} else {
		jsonData, err = json.Marshal(value)
	}

	if err != nil {
		// This can happen with complex types not representable in JSON
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("ToJsonString: failed to marshal value to JSON: %v", err), lang.ErrInvalidArgument)
	}

	jsonString := string(jsonData)
	interpreter.GetLogger().Debug("Tool: ToJsonString", "output_len", len(jsonString), "pretty", prettyPrint)
	return jsonString, nil
}
