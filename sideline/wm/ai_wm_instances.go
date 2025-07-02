package core

import (
	"fmt"
	"log" // Standard log for critical panics
	"strings"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/google/uuid"
	// "github.com/aprice2704/neuroscript/pkg/logging"
)

func (m *AIWorkerManager) SpawnWorkerInstance(
	definitionID string,
	instanceConfigOverrides map[string]interface{},
	initialFileContexts []string,
) (*AIWorkerInstance, error) {
	// Initial nil check for m and m.logger
	if m == nil {
		log.Fatalf("CRITICAL PANIC (SpawnWorkerInstance): AIWorkerManager 'm' receiver is nil.")
	}
	if m.logger == nil {
		log.Fatalf("CRITICAL PANIC (SpawnWorkerInstance): AIWorkerManager's logger (m.logger) is nil.")
	}
	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): ENTER SpawnWorkerInstance for DefID '%s', AIWM Addr: %p", definitionID, m)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): Mutex locked in SpawnWorkerInstance.")

	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): Accessing m.definitions (map Addr: %p, len: %d) for DefID '%s'", m.definitions, len(m.definitions), definitionID)
	def, exists := m.definitions[definitionID]
	if !exists {
		m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): Definition ID '%s' not found in m.definitions.", definitionID)
		return nil, lang.NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition ID '%s' not found for spawning instance", definitionID), ErrNotFound)
	}
	if def == nil {
		errMsg := fmt.Sprintf("CRITICAL PANIC (SpawnWorkerInstance): Definition '%s' found in map but pointer is nil.", definitionID)
		m.logger.Errorf(errMsg)
		panic(errMsg)
	}
	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): Fetched 'def' for DefID '%s'. Def Addr: %p, Def Name: '%s'", definitionID, def, def.Name)

	if def.Status != DefinitionStatusActive {
		m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): DefID '%s' is not active (Status: %s).", definitionID, def.Status)
		return nil, lang.NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("worker definition '%s' is not active (status: %s), cannot spawn instance", definitionID, def.Status), ErrFailedPrecondition)
	}

	supportsConversational := false
	if def.InteractionModels != nil {
		for _, model := range def.InteractionModels {
			if model == InteractionModelConversational || model == InteractionModelBoth {
				supportsConversational = true
				break
			}
		}
	} else {
		m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): DefID '%s' has nil InteractionModels slice.", definitionID)
	}

	if !supportsConversational {
		m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): DefID '%s' does not support conversational interaction.", definitionID)
		return nil, lang.NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("worker definition '%s' does not support conversational interaction, cannot spawn instance", definitionID), ErrFailedPrecondition)
	}

	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): ---- Immediately BEFORE CALL to m.getOrCreateRateTrackerUnsafe ----")
	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): Current AIWorkerManager (m) Addr: %p", m)
	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): Current m.logger Addr: %p", m.logger)
	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): Current m.rateTrackers map Addr: %p, Len: %d", m.rateTrackers, len(m.rateTrackers))
	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): Current AIWorkerDefinition (def) Addr: %p", def)

	tracker := m.getOrCreateRateTrackerUnsafe(def)

	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): ---- Immediately AFTER CALL to m.getOrCreateRateTrackerUnsafe ----")
	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): Returned tracker Addr: %p", tracker)

	if tracker == nil {
		errMsg := fmt.Sprintf("CRITICAL PANIC (SpawnWorkerInstance): getOrCreateRateTrackerUnsafe returned nil tracker for DefID '%s'. This should have been caught by internal panics in getOrCreateRateTrackerUnsafe.", def.DefinitionID)
		m.logger.Errorf(errMsg)
		panic(errMsg)
	}

	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): About to check concurrency. Tracker Addr: %p. Def Addr: %p", tracker, def)
	// The linter correctly identifies that `def != nil && tracker != nil` is always true here.
	if def.RateLimits.MaxConcurrentActiveInstances > 0 && tracker.CurrentActiveInstances >= def.RateLimits.MaxConcurrentActiveInstances {
		m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): Max concurrent instances (%d) reached for DefID '%s'. Current: %d", def.RateLimits.MaxConcurrentActiveInstances, definitionID, tracker.CurrentActiveInstances)
		m.logger.Warnf("SpawnWorkerInstance: Max concurrent instances (%d) reached for definition '%s'", def.RateLimits.MaxConcurrentActiveInstances, definitionID)
		return nil, lang.NewRuntimeError(ErrorCodeRateLimited, fmt.Sprintf("max concurrent instances (%d) reached for definition '%s'", def.RateLimits.MaxConcurrentActiveInstances, definitionID), ErrRateLimited)
	}

	instanceID := uuid.NewString()
	effectiveConfig := make(map[string]interface{})
	// The linter correctly identifies `def != nil` is always true here.
	if def.BaseConfig != nil {
		for k, v := range def.BaseConfig {
			effectiveConfig[k] = v
		}
	}
	for k, v := range instanceConfigOverrides {
		effectiveConfig[k] = v
	}

	var effectiveFileContexts []string
	// The linter correctly identifies `def != nil` is always true here.
	effectiveFileContexts = def.DefaultFileContexts

	if len(initialFileContexts) > 0 {
		effectiveFileContexts = initialFileContexts
	}

	// *** START OF THE PRIMARY FIX ***
	// Ensure the manager's LLM client is available and assign it.
	instanceLLMClient := m.llmClient // Get the manager's LLM client from AIWorkerManager struct
	if instanceLLMClient == nil {
		// This indicates a problem with AIWorkerManager initialization.
		errMsg := fmt.Sprintf("CRITICAL: AIWorkerManager's default LLM client (m.llmClient) is nil. Cannot spawn instance %s for DefID %s. This is likely NeuroScript Error 19.", instanceID, definitionID)
		m.logger.Errorf(errMsg)
		// Use a generic error code if ErrorCodeLLMClientNotSet is not defined,
		// or use the numeric code if that's how your errors are structured.
		// For now, using ErrorCodeInternal as a placeholder for "Error 19".
		return nil, lang.NewRuntimeError(ErrorCodeInternal, errMsg, nil) // Adjusted Error Code
	}
	// *** END OF THE PRIMARY FIX (Part 1: Getting and checking manager's client) ***

	now := time.Now()
	instance := &AIWorkerInstance{
		InstanceID:            instanceID,
		DefinitionID:          definitionID,
		Status:                InstanceStatusIdle,
		ConversationHistory:   make([]*interfaces.ConversationTurn, 0),
		CreationTimestamp:     now,
		LastActivityTimestamp: now,
		SessionTokenUsage:     TokenUsageMetrics{},
		CurrentConfig:         effectiveConfig,
		ActiveFileContexts:    effectiveFileContexts,
		// *** START OF THE PRIMARY FIX (Part 2: Assigning to instance fields) ***
		llmClient: instanceLLMClient, // Assign the LLM client to the instance
		// Logger field removed as it's not in AIWorkerInstance struct
		// *** END OF THE PRIMARY FIX (Part 2) ***
	}

	m.activeInstances[instanceID] = instance
	// Adjusted log message as instance.Logger is not a field
	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): Instance %s created. LLMClient assigned: %T", instanceID, instance.llmClient)

	// tracker.CurrentActiveInstances++ // STUBBED OUT by user in uploaded file
	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): CurrentActiveInstances increment commented out for DefID '%s'.", definitionID)

	// AggregatePerformanceSummary logic commented out by user in uploaded file
	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): AggregatePerformanceSummary update skipped/commented due to prior compiler error.")

	// The linter correctly identifies that tracker is never nil here.
	m.logger.Infof("AIWorkerManager: Spawned AIWorkerInstance ID=%s from DefinitionID=%s. Active instances for def (if tracked by stub): %d", instanceID, definitionID, tracker.CurrentActiveInstances)

	m.logger.Debugf("DEBUG_INSTANCE (v0.1.10+fix2): EXIT SpawnWorkerInstance for DefID '%s', returning instance Addr: %p", definitionID, instance)
	return instance, nil
}

// GetWorkerInstance retrieves an active AI worker instance by its ID.
func (m *AIWorkerManager) GetWorkerInstance(instanceID string) (*AIWorkerInstance, error) {
	if m == nil {
		log.Fatalf("CRITICAL PANIC (GetWorkerInstance): AIWorkerManager 'm' receiver is nil.")
	}
	if m.logger == nil {
		log.Fatalf("CRITICAL PANIC (GetWorkerInstance): AIWorkerManager's logger is nil.")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.activeInstances[instanceID]
	if !exists {
		return nil, lang.NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("active AI worker instance with ID '%s' not found", instanceID), ErrNotFound)
	}
	if instance == nil {
		errMsg := fmt.Sprintf("CRITICAL PANIC (GetWorkerInstance): Instance '%s' found in map but is nil.", instanceID)
		m.logger.Errorf(errMsg)
		panic(errMsg)
	}
	return instance, nil
}

// ListActiveWorkerInstances returns a list of currently active AI worker instances, optionally filtered.
func (m *AIWorkerManager) ListActiveWorkerInstances(filters map[string]interface{}) []*AIWorkerInstance {
	if m == nil {
		log.Fatalf("CRITICAL PANIC (ListActiveWorkerInstances): AIWorkerManager 'm' receiver is nil.")
	}
	if m.logger == nil {
		log.Fatalf("CRITICAL PANIC (ListActiveWorkerInstances): AIWorkerManager's logger is nil.")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*AIWorkerInstance, 0, len(m.activeInstances))
	for id, instance := range m.activeInstances {
		if instance == nil {
			fmt.Printf("WARN_INSTANCE (v0.1.10): ListActiveWorkerInstances: Found a nil instance in activeInstances map with ID '%s'. Skipping.\n", id)
			// Linter correctly identifies m.logger is not nil here
			m.logger.Warnf("ListActiveWorkerInstances: Found a nil instance in activeInstances map with ID '%s'. Skipping.", id)
			continue
		}
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
	finalSessionUsage TokenUsageMetrics,
	instancePerformanceRecords []*PerformanceRecord,
) error {
	if m == nil {
		log.Fatalf("CRITICAL PANIC (RetireWorkerInstance): AIWorkerManager 'm' receiver is nil.")
	}
	if m.logger == nil {
		log.Fatalf("CRITICAL PANIC (RetireWorkerInstance): AIWorkerManager's logger is nil.")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.activeInstances[instanceID]
	if !exists {
		m.logger.Warnf("RetireWorkerInstance: Active instance ID '%s' not found for retirement.", instanceID)
		return lang.NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("active AI worker instance ID '%s' not found for retirement", instanceID), ErrNotFound)
	}
	if instance == nil {
		errMsg := fmt.Sprintf("CRITICAL PANIC (RetireWorkerInstance): Instance '%s' found in map but is nil.", instanceID)
		m.logger.Errorf(errMsg)
		panic(errMsg)
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
		if _, ok := err.(*RuntimeError); ok {
			return err
		}
		return lang.NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to persist retired instance info for %s", instanceID), err)
	}

	delete(m.activeInstances, instanceID)

	// CurrentActiveInstances decrement is part of rate limiting, which is stubbed.
	// defToUpdate, defToUpdateExists := m.definitions[instance.DefinitionID]
	// if defToUpdateExists && defToUpdate != nil {
	// 	tracker := m.getOrCreateRateTrackerUnsafe(defToUpdate)
	// 	if tracker != nil {
	// 		// tracker.CurrentActiveInstances-- // STUBBED OUT
	// 	}
	// }
	fmt.Printf("DEBUG_INSTANCE (v0.1.10): CurrentActiveInstances decrement commented out for DefID '%s' in RetireWorkerInstance.\n", instance.DefinitionID)

	m.logger.Infof("AIWorkerManager: Retired AIWorkerInstance: ID=%s, Reason: %s, Final Status: %s", instanceID, reason, finalStatus)
	return nil
}

// UpdateInstanceStatus updates the status and potentially LastError of an active instance.
func (m *AIWorkerManager) UpdateInstanceStatus(instanceID string, newStatus AIWorkerInstanceStatus, lastError string) error {
	if m == nil {
		log.Fatalf("CRITICAL PANIC (UpdateInstanceStatus): AIWorkerManager 'm' receiver is nil.")
	}
	if m.logger == nil {
		log.Fatalf("CRITICAL PANIC (UpdateInstanceStatus): AIWorkerManager's logger is nil.")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.activeInstances[instanceID]
	if !exists {
		return lang.NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("active AI worker instance ID '%s' not found for status update", instanceID), ErrNotFound)
	}
	if instance == nil {
		errMsg := fmt.Sprintf("CRITICAL PANIC (UpdateInstanceStatus): Instance '%s' found in map but is nil.", instanceID)
		m.logger.Errorf(errMsg)
		panic(errMsg)
	}

	if instance.Status == newStatus && (newStatus != InstanceStatusError || instance.LastError == lastError) {
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
	if instance == nil { // Should be checked by caller ListActiveWorkerInstances
		// Linter correctly identifies m.logger is not nil here
		m.logger.Warnf("matchesInstanceFilters called with nil instance.")
		return false
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
			// Linter correctly identifies m.logger is not nil here
			m.logger.Debugf("AIWorkerManager.matchesInstanceFilters: Unknown filter key '%s'", filterKey)
		}

		if !match {
			return false
		}
	}
	return true
}

// UpdateInstanceSessionTokenUsage updates the token usage for an active instance.
func (m *AIWorkerManager) UpdateInstanceSessionTokenUsage(instanceID string, inputTokens int64, outputTokens int64) error {
	if m == nil {
		log.Fatalf("CRITICAL PANIC (UpdateInstanceSessionTokenUsage): AIWorkerManager 'm' receiver is nil.")
	}
	if m.logger == nil {
		log.Fatalf("CRITICAL PANIC (UpdateInstanceSessionTokenUsage): AIWorkerManager's logger is nil.")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.activeInstances[instanceID]
	if !exists {
		return lang.NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("active AI worker instance ID '%s' not found for token usage update", instanceID), ErrNotFound)
	}
	if instance == nil {
		errMsg := fmt.Sprintf("CRITICAL PANIC (UpdateInstanceSessionTokenUsage): Instance '%s' found in map but is nil.", instanceID)
		m.logger.Errorf(errMsg)
		panic(errMsg)
	}

	instance.SessionTokenUsage.InputTokens += inputTokens
	instance.SessionTokenUsage.OutputTokens += outputTokens
	instance.SessionTokenUsage.TotalTokens = instance.SessionTokenUsage.InputTokens + instance.SessionTokenUsage.OutputTokens
	instance.LastActivityTimestamp = time.Now()

	m.logger.Debugf("Updated session token usage for InstanceID=%s: Input=%d, Output=%d, TotalSession=%d (Rate limit counter updates are stubbed)",
		instanceID, inputTokens, outputTokens, instance.SessionTokenUsage.TotalTokens)

	return nil
}
