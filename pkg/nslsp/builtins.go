// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 3
// :: description: Defines signatures and help text for all built-in NeuroScript functions and predefined variables.
// :: latestChange: Added MinArgs/MaxArgs for arity checking and PredefinedVariables map.
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
	MinArgs     int
	MaxArgs     int
}

// BuiltInFunctions is the canonical map of all built-in functions and their help info.
var BuiltInFunctions = map[string]BuiltInFunctionInfo{
	"sin": {
		Signature:   "sin(value: number) -> number",
		Description: "Returns the sine of the 'value' (in radians).",
		MinArgs:     1, MaxArgs: 1,
	},
	"cos": {
		Signature:   "cos(value: number) -> number",
		Description: "Returns the cosine of the 'value' (in radians).",
		MinArgs:     1, MaxArgs: 1,
	},
	"tan": {
		Signature:   "tan(value: number) -> number",
		Description: "Returns the tangent of the 'value' (in radians).",
		MinArgs:     1, MaxArgs: 1,
	},
	"asin": {
		Signature:   "asin(value: number) -> number",
		Description: "Returns the arcsine of the 'value'.",
		MinArgs:     1, MaxArgs: 1,
	},
	"acos": {
		Signature:   "acos(value: number) -> number",
		Description: "Returns the arccosine of the 'value'.",
		MinArgs:     1, MaxArgs: 1,
	},
	"atan": {
		Signature:   "atan(value: number) -> number",
		Description: "Returns the arctangent of the 'value'.",
		MinArgs:     1, MaxArgs: 1,
	},
	"ln": {
		Signature:   "ln(value: number) -> number",
		Description: "Returns the natural logarithm of the 'value'.",
		MinArgs:     1, MaxArgs: 1,
	},
	"log": {
		Signature:   "log(value: number) -> number",
		Description: "Returns the base-10 logarithm of the 'value'.",
		MinArgs:     1, MaxArgs: 1,
	},
	"len": {
		Signature:   "len(value: any) -> number",
		Description: "Returns the length of a string, list, or map.",
		MinArgs:     1, MaxArgs: 1,
	},
	"typeof": {
		Signature:   "typeof(value: any) -> string",
		Description: "Returns the type of the 'value' as a string (e.g., \"number\", \"string\", \"list\", \"map\", \"nil\").",
		MinArgs:     1, MaxArgs: 1,
	},
	"eval": {
		Signature:   "eval(script_string: string) -> any",
		Description: "Executes a dynamic string of NeuroScript code. This is a powerful and high-risk function.",
		MinArgs:     1, MaxArgs: 1,
	},
	"char": {
		Signature:   "char(codepoint: number) -> string",
		Description: "Returns a string containing the character corresponding to the numeric Unicode codepoint.",
		MinArgs:     1, MaxArgs: 1,
	},
	"ord": {
		Signature:   "ord(char: string) -> number",
		Description: "Returns the numeric Unicode codepoint of the first character in the string.",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_string": {
		Signature:   "is_string(value: any) -> bool",
		Description: "Returns true if the value is a string.",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_number": {
		Signature:   "is_number(value: any) -> bool",
		Description: "Returns true if the value is a number.",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_bool": {
		Signature:   "is_bool(value: any) -> bool",
		Description: "Returns true if the value is a boolean.",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_list": {
		Signature:   "is_list(value: any) -> bool",
		Description: "Returns true if the value is a list.",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_map": {
		Signature:   "is_map(value: any) -> bool",
		Description: "Returns true if the value is a map.",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_nil": {
		Signature:   "is_nil(value: any) -> bool",
		Description: "Returns true if the value is nil.",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_int": {
		Signature:   "is_int(value: any) -> bool",
		Description: "Returns true if the value is an integer (a number with no fractional part).",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_float": {
		Signature:   "is_float(value: any) -> bool",
		Description: "Returns true if the value is a float (a number with a fractional part).",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_error": {
		Signature:   "is_error(value: any) -> bool",
		Description: "Returns true if the value is an error object.",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_function": {
		Signature:   "is_function(value: any) -> bool",
		Description: "Returns true if the value is a function.",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_tool": {
		Signature:   "is_tool(value: any) -> bool",
		Description: "Returns true if the value is a tool.",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_event": {
		Signature:   "is_event(value: any) -> bool",
		Description: "Returns true if the value is an event.",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_timedate": {
		Signature:   "is_timedate(value: any) -> bool",
		Description: "Returns true if the value is a timedate.",
		MinArgs:     1, MaxArgs: 1,
	},
	"is_fuzzy": {
		Signature:   "is_fuzzy(value: any) -> bool",
		Description: "Returns true if the value is a fuzzy value.",
		MinArgs:     1, MaxArgs: 1,
	},
}

// PredefinedVariableInfo holds the definition for a predefined system variable.
type PredefinedVariableInfo struct {
	Type        string
	Description string
}

// PredefinedVariables is the canonical map of all predefined variables.
var PredefinedVariables = map[string]PredefinedVariableInfo{
	"self": {
		Type:        "handle",
		Description: "The handle to the interpreter's default internal text buffer. Its primary purpose is to serve as a channel for providing contextual information to the `ask` statement via the `whisper` command.",
	},
	"system_error_message": {
		Type:        "string",
		Description: "Contains a string describing the failure when inside an `on error` block.",
	},
	"stdout": {
		Type:        "stream",
		Description: "Standard output stream.",
	},
	"stderr": {
		Type:        "stream",
		Description: "Standard error stream.",
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

	bt3 := "``" + "`"
	var hoverContent strings.Builder
	hoverContent.WriteString(fmt.Sprintf("%sneuroscript\n(built-in) %s\n%s\n", bt3, info.Signature, bt3))
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

// getPredefinedVariableCompletions creates a completion list for predefined variables.
func (s *Server) getPredefinedVariableCompletions() *lsp.CompletionList {
	items := make([]lsp.CompletionItem, 0, len(PredefinedVariables))
	for name, info := range PredefinedVariables {
		items = append(items, lsp.CompletionItem{
			Label:         name,
			Kind:          lsp.CIKVariable,
			Detail:        info.Type,
			Documentation: info.Description,
		})
	}
	return &lsp.CompletionList{IsIncomplete: false, Items: items}
}

// getPredefinedVariableHover creates hover info for a predefined variable.
func (s *Server) getPredefinedVariableHover(name string) *lsp.Hover {
	info, found := PredefinedVariables[name]
	if !found {
		return nil
	}

	bt3 := "``" + "`"
	var hoverContent strings.Builder
	hoverContent.WriteString(fmt.Sprintf("%sneuroscript\n(predefined) %s: %s\n%s\n", bt3, name, info.Type, bt3))
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
