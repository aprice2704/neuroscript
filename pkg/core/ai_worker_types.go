// NeuroScript Version: 0.4.0
// File version: 0.3.7
// Description: Defines types for the AI Worker Management system, including workers, data sources, pools, queues, and work items.
// filename: pkg/core/ai_worker_types.go

package core

import (
	"context"
	"fmt"
	"time"
)

// --- Enums and Basic Types ---

// InteractionModelType specifies the primary intended use of an AIWorkerDefinition.
type InteractionModelType string

const (
	InteractionModelConversational InteractionModelType = "conversational"
	InteractionModelStateless      InteractionModelType = "stateless_task"
	InteractionModelBoth           InteractionModelType = "both"
)

// APIKeySourceMethod defines how an API key is to be retrieved.
type APIKeySourceMethod string

const (
	APIKeyMethodEnvVar     APIKeySourceMethod = "env_var"
	APIKeyMethodInline     APIKeySourceMethod = "inline"      // Use with caution
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
	ProviderOllama    AIWorkerProvider = "ollama"
	ProviderLocal     AIWorkerProvider = "local"  // For other direct local model integrations
	ProviderCustom    AIWorkerProvider = "custom" // For other, unspecified types
)

// AIWorkerDefinitionStatus indicates the general status of a definition.
type AIWorkerDefinitionStatus string

const (
	DefinitionStatusActive   AIWorkerDefinitionStatus = "active"
	DefinitionStatusDisabled AIWorkerDefinitionStatus = "disabled"
	DefinitionStatusArchived AIWorkerDefinitionStatus = "archived"
)

// AIWorkerInstanceStatus indicates the current state of an AI worker instance.
type AIWorkerInstanceStatus string

const (
	InstanceStatusInitializing      AIWorkerInstanceStatus = "initializing"
	InstanceStatusIdle              AIWorkerInstanceStatus = "idle"
	InstanceStatusBusy              AIWorkerInstanceStatus = "busy"
	InstanceStatusContextFull       AIWorkerInstanceStatus = "context_full"
	InstanceStatusRateLimited       AIWorkerInstanceStatus = "rate_limited"
	InstanceStatusTokenLimitReached AIWorkerInstanceStatus = "token_limit_reached"
	InstanceStatusRetiredCompleted  AIWorkerInstanceStatus = "retired_completed"
	InstanceStatusRetiredExhausted  AIWorkerInstanceStatus = "retired_exhausted"
	InstanceStatusRetiredError      AIWorkerInstanceStatus = "retired_error"
	InstanceStatusError             AIWorkerInstanceStatus = "error"
)

// APIKeyStatus indicates the resolution status of an API key for an AIWorkerDefinition.
type APIKeyStatus string

const (
	APIKeyStatusUnknown       APIKeyStatus = "Unknown"
	APIKeyStatusFound         APIKeyStatus = "Found"
	APIKeyStatusNotFound      APIKeyStatus = "Not Found"
	APIKeyStatusNotConfigured APIKeyStatus = "Not Configured"
	APIKeyStatusError         APIKeyStatus = "Error Resolving"
)

// AIWorkerDefinitionDisplayInfo provides a snapshot of an AIWorkerDefinition
// along with transient status information useful for display or selection.
type AIWorkerDefinitionDisplayInfo struct {
	Definition    *AIWorkerDefinition
	IsChatCapable bool
	APIKeyStatus  APIKeyStatus
	// Add other TUI-relevant transient info here if needed
}

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

// TokenUsageMetrics tracks token consumption.
type TokenUsageMetrics struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	TotalTokens  int64 `json:"total_tokens"`
}

// SupervisorFeedback holds feedback, potentially from an SAI.
type SupervisorFeedback struct {
	Rating                 float64   `json:"rating,omitempty"`
	Comments               string    `json:"comments,omitempty"`
	CorrectionInstructions string    `json:"correction_instructions,omitempty"`
	SupervisorAgentID      string    `json:"supervisor_agent_id,omitempty"` // Could be an AIWorkerInstanceID or AIWorkerDefinition name
	FeedbackTimestamp      time.Time `json:"feedback_timestamp,omitempty"`
}

// AIWorkerPerformanceSummary provides aggregated performance statistics.
// This is typically embedded within AIWorkerDefinition.
type AIWorkerPerformanceSummary struct {
	TotalTasksAttempted   int       `json:"total_tasks_attempted"`
	SuccessfulTasks       int       `json:"successful_tasks"`
	FailedTasks           int       `json:"failed_tasks"`
	AverageSuccessRate    float64   `json:"average_success_rate"`            // Calculated: SuccessfulTasks / TotalTasksAttempted
	AverageDurationMs     float64   `json:"average_duration_ms"`             // For successful tasks
	TotalTokensProcessed  int64     `json:"total_tokens_processed"`          // Sum of all TokenUsageMetrics.TotalTokens
	TotalCostIncurred     float64   `json:"total_cost_incurred"`             // Sum of all PerformanceRecord.CostIncurred
	AverageQualityScore   float64   `json:"average_quality_score,omitempty"` // If supervisor feedback ratings are used
	LastActivityTimestamp time.Time `json:"last_activity_timestamp,omitempty"`
	TotalInstancesSpawned int       `json:"total_instances_spawned,omitempty"`
	ActiveInstancesCount  int       `json:"active_instances_count,omitempty"` // Runtime info, might not be persisted directly here
}

// --- New GlobalDataSourceDefinition ---
type DataSourceType string

const (
	DataSourceTypeLocalDirectory DataSourceType = "local_directory"
	DataSourceTypeFileAPI        DataSourceType = "file_api"
	// Future: DataSourceTypeGitRepo, DataSourceTypeS3Bucket etc.
)

type GlobalDataSourceDefinition struct {
	Name                    string                 `json:"name"` // Unique Name/ID for global reference by other definitions
	Type                    DataSourceType         `json:"type"`
	Description             string                 `json:"description,omitempty"`
	LocalPath               string                 `json:"local_path,omitempty"`                 // Relevant for DataSourceTypeLocalDirectory
	AllowExternalReadAccess bool                   `json:"allow_external_read_access,omitempty"` // CRITICAL: If true, LocalPath can be outside the main interpreter sandbox. Must be validated by AIWorkerManager.
	FileAPIPath             string                 `json:"file_api_path,omitempty"`              // Relevant for DataSourceTypeFileAPI (e.g., "fm:/shared_data/my_folder")
	RemoteTargetPath        string                 `json:"remote_target_path,omitempty"`         // Default target path in File API if this source (e.g. a local dir) is synced. Example: "synced_sources/<DataSourceName>"
	ReadOnly                bool                   `json:"read_only"`                            // Hint for usage; write operations are always sandboxed.
	Filters                 []string               `json:"filters,omitempty"`                    // Glob patterns for file inclusion, e.g., ["*.txt", "*.log"]
	Recursive               bool                   `json:"recursive,omitempty"`                  // Whether to process directories recursively.
	Metadata                map[string]interface{} `json:"metadata,omitempty"`                   // For custom annotations
	CreatedTimestamp        time.Time              `json:"created_timestamp,omitempty"`          // Set by manager
	ModifiedTimestamp       time.Time              `json:"modified_timestamp,omitempty"`         // Set by manager
}

// AIWorkerDefinition represents the blueprint for an AI worker.
type AIWorkerDefinition struct {
	DefinitionID                string                      `json:"definitionID"`
	Name                        string                      `json:"name"`
	Provider                    AIWorkerProvider            `json:"provider"`
	ModelName                   string                      `json:"modelName"`
	Auth                        APIKeySource                `json:"auth"`
	InteractionModels           []InteractionModelType      `json:"interactionModels,omitempty"`
	Capabilities                []string                    `json:"capabilities,omitempty"`
	BaseConfig                  map[string]interface{}      `json:"baseConfig,omitempty"`
	CostMetrics                 map[string]float64          `json:"costMetrics,omitempty"`
	RateLimits                  RateLimitPolicy             `json:"rateLimits,omitempty"`
	Status                      AIWorkerDefinitionStatus    `json:"status,omitempty"`
	DefaultFileContexts         []string                    `json:"defaultFileContexts,omitempty"`
	DataSourceRefs              []string                    `json:"dataSourceRefs,omitempty"`
	ToolAllowlist               []string                    `json:"toolAllowlist,omitempty"`
	ToolDenylist                []string                    `json:"toolDenylist,omitempty"`
	DefaultSupervisoryAIRef     string                      `json:"defaultSupervisoryAIRef,omitempty"`
	AggregatePerformanceSummary *AIWorkerPerformanceSummary `json:"aggregatePerformanceSummary,omitempty"`
	Metadata                    map[string]interface{}      `json:"metadata,omitempty"`
	CreatedTimestamp            time.Time                   `json:"createdTimestamp,omitempty"`
	ModifiedTimestamp           time.Time                   `json:"modifiedTimestamp,omitempty"`
}

// --- AIWorkerInstance (incorporates existing fields and new references) ---
// Represents an active or retired conversational session or a pooled worker.
type AIWorkerInstance struct {
	InstanceID            string                 `json:"instance_id"`   // System-generated UUID
	DefinitionID          string                 `json:"definition_id"` // ID of the AIWorkerDefinition it's based on
	Status                AIWorkerInstanceStatus `json:"status"`
	ConversationHistory   []*ConversationTurn    `json:"-"` // Holds the history of conversation turns for this instance.
	CreationTimestamp     time.Time              `json:"creation_timestamp"`
	LastActivityTimestamp time.Time              `json:"last_activity_timestamp"`
	SessionTokenUsage     TokenUsageMetrics      `json:"session_token_usage"`
	CurrentConfig         map[string]interface{} `json:"current_config,omitempty"` // Effective config, may include overrides
	ActiveFileContexts    []string               `json:"-"`                        // Runtime only, not persisted with instance JSON
	LastError             string                 `json:"last_error,omitempty"`
	RetirementReason      string                 `json:"retirement_reason,omitempty"`

	// New fields for v0.5
	PoolID           string   `json:"pool_id,omitempty"`            // If this instance is managed by an AIWorkerPool
	CurrentTaskID    string   `json:"current_task_id,omitempty"`    // If currently processing a WorkItem
	DataSourceRefs   []string `json:"data_source_refs,omitempty"`   // Instance-specific dynamically attached GlobalDataSourceDefinition names (augments definition/pool/queue/item refs)
	SupervisoryAIRef string   `json:"supervisory_ai_ref,omitempty"` // Instance-specific SAI (Future)

	llmClient LLMClient `json:"-"` // Runtime: LLM client for this instance to use.
}

// ProcessChatMessage sends a user message to the LLM for this instance and updates its history.
// It uses the llmClient associated with the instance.
func (instance *AIWorkerInstance) ProcessChatMessage(ctx context.Context, userMessageText string) (*ConversationTurn, error) {
	if instance.llmClient == nil {
		return nil, NewRuntimeError(ErrorCodePreconditionFailed, "LLM client not set for instance", nil)
	}
	if instance.Status == InstanceStatusRetiredCompleted || instance.Status == InstanceStatusRetiredError || instance.Status == InstanceStatusRetiredExhausted {
		return nil, NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("instance %s is retired and cannot process messages", instance.InstanceID), nil)
	}
	if instance.Status == InstanceStatusBusy {
		return nil, NewRuntimeError(ErrorCodePreconditionFailed, fmt.Sprintf("instance %s is currently busy", instance.InstanceID), nil)
	}

	// 1. Construct and add user message to instance.ConversationHistory
	userTurn := &ConversationTurn{
		Role:    RoleUser, // Assumes RoleUser is defined in llm_types.go or similar
		Content: userMessageText,
	}
	instance.ConversationHistory = append(instance.ConversationHistory, userTurn)
	instance.LastActivityTimestamp = time.Now()
	previousStatus := instance.Status
	instance.Status = InstanceStatusBusy // Mark as busy

	// 2. Call llmClient.Ask
	// The llmClient.Ask method takes the current full history.
	modelResponseTurn, err := instance.llmClient.Ask(ctx, instance.ConversationHistory)

	if err != nil {
		instance.Status = InstanceStatusError
		instance.LastError = err.Error()
		// Do not return the error directly if it's already a RuntimeError
		if rtErr, ok := err.(*RuntimeError); ok {
			return nil, rtErr
		}
		return nil, NewRuntimeError(ErrorCodeLLMError, fmt.Sprintf("LLM Ask failed for instance %s: %v", instance.InstanceID, err), err)
	}

	// 3. Add model response to instance.ConversationHistory
	if modelResponseTurn != nil {
		instance.ConversationHistory = append(instance.ConversationHistory, modelResponseTurn)
		// Update token usage from the model's response turn
		// Ensure TokenUsage field exists and is populated by the LLMClient implementation.
		instance.SessionTokenUsage.InputTokens += modelResponseTurn.TokenUsage.InputTokens
		instance.SessionTokenUsage.OutputTokens += modelResponseTurn.TokenUsage.OutputTokens
		instance.SessionTokenUsage.TotalTokens = instance.SessionTokenUsage.InputTokens + instance.SessionTokenUsage.OutputTokens
	}

	instance.LastActivityTimestamp = time.Now()
	// If it was busy due to this call, set it back to what it was before, or idle.
	if previousStatus == InstanceStatusBusy || previousStatus == InstanceStatusInitializing {
		instance.Status = InstanceStatusIdle
	} else {
		instance.Status = previousStatus
	}
	if modelResponseTurn == nil && err == nil {
		// This case (nil response, nil error) should ideally be handled by the LLMClient
		// or represent a specific scenario (e.g., content filtered by safety settings).
		// For now, we'll assume the LLMClient might return this and we pass nil back.
		// Log a warning or debug message here.
		// instance.logger.Debugf("ProcessChatMessage for instance %s received nil modelResponseTurn and nil error from LLMClient.", instance.InstanceID)
	}

	return modelResponseTurn, nil
}

// --- PerformanceRecord (existing struct, ensure TaskID can link to WorkItem.TaskID) ---
type PerformanceRecord struct {
	TaskID             string                 `json:"task_id"`     // Can be WorkItem.TaskID if applicable
	InstanceID         string                 `json:"instance_id"` // Can be "stateless-<uuid>" for direct calls
	DefinitionID       string                 `json:"definition_id"`
	TimestampStart     time.Time              `json:"timestamp_start"`
	TimestampEnd       time.Time              `json:"timestamp_end"`
	DurationMs         int64                  `json:"duration_ms"`
	Success            bool                   `json:"success"`
	InputContext       map[string]interface{} `json:"input_context,omitempty"` // e.g., prompt hash, summary of WorkItem payload
	LLMMetrics         map[string]interface{} `json:"llm_metrics,omitempty"`   // Raw metrics from LLM (tokens, finish reason etc.)
	CostIncurred       float64                `json:"cost_incurred,omitempty"`
	OutputSummary      string                 `json:"output_summary,omitempty"` // Trimmed or hashed output
	ErrorDetails       string                 `json:"error_details,omitempty"`
	SupervisorFeedback *SupervisorFeedback    `json:"supervisor_feedback,omitempty"`
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
	InitialFileContexts []string               `json:"initial_file_contexts,omitempty"`
	PerformanceRecords  []*PerformanceRecord   `json:"performance_records"`
}

// --- AIWorkerPoolDefinition ---
type InstanceRetirementPolicy struct {
	MaxTasksPerInstance int `json:"max_tasks_per_instance,omitempty"`
	MaxInstanceAgeHours int `json:"max_instance_age_hours,omitempty"`
}

type AIWorkerPoolDefinition struct {
	PoolID                       string                   `json:"pool_id"`                          // System-generated UUID
	Name                         string                   `json:"name"`                             // User-defined, unique
	TargetAIWorkerDefinitionName string                   `json:"target_ai_worker_definition_name"` // Name of an AIWorkerDefinition
	MinIdleInstances             int                      `json:"min_idle_instances,omitempty"`
	MaxTotalInstances            int                      `json:"max_total_instances,omitempty"`
	InstanceRetirementPolicy     InstanceRetirementPolicy `json:"instance_retirement_policy,omitempty"`
	DataSourceRefs               []string                 `json:"data_source_refs,omitempty"`   // Names of GlobalDataSourceDefinitions applicable to all instances in this pool
	SupervisoryAIRef             string                   `json:"supervisory_ai_ref,omitempty"` // Name of an AIWorkerDefinition for SAI (Future)
	Metadata                     map[string]interface{}   `json:"metadata,omitempty"`
	CreatedTimestamp             time.Time                `json:"created_timestamp,omitempty"`
	ModifiedTimestamp            time.Time                `json:"modified_timestamp,omitempty"`
}

// --- WorkQueueDefinition ---
type RetryPolicy struct {
	MaxRetries        int `json:"max_retries,omitempty"`
	RetryDelaySeconds int `json:"retry_delay_seconds,omitempty"`
}

type WorkQueueDefinition struct {
	QueueID             string                 `json:"queue_id"`              // System-generated UUID
	Name                string                 `json:"name"`                  // User-defined, unique
	AssociatedPoolNames []string               `json:"associated_pool_names"` // Names of AIWorkerPoolDefinitions that service this queue
	DefaultPriority     int                    `json:"default_priority,omitempty"`
	RetryPolicy         RetryPolicy            `json:"retry_policy,omitempty"`
	PersistTasks        bool                   `json:"persist_tasks,omitempty"`      // Hint for future: if true, WorkItems should be persisted
	DataSourceRefs      []string               `json:"data_source_refs,omitempty"`   // Names of GlobalDataSourceDefinitions relevant to all tasks in this queue
	SupervisoryAIRef    string                 `json:"supervisory_ai_ref,omitempty"` // Name of an AIWorkerDefinition for SAI (Future)
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	CreatedTimestamp    time.Time              `json:"created_timestamp,omitempty"`
	ModifiedTimestamp   time.Time              `json:"modified_timestamp,omitempty"`
}

// --- WorkItemDefinition ---
type WorkItemDefinition struct {
	WorkItemDefinitionID        string                 `json:"work_item_definition_id"` // System-generated UUID
	Name                        string                 `json:"name"`                    // User-defined, unique
	Description                 string                 `json:"description,omitempty"`
	DefaultTargetWorkerCriteria map[string]interface{} `json:"default_target_worker_criteria,omitempty"` // e.g., {"definitionName": "panel_analyzer_v2", "capabilities": ["stress_analysis"]}
	DefaultPayloadSchema        map[string]interface{} `json:"default_payload_schema,omitempty"`         // Map of default values or a more formal JSON schema
	DefaultDataSourceRefs       []string               `json:"default_data_source_refs,omitempty"`       // Names of GlobalDataSourceDefinitions
	DefaultPriority             int                    `json:"default_priority,omitempty"`
	DefaultSupervisoryAIRef     string                 `json:"default_supervisory_ai_ref,omitempty"` // Name of an AIWorkerDefinition for SAI (Future)
	Metadata                    map[string]interface{} `json:"metadata,omitempty"`
	CreatedTimestamp            time.Time              `json:"created_timestamp,omitempty"`
	ModifiedTimestamp           time.Time              `json:"modified_timestamp,omitempty"`
}

// --- WorkItem ---
type WorkItemStatus string

const (
	WorkItemStatusPending    WorkItemStatus = "pending"
	WorkItemStatusProcessing WorkItemStatus = "processing"
	WorkItemStatusCompleted  WorkItemStatus = "completed"
	WorkItemStatusFailed     WorkItemStatus = "failed"
	WorkItemStatusRetrying   WorkItemStatus = "retrying"
	WorkItemStatusCancelled  WorkItemStatus = "cancelled" // Future
)

// WorkItem represents a task to be processed.
type WorkItem struct {
	TaskID                 string                 `json:"task_id"`                             // System-generated UUID if not provided on submission
	WorkItemDefinitionName string                 `json:"work_item_definition_name,omitempty"` // Optional: Name of WorkItemDefinition to use as template
	QueueName              string                 `json:"queue_name"`                          // Name of the WorkQueueDefinition it's submitted to
	TargetWorkerCriteria   map[string]interface{} `json:"target_worker_criteria,omitempty"`    // Overrides criteria from WorkItemDefinition or Queue
	Payload                map[string]interface{} `json:"payload"`                             // Task-specific data; merged with/validated by WorkItemDefinition's schema
	DataSourceRefs         []string               `json:"data_source_refs,omitempty"`          // Augments/overrides DataSourceRefs from other contexts
	Priority               int                    `json:"priority,omitempty"`                  // Overrides defaults
	Status                 WorkItemStatus         `json:"status,omitempty"`                    // Set by system; typically "pending" on submission
	SubmitTimestamp        time.Time              `json:"submit_timestamp,omitempty"`          // Set by system
	StartTimestamp         time.Time              `json:"start_timestamp,omitempty"`           // Set by system
	EndTimestamp           time.Time              `json:"end_timestamp,omitempty"`             // Set by system
	RetryCount             int                    `json:"retry_count,omitempty"`               // Managed by system
	Result                 interface{}            `json:"result,omitempty"`                    // Stored upon successful completion
	Error                  string                 `json:"error,omitempty"`                     // Error message if failed
	PerformanceRecordID    string                 `json:"performance_record_id,omitempty"`     // Link to the PerformanceRecord
	SupervisoryAIRef       string                 `json:"supervisory_ai_ref,omitempty"`        // Specific SAI for this item (Future)
	Metadata               map[string]interface{} `json:"metadata,omitempty"`
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
