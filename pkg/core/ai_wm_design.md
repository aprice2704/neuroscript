# AI Worker Management (ai_wm_*) System Design

## 1. Overview

The AI Worker Management system (ai_wm_*) within the pkg/core package of NeuroScript provides a comprehensive framework for defining, managing, executing, and monitoring AI-powered workers. These workers typically represent Large Language Models (LLMs) or other model-based agents. The system is designed to support both stateful, instance-based interactions (e.g., for ongoing conversations) and stateless, one-shot task executions. Key features include persistent worker definitions, API key management, configurable rate limiting, detailed performance tracking with aggregation, and exposure of its functionalities to NeuroScript via a dedicated toolset.

## 2. Core Components and Concepts

### 2.1. AIWorkerManager (ai_wm.go)

**Role**: This is the central orchestrator and entry point for the entire AI Worker Management system. It is responsible for the lifecycle management of worker definitions and active instances, handling persistence of configurations and performance data, enforcing rate limits, and resolving API keys.

**State**: The manager maintains in-memory collections of:
- AIWorkerDefinitions (blueprints for workers).
- Active AIWorkerInstances (live, stateful worker sessions).
- WorkerRateTrackers (runtime counters for rate limiting per definition).

**Concurrency**: It utilizes a sync.RWMutex to ensure thread-safe access and modification of its internal state, crucial for concurrent operations initiated by NeuroScript tools or other parts of the system.

**Persistence**: The manager is responsible for loading and saving:
- Worker definitions (including their aggregated performance summaries) to/from `ai_worker_definitions.json`.
- Performance data of retired instances to/from `ai_worker_performance_data.json`.
These files are stored within a specified sandbox directory, ensuring data isolation and portability.

**Initialization**: The `NewAIWorkerManager` constructor initializes the manager, attempts to load existing definitions and historical performance data from the sandbox, and sets up initial rate trackers for each loaded definition.

### 2.2. AIWorkerDefinition (ai_worker_types.go, ai_wm_definitions.go)

**Role**: Serves as a static blueprint or template that defines all configurable aspects of a particular type of AI worker. It dictates how instances of this worker behave and how stateless calls using this definition are processed.

**Key Attributes**:
- **DefinitionID** (string): A unique identifier (UUID) for the definition.
- **Name** (string): A human-readable name for easy identification.
- **Provider** (AIWorkerProvider): Specifies the source or vendor of the AI model (e.g., "google", "openai", "ollama").
- **ModelName** (string): The specific model identifier from the provider (e.g., "gemini-1.5-pro-latest", "gpt-4o").
- **Auth** (APIKeySource): Defines the method and necessary value (e.g., environment variable name) for retrieving the API key required by the provider.
- **InteractionModels** ([]InteractionModelType): An array specifying the supported modes of interaction (e.g., conversational, stateless_task, or both). Defaults to conversational if empty.
- **Capabilities** ([]string): A list of strings describing the functional capabilities of the worker (e.g., "text_generation", "code_completion", "image_analysis").
- **BaseConfig** (map[string]interface{}): A set of default configuration parameters passed to the LLM during calls (e.g., temperature, top_p, max output tokens).
- **CostMetrics** (map[string]float64): Defines the cost associated with using the worker, typically per token (input/output) or per API call.
- **RateLimits** (RateLimitPolicy): Specifies usage limits such as maximum requests per minute/day, maximum tokens per minute/day, and maximum concurrent active instances.
- **Status** (AIWorkerDefinitionStatus): Indicates the operational status of the definition (active, disabled, archived). Only active definitions can spawn new instances.
- **DefaultFileContexts** ([]string): A list of file paths or URIs that are implicitly included in the context for new instances spawned from this definition.
- **AggregatePerformanceSummary** (AIWorkerPerformanceSummary): Stores aggregated performance metrics derived from all tasks executed using this definition (both instance-based and stateless).
- **Metadata** (map[string]interface{}): A flexible key-value store for any additional, custom information related to the definition.

**Management**: The AIWorkerManager provides a suite of methods for CRUD (Create, Read, Update, Delete) operations on definitions. All changes are persisted to the JSON file.

### 2.3. AIWorkerInstance (ai_worker_types.go, ai_wm_instances.go)

**Role**: Represents a live, stateful session with an AI worker, typically for conversational interactions. Each instance is derived from an AIWorkerDefinition.

**Key Attributes**:
- **InstanceID** (string): A unique identifier (UUID) for the active session.
- **DefinitionID** (string): Links the instance back to its parent AIWorkerDefinition.
- **Status** (AIWorkerInstanceStatus): Reflects the current operational state of the instance (e.g., initializing, idle, busy, retired_completed, error).
- **ConversationHistory** ([]*ConversationTurn): This field is tagged `json:"-"` meaning it's not directly persisted with the instance metadata. It's managed at runtime, typically by an associated ConversationManager within the NeuroScript Interpreter.
- **CreationTimestamp, LastActivityTimestamp** (time.Time): Track the instance's lifecycle.
- **SessionTokenUsage** (TokenUsageMetrics): Accumulates token consumption (input, output, total) for the current active session of this instance.
- **CurrentConfig** (map[string]interface{}): The effective configuration for this instance, derived from the definition's BaseConfig and any overrides applied at spawn time.
- **ActiveFileContexts** ([]string): Runtime-only list of files currently active in the instance's context.
- **LastError** (string), **RetirementReason** (string): Store information about errors or why an instance was retired.

**Lifecycle**:
- **Spawning**: Instances are created via `AIWorkerManager.SpawnWorkerInstance()`. This process checks the definition's status and rate limits (specifically MaxConcurrentActiveInstances).
- **Retirement**: Instances are moved from the active pool upon completion, error, exhaustion (e.g., context window full), or explicit command. `AIWorkerManager.RetireWorkerInstance()` handles this, persisting the instance's final metadata, session token usage, and all associated PerformanceRecords into RetiredInstanceInfo.

**Association with ConversationManager**: For conversational interactions, an AIWorkerInstance is typically paired with a ConversationManager instance within the NeuroScript Interpreter. The ConversationManager handle is returned by the spawn tool, allowing the script to manage the dialogue.

### 2.4. Stateless Task Execution (ai_wm_stateless.go)

**Role**: Provides a mechanism for executing one-shot tasks using an AIWorkerDefinition without the overhead of creating, managing, and retiring a full AIWorkerInstance. This is suitable for tasks that don't require conversational memory.

**Method**: `AIWorkerManager.ExecuteStatelessTask()`.

**Process**:
- Validates that the specified AIWorkerDefinition exists, is active, and supports stateless interaction.
- Checks rate limits defined in the RateLimitPolicy of the definition (requests/tokens per minute/day).
- Resolves the API key using the definition's Auth configuration.
- Makes a direct call to the underlying LLM via the LLMClient provided to the manager (typically from the Interpreter). A temporary, minimal conversation history is constructed for the call.
- Calculates the cost of the operation based on CostMetrics in the definition and token usage from the LLM call.
- Creates and logs a PerformanceRecord for the stateless task. This updates the AggregatePerformanceSummary on the parent AIWorkerDefinition.
- Updates the relevant rate limit counters in the WorkerRateTracker for the definition.
- Returns the model's output and the generated PerformanceRecord.

### 2.5. Performance Tracking (ai_worker_types.go, ai_wm_performance.go)

**PerformanceRecord**: A detailed log entry for each individual task performed by either a stateful instance or a stateless call. It captures:
- **Identifiers**: TaskID, InstanceID (can be a special "stateless-\<uuid\>" for stateless tasks), DefinitionID.
- **Timings**: TimestampStart, TimestampEnd, DurationMs.
- **Outcome**: Success (bool), ErrorDetails (string).
- **Context**: InputContext (map for arbitrary input details).
- **LLM Metrics**: LLMMetrics (map for raw metrics from LLM like tokens, finish reason).
- **Cost**: CostIncurred (float64).
- **Output**: OutputSummary (string, potentially trimmed).
- **Feedback**: SupervisorFeedback (optional, for quality ratings/comments).

**AIWorkerPerformanceSummary**: Embedded within each AIWorkerDefinition. It aggregates data from all PerformanceRecords associated with that definition to provide long-term statistics:
- **Counts**: TotalTasksAttempted, SuccessfulTasks, FailedTasks, TotalInstancesSpawned.
- **Averages**: AverageSuccessRate, AverageDurationMs, AverageQualityScore.
- **Usage**: TotalTokensProcessed, TotalCostIncurred.
- **Timestamps**: LastActivityTimestamp.
- **Runtime**: ActiveInstancesCount (reflects current active instances, updated dynamically).

**Logging Process**: `AIWorkerManager.logPerformanceRecordUnsafe()` is an internal method that takes a PerformanceRecord and updates the AggregatePerformanceSummary of the corresponding AIWorkerDefinition.

**Persistence**:
- PerformanceRecords for stateful instances are grouped under RetiredInstanceInfo and saved to `ai_worker_performance_data.json` when an instance is retired.
- PerformanceRecords for stateless tasks are also logged and contribute to the definition's summary, which is saved in `ai_worker_definitions.json`. The individual stateless performance records are effectively persisted via the `logPerformanceRecordUnsafe` mechanism if they are included in a RetiredInstanceInfo structure designed for them or if the design implies they are written to the performance log file through another path (currently, they are logged to the definition's summary, and the record itself is returned by ExecuteStatelessTask).

### 2.6. Rate Limiting (ai_worker_types.go, ai_wm_ratelimit.go)

**RateLimitPolicy**: A struct embedded in AIWorkerDefinition. It specifies:
- MaxRequestsPerMinute
- MaxTokensPerMinute (input + output)
- MaxTokensPerDay (input + output)
- MaxConcurrentActiveInstances

**WorkerRateTracker**: An in-memory struct, managed by AIWorkerManager (one per definition). It dynamically tracks current usage against the policy limits using counters and time markers for minute/day windows.

**Enforcement**:
- MaxConcurrentActiveInstances is checked by `SpawnWorkerInstance`.
- Request and token limits (per minute/day) are checked by `ExecuteStatelessTask` and potentially by instance-based operations before making LLM calls.
- The AIWorkerManager includes methods like `checkAndRecordUsageUnsafe` (primarily a check) and `recordUsageUnsafe` (to update counters after a call) which are used internally. `updateTokenCountForRateLimitsUnsafe` is another helper for token updates.

### 2.7. API Key Management (ai_worker_types.go, ai_wm.go)

**APIKeySource**: A struct within AIWorkerDefinition that specifies how the API key for the worker should be obtained.

**APIKeySourceMethod**: An enum defining the retrieval strategy:
- **APIKeyMethodEnvVar**: Key is read from an environment variable (name specified in `APIKeySource.Value`).
- **APIKeyMethodInline**: Key is provided directly in `APIKeySource.Value` (use with extreme caution due to security implications of storing keys in configuration files).
- **APIKeyMethodNone**: No API key is required (e.g., for local models not needing auth).
- **APIKeyMethodConfigPath**, **APIKeyMethodVault**: Planned for future implementation, for more secure key retrieval from dedicated config files or secret vaults.

**Resolution**: `AIWorkerManager.resolveAPIKey()` is an internal method that takes an APIKeySource and returns the actual API key string.

## 3. Data Structures and Types (Key Types from ai_worker_types.go)

- **AIWorkerProvider**: Enum (google, openai, anthropic, ollama, local, custom) identifying the LLM provider.
- **InteractionModelType**: Enum (conversational, stateless_task, both) defining how a worker definition is primarily used.
- **AIWorkerDefinitionStatus**: Enum (active, disabled, archived) for the lifecycle state of a definition.
- **AIWorkerInstanceStatus**: Enum (e.g., initializing, idle, busy, retired_completed, error) for the runtime state of an instance.
- **TokenUsageMetrics**: Struct (InputTokens, OutputTokens, TotalTokens) for tracking token consumption.
- **SupervisorFeedback**: Struct (Rating, Comments, etc.) for capturing qualitative feedback on task outputs.
- **LLMCallMetrics**: An internal struct used primarily during stateless execution to temporarily hold detailed metrics from an LLM API call before they are incorporated into a PerformanceRecord.
- **RetiredInstanceInfo**: A data transfer object used for persisting the final state and all associated PerformanceRecords of an instance once it's retired.

## 4. Persistence Strategy

**Format**: JSON is used for serializing data to files.

**Files**:
- `ai_worker_definitions.json`: Stores an array of AIWorkerDefinition objects. This file includes the AggregatePerformanceSummary for each definition, which is updated as tasks are completed.
- `ai_worker_performance_data.json`: Stores an array of RetiredInstanceInfo objects. Each RetiredInstanceInfo contains a list of PerformanceRecords specific to that retired instance.

**Location**: Both primary data files are located in the `sandboxDir` configured for the AIWorkerManager. This allows for environment-specific data sets.

**Loading**: Upon initialization, AIWorkerManager attempts to load data from these files. `loadWorkerDefinitionsFromFileInternal` handles definitions, and `loadRetiredInstancePerformanceDataInternal` processes the performance data to update the in-memory aggregate summaries on the loaded definitions.

**Saving**:
- Definitions are saved (via `saveDefinitionsToFileUnsafe`) whenever a definition is added, updated, or removed, and also after a stateless task logs performance (as this updates the definition's summary).
- Retired instance data (including all its performance records) is appended to `ai_worker_performance_data.json` (via `appendRetiredInstanceToFileUnsafe`) when an AIWorkerInstance is retired. This method reads the existing file, appends/updates the new record, and writes the entire collection back.

## 5. NeuroScript Tool Integration (ai_wm_tools.go)

The AI Worker Management system's capabilities are exposed to NeuroScript programs through a comprehensive set of tools, allowing scripts to dynamically manage and utilize AI workers.

**Tool Categories**:
- **Definition Management**: Tools like `AIWorkerDefinition.Add`, `AIWorkerDefinition.Get`, `AIWorkerDefinition.List`, `AIWorkerDefinition.Update`, `AIWorkerDefinition.Remove`, `AIWorkerDefinition.LoadAll` (reloads from file), and `AIWorkerDefinition.SaveAll` (persists to file).
- **Instance Management**: Tools like `AIWorkerInstance.Spawn` (creates a new instance and returns a handle to its ConversationManager), `AIWorkerInstance.Get`, `AIWorkerInstance.ListActive`, `AIWorkerInstance.Retire`, `AIWorkerInstance.UpdateStatus`, and `AIWorkerInstance.UpdateTokenUsage`.
- **Stateless Execution**: The `AIWorker.ExecuteStatelessTask` tool allows direct, one-off calls using a definition.
- **Performance & Logging**: Tools like `AIWorker.LogPerformance` (for explicitly logging a performance record, though often handled internally), `AIWorker.GetPerformanceRecords` (to query persisted records for a definition), `AIWorker.SavePerformanceData` (intended for manual persistence of the raw performance log, though current implementation notes a dependency on a specific manager method), and `AIWorker.LoadPerformanceData` (reloads definitions which in turn re-processes performance summaries).

**Registration**: The `RegisterAIWorkerTools(*Interpreter)` function is responsible for registering all these tools with the NeuroScript interpreter's ToolRegistry. During this process, it also ensures that an AIWorkerManager instance is initialized and associated with the interpreter (via the `Interpreter.aiWorkerManager` field).

**Argument Handling**: Tool functions utilize `ValidateAndConvertArgs` (from `tools_validation.go`) for robust argument checking and type coercion based on their ToolSpec. A helper `mapValidatedArgsListToMapByName` is used internally by tools to convert the validated argument list (slice) into a map keyed by argument names for more convenient access.

## 6. Error Handling

The system employs a structured error handling approach using the custom `RuntimeError` type (defined in `pkg/core/errors.go`).

`RuntimeError` encapsulates an `ErrorCode` (an integer code for categorizing errors), a descriptive message, and an optional wrapped underlying error (for context preservation, adhering to Go's error wrapping conventions).

A set of predefined `ErrorCode` constants (e.g., `ErrorCodeKeyNotFound`, `ErrorCodeRateLimited`, `ErrorCodePreconditionFailed`, `ErrorCodeArgMismatch`, `ErrorCodeInternal`) and corresponding sentinel error variables (e.g., `ErrNotFound`, `ErrRateLimited`, `ErrInvalidArgument`) are used throughout the ai_wm_* modules to provide specific and checkable error conditions.

## 7. Key Design Principles and Considerations

- **Modularity and Separation of Concerns**:
  - Clear distinction between the static `AIWorkerDefinition` (the blueprint) and the dynamic `AIWorkerInstance` (the runtime entity).
  - The `AIWorkerManager` acts as a facade and central point of control, encapsulating complex logic.
  - Functionality is broken down into logical Go files: core types (`ai_worker_types.go`), main manager (`ai_wm.go`), and separate files for managing definitions (`ai_wm_definitions.go`), instances (`ai_wm_instances.go`), stateless calls (`ai_wm_stateless.go`), performance (`ai_wm_performance.go`), rate limiting (`ai_wm_ratelimit.go`), and NeuroScript tools (`ai_wm_tools.go`).
- **State Management and Persistence**:
  - Critical configuration (definitions) and historical data (performance) are persisted, allowing the system's state to survive across application restarts.
  - In-memory caches and trackers (for active instances, rate limits) provide efficient runtime operation.
- **Extensibility**:
  - The use of enums/types like `AIWorkerProvider`, `APIKeySourceMethod`, and `InteractionModelType` allows for future expansion with new LLM providers or interaction paradigms.
  - `BaseConfig` and `Metadata` fields (maps) in `AIWorkerDefinition` offer flexibility for provider-specific settings and custom annotations.
- **Monitoring and Control**:
  - Integrated performance tracking (per-task records and aggregated summaries) enables monitoring of worker effectiveness and cost.
  - Rate limiting provides essential control over API usage to manage costs and adhere to provider limits.
- **Integration with NeuroScript**:
  - The system is designed to be primarily controlled and utilized from within NeuroScript scripts via the `ai_wm_tools.go` toolset.
  - Handles (e.g., for ConversationManager) are used to bridge NeuroScript's execution environment with Go objects managed by the AI Worker system.
- **Concurrency Safety**: The `AIWorkerManager` uses `sync.RWMutex` to protect its shared internal state, making it safe for use in concurrent environments where multiple NeuroScript routines might interact with it.