// NeuroScript Version: 0.3.0
// File version: 0.1.0
// AI Worker Management: Tool Helper Functions
// filename: pkg/wm/ai_wm_tools_helpers.go

package wm

import (
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/sideline/nspatch"
)

// getAIWorkerManager retrieves the AIWorkerManager from the interpreter.
// It returns an error if the manager is not initialized.
func getAIWorkerManager(i tool.RunTime) (*AIWorkerManager, error) {
	if i.aiWorkerManager == nil {
		i.Logger().Error("AIWorkerManager not initialized in Interpreter context!")
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "AIWorkerManager not available in Interpreter", nspatch.ErrInternal)
	}
	return i.aiWorkerManager, nil
}

// mapValidatedArgsListToMapByName converts the list of validated arguments
// (from ValidateAndConvertArgs) into a map where keys are argument names,
// for easier access within tool functions.
func mapValidatedArgsListToMapByName(specArgs []tool.ArgSpec, validatedArgsList []interface{}) map[string]interface{} {
	argsMap := make(map[string]interface{})
	for idx, argSpec := range specArgs {
		if idx < len(validatedArgsList) {
			argsMap[argSpec.Name] = validatedArgsList[idx]
		} else {
			// If arg is not in validatedArgsList, it means it was optional and not provided,
			// or required and ValidateAndConvertArgs would have errored.
			// Setting to nil is a safe default for optional args.
			argsMap[argSpec.Name] = nil
		}
	}
	return argsMap
}

// convertAIWorkerDefinitionToMap converts an AIWorkerDefinition to a map for tool output.
func convertAIWorkerDefinitionToMap(def *AIWorkerDefinition) map[string]interface{} {
	if def == nil {
		return nil
	}
	interactionModelsStr := make([]interface{}, len(def.InteractionModels))
	for i, im := range def.InteractionModels {
		interactionModelsStr[i] = string(im)
	}
	capabilitiesInterface := make([]interface{}, len(def.Capabilities))
	for i, capability := range def.Capabilities {
		capabilitiesInterface[i] = capability
	}
	defaultFileContextsInterface := make([]interface{}, len(def.DefaultFileContexts))
	for i, contextVal := range def.DefaultFileContexts { // Renamed 'context' to 'contextVal' to avoid conflict
		defaultFileContextsInterface[i] = contextVal
	}
	costMetricsInterface := make(map[string]interface{})
	for k, v := range def.CostMetrics {
		costMetricsInterface[k] = v
	}

	var perfSummaryMap map[string]interface{}
	if def.AggregatePerformanceSummary != nil {
		// Pass the pointer directly, as convertPerformanceSummaryToMap expects *AIWorkerPerformanceSummary
		perfSummaryMap = convertPerformanceSummaryToMap(def.AggregatePerformanceSummary)
	}

	return map[string]interface{}{
		"definition_id":                 def.DefinitionID,
		"name":                          def.Name,
		"provider":                      string(def.Provider),
		"model_name":                    def.ModelName,
		"auth":                          map[string]interface{}{"method": string(def.Auth.Method), "value": def.Auth.Value},
		"interaction_models":            interactionModelsStr,
		"capabilities":                  capabilitiesInterface,
		"base_config":                   def.BaseConfig,
		"cost_metrics":                  costMetricsInterface,
		"rate_limits":                   convertRateLimitPolicyToMap(&def.RateLimits), // RateLimits is a struct
		"status":                        string(def.Status),
		"default_file_contexts":         defaultFileContextsInterface,
		"aggregate_performance_summary": perfSummaryMap,
		"metadata":                      def.Metadata,
	}
}

// convertAIWorkerInstanceToMap converts an AIWorkerInstance to a map for tool output.
func convertAIWorkerInstanceToMap(instance *AIWorkerInstance) map[string]interface{} {
	if instance == nil {
		return nil
	}
	activeFileContextsInterface := make([]interface{}, len(instance.ActiveFileContexts))
	for i, contextVal := range instance.ActiveFileContexts { // Renamed 'context' to 'contextVal'
		activeFileContextsInterface[i] = contextVal
	}
	return map[string]interface{}{
		"instance_id":             instance.InstanceID,
		"definition_id":           instance.DefinitionID,
		"status":                  string(instance.Status),
		"creation_timestamp":      instance.CreationTimestamp.Format(time.RFC3339Nano),
		"last_activity_timestamp": instance.LastActivityTimestamp.Format(time.RFC3339Nano),
		"session_token_usage":     convertTokenUsageMetricsToMap(&instance.SessionTokenUsage), // SessionTokenUsage is a struct
		"current_config":          instance.CurrentConfig,
		"active_file_contexts":    activeFileContextsInterface, // Note: This is runtime only, not persisted with instance struct
		"last_error":              instance.LastError,
		"retirement_reason":       instance.RetirementReason,
	}
}

// convertPerformanceRecordToMap converts a PerformanceRecord to a map for tool output.
func convertPerformanceRecordToMap(pr *PerformanceRecord) map[string]interface{} {
	if pr == nil {
		return nil
	}
	var supervisorFeedbackMap map[string]interface{}
	if pr.SupervisorFeedback != nil {
		supervisorFeedbackMap = map[string]interface{}{
			"rating":                  pr.SupervisorFeedback.Rating,
			"comments":                pr.SupervisorFeedback.Comments,
			"correction_instructions": pr.SupervisorFeedback.CorrectionInstructions,
			"supervisor_agent_id":     pr.SupervisorFeedback.SupervisorAgentID,
			"feedback_timestamp":      pr.SupervisorFeedback.FeedbackTimestamp.Format(time.RFC3339Nano),
		}
	}
	return map[string]interface{}{
		"task_id":             pr.TaskID,
		"instance_id":         pr.InstanceID,
		"definition_id":       pr.DefinitionID,
		"timestamp_start":     pr.TimestampStart.Format(time.RFC3339Nano),
		"timestamp_end":       pr.TimestampEnd.Format(time.RFC3339Nano),
		"duration_ms":         pr.DurationMs,
		"success":             pr.Success,
		"input_context":       pr.InputContext,
		"llm_metrics":         pr.LLMMetrics, // This is already a map[string]interface{}
		"cost_incurred":       pr.CostIncurred,
		"output_summary":      pr.OutputSummary,
		"error_details":       pr.ErrorDetails,
		"supervisor_feedback": supervisorFeedbackMap,
	}
}

// convertPerformanceSummaryToMap converts an AIWorkerPerformanceSummary to a map for tool output.
func convertPerformanceSummaryToMap(s *AIWorkerPerformanceSummary) map[string]interface{} {
	if s == nil {
		return nil
	}
	return map[string]interface{}{
		"total_tasks_attempted":   s.TotalTasksAttempted,
		"successful_tasks":        s.SuccessfulTasks,
		"failed_tasks":            s.FailedTasks,
		"average_success_rate":    s.AverageSuccessRate,
		"average_duration_ms":     s.AverageDurationMs,
		"total_tokens_processed":  s.TotalTokensProcessed,
		"total_cost_incurred":     s.TotalCostIncurred,
		"average_quality_score":   s.AverageQualityScore,
		"last_activity_timestamp": s.LastActivityTimestamp.Format(time.RFC3339Nano),
		"total_instances_spawned": s.TotalInstancesSpawned,
		"active_instances_count":  s.ActiveInstancesCount, // This is runtime info
	}
}

// convertRateLimitPolicyToMap converts a RateLimitPolicy to a map for tool output.
func convertRateLimitPolicyToMap(p *RateLimitPolicy) map[string]interface{} {
	if p == nil {
		return nil
	}
	return map[string]interface{}{
		"max_requests_per_minute":         p.MaxRequestsPerMinute,
		"max_tokens_per_minute":           p.MaxTokensPerMinute,
		"max_tokens_per_day":              p.MaxTokensPerDay,
		"max_concurrent_active_instances": p.MaxConcurrentActiveInstances,
	}
}

// convertTokenUsageMetricsToMap converts TokenUsageMetrics to a map for tool output.
func convertTokenUsageMetricsToMap(p *TokenUsageMetrics) map[string]interface{} {
	if p == nil {
		return nil
	}
	return map[string]interface{}{
		"input_tokens":  p.InputTokens,
		"output_tokens": p.OutputTokens,
		"total_tokens":  p.TotalTokens,
	}
}
