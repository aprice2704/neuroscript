// NeuroScript Version: 0.3.0
// File version: 0.1.6
// AI Worker Management: Stateless Task Execution (Relies on LLMCallMetrics in types)
// filename: pkg/core/ai_wm_stateless.go

package core

import (
	"context" // Added for llmClient.Ask
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai" // For genai.Content and genai.Text
	"github.com/google/uuid"
	// "github.com/aprice2704/neuroscript/pkg/logging"
)

// Helper function to convert []*genai.Content to []*ConversationTurn
// This is a simplified conversion focusing on text content.
func convertGenaiContentsToConversationTurns(genaiContents []*genai.Content) []*ConversationTurn {
	if genaiContents == nil {
		return nil
	}
	turns := make([]*ConversationTurn, 0, len(genaiContents))
	for _, gc := range genaiContents {
		if gc == nil {
			continue
		}
		var contentBuilder strings.Builder
		for _, part := range gc.Parts {
			if textPart, ok := part.(genai.Text); ok {
				contentBuilder.WriteString(string(textPart))
			}
		}
		turns = append(turns, &ConversationTurn{
			Role:    Role(gc.Role),
			Content: contentBuilder.String(),
		})
	}
	return turns
}

// ExecuteStatelessTask allows making a direct call using an AIWorkerDefinition
// without creating or managing a full AIWorkerInstance. This is suitable for one-shot tasks.
// It still respects the definition's rate limits and logs performance.
// The llmClient is passed in, typically from the Interpreter.
func (m *AIWorkerManager) ExecuteStatelessTask(
	definitionID string,
	llmClient LLMClient, // LLMClient is an interface, expected to be provided
	prompt string,
	configOverrides map[string]interface{},
) (string /* modelOutput */, *PerformanceRecord, error) {

	m.mu.Lock()

	def, exists := m.definitions[definitionID]
	if !exists {
		m.mu.Unlock()
		return "", nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition ID '%s' not found for stateless task", definitionID), ErrNotFound)
	}

	if def.Status != DefinitionStatusActive {
		m.mu.Unlock()
		return "", nil, NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("worker definition '%s' is not active (status: %s), cannot execute stateless task", definitionID, def.Status), ErrFailedPrecondition)
	}

	supportsStateless := false
	for _, modelType := range def.InteractionModels {
		if modelType == InteractionModelStateless || modelType == InteractionModelBoth {
			supportsStateless = true
			break
		}
	}
	if !supportsStateless {
		m.mu.Unlock()
		return "", nil, NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("worker definition '%s' does not support stateless interaction", definitionID), ErrFailedPrecondition)
	}

	tracker := m.getOrCreateRateTrackerUnsafe(def)
	now := time.Now()
	var rateLimitErr *RuntimeError

	if def.RateLimits.MaxRequestsPerMinute > 0 {
		if now.Sub(tracker.RequestsMinuteMarker).Minutes() >= 1 {
			tracker.RequestsThisMinuteCount = 0
			tracker.RequestsMinuteMarker = now
		}
		if tracker.RequestsThisMinuteCount >= def.RateLimits.MaxRequestsPerMinute {
			rateLimitErr = NewRuntimeError(ErrorCodeRateLimited, fmt.Sprintf("requests_per_minute limit (%d) reached for definition '%s'", def.RateLimits.MaxRequestsPerMinute, definitionID), ErrRateLimited)
		}
	}

	if rateLimitErr != nil {
		m.mu.Unlock()
		m.logger.Warnf("ExecuteStatelessTask: Rate limit hit for DefID %s: %v", definitionID, rateLimitErr)
		return "", nil, rateLimitErr
	}
	m.mu.Unlock()

	if _, keyErr := m.resolveAPIKey(def.Auth); keyErr != nil {
		if re, ok := keyErr.(*RuntimeError); ok {
			return "", nil, re
		}
		return "", nil, NewRuntimeError(ErrorCodeConfiguration, fmt.Sprintf("failed to resolve API key for definition '%s' (stateless task)", def.DefinitionID), keyErr)
	}

	effectiveConfig := make(map[string]interface{})
	for k, v := range def.BaseConfig {
		effectiveConfig[k] = v
	}
	for k, v := range configOverrides {
		effectiveConfig[k] = v
	}

	startTime := time.Now()
	var responseContent string
	var llmCallMetrics LLMCallMetrics // Now this type should be defined
	var callErr error

	if llmClient == nil {
		m.logger.Warnf("AIWorkerManager: No LLMClient provided for stateless task on definition %s. Using mock response.", definitionID)
		responseContent = "Mocked LLM response for stateless prompt: " + prompt
		llmCallMetrics = LLMCallMetrics{
			InputTokens:  int64(len(prompt) / 4),
			OutputTokens: int64(len(responseContent) / 4),
			FinishReason: "stop",
			ModelUsed:    def.ModelName,
		}
		llmCallMetrics.TotalTokens = llmCallMetrics.InputTokens + llmCallMetrics.OutputTokens
	} else {
		tempConvoMgr := NewConversationManager(m.logger)
		tempConvoMgr.AddUserMessage(prompt)
		turnsForLLM := convertGenaiContentsToConversationTurns(tempConvoMgr.GetHistory())

		// The LLMClient's Ask method is expected to return populated LLMCallMetrics on success.
		// This part depends on the actual signature and return values of your LLMClient.Ask method.
		// For this example, I'm assuming it returns a ConversationTurn and an error,
		// and we derive LLMCallMetrics from that or other sources if available.
		llmResponseTurn, askErr := llmClient.Ask(context.Background(), turnsForLLM)

		if askErr != nil {
			if _, ok := askErr.(*RuntimeError); ok {
				callErr = askErr
			} else {
				callErr = NewRuntimeError(ErrorCodeLLMError, fmt.Sprintf("LLM Ask failed for stateless task (DefID: %s)", definitionID), askErr)
			}
		} else if llmResponseTurn == nil || llmResponseTurn.Content == "" {
			callErr = NewRuntimeError(ErrorCodeLLMError, fmt.Sprintf("LLM response was empty or malformed for stateless task (DefID: %s)", definitionID), ErrLLMError)
		} else {
			responseContent = llmResponseTurn.Content
			// Ideally, llmClient.Ask or llmResponseTurn would provide detailed metrics.
			// If not, we use estimates.
			llmCallMetrics = LLMCallMetrics{
				InputTokens:  int64(len(prompt) / 4),
				OutputTokens: int64(len(responseContent) / 4),
				FinishReason: "stop",
				ModelUsed:    def.ModelName,
			}
			// If llmResponseTurn.Metrics is available and is of type LLMCallMetrics:
			// if metrics, ok := llmResponseTurn.Metrics.(LLMCallMetrics); ok {
			//    llmCallMetrics = metrics
			// } else { ... handle missing metrics ... }
			llmCallMetrics.TotalTokens = llmCallMetrics.InputTokens + llmCallMetrics.OutputTokens
		}
	}

	endTime := time.Now()
	durationMs := endTime.Sub(startTime).Milliseconds()

	cost := 0.0
	if def.CostMetrics != nil {
		if perTokenCostIn, okIn := def.CostMetrics["input_cost_per_token"]; okIn {
			cost += float64(llmCallMetrics.InputTokens) * perTokenCostIn
		}
		if perTokenCostOut, okOut := def.CostMetrics["output_cost_per_token"]; okOut {
			cost += float64(llmCallMetrics.OutputTokens) * perTokenCostOut
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	tracker = m.getOrCreateRateTrackerUnsafe(def)
	tokensUsedForRateLimit := llmCallMetrics.TotalTokens
	if callErr != nil && tokensUsedForRateLimit == 0 {
		tokensUsedForRateLimit = int64(len(prompt) / 4)
	}

	if now.Sub(tracker.RequestsMinuteMarker).Minutes() >= 1 {
		tracker.RequestsThisMinuteCount = 0
		tracker.RequestsMinuteMarker = now
	}
	tracker.RequestsThisMinuteCount++

	m.updateTokenCountForRateLimitsUnsafe(tracker, tokensUsedForRateLimit)

	// Populate the LLMMetrics map for the PerformanceRecord
	llmMetricsForRecord := map[string]interface{}{
		"input_tokens":       llmCallMetrics.InputTokens,
		"output_tokens":      llmCallMetrics.OutputTokens,
		"total_tokens":       llmCallMetrics.TotalTokens,
		"finish_reason":      llmCallMetrics.FinishReason,
		"model_used_at_call": llmCallMetrics.ModelUsed,
	}

	perfRecord := &PerformanceRecord{
		TaskID:         uuid.NewString(),
		InstanceID:     statelessInstanceIDPrefix + uuid.NewString(),
		DefinitionID:   definitionID,
		TimestampStart: startTime,
		TimestampEnd:   endTime,
		DurationMs:     durationMs,
		Success:        callErr == nil,
		InputContext: map[string]interface{}{
			"prompt_char_length": len(prompt),
		},
		LLMMetrics:    llmMetricsForRecord, // Use the map here
		CostIncurred:  cost,
		OutputSummary: smartTrim(responseContent, 256),
		ErrorDetails:  ifErrorToString(callErr),
	}

	if logErr := m.logPerformanceRecordUnsafe(perfRecord); logErr != nil {
		m.logger.Errorf("AIWorkerManager: Failed to log performance for stateless task (DefID: %s, TaskID: %s): %v", definitionID, perfRecord.TaskID, logErr)
	} else {
		if errSave := m.persistDefinitionsUnsafe(); errSave != nil {
			m.logger.Errorf("AIWorkerManager: Failed to save definitions after logging stateless performance (DefID: %s): %v", definitionID, errSave)
		}
	}

	if callErr != nil {
		if _, ok := callErr.(*RuntimeError); !ok && callErr != nil {
			callErr = NewRuntimeError(ErrorCodeLLMError, "stateless LLM task failed", callErr)
		}
		return "", perfRecord, callErr
	}

	return responseContent, perfRecord, nil
}
