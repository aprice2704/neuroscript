// NeuroScript Version: 0.8.0
// File version: 7
// Purpose: Corrects the listTools implementation to return a slice of maps, making it compatible with the lang.Wrap function.
// filename: pkg/tool/meta/tools_meta.go
// nlines: 106
// risk_rating: MEDIUM

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

// RegisterTools registers all the tools in the meta package with the provided registrar.
func RegisterTools(registrar tool.ToolRegistrar) error {
	for _, t := range metaToolsToRegister {
		if _, err := registrar.RegisterTool(t); err != nil {
			return fmt.Errorf("failed to register meta tool '%s': %w", t.Spec.Name, err)
		}
	}
	return nil
}

// GetTool is the implementation for the 'tool.meta.getTool' tool.
func GetTool(rt tool.Runtime, args []any) (any, error) {
	if len(args) < 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "getTool: expects one argument: fullName", lang.ErrArgumentMismatch)
	}
	toolName, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("getTool: 'fullName' argument must be a string, got %T", args[0]), lang.ErrInvalidArgument)
	}

	impl, found := rt.ToolRegistry().GetTool(types.FullName(toolName))

	return map[string]any{
		"found": found,
		"spec":  impl.Spec,
	}, nil
}

// ListTools is the implementation for the 'tool.meta.listTools' tool.
// It returns a slice of maps to be compatible with lang.Wrap.
func ListTools(rt tool.Runtime, args []any) (any, error) {
	tools := rt.ToolRegistry().ListTools()
	specs := make([]tool.ToolSpec, len(tools))
	for i, t := range tools {
		specs[i] = t.Spec
	}
	// Sort for deterministic output
	sort.Slice(specs, func(i, j int) bool {
		return strings.ToLower(string(specs[i].FullName)) < strings.ToLower(string(specs[j].FullName))
	})

	// Convert slice of structs to slice of maps via JSON
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

	specs, err := ListTools(rt, nil)
	if err != nil {
		return nil, err // Should not happen
	}

	jsonData, err := json.MarshalIndent(specs, "", "  ")
	if err != nil {
		rt.GetLogger().Errorf("GetToolSpecificationsJSON: Failed to marshal tool specs: %v", err)
		return "", lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal tool specifications to JSON", err)
	}

	return string(jsonData), nil
}
