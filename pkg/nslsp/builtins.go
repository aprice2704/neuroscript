// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines signatures and help text for all built-in NeuroScript functions.
// filename: pkg/nslsp/builtins.go
// nlines: 70
// risk_rating: LOW

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
