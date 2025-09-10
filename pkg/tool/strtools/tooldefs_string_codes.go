// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines codec and compression tools for the strtools package.
// filename: pkg/tool/strtools/tooldefs_string_codecs.go
// nlines: 108
// risk_rating: LOW

package strtools

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// stringCodecToolsToRegister contains ToolImplementation definitions for String codec and compression tools.
var stringCodecToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "ToBase64",
			Group:       group,
			Description: "Encodes a string using standard Base64 encoding.",
			Category:    "String Codecs",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to encode."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the Base64 encoded string.",
			Example:         `tool.ToBase64("hello world") // Returns "aGVsbG8gd29ybGQ="`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` if `input_string` is not a string.",
		},
		Func: toolStringToBase64,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "FromBase64",
			Group:       group,
			Description: "Decodes a Base64-encoded string.",
			Category:    "String Codecs",
			Args: []tool.ArgSpec{
				{Name: "encoded_string", Type: tool.ArgTypeString, Required: true, Description: "The Base64 string to decode."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the decoded string.",
			Example:         `tool.FromBase64("aGVsbG8gd29ybGQ=") // Returns "hello world"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` if `encoded_string` is not a string or is invalid Base64.",
		},
		Func: toolStringFromBase64,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "ToHex",
			Group:       group,
			Description: "Encodes a string into a hexadecimal representation.",
			Category:    "String Codecs",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to encode."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the hex-encoded string.",
			Example:         `tool.ToHex("hello") // Returns "68656c6c6f"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` if `input_string` is not a string.",
		},
		Func: toolStringToHex,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "FromHex",
			Group:       group,
			Description: "Decodes a string from its hexadecimal representation.",
			Category:    "String Codecs",
			Args: []tool.ArgSpec{
				{Name: "encoded_string", Type: tool.ArgTypeString, Required: true, Description: "The hex string to decode."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the decoded string.",
			Example:         `tool.FromHex("68656c6c6f") // Returns "hello"`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` if `encoded_string` is not a string or is invalid hex.",
		},
		Func: toolStringFromHex,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Compress",
			Group:       group,
			Description: "Compresses a string using Gzip and returns the result as a Base64-encoded string.",
			Category:    "String Compression",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to compress."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the Gzip compressed and Base64 encoded string.",
			Example:         `tool.Compress("some repeating text...")`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` if `input_string` is not a string.",
		},
		Func: toolStringCompress,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Decompress",
			Group:       group,
			Description: "Decodes a Base64 string and then decompresses the Gzip data to the original string.",
			Category:    "String Compression",
			Args: []tool.ArgSpec{
				{Name: "base64_encoded_string", Type: tool.ArgTypeString, Required: true, Description: "The Base64 encoded Gzip data to decompress."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the decompressed original string.",
			Example:         `tool.Decompress("H4sIAAAAAAAA/...")`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` if the input is not a valid Base64/Gzip string.",
		},
		Func: toolStringDecompress,
	},
}
