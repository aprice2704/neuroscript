// NeuroScript Version: 0.3.0
// File version: 0.1.2
// AI Worker Management: Definition Management Methods (Error Handling Corrected, I/O Refactored)
// filename: pkg/core/ai_wm_definitions.go

package core

import (
	"fmt"
	"os"
	"path/filepath" // Added for directory creation
	"reflect"       // For deep comparison in update
	"strings"

	"github.com/google/uuid"
	// Ensure logging.Logger is correctly imported if not already covered by the main manager file
	// "github.com/aprice2704/neuroscript/pkg/logging"
)

// persistDefinitionsUnsafe prepares and writes the current definitions to their file.
// Assumes caller holds the write lock.
func (m *AIWorkerManager) persistDefinitionsUnsafe() error {
	jsonString, err := m.prepareDefinitionsForSaving() // From ai_wm.go
	if err != nil {
		// error already logged by prepareDefinitionsForSaving
		return err // Should be a RuntimeError
	}

	defPath := m.FullPathForDefinitions() // From ai_wm.go
	if defPath == "" {
		m.logger.Error("Cannot save definitions: file path is not configured in AIWorkerManager.")
		return NewRuntimeError(ErrorCodeConfiguration, "definitions file path not configured for saving", ErrConfiguration)
	}

	// Ensure directory exists
	dir := filepath.Dir(defPath)
	if mkDirErr := os.MkdirAll(dir, 0755); mkDirErr != nil {
		m.logger.Errorf("Failed to create directory '%s' for definitions file: %v", dir, mkDirErr)
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to create directory for definitions file '%s'", dir), mkDirErr)
	}

	if err := os.WriteFile(defPath, []byte(jsonString), 0644); err != nil {
		m.logger.Errorf("Failed to write definitions to file '%s': %v", defPath, err)
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to write definitions to file '%s'", defPath), err)
	}
	m.logger.Debugf("Successfully saved definitions to %s", defPath)
	return nil
}

// LoadWorkerDefinitionsFromFile is a public method for tools/external calls to reload definitions.
// It replaces all current definitions and re-initializes rate trackers.
func (m *AIWorkerManager) LoadWorkerDefinitionsFromFile() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	defPath := m.FullPathForDefinitions()
	if defPath == "" {
		m.logger.Error("Cannot load definitions: file path is not configured in AIWorkerManager.")
		// Consistent with NewAIWorkerManager: if path isn't set, proceed with empty but log error.
		m.definitions = make(map[string]*AIWorkerDefinition)
		m.activeInstances = make(map[string]*AIWorkerInstance)
		m.initializeRateTrackersUnsafe()
		return NewRuntimeError(ErrorCodeConfiguration, "definitions file path not configured, cannot load", ErrConfiguration)
	}

	m.logger.Infof("AIWorkerManager: Public request to load worker definitions from %s", defPath)

	// Clear existing state carefully
	m.definitions = make(map[string]*AIWorkerDefinition)
	m.activeInstances = make(map[string]*AIWorkerInstance) // Instances are ephemeral on full reload

	contentBytes, err := os.ReadFile(defPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.logger.Infof("AIWorkerManager: Definitions file '%s' not found. Manager will have no definitions.", defPath)
			m.initializeRateTrackersUnsafe() // Initialize for empty state
			return nil                       // Not an error if file simply doesn't exist
		}
		m.logger.Errorf("AIWorkerManager: Error reading definitions file '%s': %v", defPath, err)
		m.initializeRateTrackersUnsafe() // Initialize for empty state
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to read definitions file '%s'", defPath), err)
	}

	if loadErr := m.loadWorkerDefinitionsFromContent(contentBytes); loadErr != nil {
		// loadWorkerDefinitionsFromContent already logs errors and handles m.definitions state.
		m.initializeRateTrackersUnsafe() // Ensure trackers are set up even if loading had issues.
		return loadErr                   // This should be a RuntimeError
	}

	// Re-initialize rate trackers for the newly loaded definitions
	m.initializeRateTrackersUnsafe()

	m.logger.Infof("AIWorkerManager: Public load definitions complete. %d definitions loaded from %s.", len(m.definitions), defPath)
	return nil
}

// SaveWorkerDefinitionsToFile is a public method to persist the current state of definitions.
func (m *AIWorkerManager) SaveWorkerDefinitionsToFile() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	defPath := m.FullPathForDefinitions()
	m.logger.Infof("AIWorkerManager: Public request to save worker definitions to %s", defPath)
	return m.persistDefinitionsUnsafe()
}

// AddWorkerDefinition adds a new AI worker definition to the manager and persists it.
func (m *AIWorkerManager) AddWorkerDefinition(def AIWorkerDefinition) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if def.DefinitionID == "" {
		def.DefinitionID = uuid.NewString()
		m.logger.Debugf("AddWorkerDefinition: No DefinitionID provided for '%s', generated new: %s", def.Name, def.DefinitionID)
	} else if _, exists := m.definitions[def.DefinitionID]; exists {
		m.logger.Warnf("AddWorkerDefinition: Attempt to add definition with existing ID '%s'", def.DefinitionID)
		return "", NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("worker definition with ID '%s' already exists", def.DefinitionID), ErrInvalidArgument)
	}

	// Apply defaults
	if len(def.InteractionModels) == 0 {
		def.InteractionModels = []InteractionModelType{InteractionModelConversational}
	}
	if def.Status == "" {
		def.Status = DefinitionStatusActive
	}
	if def.Auth.Method == "" {
		def.Auth = APIKeySource{Method: APIKeyMethodNone}
	}
	// Ensure AggregatePerformanceSummary is initialized if not provided
	if def.AggregatePerformanceSummary == nil {
		def.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{}
	}

	m.definitions[def.DefinitionID] = &def
	m.initializeRateTrackerForDefinitionUnsafe(&def) // Initialize rate tracking for the new definition
	m.logger.Infof("AIWorkerManager: Added AIWorkerDefinition: ID=%s, Name=%s", def.DefinitionID, def.Name)

	if err := m.persistDefinitionsUnsafe(); err != nil {
		m.logger.Errorf("AIWorkerManager: Failed to save definitions after adding ID %s: %v", def.DefinitionID, err)
		// Return the original error from persistDefinitionsUnsafe if it's a RuntimeError
		if _, ok := err.(*RuntimeError); ok {
			return def.DefinitionID, err
		}
		return def.DefinitionID, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("definition added in memory but failed to save ID %s", def.DefinitionID), err)
	}
	return def.DefinitionID, nil
}

// GetWorkerDefinition retrieves a copy of an AI worker definition by its ID.
// It includes runtime information like current active instance count in the summary.
func (m *AIWorkerManager) GetWorkerDefinition(definitionID string) (*AIWorkerDefinition, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	def, exists := m.definitions[definitionID]
	if !exists {
		return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition with ID '%s' not found", definitionID), ErrNotFound)
	}

	// Create a copy to return, including dynamic runtime info in the summary
	defCopy := *def
	if def.AggregatePerformanceSummary != nil {
		summaryCopy := *def.AggregatePerformanceSummary // copy the struct
		defCopy.AggregatePerformanceSummary = &summaryCopy
	} else {
		// If the source summary is nil, the copy's summary should also be nil, or an empty one
		defCopy.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{} // Or nil, depending on desired behavior
		m.logger.Warnf("GetWorkerDefinition: DefinitionID %s has nil AggregatePerformanceSummary.", definitionID)
	}

	if tracker, ok := m.rateTrackers[def.DefinitionID]; ok && defCopy.AggregatePerformanceSummary != nil {
		defCopy.AggregatePerformanceSummary.ActiveInstancesCount = tracker.CurrentActiveInstances
	} else if defCopy.AggregatePerformanceSummary != nil {
		// This case should ideally not be hit if trackers are managed consistently
		defCopy.AggregatePerformanceSummary.ActiveInstancesCount = 0
		m.logger.Warnf("GetWorkerDefinition: No rate tracker found for definition ID '%s' when fetching active instance count, or summary was nil.", definitionID)
	}
	return &defCopy, nil
}

// ListWorkerDefinitions returns a list of AI worker definitions, optionally filtered.
// Includes runtime information like current active instance count in summaries.
func (m *AIWorkerManager) ListWorkerDefinitions(filters map[string]interface{}) []*AIWorkerDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*AIWorkerDefinition, 0, len(m.definitions))
	for _, def := range m.definitions {
		if m.matchesDefinitionFilters(def, filters) {
			defCopy := *def // Create a copy
			if def.AggregatePerformanceSummary != nil {
				summaryCopy := *def.AggregatePerformanceSummary // copy the struct
				defCopy.AggregatePerformanceSummary = &summaryCopy
			} else {
				defCopy.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{} // Or nil
				m.logger.Warnf("ListWorkerDefinitions: DefinitionID %s has nil AggregatePerformanceSummary.", def.DefinitionID)
			}

			if tracker, ok := m.rateTrackers[def.DefinitionID]; ok && defCopy.AggregatePerformanceSummary != nil {
				defCopy.AggregatePerformanceSummary.ActiveInstancesCount = tracker.CurrentActiveInstances
			} else if defCopy.AggregatePerformanceSummary != nil {
				defCopy.AggregatePerformanceSummary.ActiveInstancesCount = 0
				m.logger.Warnf("ListWorkerDefinitions: No rate tracker for definition ID '%s', or summary was nil. Active count set to 0.", def.DefinitionID)
			}
			list = append(list, &defCopy)
		}
	}
	return list
}

// UpdateWorkerDefinition updates an existing AI worker definition and persists changes.
// The 'updates' map contains fields to be changed.
func (m *AIWorkerManager) UpdateWorkerDefinition(definitionID string, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	def, exists := m.definitions[definitionID]
	if !exists {
		return NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition ID '%s' not found for update", definitionID), ErrNotFound)
	}

	changedFields := []string{}

	setChanged := func(fieldName string) {
		isNew := true
		for _, f := range changedFields {
			if f == fieldName {
				isNew = false
				break
			}
		}
		if isNew {
			changedFields = append(changedFields, fieldName)
		}
	}

	if val, _ok := updates["name"]; _ok {
		if v, ok := val.(string); ok && def.Name != v {
			def.Name = v
			setChanged("Name")
		}
	}
	if val, ok := updates["provider"]; ok {
		if v, ok := val.(string); ok && def.Provider != AIWorkerProvider(v) {
			def.Provider = AIWorkerProvider(v)
			setChanged("Provider")
		}
	}
	if val, ok := updates["model_name"]; ok {
		if v, ok := val.(string); ok && def.ModelName != v {
			def.ModelName = v
			setChanged("ModelName")
		}
	}

	if authMapVal, ok := updates["auth"]; ok {
		if authMap, mapOk := authMapVal.(map[string]interface{}); mapOk {
			newAuth := def.Auth // auth is a struct, so this is a copy
			authUpdated := false
			if methodVal, mOk := authMap["method"]; mOk {
				if methodStr, sOk := methodVal.(string); sOk && newAuth.Method != APIKeySourceMethod(methodStr) {
					newAuth.Method = APIKeySourceMethod(methodStr)
					authUpdated = true
				}
			}
			if valueVal, vOk := authMap["value"]; vOk {
				if valueStr, sOk := valueVal.(string); sOk && newAuth.Value != valueStr {
					newAuth.Value = valueStr
					authUpdated = true
				}
			}
			if authUpdated {
				def.Auth = newAuth
				setChanged("Auth")
			}
		}
	}

	if val, ok := updates["interaction_models"]; ok {
		if vSlice, ok := val.([]interface{}); ok {
			newIMs := []InteractionModelType{}
			for _, item := range vSlice {
				if s, sOk := item.(string); sOk {
					newIMs = append(newIMs, InteractionModelType(s))
				}
			}
			if !reflect.DeepEqual(def.InteractionModels, newIMs) {
				def.InteractionModels = newIMs
				setChanged("InteractionModels")
			}
		}
	}
	if val, ok := updates["capabilities"]; ok {
		if vSlice, ok := val.([]interface{}); ok {
			newCaps := []string{}
			for _, item := range vSlice {
				if s, sOk := item.(string); sOk {
					newCaps = append(newCaps, s)
				}
			}
			if !reflect.DeepEqual(def.Capabilities, newCaps) {
				def.Capabilities = newCaps
				setChanged("Capabilities")
			}
		}
	}
	if val, ok := updates["base_config"]; ok {
		if v, ok := val.(map[string]interface{}); ok && !reflect.DeepEqual(def.BaseConfig, v) {
			def.BaseConfig = v
			setChanged("BaseConfig")
		}
	}
	if val, ok := updates["cost_metrics"]; ok {
		if vMap, ok := val.(map[string]interface{}); ok {
			newMetrics := make(map[string]float64)
			changed := false
			for k, vi := range vMap {
				if f, fOk := toFloat64(vi); fOk {
					if currentVal, currentOk := def.CostMetrics[k]; !currentOk || currentVal != f {
						changed = true
					}
					newMetrics[k] = f
				}
			}
			// Check if any keys were removed
			if len(def.CostMetrics) != len(newMetrics) && !changed { // if lengths differ, it's a change unless already caught
				for k := range def.CostMetrics {
					if _, existsInNew := newMetrics[k]; !existsInNew {
						changed = true
						break
					}
				}
			}
			if changed || !reflect.DeepEqual(def.CostMetrics, newMetrics) { // Final deep equal for safety
				def.CostMetrics = newMetrics
				setChanged("CostMetrics")
			}
		}
	}

	if val, ok := updates["rate_limits"]; ok {
		if vMap, ok := val.(map[string]interface{}); ok {
			newLimits := def.RateLimits // RateLimits is a struct, this is a copy
			updatedLimits := false
			if v, fOk := toInt64(vMap["max_requests_per_minute"]); fOk && newLimits.MaxRequestsPerMinute != int(v) {
				newLimits.MaxRequestsPerMinute = int(v)
				updatedLimits = true
			}
			if v, fOk := toInt64(vMap["max_tokens_per_minute"]); fOk && newLimits.MaxTokensPerMinute != int(v) {
				newLimits.MaxTokensPerMinute = int(v)
				updatedLimits = true
			}
			if v, fOk := toInt64(vMap["max_tokens_per_day"]); fOk && newLimits.MaxTokensPerDay != int(v) {
				newLimits.MaxTokensPerDay = int(v)
				updatedLimits = true
			}
			if v, fOk := toInt64(vMap["max_concurrent_active_instances"]); fOk && newLimits.MaxConcurrentActiveInstances != int(v) {
				newLimits.MaxConcurrentActiveInstances = int(v)
				updatedLimits = true
			}

			if updatedLimits {
				def.RateLimits = newLimits
				setChanged("RateLimits")
			}
		}
	}
	if val, ok := updates["status"]; ok {
		if v, ok := val.(string); ok && def.Status != AIWorkerDefinitionStatus(v) {
			def.Status = AIWorkerDefinitionStatus(v)
			setChanged("Status")
		}
	}
	if val, ok := updates["default_file_contexts"]; ok {
		if vSlice, ok := val.([]interface{}); ok {
			newCtx := []string{}
			for _, item := range vSlice {
				if s, sOk := item.(string); sOk {
					newCtx = append(newCtx, s)
				}
			}
			if !reflect.DeepEqual(def.DefaultFileContexts, newCtx) {
				def.DefaultFileContexts = newCtx
				setChanged("DefaultFileContexts")
			}
		}
	}
	if val, ok := updates["metadata"]; ok {
		if v, ok := val.(map[string]interface{}); ok && !reflect.DeepEqual(def.Metadata, v) {
			def.Metadata = v
			setChanged("Metadata")
		}
	}

	if len(changedFields) > 0 {
		m.logger.Infof("AIWorkerManager: Updating AIWorkerDefinition ID '%s'. Changed fields: %v", definitionID, changedFields)
		err := m.persistDefinitionsUnsafe()
		if err != nil {
			// Return the original error from persistDefinitionsUnsafe if it's a RuntimeError
			if _, ok := err.(*RuntimeError); ok {
				return err
			}
			return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to save updated definition ID %s", definitionID), err)
		}
		return nil
	}
	m.logger.Infof("AIWorkerManager: UpdateWorkerDefinition ID '%s': No effective changes.", definitionID)
	return nil
}

// RemoveWorkerDefinition removes a definition if no active instances are using it.
func (m *AIWorkerManager) RemoveWorkerDefinition(definitionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.definitions[definitionID]; !exists {
		return NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition ID '%s' not found for removal", definitionID), ErrNotFound)
	}

	if tracker, ok := m.rateTrackers[definitionID]; ok && tracker.CurrentActiveInstances > 0 {
		m.logger.Warnf("RemoveWorkerDefinition: Cannot remove definition '%s', it has %d active instances.", definitionID, tracker.CurrentActiveInstances)
		return NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("cannot remove definition '%s', it has %d active instances. Retire instances first.", definitionID, tracker.CurrentActiveInstances), ErrFailedPrecondition)
	}

	delete(m.definitions, definitionID)
	delete(m.rateTrackers, definitionID) // Also remove its rate tracker
	m.logger.Infof("AIWorkerManager: Removed AIWorkerDefinition: ID=%s", definitionID)
	err := m.persistDefinitionsUnsafe()
	if err != nil {
		// Return the original error from persistDefinitionsUnsafe if it's a RuntimeError
		if _, ok := err.(*RuntimeError); ok {
			return err
		}
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to save after removing definition ID %s", definitionID), err)
	}
	return nil
}

// matchesDefinitionFilters is a helper to check if a definition matches given criteria.
// Called with RLock typically.
func (m *AIWorkerManager) matchesDefinitionFilters(def *AIWorkerDefinition, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}
	if def == nil { // Should not happen if called from ListWorkerDefinitions where defs are from m.definitions
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
				if len(capListInterfaces) == 0 { // an empty list of required capabilities means all definitions match this criterion
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
			m.logger.Debugf("AIWorkerManager.matchesDefinitionFilters: Unknown or unhandled filter key '%s'", filterKey)
		}

		if !match {
			return false
		}
	}
	return true
}
