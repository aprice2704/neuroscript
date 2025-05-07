// NeuroScript Version: 0.3.0
// File version: 0.1.1
// AI Worker Management: Instance Management Methods (Error Handling Corrected)
// filename: pkg/core/ai_wm_instances.go

package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	// "github.com/aprice2704/neuroscript/pkg/logging"
)

// SpawnWorkerInstance creates a new AIWorkerInstance from a given definition.
// It handles checking definition status, supported interaction models, and concurrent instance limits.
// The caller is responsible for associating a ConversationManager with the returned instance if needed.
func (m *AIWorkerManager) SpawnWorkerInstance(
	definitionID string,
	instanceConfigOverrides map[string]interface{},
	initialFileContexts []string, // Optional: if nil or empty, definition defaults are used
) (*AIWorkerInstance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	def, exists := m.definitions[definitionID]
	if !exists {
		return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition ID '%s' not found for spawning instance", definitionID), ErrNotFound)
	}

	if def.Status != DefinitionStatusActive {
		return nil, NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("worker definition '%s' is not active (status: %s), cannot spawn instance", definitionID, def.Status), ErrFailedPrecondition)
	}

	// Check if definition supports conversational interaction model
	supportsConversational := false
	for _, model := range def.InteractionModels {
		if model == InteractionModelConversational || model == InteractionModelBoth {
			supportsConversational = true
			break
		}
	}
	if !supportsConversational {
		return nil, NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("worker definition '%s' does not support conversational interaction, cannot spawn instance", definitionID), ErrFailedPrecondition)
	}

	// Rate Limiting: Check MaxConcurrentActiveInstances
	tracker := m.getOrCreateRateTrackerUnsafe(def) // Ensures tracker exists
	if def.RateLimits.MaxConcurrentActiveInstances > 0 && tracker.CurrentActiveInstances >= def.RateLimits.MaxConcurrentActiveInstances {
		m.logger.Warnf("SpawnWorkerInstance: Max concurrent instances (%d) reached for definition '%s'", def.RateLimits.MaxConcurrentActiveInstances, definitionID)
		return nil, NewRuntimeError(ErrorCodeRateLimited, fmt.Sprintf("max concurrent instances (%d) reached for definition '%s'", def.RateLimits.MaxConcurrentActiveInstances, definitionID), ErrRateLimited)
	}

	instanceID := uuid.NewString()
	effectiveConfig := make(map[string]interface{}) // Combine base and overrides
	for k, v := range def.BaseConfig {
		effectiveConfig[k] = v
	}
	for k, v := range instanceConfigOverrides { // Instance-specific overrides take precedence
		effectiveConfig[k] = v
	}

	effectiveFileContexts := def.DefaultFileContexts // Start with definition defaults
	if len(initialFileContexts) > 0 {                // If spawn-time contexts are provided, they override
		effectiveFileContexts = initialFileContexts
	}

	now := time.Now()
	instance := &AIWorkerInstance{
		InstanceID:            instanceID,
		DefinitionID:          definitionID,
		Status:                InstanceStatusIdle,           // Start as idle, ready for work
		ConversationHistory:   make([]*ConversationTurn, 0), // To be populated by ConversationManager interaction
		CreationTimestamp:     now,
		LastActivityTimestamp: now,
		SessionTokenUsage:     TokenUsageMetrics{},
		CurrentConfig:         effectiveConfig,
		ActiveFileContexts:    effectiveFileContexts, // Store the determined contexts
	}

	m.activeInstances[instanceID] = instance
	tracker.CurrentActiveInstances++
	def.AggregatePerformanceSummary.TotalInstancesSpawned++ // Update stat on definition

	m.logger.Infof("AIWorkerManager: Spawned AIWorkerInstance ID=%s from DefinitionID=%s. Active instances for def: %d", instanceID, definitionID, tracker.CurrentActiveInstances)
	return instance, nil
}

// GetWorkerInstance retrieves an active AI worker instance by its ID.
// Returns a pointer to the instance in the manager's map.
func (m *AIWorkerManager) GetWorkerInstance(instanceID string) (*AIWorkerInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.activeInstances[instanceID]
	if !exists {
		return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("active AI worker instance with ID '%s' not found", instanceID), ErrNotFound)
	}
	return instance, nil
}

// ListActiveWorkerInstances returns a list of currently active AI worker instances, optionally filtered.
// Returns copies of the instances to prevent modification of internal state through the list.
func (m *AIWorkerManager) ListActiveWorkerInstances(filters map[string]interface{}) []*AIWorkerInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*AIWorkerInstance, 0, len(m.activeInstances))
	for _, instance := range m.activeInstances {
		if m.matchesInstanceFilters(instance, filters) {
			instanceCopy := *instance
			list = append(list, &instanceCopy)
		}
	}
	return list
}

// RetireWorkerInstance moves an instance from the active list, updates its status,
// and persists its final metadata and performance records.
func (m *AIWorkerManager) RetireWorkerInstance(
	instanceID string,
	reason string,
	finalStatus AIWorkerInstanceStatus,
	finalSessionUsage TokenUsageMetrics, // Pass in the final token usage for this session
	instancePerformanceRecords []*PerformanceRecord, // Pass in all records for this instance
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.activeInstances[instanceID]
	if !exists {
		m.logger.Warnf("RetireWorkerInstance: Active instance ID '%s' not found for retirement.", instanceID)
		return NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("active AI worker instance ID '%s' not found for retirement", instanceID), ErrNotFound)
	}

	isRetirementStatus := false
	switch finalStatus {
	case InstanceStatusRetiredCompleted, InstanceStatusRetiredError, InstanceStatusRetiredExhausted, InstanceStatusContextFull, InstanceStatusTokenLimitReached:
		isRetirementStatus = true
	}
	if !isRetirementStatus {
		m.logger.Warnf("RetireWorkerInstance: Final status '%s' for instance '%s' is not a recognized retirement status. Defaulting to RetiredExhausted.", finalStatus, instanceID)
		finalStatus = InstanceStatusRetiredExhausted
	}

	instance.Status = finalStatus
	instance.RetirementReason = reason
	instance.LastActivityTimestamp = time.Now()
	instance.SessionTokenUsage = finalSessionUsage

	retiredInfo := RetiredInstanceInfo{
		InstanceID:          instance.InstanceID,
		DefinitionID:        instance.DefinitionID,
		CreationTimestamp:   instance.CreationTimestamp,
		RetirementTimestamp: instance.LastActivityTimestamp,
		FinalStatus:         instance.Status,
		RetirementReason:    instance.RetirementReason,
		SessionTokenUsage:   instance.SessionTokenUsage,
		InitialFileContexts: instance.ActiveFileContexts,
		PerformanceRecords:  instancePerformanceRecords,
	}

	if err := m.appendRetiredInstanceToFileUnsafe(retiredInfo); err != nil {
		m.logger.Errorf("AIWorkerManager: Failed to persist retired instance info for %s: %v. Instance will still be removed from active list.", instanceID, err)
		// If err is already a RuntimeError, return it, otherwise wrap.
		if _, ok := err.(*RuntimeError); ok {
			return err
		}
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to persist retired instance info for %s", instanceID), err)
	}
	m.logger.Infof("AIWorkerManager: Successfully persisted RetiredInstanceInfo for ID=%s", instanceID)

	delete(m.activeInstances, instanceID)

	if def, defExists := m.definitions[instance.DefinitionID]; defExists {
		tracker := m.getOrCreateRateTrackerUnsafe(def)
		if tracker.CurrentActiveInstances > 0 {
			tracker.CurrentActiveInstances--
		}
		m.logger.Debugf("AIWorkerManager: Decremented active instance count for DefinitionID=%s. Now: %d", instance.DefinitionID, tracker.CurrentActiveInstances)
	} else {
		m.logger.Warnf("AIWorkerManager: DefinitionID '%s' not found when retiring instance '%s'. Cannot update active instance count on definition summary.", instance.DefinitionID, instanceID)
	}

	m.logger.Infof("AIWorkerManager: Retired AIWorkerInstance: ID=%s, Reason: %s, Final Status: %s", instanceID, reason, finalStatus)
	return nil
}

// UpdateInstanceStatus updates the status and potentially LastError of an active instance.
func (m *AIWorkerManager) UpdateInstanceStatus(instanceID string, newStatus AIWorkerInstanceStatus, lastError string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.activeInstances[instanceID]
	if !exists {
		return NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("active AI worker instance ID '%s' not found for status update", instanceID), ErrNotFound)
	}

	if instance.Status == newStatus && (newStatus != InstanceStatusError || instance.LastError == lastError) {
		m.logger.Debugf("UpdateInstanceStatus: No change for instance %s, status %s.", instanceID, newStatus)
		return nil
	}

	instance.Status = newStatus
	instance.LastActivityTimestamp = time.Now()
	if newStatus == InstanceStatusError {
		instance.LastError = lastError
	} else {
		instance.LastError = ""
	}

	m.logger.Infof("AIWorkerManager: Updated status for InstanceID=%s to %s. Error: '%s'", instanceID, newStatus, instance.LastError)
	return nil
}

// matchesInstanceFilters is a helper to check if an active instance matches given criteria.
func (m *AIWorkerManager) matchesInstanceFilters(instance *AIWorkerInstance, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}

	for key, expectedValue := range filters {
		filterKey := strings.ToLower(key)
		match := false

		switch filterKey {
		case "instanceid", "id":
			if id, ok := expectedValue.(string); ok && instance.InstanceID == id {
				match = true
			}
		case "definitionid":
			if id, ok := expectedValue.(string); ok && instance.DefinitionID == id {
				match = true
			}
		case "status":
			if statusStr, ok := expectedValue.(string); ok && instance.Status == AIWorkerInstanceStatus(statusStr) {
				match = true
			}
		default:
			m.logger.Debugf("AIWorkerManager.matchesInstanceFilters: Unknown or unhandled filter key '%s'", filterKey)
		}

		if !match {
			return false
		}
	}
	return true
}

// UpdateInstanceSessionTokenUsage updates the token usage for an active instance.
func (m *AIWorkerManager) UpdateInstanceSessionTokenUsage(instanceID string, inputTokens int64, outputTokens int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.activeInstances[instanceID]
	if !exists {
		return NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("active AI worker instance ID '%s' not found for token usage update", instanceID), ErrNotFound)
	}

	instance.SessionTokenUsage.InputTokens += inputTokens
	instance.SessionTokenUsage.OutputTokens += outputTokens
	instance.SessionTokenUsage.TotalTokens = instance.SessionTokenUsage.InputTokens + instance.SessionTokenUsage.OutputTokens
	instance.LastActivityTimestamp = time.Now()

	m.logger.Debugf("Updated session token usage for InstanceID=%s: Input=%d, Output=%d, TotalSession=%d",
		instanceID, inputTokens, outputTokens, instance.SessionTokenUsage.TotalTokens)

	return nil
}
