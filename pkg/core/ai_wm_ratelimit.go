// NeuroScript Version: 0.3.0
// File version: 0.1.1
// AI Worker Management: Rate Limiting Logic (Error Handling Corrected)
// filename: pkg/core/ai_wm_ratelimit.go

package core

import (
	"fmt"
	"time"
	// "github.com/aprice2704/neuroscript/pkg/logging"
)

// WorkerRateTracker holds runtime info for rate limiting an AIWorkerDefinition.
// It's managed internally by AIWorkerManager.
type WorkerRateTracker struct {
	DefinitionID string // For context, though map key in AIWorkerManager is DefinitionID

	// For MaxRequestsPerMinute
	RequestsThisMinuteCount int
	RequestsMinuteMarker    time.Time

	// For MaxTokensPerMinute
	TokensThisMinuteCount int64
	TokensMinuteMarker    time.Time

	// For MaxTokensPerDay
	TokensThisDayCount int64
	TokensDayMarker    time.Time

	// For MaxConcurrentActiveInstances
	CurrentActiveInstances int
}

// initializeRateTrackerForDefinitionUnsafe creates a new rate tracker for a definition if one doesn't exist.
// Called with Write Lock.
func (m *AIWorkerManager) initializeRateTrackerForDefinitionUnsafe(def *AIWorkerDefinition) {
	if _, exists := m.rateTrackers[def.DefinitionID]; !exists {
		now := time.Now()
		m.rateTrackers[def.DefinitionID] = &WorkerRateTracker{
			DefinitionID:           def.DefinitionID, // Store for context if needed, though key is def.DefinitionID
			RequestsMinuteMarker:   now,
			TokensMinuteMarker:     now,
			TokensDayMarker:        now,
			CurrentActiveInstances: 0,
		}
		m.logger.Debugf("AIWorkerManager: Initialized new rate tracker for DefinitionID: %s", def.DefinitionID)
	}
}

// getOrCreateRateTrackerUnsafe retrieves or creates a rate tracker for a definition.
// Called with Write Lock or when it's known the definition exists and tracker should be present.
func (m *AIWorkerManager) getOrCreateRateTrackerUnsafe(def *AIWorkerDefinition) *WorkerRateTracker {
	tracker, exists := m.rateTrackers[def.DefinitionID]
	if !exists {
		m.logger.Warnf("AIWorkerManager: Rate tracker for DefinitionID '%s' not found, creating ad-hoc. This might indicate an initialization gap.", def.DefinitionID)
		m.initializeRateTrackerForDefinitionUnsafe(def)
		tracker = m.rateTrackers[def.DefinitionID] // Re-fetch after creation
	}
	return tracker
}

// checkAndRecordUsageUnsafe checks if a call (request with estimated tokens) can proceed
// based on the definition's rate limits.
// This function is primarily a CHECK function. Recording is handled by recordUsageUnsafe.
// Called with Write Lock.
func (m *AIWorkerManager) checkAndRecordUsageUnsafe(
	def *AIWorkerDefinition,
	tracker *WorkerRateTracker,
	isInstanceSpawn bool,
	tokensForCall int64,
) error {
	now := time.Now()
	policy := def.RateLimits

	// 1. Check MaxConcurrentActiveInstances (only if spawning)
	if isInstanceSpawn {
		if policy.MaxConcurrentActiveInstances > 0 && tracker.CurrentActiveInstances >= policy.MaxConcurrentActiveInstances {
			m.logger.Warnf("RateLimitCheck: Max concurrent instances (%d) reached for DefID %s.", policy.MaxConcurrentActiveInstances, def.DefinitionID)
			return NewRuntimeError(ErrorCodeRateLimited, fmt.Sprintf("max concurrent instances (%d) reached for definition '%s'", policy.MaxConcurrentActiveInstances, def.DefinitionID), ErrRateLimited)
		}
	}

	// --- Per-Minute Limits ---
	if now.Sub(tracker.RequestsMinuteMarker).Minutes() >= 1 {
		tracker.RequestsThisMinuteCount = 0
		tracker.RequestsMinuteMarker = now
	}
	if now.Sub(tracker.TokensMinuteMarker).Minutes() >= 1 {
		tracker.TokensThisMinuteCount = 0
		tracker.TokensMinuteMarker = now
	}

	if !isInstanceSpawn && policy.MaxRequestsPerMinute > 0 {
		if tracker.RequestsThisMinuteCount+1 > policy.MaxRequestsPerMinute {
			m.logger.Warnf("RateLimitCheck: MaxRequestsPerMinute (%d) would be exceeded for DefID %s.", policy.MaxRequestsPerMinute, def.DefinitionID)
			return NewRuntimeError(ErrorCodeRateLimited, fmt.Sprintf("requests_per_minute limit (%d) reached for definition '%s'", policy.MaxRequestsPerMinute, def.DefinitionID), ErrRateLimited)
		}
	}

	if !isInstanceSpawn && policy.MaxTokensPerMinute > 0 {
		if tracker.TokensThisMinuteCount+tokensForCall > int64(policy.MaxTokensPerMinute) {
			m.logger.Warnf("RateLimitCheck: MaxTokensPerMinute (%d) would be exceeded for DefID %s (current: %d, adding: %d).", policy.MaxTokensPerMinute, def.DefinitionID, tracker.TokensThisMinuteCount, tokensForCall)
			return NewRuntimeError(ErrorCodeRateLimited, fmt.Sprintf("tokens_per_minute limit (%d) would be exceeded for definition '%s'", policy.MaxTokensPerMinute, def.DefinitionID), ErrRateLimited)
		}
	}

	// --- Per-Day Limits ---
	if now.Sub(tracker.TokensDayMarker).Hours() >= 24 {
		tracker.TokensThisDayCount = 0
		tracker.TokensDayMarker = now
	}

	if !isInstanceSpawn && policy.MaxTokensPerDay > 0 {
		if tracker.TokensThisDayCount+tokensForCall > int64(policy.MaxTokensPerDay) {
			m.logger.Warnf("RateLimitCheck: MaxTokensPerDay (%d) would be exceeded for DefID %s (current: %d, adding: %d).", policy.MaxTokensPerDay, def.DefinitionID, tracker.TokensThisDayCount, tokensForCall)
			return NewRuntimeError(ErrorCodeRateLimited, fmt.Sprintf("tokens_per_day limit (%d) would be exceeded for definition '%s'", policy.MaxTokensPerDay, def.DefinitionID), ErrRateLimited)
		}
	}

	m.logger.Debugf("RateLimitCheck: PASSED for DefID %s. IsSpawn: %t, TokensForCall: %d", def.DefinitionID, isInstanceSpawn, tokensForCall)
	return nil
}

// recordUsageUnsafe updates the rate limit counters after a call has been made.
// Called with Write Lock.
func (m *AIWorkerManager) recordUsageUnsafe(tracker *WorkerRateTracker, tokensUsed int64, isNewRequest bool) {
	now := time.Now()

	if now.Sub(tracker.RequestsMinuteMarker).Minutes() >= 1 {
		tracker.RequestsThisMinuteCount = 0
		tracker.RequestsMinuteMarker = now
	}
	if now.Sub(tracker.TokensMinuteMarker).Minutes() >= 1 {
		tracker.TokensThisMinuteCount = 0
		tracker.TokensMinuteMarker = now
	}
	if now.Sub(tracker.TokensDayMarker).Hours() >= 24 {
		tracker.TokensThisDayCount = 0
		tracker.TokensDayMarker = now
	}

	if isNewRequest {
		tracker.RequestsThisMinuteCount++
	}
	tracker.TokensThisMinuteCount += tokensUsed
	tracker.TokensThisDayCount += tokensUsed

	m.logger.Debugf("RateLimitRecord: Usage recorded for DefID %s. NewRequest: %t, Tokens: %d. MinReq: %d, MinTok: %d, DayTok: %d",
		tracker.DefinitionID, isNewRequest, tokensUsed,
		tracker.RequestsThisMinuteCount, tracker.TokensThisMinuteCount, tracker.TokensThisDayCount)
}

// updateTokenCountForRateLimitsUnsafe is a specific helper called by ExecuteStatelessTask or instance task completion.
// Called with Write Lock.
func (m *AIWorkerManager) updateTokenCountForRateLimitsUnsafe(tracker *WorkerRateTracker, tokensUsed int64) {
	now := time.Now()

	if now.Sub(tracker.TokensMinuteMarker).Minutes() >= 1 {
		m.logger.Debugf("RateLimitTokenUpdate: Resetting TokensThisMinuteCount for DefID %s (was %d).", tracker.DefinitionID, tracker.TokensThisMinuteCount)
		tracker.TokensThisMinuteCount = 0
		tracker.TokensMinuteMarker = now
	}
	if now.Sub(tracker.TokensDayMarker).Hours() >= 24 {
		m.logger.Debugf("RateLimitTokenUpdate: Resetting TokensThisDayCount for DefID %s (was %d).", tracker.DefinitionID, tracker.TokensThisDayCount)
		tracker.TokensThisDayCount = 0
		tracker.TokensDayMarker = now
	}

	tracker.TokensThisMinuteCount += tokensUsed
	tracker.TokensThisDayCount += tokensUsed

	m.logger.Debugf("RateLimitTokenUpdate: Tokens updated for DefID %s. Added: %d. MinTok: %d, DayTok: %d",
		tracker.DefinitionID, tokensUsed, tracker.TokensThisMinuteCount, tracker.TokensThisDayCount)
}
