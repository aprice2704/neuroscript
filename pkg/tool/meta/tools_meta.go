// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Implements meta tools. Removed deprecated GetTool function.
// filename: pkg/tool/meta/tools_meta.go
// nlines: 91
// risk_rating: MEDIUM

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
	fmt.Fprintf(os.Stderr, "DEBUG: Entered ListTools implementation\n")
	tools := rt.ToolRegistry().ListTools()
	specs := make([]tool.ToolSpec, len(tools))
	for i, t := range tools {
		specs[i] = t.Spec
	}
	fmt.Fprintf(os.Stderr, "DEBUG: ListTools: Found %d tools in registry\n", len(specs))

	// Sort for deterministic output
	sort.Slice(specs, func(i, j int) bool {
		return strings.ToLower(string(specs[i].FullName)) < strings.ToLower(string(specs[j].FullName))
	})

	// Convert slice of structs to slice of maps via JSON
	var specMaps []any
	jsonData, err := json.Marshal(specs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: ListTools: Failed to marshal specs: %v\n", err)
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal tool specs to JSON for wrapping", err)
	}
	if err := json.Unmarshal(jsonData, &specMaps); err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: ListTools: Failed to unmarshal specs: %v\n", err)
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to unmarshal tool spec JSON for wrapping", err)
	}

	fmt.Fprintf(os.Stderr, "DEBUG: ListTools: Successfully converted specs to %d maps\n", len(specMaps))
	return specMaps, nil
}

// GetToolSpecificationsJSON provides a JSON string of all available tool specifications.
func GetToolSpecificationsJSON(rt tool.Runtime, args []interface{}) (interface{}, error) {
	fmt.Fprintf(os.Stderr, "DEBUG: Entered GetToolSpecificationsJSON implementation\n")
	if len(args) != 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "GetToolSpecificationsJSON: expects no arguments", lang.ErrArgumentMismatch)
	}

	registry := rt.ToolRegistry()
	if registry == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "GetToolSpecificationsJSON: ToolRegistry is not available", lang.ErrConfiguration)
	}

	specs, err := ListTools(rt, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: GetToolSpecificationsJSON: Error calling ListTools: %v\n", err)
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

	fmt.Fprintf(os.Stderr, "DEBUG: GetToolSpecificationsJSON: Returning JSON string of %d bytes\n", len(jsonData))
	return string(jsonData), nil
}
