// NeuroScript Major Version: 1
// File version: 11
// Purpose: Implements meta tools. Added ListToolNames and ToolsHelp functions.
// Latest change: Removed fmt.Fprintf debug output from ToolsHelp after tests passed.
// filename: pkg/tool/meta/tools_meta.go
// nlines: 194

package meta

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// RegisterTools registers all the tools in the meta package with the provided registrar.
func RegisterTools(registrar tool.ToolRegistrar) error {
	for _, t := range metaToolsToRegister {
		if _, err := registrar.RegisterTool(t); err != nil {
			return fmt.Errorf("failed to register meta tool '%s': %w", t.Spec.Name, err)
		}
	}
	return nil
}

// ListTools is the implementation for the 'tool.meta.listTools' tool.
// It returns a slice of maps to be compatible with lang.Wrap.
func ListTools(rt tool.Runtime, args []any) (any, error) {
	// fmt.Fprintf(os.Stderr, "DEBUG: Entered ListTools implementation\n")
	tools := rt.ToolRegistry().ListTools()
	specs := make([]tool.ToolSpec, len(tools))
	for i, t := range tools {
		specs[i] = t.Spec
	}
	// fmt.Fprintf(os.Stderr, "DEBUG: ListTools: Found %d tools in registry\n", len(specs))

	// Sort for deterministic output
	sort.Slice(specs, func(i, j int) bool {
		return strings.ToLower(string(specs[i].FullName)) < strings.ToLower(string(specs[j].FullName))
	})

	// Convert slice of structs to slice of maps via JSON
	var specMaps []any
	jsonData, err := json.Marshal(specs)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "DEBUG: ListTools: Failed to marshal specs: %v\n", err)
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal tool specs to JSON for wrapping", err)
	}
	if err := json.Unmarshal(jsonData, &specMaps); err != nil {
		// fmt.Fprintf(os.Stderr, "DEBUG: ListTools: Failed to unmarshal specs: %v\n", err)
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to unmarshal tool spec JSON for wrapping", err)
	}

	// fmt.Fprintf(os.Stderr, "DEBUG: ListTools: Successfully converted specs to %d maps\n", len(specMaps))
	return specMaps, nil
}

// GetToolSpecificationsJSON provides a JSON string of all available tool specifications.
func GetToolSpecificationsJSON(rt tool.Runtime, args []interface{}) (interface{}, error) {
	// fmt.Fprintf(os.Stderr, "DEBUG: Entered GetToolSpecificationsJSON implementation\n")
	if len(args) != 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "GetToolSpecificationsJSON: expects no arguments", lang.ErrArgumentMismatch)
	}

	registry := rt.ToolRegistry()
	if registry == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "GetToolSpecificationsJSON: ToolRegistry is not available", lang.ErrConfiguration)
	}

	specs, err := ListTools(rt, nil)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "DEBUG: GetToolSpecificationsJSON: Error calling ListTools: %v\n", err)
		return nil, err
	}

	jsonData, err := json.MarshalIndent(specs, "", "  ")
	if err != nil {
		// Use rt.GetLogger() if available, otherwise stderr
		if rt.GetLogger() != nil {
			rt.GetLogger().Errorf("GetToolSpecificationsJSON: Failed to marshal tool specs: %v", err)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: GetToolSpecificationsJSON: Failed to marshal tool specs: %v\n", err)
		}
		return "", lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal tool specifications to JSON", err)
	}

	// fmt.Fprintf(os.Stderr, "DEBUG: GetToolSpecificationsJSON: Returning JSON string of %d bytes\n", len(jsonData))
	return string(jsonData), nil
}

// ListToolNames provides a simple, newline-separated list of all available tool signatures.
func ListToolNames(rt tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "ListToolNames: expects no arguments", lang.ErrArgumentMismatch)
	}

	tools := rt.ToolRegistry().ListTools()
	specs := make([]tool.ToolSpec, len(tools))
	for i, t := range tools {
		specs[i] = t.Spec
	}

	// Sort for deterministic output
	sort.Slice(specs, func(i, j int) bool {
		return strings.ToLower(string(specs[i].FullName)) < strings.ToLower(string(specs[j].FullName))
	})

	var b strings.Builder
	for _, spec := range specs {
		b.WriteString(formatSignature(spec))
		b.WriteRune('\n')
	}

	return b.String(), nil
}

// ToolsHelp provides formatted Markdown help text for tools, optionally filtered by name.
func ToolsHelp(rt tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) > 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "ToolsHelp: expects 0 or 1 argument (filter)", lang.ErrArgumentMismatch)
	}

	filter := ""
	if len(args) == 1 {
		if args[0] != nil {
			var ok bool
			filter, ok = args[0].(string)
			if !ok {
				return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ToolsHelp: filter argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
			}
		}
	}

	tools := rt.ToolRegistry().ListTools()
	specs := make([]tool.ToolSpec, len(tools))
	for i, t := range tools {
		specs[i] = t.Spec
	}

	// Sort for deterministic output
	sort.Slice(specs, func(i, j int) bool {
		return strings.ToLower(string(specs[i].FullName)) < strings.ToLower(string(specs[j].FullName))
	})

	var b strings.Builder
	count := 0
	for _, spec := range specs {
		if filter == "" || strings.Contains(string(spec.FullName), filter) {
			if count > 0 {
				b.WriteString("\n---\n\n")
			}
			b.WriteString(formatToolHelp(spec))
			count++
		}
	}

	if count == 0 {
		return fmt.Sprintf("No tools found matching filter: %q", filter), nil
	}

	return b.String(), nil
}

// formatSignature creates a human-readable signature string for a tool.
func formatSignature(spec tool.ToolSpec) string {
	var argParts []string
	for _, arg := range spec.Args {
		argParts = append(argParts, fmt.Sprintf("%s:%s", arg.Name, arg.Type))
	}
	argString := strings.Join(argParts, ", ")
	return fmt.Sprintf("%s(%s) -> %s", spec.FullName, argString, spec.ReturnType)
}

// formatToolHelp creates a detailed Markdown-formatted help block for a single tool.
func formatToolHelp(spec tool.ToolSpec) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("## %s\n\n", spec.FullName))
	b.WriteString(fmt.Sprintf("> %s\n\n", spec.Description))
	if spec.Category != "" {
		b.WriteString(fmt.Sprintf("**Category:** `%s`\n", spec.Category))
	}
	b.WriteString(fmt.Sprintf("**Signature:** `%s`\n\n", formatSignature(spec)))

	if len(spec.Args) > 0 {
		b.WriteString("**Arguments:**\n\n| Name | Type | Required | Description |\n| :--- | :--- | :--- | :--- |\n")
		for _, arg := range spec.Args {
			b.WriteString(fmt.Sprintf("| `%s` | `%s` | %t | %s |\n", arg.Name, arg.Type, arg.Required, arg.Description))
		}
		b.WriteString("\n")
	} else {
		b.WriteString("**Arguments:** None\n\n")
	}

	b.WriteString("**Returns:**\n")
	b.WriteString(fmt.Sprintf("- **Type:** `%s`", spec.ReturnType))
	if spec.ReturnHelp != "" {
		b.WriteString(fmt.Sprintf("\n- **Help:** %s\n", spec.ReturnHelp))
	} else {
		b.WriteString("\n")
	}

	if spec.Example != "" {
		b.WriteString(fmt.Sprintf("\n**Example:**\n```neuroscript\n%s\n```\n", spec.Example))
	}

	if spec.ErrorConditions != "" {
		b.WriteString(fmt.Sprintf("\n**Error Conditions:** %s\n", spec.ErrorConditions))
	}

	return b.String()
}
