// NeuroScript Version: 0.3.8
// File version: 0.3.2
// Purpose: Corrected type casting for case-insensitive sorting and fixed a bad import path.
// filename: pkg/tool/meta/tools_meta.go
// nlines: 206 // Approximate
// risk_rating: MEDIUM

// filename: pkg/tool/meta/tools_meta.go
package meta

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
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
		capsStr := formatCapsSimple(impl.RequiredCaps)
		toolStrings = append(toolStrings, fmt.Sprintf("%s(%s) -> %s%s", fullName, paramsStr, spec.ReturnType, capsStr))
	}

	// Use a case-insensitive sort for predictable alphabetical order.
	sort.Slice(toolStrings, func(i, j int) bool {
		return strings.ToLower(toolStrings[i]) < strings.ToLower(toolStrings[j])
	})

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
	var filteredImpls []tool.ToolImplementation
	for _, impl := range allToolImpls {
		spec := impl.Spec
		fullname := types.MakeFullName(string(spec.Group), string(spec.Name))
		if normalizedFilter == "" || strings.Contains(strings.ToLower(string(fullname)), normalizedFilter) {
			filteredImpls = append(filteredImpls, impl)
		}
	}

	if len(filteredImpls) == 0 {
		if displayFilter != "" {
			return fmt.Sprintf("No tools found matching filter: `%s`", displayFilter), nil
		}
		return "No tools are currently registered.", nil
	}

	// Use a case-insensitive sort for predictable alphabetical order.
	sort.Slice(filteredImpls, func(i, j int) bool {
		fullNameI := types.MakeFullName(string(filteredImpls[i].Spec.Group), string(filteredImpls[i].Spec.Name))
		fullNameJ := types.MakeFullName(string(filteredImpls[j].Spec.Group), string(filteredImpls[j].Spec.Name))
		return strings.ToLower(string(fullNameI)) < strings.ToLower(string(fullNameJ))
	})

	var mdBuilder strings.Builder
	mdBuilder.WriteString("# NeuroScript Tools Help\n\n")
	if displayFilter != "" {
		mdBuilder.WriteString(fmt.Sprintf("Showing tools matching filter: `%s`\n\n", displayFilter))
	}

	for _, impl := range filteredImpls {
		spec := impl.Spec
		fullName := types.MakeFullName(string(spec.Group), string(spec.Name))
		mdBuilder.WriteString(fmt.Sprintf("## `%s`\n", fullName))
		mdBuilder.WriteString(fmt.Sprintf("**Description:** %s\n\n", spec.Description))
		if spec.Category != "" {
			mdBuilder.WriteString(fmt.Sprintf("**Category:** %s\n\n", spec.Category))
		}
		if len(impl.RequiredCaps) > 0 {
			mdBuilder.WriteString("**Required Capabilities:**\n")
			mdBuilder.WriteString(formatCapsMarkdown(impl.RequiredCaps))
			mdBuilder.WriteString("\n")
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
	// Use a case-insensitive sort for predictable alphabetical order.
	sort.Slice(allToolImpls, func(i, j int) bool {
		fullNameI := types.MakeFullName(string(allToolImpls[i].Spec.Group), string(allToolImpls[i].Spec.Name))
		fullNameJ := types.MakeFullName(string(allToolImpls[j].Spec.Group), string(allToolImpls[j].Spec.Name))
		return strings.ToLower(string(fullNameI)) < strings.ToLower(string(fullNameJ))
	})

	// Create a serializable version without the Func pointer
	type serializableTool struct {
		tool.ToolSpec
		RequiresTrust bool                    `json:"requires_trust"`
		RequiredCaps  []capability.Capability `json:"required_caps"`
		Effects       []string                `json:"effects"`
	}
	serializableTools := make([]serializableTool, len(allToolImpls))
	for i, impl := range allToolImpls {
		serializableTools[i] = serializableTool{
			ToolSpec:      impl.Spec,
			RequiresTrust: impl.RequiresTrust,
			RequiredCaps:  impl.RequiredCaps,
			Effects:       impl.Effects,
		}
	}

	jsonData, err := json.MarshalIndent(serializableTools, "", "  ")
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

func formatCapsSimple(caps []capability.Capability) string {
	if len(caps) == 0 {
		return ""
	}
	var parts []string
	for _, c := range caps {
		scopePart := ""
		if len(c.Scopes) > 0 {
			scopePart = ":" + strings.Join(c.Scopes, ",")
		}
		parts = append(parts, fmt.Sprintf("%s:%s%s", c.Resource, strings.Join(c.Verbs, ","), scopePart))
	}
	return " [caps: " + strings.Join(parts, "; ") + "]"
}

func formatCapsMarkdown(caps []capability.Capability) string {
	var mdBuilder strings.Builder
	for _, c := range caps {
		scopePart := ""
		if len(c.Scopes) > 0 {
			scopePart = ":" + strings.Join(c.Scopes, ",")
		}
		mdBuilder.WriteString(fmt.Sprintf("* `%s:%s%s`\n", c.Resource, strings.Join(c.Verbs, ","), scopePart))
	}
	return mdBuilder.String()
}
