// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 13:55:02 PDT // Add package comment
// filename: pkg/toolsets/register.go

// Package toolsets provides central registration for extended NeuroScript toolsets.
// It decouples core interpreter initialization from specific non-core tool implementations
// by allowing tool packages to register themselves via init().
package toolsets

import (
	"errors"
	"fmt"
	"sync" // Added for safe concurrent access to the registry map

	// Core package for registry type and registrar interface
	"github.com/aprice2704/neuroscript/pkg/core"
	// --- REMOVED direct imports of specific tool packages ---
	// "github.com/aprice2704/neuroscript/pkg/neurodata/blocks"
	// "github.com/aprice2704/neuroscript/pkg/neurodata/checklist"
)

// ToolRegisterFunc defines the function signature expected for registering a toolset.
// It matches the signature of functions like checklist.RegisterChecklistTools.
type ToolRegisterFunc func(registry core.ToolRegistrar) error // Use core.ToolRegistrar interface

// --- Registry for Toolset Registration Functions ---

var (
	// Use a mutex for safe concurrent access during init potentially
	registrationMu       sync.RWMutex
	toolsetRegistrations = make(map[string]ToolRegisterFunc)
)

// AddToolsetRegistration is called by tool packages (typically in their init() function)
// to register their main registration function with the toolsets package.
func AddToolsetRegistration(name string, regFunc ToolRegisterFunc) {
	registrationMu.Lock()
	defer registrationMu.Unlock()

	if _, exists := toolsetRegistrations[name]; exists {
		// Log or handle duplicate registration attempts if necessary
		// For now, allow overwrite (last one wins) but could panic or log warning.
		fmt.Printf("[WARN] Toolset registration function for '%s' overwritten.\n", name)
	}
	if regFunc == nil {
		panic(fmt.Sprintf("attempted to register nil registration function for toolset '%s'", name))
	}
	toolsetRegistrations[name] = regFunc
	fmt.Printf("Toolset registration function added for: %s\n", name) // Debug output
}

// RegisterExtendedTools registers all non-core toolsets that have added themselves
// via AddToolsetRegistration.
func RegisterExtendedTools(registry core.ToolRegistrar) error { // Accept interface
	registrationMu.RLock() // Read lock while iterating
	defer registrationMu.RUnlock()

	if registry == nil {
		return fmt.Errorf("cannot register extended tools: registry is nil")
	}

	fmt.Printf("Registering %d discovered extended toolsets...\n", len(toolsetRegistrations)) // Debug

	var allErrors []error

	// --- Iterate and call registered functions ---
	for name, regFunc := range toolsetRegistrations {
		fmt.Printf("Calling registration function for: %s\n", name) // Debug
		if err := regFunc(registry); err != nil {
			// Wrap the error with context about which toolset failed
			allErrors = append(allErrors, fmt.Errorf("failed registering %s tools: %w", name, err))
		}
	}

	// --- Error Handling ---
	if len(allErrors) > 0 {
		// Use errors.Join (available since Go 1.20)
		return errors.Join(allErrors...)
	}

	fmt.Println("Extended tools registered successfully via toolsets.") // Debug output
	return nil                                                          // Success
}
