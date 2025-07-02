// NeuroScript Version: 0.4.0
// File version: 0.3.10
// Description: Defines types for the AI Worker Management system. Stringer methods moved to core/ai_worker_stringers.go.
// filename: pkg/core/ai_worker_types.go
// nlines: 305 // Estimate, will be less than original 453

package core

import (
	"context"
	"fmt"

	// "strings" // No longer needed here if all String methods are moved
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
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
	APIKeyMethodInline     APIKeySourceMethod = "inline"
	APIKeyMethodConfigPath APIKeySourceMethod = "config_path"
	APIKeyMethodVault      APIKeySourceMethod = "vault"
	APIKeyMethodNone       APIKeySourceMethod = "none"
)

// AIWorkerProvider represents the source or type of the AI worker.
type AIWorkerProvider string

const (
	ProviderGoogle    AIWorkerProvider = "google"
	ProviderOpenAI    AIWorkerProvider = "openai"
	ProviderAnthropic AIWorkerProvider = "anthropic"
	ProviderOllama    AIWorkerProvider = "ollama"
	ProviderLocal     AIWorkerProvider = "local"
	ProviderCustom    AIWorkerProvider = "custom"
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
}

// APIKeySource specifies where to find the API key.
type APIKeySource struct {
	Method APIKeySourceMethod `json:"method"`
	Value  string             `json:"value,omitempty"`
}

// RateLimitPolicy defines usage limits for an AIWorkerDefinition.
type RateLimitPolicy struct {
	MaxRequestsPerMinute         int `json:"max_requests_per_minute,omitempty"`
	MaxTokensPerMinute           int `json:"max_tokens_per_minute,omitempty"`
	MaxTokensPerDay              int `json:"max_tokens_per_day,omitempty"`
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
	SupervisorAgentID      string    `json:"supervisor_agent_id,omitempty"`
	FeedbackTimestamp      time.Time `json:"feedback_timestamp,omitempty"`
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
	TotalInstancesSpawned int       `json:"total_instances_spawned,omitempty"`
	ActiveInstancesCount  int       `json:"active_instances_count,omitempty"`
}

// --- New GlobalDataSourceDefinition ---
type DataSourceType string

const (
	DataSourceTypeLocalDirectory DataSourceType = "local_directory"
	DataSourceTypeFileAPI        DataSourceType = "file_api"
)

type GlobalDataSourceDefinition struct {
	Name                    string                 `json:"name"`
	Type                    DataSourceType         `json:"type"`
	Description             string                 `json:"description,omitempty"`
	LocalPath               string                 `json:"local_path,omitempty"`
	AllowExternalReadAccess bool                   `json:"allow_external_read_access,omitempty"`
	FileAPIPath             string                 `json:"file_api_path,omitempty"`
	RemoteTargetPath        string                 `json:"remote_target_path,omitempty"`
	ReadOnly                bool                   `json:"read_only"`
	Filters                 []string               `json:"filters,omitempty"`
	Recursive               bool                   `json:"recursive,omitempty"`
	Metadata                map[string]interface{} `json:"metadata,omitempty"`
	CreatedTimestamp        time.Time              `json:"created_timestamp,omitempty"`
	ModifiedTimestamp       time.Time              `json:"modified_timestamp,omitempty"`
}

// AIWorkerDefinition represents the blueprint for an AI worker.
type AIWorkerDefinition struct {
	DefinitionID                string                      `json:"definitionID"`
	Name                        string                      `json:"name"`
	Provider                    AIWorkerProvider            `json:"provider"`
	ModelName                   string                      `json:"modelName"`
	Auth                        APIKeySource                `json:"auth"`
	InteractionModels           []InteractionModelType      `json:"interaction_models,omitempty"`
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

// --- AIWorkerInstance ---
type AIWorkerInstance struct {
	InstanceID            string                         `json:"instance_id"`
	DefinitionID          string                         `json:"definition_id"`
	Status                AIWorkerInstanceStatus         `json:"status"`
	ConversationHistory   []*interfaces.ConversationTurn `json:"-"`
	CreationTimestamp     time.Time                      `json:"creation_timestamp"`
	LastActivityTimestamp time.Time                      `json:"last_activity_timestamp"`
	SessionTokenUsage     TokenUsageMetrics              `json:"session_token_usage"`
	CurrentConfig         map[string]interface{}         `json:"current_config,omitempty"`
	ActiveFileContexts    []string                       `json:"-"`
	LastError             string                         `json:"last_error,omitempty"`
	RetirementReason      string                         `json:"retirement_reason,omitempty"`
	PoolID                string                         `json:"pool_id,omitempty"`
	CurrentTaskID         string                         `json:"current_task_id,omitempty"`
	DataSourceRefs        []string                       `json:"data_source_refs,omitempty"`
	SupervisoryAIRef      string                         `json:"supervisory_ai_ref,omitempty"`
	llmClient             interfaces.LLMClient           `json:"-"`
}

// ProcessChatMessage method remains here as it's operational logic.
func (instance *AIWorkerInstance) ProcessChatMessage(ctx context.Context, userMessageText string) (*interfaces.ConversationTurn, error) {
	if instance.llmClient == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodePreconditionFailed, "LLM client not set for instance", nil)
	}
	if instance.Status == InstanceStatusRetiredCompleted || instance.Status == InstanceStatusRetiredError || instance.Status == InstanceStatusRetiredExhausted {
		return nil, lang.NewRuntimeError(lang.ErrorCodePreconditionFailed, fmt.Sprintf("instance %s is retired and cannot process messages", instance.InstanceID), nil)
	}
	if instance.Status == InstanceStatusBusy {
		return nil, lang.NewRuntimeError(lang.ErrorCodePreconditionFailed, fmt.Sprintf("instance %s is currently busy", instance.InstanceID), nil)
	}

	userTurn := &interfaces.ConversationTurn{
		Role:    interfaces.RoleUser,
		Content: userMessageText,
	}
	instance.ConversationHistory = append(instance.ConversationHistory, userTurn)
	instance.LastActivityTimestamp = time.Now()
	previousStatus := instance.Status
	instance.Status = InstanceStatusBusy

	modelResponseTurn, err := instance.llmClient.Ask(ctx, instance.ConversationHistory)

	if err != nil {
		instance.Status = InstanceStatusError
		instance.LastError = err.Error()
		if rtErr, ok := err.(*lang.RuntimeError); ok {
			return nil, rtErr
		}
		return nil, lang.NewRuntimeError(lang.ErrorCodeLLMError, fmt.Sprintf("LLM Ask failed for instance %s: %v", instance.InstanceID, err), err)
	}

	// if modelResponseTurn != nil {
	// 	instance.ConversationHistory = append(instance.ConversationHistory, modelResponseTurn)
	// 	instance.SessionTokenUsage.InputTokens += modelResponseTurn.TokenUsage.InputTokens
	// 	instance.SessionTokenUsage.OutputTokens += modelResponseTurn.TokenUsage.OutputTokens
	// 	instance.SessionTokenUsage.TotalTokens = instance.SessionTokenUsage.InputTokens + instance.SessionTokenUsage.OutputTokens
	// }

	instance.LastActivityTimestamp = time.Now()
	if previousStatus == InstanceStatusBusy || previousStatus == InstanceStatusInitializing {
		instance.Status = InstanceStatusIdle
	} else {
		instance.Status = previousStatus
	}
	return modelResponseTurn, nil
}

// --- PerformanceRecord ---
type PerformanceRecord struct {
	TaskID             string                 `json:"task_id"`
	InstanceID         string                 `json:"instance_id"`
	DefinitionID       string                 `json:"definition_id"`
	TimestampStart     time.Time              `json:"timestamp_start"`
	TimestampEnd       time.Time              `json:"timestamp_end"`
	DurationMs         int64                  `json:"duration_ms"`
	Success            bool                   `json:"success"`
	InputContext       map[string]interface{} `json:"input_context,omitempty"`
	LLMMetrics         map[string]interface{} `json:"llm_metrics,omitempty"`
	CostIncurred       float64                `json:"cost_incurred,omitempty"`
	OutputSummary      string                 `json:"output_summary,omitempty"`
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
	PoolID                       string                   `json:"pool_id"`
	Name                         string                   `json:"name"`
	TargetAIWorkerDefinitionName string                   `json:"target_ai_worker_definition_name"`
	MinIdleInstances             int                      `json:"min_idle_instances,omitempty"`
	MaxTotalInstances            int                      `json:"max_total_instances,omitempty"`
	InstanceRetirementPolicy     InstanceRetirementPolicy `json:"instance_retirement_policy,omitempty"`
	DataSourceRefs               []string                 `json:"data_source_refs,omitempty"`
	SupervisoryAIRef             string                   `json:"supervisory_ai_ref,omitempty"`
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
	QueueID             string                 `json:"queue_id"`
	Name                string                 `json:"name"`
	AssociatedPoolNames []string               `json:"associated_pool_names"`
	DefaultPriority     int                    `json:"default_priority,omitempty"`
	RetryPolicy         RetryPolicy            `json:"retry_policy,omitempty"`
	PersistTasks        bool                   `json:"persist_tasks,omitempty"`
	DataSourceRefs      []string               `json:"data_source_refs,omitempty"`
	SupervisoryAIRef    string                 `json:"supervisory_ai_ref,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	CreatedTimestamp    time.Time              `json:"created_timestamp,omitempty"`
	ModifiedTimestamp   time.Time              `json:"modified_timestamp,omitempty"`
}

// --- WorkItemDefinition ---
type WorkItemDefinition struct {
	WorkItemDefinitionID        string                 `json:"work_item_definition_id"`
	Name                        string                 `json:"name"`
	Description                 string                 `json:"description,omitempty"`
	DefaultTargetWorkerCriteria map[string]interface{} `json:"default_target_worker_criteria,omitempty"`
	DefaultPayloadSchema        map[string]interface{} `json:"default_payload_schema,omitempty"`
	DefaultDataSourceRefs       []string               `json:"default_data_source_refs,omitempty"`
	DefaultPriority             int                    `json:"default_priority,omitempty"`
	DefaultSupervisoryAIRef     string                 `json:"default_supervisory_ai_ref,omitempty"`
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
	WorkItemStatusCancelled  WorkItemStatus = "cancelled"
)

type WorkItem struct {
	TaskID                 string                 `json:"task_id"`
	WorkItemDefinitionName string                 `json:"work_item_definition_name,omitempty"`
	QueueName              string                 `json:"queue_name"`
	TargetWorkerCriteria   map[string]interface{} `json:"target_worker_criteria,omitempty"`
	Payload                map[string]interface{} `json:"payload"`
	DataSourceRefs         []string               `json:"data_source_refs,omitempty"`
	Priority               int                    `json:"priority,omitempty"`
	Status                 WorkItemStatus         `json:"status,omitempty"`
	SubmitTimestamp        time.Time              `json:"submit_timestamp,omitempty"`
	StartTimestamp         time.Time              `json:"start_timestamp,omitempty"`
	EndTimestamp           time.Time              `json:"end_timestamp,omitempty"`
	RetryCount             int                    `json:"retry_count,omitempty"`
	Result                 interface{}            `json:"result,omitempty"`
	Error                  string                 `json:"error,omitempty"`
	PerformanceRecordID    string                 `json:"performance_record_id,omitempty"`
	SupervisoryAIRef       string                 `json:"supervisory_ai_ref,omitempty"`
	Metadata               map[string]interface{} `json:"metadata,omitempty"`
}

// LLMCallMetrics holds detailed metrics from a specific LLM API call.
type LLMCallMetrics struct {
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	TotalTokens  int64  `json:"total_tokens"`
	FinishReason string `json:"finish_reason,omitempty"`
	ModelUsed    string `json:"model_used,omitempty"`
}
