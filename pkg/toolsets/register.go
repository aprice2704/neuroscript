// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 13:55:02 PDT // Add package comment
// filename: pkg/toolsets/register.go

// Package toolsets provides central registration for extended NeuroScript toolsets.
// It decouples core interpreter initialization from specific non-core tool implementations.
package toolsets

import (
	"errors" // Using errors package for Join
	"fmt"

	// Core package for registry type
	"github.com/aprice2704/neuroscript/pkg/core"

	// --- Import ALL packages that provide tool registration functions ---
	"github.com/aprice2704/neuroscript/pkg/neurodata/blocks"
	"github.com/aprice2704/neuroscript/pkg/neurodata/checklist"
	// Import other tool packages here as they are created...
	// e.g., "github.com/aprice2704/neuroscript/pkg/goast"
)

// RegisterExtendedTools registers all non-core toolsets.
// It assumes core tools (FS, String, Math, Tree, etc.) are registered elsewhere
// (e.g., during interpreter initialization via core.RegisterCoreTools).
func RegisterExtendedTools(registry *core.ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("cannot register extended tools: registry is nil")
	}

	var allErrors []error

	// Helper function to append errors gracefully
	collectErr := func(toolsetName string, err error) {
		if err != nil {
			// Wrap the error with context about which toolset failed
			allErrors = append(allErrors, fmt.Errorf("failed registering %s tools: %w", toolsetName, err))
		}
	}

	// --- Register Individual Extended Toolsets ---
	fmt.Println("Registering Checklist tools...") // Temporary debug output
	collectErr("Checklist", checklist.RegisterChecklistTools(registry))

	fmt.Println("Registering Blocks tools...") // Temporary debug output
	collectErr("Blocks", blocks.RegisterBlockTools(registry))

	// Add calls for other toolsets here...
	// fmt.Println("Registering GoAST tools...")
	// collectErr("GoAST", goast.RegisterGoASTTools(registry))

	// --- Error Handling ---
	if len(allErrors) > 0 {
		// Use errors.Join (available since Go 1.20)
		return errors.Join(allErrors...)
	}

	fmt.Println("Extended tools registered successfully.") // Temporary debug output
	return nil                                             // Success
}
