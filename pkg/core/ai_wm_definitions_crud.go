// NeuroScript Version: 0.3.0
// File version: 0.1.0
// AI Worker Management: Definition CRUD and Listing Methods
// filename: pkg/core/ai_wm_definitions_crud.go
// nlines: 230 // Approximate
// risk_rating: MEDIUM
package core

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AddWorkerDefinition adds a new AI worker definition to the manager and persists it.
func (m *AIWorkerManager) AddWorkerDefinition(def AIWorkerDefinition) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if def.DefinitionID == "" {
		def.DefinitionID = uuid.NewString()
		m.logger.Debugf("AddWorkerDefinition: No DefinitionID provided for (Name: '%s'), generated new: %s", def.Name, def.DefinitionID)
	} else if _, exists := m.definitions[def.DefinitionID]; exists {
		m.logger.Warnf("AddWorkerDefinition: Attempt to add definition with existing ID '%s' (Name: '%s')", def.DefinitionID, def.Name)
		return "", NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("worker definition with ID '%s' (Name: '%s') already exists", def.DefinitionID, def.Name), ErrInvalidArgument)
	}

	for id, existingDef := range m.definitions {
		if existingDef.Name == def.Name {
			m.logger.Warnf("AddWorkerDefinition: Attempt to add definition with name '%s' which is already used by definition ID '%s'.", def.Name, id)
			return "", NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("worker definition name '%s' already exists for ID '%s'", def.Name, id), ErrInvalidArgument)
		}
	}

	if len(def.InteractionModels) == 0 {
		def.InteractionModels = []InteractionModelType{InteractionModelConversational}
	}
	if def.Status == "" {
		def.Status = DefinitionStatusActive
	}
	if def.Auth.Method == "" {
		def.Auth = APIKeySource{Method: APIKeyMethodNone}
	}
	if def.AggregatePerformanceSummary == nil {
		def.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{}
	}
	def.CreatedTimestamp = time.Now()
	def.ModifiedTimestamp = def.CreatedTimestamp

	m.definitions[def.DefinitionID] = &def
	m.initializeRateTrackerForDefinitionUnsafe(&def)
	m.logger.Infof("AIWorkerManager: Added AIWorkerDefinition: Name='%s', ID=%s", def.Name, def.DefinitionID)

	if err := m.persistDefinitionsUnsafe(); err != nil {
		m.logger.Errorf("AIWorkerManager: Failed to save definitions after adding (Name: '%s', ID: %s): %v", def.Name, def.DefinitionID, err)
		if _, ok := err.(*RuntimeError); ok {
			return def.DefinitionID, err
		}
		return def.DefinitionID, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("definition (Name: '%s', ID: %s) added in memory but failed to save", def.Name, def.DefinitionID), err)
	}
	return def.DefinitionID, nil
}

// GetWorkerDefinition retrieves a copy of an AI worker definition by its ID.
func (m *AIWorkerManager) GetWorkerDefinition(definitionID string) (*AIWorkerDefinition, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	def, exists := m.definitions[definitionID]
	if !exists {
		return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition with ID '%s' not found", definitionID), ErrNotFound)
	}

	defCopy := *def
	if def.AggregatePerformanceSummary != nil {
		summaryCopy := *def.AggregatePerformanceSummary
		defCopy.AggregatePerformanceSummary = &summaryCopy
	} else {
		defCopy.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{}
		m.logger.Debugf("GetWorkerDefinition: Definition (Name: '%s', ID: '%s') has nil AggregatePerformanceSummary, initialized empty for copy.", def.Name, definitionID)
	}

	if tracker, ok := m.rateTrackers[def.DefinitionID]; ok && defCopy.AggregatePerformanceSummary != nil {
		defCopy.AggregatePerformanceSummary.ActiveInstancesCount = tracker.CurrentActiveInstances
	} else if defCopy.AggregatePerformanceSummary != nil {
		defCopy.AggregatePerformanceSummary.ActiveInstancesCount = 0
		m.logger.Debugf("GetWorkerDefinition: No rate tracker found for definition (Name: '%s', ID: '%s') when fetching active instance count, or summary was nil.", def.Name, definitionID)
	}
	return &defCopy, nil
}

// ListWorkerDefinitions returns a list of AI worker definitions, optionally filtered.
func (m *AIWorkerManager) ListWorkerDefinitions(filters map[string]interface{}) []*AIWorkerDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*AIWorkerDefinition, 0, len(m.definitions))
	for _, def := range m.definitions {
		if m.matchesDefinitionFilters(def, filters) {
			defCopy := *def
			if def.AggregatePerformanceSummary != nil {
				summaryCopy := *def.AggregatePerformanceSummary
				defCopy.AggregatePerformanceSummary = &summaryCopy
			} else {
				defCopy.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{}
				m.logger.Debugf("ListWorkerDefinitions: Definition (Name: '%s', ID: '%s') has nil AggregatePerformanceSummary, initialized empty for copy.", def.Name, def.DefinitionID)
			}

			if tracker, ok := m.rateTrackers[def.DefinitionID]; ok && defCopy.AggregatePerformanceSummary != nil {
				defCopy.AggregatePerformanceSummary.ActiveInstancesCount = tracker.CurrentActiveInstances
			} else if defCopy.AggregatePerformanceSummary != nil {
				defCopy.AggregatePerformanceSummary.ActiveInstancesCount = 0
				m.logger.Debugf("ListWorkerDefinitions: No rate tracker for definition (Name: '%s', ID: '%s'), or summary was nil. Active count set to 0.", def.Name, def.DefinitionID)
			}
			list = append(list, &defCopy)
		}
	}
	return list
}

// UpdateWorkerDefinition updates an existing AI worker definition and persists changes.
func (m *AIWorkerManager) UpdateWorkerDefinition(definitionID string, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	def, exists := m.definitions[definitionID]
	if !exists {
		if nameVal, nameInUpdates := updates["name"]; nameInUpdates {
			if nameStr, nameIsStr := nameVal.(string); nameIsStr {
				for _, d := range m.definitions {
					if d.Name == nameStr {
						m.logger.Warnf("UpdateWorkerDefinition: Original ID '%s' not found, but found definition by name '%s' (ID: '%s'). Proceeding with update on found definition.", definitionID, nameStr, d.DefinitionID)
						def = d
						definitionID = d.DefinitionID
						exists = true
						break
					}
				}
			}
		}
		if !exists {
			return NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition ID '%s' not found for update", definitionID), ErrNotFound)
		}
	}

	changedFields := []string{}
	originalName := def.Name

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
			for id, existingDef := range m.definitions {
				if id != definitionID && existingDef.Name == v {
					m.logger.Errorf("UpdateWorkerDefinition: Attempt to rename definition (ID: '%s', Original Name: '%s') to '%s', but this name is already used by definition (ID: '%s').", definitionID, originalName, v, id)
					return NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("worker definition name '%s' already exists for ID '%s'", v, id), ErrInvalidArgument)
				}
			}
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
			newAuth := def.Auth
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
			if len(def.CostMetrics) != len(newMetrics) && !changed {
				for k_1 := range def.CostMetrics {
					if _, existsInNew := newMetrics[k_1]; !existsInNew {
						changed = true
						break
					}
				}
			}
			if changed || !reflect.DeepEqual(def.CostMetrics, newMetrics) {
				def.CostMetrics = newMetrics
				setChanged("CostMetrics")
			}
		}
	}
	if val, ok := updates["rate_limits"]; ok {
		if vMap, ok := val.(map[string]interface{}); ok {
			newLimits := def.RateLimits
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
				m.initializeRateTrackerForDefinitionUnsafe(def)
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
	if val, ok := updates["data_source_refs"]; ok {
		if vSlice, ok := val.([]interface{}); ok {
			newRefs := []string{}
			for _, item := range vSlice {
				if s, sOk := item.(string); sOk {
					newRefs = append(newRefs, s)
				}
			}
			if !reflect.DeepEqual(def.DataSourceRefs, newRefs) {
				def.DataSourceRefs = newRefs
				setChanged("DataSourceRefs")
			}
		}
	}
	if val, ok := updates["tool_allowlist"]; ok {
		if vSlice, ok := val.([]interface{}); ok {
			newAllow := []string{}
			for _, item := range vSlice {
				if s, sOk := item.(string); sOk {
					newAllow = append(newAllow, s)
				}
			}
			if !reflect.DeepEqual(def.ToolAllowlist, newAllow) {
				def.ToolAllowlist = newAllow
				setChanged("ToolAllowlist")
			}
		}
	}
	if val, ok := updates["tool_denylist"]; ok {
		if vSlice, ok := val.([]interface{}); ok {
			newDeny := []string{}
			for _, item := range vSlice {
				if s, sOk := item.(string); sOk {
					newDeny = append(newDeny, s)
				}
			}
			if !reflect.DeepEqual(def.ToolDenylist, newDeny) {
				def.ToolDenylist = newDeny
				setChanged("ToolDenylist")
			}
		}
	}
	if val, ok := updates["default_supervisory_ai_ref"]; ok {
		if v, ok := val.(string); ok && def.DefaultSupervisoryAIRef != v {
			def.DefaultSupervisoryAIRef = v
			setChanged("DefaultSupervisoryAIRef")
		}
	}

	if len(changedFields) > 0 {
		def.ModifiedTimestamp = time.Now()
		setChanged("ModifiedTimestamp")
		m.logger.Infof("AIWorkerManager: Updating AIWorkerDefinition (Name: '%s', ID: '%s'). Changed fields: %v", def.Name, definitionID, changedFields)
		err := m.persistDefinitionsUnsafe()
		if err != nil {
			if _, ok := err.(*RuntimeError); ok {
				return err
			}
			return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to save updated definition (Name: '%s', ID: %s)", def.Name, definitionID), err)
		}
		return nil
	}
	m.logger.Infof("AIWorkerManager: UpdateWorkerDefinition (Name: '%s', ID: '%s'): No effective changes.", originalName, definitionID)
	return nil
}

// RemoveWorkerDefinition removes a definition if no active instances are using it.
func (m *AIWorkerManager) RemoveWorkerDefinition(definitionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	def, exists := m.definitions[definitionID]
	if !exists {
		return NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition ID '%s' not found for removal", definitionID), ErrNotFound)
	}
	defName := def.Name

	if tracker, ok := m.rateTrackers[definitionID]; ok && tracker.CurrentActiveInstances > 0 {
		m.logger.Warnf("RemoveWorkerDefinition: Cannot remove definition (Name: '%s', ID: '%s'), it has %d active instances.", defName, definitionID, tracker.CurrentActiveInstances)
		return NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("cannot remove definition (Name: '%s', ID: '%s'), it has %d active instances. Retire instances first.", defName, definitionID, tracker.CurrentActiveInstances), ErrFailedPrecondition)
	}

	delete(m.definitions, definitionID)
	delete(m.rateTrackers, definitionID)
	m.logger.Infof("AIWorkerManager: Removed AIWorkerDefinition: Name='%s', ID=%s", defName, definitionID)
	err := m.persistDefinitionsUnsafe()
	if err != nil {
		if _, ok := err.(*RuntimeError); ok {
			return err
		}
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to save after removing definition (Name: '%s', ID: %s)", defName, definitionID), err)
	}
	return nil
}

// matchesDefinitionFilters is a helper to check if a definition matches given criteria.
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
