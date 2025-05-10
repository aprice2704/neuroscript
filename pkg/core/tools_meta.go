// NeuroScript Version: 0.3.8
// File version: 0.1.3 // Further refine nil argument handling for filter in toolToolsHelp
// Filename: pkg/core/tools_meta.go
// nlines: 145 // Approximate
// risk_rating: MEDIUM

package core

import (
	"fmt"
	"sort"
	"strings"
)

// toolListTools provides a compact list of available tools.
// NeuroScript: call Meta.ListTools() -> string
func toolListTools(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "Meta.ListTools: expects no arguments", ErrArgumentMismatch)
	}

	registry := interpreter.ToolRegistry()
	if registry == nil {
		return nil, NewRuntimeError(ErrorCodeConfiguration, "Meta.ListTools: ToolRegistry is not available", ErrConfiguration)
	}

	toolSpecs := registry.ListTools()
	if len(toolSpecs) == 0 {
		return "No tools are currently registered.", nil
	}

	sort.Slice(toolSpecs, func(i, j int) bool {
		return toolSpecs[i].Name < toolSpecs[j].Name
	})

	var output strings.Builder
	for _, spec := range toolSpecs {
		paramsStr := formatParamsSimpleForSpec(spec.Args)
		output.WriteString(fmt.Sprintf("%s(%s) -> %s\n", spec.Name, paramsStr, spec.ReturnType))
	}

	return output.String(), nil
}

// toolToolsHelp provides detailed help for available tools in Markdown format.
// NeuroScript: call Meta.ToolsHelp(filter?:string) -> string
func toolToolsHelp(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	var filterValue string // Defaults to empty string, meaning "no filter"

	// The 'args' slice is prepared by Interpreter.ExecuteTool based on the ToolSpec.
	// For an optional argument like 'filter', if not provided in the NeuroScript call,
	// Interpreter.ExecuteTool will place a `nil` at its position in `args`.
	if len(args) > 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "Meta.ToolsHelp: expects at most 1 argument (filter)", ErrArgumentMismatch)
	}

	if len(args) == 1 {
		// An argument was provided for the filter (it's the only argument).
		// This could be an actual string, or it could be `nil` if the NeuroScript call
		// was `tool.Meta.ToolsHelp()` and the interpreter inserted `nil` for the optional arg.
		if args[0] == nil {
			// Filter argument was explicitly nil (or an omitted optional arg became nil).
			// Treat as no filter; filterValue remains "".
		} else {
			// Argument is not nil, so it must be a string.
			strVal, ok := args[0].(string)
			if !ok {
				// This path is taken if args[0] is not nil AND not a string.
				return nil, NewRuntimeError(ErrorCodeType,
					fmt.Sprintf("Meta.ToolsHelp: 'filter' argument must be a string, got %T", args[0]), ErrInvalidArgument)
			}
			filterValue = strVal
		}
	}
	// If len(args) == 0 (should not happen if ToolSpec has one optional arg,
	// as ExecuteTool would likely create a slice of length 1 with nil),
	// filterValue remains "", which is correct for "no filter".

	displayFilter := filterValue                     // Store original casing for messages
	normalizedFilter := strings.ToLower(filterValue) // Use for case-insensitive matching

	registry := interpreter.ToolRegistry()
	if registry == nil {
		return nil, NewRuntimeError(ErrorCodeConfiguration, "Meta.ToolsHelp: ToolRegistry is not available", ErrConfiguration)
	}

	allToolSpecs := registry.ListTools()
	var filteredSpecs []ToolSpec
	for _, spec := range allToolSpecs {
		if normalizedFilter == "" || strings.Contains(strings.ToLower(spec.Name), normalizedFilter) {
			filteredSpecs = append(filteredSpecs, spec)
		}
	}

	if len(filteredSpecs) == 0 {
		if displayFilter != "" {
			return fmt.Sprintf("No tools found matching filter: `%s`", displayFilter), nil
		}
		return "No tools are currently registered.", nil
	}

	sort.Slice(filteredSpecs, func(i, j int) bool {
		return filteredSpecs[i].Name < filteredSpecs[j].Name
	})

	var mdBuilder strings.Builder
	mdBuilder.WriteString("# NeuroScript Tools Help\n\n")
	if displayFilter != "" {
		mdBuilder.WriteString(fmt.Sprintf("Showing tools matching filter: `%s`\n\n", displayFilter))
	}

	for _, spec := range filteredSpecs {
		mdBuilder.WriteString(fmt.Sprintf("## `tool.%s`\n", spec.Name))
		mdBuilder.WriteString(fmt.Sprintf("**Description:** %s\n\n", spec.Description))

		mdBuilder.WriteString("**Parameters:**\n")
		if len(spec.Args) > 0 {
			mdBuilder.WriteString(formatParamsMarkdownForSpec(spec.Args))
		} else {
			mdBuilder.WriteString("_None_\n")
		}
		mdBuilder.WriteString("\n")

		mdBuilder.WriteString(fmt.Sprintf("**Returns:** (`%s`)\n", spec.ReturnType))
		mdBuilder.WriteString("---\n\n")
	}

	return mdBuilder.String(), nil
}

// formatParamsSimpleForSpec formats ArgSpec for toolListTools
func formatParamsSimpleForSpec(params []ArgSpec) string {
	var parts []string
	for _, p := range params {
		part := fmt.Sprintf("%s:%s", p.Name, p.Type)
		if !p.Required {
			part += "?"
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, ", ")
}

// formatParamsMarkdownForSpec formats ArgSpec for toolToolsHelp
func formatParamsMarkdownForSpec(params []ArgSpec) string {
	if len(params) == 0 {
		return "_None_"
	}
	var mdBuilder strings.Builder
	for _, p := range params {
		requiredStr := ""
		if !p.Required {
			requiredStr = "(optional) "
		}
		mdBuilder.WriteString(fmt.Sprintf("* `%s` (`%s`): %s%s\n", p.Name, p.Type, requiredStr, p.Description))
	}
	return mdBuilder.String()
}
