// NeuroScript Version: 0.3.1
// File version: 0.2.15
// filename: pkg/core/ai_wm_helpers.go
// Changes:
// - Replaced undefined ErrorCodeItemNotFound with ErrorCodeKeyNotFound.
// - Replaced undefined ErrItemNotFound with ErrNotFound or ErrConfiguration as appropriate.
package core

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
	// "sync" // Not used in the provided snippet for these functions
	// "github.com/aprice2704/neuroscript/pkg/logging" // Not used in the provided snippet
	// "github.com/google/uuid" // Not used in the provided snippet
)

// GetDefinitionIDByName resolves an AIWorkerDefinition's ID from its unique name.
// It returns the DefinitionID and nil if found, or an empty string and an error if not found.
// This function assumes it's a method of AIWorkerManager or has access to its fields (m.definitions, m.logger, m.mu).
// The provided snippet doesn't show the AIWorkerManager struct definition here, so I'm keeping the method signature
// as it was in the snippet. If this is not a method of AIWorkerManager, `m.` accesses will fail.
// Assuming 'm' is a pointer to AIWorkerManager and relevant fields (definitions, logger, mu) are accessible.
func (m *AIWorkerManager) GetDefinitionIDByName(name string) (string, error) {
	if name == "" {
		return "", lang.NewRuntimeError(ErrorCodeArgMismatch, "worker definition name cannot be empty", ErrInvalidArgument)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for id, def := range m.definitions {
		if def != nil && def.Name == name {
			if m.logger != nil { // Defensive check for logger
				m.logger.Debugf("GetDefinitionIDByName: Found ID '%s' for name '%s'", id, name)
			}
			return id, nil
		}
	}

	if m.logger != nil { // Defensive check
		m.logger.Warnf("GetDefinitionIDByName: Worker definition with name '%s' not found.", name)
	}
	// Corrected: ErrorCodeItemNotFound -> ErrorCodeKeyNotFound, ErrItemNotFound -> ErrNotFound
	return "", lang.NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition name '%s' not found", name), ErrNotFound)
}

// GetDefinition retrieves an active AIWorkerDefinition by its ID.
// This is a helper that might be called after resolving name to ID.
// Assuming 'm' is a pointer to AIWorkerManager.
func (m *AIWorkerManager) GetDefinition(definitionID string) (*AIWorkerDefinition, error) {
	if definitionID == "" {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "worker definition ID cannot be empty", ErrInvalidArgument)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	def, exists := m.definitions[definitionID]
	if !exists || def == nil {
		// Corrected: ErrorCodeItemNotFound -> ErrorCodeKeyNotFound, ErrItemNotFound -> ErrNotFound
		return nil, lang.NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition ID '%s' not found", definitionID), ErrNotFound)
	}

	if def.Status != DefinitionStatusActive {
		// Corrected: ErrItemNotFound -> ErrConfiguration (to match ErrorCodeConfiguration)
		return nil, lang.NewRuntimeError(ErrorCodeConfiguration, fmt.Sprintf("worker definition '%s' (ID: %s) is not active (status: %s)", def.Name, def.DefinitionID, def.Status), ErrConfiguration)
	}
	return def, nil
}

// Note: The rest of the AIWorkerManager methods or other helper functions that might have been
// in the original ai_wm_helpers.go are not included here as they were not in the
// snippet provided in the context for this specific correction task.
// If other functions in this file used the undefined errors, they would need similar corrections.
