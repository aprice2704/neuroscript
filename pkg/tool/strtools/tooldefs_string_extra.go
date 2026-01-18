// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 5
// :: description: Defines extra string/codec tools. Updated ToJsonString description to reflect it accepts any type.
// :: latestChange: Updated ToJsonString docs for Any type support.
// :: filename: pkg/tool/strtools/tooldefs_string_extra.go
// :: serialization: go

package strtools

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// stringExtraToolsToRegister contains ToolImplementation definitions for extra String codec tools.
var stringExtraToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "BytesFromBase64",
			Group:       group,
			Description: "Decodes a Base64 string (representing bytes) into a string, assuming UTF-8 encoding.",
			Category:    "String Codecs",
			Args: []tool.ArgSpec{
				{Name: "base64_string", Type: tool.ArgTypeString, Required: true, Description: "The Base64 encoded string representing byte data."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the decoded UTF-8 string.",
			Example:         `str.BytesFromBase64("SGVsbG8gV29ybGQ=") // Returns "Hello World"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` if input is not a string, invalid Base64, or not valid UTF-8.",
		},
		Func: toolBytesFromBase64,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "BytesToBase64",
			Group:       group,
			Description: "Converts a string into a Base64 encoded string representing its UTF-8 bytes.",
			Category:    "String Codecs",
			Args: []tool.ArgSpec{
				{Name: "string_data", Type: tool.ArgTypeString, Required: true, Description: "The string to convert."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the Base64 encoded string representing the UTF-8 bytes.",
			Example:         `str.BytesToBase64("Hello World") // Returns "SGVsbG8gV29ybGQ="`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided or input is not a string.",
		},
		Func: toolBytesToBase64,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "ParseFromJsonBase64",
			Group:       group,
			Description: "Parses JSON data from a Base64 encoded string (representing bytes) into a map or list.",
			Category:    "String Codecs",
			Args: []tool.ArgSpec{
				{Name: "base64_string", Type: tool.ArgTypeString, Required: true, Description: "The Base64 encoded string representing JSON byte data."},
			},
			ReturnType:      tool.ArgTypeAny, // Can return map or list
			ReturnHelp:      "Returns the parsed map or list.",
			Example:         `str.ParseFromJsonBase64("eyJrZXkiOiAidmFsdWUifQ==") // Returns {"key": "value"}`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` if input is not a string, invalid Base64, or invalid JSON.",
		},
		Func: toolParseFromJsonBase64,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "ParseJsonString",
			Group:       group,
			Description: "Parses JSON data from a plain string into a map or list.",
			Category:    "String Codecs",
			Args: []tool.ArgSpec{
				{Name: "json_string", Type: tool.ArgTypeString, Required: true, Description: "The plain string representing JSON data."},
			},
			ReturnType:      tool.ArgTypeAny, // Can return map or list
			ReturnHelp:      "Returns the parsed map or list.",
			Example:         `str.ParseJsonString("{\"key\": \"value\"}") // Returns {"key": "value"}`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` if input is not a string or is invalid JSON.",
		},
		Func: toolParseJsonString,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "ToJsonString",
			Group:       group,
			Description: "Converts any value (map, list, string, number, etc.) into a JSON formatted string.",
			Category:    "String Codecs",
			Args: []tool.ArgSpec{
				{Name: "value", Type: tool.ArgTypeAny, Required: true, Description: "The value to stringify."},
				{Name: "pretty_print", Type: tool.ArgTypeBool, Required: false, Description: "If true, formats the JSON with indentation. Default: false."},
				{Name: "prefix", Type: tool.ArgTypeString, Required: false, Description: "The line prefix for pretty-printing. Default: \"\". Only used if pretty_print is true."},
				{Name: "indent", Type: tool.ArgTypeString, Required: false, Description: "The indentation string for pretty-printing. Default: \"  \". Only used if pretty_print is true."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the JSON formatted string.",
			Example:         `str.ToJsonString({"key": "value"}) // Returns "{\"key\":\"value\"}"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` if the value cannot be represented as JSON or if optional arguments have wrong types.",
		},
		Func: toolToJsonString,
	},
}
