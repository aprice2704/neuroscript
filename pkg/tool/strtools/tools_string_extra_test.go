// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 7
// :: description: Tests for extra string/codec tools. Added tests for primitive types in ToJsonString.
// :: latestChange: Added ToJsonString primitive tests.
// :: filename: pkg/tool/strtools/tools_string_extra_test.go
// :: serialization: go

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
		{name: "BytesFromBase64 Invalid UTF8", toolName: "BytesFromBase64", args: MakeArgs("gA=="), wantErrIs: lang.ErrInvalidArgument},
		{name: "BytesFromBase64 Wrong Type", toolName: "BytesFromBase64", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},

		// BytesToBase64
		{name: "BytesToBase64 Simple", toolName: "BytesToBase64", args: MakeArgs("Hello World"), wantResult: "SGVsbG8gV29ybGQ="},
		{name: "BytesToBase64 Empty", toolName: "BytesToBase64", args: MakeArgs(""), wantResult: ""},
		{name: "BytesToBase64 Wrong Type", toolName: "BytesToBase64", args: MakeArgs(true), wantErrIs: lang.ErrArgumentMismatch},

		// ParseFromJsonBase64
		{
			name:       "ParseFromJsonBase64 Map",
			toolName:   "ParseFromJsonBase64",
			args:       MakeArgs("eyJrZXkiOiAidmFsdWUifQ=="),
			wantResult: map[string]interface{}{"key": "value"},
		},
		{
			name:       "ParseFromJsonBase64 List",
			toolName:   "ParseFromJsonBase64",
			args:       MakeArgs("WyIxIiwyLCJ0aHJlZSJd"),
			wantResult: []interface{}{"1", float64(2), "three"},
		},
		{
			name:       "ParseFromJsonBase64 Empty JSON Object",
			toolName:   "ParseFromJsonBase64",
			args:       MakeArgs("e30="),
			wantResult: map[string]interface{}{},
		},
		{
			name:       "ParseFromJsonBase64 Empty JSON Array",
			toolName:   "ParseFromJsonBase64",
			args:       MakeArgs("W10="),
			wantResult: []interface{}{},
		},
		{name: "ParseFromJsonBase64 Invalid Base64", toolName: "ParseFromJsonBase64", args: MakeArgs("---"), wantErrIs: lang.ErrInvalidArgument},
		{name: "ParseFromJsonBase64 Invalid JSON", toolName: "ParseFromJsonBase64", args: MakeArgs("eyJrZXkiOiB9"), wantErrIs: lang.ErrInvalidArgument},
		{name: "ParseFromJsonBase64 Wrong Type", toolName: "ParseFromJsonBase64", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},

		// ParseJsonString
		{
			name:       "ParseJsonString Map",
			toolName:   "ParseJsonString",
			args:       MakeArgs(`{"key": "value"}`),
			wantResult: map[string]interface{}{"key": "value"},
		},
		{
			name:       "ParseJsonString List",
			toolName:   "ParseJsonString",
			args:       MakeArgs(`["1",2,"three",{"nested": true}]`),
			wantResult: []interface{}{"1", float64(2), "three", map[string]interface{}{"nested": true}},
		},
		{
			name:       "ParseJsonString Empty Map",
			toolName:   "ParseJsonString",
			args:       MakeArgs(`{}`),
			wantResult: map[string]interface{}{},
		},
		{
			name:       "ParseJsonString Empty List",
			toolName:   "ParseJsonString",
			args:       MakeArgs(`[]`),
			wantResult: []interface{}{},
		},
		{name: "ParseJsonString Invalid JSON", toolName: "ParseJsonString", args: MakeArgs(`{"key": }`), wantErrIs: lang.ErrInvalidArgument},
		{name: "ParseJsonString Invalid Type", toolName: "ParseJsonString", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
		{name: "ParseJsonString Empty String", toolName: "ParseJsonString", args: MakeArgs(""), wantErrIs: lang.ErrInvalidArgument},

		// ToJsonString (Complex)
		{
			name:       "ToJsonString Map (Compact)",
			toolName:   "ToJsonString",
			args:       MakeArgs(map[string]interface{}{"b": 2.0, "a": "one"}),
			wantResult: `{"a":"one","b":2}`,
		},
		{
			name:       "ToJsonString List (Compact)",
			toolName:   "ToJsonString",
			args:       MakeArgs([]interface{}{"1", float64(2), true}),
			wantResult: `["1",2,true]`,
		},

		// --- ToJsonString Primitives (New Tolerance) ---
		{
			name:       "ToJsonString String",
			toolName:   "ToJsonString",
			args:       MakeArgs("just a string"),
			wantResult: `"just a string"`, // Should wrap in quotes
		},
		{
			name:       "ToJsonString Number",
			toolName:   "ToJsonString",
			args:       MakeArgs(123.45),
			wantResult: `123.45`,
		},
		{
			name:       "ToJsonString Bool",
			toolName:   "ToJsonString",
			args:       MakeArgs(true),
			wantResult: `true`,
		},
		{
			name:       "ToJsonString Nil",
			toolName:   "ToJsonString",
			args:       MakeArgs(nil),
			wantResult: `null`,
		},

		// ToJsonString (Pretty Print)
		{
			name:       "ToJsonString Map (Pretty)",
			toolName:   "ToJsonString",
			args:       MakeArgs(map[string]interface{}{"a": "one"}, true),
			wantResult: "{\n  \"a\": \"one\"\n}",
		},
		{
			name:       "ToJsonString List (Pretty)",
			toolName:   "ToJsonString",
			args:       MakeArgs([]interface{}{"1", 2.0}, true),
			wantResult: "[\n  \"1\",\n  2\n]",
		},
		{
			name:       "ToJsonString Map (Custom Indent)",
			toolName:   "ToJsonString",
			args:       MakeArgs(map[string]interface{}{"a": "one"}, true, "", "\t"),
			wantResult: "{\n\t\"a\": \"one\"\n}",
		},
		{
			name:      "ToJsonString Wrong Arg Type (pretty_print)",
			toolName:  "ToJsonString",
			args:      MakeArgs(map[string]interface{}{}, "not-a-bool"),
			wantErrIs: lang.ErrArgumentMismatch,
		},
	}
	for _, tt := range tests {
		interp := newStringTestInterpreter(t)
		testStringToolHelper(t, interp, tt)
	}
}
