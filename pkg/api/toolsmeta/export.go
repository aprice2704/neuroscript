// NeuroScript Version: 0.6.0
// File version: 3
// Purpose: Provides an exportable function for external projects to generate tool metadata. FIX: Now accepts a tool.ToolRegistry to decouple it from interpreter creation.
// filename: pkg/api/toolsmeta/export.go
// nlines: 45
// risk_rating: LOW

package toolsmeta

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/tool"
)

// ExportTools generates a JSON metadata file containing the definitions of all
// tools in the provided tool registry.
// outputFile is the path where the JSON file will be written.
func ExportTools(registry tool.ToolRegistry, outputFile string) error {
	if registry == nil {
		return fmt.Errorf("cannot export tools from a nil registry")
	}

	// 1. Get the list of all registered tools from the provided registry.
	allTools := registry.ListTools()
	if len(allTools) == 0 {
		return fmt.Errorf("no tools found in the registry; ensure the registry is populated before exporting")
	}

	// 2. Serialize the full ToolImplementation for each tool.
	jsonData, err := json.MarshalIndent(allTools, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tool implementations to JSON: %w", err)
	}

	// 3. Write to the output file.
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write to output file %s: %w", outputFile, err)
	}

	return nil
}
