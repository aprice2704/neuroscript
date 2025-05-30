// NeuroScript Version: 0.3.0
// File version: 0.1.1
// Purpose: AI Worker Management: Read-only access to loaded AIWorkerDefinitions. CRUD operations removed as definitions are immutable post-load.
// filename: pkg/core/ai_wm_definitions_crud.go
// nlines: 85 // Approximate
// risk_rating: LOW
package core

import (
	"fmt"
	"strings"
	// "time" // No longer needed by functions in this file
	// "github.com/google/uuid" // No longer needed by functions in this file
	// "reflect" // No longer needed by functions in this file
)

// GetWorkerDefinition retrieves a copy of an AI worker definition by its ID.
// The returned definition is a copy to ensure the manager's internal state is not inadvertently modified,
// especially the mutable AggregatePerformanceSummary.
func (m *AIWorkerManager) GetWorkerDefinition(definitionID string) (*AIWorkerDefinition, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	def, exists := m.definitions[definitionID]
	if !exists {
		return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition with ID '%s' not found", definitionID), ErrNotFound)
	}

	// Return a copy, primarily to snapshot AggregatePerformanceSummary which can be updated.
	defCopy := *def
	if def.AggregatePerformanceSummary != nil {
		summaryCopy := *def.AggregatePerformanceSummary
		defCopy.AggregatePerformanceSummary = &summaryCopy
	} else {
		// This case should ideally not happen if definitions are well-formed at load time.
		defCopy.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{}
		m.logger.Debugf("GetWorkerDefinition: Definition (Name: '%s', ID: '%s') has nil AggregatePerformanceSummary, initialized empty for copy.", def.Name, definitionID)
	}

	// Populate dynamic count from rate tracker
	if tracker, ok := m.rateTrackers[def.DefinitionID]; ok && defCopy.AggregatePerformanceSummary != nil {
		defCopy.AggregatePerformanceSummary.ActiveInstancesCount = tracker.CurrentActiveInstances
	} else if defCopy.AggregatePerformanceSummary != nil {
		defCopy.AggregatePerformanceSummary.ActiveInstancesCount = 0
	}
	return &defCopy, nil
}

// ListWorkerDefinitions returns a list of copies of AI worker definitions, optionally filtered.
// Copies are returned to ensure the manager's internal state is not inadvertently modified.
func (m *AIWorkerManager) ListWorkerDefinitions(filters map[string]interface{}) []*AIWorkerDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*AIWorkerDefinition, 0, len(m.definitions))
	for _, def := range m.definitions {
		if m.matchesDefinitionFilters(def, filters) {
			defCopy := *def // Create a shallow copy of the definition
			if def.AggregatePerformanceSummary != nil {
				summaryCopy := *def.AggregatePerformanceSummary // Deep copy the summary
				defCopy.AggregatePerformanceSummary = &summaryCopy
			} else {
				defCopy.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{}
				m.logger.Debugf("ListWorkerDefinitions: Definition (Name: '%s', ID: '%s') has nil AggregatePerformanceSummary, initialized empty for copy.", def.Name, def.DefinitionID)
			}

			if tracker, ok := m.rateTrackers[def.DefinitionID]; ok && defCopy.AggregatePerformanceSummary != nil {
				defCopy.AggregatePerformanceSummary.ActiveInstancesCount = tracker.CurrentActiveInstances
			} else if defCopy.AggregatePerformanceSummary != nil {
				defCopy.AggregatePerformanceSummary.ActiveInstancesCount = 0
			}
			list = append(list, &defCopy)
		}
	}
	return list
}

// matchesDefinitionFilters is a helper to check if a definition matches given criteria.
// This function operates on the principle that definitions are read-only.
func (m *AIWorkerManager) matchesDefinitionFilters(def *AIWorkerDefinition, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}
	if def == nil {
		m.logger.Warnf("matchesDefinitionFilters called with a nil definition.")
		return false
	}

	for key, expectedValue := range filters {
		filterKey := strings.ToLower(key)
		match := false

		switch filterKey {
		case "definitionid", "id":
			if id, ok := expectedValue.(string); ok && def.DefinitionID == id {
				match = true
			}
		case "name":
			if name, ok := expectedValue.(string); ok && strings.Contains(strings.ToLower(def.Name), strings.ToLower(name)) {
				match = true
			}
		case "provider":
			if providerStr, ok := expectedValue.(string); ok && def.Provider == AIWorkerProvider(providerStr) {
				match = true
			}
		case "modelname", "model_name":
			if model, ok := expectedValue.(string); ok && strings.Contains(strings.ToLower(def.ModelName), strings.ToLower(model)) {
				match = true
			}
		case "status":
			if statusStr, ok := expectedValue.(string); ok && def.Status == AIWorkerDefinitionStatus(statusStr) {
				match = true
			}
		case "interactionmodels_contains", "interactionmodel_contains":
			if modelStr, ok := expectedValue.(string); ok {
				for _, im := range def.InteractionModels {
					if im == InteractionModelType(modelStr) {
						match = true
						break
					}
				}
			}
		case "capabilities_contains":
			if capValStr, ok := expectedValue.(string); ok {
				capValLower := strings.ToLower(capValStr)
				for _, c := range def.Capabilities {
					if strings.ToLower(c) == capValLower {
						match = true
						break
					}
				}
			}
		case "capabilities_contains_all":
			if capListInterfaces, ok := expectedValue.([]interface{}); ok {
				if len(capListInterfaces) == 0 {
					match = true
					break
				}
				allFound := true
				for _, capInterface := range capListInterfaces {
					capStr, capStrOk := capInterface.(string)
					if !capStrOk {
						allFound = false
						break
					}
					foundThisCap := false
					capStrLower := strings.ToLower(capStr)
					for _, workerCap := range def.Capabilities {
						if strings.ToLower(workerCap) == capStrLower {
							foundThisCap = true
							break
						}
					}
					if !foundThisCap {
						allFound = false
						break
					}
				}
				if allFound {
					match = true
				}
			}
		default:
			m.logger.Debugf("AIWorkerManager.matchesDefinitionFilters: Unknown or unhandled filter key '%s' for definition (Name: '%s', ID: '%s')", filterKey, def.Name, def.DefinitionID)
		}

		if !match {
			return false
		}
	}
	return true
}
