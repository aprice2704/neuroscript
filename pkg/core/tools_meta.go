// NeuroScript Version: 0.3.8
// File version: 0.1.4 // Added toolGetToolSpecificationsJSON implementation
// Filename: pkg/core/tools_meta.go
// nlines: 175 // Approximate
// risk_rating: MEDIUM

package core

import (
	"encoding/json" // Added for JSON marshalling
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

	if len(args) > 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "Meta.ToolsHelp: expects at most 1 argument (filter)", ErrArgumentMismatch)
	}

	if len(args) == 1 {
		if args[0] == nil {
			// Filter argument was explicitly nil (or an omitted optional arg became nil).
		} else {
			strVal, ok := args[0].(string)
			if !ok {
				return nil, NewRuntimeError(ErrorCodeType,
					fmt.Sprintf("Meta.ToolsHelp: 'filter' argument must be a string, got %T", args[0]), ErrInvalidArgument)
			}
			filterValue = strVal
		}
	}

	displayFilter := filterValue
	normalizedFilter := strings.ToLower(filterValue)

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
		// --- START: Include new fields in ToolsHelp ---
		if spec.Category != "" {
			mdBuilder.WriteString(fmt.Sprintf("**Category:** %s\n\n", spec.Category))
		}

		mdBuilder.WriteString("**Parameters:**\n")
		if len(spec.Args) > 0 {
			mdBuilder.WriteString(formatParamsMarkdownForSpec(spec.Args)) // formatParamsMarkdownForSpec already handles DefaultValue if present
		} else {
			mdBuilder.WriteString("_None_\n")
		}
		mdBuilder.WriteString("\n")

		mdBuilder.WriteString(fmt.Sprintf("**Returns:** (`%s`) %s\n", spec.ReturnType, spec.ReturnHelp))
		if spec.Variadic {
			mdBuilder.WriteString("**Variadic:** Yes\n")
		}
		if spec.Example != "" {
			mdBuilder.WriteString(fmt.Sprintf("\n**Example:**\n```neuroscript\n%s\n```\n", spec.Example))
		}
		if spec.ErrorConditions != "" {
			mdBuilder.WriteString(fmt.Sprintf("\n**Error Conditions:** %s\n", spec.ErrorConditions))
		}
		// --- END: Include new fields in ToolsHelp ---
		mdBuilder.WriteString("---\n\n")
	}

	return mdBuilder.String(), nil
}

// toolGetToolSpecificationsJSON provides a JSON string of all available tool specifications.
// NeuroScript: call Meta.GetToolSpecificationsJSON() -> string
func toolGetToolSpecificationsJSON(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "Meta.GetToolSpecificationsJSON: expects no arguments", ErrArgumentMismatch)
	}

	registry := interpreter.ToolRegistry()
	if registry == nil {
		return nil, NewRuntimeError(ErrorCodeConfiguration, "Meta.GetToolSpecificationsJSON: ToolRegistry is not available", ErrConfiguration)
	}

	toolSpecs := registry.ListTools() // This returns []ToolSpec

	// Sort toolSpecs by name for consistent output order in the JSON array.
	sort.Slice(toolSpecs, func(i, j int) bool {
		return toolSpecs[i].Name < toolSpecs[j].Name
	})

	jsonData, err := json.MarshalIndent(toolSpecs, "", "  ") // Using MarshalIndent for readability
	if err != nil {
		if interpreter.logger != nil {
			interpreter.logger.Errorf("Meta.GetToolSpecificationsJSON: Failed to marshal tool specifications to JSON: %v", err)
		}
		return "", NewRuntimeError(ErrorCodeInternal, "failed to marshal tool specifications to JSON", err)
	}

	return string(jsonData), nil
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
			requiredStr = "(optional"
			if p.DefaultValue != nil {
				// Check if default value is string and needs quoting
				if _, ok := p.DefaultValue.(string); ok {
					requiredStr += fmt.Sprintf(", default: \"%v\"", p.DefaultValue)
				} else {
					requiredStr += fmt.Sprintf(", default: %v", p.DefaultValue)
				}
			}
			requiredStr += ") "
		}
		mdBuilder.WriteString(fmt.Sprintf("* `%s` (`%s`): %s%s\n", p.Name, p.Type, requiredStr, p.Description))
	}
	return mdBuilder.String()
}
