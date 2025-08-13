// NeuroScript Version: 0.3.8
// File version: 0.2.0
// Purpose: Corrected all compiler errors by accessing the nested 'Spec' field for tool properties.
// nlines: 175 // Approximate
// risk_rating: MEDIUM

// filename: pkg/tool/meta/tools_meta.go
package meta

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// toolListTools provides a compact list of available tools.
// NeuroScript: call Meta.ListTools() -> string
func toolListTools(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Meta.ListTools: expects no arguments", lang.ErrArgumentMismatch)
	}

	registry := interpreter.ToolRegistry()
	if registry == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "Meta.ListTools: ToolRegistry is not available from runtime", lang.ErrConfiguration)
	}

	toolImpls := registry.ListTools()
	if len(toolImpls) == 0 {
		return "No tools are currently registered.", nil
	}

	var toolStrings []string
	for _, impl := range toolImpls {
		spec := impl.Spec
		fullName := types.MakeFullName(string(spec.Group), string(spec.Name))
		paramsStr := formatParamsSimpleForSpec(spec.Args)
		toolStrings = append(toolStrings, fmt.Sprintf("%s(%s) -> %s", fullName, paramsStr, spec.ReturnType))
	}

	sort.Strings(toolStrings)

	return strings.Join(toolStrings, "\n") + "\n", nil
}

// toolToolsHelp provides detailed help for available tools in Markdown format.
// NeuroScript: call Meta.ToolsHelp(filter?:string) -> string
func toolToolsHelp(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	var filterValue string

	if len(args) > 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Meta.ToolsHelp: expects at most 1 argument (filter)", lang.ErrArgumentMismatch)
	}

	if len(args) == 1 && args[0] != nil {
		strVal, ok := args[0].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType,
				fmt.Sprintf("ToolsHelp: 'filter' argument must be a string, got %T", args[0]), lang.ErrInvalidArgument)
		}
		filterValue = strVal
	}

	displayFilter := filterValue
	normalizedFilter := strings.ToLower(filterValue)

	registry := interpreter.ToolRegistry()
	if registry == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "Meta.ToolsHelp: ToolRegistry is not available from runtime", lang.ErrConfiguration)
	}

	allToolImpls := registry.ListTools()
	var filteredSpecs []tool.ToolSpec
	for _, impl := range allToolImpls {
		spec := impl.Spec
		fullname := types.MakeFullName(string(spec.Group), string(spec.Name))
		if normalizedFilter == "" || strings.Contains(strings.ToLower(string(fullname)), normalizedFilter) {
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
		fullNameI := types.MakeFullName(string(filteredSpecs[i].Group), string(filteredSpecs[i].Name))
		fullNameJ := types.MakeFullName(string(filteredSpecs[j].Group), string(filteredSpecs[j].Name))
		return fullNameI < fullNameJ
	})

	var mdBuilder strings.Builder
	mdBuilder.WriteString("# NeuroScript Tools Help\n\n")
	if displayFilter != "" {
		mdBuilder.WriteString(fmt.Sprintf("Showing tools matching filter: `%s`\n\n", displayFilter))
	}

	for _, spec := range filteredSpecs {
		fullName := types.MakeFullName(string(spec.Group), string(spec.Name))
		mdBuilder.WriteString(fmt.Sprintf("## `%s`\n", fullName))
		mdBuilder.WriteString(fmt.Sprintf("**Description:** %s\n\n", spec.Description))
		if spec.Category != "" {
			mdBuilder.WriteString(fmt.Sprintf("**Category:** %s\n\n", spec.Category))
		}
		mdBuilder.WriteString("**Parameters:**\n")
		mdBuilder.WriteString(formatParamsMarkdownForSpec(spec.Args))
		mdBuilder.WriteString("\n")
		mdBuilder.WriteString(fmt.Sprintf("**Returns:** (`%s`) %s\n", spec.ReturnType, spec.ReturnHelp))
		if spec.Example != "" {
			mdBuilder.WriteString(fmt.Sprintf("\n**Example:**\n```neuroscript\n%s\n```\n", spec.Example))
		}
		mdBuilder.WriteString("---\n\n")
	}

	return mdBuilder.String(), nil
}

// toolGetToolSpecificationsJSON provides a JSON string of all available tool specifications.
func toolGetToolSpecificationsJSON(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Meta.GetToolSpecificationsJSON: expects no arguments", lang.ErrArgumentMismatch)
	}

	registry := interpreter.ToolRegistry()
	if registry == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "Meta.GetToolSpecificationsJSON: ToolRegistry is not available", lang.ErrConfiguration)
	}

	allToolImpls := registry.ListTools()
	toolSpecs := make([]tool.ToolSpec, len(allToolImpls))
	for i, impl := range allToolImpls {
		toolSpecs[i] = impl.Spec
	}

	sort.Slice(toolSpecs, func(i, j int) bool {
		fullNameI := types.MakeFullName(string(toolSpecs[i].Group), string(toolSpecs[i].Name))
		fullNameJ := types.MakeFullName(string(toolSpecs[j].Group), string(toolSpecs[j].Name))
		return fullNameI < fullNameJ
	})

	jsonData, err := json.MarshalIndent(toolSpecs, "", "  ")
	if err != nil {
		interpreter.GetLogger().Errorf("Meta.GetToolSpecificationsJSON: Failed to marshal tool specs: %v", err)
		return "", lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal tool specifications to JSON", err)
	}

	return string(jsonData), nil
}

// formatParamsSimpleForSpec formats ArgSpec for toolListTools
func formatParamsSimpleForSpec(params []tool.ArgSpec) string {
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
func formatParamsMarkdownForSpec(params []tool.ArgSpec) string {
	if len(params) == 0 {
		return "_None_\n"
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
