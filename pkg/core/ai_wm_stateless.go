// NeuroScript Version: 0.3.0
// File version: 0.1.10 // Minimal Lock Refactor on v0.1.9: RLock for read, copy key def data, removed persistDefinitionsUnsafe
// AI Worker Management: Stateless Task Execution (Relies on LLMCallMetrics in types)
// filename: pkg/core/ai_wm_stateless.go

package core

import (
	"context"
	"errors" //
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai" //
	"github.com/google/uuid"                   //
	// "github.com/aprice2704/neuroscript/pkg/logging"
)

// Helper function to convert []*genai.Content to []*ConversationTurn
// (Copied from your provided v0.1.9)
func convertGenaiContentsToConversationTurns(genaiContents []*genai.Content) []*ConversationTurn { //
	if genaiContents == nil { //
		return nil
	}
	turns := make([]*ConversationTurn, 0, len(genaiContents)) //
	for _, gc := range genaiContents {                        //
		if gc == nil { //
			continue
		}
		var contentBuilder strings.Builder //
		for _, part := range gc.Parts {    //
			if textPart, ok := part.(genai.Text); ok { //
				contentBuilder.WriteString(string(textPart)) //
			}
		}
		turns = append(turns, &ConversationTurn{ //
			Role:    Role(gc.Role),           //
			Content: contentBuilder.String(), //
		})
	}
	return turns //
}

// ExecuteStatelessTask allows making a direct call using an AIWorkerDefinition
// without creating or managing a full AIWorkerInstance. This is suitable for one-shot tasks.
// It still respects the definition's rate limits and logs performance.
// The llmClient is passed in, typically from the Interpreter.
func (m *AIWorkerManager) ExecuteStatelessTask( //
	definitionNameIn string, // Changed from 'name' to avoid conflict if 'name' is used locally
	llmClient LLMClient, //
	prompt string, //
	configOverrides map[string]interface{}, //
) (string /* modelOutput */, *PerformanceRecord, error) { //

	m.logger.Infof("ExecuteStatelessTask: Entered function for Definition Name: '%s'", definitionNameIn)

	// --- MINIMAL CHANGE: Define local vars to hold copies of def data ---
	var originalDefinitionID string
	var defAuthCopy APIKeySource // Assuming AIWorkerDefinition.Auth is APIKeySource
	var defBaseConfigCopy map[string]interface{}
	var defModelNameCopy string
	var defCostMetricsCopy map[string]float64
	//	var defStatusCopy DefinitionStatus
	var defInteractionModelsCopy []InteractionModelType
	// RateLimits are used by getOrCreateRateTrackerUnsafe, which is called with the *live* def later.

	// --- Phase 1: Read definition details under RLock ---
	m.mu.RLock() // CHANGED to RLock for reading
	m.logger.Infof("ExecuteStatelessTask: RLock acquired for Definition Name: '%s'", definitionNameIn)

	// Using direct map iteration as in your ai_wm.go v0.2.12 since GetDefinitionIDByName might not exist
	// or its locking behavior is unconfirmed for this context.
	var internalDefPtr *AIWorkerDefinition
	foundByName := false
	for id, dPtr := range m.definitions {
		if dPtr != nil && dPtr.Name == definitionNameIn {
			internalDefPtr = dPtr
			originalDefinitionID = id // Store the ID of the found definition
			foundByName = true
			break
		}
	}

	if !foundByName {
		m.logger.Errorf("ExecuteStatelessTask: Definition name '%s' not found in m.definitions.", definitionNameIn)
		m.mu.RUnlock()
		m.logger.Infof("ExecuteStatelessTask: RUnlock after definition name '%s' not found.", definitionNameIn)
		return "", nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker name '%s' not found for stateless task", definitionNameIn), ErrNotFound)
	}
	m.logger.Infof("ExecuteStatelessTask: Found definition for ID: '%s' (Name: '%s')", originalDefinitionID, definitionNameIn)

	// CRITICAL: Copy necessary data from 'internalDefPtr' before releasing RLock
	//defStatusCopy = internalDefPtr.Status
	defInteractionModelsCopy = append([]InteractionModelType(nil), internalDefPtr.InteractionModels...) // Deep copy slice
	defAuthCopy = internalDefPtr.Auth                                                                   // APIKeySource is a struct, direct copy
	defBaseConfigCopy = make(map[string]interface{}, len(internalDefPtr.BaseConfig))
	for k, v := range internalDefPtr.BaseConfig { // Deep copy map
		defBaseConfigCopy[k] = v
	}
	defModelNameCopy = internalDefPtr.ModelName
	defCostMetricsCopy = make(map[string]float64, len(internalDefPtr.CostMetrics))
	for k, v := range internalDefPtr.CostMetrics { // Deep copy map
		defCostMetricsCopy[k] = v
	}
	// Note: The call to getOrCreateRateTrackerUnsafe is removed from this RLock section.
	// The 'var rateLimitErr *RuntimeError' from original v0.1.9 is also removed as its check was commented out.

	m.mu.RUnlock() // Release RLock
	m.logger.Infof("ExecuteStatelessTask: RUnlock released for Definition ID: '%s'. Copied data will be used.", originalDefinitionID)

	// // --- Perform checks on copied data (no m.mu lock held) ---
	// if defStatusCopy != DefinitionStatusActive { // Using copied status
	// 	m.logger.Warnf("ExecuteStatelessTask: Definition '%s' is not active (status: %s)", definitionNameIn, defStatusCopy)
	// 	return "", nil, NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("worker definition '%s' is not active (status: %s), cannot execute stateless task", definitionNameIn, defStatusCopy), ErrFailedPrecondition)
	// }

	supportsStateless := false
	for _, modelType := range defInteractionModelsCopy { // Using copied interaction models
		if modelType == InteractionModelStateless || modelType == InteractionModelBoth {
			supportsStateless = true
			break
		}
	}
	if !supportsStateless {
		m.logger.Warnf("ExecuteStatelessTask: Definition '%s' does not support stateless interaction", definitionNameIn) // Name is from input, reliable
		return "", nil, NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("worker definition '%s' does not support stateless interaction", definitionNameIn), ErrFailedPrecondition)
	}

	m.logger.Infof("ExecuteStatelessTask: About to resolve API key for Definition ID: '%s'", originalDefinitionID)

	resolvedAPIKey, keyErr := m.resolveAPIKey(defAuthCopy) // Using copied Auth
	if keyErr != nil {
		m.logger.Errorf("ExecuteStatelessTask: API key resolution failed for DefID '%s': %v", originalDefinitionID, keyErr)
		if re, ok := keyErr.(*RuntimeError); ok {
			return "", nil, re
		}
		return "", nil, NewRuntimeError(ErrorCodeConfiguration, fmt.Sprintf("failed to resolve API key for definition '%s' (stateless task)", originalDefinitionID), keyErr)
	}
	m.logger.Infof("ExecuteStatelessTask: API key resolution successful for DefID '%s'. Key presence: %t", originalDefinitionID, resolvedAPIKey != "")

	effectiveConfig := make(map[string]interface{})
	for k, v := range defBaseConfigCopy { // Using copied BaseConfig
		effectiveConfig[k] = v
	}
	for k, v := range configOverrides {
		effectiveConfig[k] = v
	}

	taskStartTime := time.Now() // Moved to before the LLM call
	var responseContent string
	var llmCallMetrics LLMCallMetrics
	var callErr error

	if llmClient == nil { //
		m.logger.Warnf("AIWorkerManager: No LLMClient provided for stateless task on definition %s. Using mock response.", originalDefinitionID) //
		responseContent = "Mocked LLM response for stateless prompt: " + prompt
		llmCallMetrics = LLMCallMetrics{
			InputTokens:  int64(len(prompt) / 4),
			OutputTokens: int64(len(responseContent) / 4),
			FinishReason: "stop",
			ModelUsed:    defModelNameCopy, // Using copied ModelName
		}
		llmCallMetrics.TotalTokens = llmCallMetrics.InputTokens + llmCallMetrics.OutputTokens
	} else {
		tempConvoMgr := NewConversationManager(m.logger)
		tempConvoMgr.AddUserMessage(prompt)
		turnsForLLM := convertGenaiContentsToConversationTurns(tempConvoMgr.GetHistory())

		m.logger.Infof("ExecuteStatelessTask: Preparing to call llmClient.Ask for DefID %s with a 60s timeout. Prompt char length: %d", originalDefinitionID, len(prompt))

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var llmResponseTurn *ConversationTurn
		llmResponseTurn, callErr = llmClient.Ask(ctx, turnsForLLM)

		m.logger.Infof("ExecuteStatelessTask: llmClient.Ask call completed for DefID %s. Error returned: %v", originalDefinitionID, callErr)

		if callErr != nil {
			if errors.Is(callErr, context.DeadlineExceeded) {
				m.logger.Errorf("ExecuteStatelessTask: llmClient.Ask timed out for DefID %s: %v", originalDefinitionID, callErr)
				callErr = NewRuntimeError(ErrorCodeTimeout, fmt.Sprintf("LLM Ask timed out for stateless task (DefID: %s)", originalDefinitionID), callErr) // Ensure ErrorCodeTimeout exists
			} else {
				m.logger.Errorf("ExecuteStatelessTask: llmClient.Ask reported an error for DefID %s: %v", originalDefinitionID, callErr)
				if _, ok := callErr.(*RuntimeError); !ok {
					callErr = NewRuntimeError(ErrorCodeLLMError, fmt.Sprintf("LLM Ask failed for stateless task (DefID: %s)", originalDefinitionID), callErr)
				}
			}
		} else if llmResponseTurn == nil {
			m.logger.Errorf("ExecuteStatelessTask: llmClient.Ask returned a nil llmResponseTurn for DefID %s", originalDefinitionID)
			callErr = NewRuntimeError(ErrorCodeLLMError, fmt.Sprintf("LLM response turn was nil for stateless task (DefID: %s)", originalDefinitionID), ErrLLMError)
		} else if llmResponseTurn.Content == "" {
			responseContent = "" // Allow empty content if no error
			m.logger.Warnf("ExecuteStatelessTask: llmClient.Ask returned empty content for DefID %s", originalDefinitionID)
		} else {
			responseContent = llmResponseTurn.Content
			m.logger.Infof("ExecuteStatelessTask: llmClient.Ask successful for DefID %s. Response char length: %d", originalDefinitionID, len(responseContent))
		}

		// Populate llmCallMetrics based on llmResponseTurn
		if llmResponseTurn != nil {
			if llmResponseTurn.TokenUsage.TotalTokens > 0 || llmResponseTurn.TokenUsage.InputTokens > 0 || llmResponseTurn.TokenUsage.OutputTokens > 0 {
				llmCallMetrics.InputTokens = llmResponseTurn.TokenUsage.InputTokens
				llmCallMetrics.OutputTokens = llmResponseTurn.TokenUsage.OutputTokens
				llmCallMetrics.TotalTokens = llmResponseTurn.TokenUsage.TotalTokens
				llmCallMetrics.ModelUsed = defModelNameCopy // Using copied ModelName
			} else { // Fallback to estimates if not available from llmResponseTurn
				llmCallMetrics = LLMCallMetrics{
					InputTokens:  int64(len(prompt) / 4),
					OutputTokens: int64(len(responseContent) / 4),
					FinishReason: "stop",           // Or from llmResponseTurn if available
					ModelUsed:    defModelNameCopy, // Using copied ModelName
				}
				llmCallMetrics.TotalTokens = llmCallMetrics.InputTokens + llmCallMetrics.OutputTokens
			}
		} else if callErr != nil { // If callErr happened and llmResponseTurn is nil
			llmCallMetrics.InputTokens = int64(len(prompt) / 4) // Estimate input tokens
			llmCallMetrics.ModelUsed = defModelNameCopy
		}
	}

	taskEndTime := time.Now()
	durationMs := taskEndTime.Sub(taskStartTime).Milliseconds()

	cost := 0.0
	if defCostMetricsCopy != nil { // Using copied CostMetrics
		if perTokenCostIn, okIn := defCostMetricsCopy["input_cost_per_token"]; okIn {
			cost += float64(llmCallMetrics.InputTokens) * perTokenCostIn
		}
		if perTokenCostOut, okOut := defCostMetricsCopy["output_cost_per_token"]; okOut {
			cost += float64(llmCallMetrics.OutputTokens) * perTokenCostOut
		}
	}

	perfRecord := &PerformanceRecord{ //
		TaskID:         uuid.NewString(),                                          //
		InstanceID:     statelessInstanceIDPrefix + uuid.NewString(),              //
		DefinitionID:   originalDefinitionID,                                      // Use the ID found
		TimestampStart: taskStartTime,                                             //
		TimestampEnd:   taskEndTime,                                               //
		DurationMs:     durationMs,                                                //
		Success:        callErr == nil,                                            //
		InputContext:   map[string]interface{}{"prompt_char_length": len(prompt)}, //
		LLMMetrics: map[string]interface{}{ //
			"input_tokens":       llmCallMetrics.InputTokens,  //
			"output_tokens":      llmCallMetrics.OutputTokens, //
			"total_tokens":       llmCallMetrics.TotalTokens,  //
			"finish_reason":      llmCallMetrics.FinishReason, //
			"model_used_at_call": llmCallMetrics.ModelUsed,    //
		},
		CostIncurred:  cost,                            //
		OutputSummary: smartTrim(responseContent, 256), //
		ErrorDetails:  ifErrorToString(callErr),        //
	}

	// --- Phase 2: Update in-memory summaries and rate limits under Write Lock ---
	m.mu.Lock() // Acquire WRITE lock
	defer m.mu.Unlock()
	m.logger.Debugf("ExecuteStatelessTask: Write Lock acquired for rate limits and summary, DefID: %s", originalDefinitionID)

	// Re-fetch the live definition to update its summary and use its current RateLimits config
	liveDef, defStillExists := m.definitions[originalDefinitionID]
	if defStillExists {
		// logPerformanceRecordUnsafe updates liveDef.AggregatePerformanceSummary
		// This method (from your ai_wm_performance.go v0.1.3) does no I/O.
		summaryUpdateErr := m.logPerformanceRecordUnsafe(perfRecord) //
		if summaryUpdateErr != nil {
			// This error is from an in-memory update, so log it but don't overwrite primary callErr
			m.logger.Warnf("ExecuteStatelessTask: Failed to update in-memory performance summary for DefID %s, TaskID %s: %v", originalDefinitionID, perfRecord.TaskID, summaryUpdateErr)
		}

		// Update rate limits using the live definition
		currentRateTracker := m.getOrCreateRateTrackerUnsafe(liveDef) //
		tokensUsedForRateLimit := llmCallMetrics.TotalTokens          //
		if callErr != nil && tokensUsedForRateLimit == 0 {            //
			tokensUsedForRateLimit = llmCallMetrics.InputTokens // Account for input tokens if call failed
		}
		m.updateTokenCountForRateLimitsUnsafe(currentRateTracker, tokensUsedForRateLimit) //
	} else {
		m.logger.Warnf("ExecuteStatelessTask: Definition ID %s (Name: '%s') no longer exists after LLM call. Skipping rate limit and summary update.", originalDefinitionID, definitionNameIn)
	}

	// REMOVED: m.persistDefinitionsUnsafe() call as per user request and to avoid I/O in lock.
	m.logger.Debugf("ExecuteStatelessTask: Write Lock to be released for DefID: %s. (Definition persistence was skipped).", originalDefinitionID)

	// Note: The actual writing of `perfRecord` to a persistent file log is NOT done here.
	// That would require a separate mechanism outside this m.mu lock,
	// e.g., by calling a dedicated file-writing helper method on 'm'.
	// Since "forget performance logging etc." was mentioned, this is deferred.

	if callErr != nil { //
		if _, ok := callErr.(*RuntimeError); !ok { //
			callErr = NewRuntimeError(ErrorCodeLLMError, "stateless LLM task failed", callErr) // Using ErrorCodeLLMError for wrapping
		}
		return "", perfRecord, callErr //
	}

	return responseContent, perfRecord, nil //
}

// Ensure smartTrim and ifErrorToString are available in the 'core' package
// (e.g., from ai_wm.go or a utils.go file).
// type APIKeySource struct { ... } // Should be defined in ai_worker_types.go
// type DefinitionStatus string // Should be defined in ai_worker_types.go
// type RateLimitConfig struct { ... } // Should be defined in ai_worker_types.go
// const DefinitionStatusActive DefinitionStatus = "active" // Should be defined in ai_worker_types.go
// var ErrorCodeTimeout ErrorCode = ... // Should be defined in errors.go
