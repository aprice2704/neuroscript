 # AI Worker Management (ai_wm_*) System Design

 ## 1. Overview
 *** MODIFIED ***
 The AI Worker Management system (ai_wm_*) within the pkg/core package of NeuroScript provides a comprehensive framework for defining, managing, executing, and monitoring AI-powered workers. These workers typically represent Large Language Models (LLMs) or other model-based agents. The system is designed to support both stateful, instance-based interactions and stateless, one-shot task executions, potentially managed through worker pools and work queues. Key features include persistent worker definitions; flexible and shared data source configurations; **definitions for templatizing work items (WorkItemDefinition - NEW)**; API key management; worker-specific tool allow/deny lists; configurable rate limiting; detailed performance tracking; and provisions for Supervisory AI (SAI) attachment. The system exposes its functionalities to NeuroScript via a dedicated toolset.

 ## 2. Core Components and Concepts

 ### 2.1. AIWorkerManager (ai_wm.go)
 *** MODIFIED ***
 **Role**: This is the central orchestrator and entry point for the entire AI Worker Management system. It is responsible for the lifecycle management of:
 -   `AIWorkerDefinition`s (blueprints for workers).
 -   `GlobalDataSourceDefinition`s **(NEW)**: Centrally defined data sources.
 -   `AIWorkerPoolDefinition`s and active `AIWorkerPool`s **(NEW)**: For managing groups of worker instances.
 -   `WorkQueueDefinition`s and active `WorkQueue`s **(NEW)**: For managing task submission and dispatch.
 -   **`WorkItemDefinition`s (NEW)**: For templatizing tasks.
 -   Active `AIWorkerInstance`s (live, stateful worker sessions, potentially managed by pools).

 It handles persistence of these configurations, performance data, enforces rate limits, resolves API keys, and dispatches tasks from queues to pools.

 **State**: The manager maintains in-memory collections of:
 -   `AIWorkerDefinition`s.
 -   Active `AIWorkerInstance`s.
 -   `WorkerRateTracker`s.
 -   `GlobalDataSourceDefinition`s **(NEW)**.
 -   `AIWorkerPoolDefinition`s and runtime `AIWorkerPool` states **(NEW)**.
 -   `WorkQueueDefinition`s and runtime `WorkQueue` states (including `WorkItem`s if queues are in-memory) **(NEW)**.
 -   **`WorkItemDefinition`s (NEW)**.
 -   **(Future)** References to active Supervisory AI (SAI) instances or configurations.

 **Persistence**:
 *** MODIFIED ***
 The manager is responsible for loading and saving:
 -   Worker definitions to/from `ai_worker_definitions.json`.
 -   Performance data of retired instances to/from `ai_worker_performance_data.json`.
 -   Global Data Source definitions to/from `ai_global_data_source_definitions.json`.
 -   **`AIWorkerPoolDefinition`s to/from `ai_worker_pool_definitions.json` (NEW)**.
 -   **`WorkQueueDefinition`s to/from `ai_work_queue_definitions.json` (NEW)**.
 -   **`WorkItemDefinition`s to/from `ai_work_item_definitions.json` (NEW)**.
 -   **(Future)** `WorkItem`s if queue persistence is enabled.
 These files are stored within a specified sandbox directory.

 **Initialization**:
 *** MODIFIED ***
 The `NewAIWorkerManager` constructor initializes the manager, attempts to load all defined configurations (worker definitions, data sources, **work item definitions (NEW)**, pools, queues) and historical performance data.

 ---
 ### 2.2. Global Data Source Management (ai_wm_datasources.go - *new conceptual file*)

 This new component, managed by the `AIWorkerManager`, handles the definition and persistence of shared data sources that AI Workers can reference.

 **GlobalDataSourceDefinition (`ai_worker_types.go`)**:
 Serves as a template for a data access area. Each definition includes:
 -   **`Name`** (string): A unique, human-readable name/ID for this data source (e.g., "project_schematics", "shared_part_files"). This name is used by `AIWorkerDefinition`s to reference it.
 -   **`Type`** (DataSourceType): Indicates the nature of the data source (e.g., `DataSourceTypeLocalDirectory`, `DataSourceTypeFileAPI`).
 -   **`LocalPath`** (string, optional): For `DataSourceTypeLocalDirectory`, this is the absolute path on the local filesystem.
 -   **`AllowExternalReadAccess`** (bool, optional): For `DataSourceTypeLocalDirectory`, if true, `LocalPath` can be outside the primary interpreter sandbox. **This requires strict validation by `AIWorkerManager` against a system-administrator-defined list of permissible external base paths.**
 -   **`FileAPIPath`** (string, optional): For `DataSourceTypeFileAPI`, this is the path within the Google File API (e.g., "fm:/shared/designs/").
 -   **`RemoteTargetPath`** (string, optional): For `DataSourceTypeLocalDirectory`, suggests a default target path in the remote File API if this source is synced (e.g., "synced_data_sources/<DataSourceName>").
 -   **`ReadOnly`** (bool): If true, the worker should treat this source as read-only. Critically, NeuroScript file *write* tools (`toolWriteFile`, etc.) **always** operate strictly within the primary interpreter sandbox, regardless of this flag for external sources.
 -   **`Filters`** ([]string, optional): Glob patterns for files to include from this source.
 -   **`Recursive`** (bool, optional): Whether to consider files recursively from this source.
 -   **`Metadata`** (map[string]interface{}, optional): Additional custom information.

 **Management**: The `AIWorkerManager` will provide CRUD operations for `GlobalDataSourceDefinition`s. All changes will be persisted to `ai_global_data_source_definitions.json`.

 **Security**: The `AIWorkerManager`, upon loading or adding a `GlobalDataSourceDefinition` of type `DataSourceTypeLocalDirectory` with `AllowExternalReadAccess` set to true, **must** validate its `LocalPath` against a pre-configured whitelist of allowed external base directories. This whitelist is a system-level configuration, not part of the NeuroScript-modifiable definitions, ensuring that workers cannot be configured to access arbitrary filesystem locations. If validation fails, the data source definition should be rejected or marked as unusable.
 ---

 ### 2.3. AIWorkerDefinition (ai_worker_types.go, ai_wm_definitions.go)
 **Role**: Serves as a static blueprint or template that defines all configurable aspects of a particular type of AI worker. It dictates how instances of this worker behave and how stateless calls using this definition are processed.

 **Key Attributes**:
 - **DefinitionID** (string): A unique identifier (UUID) for the definition.
 - **Name** (string): A human-readable name for easy identification.
 - **Provider** (AIWorkerProvider): Specifies the source or vendor of the AI model.
 - **ModelName** (string): The specific model identifier from the provider.
 - **Auth** (APIKeySource): Defines the method and value for retrieving the API key.
 - **InteractionModels** ([]InteractionModelType): Supported modes of interaction.
 - **Capabilities** ([]string): Functional capabilities of the worker.
 - **BaseConfig** (map[string]interface{}): Default LLM configuration parameters.
 - **CostMetrics** (map[string]float64): Cost associated with using the worker.
 - **RateLimits** (RateLimitPolicy): Usage limits.
 - **Status** (AIWorkerDefinitionStatus): Operational status (active, disabled, archived).
 - **DefaultFileContexts** ([]string): List of file paths or URIs (potentially referencing data sources, e.g., `datasource://<DataSourceName>/path/to/file`) implicitly included in the context.
 - **AggregatePerformanceSummary** (AIWorkerPerformanceSummary): Aggregated performance metrics.
 - **Metadata** (map[string]interface{}): Flexible key-value store.
 -   `DataSourceRefs` ([]string): An array of names/IDs referencing `GlobalDataSourceDefinition`s. **This provides a baseline set of data sources for this worker type (NEW)**.
 -   `ToolAllowlist` ([]string): Worker-definition-specific tool allowlist.
 -   `ToolDenylist` ([]string): Worker-definition-specific tool denylist.
 -   **`DefaultSupervisoryAIRef` (string, optional, Future)**: References the name/ID of an `AIWorkerDefinition` suitable for supervising tasks run by this worker type, or a pre-configured SAI instance.

 **Management**: The AIWorkerManager provides CRUD operations on definitions. Changes are persisted to `ai_worker_definitions.json`.

 ---
 ### 2.4. AIWorkerPools (ai_wm_pools.go - *new conceptual file*)

 **AIWorkerPoolDefinition (`ai_worker_types.go`)**:
 A persistent configuration for a worker pool.
 -   **`PoolID`** (string): Unique identifier.
 -   **`Name`** (string): Human-readable name.
 -   **`TargetAIWorkerDefinitionName`** (string): Name of the `AIWorkerDefinition` used to spawn instances in this pool.
 -   **`MinIdleInstances`** (int): Desired minimum number of idle instances.
 -   **`MaxTotalInstances`** (int): Maximum total instances allowed in this pool (also subject to global `AIWorkerManager` limits and definition rate limits).
 -   **`InstanceRetirementPolicy`** (struct, e.g., `MaxTasksPerInstance`, `MaxInstanceAgeHours`): Policy for when to retire instances.
 -   **`DataSourceRefs`** ([]string, optional): References to `GlobalDataSourceDefinition`s that are common to all workers in this pool, supplementing or being overridden by more specific contexts (see Section 9).
 -   **`SupervisoryAIRef` (string, optional, Future)**: Reference to an SAI for this pool.

 **AIWorkerPool (Runtime Structure)**:
 Managed by `AIWorkerManager`.
 -   Tracks its `AIWorkerPoolDefinition`.
 -   Maintains a list of active `AIWorkerInstance` IDs belonging to the pool and their statuses.
 -   Handles scaling logic (spawning/retiring instances based on demand and definition).
 -   Provides idle instances to the Task Dispatcher.

 ---
 *** NEW ***
 ### 2.5. WorkItemDefinition (ai_worker_types.go - new type)

 **Role**: Serves as a persistent template or blueprint for creating similar `WorkItem`s, reducing redundancy and ensuring consistency when submitting common types of tasks to a `WorkQueue`.

 **WorkItemDefinition (`ai_worker_types.go`)**:
 -   **`WorkItemDefinitionID`** (string): Unique identifier (e.g., UUID).
 -   **`Name`** (string): A human-readable, unique name for this template (e.g., "AnalyzePanelStress", "GenerateToolpathVariant").
 -   **`Description`** (string, optional): A brief description of the purpose of tasks created from this definition.
 -   **`DefaultTargetWorkerCriteria`** (map[string]interface{}, optional): Pre-defined criteria for selecting a worker or pool (e.g., `{"definitionName": "panel_analyzer_v2", "capabilities": ["stress_analysis"]}`). These can be overridden at submission time.
 -   **`DefaultPayloadSchema`** (map[string]interface{}, optional): A JSON schema map defining the expected structure, types, and potentially default values for the `WorkItem.Payload`. This helps validate `WorkItem`s created from this definition and can provide boilerplate for payload construction. For example: `{"input_file_uri": {"type": "string", "description": "URI of the panel design file"}, "material_type": {"type": "string", "default": "steel_grade_a"}}`.
 -   **`DefaultDataSourceRefs`** ([]string, optional): A list of `GlobalDataSourceDefinition` names that are typically required or useful for tasks of this type. These are merged with other contextually relevant data sources (see Section 9).
 -   **`DefaultPriority`** (int, optional): A default priority for `WorkItem`s created from this template.
 -   **`DefaultSupervisoryAIRef`** (string, optional, Future): Default SAI reference for tasks of this type.
 -   **`Metadata`** (map[string]interface{}, optional): Additional custom metadata.

 **Management**: The `AIWorkerManager` will provide CRUD operations for `WorkItemDefinition`s. All changes will be persisted to `ai_work_item_definitions.json`.

 **Usage**: When submitting a task to a queue, a `WorkItemDefinitionName` can be provided. The system (e.g., `WorkQueue.SubmitTask` tool or underlying manager logic) will use this definition as a base, and any explicitly provided parameters for the `WorkItem` (like payload values, specific data source refs) will override or merge with the defaults from the `WorkItemDefinition`.

 ---
 ### 2.6. WorkQueues & WorkItems (ai_wm_queues.go - *new conceptual file*)
 *(Previously 2.5)*

 **WorkQueueDefinition (`ai_worker_types.go`)**:
 A persistent configuration for a work queue.
 -   **`QueueID`** (string): Unique identifier.
 -   **`Name`** (string): Human-readable name.
 -   **`AssociatedPoolNames`** ([]string): Names of `AIWorkerPool`(s) that service this queue.
 -   **`DefaultPriority`** (int): Default priority for tasks if not specified in the `WorkItem`.
 -   **`RetryPolicy`** (struct: `MaxRetries`, `RetryDelaySeconds`).
 -   **`PersistTasks`** (bool): If true, `WorkItem`s in this queue should be persisted to survive restarts.
 -   **`DataSourceRefs`** ([]string, optional): References to `GlobalDataSourceDefinition`s relevant to all tasks in this queue, supplementing or being overridden by more specific contexts (see Section 9).
 -   **`SupervisoryAIRef` (string, optional, Future)**: Reference to an SAI for this queue.

 **WorkItem (`ai_worker_types.go`)**:
 *** MODIFIED ***
 Represents a unit of work submitted to a queue.
 -   **`WorkItemDefinitionName`** (string, optional): **(NEW)** The name of the `WorkItemDefinition` used as a template for this work item.
 -   **`TaskID`** (string): Unique identifier.
 -   **`QueueName`** (string): Name of the queue this item was submitted to.
 -   **`TargetWorkerCriteria`** (map[string]interface{}, optional): Specifies requirements for the worker/pool (e.g., `definitionName`, `capabilities`). If not specified, uses queue's associated pools or defaults from `WorkItemDefinitionName`.
 -   **`Payload`** (map[string]interface{}): Task-specific data. If `WorkItemDefinitionName` is used, this payload is merged with/validated against the definition's `DefaultPayloadSchema`.
 -   **`DataSourceRefs`** ([]string, optional): References to `GlobalDataSourceDefinition`s specific to this work item, supplementing or overriding broader contexts (see Section 9).
 -   **`Priority`** (int, optional): Can override default from queue or `WorkItemDefinitionName`.
 -   **`Status`** (WorkItemStatus: e.g., `Pending`, `Processing`, `Completed`, `Failed`, `Retrying`).
 -   **`SubmitTimestamp`, `StartTimestamp`, `EndTimestamp`** (time.Time).
 -   **`RetryCount`** (int).
 -   **`Result`** (interface{}, optional): Outcome of the task.
 -   **`Error`** (string, optional): Error details if failed.
 -   **`PerformanceRecordID`** (string, optional): Link to the generated `PerformanceRecord`.
 -   **`SupervisoryAIRef` (string, optional, Future)**: Reference to an SAI for this specific work item, potentially overriding defaults.

 **WorkQueue (Runtime Structure)**:
 Managed by `AIWorkerManager`.
 -   Tracks its `WorkQueueDefinition`.
 -   Holds pending `WorkItem`s (in-memory or backed by persistent store).
 -   Provides tasks to the Task Dispatcher.

 ---
 ### 2.7. AIWorkerInstance (ai_worker_types.go, ai_wm_instances.go)
 *(Previously 2.6)*
 **Role**: Represents a live, stateful session with an AI worker. Can be standalone (for explicit spawning) or part of an `AIWorkerPool`.

 **Key Attributes**:
 - **InstanceID** (string): A unique identifier (UUID) for the active session.
 - **DefinitionID** (string): Links the instance back to its parent AIWorkerDefinition.
 - **Status** (AIWorkerInstanceStatus): Reflects the current operational state of the instance.
 - **ConversationHistory** ([]*ConversationTurn): This field is tagged `json:"-"`. Managed by an associated ConversationManager.
 - **CreationTimestamp, LastActivityTimestamp** (time.Time): Track the instance's lifecycle.
 - **SessionTokenUsage** (TokenUsageMetrics): Accumulates token consumption for the current active session.
 - **CurrentConfig** (map[string]interface{}): The effective configuration for this instance.
 - **LastError** (string), **RetirementReason** (string): Store information about errors or why an instance was retired.
 - `ResolvedDataSources` ([]*GlobalDataSourceDefinition): Runtime field. **(Modified)** When an instance processes a task, this is the effective set of data sources derived from the task (and its potential `WorkItemDefinition`), instance, pool, queue, and definition (see Section 9).
 - **`CurrentTaskID` (string, optional)**: If processing a task from a queue, the ID of that `WorkItem`.
 - **`PoolID` (string, optional)**: If part of a pool.
 - **`DataSourceRefs`** ([]string, optional): **(NEW)** References to `GlobalDataSourceDefinition`s that are specifically configured for this instance (e.g., dynamically attached), supplementing or being overridden by `WorkItem` context.
 - **`SupervisoryAIRef` (string, optional, Future)**: Reference to an SAI for this instance.
 - **`ActiveFileContexts`** ([]string): Runtime-only list of files currently active in the instance's context (may reference files from `ResolvedDataSources`).

 **Lifecycle**:
 -   **Spawning**: Can be spawned directly (`AIWorkerManager.SpawnWorkerInstance()`) or by an `AIWorkerPool`.
 -   **Task Execution**: When an instance from a pool is assigned a `WorkItem`, its context (data sources, SAI) is established based on the resolution logic (see Section 9).
 -   **Retirement**: Can be retired manually or by pool policy.

 ---
 ### 2.8. Stateless Task Execution & Task Dispatching
 *(Previously 2.7)*
 **Role**: The system now supports a more robust task execution model via queues and pools. The original `ExecuteStatelessTask` method on `AIWorkerManager` might evolve.

 **Process (Task via Queue)**:
 1.  A task is submitted to a `WorkQueue`. This submission can reference a `WorkItemDefinitionName` and provide overriding/additional parameters to create a `WorkItem`.
 2.  The `AIWorkerManager`'s **Task Dispatcher** component:
     a.  Selects a pending `WorkItem` from a queue (considering priority, etc.).
     b.  Identifies a compatible and available `AIWorkerPool` for the task (based on `WorkItem.TargetWorkerCriteria` or `WorkQueue.AssociatedPoolNames`).
     c.  Requests an idle `AIWorkerInstance` from the selected pool. The pool manager might spawn a new instance if needed and allowed.
     d.  Resolves the **effective context** for the task:
         i.  **Effective DataSources**: Combines `DataSourceRefs` from the `WorkItem` (which may have inherited from its `WorkItemDefinition`), `AIWorkerInstance` (if any), `AIWorkerPoolDefinition`, `WorkQueueDefinition`, and the instance's `AIWorkerDefinition` (see Section 9).
         ii. **Effective Tool Permissions**: Derived from the instance's `AIWorkerDefinition` (`ToolAllowlist`, `ToolDenylist`).
         iii. **Effective SAI**: Resolved from `WorkItem` (and its potential `WorkItemDefinition`), `Instance`, `Pool`, `Queue`, `Definition` (see Section 9).
     e.  Assigns the `WorkItem` and its effective context to the instance. The instance status changes to `Busy`.
 3.  The `AIWorkerInstance` executes the task using its `AIWorkerDefinition`'s model, provider, auth, and the effective context. This is similar to the current direct LLM call but now framed within a `WorkItem`.
 4.  During execution, if the LLM attempts a tool call, the `SecurityLayer` validates it against the worker's **effective tool permissions**. File access tools use the **effective data sources**.
 5.  Upon completion/failure:
     a.  A `PerformanceRecord` is generated and associated with the `WorkItem`.
     b.  The `WorkItem` status and result/error are updated.
     c.  The `AIWorkerInstance` status changes back to `Idle` (or it might be retired by the pool).
     d.  **(Future)** Relevant information/events are fed to the effective SAI.
 6.  The originator of the task can query the `WorkItem`'s status and result using its `TaskID`.

 The original `AIWorkerManager.ExecuteStatelessTask()` could be a convenience wrapper that creates a `WorkItem` (possibly using a default "stateless_task" `WorkItemDefinition`) and submits it, then waits for the result if synchronous behavior is desired.

 ---
 ### 2.9. Performance Tracking (ai_worker_types.go, ai_wm_performance.go)
 *(Previously 2.8. Performance records will be associated with WorkItems and thus implicitly with their originating queues/pools.)*

 ### 2.10. Rate Limiting (ai_worker_types.go, ai_wm_ratelimit.go)
 *(Previously 2.9. Rate limits defined in AIWorkerDefinition still apply to instances spawned by pools.)*

 ### 2.11. API Key Management (ai_worker_types.go, ai_wm.go)
 *(Previously 2.10. Unchanged in principle.)*

 ---
 ## 3. Data Structures and Types (Key Types from `ai_worker_types.go`)
 *** MODIFIED ***
 -   `AIWorkerProvider`, `InteractionModelType`, `APIKeySourceMethod`, `AIWorkerDefinitionStatus`, `AIWorkerInstanceStatus`
 -   `APIKeySource`, `RateLimitPolicy`, `TokenUsageMetrics`, `SupervisorFeedback`
 -   `PerformanceRecord`, `AIWorkerPerformanceSummary`
 -   `AIWorkerDefinition` (now with `DataSourceRefs`, `ToolAllowlist`, `ToolDenylist`, `DefaultSupervisoryAIRef`)
 -   `AIWorkerInstance` (now with `PoolID`, `CurrentTaskID`, instance-level `DataSourceRefs` and `SupervisoryAIRef`)
 -   `RetiredInstanceInfo`
 -   `LLMCallMetrics`
 -   `DataSourceType` **(NEW)**
 -   `GlobalDataSourceDefinition` **(NEW)**
 -   `AIWorkerPoolDefinition` **(NEW)** (with `DataSourceRefs`, `SupervisoryAIRef`)
 -   `WorkQueueDefinition` **(NEW)** (with `DataSourceRefs`, `SupervisoryAIRef`)
 -   **`WorkItemDefinition` (NEW)** (with `DefaultTargetWorkerCriteria`, `DefaultPayloadSchema`, `DefaultDataSourceRefs`, `DefaultPriority`, `DefaultSupervisoryAIRef`)
 -   `WorkItem` **(NEW)** (now with `WorkItemDefinitionName?`, and its fields potentially overriding/merging with `WorkItemDefinition` defaults)
 -   `WorkItemStatus` (e.g., `Pending`, `Processing`, `Completed`, `Failed`) **(NEW)**

 ## 4. Persistence Strategy
 *** MODIFIED ***
 **Format**: JSON for definitions. A more robust solution (e.g., embedded DB) might be needed for high-volume `WorkItem` persistence.

 **Files**:
 -   `ai_worker_definitions.json`
 -   `ai_worker_performance_data.json`
 -   `ai_global_data_source_definitions.json`
 -   `ai_worker_pool_definitions.json`
 -   `ai_work_queue_definitions.json`
 -   **`ai_work_item_definitions.json` (NEW)**
 -   **(Future)** `ai_work_items.db` or similar for persistent `WorkItem`s.

 **Location**: All primary data files are located in the `sandboxDir` configured for the `AIWorkerManager`.

 **Loading/Saving**: `AIWorkerManager` handles loading and saving all these definition types.

 ## 5. NeuroScript Tool Integration (ai_wm_tools.go)
 *** MODIFIED ***
 **Tool Categories**:
 -   **AIWorkerDefinition Management**:
     -   `AIWorkerDefinition.Add`, `Get`, `List`, `Remove`
     -   `AIWorkerDefinition.Update`: Now handles updates to `dataSourceRefs`, `toolAllowlist`, `toolDenylist`, `defaultSupervisoryAIRef`.
 -   **Global Data Source Management Tools**:
     -   `AIWorkerDataSource.AddGlobal (name, type, localPath?, fileAPIPath?, allowExternalRead?, remoteTargetPath?, readOnly?, filters?, recursive?)`
     -   `AIWorkerDataSource.GetGlobal (name)`
     -   `AIWorkerDataSource.ListGlobal ()`
     -   `AIWorkerDataSource.UpdateGlobal (name, ...)`
     -   `AIWorkerDataSource.RemoveGlobal (name)`
 *** NEW ***
 -   **Work Item Definition Tools (NEW)**:
     -   `AIWorkerWorkItemDefinition.Add (name, description?, defaultTargetCriteriaJSON?, defaultPayloadSchemaJSON?, defaultDataSourceRefsList?, defaultPriority?, defaultSAIRef?)`
     -   `AIWorkerWorkItemDefinition.Get (name)`
     -   `AIWorkerWorkItemDefinition.List ()`
     -   `AIWorkerWorkItemDefinition.Update (name, ...)`
     -   `AIWorkerWorkItemDefinition.Remove (name)`
 -   **Worker Pool Management Tools**:
     -   `AIWorkerPool.Create (poolName, targetDefinitionName, minIdle, maxTotal, instanceRetirementPolicyJSON?, dataSourceRefsList?, supervisoryAIRef?)`
     -   `AIWorkerPool.Get (poolName)`
     -   `AIWorkerPool.Update (poolName, minIdle?, maxTotal?, ...)`
     -   `AIWorkerPool.Delete (poolName)`
     -   `AIWorkerPool.List ()`
     -   `AIWorkerPool.GetInstanceStatus (poolName)`: Provides runtime status of instances in the pool.
 -   **Work Queue Management Tools**:
     -   `WorkQueue.Create (queueName, associatedPoolNamesList, defaultPriority?, retryPolicyJSON?, persistTasks?, dataSourceRefsList?, supervisoryAIRef?)`
     -   `WorkQueue.Get (queueName)`
     -   `WorkQueue.Update (queueName, ...)`
     -   `WorkQueue.Delete (queueName)`
     -   `WorkQueue.List ()`
     -   `WorkQueue.GetStatus (queueName)`: (e.g., length, number of processing items).
 -   **Task Management Tools**:
     -   `WorkQueue.SubmitTask (queueName, workItemDefinitionName?, payloadOverrideJSON?, targetWorkerCriteriaOverrideJSON?, dataSourceRefsOverrideList?, priorityOverride?, supervisoryAIRef?)`: Returns `TaskID`. `payloadOverrideJSON` and `targetWorkerCriteriaOverrideJSON` are JSON strings for simplicity in NeuroScript, which will be merged with the `WorkItemDefinition` if provided.
     -   `WorkQueue.GetTaskInfo (taskID)`: Returns `WorkItem` details as a map.
     -   *(Future) `WorkQueue.CancelTask (taskID)`*
 -   **Instance Management (Standalone/Direct)**:
     -   `AIWorkerInstance.Spawn` (for cases not using pools, or for specialized instance needs).
     -   `AIWorkerInstance.Get`, `ListActive`, `Retire`, `UpdateStatus`, `UpdateTokenUsage`.
 -   **File Sync Tools**:
     -   `AIWorker.SyncDataSource (dataSourceName)`: Uses the effective `GlobalDataSourceDefinition` (based on current worker/task context if applicable, or a globally specified one) to sync its `LocalPath` to its `RemoteTargetPath` in the File API.
 -   **Performance & Logging**: (as before).

 **Tool Implementation Notes for File Access**:
 -   Standard file system tools (`toolReadFile`, `toolWriteFile`, `toolListDirectory`, etc.) will continue to operate.
 -   When accessing files related to a task being processed by a worker (especially one from a pool/queue), these tools need to be aware of the task's **effective data sources** (see Section 9).
 -   **Path Resolution**:
     -   **Initial**: Tools might require explicit `dataSourceName` argument in addition to a `relativePath` within that source if ambiguity exists.
     -   **Future URI Scheme**: `datasource://<DataSourceName>/path/to/file` would be the preferred way to reference files within these configured sources. The `FileAPI` and/or `SecurityLayer` would resolve these. The `<DataSourceName>` refers to a name from the *effective set* of `GlobalDataSourceDefinition`s for the current context.
 -   **Write Sandboxing**: All direct file writes by tools like `toolWriteFile` remain strictly within the primary interpreter sandbox.

 ## 6. Error Handling
 *(As before, but new error codes for pool/queue/task/item_definition issues, e.g., `ErrPoolNotFound`, `ErrQueueFull`, `ErrTaskDispatchFailed`, `ErrTaskNotFound`, `ErrWorkItemDefinitionNotFound`, `ErrDataSourceResolutionFailed`, `ErrSAIResolutionFailed`)*

 ## 7. Key Design Principles and Considerations
 *** MODIFIED ***
 -   **Modularity and Separation of Concerns**:
     -   Clear distinction between static definitions (`AIWorkerDefinition`, `GlobalDataSourceDefinition`, `AIWorkerPoolDefinition`, `WorkQueueDefinition`, `WorkItemDefinition`) and dynamic entities (`AIWorkerInstance`, runtime pools/queues, `WorkItem`s).
     -   Centralized management of global definitions promotes reusability.
 -   **State Management and Persistence**: All configurations are persisted. Task persistence is optional per queue.
 -   **Extensibility**: `DataSourceType` allows future data source types. SAI integration is designed for future expansion.
 -   **Resource Pooling (NEW)**: Efficiently manages and reuses worker instances.
 -   **Asynchronous Task Processing (NEW)**: Decouples task submission from execution via queues.
 -   **Task Templating (NEW)**: `WorkItemDefinition`s allow for standardization and simplification of repetitive task submissions.
 -   **Context Layering (NEW)**: Provides a flexible and hierarchical way to define data access, tool permissions, and supervision across tasks, instances, pools, queues, and worker definitions.
 -   **Monitoring and Control**: Performance tracking, rate limiting, worker-specific tool permissions, controlled data source access.
 -   **Integration with NeuroScript**: System primarily controlled via NeuroScript tools.
 -   **Concurrency Safety**: Maintained for all shared structures within `AIWorkerManager`.

 ## 8. Security Considerations for DataSources and Tool Permissions
 *(This section remains critical and largely as defined previously, emphasizing external LocalPath validation, how definition-specific tool permissions are enforced by the SecurityLayer possibly by being passed the worker's specific permission set for a given call, and the strict sandboxing of all direct file writes.)*

 ---
 ## 9. Context Resolution (DataSources and Supervisory AI)

 With `DataSourceRefs` and `SupervisoryAIRef` potentially defined at multiple levels (`AIWorkerDefinition`, `WorkQueueDefinition`, `AIWorkerPoolDefinition`, `AIWorkerInstance`, `WorkItem`, and now `WorkItemDefinition`), a clear resolution strategy is needed to determine the "effective" context for a task. The general principle is that more specific contexts can augment or, in some cases (like SAI), override more general contexts.

 **9.1. Effective DataSources**
 -   When a `WorkItem` is being prepared for execution by an `AIWorkerInstance`, the system determines the **effective set of `GlobalDataSourceDefinition` names** through an **additive merge**:
     1.  Start with `DataSourceRefs` from the instance's `AIWorkerDefinition`.
     2.  Merge (add unique) `DataSourceRefs` from the `WorkQueueDefinition` (if the task is from a queue).
     3.  Merge (add unique) `DataSourceRefs` from the `AIWorkerPoolDefinition` (if the instance is part of a pool).
     4.  Merge (add unique) `DataSourceRefs` from the `AIWorkerInstance` itself (for dynamically attached sources, if any).
     5.  Merge (add unique) `DataSourceRefs` from the `WorkItemDefinition` (if the `WorkItem` was created from one, these are its defaults).
     6.  Finally, merge (add unique) `DataSourceRefs` from the `WorkItem`'s own `DataSourceRefs` (these act as specific additions or overrides to what its `WorkItemDefinition` might have provided).
 -   This produces a unique list of `GlobalDataSourceDefinition` names. These names are then resolved by the `AIWorkerManager` into the full `GlobalDataSourceDefinition` objects. These resolved objects constitute the `ResolvedDataSources` available to the instance for that task.
 -   **Uniqueness**: `GlobalDataSourceDefinition.Name` must be globally unique within the `AIWorkerManager`'s registry. The merging process collects unique names.
 -   **File Access**: Tools accessing files (e.g., via the future `datasource://<DataSourceName>/path` URI) will use this effective set. `<DataSourceName>` must be one of the names in the `ResolvedDataSources`.

 **9.2. Effective Supervisory AI (Future)**
 -   A hierarchical override model is used to determine the single effective `SupervisoryAIRef` for a given context (e.g., a task execution):
     1.  `WorkItem.SupervisoryAIRef` (highest precedence)
     2.  `WorkItemDefinition.DefaultSupervisoryAIRef` (if `WorkItem` used a definition and its own ref is empty)
     3.  `AIWorkerInstance.SupervisoryAIRef`
     4.  `AIWorkerPoolDefinition.SupervisoryAIRef` (if instance is part of a pool)
     5.  `WorkQueueDefinition.SupervisoryAIRef` (if task is from a queue)
     6.  `AIWorkerDefinition.DefaultSupervisoryAIRef` (lowest precedence for the instance's type)
 -   The first non-empty `SupervisoryAIRef` found in this order is chosen as the effective SAI. This reference would typically be the name of an `AIWorkerDefinition` suitable for supervision, or the ID of a specific, active SAI worker instance.
 -   The `AIWorkerManager` (or a dedicated monitoring/eventing subsystem) is responsible for routing "key info" (e.g., `PerformanceRecord`s, status updates, errors) from the relevant component (task, instance, pool, queue) to its effective SAI.

 **9.3. Computer (`ng`) vs. SAI Management of Queues/Pools**
 -   **Primary Management**: `WorkQueueDefinition`s, `AIWorkerPoolDefinition`s, and `WorkItemDefinition`s are created, configured, and their lifecycles managed primarily through NeuroScript tools, typically invoked by an `ng` instance or a master control script. This includes defining associations, submitting tasks, etc.
 -   **SAI Role**: The `SupervisoryAIRef` fields designate an SAI to act as an observer or a more advanced feedback/control loop. The SAI does not replace `ng` for fundamental administrative control but rather consumes information and can potentially react by:
     * Logging/Reporting: Analyzing trends, errors, performance.
     * Generating `SupervisorFeedback` for `PerformanceRecord`s.
     * Alerting: Notifying human operators of critical issues.
     * Reactive Control (Advanced): If permitted by its own toolset and `AIWorkerDefinition`, an SAI might execute NeuroScript tools to, for example, adjust queue priorities, suggest scaling a pool, or even re-submit a failed task with modifications. This interaction is still mediated through the standard NeuroScript tool execution framework.

 ---
 ## 10. Document Metadata

 :: version: 0.4.0
 :: type: NSproject
 :: subtype: design_document
 :: project: NeuroScript
 :: purpose: Describes the design for the AI Worker Management (ai_wm_*) subsystem, including worker definitions, instances, global data sources, worker pools, work queues, work item definitions, and supervisory AI provisions.
 :: status: under_review
 :: author: AJP (Primary), Gemini (Contributor)
 :: created: 2025-05-09
 :: modified: 2025-05-10
 :: dependsOn: pkg/core/ai_worker_types.go, pkg/core/ai_wm.go, (conceptual: pkg/core/ai_wm_datasources.go, pkg/core/ai_wm_pools.go, pkg/core/ai_wm_queues.go, pkg/core/ai_wm_item_defs.go)
 :: reviewCycle: 3
 :: nextReviewDate: 2025-05-17
 ---