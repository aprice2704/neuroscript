 # NeuroScript v0.5.0 Development Tasks

 ## Vision (Summary)

 - Worker Management System able to perform symphonic tasks, with robust configuration, operational controls, and foundational learning capabilities.
 - gopls integration allows fast, efficient, and reliable AI-driven Go code development.

## Language

- [ ] Fix persisting fns
- [ ] add final and friends (persisting_fns.md)
- [ ] add keep_calm_carry_on as alias for clear_error
- [ ] add proc as alias for func, endproc for endfunc
- [ ] remove need for line continuation if easy (check at least)
- [ ] fn built in type?
 
 ## Worker Management System

 - See design document: [docs/core/ai_wm_design.md]

 ### From version 0.3.0 Review

   - |x| AI Worker Management System (Core v0.3 Functionality)
     - [x] Core AI Worker Manager (Initialization, basic structure)
     - |x| AI Worker Definition Tools
       - [x] AIWorkerDefinition.Add
       - [x] AIWorkerDefinition.Get
       - [x] AIWorkerDefinition.List
       - [x] AIWorkerDefinition.Update
       - [x] AIWorkerDefinition.Remove
       - [x] AIWorkerDefinition.LoadAll
       - [x] AIWorkerDefinition.SaveAll
     - |x| AI Worker Instance Tools
       - [x] AIWorkerInstance.Spawn
       - [x] AIWorkerInstance.Get
       - [x] AIWorkerInstance.ListActive
       - [x] AIWorkerInstance.Retire
       - [x] AIWorkerInstance.UpdateStatus
       - [x] AIWorkerInstance.UpdateTokenUsage
     - |x| AI Worker Execution Tools
       - [x] AIWorker.ExecuteStatelessTask
     - |x| AI Worker Performance Tools
       - [x] AIWorker.SavePerformanceData
       - [x] AIWorker.LoadPerformanceData
       - [x] AIWorker.LogPerformance
       - [x] AIWorker.GetPerformanceRecords
   - [-] Design for stateful worker interaction and task lifecycle. @ai_wm
     :: note: Superseded and expanded by v0.5 worker pool and queue design. Stateful interaction will leverage these new components.
   - [-] Tooling for SAI to assign/monitor tasks on workers. @ai_wm @sai
     :: note: Partially addressed by foundational SAI hooks in v0.5 (Section M of WMS enhancements), further SAI-specific tooling will be a future focus. This also ties into the "Shared Memory & Lessons Learned" (Section N).
   - [-] Agent mode permitted, allow & deny lists @ai_wm @security
     :: note: Addressed by v0.5 enhancements: definition-specific ToolAllowlist/ToolDenylist (Section A.10, E, K.1 of WMS enhancements). Agent mode itself is a higher-level application concern using these features.

 ### New for version 0.5.0

 - | | 6. AI Worker Management System Enhancements (v0.5 Focus) @ai_wm @core @roadmap_v0.5
   :: description: Implement advanced features for AI Worker Management including global data sources, work item definitions, worker pools, work queues, configurable loading policies, enhanced context resolution, and foundational support for lessons learned.
   :: dependsOn: docs/core/ai_wm_design.md

   - | | **A. Foundational Type Definitions (`pkg/core/ai_worker_types.go`)**
     - [ ] A.1. Define `GlobalDataSourceDefinition` struct (Name, Type, Description, LocalPath, AllowExternalReadAccess, FileAPIPath, RemoteTargetPath, ReadOnly, Filters, Recursive, Metadata, Timestamps).
     - [ ] A.2. Define `DataSourceType` enum (`DataSourceTypeLocalDirectory`, `DataSourceTypeFileAPI`).
     - [ ] A.3. Define `InstanceRetirementPolicy` struct (MaxTasksPerInstance, MaxInstanceAgeHours).
     - [ ] A.4. Define `AIWorkerPoolDefinition` struct (PoolID, Name, TargetAIWorkerDefinitionName, MinIdleInstances, MaxTotalInstances, InstanceRetirementPolicy, DataSourceRefs, SupervisoryAIRef, IsMissionCritical, Metadata, Timestamps).
     - [ ] A.5. Define `RetryPolicy` struct (MaxRetries, RetryDelaySeconds).
     - [ ] A.6. Define `WorkQueueDefinition` struct (QueueID, Name, AssociatedPoolNames, DefaultPriority, RetryPolicy, PersistTasks, DataSourceRefs, SupervisoryAIRef, IsMissionCritical, Metadata, Timestamps).
     - [ ] A.7. Define `WorkItemDefinition` struct (WorkItemDefinitionID, Name, Description, DefaultTargetWorkerCriteria, DefaultPayloadSchema, DefaultDataSourceRefs, DefaultPriority, DefaultSupervisoryAIRef, Metadata, Timestamps).
     - [ ] A.8. Define `WorkItemStatus` enum (e.g., `Pending`, `Processing`, `Completed`, `Failed`, `Retrying`, `Cancelled`).
     - [ ] A.9. Define `WorkItem` struct (TaskID, WorkItemDefinitionName?, QueueName, TargetWorkerCriteria, Payload, DataSourceRefs, Priority, Status, Timestamps, RetryCount, Result, Error, PerformanceRecordID, SupervisoryAIRef, Metadata).
     - [ ] A.10. Modify `AIWorkerDefinition` to include `DataSourceRefs []string`, `ToolAllowlist []string`, `ToolDenylist []string`, and `DefaultSupervisoryAIRef string`.
     - [ ] A.11. Modify `AIWorkerInstance` to include `PoolID string`, `CurrentTaskID string`, instance-level `DataSourceRefs []string`, `SupervisoryAIRef string`, `CurrentOperationalLog []OperationalLogEntry` (runtime), and `LastTaskSummary string`.
     - [ ] A.12. Define `ConfigLoadPolicy` enum for `AIWorkerManager` (e.g., `FailFastOnError`, `LoadValidAndReportErrors`).
     - [ ] A.13. Define `AIWorkerManagementConfigBundle` struct for omni-loading.
     - [ ] A.14. Add `WorkerGeneratedSummary string` and optional `OperationalLogRef string` fields to `PerformanceRecord` struct.
     - [ ] A.15. Define `DistilledLesson` struct (LessonID, Title, ApplicabilityContext, Insight, ActionableSteps, Confidence, EvidenceLinks, CreatedBySAI, Version, Timestamp) - for future use, define type now.
     - [ ] A.16. (NEW) Define `OperationalLogEntry` struct (Timestamp, StepName, ObservationDecision, ToolCallDetails, IntermediateResultHash, KeyInsight) - for `AIWorkerInstance.CurrentOperationalLog`.

   - | | **B. AIWorkerManager Core Logic Updates (`pkg/core/ai_wm.go`)**
     - [ ] B.1. Implement `ConfigLoadPolicy` setting within `AIWorkerManager` and its effect on `LoadConfigBundleFromString`.
         - [ ] B.1a. Implement "Validate Phase" for `LoadConfigBundleFromString`: validate all definitions in a bundle (structural, security, inter-ref within bundle & existing) before activating any. Collect all errors.
         - [ ] B.1b. Implement "Activate Phase" for `LoadConfigBundleFromString`: commit validated definitions to live registries and persist, based on `ConfigLoadPolicy` and validation results.
     - [ ] B.2. Implement internal `validate<Type>DefinitionUnsafe` methods for each definition type.
     - [ ] B.3. Implement internal `activate<Type>DefinitionUnsafe` methods for each definition type (assigns IDs, timestamps, stores in map, calls persist).
     - [ ] B.4. Design and implement the **Task Dispatcher** component/logic.
       - [ ] B.4a. Monitor queues for pending `WorkItem`s.
       - [ ] B.4b. Identify compatible `AIWorkerPool`s for a `WorkItem`, considering `IsMissionCritical` status.
       - [ ] B.4c. Request idle instances from pools.
       - [ ] B.4d. Assign `WorkItem`s to instances, assembling and providing full effective context (including future knowledge injection).
     - [ ] B.5. Refactor or adapt existing `ExecuteStatelessTask` logic to align with the new queueing system.
     - [ ] B.6. Ensure `AIWorkerManager` correctly initializes its own configuration (like `ConfigLoadPolicy`) and manages runtime states of all definitions, pools, and queues.

   - | | **C. Global DataSource Management (`pkg/core/ai_wm_datasources.go` - new conceptual file, logic in `ai_wm.go`)**
     - [ ] C.1. Implement CRUD logic (validate, activate, store, persist) in `AIWorkerManager` for `GlobalDataSourceDefinition`s.
     - [ ] C.2. Implement security validation for `GlobalDataSourceDefinition.LocalPath` with `AllowExternalReadAccess` (against admin whitelist).

   - | | **D. WorkItemDefinition Management (`pkg/core/ai_wm_item_defs.go` - new conceptual file, logic in `ai_wm.go`)**
     - [ ] D.1. Implement CRUD logic in `AIWorkerManager` for `WorkItemDefinition`s.

   - | | **E. AIWorkerDefinition Enhancements (Logic in `ai_wm_definitions.go`, `ai_wm.go`)**
     - [ ] E.1. Update `AIWorkerManager` CRUD logic for `AIWorkerDefinition` to handle new fields: `DataSourceRefs`, `ToolAllowlist`, `ToolDenylist`, `DefaultSupervisoryAIRef`.
     - [ ] E.2. Ensure persistence for `AIWorkerDefinition` includes these new fields.

   - | | **F. AIWorkerPool Implementation (`pkg/core/ai_wm_pools.go` - new conceptual file, logic in `ai_wm.go`)**
     - [ ] F.1. Implement CRUD logic in `AIWorkerManager` for `AIWorkerPoolDefinition`s (including `IsMissionCritical`).
     - [ ] F.2. Develop runtime `AIWorkerPool` management logic within `AIWorkerManager` (instance tracking, scaling, retirement, providing idle instances).

   - | | **G. WorkQueue & WorkItem Implementation (`pkg/core/ai_wm_queues.go` - new conceptual file, logic in `ai_wm.go`)**
     - [ ] G.1. Implement CRUD logic in `AIWorkerManager` for `WorkQueueDefinition`s (including `IsMissionCritical`).
     - [ ] G.2. Develop runtime `WorkQueue` management logic within `AIWorkerManager` (holding `WorkItem`s - in-memory for Phase 1).
         - [ ] G.2a. Logic for constructing a `WorkItem` (assigning `TaskID`, `SubmitTimestamp`, `Status=Pending`) using an optional `WorkItemDefinitionName` as a template, merging/overriding with explicitly provided fields.
     - [ ] G.3. Implement logic for adding `WorkItem`s to a queue and retrieving them for the dispatcher.
     - [ ] G.4. Logic for `AIWorkerManager` (or instance via callback) to update `WorkItem` status, result, error, timestamps, and link to `PerformanceRecordID`.
     - [ ] G.5. (Future Phase) Implement `WorkItem` persistence if `WorkQueueDefinition.PersistTasks` is true.

   - | | **H. Context Resolution Logic (`pkg/core/ai_wm.go`, task execution path)**
     - [ ] H.1. Implement logic in `AIWorkerManager` (Task Dispatcher) to resolve **Effective DataSources** for a task (additive merge: `WorkItem` -> `WorkItemDefinition` -> `AIWorkerInstance` -> `AIWorkerPoolDefinition` -> `WorkQueueDefinition` -> `AIWorkerDefinition`). Store/pass as `ResolvedDataSources` to instance.
     - [ ] H.2. Ensure `AIWorkerInstance` execution path uses `ResolvedDataSources`.
     - [ ] H.3. Ensure tool permissions (`ToolAllowlist`/`ToolDenylist` from `AIWorkerDefinition`) are available to `SecurityLayer` for the current task context.

   - | | **I. Persistence for New Definitions (`pkg/core/ai_wm.go`)**
     - [ ] I.1. Implement persistence for `GlobalDataSourceDefinition`s to/from `ai_global_data_source_definitions.json`.
     - [ ] I.2. Implement persistence for `WorkItemDefinition`s to/from `ai_work_item_definitions.json`.
     - [ ] I.3. Implement persistence for `AIWorkerPoolDefinition`s to/from `ai_worker_pool_definitions.json`.
     - [ ] I.4. Implement persistence for `WorkQueueDefinition`s to/from `ai_work_queue_definitions.json`.
     - [ ] I.5. Ensure `AIWorkerDefinition` persistence handles new fields correctly.
     - [ ] I.6. (Future) Persistence for `DistilledLesson`s to `ai_knowledge_base.json` or DB.

   - | | **J. NeuroScript Tooling (`pkg/core/ai_wm_tools_*.go` - new/modified files)**
     - [ ] J.1. **Omni-Loader Tool**:
         - [ ] `aiWorkerManager.loadConfigBundleFromString(jsonBundleString string, loadPolicyOverride? string)`
     - [ ] J.2. **Individual Loader Tools** (e.g., `aiWorkerGlobalDataSource.loadFromString`, etc. for all 5 definition types: GlobalDataSource, AIWorker, AIWorkerPool, WorkQueue, WorkItemDefinition).
     - [ ] J.3. **CRUD Tools** for all 5 definition types (`.add`, `.get`, `.list`, `.update`, `.remove` - where `add` might be covered by `loadFromString` if it handles creation, and `update` would take ID/Name and JSON string for respective fields).
     - [ ] J.4. **Pool/Queue Runtime Info Tools**: `aiWorkerPool.getInstanceStatus`, `aiWorkerWorkQueue.getStatus`.
     - [ ] J.5. **Task Management Tools**: `aiWorkerWorkQueue.submitTask` (updated for `WorkItemDefinitionName`), `aiWorkerWorkQueue.getTaskInfo`.
     - [ ] J.6. **File Sync Tool**: `aiWorker.syncDataSource`.
     - [ ] J.7. **Manager Config Tools**: `aiWorkerManager.setConfigLoadPolicy`, `aiWorkerManager.getConfigLoadPolicy`.
     - [ ] J.8. Register all new/updated tools.

   - | | **K. Security Enhancements (`pkg/core/security.go`, `pkg/core/tools_file_api.go`)**
     - [ ] K.1. Enhance `SecurityLayer`'s `ValidateToolCall` and `GetToolDeclarations` to accept/use the active worker's definition-specific `ToolAllowlist` and `ToolDenylist`.
     - [ ] K.2. Update path resolution logic in `FileAPI` and `SecurityLayer` to handle `GlobalDataSourceDefinition.LocalPath` with `AllowExternalReadAccess`, using the effective data sources context.
     - [ ] K.3. Rigorously ensure all file write tools remain confined to the primary interpreter sandbox.

   - | | **L. Supervisory AI (SAI) Integration - Foundational Hooks (Future Focus)**
     - [ ] L.1. Define event emission points for key lifecycle events.
     - [ ] L.2. Implement logic to resolve Effective SAI (including from `WorkItemDefinition`).
     - [ ] L.3. Basic event routing mechanism to the resolved SAI.
     - [ ] L.4. Ensure `SupervisorFeedback` in `PerformanceRecord` can be populated by an SAI, and that `PerformanceRecord` (including new `WorkerGeneratedSummary`) is part of the info feed to SAIs.
     - [ ] L.5. (Design for Future) Outline `KnowledgeBaseManager` component responsible for storing and retrieving `DistilledLesson`s.

   - | | **M. Shared Memory & Lessons Learned (Foundational Work for v0.5 - Full System in v0.6+)** @ai_wm @sai @learning
     :: description: Lay groundwork for workers to record operational insights and for SAIs to distill these into actionable knowledge.
     - [ ] M.1. **Worker Insight Logging & Summary**:
         - [ ] M.1a. Define NeuroScript tool `aiWorker.logOperationalInsight(level string, message string, detailsMap? map)` for workers to record significant observations during task execution.
         - [ ] M.1b. Implement logic for `AIWorkerInstance` to collect insights logged via this tool into `CurrentOperationalLog`.
         - [ ] M.1c. Implement logic for `AIWorkerInstance` to produce a `WorkerGeneratedSummary` (e.g., by concatenating logged insights or an LLM call to summarize its log) and ensure it's included in the `PerformanceRecord`.
     - [ ] M.2. **SAI Data Collection Point for Learning**:
         - [ ] M.2a. Ensure `PerformanceRecord`s, including `WorkerGeneratedSummary` and any `SupervisorFeedback`, are made available to the designated SAI (e.g., via event feed or queryable mechanism).
     - [ ] M.3. **Knowledge Base (Design Only for v0.5)**:
         - [ ] M.3a. Finalize the design for the `KnowledgeBaseManager` component and the structure of `DistilledLesson`s.
         - [ ] M.3b. Define the API for SAIs to submit lessons and for workers/dispatchers to query lessons (tool stubs, no backend implementation in v0.5).
     - [ ] M.4. **Contextual Knowledge Injection (Design Only for v0.5)**:
         - [ ] M.4a. Design how the Task Dispatcher or `AIWorkerInstance` will query the (future) `KnowledgeBaseManager` and how relevant lessons will be formatted and injected into a worker's prompt/context. (No implementation in v0.5)

   - | | **N. Testing** (Renumbered from M)
     - [ ] N.1. Unit tests for CRUD and validation of all 5 definition types (GDS, AWD, WID, AWPD, WQD - including `IsMissionCritical`, external path security).
     - [ ] N.2. Unit tests for `AIWorkerPool` runtime logic.
     - [ ] N.3. Unit tests for `WorkQueue` runtime logic (including item creation from `WorkItemDefinition`).
     - [ ] N.4. Unit tests for `AIWorkerManager.LoadConfigBundleFromString` with different `ConfigLoadPolicy` values and error conditions.
     - [ ] N.5. Unit tests for Task Dispatcher logic.
     - [ ] N.6. Unit tests for Context Resolution (DataSources, Tool Permissions, basic SAI ref, considering all definition inputs).
     - [ ] N.7. Unit tests for `aiWorker.logOperationalInsight` and `WorkerGeneratedSummary` population in `PerformanceRecord`.
     - [ ] N.8. Integration tests: task submission (using `WorkItemDefinition`) to queue, execution by pooled worker, result retrieval, and `PerformanceRecord` checks.
     - [ ] N.9. Tests for all new/updated NeuroScript tools for AI WM.
     - [ ] N.10. Security tests: `AllowExternalReadAccess` validation, tool permissions, write sandboxing.

 ## gopls Integration for Advanced Go Development (v0.5 Focus) @gopls @core @codegen
   :: description: Enable NeuroScript to leverage gopls for precise Go code diagnostics and semantic understanding, empowering AI to write compiling code reliably.
   :: objective: Allow an AI to write compiling Go code **without fail** as quickly and efficiently as possible using gopls feedback.
   :: dependsOn: docs/gopls_integration_design.md, pkg/core/tools/gopls_client.go (new), pkg/core/tools_gopls.go (new)

   - | | **A. LSP Client Infrastructure (`pkg/core/tools/gopls_client.go` - new file/package)**
     :: description: Establish foundational communication with a gopls server instance.
     - [ ] A.1. Design `GoplsClient` interface and basic implementation.
         - [ ] A.1a. Method to start and manage a `gopls` subprocess.
         - [ ] A.1b. Implement JSON-RPC 2.0 stream handling for LSP messages.
         - [ ] A.1c. Basic LSP request/response/notification handling logic.
     - [ ] A.2. Implement LSP `initialize` / `initialized` handshake sequence.
         - [ ] A.2a. Send `initialize` request with client capabilities.
         - [ ] A.2b. Process `initialize` response from gopls and store server capabilities.
         - [ ] A.2c. Send `initialized` notification.
     - [ ] A.3. Implement LSP `shutdown` / `exit` sequence for graceful termination.
     - [ ] A.4. Basic error handling and logging for LSP communication issues.

   - | | **B. Workspace and Document Synchronization**
     :: description: Ensure gopls has an accurate view of the workspace and file contents.
     - [ ] B.1. Implement `GoplsClient` methods for LSP `textDocument/didOpen` notification.
         - [ ] B.1a. Tool `Gopls.NotifyDidOpen (filePath string, content string)`.
     - [ ] B.2. Implement `GoplsClient` methods for LSP `textDocument/didChange` notification.
         - [ ] B.2a. Tool `Gopls.NotifyDidChange (filePath string, newContent string)`.
         - [ ] B.2b. (Future) Support for incremental changes.
     - [ ] B.3. Implement `GoplsClient` methods for LSP `textDocument/didSave` notification.
         - [ ] B.3a. Tool `Gopls.NotifyDidSave (filePath string)`.
     - [ ] B.4. Implement `GoplsClient` methods for LSP `textDocument/didClose` notification.
         - [ ] B.4a. Tool `Gopls.NotifyDidClose (filePath string)`.
     - [ ] B.5. Tool `Gopls.SetWorkspaceRoot (workspacePath string)`.

   - | | **C. Diagnostic Tools (`pkg/core/tools_gopls_diags.go` - new conceptual file)**
     :: description: Provide NeuroScript tools to fetch and interpret diagnostics from gopls.
     - [ ] C.1. Define `DiagnosticInfo` struct in Go (mirroring LSP `Diagnostic`).
     - [ ] C.2. Implement `GoplsClient` logic to receive/parse `textDocument/publishDiagnostics`.
         - [ ] C.2a. Store diagnostics per file URI.
     - [ ] C.3. NeuroScript Tool: `Gopls.GetDiagnostics (filePath string) (diagnostics []DiagnosticInfo)`.
     - [ ] C.4. NeuroScript Tool: `Gopls.GetAllProjectDiagnostics () (map[string][]DiagnosticInfo)`.

   - | | **D. Semantic Information & Contextual Tools (AI-Focused)** @gopls @semantic
     :: description: Provide tools for AI to get rich semantic information about Go code elements related to diagnostics, using stable addressing.
     - [ ] D.1. Define NeuroScript internal representation for "SemanticReference".
         - [ ] D.1a. `ResourceURI` (string).
         - [ ] D.1b. `FullyQualifiedName` (string, optional).
         - [ ] D.1c. `SymbolKind` (string, optional).
         - [ ] D.1d. `DeclarationLocation` (`LSPLocation`, optional, precise byte offsets).
         - [ ] D.1e. `ByteOffsetRange` (`LSPRange`, optional, precise byte offsets).
         - [ ] D.1f. `Signature` (string, optional).
         - [ ] D.1g. `PackagePath` (string, optional).
     - [ ] D.2. Enhance `DiagnosticInfo` (from C.1) to include `SemanticReferenceGuess`.
     - [ ] D.3. NeuroScript Tool: `Gopls.GetSymbolInfoAt(filePath string, byteOffset int) (symbolInfo map)`.
         - [ ] D.3a. Internal LSP `textDocument/hover` and `textDocument/definition`.
         - [ ] D.3b. Convert responses to `SemanticReference` maps.
         - [ ] D.3c. Consistent byte offset handling.
     - [ ] D.4. NeuroScript Tool: `Gopls.GetSymbolInfoByName(fullyQualifiedName string) (symbolInfo map)`.
         - [ ] D.4a. Investigate gopls `workspace/symbol` or sequence of calls.
         - [ ] D.4b. Synergize with `pkg/core/tools/gosemantic`.
     - [ ] D.5. NeuroScript Tool: `Gopls.ListSymbolsInFile(filePath string) (symbols []map)`.
     - [ ] D.6. NeuroScript Tool: `Gopls.FindWorkspaceSymbols(queryString string) (symbols []map)`.
     - [ ] D.7. Ensure tools return `SemanticReference` format.

   - | | **E. NeuroScript Toolset Registration (`pkg/core/tools_gopls.go` - new file)**
     - [ ] E.1. Create `RegisterGoplsTools(*core.Interpreter)`.
     - [ ] E.2. Instantiate and manage `GoplsClient` lifecycle.
     - [ ] E.3. Register all new `Gopls.*` tools.

   - | | **F. Testing (gopls)**
     - [ ] F.1. Unit tests for `GoplsClient` (mocking LSP).
     - [ ] F.2. Integration tests with real `gopls` server:
         - [ ] F.2a. `initialize` handshake.
         - [ ] F.2b. Document sync notifications and diagnostic updates.
         - [ ] F.2c. `Gopls.GetDiagnostics` with known errors.
         - [ ] F.2d. `Gopls.GetSymbolInfoAt/ByName` on known symbols.
         - [ ] F.2e. `Gopls.ListSymbolsInFile` and `Gopls.FindWorkspaceSymbols`.
     - [ ] F.3. Test error handling for gopls communication.

 ## Other bits

 - [ ] Matrix tools @lang_features @unprioritized
 - [ ] Named arguments @lang_features @unprioritized
 - [ ] Default arguments @lang_features @unprioritized
 - [ ] Line continuation etc. @lang_features @unprioritized



 :: title: NeuroScript Road to v0.5.0 Checklist
 :: version: 0.5.0
 :: id: ns-roadmap-v0.5.0
 :: status: draft
 :: description: Prioritized tasks to implement NeuroScript v0.5.0 based on design discussions and Vision (May 10, 2025), including foundations for lessons learned and gopls integration for AI-driven Go development.
 :: updated: 2025-05-10