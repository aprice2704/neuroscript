 # AI Worker Management (ai_wm_*) System Design

 ## 1. Overview

 The AI Worker Management system (ai_wm_*) within the pkg/core package of NeuroScript provides a comprehensive framework for defining, managing, executing, and monitoring AI-powered workers. These workers typically represent Large Language Models (LLMs) or other model-based agents. The system is designed to support both stateful, instance-based interactions and stateless, one-shot task executions, potentially managed through worker pools and work queues. Key features include persistent worker definitions; flexible and shared data source configurations with controlled external file access capabilities; definitions for templatizing work items (WorkItemDefinition); API key management; worker-specific tool allow/deny lists; configurable rate limiting; detailed performance tracking; and provisions for Supervisory AI (SAI) attachment for monitoring and feedback. The system exposes its functionalities to NeuroScript via a dedicated toolset.

 ## 2. Core Components and Concepts

 ### 2.1. AIWorkerManager (ai_wm.go)

 **Role**: This is the central orchestrator and entry point for the entire AI Worker Management system. It is responsible for the lifecycle management of:
 -   `AIWorkerDefinition`s (blueprints for workers).
 -   `GlobalDataSourceDefinition`s: Centrally defined data sources.
 -   `AIWorkerPoolDefinition`s and active `AIWorkerPool`s: For managing groups of worker instances.
 -   `WorkQueueDefinition`s and active `WorkQueue`s: For managing task submission and dispatch.
 -   `WorkItemDefinition`s: For templatizing tasks.
 -   Active `AIWorkerInstance`s (live, stateful worker sessions, potentially managed by pools).

 It handles persistence of these configurations, performance data, enforces rate limits, resolves API keys, and dispatches tasks from queues to pools.

 **Configuration**:
 -   The `AIWorkerManager` itself will have configurable behaviors, potentially set during initialization or via specific NeuroScript tools:
     -   **`ConfigLoadPolicy` (enum, e.g., `FailFastOnError`, `LoadValidAndReportErrors`)**: Determines how the `LoadConfigBundleFromString` tool behaves.
         -   `FailFastOnError`: If any definition in a bundle fails validation, the entire bundle loading operation is aborted, and no changes are activated.
         -   `LoadValidAndReportErrors`: Activates all valid definitions from the bundle and reports errors for those that failed validation. This is generally more user-friendly for iterative setup.
     -   Whitelist of allowed base paths for `GlobalDataSourceDefinition.LocalPath` when `AllowExternalReadAccess` is true. This is a critical security setting configured at the system/deployment level, not typically via dynamic NeuroScript calls.

 **State**: The manager maintains in-memory collections of:
 -   `AIWorkerDefinition`s.
 -   Active `AIWorkerInstance`s.
 -   `WorkerRateTracker`s (runtime counters for rate limiting per definition).
 -   `GlobalDataSourceDefinition`s.
 -   `AIWorkerPoolDefinition`s and runtime `AIWorkerPool` states.
 -   `WorkQueueDefinition`s and runtime `WorkQueue` states (including `WorkItem`s if queues are in-memory).
 -   `WorkItemDefinition`s.
 -   **(Future)** References to active Supervisory AI (SAI) instances or configurations.

 **Concurrency**: It utilizes a `sync.RWMutex` to ensure thread-safe access and modification of its internal state.

 **Persistence**:
 The manager is responsible for loading and saving:
 -   Worker definitions (including their aggregated performance summaries and references) to/from `ai_worker_definitions.json`.
 -   Performance data of retired instances to/from `ai_worker_performance_data.json`.
 -   Global Data Source definitions to/from `ai_global_data_source_definitions.json`.
 -   `AIWorkerPoolDefinition`s to/from `ai_worker_pool_definitions.json`.
 -   `WorkQueueDefinition`s to/from `ai_work_queue_definitions.json`.
 -   `WorkItemDefinition`s to/from `ai_work_item_definitions.json`.
 -   **(Future)** `WorkItem`s if queue persistence is enabled (e.g., to `ai_work_items.db`).
 These files are stored within a specified sandbox directory.

 **Initialization**:
 The `NewAIWorkerManager` constructor initializes the manager, loads its own operational configuration (like `ConfigLoadPolicy`), attempts to load all defined configurations (worker definitions, data sources, work item definitions, pools, queues) and historical performance data from the sandbox, and sets up initial rate trackers for each loaded worker definition.

 ---
 ### 2.2. Global Data Source Management (ai_wm_datasources.go - *new conceptual file*)

 This component, managed by the `AIWorkerManager`, handles the definition and persistence of shared data sources that AI Workers can reference.

 **GlobalDataSourceDefinition (`ai_worker_types.go`)**:
 Serves as a template for a data access area. Each definition includes:
 -   **`Name`** (string): A unique, human-readable name/ID for this data source (e.g., "project_schematics", "shared_part_files"). This name is used by other definitions to reference it.
 -   **`Type`** (DataSourceType): Indicates the nature of the data source (e.g., `DataSourceTypeLocalDirectory`, `DataSourceTypeFileAPI`).
 -   **`Description`** (string, optional): A brief description of the data source.
 -   **`LocalPath`** (string, optional): For `DataSourceTypeLocalDirectory`, this is the absolute path on the local filesystem.
 -   **`AllowExternalReadAccess`** (bool, optional): For `DataSourceTypeLocalDirectory`, if true, `LocalPath` can be outside the primary interpreter sandbox. **This requires strict validation by `AIWorkerManager` against a system-administrator-defined list of permissible external base paths.**
 -   **`FileAPIPath`** (string, optional): For `DataSourceTypeFileAPI`, this is the path within the Google File API (e.g., "fm:/shared/designs/").
 -   **`RemoteTargetPath`** (string, optional): For `DataSourceTypeLocalDirectory`, suggests a default target path in the remote File API if this source is synced (e.g., "synced_data_sources/<DataSourceName>").
 -   **`ReadOnly`** (bool): If true, the worker should treat this source as read-only. Critically, NeuroScript file *write* tools (`toolWriteFile`, etc.) **always** operate strictly within the primary interpreter sandbox, regardless of this flag for external sources.
 -   **`Filters`** ([]string, optional): Glob patterns for files to include from this source.
 -   **`Recursive`** (bool, optional): Whether to consider files recursively from this source.
 -   **`Metadata`** (map[string]interface{}, optional): Additional custom information.
 -   **`CreatedTimestamp`** (time.Time, omitempty): Set by manager on creation.
 -   **`ModifiedTimestamp`** (time.Time, omitempty): Set by manager on update.

 **Management**: The `AIWorkerManager` will provide CRUD operations for `GlobalDataSourceDefinition`s. All changes will be persisted to `ai_global_data_source_definitions.json`.

 **Security**: The `AIWorkerManager`, upon loading or adding a `GlobalDataSourceDefinition` of type `DataSourceTypeLocalDirectory` with `AllowExternalReadAccess` set to true, **must** validate its `LocalPath` against a pre-configured whitelist of allowed external base directories. This whitelist is a system-level configuration, not part of the NeuroScript-modifiable definitions. If validation fails, the data source definition should be rejected or marked as unusable.
 ---

 ### 2.3. AIWorkerDefinition (ai_worker_types.go, ai_wm_definitions.go)
 **Role**: Serves as a static blueprint or template that defines all configurable aspects of a particular type of AI worker.

 **Key Attributes**:
 -   **`DefinitionID`** (string): System-generated UUID, primary key for storage.
 -   **`Name`** (string): User-provided, unique, human-readable name.
 -   **`Provider`** (AIWorkerProvider): Specifies the source or vendor of the AI model.
 -   **`ModelName`** (string): The specific model identifier from the provider.
 -   **`Auth`** (APIKeySource): Defines the method and value for retrieving the API key.
 -   **`InteractionModels`** ([]InteractionModelType, omitempty): Supported modes of interaction. Defaults to `["conversational"]` if empty.
 -   **`Capabilities`** ([]string, omitempty): Functional capabilities of the worker (e.g., "code_generation", "panel_design_analysis").
 -   **`BaseConfig`** (map[string]interface{}, omitempty): Default LLM configuration parameters (e.g., temperature, max_output_tokens).
 -   **`CostMetrics`** (map[string]float64, omitempty): Cost associated with using the worker (e.g., {"input_tokens_usd_per_1k": 0.001}).
 -   **`RateLimits`** (RateLimitPolicy, omitempty): Usage limits.
 -   **`Status`** (AIWorkerDefinitionStatus, omitempty): Operational status (active, disabled, archived). Default: "active".
 -   **`DefaultFileContexts`** ([]string, omitempty): List of file paths or URIs (e.g., `datasource://project_assets/file.txt`, `fm:/permanent/context.md`) implicitly included in the context for new instances.
 -   **`DataSourceRefs`** ([]string, omitempty): An array of names/IDs referencing `GlobalDataSourceDefinition`s. This provides a baseline set of data sources for this worker type.
 -   **`ToolAllowlist`** ([]string, omitempty): Worker-definition-specific tool allowlist (qualified tool names).
 -   **`ToolDenylist`** ([]string, omitempty): Worker-definition-specific tool denylist (qualified tool names, overrides allowlist).
 -   **`DefaultSupervisoryAIRef`** (string, optional, Future): References the name/ID of an `AIWorkerDefinition` suitable for supervising tasks run by this worker type, or a pre-configured SAI instance.
 -   **`AggregatePerformanceSummary`** (*AIWorkerPerformanceSummary, omitempty): Aggregated performance metrics, managed by the system.
 -   **`Metadata`** (map[string]interface{}, omitempty): Flexible key-value store for additional custom information.
 -   **`CreatedTimestamp`** (time.Time, omitempty): Set by manager.
 -   **`ModifiedTimestamp`** (time.Time, omitempty): Set by manager.

 **Management**: The `AIWorkerManager` provides CRUD operations on definitions. Changes are persisted to `ai_worker_definitions.json`.

 ---
 ### 2.4. AIWorkerPools (ai_wm_pools.go - *new conceptual file*)

 **AIWorkerPoolDefinition (`ai_worker_types.go`)**:
 A persistent configuration for a worker pool.
 -   **`PoolID`** (string): System-generated UUID.
 -   **`Name`** (string): User-defined, unique human-readable name.
 -   **`TargetAIWorkerDefinitionName`** (string): Name of the `AIWorkerDefinition` used to spawn instances in this pool.
 -   **`MinIdleInstances`** (int, omitempty): Desired minimum number of idle instances.
 -   **`MaxTotalInstances`** (int, omitempty): Maximum total instances allowed in this pool (also subject to global `AIWorkerManager` limits and definition rate limits).
 -   **`InstanceRetirementPolicy`** (InstanceRetirementPolicy, omitempty): Policy for when to retire instances (e.g., `MaxTasksPerInstance`, `MaxInstanceAgeHours`).
 -   **`DataSourceRefs`** ([]string, omitempty): References to `GlobalDataSourceDefinition`s that are common to all workers in this pool.
 -   **`SupervisoryAIRef`** (string, omitempty, Future): Reference to an SAI (AIWorkerDefinition name) for this pool.
 -   **`IsMissionCritical`** (bool, omitempty): If true, indicates that this pool's operational status is critical.
 -   **`Metadata`** (map[string]interface{}, omitempty).
 -   **`CreatedTimestamp`** (time.Time, omitempty): Set by manager.
 -   **`ModifiedTimestamp`** (time.Time, omitempty): Set by manager.

 **AIWorkerPool (Runtime Structure)**:
 Managed by `AIWorkerManager`.
 -   Tracks its `AIWorkerPoolDefinition`.
 -   Maintains a list of active `AIWorkerInstance` IDs belonging to the pool and their statuses.
 -   Handles scaling logic (spawning/retiring instances based on demand and definition).
 -   Provides idle instances to the Task Dispatcher.

 ---
 ### 2.5. WorkItemDefinition (ai_worker_types.go - new type)

 **Role**: Serves as a persistent template or blueprint for creating similar `WorkItem`s, reducing redundancy and ensuring consistency when submitting common types of tasks to a `WorkQueue`.

 **WorkItemDefinition (`ai_worker_types.go`)**:
 -   **`WorkItemDefinitionID`** (string): System-generated UUID.
 -   **`Name`** (string): A user-defined, unique, human-readable name for this template (e.g., "AnalyzePanelStress", "GenerateToolpathVariant").
 -   **`Description`** (string, omitempty): A brief description of the purpose of tasks created from this definition.
 -   **`DefaultTargetWorkerCriteria`** (map[string]interface{}, omitempty): Pre-defined criteria for selecting a worker or pool (e.g., `{"definitionName": "panel_analyzer_v2", "capabilities": ["stress_analysis"]}`). These can be overridden at submission time.
 -   **`DefaultPayloadSchema`** (map[string]interface{}, omitempty): A JSON schema map or a map of default values defining the expected structure/defaults for the `WorkItem.Payload`. (e.g., `{"input_file_uri": {"type": "string"}, "material": {"type": "string", "default": "steel"}}`).
 -   **`DefaultDataSourceRefs`** ([]string, omitempty): A list of `GlobalDataSourceDefinition` names typically required for tasks of this type.
 -   **`DefaultPriority`** (int, omitempty): A default priority for `WorkItem`s created from this template.
 -   **`DefaultSupervisoryAIRef`** (string, omitempty, Future): Default SAI reference (AIWorkerDefinition name) for tasks of this type.
 -   **`Metadata`** (map[string]interface{}, omitempty).
 -   **`CreatedTimestamp`** (time.Time, omitempty): Set by manager.
 -   **`ModifiedTimestamp`** (time.Time, omitempty): Set by manager.

 **Management**: The `AIWorkerManager` will provide CRUD operations for `WorkItemDefinition`s. Changes persisted to `ai_work_item_definitions.json`.

 **Usage**: When submitting a task, a `WorkItemDefinitionName` can be provided. The system uses this as a base, and explicitly provided `WorkItem` parameters override or merge with the defaults.

 ---
 ### 2.6. WorkQueues & WorkItems (ai_wm_queues.go - *new conceptual file*)

 **WorkQueueDefinition (`ai_worker_types.go`)**:
 A persistent configuration for a work queue.
 -   **`QueueID`** (string): System-generated UUID.
 -   **`Name`** (string): User-defined, unique human-readable name.
 -   **`AssociatedPoolNames`** ([]string): Names of `AIWorkerPoolDefinition`(s) that service this queue.
 -   **`DefaultPriority`** (int, omitempty): Default priority for tasks if not specified in the `WorkItem`.
 -   **`RetryPolicy`** (RetryPolicy, omitempty): Defines retry behavior (e.g., `MaxRetries`, `RetryDelaySeconds`).
 -   **`PersistTasks`** (bool, omitempty): If true, `WorkItem`s in this queue should be persisted to survive restarts (Future).
 -   **`DataSourceRefs`** ([]string, omitempty): References to `GlobalDataSourceDefinition`s relevant to all tasks in this queue.
 -   **`SupervisoryAIRef`** (string, omitempty, Future): Reference to an SAI (AIWorkerDefinition name) for this queue.
 -   **`IsMissionCritical`** (bool, omitempty): If true, indicates criticality.
 -   **`Metadata`** (map[string]interface{}, omitempty).
 -   **`CreatedTimestamp`** (time.Time, omitempty): Set by manager.
 -   **`ModifiedTimestamp`** (time.Time, omitempty): Set by manager.

 **WorkItem (`ai_worker_types.go`)**:
 Represents a unit of work submitted to a queue.
 -   **`TaskID`** (string): Unique identifier (system-generated if not provided).
 -   **`WorkItemDefinitionName`** (string, omitempty): Optional name of the `WorkItemDefinition` used as a template.
 -   **`QueueName`** (string): Name of the `WorkQueueDefinition` this item was submitted to.
 -   **`TargetWorkerCriteria`** (map[string]interface{}, omitempty): Overrides criteria from `WorkItemDefinition` or `WorkQueueDefinition`.
 -   **`Payload`** (map[string]interface{}): Task-specific data. Merged with/validated by `WorkItemDefinition`'s schema if applicable.
 -   **`DataSourceRefs`** ([]string, omitempty): Specific `GlobalDataSourceDefinition` names for this item, augmenting other contexts.
 -   **`Priority`** (int, omitempty): Overrides default priorities.
 -   **`Status`** (WorkItemStatus, omitempty): Current status (e.g., `Pending`, `Processing`). Set by system.
 -   **`SubmitTimestamp`** (time.Time, omitempty): Set by system on submission.
 -   **`StartTimestamp`** (time.Time, omitempty): Set by system when processing begins.
 -   **`EndTimestamp`** (time.Time, omitempty): Set by system on completion or final failure.
 -   **`RetryCount`** (int, omitempty): Managed by the system based on `RetryPolicy`.
 -   **`Result`** (interface{}, omitempty): Outcome of the task, stored upon successful completion.
 -   **`Error`** (string, omitempty): Error details if the task failed.
 -   **`PerformanceRecordID`** (string, omitempty): Link to the `PerformanceRecord` generated for this task.
 -   **`SupervisoryAIRef`** (string, omitempty, Future): Specific SAI for this work item.
 -   **`Metadata`** (map[string]interface{}, omitempty).

 **WorkQueue (Runtime Structure)**:
 Managed by `AIWorkerManager`. Tracks its `WorkQueueDefinition` and holds pending `WorkItem`s.

 ---
 ### 2.7. AIWorkerInstance (ai_worker_types.go, ai_wm_instances.go)

 **Role**: Represents a live, stateful session with an AI worker. Can be standalone or part of an `AIWorkerPool`.

 **Key Attributes**:
 -   **`InstanceID`** (string): System-generated UUID.
 -   **`DefinitionID`** (string): ID of the `AIWorkerDefinition` it's based on.
 -   **`Status`** (AIWorkerInstanceStatus): Reflects the current operational state.
 -   **`ConversationHistory`** ([]*ConversationTurn): Tagged `json:"-"`. Managed by an associated `ConversationManager`.
 -   **`CreationTimestamp`, `LastActivityTimestamp`** (time.Time).
 -   **`SessionTokenUsage`** (TokenUsageMetrics): Accumulated for this instance's session.
 -   **`CurrentConfig`** (map[string]interface{}, omitempty): Effective configuration.
 -   **`ActiveFileContexts`** ([]string): Runtime-only list of URIs. Tagged `json:"-"`.
 -   **`LastError`** (string, omitempty).
 -   **`RetirementReason`** (string, omitempty).
 -   **`PoolID`** (string, omitempty): If managed by an `AIWorkerPool`.
 -   **`CurrentTaskID`** (string, omitempty): If processing a `WorkItem`.
 -   **`DataSourceRefs`** ([]string, omitempty): Instance-specific dynamically attached `GlobalDataSourceDefinition` names.
 -   **`SupervisoryAIRef`** (string, omitempty, Future): Instance-specific SAI.
 -   **`ResolvedDataSources`** ([]*GlobalDataSourceDefinition): Runtime field, not directly persisted in instance JSON. Derived from all applicable contexts.

 **Lifecycle**: Spawning can be direct or by a pool. Task execution uses context resolution (Section 9). Retirement can be manual or by pool policy.

 ---
 ### 2.8. Task Execution & Dispatching

 **Role**: The system uses queues and pools for robust task execution.

 **Process (Task via Queue)**:
 1.  A task is submitted to a `WorkQueue` (e.g., via `WorkQueue.SubmitTask` tool), potentially referencing a `WorkItemDefinitionName` and providing overrides to create a `WorkItem`.
 2.  The `AIWorkerManager`'s **Task Dispatcher** component:
     a.  Selects a pending `WorkItem` from a queue (considering priority, etc.).
     b.  Identifies a compatible `AIWorkerPool` (based on `WorkItem.TargetWorkerCriteria` or `WorkQueue.AssociatedPoolNames`). Considers `IsMissionCritical` status of pools/queues in decision-making or alerting.
     c.  Requests an idle `AIWorkerInstance` from the pool. The pool manager might spawn a new instance if needed and allowed.
     d.  Resolves the **effective context** for the task:
         i.  **Effective DataSources**: Combines `DataSourceRefs` from `WorkItem` (and its `WorkItemDefinition`), `AIWorkerInstance`, `AIWorkerPoolDefinition`, `WorkQueueDefinition`, and `AIWorkerDefinition` (see Section 9).
         ii. **Effective Tool Permissions**: Derived from instance's `AIWorkerDefinition` (`ToolAllowlist`, `ToolDenylist`).
         iii. **Effective SAI**: Resolved from `WorkItem` (and its `WorkItemDefinition`), `Instance`, `Pool`, `Queue`, `Definition` (see Section 9).
     e.  Assigns the `WorkItem` and its context to the instance. Instance status changes to `Busy`, `CurrentTaskID` is set.
 3.  The `AIWorkerInstance` executes the task.
 4.  Tool calls are validated by `SecurityLayer` against effective tool permissions and use effective data sources.
 5.  Upon completion/failure: A `PerformanceRecord` is generated (linked by `WorkItem.PerformanceRecordID`), `WorkItem` status/result/error updated, instance becomes `Idle` or is retired. (Future) Info fed to SAI.
 6.  Task originator can query `WorkItem` status/result via `TaskID`.

 The original `AIWorkerManager.ExecuteStatelessTask()` method may be refactored as a synchronous convenience wrapper around this queue submission process, using a default or implicit `WorkItemDefinition`.

 ---
 ### 2.9. Performance Tracking (ai_worker_types.go, ai_wm_performance.go)

 `PerformanceRecord`s are generated for each task. If a task originates as a `WorkItem`, the `PerformanceRecord.TaskID` should match `WorkItem.TaskID`. These records contribute to the `AggregatePerformanceSummary` on the `AIWorkerDefinition` used by the executing instance.

 ### 2.10. Rate Limiting (ai_worker_types.go, ai_wm_ratelimit.go)

 `RateLimitPolicy` within `AIWorkerDefinition` continues to govern instances spawned from that definition, whether standalone or by a pool. Pools must respect these limits when scaling.

 ### 2.11. API Key Management (ai_worker_types.go, ai_wm.go)

 API key resolution via `AIWorkerDefinition.Auth` remains unchanged.

 ---
 ## 3. Data Structures and Types (Key Types from `ai_worker_types.go`)

 -   `InteractionModelType`, `APIKeySourceMethod`, `AIWorkerProvider`, `AIWorkerDefinitionStatus`, `AIWorkerInstanceStatus`, `ConfigLoadPolicy` (NEW)
 -   `APIKeySource`, `RateLimitPolicy`, `TokenUsageMetrics`, `SupervisorFeedback`, `AIWorkerPerformanceSummary`
 -   `DataSourceType` (NEW)
 -   `GlobalDataSourceDefinition` (NEW)
 -   `AIWorkerDefinition` (Modified: `DataSourceRefs`, `ToolAllowlist`, `ToolDenylist`, `DefaultSupervisoryAIRef`)
 -   `AIWorkerInstance` (Modified: `PoolID`, `CurrentTaskID`, instance-level `DataSourceRefs` & `SupervisoryAIRef`)
 -   `PerformanceRecord`, `RetiredInstanceInfo`
 -   `InstanceRetirementPolicy` (NEW)
 -   `AIWorkerPoolDefinition` (NEW, includes `IsMissionCritical`)
 -   `RetryPolicy` (NEW)
 -   `WorkQueueDefinition` (NEW, includes `IsMissionCritical`)
 -   `WorkItemDefinition` (NEW)
 -   `WorkItemStatus` (NEW)
 -   `WorkItem` (NEW, includes `WorkItemDefinitionName`)
 -   `AIWorkerManagementConfigBundle` (NEW, for omni-loading)
 -   `LLMCallMetrics`

 ## 4. Persistence Strategy

 **Format**: JSON for definitions. (Future: consider embedded DB for `WorkItem`s if `PersistTasks` is true and volume is high).

 **Files** (in `sandboxDir`):
 -   `ai_worker_definitions.json` (Array of `AIWorkerDefinition`)
 -   `ai_worker_performance_data.json` (Array of `RetiredInstanceInfo`)
 -   `ai_global_data_source_definitions.json` (Array of `GlobalDataSourceDefinition`)
 -   `ai_worker_pool_definitions.json` (Array of `AIWorkerPoolDefinition`)
 -   `ai_work_queue_definitions.json` (Array of `WorkQueueDefinition`)
 -   `ai_work_item_definitions.json` (Array of `WorkItemDefinition`)

 **Loading/Saving**: `AIWorkerManager` handles loading and saving all definition types, respecting its `ConfigLoadPolicy` during bundle loads. Individual `add/update/remove` operations for definitions trigger persistence for that definition type.

 ## 5. NeuroScript Tool Integration (ai_wm_tools.go)

 **Tool Categories**:
 -   **AIWorkerDefinition Management**:
     -   `aiWorkerDefinition.addFromString(jsonString string) (id string, error string)`
     -   `aiWorkerDefinition.get(idOrName string) (definitionMap map, error string)`
     -   `aiWorkerDefinition.list(filterMap? map) (definitionsList []map, error string)`
     -   `aiWorkerDefinition.updateFromString(idOrName string, jsonString string) (id string, error string)`
     -   `aiWorkerDefinition.remove(idOrName string) (error string)`
 -   **Global Data Source Management**:
     -   `aiWorkerGlobalDataSource.addFromString(jsonString string) (name string, error string)`
     -   `aiWorkerGlobalDataSource.get(name string) (dataSourceMap map, error string)`
     -   `aiWorkerGlobalDataSource.list(filterMap? map) (dataSourcesList []map, error string)`
     -   `aiWorkerGlobalDataSource.updateFromString(name string, jsonString string) (name string, error string)`
     -   `aiWorkerGlobalDataSource.remove(name string) (error string)`
 -   **Work Item Definition Management**:
     -   `aiWorkerWorkItemDefinition.addFromString(jsonString string) (id string, error string)`
     -   `aiWorkerWorkItemDefinition.get(idOrName string) (itemDefMap map, error string)`
     -   `aiWorkerWorkItemDefinition.list(filterMap? map) (itemDefsList []map, error string)`
     -   `aiWorkerWorkItemDefinition.updateFromString(idOrName string, jsonString string) (id string, error string)`
     -   `aiWorkerWorkItemDefinition.remove(idOrName string) (error string)`
 -   **Worker Pool Management**:
     -   `aiWorkerPoolDefinition.addFromString(jsonString string) (id string, error string)`
     -   `aiWorkerPoolDefinition.get(idOrName string) (poolDefMap map, error string)`
     -   `aiWorkerPoolDefinition.list(filterMap? map) (poolDefsList []map, error string)`
     -   `aiWorkerPoolDefinition.updateFromString(idOrName string, jsonString string) (id string, error string)`
     -   `aiWorkerPoolDefinition.remove(idOrName string) (error string)`
     -   `aiWorkerPool.getInstanceStatus(poolName string) (statusMap map, error string)`
 -   **Work Queue Management**:
     -   `aiWorkerWorkQueueDefinition.addFromString(jsonString string) (id string, error string)`
     -   `aiWorkerWorkQueueDefinition.get(idOrName string) (queueDefMap map, error string)`
     -   `aiWorkerWorkQueueDefinition.list(filterMap? map) (queueDefsList []map, error string)`
     -   `aiWorkerWorkQueueDefinition.updateFromString(idOrName string, jsonString string) (id string, error string)`
     -   `aiWorkerWorkQueueDefinition.remove(idOrName string) (error string)`
     -   `aiWorkerWorkQueue.getStatus(queueName string) (statusMap map, error string)`
 -   **Task Management**:
     -   `aiWorkerWorkQueue.submitTask(queueName string, workItemDefinitionName? string, payloadOverrideJSON? string, targetCriteriaOverrideJSON? string, dataSourceRefsOverrideList? []string, priorityOverride? int, ...) (taskID string, error string)`
     -   `aiWorkerWorkQueue.getTaskInfo(taskID string) (taskInfoMap map, error string)`
 -   **Instance Management (Standalone/Direct)**:
     -   `aiWorkerInstance.spawn(definitionNameOrID string, configOverridesJSON? string) (instanceID string, error string)`
     -   `aiWorkerInstance.get(instanceID string) (instanceMap map, error string)`
     -   `aiWorkerInstance.listActive(filterMap? map) (instancesList []map, error string)`
     -   `aiWorkerInstance.retire(instanceID string, reason? string) (error string)`
 -   **File Sync Tools**:
     -   `aiWorker.syncDataSource(dataSourceName string, definitionContextNameOrID? string) (summaryMap map, error string)`: Syncs a `GlobalDataSourceDefinition` (of type local_directory) to its `RemoteTargetPath`. If `definitionContextNameOrID` is provided, uses data sources available to that worker definition.
 -   **Performance & Logging**: (Existing tools like `AIWorker.GetPerformanceRecords` remain relevant).
 -   **AIWorkerManager Configuration Tools**:
     -   `aiWorkerManager.setConfigLoadPolicy(policyName string) (error string)`: policyName is "FailFastOnError" or "LoadValidAndReportErrors".
     -   `aiWorkerManager.getConfigLoadPolicy() (policyName string, error string)`
 -   **Omni-Loader Tool**:
     -   `aiWorkerManager.loadConfigBundleFromString(jsonBundleString string, loadPolicyOverride? string) (summaryMap map, errorsList []string)`

 **Tool Implementation Notes**: `loadFromString` tools unmarshal, validate (respecting `ConfigLoadPolicy` for bundles), and then activate/store definitions. `updateFromString` tools would fetch by ID/Name, apply changes from JSON, validate, and then activate/store.

 ## 6. Error Handling

 Structured error handling (`RuntimeError`, predefined `ErrorCode`s) will be used. New error codes for pool, queue, task, item definition, data source resolution, SAI resolution, and loading policy issues (e.g., `ErrPoolNotFound`, `ErrWorkItemDefinitionInvalid`, `ErrConfigLoadPolicyViolation`).

 ## 7. Key Design Principles and Considerations

 -   **Modularity & Separation of Concerns**: Clear distinction between static definitions and dynamic runtime entities. Centralized management of global definitions.
 -   **State Management & Persistence**: All configurations are persisted. Task persistence optional.
 -   **Extensibility**: `DataSourceType`, SAI integration designed for future expansion.
 -   **Resource Pooling**: Efficient management of worker instances.
 -   **Asynchronous Task Processing**: Decoupled submission/execution via queues.
 -   **Task Templating**: `WorkItemDefinition`s for standardizing task submissions.
 -   **Context Layering**: Flexible hierarchical definition of data access, tool permissions, supervision.
 -   **Defensive Loading (NEW)**: Load-then-activate principle for configuration changes, governed by `ConfigLoadPolicy`.
 -   **Monitoring & Control**: Performance tracking, rate limiting, granular permissions, controlled data access, criticality hints (`IsMissionCritical`).
 -   **Integration with NeuroScript**: Primary control via NeuroScript tools.
 -   **Concurrency Safety**: For shared structures in `AIWorkerManager`.

 ## 8. Security Considerations for DataSources and Tool Permissions

 -   **External File Access Validation**: `GlobalDataSourceDefinition.LocalPath` with `AllowExternalReadAccess=true` MUST be validated against a system-admin-defined whitelist of external base directories by `AIWorkerManager` before activation.
 -   **Tool Permissions Enforcement**: The `SecurityLayer` must be made aware of the active `AIWorkerDefinition`'s specific `ToolAllowlist` and `ToolDenylist` when validating any tool call. Denylists always override allowlists.
 -   **Strict Write Sandboxing**: All file write operations by standard NeuroScript tools (`tool.WriteFile`, etc.) are strictly confined to the primary interpreter sandbox, regardless of `GlobalDataSourceDefinition`s.
 -   **Sync Operations**: `aiWorker.syncDataSource` involves reading from potentially external (but validated) local paths and writing to the remote File API service, governed by File API client credentials.

 ## 9. Context Resolution (DataSources and Supervisory AI)

 A hierarchical merging/override strategy determines the "effective" context for a task.

 **9.1. Effective DataSources**
 Determined by an **additive merge** of unique `GlobalDataSourceDefinition` names referenced in the following order (later items add to the set):
 1.  `AIWorkerDefinition.DataSourceRefs` (of the instance executing the task)
 2.  `WorkQueueDefinition.DataSourceRefs` (if task is from a queue)
 3.  `AIWorkerPoolDefinition.DataSourceRefs` (if instance is from a pool)
 4.  `AIWorkerInstance.DataSourceRefs` (instance-specific, dynamic attachments)
 5.  `WorkItemDefinition.DefaultDataSourceRefs` (if `WorkItem` used a definition)
 6.  `WorkItem.DataSourceRefs` (most specific additions/overrides for the task)
 The `AIWorkerManager` resolves these names to full `GlobalDataSourceDefinition` objects for the `AIWorkerInstance.ResolvedDataSources` runtime field.

 **9.2. Effective Supervisory AI (Future)**
 Determined by **hierarchical override** (first non-empty reference found wins):
 1.  `WorkItem.SupervisoryAIRef`
 2.  `WorkItemDefinition.DefaultSupervisoryAIRef` (if `WorkItem` used a definition)
 3.  `AIWorkerInstance.SupervisoryAIRef`
 4.  `AIWorkerPoolDefinition.SupervisoryAIRef`
 5.  `WorkQueueDefinition.SupervisoryAIRef`
 6.  `AIWorkerDefinition.DefaultSupervisoryAIRef`
 The `AIWorkerManager` routes key info (events, performance data) to the effective SAI.

 **9.3. Computer (`ng`) vs. SAI Management**
 -   **Primary Management**: Definitions (DataSources, Workers, Pools, Queues, ItemDefs) and task submissions are primarily managed via NeuroScript tools (e.g., by `ng`).
 -   **SAI Role**: Observes via "key info feed," generates `SupervisorFeedback`, and can potentially react by executing its own permitted NeuroScript tools.

 ---
 ## 10. Document Metadata

 :: version: 0.4.1
 :: type: NSproject
 :: subtype: design_document
 :: project: NeuroScript
 :: purpose: Describes the design for the AI Worker Management (ai_wm_*) subsystem, including worker definitions, instances, global data sources, worker pools, work queues, work item definitions, configurable loading policies, and supervisory AI provisions.
 :: status: under_review
 :: author: AJP (Primary), Gemini (Contributor)
 :: created: 2025-05-09
 :: modified: 2025-05-10
 :: dependsOn: pkg/core/ai_worker_types.go, pkg/core/ai_wm.go, (conceptual: pkg/core/ai_wm_datasources.go, pkg/core/ai_wm_pools.go, pkg/core/ai_wm_queues.go, pkg/core/ai_wm_item_defs.go)
 :: reviewCycle: 4
 :: nextReviewDate: 2025-05-17
 ---