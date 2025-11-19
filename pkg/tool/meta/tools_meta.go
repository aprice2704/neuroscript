// NeuroScript Major Version: 1
// File version: 18
// Purpose: Implements meta tools. Uses lang.UnwrapValue for better constant display.
// Latest change: Updated ListToolNames to accept an optional filter argument.
// filename: pkg/tool/meta/tools_meta.go
// nlines: 330

package meta

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
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

// ListTools returns a list of tool specifications, optionally filtered.
func ListTools(rt tool.Runtime, args []any) (any, error) {
	filter := ""
	if len(args) > 0 && args[0] != nil {
		var ok bool
		filter, ok = args[0].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ListTools: filter argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
		}
		filter = strings.ToLower(filter)
	}

	tools := rt.ToolRegistry().ListTools()
	specs := make([]tool.ToolSpec, 0, len(tools))

	for _, t := range tools {
		if filter == "" || strings.Contains(strings.ToLower(string(t.Spec.FullName)), filter) {
			specs = append(specs, t.Spec)
		}
	}

	sort.Slice(specs, func(i, j int) bool {
		return strings.ToLower(string(specs[i].FullName)) < strings.ToLower(string(specs[j].FullName))
	})

	var specMaps []any
	jsonData, err := json.Marshal(specs)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal tool specs to JSON for wrapping", err)
	}
	if err := json.Unmarshal(jsonData, &specMaps); err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to unmarshal tool spec JSON for wrapping", err)
	}

	return specMaps, nil
}

// GetToolSpecificationsJSON provides a JSON string of all available tool specifications.
func GetToolSpecificationsJSON(rt tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "GetToolSpecificationsJSON: expects no arguments", lang.ErrArgumentMismatch)
	}

	registry := rt.ToolRegistry()
	if registry == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "GetToolSpecificationsJSON: ToolRegistry is not available", lang.ErrConfiguration)
	}

	specMaps, err := ListTools(rt, nil)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.MarshalIndent(specMaps, "", "  ")
	if err != nil {
		if rt.GetLogger() != nil {
			rt.GetLogger().Errorf("GetToolSpecificationsJSON: Failed to marshal tool specs: %v", err)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: GetToolSpecificationsJSON: Failed to marshal tool specs: %v\n", err)
		}
		return "", lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal tool specifications to JSON", err)
	}

	return string(jsonData), nil
}

// ListToolNames provides a simple, newline-separated list of all available tool signatures.
func ListToolNames(rt tool.Runtime, args []interface{}) (interface{}, error) {
	filter := ""
	if len(args) > 0 && args[0] != nil {
		var ok bool
		filter, ok = args[0].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ListToolNames: filter argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
		}
		filter = strings.ToLower(filter)
	}

	tools := rt.ToolRegistry().ListTools()
	specs := make([]tool.ToolSpec, 0, len(tools))

	for _, t := range tools {
		if filter == "" || strings.Contains(strings.ToLower(string(t.Spec.FullName)), filter) {
			specs = append(specs, t.Spec)
		}
	}

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
	if len(args) == 1 && args[0] != nil {
		var ok bool
		filter, ok = args[0].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ToolsHelp: filter argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
		}
		filter = strings.ToLower(filter)
	}

	tools := rt.ToolRegistry().ListTools()
	specs := make([]tool.ToolSpec, 0, len(tools))

	for _, t := range tools {
		specs = append(specs, t.Spec)
	}

	sort.Slice(specs, func(i, j int) bool {
		return strings.ToLower(string(specs[i].FullName)) < strings.ToLower(string(specs[j].FullName))
	})

	var b strings.Builder
	count := 0
	for _, spec := range specs {
		if filter == "" || strings.Contains(strings.ToLower(string(spec.FullName)), filter) {
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

// --- Introspection Tools ---

// constProvider is a local interface matching api.Interpreter methods needed for constants.
// This avoids importing pkg/api.
type constProvider interface {
	KnownGlobalConstants() map[string]lang.Value
}

// ListGlobalConstants returns a map of known global constants.
func ListGlobalConstants(rt tool.Runtime, args []any) (any, error) {
	filter := ""
	if len(args) > 0 && args[0] != nil {
		var ok bool
		filter, ok = args[0].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ListGlobalConstants: filter argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
		}
		filter = strings.ToLower(filter)
	}

	cp, ok := rt.(constProvider)
	if !ok {
		return map[string]any{}, nil
	}

	consts := cp.KnownGlobalConstants()
	result := make(map[string]any)

	for name, val := range consts {
		if filter == "" || strings.Contains(strings.ToLower(name), filter) {
			// FIX: Use lang.UnwrapValue to get native types (e.g., float64, string)
			// instead of stringified representations.
			result[name] = lang.UnwrapValue(val)
		}
	}

	return result, nil
}

// ListFunctions returns a list of known function names.
// Uses reflection to avoid importing pkg/ast or pkg/api.
func ListFunctions(rt tool.Runtime, args []any) (any, error) {
	filter := ""
	if len(args) > 0 && args[0] != nil {
		var ok bool
		filter, ok = args[0].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("ListFunctions: filter argument must be a string, got %T", args[0]), lang.ErrArgumentMismatch)
		}
		filter = strings.ToLower(filter)
	}

	// Use reflection to find "KnownProcedures" on the runtime object
	val := reflect.ValueOf(rt)
	method := val.MethodByName("KnownProcedures")

	if !method.IsValid() {
		return []string{}, nil
	}

	// Call the method (it takes no args)
	res := method.Call(nil)
	if len(res) == 0 {
		return []string{}, nil
	}

	// Result [0] should be the map
	mapVal := res[0]
	if mapVal.Kind() != reflect.Map {
		return []string{}, nil
	}

	keys := mapVal.MapKeys()
	names := make([]string, 0, len(keys))

	for _, k := range keys {
		name := k.String()
		if filter == "" || strings.Contains(strings.ToLower(name), filter) {
			names = append(names, name)
		}
	}

	sort.Strings(names)
	return names, nil
}

// --- Helper Functions ---

func formatSignature(spec tool.ToolSpec) string {
	var argParts []string
	for _, arg := range spec.Args {
		argParts = append(argParts, fmt.Sprintf("%s:%s", arg.Name, arg.Type))
	}
	argString := strings.Join(argParts, ", ")
	return fmt.Sprintf("%s(%s) -> %s", spec.FullName, argString, spec.ReturnType)
}

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
