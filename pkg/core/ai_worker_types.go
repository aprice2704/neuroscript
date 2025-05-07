// NeuroScript Version: 0.3.1
// File version: 0.2.4 // Uncommented LLMCallMetrics struct.
// filename: pkg/core/ai_worker_types.go

package core

import (
	"time"
	// "github.com/google/generative-ai-go/genai" // Or your internal ConversationManager path
)

// InteractionModelType specifies the primary intended use of an AIWorkerDefinition.
type InteractionModelType string

const (
	InteractionModelConversational InteractionModelType = "conversational" // Stateful, uses instances with conversation history
	InteractionModelStateless      InteractionModelType = "stateless_task" // One-shot tasks, no instance history needed beyond the call
	InteractionModelBoth           InteractionModelType = "both"           // Can be used for either
)

// APIKeySourceMethod defines how an API key is to be retrieved.
type APIKeySourceMethod string

const (
	APIKeyMethodEnvVar     APIKeySourceMethod = "env_var"     // Key is in an environment variable
	APIKeyMethodInline     APIKeySourceMethod = "inline"      // Key is provided directly in Value (use with caution)
	APIKeyMethodConfigPath APIKeySourceMethod = "config_path" // Future: Key is at a path in a secure config file
	APIKeyMethodVault      APIKeySourceMethod = "vault"       // Future: Key is in a secrets vault
	APIKeyMethodNone       APIKeySourceMethod = "none"        // For models that don't require an API key
)

// AIWorkerProvider represents the source or type of the AI worker.
type AIWorkerProvider string

const (
	ProviderGoogle    AIWorkerProvider = "google"
	ProviderOpenAI    AIWorkerProvider = "openai"
	ProviderAnthropic AIWorkerProvider = "anthropic"
	ProviderOllama    AIWorkerProvider = "ollama" // For local models via Ollama
	ProviderLocal     AIWorkerProvider = "local"  // For other direct local model integrations
	ProviderCustom    AIWorkerProvider = "custom" // For other, unspecified types
)

// AIWorkerDefinitionStatus indicates the general status of a definition.
type AIWorkerDefinitionStatus string

const (
	DefinitionStatusActive   AIWorkerDefinitionStatus = "active"   // Available for spawning instances
	DefinitionStatusDisabled AIWorkerDefinitionStatus = "disabled" // Not available for spawning
	DefinitionStatusArchived AIWorkerDefinitionStatus = "archived" // Kept for records but not active
)

// AIWorkerInstanceStatus indicates the current state of an AI worker instance.
type AIWorkerInstanceStatus string

const (
	InstanceStatusInitializing      AIWorkerInstanceStatus = "initializing"        // Being set up
	InstanceStatusIdle              AIWorkerInstanceStatus = "idle"                // Active, ready for a task
	InstanceStatusBusy              AIWorkerInstanceStatus = "busy"                // Active, currently processing a task
	InstanceStatusContextFull       AIWorkerInstanceStatus = "context_full"        // Active, but context window is (near) full, should be retired
	InstanceStatusRateLimited       AIWorkerInstanceStatus = "rate_limited"        // Temporarily paused due to rate limits
	InstanceStatusTokenLimitReached AIWorkerInstanceStatus = "token_limit_reached" // Reached a token usage cap (session or definition)
	InstanceStatusRetiredCompleted  AIWorkerInstanceStatus = "retired_completed"   // Gracefully retired after completing its work
	InstanceStatusRetiredExhausted  AIWorkerInstanceStatus = "retired_exhausted"   // Retired due to context, errors, or other exhaustion
	InstanceStatusRetiredError      AIWorkerInstanceStatus = "retired_error"       // Retired due to a persistent error state
	InstanceStatusError             AIWorkerInstanceStatus = "error"               // Unexpected error state
)

// APIKeySource specifies where to find the API key.
type APIKeySource struct {
	Method APIKeySourceMethod `json:"method"`
	Value  string             `json:"value,omitempty"` // e.g., "GOOGLE_API_KEY" or the actual key
}

// RateLimitPolicy defines usage limits for an AIWorkerDefinition.
type RateLimitPolicy struct {
	MaxRequestsPerMinute         int `json:"max_requests_per_minute,omitempty"`
	MaxTokensPerMinute           int `json:"max_tokens_per_minute,omitempty"` // Input + Output
	MaxTokensPerDay              int `json:"max_tokens_per_day,omitempty"`    // Input + Output
	MaxConcurrentActiveInstances int `json:"max_concurrent_active_instances,omitempty"`
}

// TokenUsageMetrics tracks token consumption for an AIWorkerInstance session or other contexts.
type TokenUsageMetrics struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	TotalTokens  int64 `json:"total_tokens"`
}

// SupervisorFeedback holds feedback from a supervisory AI on a task's quality.
type SupervisorFeedback struct {
	Rating                 float64   `json:"rating,omitempty"`
	Comments               string    `json:"comments,omitempty"`
	CorrectionInstructions string    `json:"correction_instructions,omitempty"`
	SupervisorAgentID      string    `json:"supervisor_agent_id,omitempty"`
	FeedbackTimestamp      time.Time `json:"feedback_timestamp,omitempty"`
}

// PerformanceRecord stores data about a single task performed by an AI worker.
type PerformanceRecord struct {
	TaskID             string                 `json:"task_id"`
	InstanceID         string                 `json:"instance_id"` // If "stateless-<uuid>", it was a direct call against a definition.
	DefinitionID       string                 `json:"definition_id"`
	TimestampStart     time.Time              `json:"timestamp_start"`
	TimestampEnd       time.Time              `json:"timestamp_end"`
	DurationMs         int64                  `json:"duration_ms"`
	Success            bool                   `json:"success"`
	InputContext       map[string]interface{} `json:"input_context,omitempty"`
	LLMMetrics         map[string]interface{} `json:"llm_metrics,omitempty"` // This map will store fields that were in LLMCallMetrics
	CostIncurred       float64                `json:"cost_incurred,omitempty"`
	OutputSummary      string                 `json:"output_summary,omitempty"`
	ErrorDetails       string                 `json:"error_details,omitempty"`
	SupervisorFeedback *SupervisorFeedback    `json:"supervisor_feedback,omitempty"`
}

// AIWorkerPerformanceSummary provides aggregated performance statistics.
type AIWorkerPerformanceSummary struct {
	TotalTasksAttempted   int       `json:"total_tasks_attempted"`
	SuccessfulTasks       int       `json:"successful_tasks"`
	FailedTasks           int       `json:"failed_tasks"`
	AverageSuccessRate    float64   `json:"average_success_rate"`
	AverageDurationMs     float64   `json:"average_duration_ms"`
	TotalTokensProcessed  int64     `json:"total_tokens_processed"`
	TotalCostIncurred     float64   `json:"total_cost_incurred"`
	AverageQualityScore   float64   `json:"average_quality_score,omitempty"`
	LastActivityTimestamp time.Time `json:"last_activity_timestamp,omitempty"`
	TotalInstancesSpawned int       `json:"total_instances_spawned,omitempty"` // Specific to Definition summary
	ActiveInstancesCount  int       `json:"active_instances_count,omitempty"`  // Runtime info on definition
}

// AIWorkerDefinition is the blueprint for AI worker instances or direct stateless calls.
type AIWorkerDefinition struct {
	DefinitionID                string                      `json:"definition_id"`
	Name                        string                      `json:"name,omitempty"`
	Provider                    AIWorkerProvider            `json:"provider"`
	ModelName                   string                      `json:"model_name"`
	Auth                        APIKeySource                `json:"auth"`
	InteractionModels           []InteractionModelType      `json:"interaction_models,omitempty"` // If empty, defaults to ["conversational"]
	Capabilities                []string                    `json:"capabilities,omitempty"`
	BaseConfig                  map[string]interface{}      `json:"base_config,omitempty"`
	CostMetrics                 map[string]float64          `json:"cost_metrics,omitempty"`
	RateLimits                  RateLimitPolicy             `json:"rate_limits,omitempty"`
	Status                      AIWorkerDefinitionStatus    `json:"status,omitempty"` // Default to "active"
	DefaultFileContexts         []string                    `json:"default_file_contexts,omitempty"`
	AggregatePerformanceSummary *AIWorkerPerformanceSummary `json:"aggregate_performance_summary,omitempty"`
	Metadata                    map[string]interface{}      `json:"metadata,omitempty"`
}

// AIWorkerInstance represents an active or retired conversational session with an LLM.
type AIWorkerInstance struct {
	InstanceID            string                 `json:"instance_id"`
	DefinitionID          string                 `json:"definition_id"`
	Status                AIWorkerInstanceStatus `json:"status"`
	ConversationHistory   []*ConversationTurn    `json:"-"` // Managed by ConversationManager, not directly persisted here
	CreationTimestamp     time.Time              `json:"creation_timestamp"`
	LastActivityTimestamp time.Time              `json:"last_activity_timestamp"`
	SessionTokenUsage     TokenUsageMetrics      `json:"session_token_usage"`
	CurrentConfig         map[string]interface{} `json:"current_config,omitempty"`
	ActiveFileContexts    []string               `json:"-"` // Runtime only, not persisted with instance
	LastError             string                 `json:"last_error,omitempty"`
	RetirementReason      string                 `json:"retirement_reason,omitempty"`
}

// RetiredInstanceInfo is used for persisting key metadata and performance of a retired instance.
type RetiredInstanceInfo struct {
	InstanceID          string                 `json:"instance_id"`
	DefinitionID        string                 `json:"definition_id"`
	CreationTimestamp   time.Time              `json:"creation_timestamp"`
	RetirementTimestamp time.Time              `json:"retirement_timestamp"`
	FinalStatus         AIWorkerInstanceStatus `json:"final_status"`
	RetirementReason    string                 `json:"retirement_reason,omitempty"`
	SessionTokenUsage   TokenUsageMetrics      `json:"session_token_usage"`
	InitialFileContexts []string               `json:"initial_file_contexts,omitempty"` // Files active at spawn time
	PerformanceRecords  []*PerformanceRecord   `json:"performance_records"`             // All records from this instance's session
}

// LLMCallMetrics holds detailed metrics from a specific LLM API call.
// This is useful as a temporary struct when processing responses before
// summarizing into PerformanceRecord.LLMMetrics.
type LLMCallMetrics struct {
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	TotalTokens  int64  `json:"total_tokens"`
	FinishReason string `json:"finish_reason,omitempty"`
	ModelUsed    string `json:"model_used,omitempty"` // The actual model identifier used for the call
}

// ConversationTurn is defined in llm_types.go
