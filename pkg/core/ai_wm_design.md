# AI Worker Management (ai_wm_*) System Design

## 1. Overview

The AI Worker Management system (ai_wm_*) within the pkg/core package of NeuroScript provides a comprehensive framework for defining, managing, executing, and monitoring AI-powered workers. These workers typically represent Large Language Models (LLMs) or other model-based agents. The system is designed to support both stateful, instance-based interactions and stateless, one-shot task executions, potentially managed through worker pools and work queues.

**Key Change (May 2025)**: `AIWorkerDefinition`s are now treated as **immutable after initial load**. They are loaded once (e.g., at startup or via a specific "reload all definitions" command) from a configuration source (like `ai_worker_definitions.json`) and are not subsequently modified, added to, or removed from memory by runtime operations, nor are they saved back. This simplifies state management significantly.

Core features include:
-   Loading worker definitions;
-   Flexible and shared data source configurations with controlled external file access capabilities;
-   Definitions for templatizing work items (`WorkItemDefinition`);
-   API key management;
-   Worker-specific tool allow/deny lists;
-   Configurable rate limiting;
-   Detailed performance tracking;
-   Provisions for Supervisory AI (SAI) attachment for monitoring and feedback;
-   **(NEW/Future Focus)** A mechanism for accumulating, distilling, and utilizing shared operational knowledge and lessons learned to improve ongoing operations.

The system exposes its functionalities to NeuroScript via a dedicated toolset.

## 2. Core Components and Concepts

### 2.1. AIWorkerManager (ai_wm.go)

**Role**: This is the central orchestrator and entry point for the entire AI Worker Management system. It is responsible for the lifecycle management of:
-   `AIWorkerDefinition`s (blueprints for workers, **loaded and treated as immutable**).
-   `GlobalDataSourceDefinition`s: Centrally defined data sources.
-   `AIWorkerPoolDefinition`s and active `AIWorkerPool`s: For managing groups of worker instances.
-   `WorkQueueDefinition`s and active `WorkQueue`s: For managing task submission and dispatch.
-   `WorkItemDefinition`s: For templatizing tasks.
-   Active `AIWorkerInstance`s (live, stateful worker sessions, potentially managed by pools).
-   **(NEW/Future)** `KnowledgeBaseManager` for storing and managing `DistilledLesson`s.

It handles persistence of configurations (excluding `AIWorkerDefinition`s post-initial load), performance data, enforces rate limits, resolves API keys, and dispatches tasks from queues to pools.

**Configuration**:
-   The `AIWorkerManager` itself will have configurable behaviors, potentially set during initialization or via specific NeuroScript tools:
    -   **`ConfigLoadPolicy` (enum, e.g., `FailFastOnError`, `LoadValidAndReportErrors`)**: Determines how the `LoadConfigBundleFromString` tool behaves.
        -   `FailFastOnError`: If any definition in a bundle fails validation, the entire bundle loading operation is aborted, and no changes are activated.
        -   `LoadValidAndReportErrors`: Activates all valid definitions from the bundle and reports errors for those that failed validation. This is generally more user-friendly for iterative setup.
    -   Whitelist of allowed base paths for `GlobalDataSourceDefinition.LocalPath` when `AllowExternalReadAccess` is true. This is a critical security setting configured at the system/deployment level, not typically via dynamic NeuroScript calls.

**State**: The manager maintains in-memory collections of:
-   `AIWorkerDefinition`s (**immutable after load**).
-   Active `AIWorkerInstance`s.
-   `WorkerRateTracker`s (runtime counters for rate limiting per definition).
-   `GlobalDataSourceDefinition`s.
-   `AIWorkerPoolDefinition`s and runtime `AIWorkerPool` states.
-   `WorkQueueDefinition`s and runtime `WorkQueue` states (including `WorkItem`s if queues are in-memory).
-   `WorkItemDefinition`s.
-   **(NEW/Future)** `DistilledLesson`s managed by the `KnowledgeBaseManager`.
-   **(Future)** References to active Supervisory AI (SAI) instances or configurations.

**Concurrency**: It utilizes a `sync.RWMutex`. Locks are primarily for mutable collections like `activeInstances`, `rateTrackers`, and dynamically loaded/managed configurations (Pools, Queues, etc.). Access to `AIWorkerDefinition`s, once loaded, is read-only and inherently thread-safe.

**Persistence**:
The manager is responsible for loading `AIWorkerDefinition`s from `ai_worker_definitions.json` at startup or on explicit reload. **It does not save `AIWorkerDefinition`s back.**
It handles loading and saving for:
-   Performance data of retired instances to/from `ai_worker_performance_data.json`.
-   Global Data Source definitions to/from `ai_global_data_source_definitions.json`.
-   `AIWorkerPoolDefinition`s to/from `ai_worker_pool_definitions.json`.
-   `WorkQueueDefinition`s to/from `ai_work_queue_definitions.json`.
-   `WorkItemDefinition`s to/from `ai_work_item_definitions.json`.
-   **(NEW/Future)** `DistilledLesson`s to/from `ai_knowledge_base.json` (or a more structured store).
-   **(Future)** `WorkItem`s if queue persistence is enabled (e.g., to `ai_work_items.db`).
These files are stored within a specified sandbox directory.

**Initialization**:
The `NewAIWorkerManager` constructor initializes the manager, loads its own operational configuration (like `ConfigLoadPolicy`), attempts to load all defined configurations (worker definitions, data sources, work item definitions, pools, queues, **distilled lessons (Future)**) and historical performance data from the sandbox, and sets up initial rate trackers for each loaded worker definition. `AIWorkerDefinition`s are loaded at this stage.

---
### 2.2. Global Data Source Management (ai_wm_datasources.go - *new conceptual file*)

This component, managed by the `AIWorkerManager`, handles the definition and persistence of shared data sources that AI Workers can reference. (Management and persistence of these definitions remain as previously designed).

**GlobalDataSourceDefinition (`ai_worker_types.go`)**:
(Structure remains as previously defined)
-   **`Name`** (string): A unique, human-readable name/ID for this data source.
-   **`Type`** (DataSourceType): Indicates the nature of the data source.
-   ... (other fields as previously defined) ...
-   **`CreatedTimestamp`** (time.Time, omitempty): Set by manager on creation.
-   **`ModifiedTimestamp`** (time.Time, omitempty): Set by manager on update.

**Management**: The `AIWorkerManager` will provide CRUD operations for `GlobalDataSourceDefinition`s. All changes will be persisted to `ai_global_data_source_definitions.json`.

**Security**: Validation of `LocalPath` for `AllowExternalReadAccess` remains critical.
---

### 2.3. AIWorkerDefinition (ai_worker_types.go, ai_wm_definitions_*.go)
**Role**: Serves as a static blueprint or template that defines all configurable aspects of a particular type of AI worker. **`AIWorkerDefinition`s are loaded at manager initialization (or via a full reload command) and are treated as immutable thereafter. They are not added, updated, or removed individually during runtime, nor are they saved back by the manager.**

**Key Attributes**:
(Attributes remain as previously defined. The `AggregatePerformanceSummary` within a definition is updated by the system with operational metrics but the definition's core configuration is immutable.)
-   **`DefinitionID`** (string): System-generated UUID, primary key.
-   **`Name`** (string): User-provided, unique, human-readable name.
-   ... (Provider, ModelName, Auth, InteractionModels, Capabilities, BaseConfig, CostMetrics, RateLimits, Status, DefaultFileContexts, DataSourceRefs, ToolAllowlist, ToolDenylist, DefaultSupervisoryAIRef) ...
-   **`AggregatePerformanceSummary`** (*AIWorkerPerformanceSummary, omitempty): Aggregated performance metrics, managed by the system. This part of the structure *is* updated in memory.
-   **`Metadata`** (map[string]interface{}, omitempty): Flexible key-value store for additional custom information.
-   **`CreatedTimestamp`** (time.Time, omitempty): Reflects load time or original definition time.
-   **`ModifiedTimestamp`** (time.Time, omitempty): Reflects load time or original definition time.

**Management**:
-   `AIWorkerDefinition`s are loaded by the `AIWorkerManager` from `ai_worker_definitions.json` (or a string bundle).
-   **No runtime CRUD operations (add, update, remove individual definitions) are supported.** To change definitions, the entire set is reloaded.
-   Access to definitions is read-only. Functions like `GetWorkerDefinition` and `ListWorkerDefinitions` provide access to the loaded (immutable) definition data. Copies might be returned to ensure callers don't accidentally attempt to modify the manager's internal state, especially for the mutable `AggregatePerformanceSummary` part.

---
### 2.4. AIWorkerPools (ai_wm_pools.go - *new conceptual file*)
(Management and persistence of `AIWorkerPoolDefinition`s remain as previously designed.)
---
### 2.5. WorkItemDefinition (ai_worker_types.go - new type)
(Management and persistence of `WorkItemDefinition`s remain as previously designed.)
---
### 2.6. WorkQueues & WorkItems (ai_wm_queues.go - *new conceptual file*)
(Management and persistence of `WorkQueueDefinition`s and `WorkItem`s remain as previously designed.)
---
### 2.7. AIWorkerInstance (ai_worker_types.go, ai_wm_instances.go)
(Structure and management remain as previously designed. Instances are dynamic and stateful.)
---
### 2.8. Task Execution & Dispatching
(Process remains largely the same. Context resolution will use the immutably loaded `AIWorkerDefinition`s.)
---
### 2.9. Performance Tracking (ai_worker_types.go, ai_wm_performance.go)
(Remains as previously designed. `PerformanceRecord`s update the `AggregatePerformanceSummary` in the in-memory `AIWorkerDefinition`.)
---
### 2.10. Rate Limiting (ai_worker_types.go, ai_wm_ratelimit.go)
(Remains as previously designed, based on the loaded `AIWorkerDefinition.RateLimits`.)
---
### 2.11. API Key Management (ai_worker_types.go, ai_wm.go)
(Remains as previously designed, based on the loaded `AIWorkerDefinition.Auth`.)
---
### 2.12. Knowledge Base & Lessons Learned (Future Focus)
(Remains as previously designed.)
---
## 3. Data Structures and Types (Key Types from `ai_worker_types.go`)
(No changes to the list of types themselves, but the handling of `AIWorkerDefinition` is different.)
---
## 4. Persistence Strategy

**Format**: JSON for definitions. (Future: consider embedded DB for `WorkItem`s if `PersistTasks` is true and volume is high, and for `DistilledLesson`s).

**Files** (in `sandboxDir`):
-   `ai_worker_definitions.json` (**Read-only by the application after initial load/reload**).
-   `ai_worker_performance_data.json` (Written to by the application).
-   `ai_global_data_source_definitions.json` (Written to).
-   `ai_worker_pool_definitions.json` (Written to).
-   `ai_work_queue_definitions.json` (Written to).
-   `ai_work_item_definitions.json` (Written to).
-   **(Future)** `ai_knowledge_base.json` (or DB) for storing `DistilledLesson`s (Written to).

**Loading/Saving**:
-   `AIWorkerDefinition`s are loaded by `AIWorkerManager` at startup or via a full reload mechanism (e.g., `aiWorkerDefinition.loadFromFile()` or from a bundle). **They are NOT saved back.**
-   Other definition types (DataSources, Pools, Queues, WorkItemDefinitions) are loaded and saved by `AIWorkerManager`, respecting its `ConfigLoadPolicy`. Individual `add/update/remove` operations for these types trigger their respective persistence.

---
## 5. NeuroScript Tool Integration (ai_wm_tools.go)

**Tool Categories & Key Tools**:
-   **AIWorkerDefinition Management**:
    -   `aiWorkerDefinition.loadFromFile()`: Reloads all definitions from the `ai_worker_definitions.json` file. This is the primary mechanism for updating definitions.
    -   `aiWorkerDefinition.loadFromString(jsonString string) (error string)`: Similar to `loadFromFile`, but takes a JSON string containing an array of definitions. Replaces all existing definitions.
    -   `aiWorkerDefinition.get(definitionNameOrID string) (definition map, error string)`: Retrieves a specific loaded definition.
    -   `aiWorkerDefinition.list(filterJSON? string) (definitionsList []map, error string)`: Lists loaded definitions.
    -   **REMOVED**: Tools for adding, updating, removing, or saving individual `AIWorkerDefinition`s.
-   **Global Data Source Management**: (CRUD tools remain, as these are mutable and persisted)
    -   `aiWorkerGlobalDataSource.loadFromString(jsonString string) (name string, error string)` (and other CRUD)
-   **Work Item Definition Management**: (CRUD tools remain)
    -   `aiWorkerWorkItemDefinition.loadFromString(jsonString string) (id string, error string)` (and other CRUD)
-   **Worker Pool Management**: (CRUD tools remain)
    -   `aiWorkerPoolDefinition.loadFromString(jsonString string) (id string, error string)` (and other CRUD including `isMissionCritical` params)
    -   `aiWorkerPool.getInstanceStatus(poolName string) (statusMap map, error string)`
-   **Work Queue Management**: (CRUD tools remain)
    -   `aiWorkerWorkQueueDefinition.loadFromString(jsonString string) (id string, error string)` (and other CRUD including `isMissionCritical` params)
    -   `aiWorkerWorkQueue.getStatus(queueName string) (statusMap map, error string)`
-   **Task Management**: (Tools remain)
    -   `aiWorkerWorkQueue.submitTask(queueName string, workItemDefinitionName? string, payloadOverrideJSON? string, ...)`
    -   `aiWorkerWorkQueue.getTaskInfo(taskID string) (taskInfoMap map, error string)`
-   **Instance Management (Standalone/Direct)**: (Tools remain)
-   **File Sync Tools**: (Tools remain)
-   **Performance & Logging**: (Tools remain)
-   **AIWorkerManager Configuration Tools**: (Tools remain)
-   **Omni-Loader Tool**:
    -   `aiWorkerManager.loadConfigBundleFromString(jsonBundleString string, loadPolicyOverride? string)`: This tool will still load various definition types. For `AIWorkerDefinition`s within the bundle, it will replace the current set of worker definitions.
-   **(NEW/Future) Worker Self-Reflection/Logging Tools**: (Tools remain)
-   **(NEW/Future) Knowledge Base Tools**: (Tools remain)

---
## 6. Error Handling
(Remains as previously designed.)
---
## 7. Key Design Principles and Considerations
-   **Modularity & Separation of Concerns**
-   **State Management & Persistence** (`AIWorkerDefinition`s are now immutable in memory after load, simplifying this aspect for them).
-   **Extensibility**
-   **Resource Pooling**
-   **Asynchronous Task Processing**
-   **Task Templating** (`WorkItemDefinition`)
-   **Context Layering** (for DataSources, SAI)
-   **Defensive Loading** (`ConfigLoadPolicy`, Load & Activate principle)
-   **Monitoring & Control** (Performance, Rate Limits, Permissions, Criticality flags)
-   **Adaptive Learning & Knowledge Sharing (NEW/Future)**: Via worker summaries, SAI condensation, and knowledge base.
-   **Integration with NeuroScript**
-   **Concurrency Safety** (Simplified for `AIWorkerDefinition` access, still relevant for other mutable state).
---
## 8. Security Considerations for DataSources and Tool Permissions
(Remains as previously designed.)
---
## 9. Context Resolution (DataSources and Supervisory AI)
(Remains as previously designed.)
---
## 10. Document Metadata
:: version: 0.4.3
:: type: NSproject
:: subtype: design_document
:: project: NeuroScript
:: purpose: Describes the design for the AI Worker Management (ai_wm_*) subsystem. Key change: AIWorkerDefinitions are immutable after load.
:: status: under_review
:: author: AJP (Primary), Gemini (Contributor)
:: created: 2025-05-09
:: modified: 2025-05-29
:: dependsOn: pkg/core/ai_worker_types.go, pkg/core/ai_wm.go, (conceptual: pkg/core/ai_wm_datasources.go, pkg/core/ai_wm_pools.go, pkg/core/ai_wm_queues.go, pkg/core/ai_wm_item_defs.go, pkg/core/ai_wm_knowledgebase.go)
:: reviewCycle: 5
:: nextReviewDate: 2025-06-05
---