// NeuroScript Version: 0.3.0
// File version: (ai_wm_instances.go - v with extensive debug & AggPerfSummary commented)
// - Added extensive fmt.Println debugging around the call to getOrCreateRateTrackerUnsafe.
// - Stricter panic if tracker is nil after call.
// - Commented out tracker.CurrentActiveInstances++ and related decrement in RetireWorkerInstance.
// - Commented out AggregatePerformanceSummary to resolve immediate compiler error.
// filename: pkg/core/ai_wm_instances.go
package core

import (
	"fmt"
	"log" // Standard log for critical panics
	"strings"
	"time"

	"github.com/google/uuid"
	// "github.com/aprice2704/neuroscript/pkg/logging"
)

// SpawnWorkerInstance creates a new AIWorkerInstance from a given definition.
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
	fmt.Printf("DEBUG_INSTANCE (v0.1.10): ENTER SpawnWorkerInstance for DefID '%s', AIWM Addr: %p\n", definitionID, m)

	m.mu.Lock()
	defer m.mu.Unlock()
	fmt.Printf("DEBUG_INSTANCE (v0.1.10): Mutex locked in SpawnWorkerInstance.\n")

	fmt.Printf("DEBUG_INSTANCE (v0.1.10): Accessing m.definitions (map Addr: %p, len: %d) for DefID '%s'\n", m.definitions, len(m.definitions), definitionID)
	def, exists := m.definitions[definitionID]
	if !exists {
		fmt.Printf("DEBUG_INSTANCE (v0.1.10): Definition ID '%s' not found in m.definitions.\n", definitionID)
		return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition ID '%s' not found for spawning instance", definitionID), ErrNotFound)
	}
	if def == nil {
		errMsg := fmt.Sprintf("CRITICAL PANIC (SpawnWorkerInstance): Definition '%s' found in map but pointer is nil.", definitionID)
		m.logger.Errorf(errMsg) // m.logger is confirmed non-nil
		panic(errMsg)
	}
	fmt.Printf("DEBUG_INSTANCE (v0.1.10): Fetched 'def' for DefID '%s'. Def Addr: %p, Def Name: '%s'\n", definitionID, def, def.Name)

	if def.Status != DefinitionStatusActive {
		fmt.Printf("DEBUG_INSTANCE (v0.1.10): DefID '%s' is not active (Status: %s).\n", definitionID, def.Status)
		return nil, NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("worker definition '%s' is not active (status: %s), cannot spawn instance", definitionID, def.Status), ErrFailedPrecondition)
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
		fmt.Printf("DEBUG_INSTANCE (v0.1.10): DefID '%s' has nil InteractionModels slice.\n", definitionID)
	}

	if !supportsConversational {
		fmt.Printf("DEBUG_INSTANCE (v0.1.10): DefID '%s' does not support conversational interaction.\n", definitionID)
		return nil, NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("worker definition '%s' does not support conversational interaction, cannot spawn instance", definitionID), ErrFailedPrecondition)
	}
	// fmt.Printf("DEBUG_INSTANCE (v0.1.10): DefID '%s' supports conversational model. Def.RateLimits: %+v\n", definitionID, def.RateLimits)

	// --- Debugging block for line 85 area ---
	fmt.Printf("DEBUG_INSTANCE (v0.1.10): ---- Immediately BEFORE CALL to m.getOrCreateRateTrackerUnsafe (Original Line 85 area) ----\n")
	fmt.Printf("DEBUG_INSTANCE (v0.1.10): Current AIWorkerManager (m) Addr: %p\n", m)
	if m != nil { // Extra safety for accessing m's fields
		fmt.Printf("DEBUG_INSTANCE (v0.1.10): Current m.logger Addr: %p\n", m.logger)
		fmt.Printf("DEBUG_INSTANCE (v0.1.10): Current m.rateTrackers map Addr: %p, Len: %d\n", m.rateTrackers, len(m.rateTrackers))
	}
	fmt.Printf("DEBUG_INSTANCE (v0.1.10): Current AIWorkerDefinition (def) Addr: %p\n", def)
	if def != nil { // Extra safety
		fmt.Printf("DEBUG_INSTANCE (v0.1.10): Current def.DefinitionID: %s\n", def.DefinitionID)
	}
	// ------------------------------------

	tracker := m.getOrCreateRateTrackerUnsafe(def) // THIS IS EFFECTIVELY LINE 85 from original stack trace

	// --- Debugging block after line 85 ---
	fmt.Printf("DEBUG_INSTANCE (v0.1.10): ---- Immediately AFTER CALL to m.getOrCreateRateTrackerUnsafe ----\n")
	fmt.Printf("DEBUG_INSTANCE (v0.1.10): Returned tracker Addr: %p\n", tracker)
	// -------------------------------------

	if tracker == nil {
		errMsg := fmt.Sprintf("CRITICAL PANIC (SpawnWorkerInstance): getOrCreateRateTrackerUnsafe returned nil tracker for DefID '%s'. This should have been caught by internal panics in getOrCreateRateTrackerUnsafe.", def.DefinitionID)
		if m != nil && m.logger != nil { // m should be non-nil here
			m.logger.Errorf(errMsg)
		} else if m == nil {
			log.Printf("CRITICAL PANIC (SpawnWorkerInstance): m is nil, and getOrCreateRateTrackerUnsafe returned nil. DefID: %s", def.DefinitionID)
		} else { // m.logger is nil
			log.Printf("CRITICAL PANIC (SpawnWorkerInstance): m.logger is nil, and getOrCreateRateTrackerUnsafe returned nil. DefID: %s", def.DefinitionID)
		}
		panic(errMsg)
	}
	// fmt.Printf("DEBUG_INSTANCE (v0.1.10): Tracker for DefID '%s' obtained. Tracker Addr: %p, Tracker.DefID: '%s', Tracker.CurrentActiveInstances: %d\n", def.DefinitionID, tracker, tracker.DefinitionID, tracker.CurrentActiveInstances)

	// Rate Limiting: Check MaxConcurrentActiveInstances (using the STUBBED tracker)
	// This check remains to see if accessing def.RateLimits or tracker.CurrentActiveInstances (if non-nil) causes issues.
	// Rate limiting itself is stubbed in ai_wm_ratelimit.go v0.1.8.
	fmt.Printf("DEBUG_INSTANCE (v0.1.10): About to check concurrency. Tracker Addr: %p. Def Addr: %p\n", tracker, def)
	if def != nil && tracker != nil { // Ensure both are non-nil before accessing fields
		// fmt.Printf("DEBUG_INSTANCE (v0.1.10): Accessing tracker.CurrentActiveInstances (%d) and def.RateLimits.MaxConcurrentActiveInstances (%d)\n", tracker.CurrentActiveInstances, def.RateLimits.MaxConcurrentActiveInstances)
		if def.RateLimits.MaxConcurrentActiveInstances > 0 && tracker.CurrentActiveInstances >= def.RateLimits.MaxConcurrentActiveInstances {
			fmt.Printf("DEBUG_INSTANCE (v0.1.10): Max concurrent instances (%d) reached for DefID '%s'. Current: %d\n", def.RateLimits.MaxConcurrentActiveInstances, definitionID, tracker.CurrentActiveInstances)
			m.logger.Warnf("SpawnWorkerInstance: Max concurrent instances (%d) reached for definition '%s'", def.RateLimits.MaxConcurrentActiveInstances, definitionID)
			return nil, NewRuntimeError(ErrorCodeRateLimited, fmt.Sprintf("max concurrent instances (%d) reached for definition '%s'", def.RateLimits.MaxConcurrentActiveInstances, definitionID), ErrRateLimited)
		}
	} else {
		fmt.Printf("DEBUG_INSTANCE (v0.1.10): Skipped concurrency check because def (%p) or tracker (%p) is nil.\n", def, tracker)
	}
	// fmt.Printf("DEBUG_INSTANCE (v0.1.10): Concurrency check passed for DefID '%s'.\n", definitionID)

	instanceID := uuid.NewString()
	effectiveConfig := make(map[string]interface{})
	if def != nil && def.BaseConfig != nil {
		for k, v := range def.BaseConfig {
			effectiveConfig[k] = v
		}
	}
	for k, v := range instanceConfigOverrides {
		effectiveConfig[k] = v
	}

	var effectiveFileContexts []string
	if def != nil { // Check def before accessing DefaultFileContexts
		effectiveFileContexts = def.DefaultFileContexts
	}
	if len(initialFileContexts) > 0 {
		effectiveFileContexts = initialFileContexts
	}

	now := time.Now()
	instance := &AIWorkerInstance{
		InstanceID:            instanceID,
		DefinitionID:          definitionID,
		Status:                InstanceStatusIdle,
		ConversationHistory:   make([]*ConversationTurn, 0),
		CreationTimestamp:     now,
		LastActivityTimestamp: now,
		SessionTokenUsage:     TokenUsageMetrics{},
		CurrentConfig:         effectiveConfig,
		ActiveFileContexts:    effectiveFileContexts,
	}

	m.activeInstances[instanceID] = instance

	// tracker.CurrentActiveInstances++ // STUBBED OUT - Rate limit tracking is disabled in ai_wm_ratelimit.go v0.1.8
	// fmt.Printf("DEBUG_INSTANCE (v0.1.10): CurrentActiveInstances increment commented out for DefID '%s'.\n", definitionID)

	// Defensively ensure AggregatePerformanceSummary is not nil before incrementing
	// TODO: Resolve 'undefined: AggregatedPerformanceSummary' by ensuring type is correctly defined and accessible.
	// For now, commenting out to allow compilation for the main panic debug.
	/*
		if def != nil && def.AggregatePerformanceSummary == nil {
		    fmt.Printf("WARN_INSTANCE (v0.1.10): def.AggregatePerformanceSummary is nil for DefID '%s'. Initializing.\n", def.DefinitionID)
		    // def.AggregatePerformanceSummary = &AggregatedPerformanceSummary{} // This line caused the compiler error.
		}
		if def != nil && def.AggregatePerformanceSummary != nil {
		    def.AggregatePerformanceSummary.TotalInstancesSpawned++
		}
	*/
	fmt.Printf("DEBUG_INSTANCE (v0.1.10): AggregatePerformanceSummary update skipped/commented due to prior compiler error.\n")

	if m.logger != nil && tracker != nil { // Check both before logging with tracker fields
		m.logger.Infof("AIWorkerManager: Spawned AIWorkerInstance ID=%s from DefinitionID=%s. Active instances for def (if tracked by stub): %d", instanceID, definitionID, tracker.CurrentActiveInstances)
	} else if m.logger != nil {
		m.logger.Infof("AIWorkerManager: Spawned AIWorkerInstance ID=%s from DefinitionID=%s. Tracker is nil or logger issue for active instances.", instanceID, definitionID)
	}
	fmt.Printf("DEBUG_INSTANCE (v0.1.10): EXIT SpawnWorkerInstance for DefID '%s', returning instance Addr: %p\n", definitionID, instance)
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
		return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("active AI worker instance with ID '%s' not found", instanceID), ErrNotFound)
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
			if m.logger != nil {
				m.logger.Warnf("ListActiveWorkerInstances: Found a nil instance in activeInstances map with ID '%s'. Skipping.", id)
			}
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
		return NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("active AI worker instance ID '%s' not found for retirement", instanceID), ErrNotFound)
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
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to persist retired instance info for %s", instanceID), err)
	}

	delete(m.activeInstances, instanceID)

	// CurrentActiveInstances decrement is part of rate limiting, which is stubbed.
	// defToUpdate, defToUpdateExists := m.definitions[instance.DefinitionID]
	// if defToUpdateExists && defToUpdate != nil {
	// 	tracker := m.getOrCreateRateTrackerUnsafe(defToUpdate)
	// 	if tracker != nil {
	// 	    // tracker.CurrentActiveInstances-- // STUBBED OUT
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
		return NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("active AI worker instance ID '%s' not found for status update", instanceID), ErrNotFound)
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
		if m.logger != nil {
			m.logger.Warnf("matchesInstanceFilters called with nil instance.")
		}
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
			if m.logger != nil {
				m.logger.Debugf("AIWorkerManager.matchesInstanceFilters: Unknown filter key '%s'", filterKey)
			}
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
		return NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("active AI worker instance ID '%s' not found for token usage update", instanceID), ErrNotFound)
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
