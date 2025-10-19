// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Contains tests for the extra string/codec tools. Removed redundant manual tool registration.
// filename: pkg/tool/strtools/tools_string_extra_test.go
// nlines: 100+
// risk_rating: LOW

package strtools

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestToolStringExtraCodecs(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		// BytesFromBase64
		{name: "BytesFromBase64 Simple", toolName: "BytesFromBase64", args: MakeArgs("SGVsbG8gV29ybGQ="), wantResult: "Hello World"},
		{name: "BytesFromBase64 Empty", toolName: "BytesFromBase64", args: MakeArgs(""), wantResult: ""},
		{name: "BytesFromBase64 Invalid Base64", toolName: "BytesFromBase64", args: MakeArgs("???"), wantErrIs: lang.ErrInvalidArgument},
		{name: "BytesFromBase64 Invalid UTF8", toolName: "BytesFromBase64", args: MakeArgs("gA=="), wantErrIs: lang.ErrInvalidArgument}, // Represents invalid UTF-8 byte 0x80
		{name: "BytesFromBase64 Wrong Type", toolName: "BytesFromBase64", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},

		// BytesToBase64
		{name: "BytesToBase64 Simple", toolName: "BytesToBase64", args: MakeArgs("Hello World"), wantResult: "SGVsbG8gV29ybGQ="},
		{name: "BytesToBase64 Empty", toolName: "BytesToBase64", args: MakeArgs(""), wantResult: ""},
		{name: "BytesToBase64 Wrong Type", toolName: "BytesToBase64", args: MakeArgs(true), wantErrIs: lang.ErrArgumentMismatch},

		// ParseFromJsonBase64
		{
			name:       "ParseFromJsonBase64 Map",
			toolName:   "ParseFromJsonBase64",
			args:       MakeArgs("eyJrZXkiOiAidmFsdWUifQ=="), // {"key": "value"}
			wantResult: map[string]interface{}{"key": "value"},
		},
		{
			name:       "ParseFromJsonBase64 List",
			toolName:   "ParseFromJsonBase64",
			args:       MakeArgs("WyIxIiwyLCJ0aHJlZSJd"),        // ["1",2,"three"]
			wantResult: []interface{}{"1", float64(2), "three"}, // Note: JSON numbers become float64
		},
		{
			name:       "ParseFromJsonBase64 Empty JSON Object",
			toolName:   "ParseFromJsonBase64",
			args:       MakeArgs("e30="), // {}
			wantResult: map[string]interface{}{},
		},
		{
			name:       "ParseFromJsonBase64 Empty JSON Array",
			toolName:   "ParseFromJsonBase64",
			args:       MakeArgs("W10="), // []
			wantResult: []interface{}{},
		},
		{name: "ParseFromJsonBase64 Invalid Base64", toolName: "ParseFromJsonBase64", args: MakeArgs("---"), wantErrIs: lang.ErrInvalidArgument},
		{name: "ParseFromJsonBase64 Invalid JSON", toolName: "ParseFromJsonBase64", args: MakeArgs("eyJrZXkiOiB9"), wantErrIs: lang.ErrInvalidArgument}, // {"key": }
		{name: "ParseFromJsonBase64 Wrong Type", toolName: "ParseFromJsonBase64", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},

		// ToJsonString
		{
			name:     "ToJsonString Map",
			toolName: "ToJsonString",
			args:     MakeArgs(map[string]interface{}{"b": 2.0, "a": "one"}),
			// Note: JSON key order is not guaranteed, but Marshal sorts them
			wantResult: `{"a":"one","b":2}`,
		},
		{
			name:       "ToJsonString List",
			toolName:   "ToJsonString",
			args:       MakeArgs([]interface{}{"1", float64(2), true}),
			wantResult: `["1",2,true]`,
		},
		{
			name:       "ToJsonString Empty Map",
			toolName:   "ToJsonString",
			args:       MakeArgs(map[string]interface{}{}),
			wantResult: `{}`,
		},
		{
			name:       "ToJsonString Empty List",
			toolName:   "ToJsonString",
			args:       MakeArgs([]interface{}{}),
			wantResult: `[]`,
		},
		{name: "ToJsonString Wrong Type (String)", toolName: "ToJsonString", args: MakeArgs("just a string"), wantErrIs: lang.ErrArgumentMismatch},
		{name: "ToJsonString Wrong Type (Number)", toolName: "ToJsonString", args: MakeArgs(123.45), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		// Use helper from tools_string_basic_test.go
		// newStringTestInterpreter already ensures ALL tools (including extra) are registered via init()
		interp := newStringTestInterpreter(t)

		// REMOVED Manual registration loop:
		/*
			for _, impl := range stringExtraToolsToRegister {
				if _, err := interp.ToolRegistry().RegisterTool(impl); err != nil {
					t.Fatalf("Failed to register extra tool '%s' for test: %v", impl.Spec.Name, err)
				}
			}
		*/

		// Run the specific test case using the interpreter provided by the helper
		testStringToolHelper(t, interp, tt)
	}
}

// Test result deep comparison might need adjustments for float64 vs int in JSON parsing.
// Using reflect.DeepEqual should generally work for map[string]interface{} and []interface{}.
