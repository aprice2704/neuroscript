// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 2
// :: description: Defines signatures and help text for all built-in NeuroScript functions.
// :: latestChange: Added new char, ord, and type-checking (is_*) built-in functions.
// :: filename: pkg/nslsp/builtins.go
// :: serialization: go

package nslsp

import (
	"fmt"
	"strings"

	lsp "github.com/sourcegraph/go-lsp"
)

// BuiltInFunctionInfo holds the definition for a built-in function.
type BuiltInFunctionInfo struct {
	Signature   string
	Description string
}

// BuiltInFunctions is the canonical map of all built-in functions and their help info.
var BuiltInFunctions = map[string]BuiltInFunctionInfo{
	"sin": {
		Signature:   "sin(value: number) -> number",
		Description: "Returns the sine of the 'value' (in radians).",
	},
	"cos": {
		Signature:   "cos(value: number) -> number",
		Description: "Returns the cosine of the 'value' (in radians).",
	},
	"tan": {
		Signature:   "tan(value: number) -> number",
		Description: "Returns the tangent of the 'value' (in radians).",
	},
	"asin": {
		Signature:   "asin(value: number) -> number",
		Description: "Returns the arcsine of the 'value'.",
	},
	"acos": {
		Signature:   "acos(value: number) -> number",
		Description: "Returns the arccosine of the 'value'.",
	},
	"atan": {
		Signature:   "atan(value: number) -> number",
		Description: "Returns the arctangent of the 'value'.",
	},
	"ln": {
		Signature:   "ln(value: number) -> number",
		Description: "Returns the natural logarithm of the 'value'.",
	},
	"log": {
		Signature:   "log(value: number) -> number",
		Description: "Returns the base-10 logarithm of the 'value'.",
	},
	"len": {
		Signature:   "len(value: any) -> number",
		Description: "Returns the length of a string, list, or map.",
	},
	"typeof": {
		Signature:   "typeof(value: any) -> string",
		Description: "Returns the type of the 'value' as a string (e.g., \"number\", \"string\", \"list\", \"map\", \"nil\").",
	},
	"eval": {
		Signature:   "eval(script_string: string) -> any",
		Description: "Executes a dynamic string of NeuroScript code. This is a powerful and high-risk function.",
	},
	"char": {
		Signature:   "char(codepoint: number) -> string",
		Description: "Returns a string containing the character corresponding to the numeric Unicode codepoint.",
	},
	"ord": {
		Signature:   "ord(char: string) -> number",
		Description: "Returns the numeric Unicode codepoint of the first character in the string.",
	},
	"is_string": {
		Signature:   "is_string(value: any) -> bool",
		Description: "Returns true if the value is a string.",
	},
	"is_number": {
		Signature:   "is_number(value: any) -> bool",
		Description: "Returns true if the value is a number.",
	},
	"is_bool": {
		Signature:   "is_bool(value: any) -> bool",
		Description: "Returns true if the value is a boolean.",
	},
	"is_list": {
		Signature:   "is_list(value: any) -> bool",
		Description: "Returns true if the value is a list.",
	},
	"is_map": {
		Signature:   "is_map(value: any) -> bool",
		Description: "Returns true if the value is a map.",
	},
	"is_nil": {
		Signature:   "is_nil(value: any) -> bool",
		Description: "Returns true if the value is nil.",
	},
	"is_int": {
		Signature:   "is_int(value: any) -> bool",
		Description: "Returns true if the value is an integer (a number with no fractional part).",
	},
	"is_float": {
		Signature:   "is_float(value: any) -> bool",
		Description: "Returns true if the value is a float (a number with a fractional part).",
	},
	"is_error": {
		Signature:   "is_error(value: any) -> bool",
		Description: "Returns true if the value is an error object.",
	},
	"is_function": {
		Signature:   "is_function(value: any) -> bool",
		Description: "Returns true if the value is a function.",
	},
	"is_tool": {
		Signature:   "is_tool(value: any) -> bool",
		Description: "Returns true if the value is a tool.",
	},
	"is_event": {
		Signature:   "is_event(value: any) -> bool",
		Description: "Returns true if the value is an event.",
	},
	"is_timedate": {
		Signature:   "is_timedate(value: any) -> bool",
		Description: "Returns true if the value is a timedate.",
	},
	"is_fuzzy": {
		Signature:   "is_fuzzy(value: any) -> bool",
		Description: "Returns true if the value is a fuzzy value.",
	},
}

// getBuiltInCompletions creates a completion list for all built-in functions.
func (s *Server) getBuiltInCompletions() *lsp.CompletionList {
	items := make([]lsp.CompletionItem, 0, len(BuiltInFunctions))
	for name, info := range BuiltInFunctions {
		items = append(items, lsp.CompletionItem{
			Label:         name,
			Kind:          lsp.CIKFunction,
			Detail:        info.Signature,
			Documentation: info.Description,
		})
	}
	return &lsp.CompletionList{IsIncomplete: false, Items: items}
}

// getBuiltInHover creates hover info for a built-in function.
func (s *Server) getBuiltInHover(name string) *lsp.Hover {
	info, found := BuiltInFunctions[name]
	if !found {
		return nil
	}

	var hoverContent strings.Builder
	hoverContent.WriteString(fmt.Sprintf("```neuroscript\n(built-in) %s\n```\n", info.Signature))
	hoverContent.WriteString("---\n")
	hoverContent.WriteString(info.Description)

	return &lsp.Hover{
		Contents: []lsp.MarkedString{
			{
				Language: "markdown",
				Value:    hoverContent.String(),
			},
		},
	}
}
