// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: Contains tests for the extra string/codec tools. Added tests for ToJsonString pretty_print, prefix, and indent. Added tests for ParseJsonString.
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

		// --- New ParseJsonString Tests ---
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
		// --- End ParseJsonString Tests ---

		// ToJsonString
		{
			name:       "ToJsonString Map (Compact, 1 arg)",
			toolName:   "ToJsonString",
			args:       MakeArgs(map[string]interface{}{"b": 2.0, "a": "one"}),
			wantResult: `{"a":"one","b":2}`,
		},
		{
			name:       "ToJsonString List (Compact, 1 arg)",
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

		// --- New Pretty-Print Tests ---
		{
			name:       "ToJsonString Map (Pretty=true, 2 args)",
			toolName:   "ToJsonString",
			args:       MakeArgs(map[string]interface{}{"a": "one"}, true),
			wantResult: "{\n  \"a\": \"one\"\n}", // Default prefix="", indent="  "
		},
		{
			name:       "ToJsonString List (Pretty=true, 2 args)",
			toolName:   "ToJsonString",
			args:       MakeArgs([]interface{}{"1", 2.0}, true),
			wantResult: "[\n  \"1\",\n  2\n]", // Default prefix="", indent="  "
		},
		{
			name:       "ToJsonString Map (Pretty=true, 4 args, custom indent)",
			toolName:   "ToJsonString",
			args:       MakeArgs(map[string]interface{}{"a": "one"}, true, "", "\t"),
			wantResult: "{\n\t\"a\": \"one\"\n}",
		},
		{
			name:       "ToJsonString Map (Pretty=true, 4 args, custom prefix/indent)",
			toolName:   "ToJsonString",
			args:       MakeArgs(map[string]interface{}{"a": "one"}, true, "> ", "| "),
			wantResult: "{\n> | \"a\": \"one\"\n> }",
		},
		{
			name:       "ToJsonString Map (Compact, pretty=false, custom ignored)",
			toolName:   "ToJsonString",
			args:       MakeArgs(map[string]interface{}{"a": "one"}, false, "> ", "\t"),
			wantResult: `{"a":"one"}`,
		},
		{
			name:       "ToJsonString Map (Pretty=true, 4 args, nil indent)",
			toolName:   "ToJsonString",
			args:       MakeArgs(map[string]interface{}{"a": "one"}, true, "", nil),
			wantResult: "{\n  \"a\": \"one\"\n}", // Nil indent -> default "  "
		},
		{
			name:       "ToJsonString Map (Pretty=true, 3 args, nil prefix)",
			toolName:   "ToJsonString",
			args:       MakeArgs(map[string]interface{}{"a": "one"}, true, nil, "\t"),
			wantResult: "{\n\t\"a\": \"one\"\n}", // Nil prefix -> default ""
		},
		{
			name:       "ToJsonString Map (Compact, 2 args, pretty=nil)",
			toolName:   "ToJsonString",
			args:       MakeArgs(map[string]interface{}{"a": "one"}, nil),
			wantResult: `{"a":"one"}`, // Nil bool -> false
		},
		{
			name:      "ToJsonString Wrong Type (pretty_print)",
			toolName:  "ToJsonString",
			args:      MakeArgs(map[string]interface{}{}, "not-a-bool"),
			wantErrIs: lang.ErrArgumentMismatch,
		},
		{
			name:      "ToJsonString Wrong Type (prefix)",
			toolName:  "ToJsonString",
			args:      MakeArgs(map[string]interface{}{}, true, 123),
			wantErrIs: lang.ErrArgumentMismatch,
		},
		{
			name:      "ToJsonString Wrong Type (indent)",
			toolName:  "ToJsonString",
			args:      MakeArgs(map[string]interface{}{}, true, "", 123),
			wantErrIs: lang.ErrArgumentMismatch,
		},
	}
	for _, tt := range tests {
		// Use helper from tools_string_basic_test.go
		// newStringTestInterpreter already ensures ALL tools (including extra) are registered via init()
		interp := newStringTestInterpreter(t)

		// Run the specific test case using the interpreter provided by the helper
		testStringToolHelper(t, interp, tt)
	}
}
