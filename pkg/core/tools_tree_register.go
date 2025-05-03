// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 17:32:05 PDT // Register TreeAddNode
// filename: pkg/core/tools_tree_register.go

// Package core contains core interpreter functionality, including built-in tools.
package core

import (
	"errors"
	"fmt"
)

// registerTreeTools registers the core tree manipulation tools (load, nav, find, modify).
// Rendering tools are registered separately in registerTreeRenderTools.
func registerTreeTools(registry *ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("registerTreeTools called with nil registry")
	}

	toolsToRegister := []ToolImplementation{
		// Load
		toolTreeLoadJSONImpl,
		// Navigation
		toolTreeGetNodeImpl,
		toolTreeGetChildrenImpl,
		toolTreeGetParentImpl,
		// Find
		toolTreeFindNodesImpl,
		// Modify
		toolTreeModifyNodeImpl,      // Modifies simple Value
		toolTreeSetAttributeImpl,    // Adds/Updates an attribute on an object node
		toolTreeRemoveAttributeImpl, // Removes an attribute from an object node
		toolTreeAddNodeImpl,         // <<< ADDED: Adds a new node as a child
		// Add TreeRemoveNodeImpl etc. here when created
	}

	var registrationErrors []error
	for _, tool := range toolsToRegister {
		if err := registry.RegisterTool(tool); err != nil {
			// Log the error and collect it
			// Consider using interpreter logger if available, fallback to fmt
			fmt.Printf("! Error registering tree tool %s: %v\n", tool.Spec.Name, err)
			registrationErrors = append(registrationErrors, fmt.Errorf("failed to register tree tool %q: %w", tool.Spec.Name, err))
		}
	}

	// Return a combined error if any registration failed
	if len(registrationErrors) > 0 {
		return errors.Join(registrationErrors...)
	}

	return nil // Success
}
