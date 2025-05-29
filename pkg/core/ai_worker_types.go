// NeuroScript Version: 0.4.0
// File version: 0.3.9
// Description: Added String() methods to various types for better string representation and TUI display.
// filename: pkg/core/ai_worker_types.go

package core

import (
	"context"
	"fmt"
	"strings"
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

func (di *AIWorkerDefinitionDisplayInfo) String() string {
	if di == nil {
		return "<nil AIWorkerDefinitionDisplayInfo>"
	}
	var defStr string
	if di.Definition != nil {
		defStr = di.Definition.String()
	} else {
		defStr = "<nil Definition>"
	}
	return fmt.Sprintf("DisplayInfo: Capable: %t, KeyStatus: %s, Def: %s", di.IsChatCapable, di.APIKeyStatus, defStr)
}

// APIKeySource specifies where to find the API key.
type APIKeySource struct {
	Method APIKeySourceMethod `json:"method"`
	Value  string             `json:"value,omitempty"` // e.g., "GOOGLE_API_KEY" or the actual key
}

func (aks *APIKeySource) String() string {
	if aks == nil {
		return "<nil APIKeySource>"
	}
	val := aks.Value
	if aks.Method == APIKeyMethodInline && val != "" {
		val = "[redacted]"
	}
	return fmt.Sprintf("Method: %s, Value: '%s'", aks.Method, val)
}

// RateLimitPolicy defines usage limits for an AIWorkerDefinition.
type RateLimitPolicy struct {
	MaxRequestsPerMinute         int `json:"max_requests_per_minute,omitempty"`
	MaxTokensPerMinute           int `json:"max_tokens_per_minute,omitempty"` // Input + Output
	MaxTokensPerDay              int `json:"max_tokens_per_day,omitempty"`    // Input + Output
	MaxConcurrentActiveInstances int `json:"max_concurrent_active_instances,omitempty"`
}

func (rlp *RateLimitPolicy) String() string {
	if rlp == nil {
		return "<nil RateLimitPolicy>"
	}
	return fmt.Sprintf("Req/Min: %d, Tok/Min: %d, Tok/Day: %d, MaxInstances: %d",
		rlp.MaxRequestsPerMinute, rlp.MaxTokensPerMinute, rlp.MaxTokensPerDay, rlp.MaxConcurrentActiveInstances)
}

// TokenUsageMetrics tracks token consumption.
type TokenUsageMetrics struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	TotalTokens  int64 `json:"total_tokens"`
}

func (tum *TokenUsageMetrics) String() string {
	if tum == nil {
		return "<nil TokenUsageMetrics>"
	}
	return fmt.Sprintf("In: %d, Out: %d, Total: %d", tum.InputTokens, tum.OutputTokens, tum.TotalTokens)
}

// SupervisorFeedback holds feedback, potentially from an SAI.
type SupervisorFeedback struct {
	Rating                 float64   `json:"rating,omitempty"`
	Comments               string    `json:"comments,omitempty"`
	CorrectionInstructions string    `json:"correction_instructions,omitempty"`
	SupervisorAgentID      string    `json:"supervisor_agent_id,omitempty"` // Could be an AIWorkerInstanceID or AIWorkerDefinition name
	FeedbackTimestamp      time.Time `json:"feedback_timestamp,omitempty"`
}

func (sf *SupervisorFeedback) String() string {
	if sf == nil {
		return "<nil SupervisorFeedback>"
	}
	return fmt.Sprintf("Rating: %.1f, Supervisor: %s, Comments: %.30s...", sf.Rating, sf.SupervisorAgentID, sf.Comments)
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
	ActiveInstancesCount  int       `json:"active_instances_count,omitempty"` // Runtime info
}

func (ps *AIWorkerPerformanceSummary) String() string {
	if ps == nil {
		return "<nil AIWorkerPerformanceSummary>"
	}
	return fmt.Sprintf("Tasks: %d (S:%d, F:%d), SuccessRate: %.2f%%, AvgDur: %.0fms, Tokens: %d, Cost: $%.4f, ActiveInst: %d",
		ps.TotalTasksAttempted, ps.SuccessfulTasks, ps.FailedTasks, ps.AverageSuccessRate*100,
		ps.AverageDurationMs, ps.TotalTokensProcessed, ps.TotalCostIncurred, ps.ActiveInstancesCount)
}

// --- New GlobalDataSourceDefinition ---
type DataSourceType string

const (
	DataSourceTypeLocalDirectory DataSourceType = "local_directory"
	DataSourceTypeFileAPI        DataSourceType = "file_api"
	// Future: DataSourceTypeGitRepo, DataSourceTypeS3Bucket etc.
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

func (dsd *GlobalDataSourceDefinition) String() string {
	if dsd == nil {
		return "<nil GlobalDataSourceDefinition>"
	}
	path := dsd.LocalPath
	if dsd.Type == DataSourceTypeFileAPI {
		path = dsd.FileAPIPath
	}
	return fmt.Sprintf("DS '%s': Type: %s, Path: '%s', RO: %t, Filters: %d",
		dsd.Name, dsd.Type, path, dsd.ReadOnly, len(dsd.Filters))
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

func (wd *AIWorkerDefinition) String() string {
	if wd == nil {
		return "<nil AIWorkerDefinition>"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Def ID: %s\n", wd.DefinitionID))
	sb.WriteString(fmt.Sprintf("  Name: %s\n", wd.Name))
	sb.WriteString(fmt.Sprintf("  Provider: %s, Model: %s\n", wd.Provider, wd.ModelName))
	sb.WriteString(fmt.Sprintf("  Status: %s\n", wd.Status))
	sb.WriteString(fmt.Sprintf("  Auth: %s\n", wd.Auth.String()))
	sb.WriteString(fmt.Sprintf("  InteractionModels: %v\n", wd.InteractionModels))
	sb.WriteString(fmt.Sprintf("  Capabilities: %d\n", len(wd.Capabilities)))
	if wd.AggregatePerformanceSummary != nil {
		sb.WriteString(fmt.Sprintf("  Performance: %s\n", wd.AggregatePerformanceSummary.String()))
	} else {
		sb.WriteString("  Performance: <nil>\n")
	}
	sb.WriteString(fmt.Sprintf("  DataSources: %d, ToolsAllowed: %d, ToolsDenied: %d\n", len(wd.DataSourceRefs), len(wd.ToolAllowlist), len(wd.ToolDenylist)))
	return sb.String()
}

// --- AIWorkerInstance (incorporates existing fields and new references) ---
type AIWorkerInstance struct {
	InstanceID            string                 `json:"instance_id"`
	DefinitionID          string                 `json:"definition_id"`
	Status                AIWorkerInstanceStatus `json:"status"`
	ConversationHistory   []*ConversationTurn    `json:"-"`
	CreationTimestamp     time.Time              `json:"creation_timestamp"`
	LastActivityTimestamp time.Time              `json:"last_activity_timestamp"`
	SessionTokenUsage     TokenUsageMetrics      `json:"session_token_usage"`
	CurrentConfig         map[string]interface{} `json:"current_config,omitempty"`
	ActiveFileContexts    []string               `json:"-"`
	LastError             string                 `json:"last_error,omitempty"`
	RetirementReason      string                 `json:"retirement_reason,omitempty"`
	PoolID                string                 `json:"pool_id,omitempty"`
	CurrentTaskID         string                 `json:"current_task_id,omitempty"`
	DataSourceRefs        []string               `json:"data_source_refs,omitempty"`
	SupervisoryAIRef      string                 `json:"supervisory_ai_ref,omitempty"`
	llmClient             LLMClient              `json:"-"`
}

func (wi *AIWorkerInstance) String() string {
	if wi == nil {
		return "<nil AIWorkerInstance>"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Instance ID: %s (Def: %s)\n", wi.InstanceID, wi.DefinitionID))
	sb.WriteString(fmt.Sprintf("  Status: %s\n", wi.Status))
	if wi.PoolID != "" {
		sb.WriteString(fmt.Sprintf("  PoolID: %s\n", wi.PoolID))
	}
	if wi.CurrentTaskID != "" {
		sb.WriteString(fmt.Sprintf("  TaskID: %s\n", wi.CurrentTaskID))
	}
	sb.WriteString(fmt.Sprintf("  Tokens: %s\n", wi.SessionTokenUsage.String()))
	sb.WriteString(fmt.Sprintf("  History Turns: %d\n", len(wi.ConversationHistory)))
	sb.WriteString(fmt.Sprintf("  Last Activity: %s\n", wi.LastActivityTimestamp.Format(time.RFC3339)))
	if wi.LastError != "" {
		sb.WriteString(fmt.Sprintf("  Error: %s\n", wi.LastError))
	}
	return sb.String()
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
		// instance.logger.Debugf("ProcessChatMessage for instance %s received nil modelResponseTurn and nil error from LLMClient.", instance.InstanceID)
	}

	return modelResponseTurn, nil
}

// --- PerformanceRecord (existing struct, ensure TaskID can link to WorkItem.TaskID) ---
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

func (pr *PerformanceRecord) String() string {
	if pr == nil {
		return "<nil PerformanceRecord>"
	}
	status := "FAIL"
	if pr.Success {
		status = "OK"
	}
	return fmt.Sprintf("Task: %s (Inst: %s, Def: %s) Status: %s, Dur: %dms, Cost: $%.4f, Error: %.30s...",
		pr.TaskID, pr.InstanceID, pr.DefinitionID, status, pr.DurationMs, pr.CostIncurred, pr.ErrorDetails)
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

func (rii *RetiredInstanceInfo) String() string {
	if rii == nil {
		return "<nil RetiredInstanceInfo>"
	}
	return fmt.Sprintf("Retired Inst: %s (Def: %s), Status: %s, Reason: %s, Tokens: %s, PerfRecs: %d",
		rii.InstanceID, rii.DefinitionID, rii.FinalStatus, rii.RetirementReason, rii.SessionTokenUsage.String(), len(rii.PerformanceRecords))
}

// --- AIWorkerPoolDefinition ---
type InstanceRetirementPolicy struct {
	MaxTasksPerInstance int `json:"max_tasks_per_instance,omitempty"`
	MaxInstanceAgeHours int `json:"max_instance_age_hours,omitempty"`
}

func (irp *InstanceRetirementPolicy) String() string {
	if irp == nil {
		return "<nil InstanceRetirementPolicy>"
	}
	return fmt.Sprintf("MaxTasks: %d, MaxAgeHours: %d", irp.MaxTasksPerInstance, irp.MaxInstanceAgeHours)
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

func (wpd *AIWorkerPoolDefinition) String() string {
	if wpd == nil {
		return "<nil AIWorkerPoolDefinition>"
	}
	return fmt.Sprintf("Pool '%s' (ID: %s): TargetDef: %s, Idle: %d, Max: %d, Retirement: [%s], DS: %d",
		wpd.Name, wpd.PoolID, wpd.TargetAIWorkerDefinitionName, wpd.MinIdleInstances, wpd.MaxTotalInstances,
		wpd.InstanceRetirementPolicy.String(), len(wpd.DataSourceRefs))
}

// --- WorkQueueDefinition ---
type RetryPolicy struct {
	MaxRetries        int `json:"max_retries,omitempty"`
	RetryDelaySeconds int `json:"retry_delay_seconds,omitempty"`
}

func (rp *RetryPolicy) String() string {
	if rp == nil {
		return "<nil RetryPolicy>"
	}
	return fmt.Sprintf("MaxRetries: %d, Delay: %ds", rp.MaxRetries, rp.RetryDelaySeconds)
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

func (wqd *WorkQueueDefinition) String() string {
	if wqd == nil {
		return "<nil WorkQueueDefinition>"
	}
	return fmt.Sprintf("Queue '%s' (ID: %s): Pools: %v, Prio: %d, Retry: [%s], Persist: %t, DS: %d",
		wqd.Name, wqd.QueueID, wqd.AssociatedPoolNames, wqd.DefaultPriority, wqd.RetryPolicy.String(),
		wqd.PersistTasks, len(wqd.DataSourceRefs))
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

func (wid *WorkItemDefinition) String() string {
	if wid == nil {
		return "<nil WorkItemDefinition>"
	}
	return fmt.Sprintf("WorkItemDef '%s' (ID: %s): Desc: %.30s..., CriteriaKeys: %d, SchemaKeys: %d, DS: %d",
		wid.Name, wid.WorkItemDefinitionID, wid.Description, len(wid.DefaultTargetWorkerCriteria),
		len(wid.DefaultPayloadSchema), len(wid.DefaultDataSourceRefs))
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

func (wi *WorkItem) String() string {
	if wi == nil {
		return "<nil WorkItem>"
	}
	return fmt.Sprintf("WorkItem TaskID: %s (DefName: %s, Queue: %s)\n  Status: %s, Prio: %d, Retries: %d\n  PayloadKeys: %d, DS: %d\n  Error: %.30s...",
		wi.TaskID, wi.WorkItemDefinitionName, wi.QueueName, wi.Status, wi.Priority, wi.RetryCount,
		len(wi.Payload), len(wi.DataSourceRefs), wi.Error)
}

// LLMCallMetrics holds detailed metrics from a specific LLM API call.
type LLMCallMetrics struct {
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	TotalTokens  int64  `json:"total_tokens"`
	FinishReason string `json:"finish_reason,omitempty"`
	ModelUsed    string `json:"model_used,omitempty"`
}

func (lcm *LLMCallMetrics) String() string {
	if lcm == nil {
		return "<nil LLMCallMetrics>"
	}
	return fmt.Sprintf("Model: %s, In: %d, Out: %d, Total: %d, Reason: %s",
		lcm.ModelUsed, lcm.InputTokens, lcm.OutputTokens, lcm.TotalTokens, lcm.FinishReason)
}

// Ensure nlines is correct after modifications
// Original nlines was not specified for this file in this turn, but the file version is 0.3.8
// The previous version for ai_worker_types.go was:
// // NeuroScript Version: 0.4.0
// // File version: 0.3.8 // Corrected JSON tag for InteractionModels
// // Description: Defines types for the AI Worker Management system, including workers, data sources, pools, queues, and work items.
// // filename: pkg/core/ai_worker_types.go
// The file content had 453 lines from the `uploaded:core/ai_worker_types.go` prompt.
// I've added quite a few String() methods. Let's estimate the new line count.
// Added String() methods for:
// AIWorkerDefinitionDisplayInfo (~5 lines)
// APIKeySource (~8 lines)
// RateLimitPolicy (~6 lines)
// TokenUsageMetrics (~6 lines)
// SupervisorFeedback (~6 lines)
// AIWorkerPerformanceSummary (~8 lines)
// GlobalDataSourceDefinition (~9 lines)
// AIWorkerDefinition (~17 lines)
// AIWorkerInstance (~16 lines)
// PerformanceRecord (~9 lines)
// RetiredInstanceInfo (~7 lines)
// InstanceRetirementPolicy (~6 lines)
// AIWorkerPoolDefinition (~8 lines)
// RetryPolicy (~6 lines)
// WorkQueueDefinition (~8 lines)
// WorkItemDefinition (~8 lines)
// WorkItem (~9 lines)
// LLMCallMetrics (~7 lines)
// Total ~150 new lines of code (including func signatures, newlines, fmt.Sprintf, etc.)
// New nlines approx 453 + 150 = 603.
// filename is pkg/core/ai_worker_types.go
// Risk rating LOW-MEDIUM for adding String methods (they don't change core logic but are used for display).
