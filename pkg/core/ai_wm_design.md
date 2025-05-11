 # AI Worker Management (ai_wm_*) System Design

 ## 1. Overview

 The AI Worker Management system (ai_wm_*) within the pkg/core package of NeuroScript provides a comprehensive framework for defining, managing, executing, and monitoring AI-powered workers. These workers typically represent Large Language Models (LLMs) or other model-based agents. The system is designed to support both stateful, instance-based interactions and stateless, one-shot task executions, potentially managed through worker pools and work queues. Key features include persistent worker definitions; flexible and shared data source configurations with controlled external file access capabilities; definitions for templatizing work items (WorkItemDefinition); API key management; worker-specific tool allow/deny lists; configurable rate limiting; detailed performance tracking; provisions for Supervisory AI (SAI) attachment for monitoring and feedback; **and a mechanism for accumulating, distilling, and utilizing shared operational knowledge and lessons learned to improve ongoing operations (NEW/Future Focus)**. The system exposes its functionalities to NeuroScript via a dedicated toolset.

 ## 2. Core Components and Concepts

 ### 2.1. AIWorkerManager (ai_wm.go)

 **Role**: This is the central orchestrator and entry point for the entire AI Worker Management system. It is responsible for the lifecycle management of:
 -   `AIWorkerDefinition`s (blueprints for workers).
 -   `GlobalDataSourceDefinition`s: Centrally defined data sources.
 -   `AIWorkerPoolDefinition`s and active `AIWorkerPool`s: For managing groups of worker instances.
 -   `WorkQueueDefinition`s and active `WorkQueue`s: For managing task submission and dispatch.
 -   `WorkItemDefinition`s: For templatizing tasks.
 -   Active `AIWorkerInstance`s (live, stateful worker sessions, potentially managed by pools).
 -   **(NEW/Future)** `KnowledgeBaseManager` for storing and managing `DistilledLesson`s.

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
 -   **(NEW/Future)** `DistilledLesson`s managed by the `KnowledgeBaseManager`.
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
 -   **(NEW/Future)** `DistilledLesson`s to/from `ai_knowledge_base.json` (or a more structured store).
 -   **(Future)** `WorkItem`s if queue persistence is enabled (e.g., to `ai_work_items.db`).
 These files are stored within a specified sandbox directory.

 **Initialization**:
 The `NewAIWorkerManager` constructor initializes the manager, loads its own operational configuration (like `ConfigLoadPolicy`), attempts to load all defined configurations (worker definitions, data sources, work item definitions, pools, queues, **distilled lessons (Future)**) and historical performance data from the sandbox, and sets up initial rate trackers for each loaded worker definition.

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
 -   **`IsMissionCritical`** (bool, omitempty): If true, indicates that this pool's operational status is critical. Failures in loading or managing this pool might trigger stricter error handling or alerts.
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
 -   **`CurrentOperationalLog` ([]OperationalLogEntry, omitempty, Future)**: Runtime log of key operations/decisions for current task. Tagged `json:"-"`.
 -   **`LastTaskSummary` (string, omitempty, Future)**: Worker-generated summary of its last task's execution.
 -   **`ResolvedDataSources`** ([]*GlobalDataSourceDefinition): Runtime field, not directly persisted in instance JSON. Derived from all applicable contexts.

 **Lifecycle**: Spawning can be direct or by a pool. Task execution uses context resolution (Section 9). **(NEW/Future)** Post-task, instance may generate `LastTaskSummary`. Retirement can be manual or by pool policy.

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
         iv. **(NEW/Future)** **Relevant Knowledge Injection**: Queries the `KnowledgeBaseManager` for `DistilledLesson`s relevant to the `WorkItem`. A summary of these lessons is provided to the `AIWorkerInstance` as part of its initial context/prompt for the task.
     e.  Assigns the `WorkItem` and its context to the instance. Instance status changes to `Busy`, `CurrentTaskID` is set.
 3.  The `AIWorkerInstance` executes the task.
 4.  Tool calls are validated by `SecurityLayer` against effective tool permissions and use effective data sources.
 5.  Upon completion/failure:
     a.  **(NEW/Future)** The instance may generate its `LastTaskSummary`.
     b.  A `PerformanceRecord` is generated (linked by `WorkItem.PerformanceRecordID`), **including any `WorkerGeneratedSummary` (NEW/Future)**.
     c.  `WorkItem` status/result/error updated. Instance becomes `Idle` or is retired.
     d.  **(MODIFIED/Future)** Relevant information/events, **including the `PerformanceRecord` with `WorkerGeneratedSummary`**, are fed to the effective SAI for learning and knowledge distillation.
 6.  Task originator can query `WorkItem` status/result via `TaskID`.

 The original `AIWorkerManager.ExecuteStatelessTask()` method may be refactored as a synchronous convenience wrapper around this queue submission process, using a default or implicit `WorkItemDefinition`.

 ---
 ### 2.9. Performance Tracking (ai_worker_types.go, ai_wm_performance.go)

 `PerformanceRecord`s are generated for each task. If a task originates as a `WorkItem`, the `PerformanceRecord.TaskID` should match `WorkItem.TaskID`.
 **(NEW/Future)** `PerformanceRecord` will include a `WorkerGeneratedSummary` field, capturing condensed operational accounts from the worker.
 These records contribute to the `AggregatePerformanceSummary` on the `AIWorkerDefinition` used by the executing instance.

 **PerformanceRecord (`ai_worker_types.go`)**:
 -   **`TaskID`** (string): Can be `WorkItem.TaskID` if applicable, or a unique ID for stateless calls.
 -   **`InstanceID`** (string): Can be "stateless-<uuid>" for direct calls not using an instance.
 -   **`DefinitionID`** (string): `AIWorkerDefinition` ID used.
 -   **`TimestampStart`, `TimestampEnd`** (time.Time).
 -   **`DurationMs`** (int64).
 -   **`Success`** (bool).
 -   **`InputContext`** (map[string]interface{}, omitempty): e.g., prompt hash, summary of `WorkItem` payload, key input parameters.
 -   **`LLMMetrics`** (map[string]interface{}, omitempty): Raw metrics from LLM (tokens, finish reason, model used etc.).
 -   **`CostIncurred`** (float64, omitempty).
 -   **`OutputSummary`** (string, omitempty): Trimmed, hashed, or representative summary of the primary output.
 -   **`ErrorDetails`** (string, omitempty).
 -   **`SupervisorFeedback`** (*SupervisorFeedback, omitempty).
 -   **`WorkerGeneratedSummary`** (string, omitempty): **(NEW/Future)** Condensed summary or key observations provided by the worker instance after completing the task.
 -   **`OperationalLogRef`** (string, omitempty): **(NEW/Future)** Reference to a more detailed (but still potentially condensed) operational log, if stored separately.

 ### 2.10. Rate Limiting (ai_worker_types.go, ai_wm_ratelimit.go)

 `RateLimitPolicy` within `AIWorkerDefinition` continues to govern instances spawned from that definition, whether standalone or by a pool. Pools must respect these limits when scaling.

 ### 2.11. API Key Management (ai_worker_types.go, ai_wm.go)

 API key resolution via `AIWorkerDefinition.Auth` remains unchanged.

 ---
 ### 2.12. Knowledge Base & Lessons Learned (Future Focus)

 **Role**: To enable the AI worker ecosystem to learn from past operations and improve future performance, efficiency, and success rates.

 **Components**:
 -   **`WorkerOperationalLog` (Conceptual / `PerformanceRecord.WorkerGeneratedSummary`)**:
     -   Each worker, after completing a task, can generate a condensed account of its operations (stored in `PerformanceRecord.WorkerGeneratedSummary`). This includes key decisions, successful tool sequences, challenges, parameters that worked well, or concise "learnings."
 -   **Supervisory AI (SAI) - Condensation Role**:
     -   The designated SAI for a task, worker type, pool, or queue receives `PerformanceRecord`s (including `WorkerGeneratedSummary` and `SupervisorFeedback`).
     -   The SAI's role includes analyzing these records over time to identify patterns, common issues, best practices, and to synthesize "distilled lessons."
 -   **`DistilledLesson` (`ai_worker_types.go` - NEW type)**:
     -   `LessonID` (string): Unique identifier.
     -   `Title` (string): A concise summary of the lesson.
     -   `ApplicabilityContext` (map[string]interface{}): Describes when this lesson is relevant (e.g., `{"workItemDefinitionName": "AnalyzeDXF", "encounteredErrorPattern": "timeout_tool_X"}`).
     -   `Insight` (string): The core learning or recommendation (e.g., "Increase timeout for tool_X to 120s when DXF complexity > Y").
     -   `ActionableSteps` ([]string, omitempty): Specific steps or parameter changes suggested.
     -   `Confidence` (float64, omitempty): How confident the SAI is in this lesson.
     -   `EvidenceLinks` ([]string, omitempty): References to `PerformanceRecord.TaskID`s or `WorkItem.TaskID`s that support this lesson.
     -   `CreatedBySAI` (string): ID/Name of the SAI worker/definition that generated this lesson.
     -   `Version` (int): Version of this lesson, allowing refinement.
     -   `Timestamp` (time.Time).
 -   **`KnowledgeBaseManager` (Conceptual component within `AIWorkerManager`)**:
     -   **Storage**: Persistently stores `DistilledLesson`s (e.g., in `ai_knowledge_base.json` or a dedicated database).
     -   **API for SAIs**: Allows SAIs to submit new `DistilledLesson`s or updates to existing ones.
     -   **API for Workers/Dispatcher**: Allows querying for relevant `DistilledLesson`s based on the current task context (`WorkItemDefinitionName`, payload characteristics, target worker capabilities).

 **Workflow**:
 1.  An `AIWorkerInstance` performs a task and generates a `WorkerGeneratedSummary`.
 2.  The `PerformanceRecord` (with this summary) is sent to the effective SAI.
 3.  The SAI analyzes records to generate/refine `DistilledLesson`s.
 4.  SAI submits `DistilledLesson`s to the `KnowledgeBaseManager`.
 5.  Before a new task, dispatcher/instance queries `KnowledgeBaseManager` for relevant lessons.
 6.  These lessons are provided as context to the `AIWorkerInstance`.

 ---
 ## 3. Data Structures and Types (Key Types from `ai_worker_types.go`)

 -   `InteractionModelType`, `APIKeySourceMethod`, `AIWorkerProvider`, `AIWorkerDefinitionStatus`, `AIWorkerInstanceStatus`, `ConfigLoadPolicy`
 -   `APIKeySource`, `RateLimitPolicy`, `TokenUsageMetrics`, `SupervisorFeedback`, `AIWorkerPerformanceSummary`
 -   `DataSourceType`
 -   `GlobalDataSourceDefinition`
 -   `AIWorkerDefinition` (Modified)
 -   `AIWorkerInstance` (Modified)
 -   `PerformanceRecord` (Modified: `WorkerGeneratedSummary`, `OperationalLogRef`)
 -   `RetiredInstanceInfo`
 -   `InstanceRetirementPolicy`
 -   `AIWorkerPoolDefinition` (Includes `IsMissionCritical`)
 -   `RetryPolicy`
 -   `WorkQueueDefinition` (Includes `IsMissionCritical`)
 -   `WorkItemDefinition`
 -   `WorkItemStatus`
 -   `WorkItem` (Includes `WorkItemDefinitionName`)
 -   `AIWorkerManagementConfigBundle`
 -   `LLMCallMetrics`
 -   **`DistilledLesson` (NEW/Future)**

 ## 4. Persistence Strategy

 **Format**: JSON for definitions. (Future: consider embedded DB for `WorkItem`s if `PersistTasks` is true and volume is high, and for `DistilledLesson`s).

 **Files** (in `sandboxDir`):
 -   `ai_worker_definitions.json`
 -   `ai_worker_performance_data.json`
 -   `ai_global_data_source_definitions.json`
 -   `ai_worker_pool_definitions.json`
 -   `ai_work_queue_definitions.json`
 -   `ai_work_item_definitions.json`
 -   **(Future)** `ai_knowledge_base.json` (or DB) for storing `DistilledLesson`s.

 **Loading/Saving**: `AIWorkerManager` handles loading and saving all definition types, respecting its `ConfigLoadPolicy` during bundle loads. Individual `add/update/remove` operations trigger persistence. A **Load & Activate** principle should be applied: definitions are validated before being committed to the live state and persisted.

 ## 5. NeuroScript Tool Integration (ai_wm_tools.go)

 **Tool Categories & Key Tools**:
 -   **AIWorkerDefinition Management**:
     -   `aiWorkerDefinition.loadFromString(jsonString string) (id string, error string)` (and other CRUD: get, list, updateFromString, remove)
 -   **Global Data Source Management**:
     -   `aiWorkerGlobalDataSource.loadFromString(jsonString string) (name string, error string)` (and other CRUD)
 -   **Work Item Definition Management**:
     -   `aiWorkerWorkItemDefinition.loadFromString(jsonString string) (id string, error string)` (and other CRUD)
 -   **Worker Pool Management**:
     -   `aiWorkerPoolDefinition.loadFromString(jsonString string) (id string, error string)` (and other CRUD including `isMissionCritical` params)
     -   `aiWorkerPool.getInstanceStatus(poolName string) (statusMap map, error string)`
 -   **Work Queue Management**:
     -   `aiWorkerWorkQueueDefinition.loadFromString(jsonString string) (id string, error string)` (and other CRUD including `isMissionCritical` params)
     -   `aiWorkerWorkQueue.getStatus(queueName string) (statusMap map, error string)`
 -   **Task Management**:
     -   `aiWorkerWorkQueue.submitTask(queueName string, workItemDefinitionName? string, payloadOverrideJSON? string, ...)`
     -   `aiWorkerWorkQueue.getTaskInfo(taskID string) (taskInfoMap map, error string)`
 -   **Instance Management (Standalone/Direct)**: (as before)
 -   **File Sync Tools**:
     -   `aiWorker.syncDataSource(dataSourceName string, definitionContextNameOrID? string)`
 -   **Performance & Logging**: (as before)
 -   **AIWorkerManager Configuration Tools**:
     -   `aiWorkerManager.setConfigLoadPolicy(policyName string)`
     -   `aiWorkerManager.getConfigLoadPolicy()`
 -   **Omni-Loader Tool**:
     -   `aiWorkerManager.loadConfigBundleFromString(jsonBundleString string, loadPolicyOverride? string)`
 -   **(NEW/Future) Worker Self-Reflection/Logging Tools**:
     -   `aiWorker.logOperationalInsight(level string, message string, detailsMap? map)`
 -   **(NEW/Future) Knowledge Base Tools**:
     -   `aiKnowledgeBase.queryLessons(taskContextJSON string) (lessonsList []map, error string)`
     -   `aiSAI.submitDistilledLesson(lessonJSON string) (lessonID string, error string)`
     -   `aiSAI.getPendingPerformanceReviews(filterJSON? string) (recordsList []map, error string)`

 ## 6. Error Handling

 Structured error handling (`RuntimeError`, predefined `ErrorCode`s) will be used. New error codes for pool, queue, task, item definition, data source resolution, SAI resolution, loading policy issues, and knowledge base operations.

 ## 7. Key Design Principles and Considerations

 -   **Modularity & Separation of Concerns**
 -   **State Management & Persistence**
 -   **Extensibility**
 -   **Resource Pooling**
 -   **Asynchronous Task Processing**
 -   **Task Templating** (`WorkItemDefinition`)
 -   **Context Layering** (for DataSources, SAI)
 -   **Defensive Loading** (`ConfigLoadPolicy`, Load & Activate principle)
 -   **Monitoring & Control** (Performance, Rate Limits, Permissions, Criticality flags)
 -   **Adaptive Learning & Knowledge Sharing (NEW/Future)**: Via worker summaries, SAI condensation, and knowledge base.
 -   **Integration with NeuroScript**
 -   **Concurrency Safety**

 ## 8. Security Considerations for DataSources and Tool Permissions

 -   **External File Access Validation**: `GlobalDataSourceDefinition.LocalPath` with `AllowExternalReadAccess=true` MUST be validated against a system-admin-defined whitelist by `AIWorkerManager`.
 -   **Tool Permissions Enforcement**: `SecurityLayer` must use the active `AIWorkerDefinition`'s `ToolAllowlist`/`ToolDenylist`.
 -   **Strict Write Sandboxing**: All file write operations by standard NeuroScript tools are confined to the primary interpreter sandbox.
 -   **Sync Operations**: `aiWorker.syncDataSource` reads from validated local paths and writes to the remote File API.

 ## 9. Context Resolution (DataSources and Supervisory AI)

 A hierarchical merging/override strategy determines the "effective" context.

 **9.1. Effective DataSources**
 Determined by an **additive merge** of unique `GlobalDataSourceDefinition` names referenced in order: `AIWorkerDefinition` -> `WorkQueueDefinition` -> `AIWorkerPoolDefinition` -> `AIWorkerInstance` -> `WorkItemDefinition` (defaults) -> `WorkItem` (specifics).

 **9.2. Effective Supervisory AI (Future)**
 Determined by **hierarchical override** (most specific wins): `WorkItem` -> `WorkItemDefinition` -> `AIWorkerInstance` -> `AIWorkerPoolDefinition` -> `WorkQueueDefinition` -> `AIWorkerDefinition`.

 **9.3. Computer (`ng`) vs. SAI Management**
 -   **Primary Management**: Definitions and task submissions primarily via NeuroScript tools.
 -   **SAI Role**: Observes via "key info feed", generates feedback/lessons, and can react by executing permitted NeuroScript tools.

 ---
 ## 10. Document Metadata

 :: version: 0.4.2
 :: type: NSproject
 :: subtype: design_document
 :: project: NeuroScript
 :: purpose: Describes the design for the AI Worker Management (ai_wm_*) subsystem, including worker definitions, instances, global data sources, worker pools, work queues, work item definitions, configurable loading policies, supervisory AI provisions, and a framework for shared memory and lessons learned.
 :: status: under_review
 :: author: AJP (Primary), Gemini (Contributor)
 :: created: 2025-05-09
 :: modified: 2025-05-10
 :: dependsOn: pkg/core/ai_worker_types.go, pkg/core/ai_wm.go, (conceptual: pkg/core/ai_wm_datasources.go, pkg/core/ai_wm_pools.go, pkg/core/ai_wm_queues.go, pkg/core/ai_wm_item_defs.go, pkg/core/ai_wm_knowledgebase.go)
 :: reviewCycle: 4
 :: nextReviewDate: 2025-05-17
 ---