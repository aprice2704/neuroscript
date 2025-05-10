 # NeuroScript v0.5.0 Development Tasks

 ## Vision (Summary)

 - Worker Management System able to perform symphonic tasks
 - gopls allows fast, efficient autonomous development

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
     :: note: Partially addressed by foundational SAI hooks in v0.5 (Section M of WMS enhancements), further SAI-specific tooling will be a future focus.
   - [-] Agent mode permitted, allow & deny lists @ai_wm @security
     :: note: Addressed by v0.5 enhancements: definition-specific ToolAllowlist/ToolDenylist (Section A8, D, J1 of WMS enhancements). Agent mode itself is a higher-level application concern using these features.

 ### New for version 0.5.0

 - | | 6. AI Worker Management System Enhancements (v0.5 Focus) @ai_wm @core @roadmap_v0.5
   :: description: Implement advanced features for AI Worker Management including global data sources, work item definitions, worker pools, work queues, and enhanced context resolution.
   :: dependsOn: docs/core/ai_wm_design.md

   - | | **A. Foundational Type Definitions (`pkg/core/ai_worker_types.go`)**
     - [ ] A1. Define `GlobalDataSourceDefinition` struct (Name, Type, LocalPath, AllowExternalReadAccess, FileAPIPath, RemoteTargetPath, ReadOnly, Filters, Recursive, Metadata).
     - [ ] A2. Define `DataSourceType` enum (`DataSourceTypeLocalDirectory`, `DataSourceTypeFileAPI`).
     - [ ] A3. Define `AIWorkerPoolDefinition` struct (PoolID, Name, TargetAIWorkerDefinitionName, MinIdleInstances, MaxTotalInstances, InstanceRetirementPolicy, DataSourceRefs, SupervisoryAIRef).
     - [ ] A4. Define `WorkQueueDefinition` struct (QueueID, Name, AssociatedPoolNames, DefaultPriority, RetryPolicy, PersistTasks, DataSourceRefs, SupervisoryAIRef).
     - [ ] A5. Define `WorkItemDefinition` struct (WorkItemDefinitionID, Name, Description, DefaultTargetWorkerCriteria, DefaultPayloadSchema, DefaultDataSourceRefs, DefaultPriority, DefaultSupervisoryAIRef, Metadata).
     - [ ] A6. Define `WorkItem` struct (TaskID, QueueName, WorkItemDefinitionName?, TargetWorkerCriteria, Payload, DataSourceRefs, Priority, Status, Timestamps, RetryCount, Result, Error, PerformanceRecordID, SupervisoryAIRef).
     - [ ] A7. Define `WorkItemStatus` enum (e.g., `Pending`, `Processing`, `Completed`, `Failed`, `Retrying`).
     - [ ] A8. Modify `AIWorkerDefinition` to include `DataSourceRefs []string`, `ToolAllowlist []string`, `ToolDenylist []string`, and `DefaultSupervisoryAIRef string`.
     - [ ] A9. Modify `AIWorkerInstance` to include `PoolID string`, `CurrentTaskID string`, instance-level `DataSourceRefs []string`, and `SupervisoryAIRef string`.

   - | | **B. Global DataSource Management (`pkg/core/ai_wm_datasources.go` - new file, `ai_wm.go`)**
     - [ ] B1. Implement CRUD operations in `AIWorkerManager` for `GlobalDataSourceDefinition`s.
     - [ ] B2. Add internal `loadGlobalDataSourceDefinitionsFromFileInternal` and `saveGlobalDataSourceDefinitionsToFileUnsafe` in `AIWorkerManager`.
     - [ ] B3. Implement security validation in `AIWorkerManager` for `GlobalDataSourceDefinition.LocalPath` when `AllowExternalReadAccess` is true (against a system-admin configured whitelist).

   - | | **C. WorkItemDefinition Management (`pkg/core/ai_wm_item_defs.go` - new file, `ai_wm.go`)**
     - [ ] C1. Implement CRUD operations in `AIWorkerManager` for `WorkItemDefinition`s.
     - [ ] C2. Add internal `loadWorkItemDefinitionsFromFileInternal` and `saveWorkItemDefinitionsToFileUnsafe` in `AIWorkerManager`.

   - | | **D. AIWorkerDefinition Enhancements (Logic in `ai_wm_definitions.go`)**
     - [ ] D1. Update `AIWorkerManager.AddWorkerDefinition` and `UpdateWorkerDefinition` to handle new fields: `DataSourceRefs`, `ToolAllowlist`, `ToolDenylist`, `DefaultSupervisoryAIRef`.
     - [ ] D2. Ensure persistence logic for `AIWorkerDefinition` correctly saves and loads these new fields.

   - | | **E. AIWorkerPool Implementation (Phase 1 - Stateless Focus) (`pkg/core/ai_wm_pools.go` - new file, `ai_wm.go`)**
     - [ ] E1. Implement `AIWorkerPoolDefinition` CRUD operations in `AIWorkerManager`.
     - [ ] E2. Develop runtime `AIWorkerPool` struct/logic for managing a collection of `AIWorkerInstance`s based on a `AIWorkerPoolDefinition`.
     - [ ] E3. Implement pool scaling logic (spawn new instances based on `MinIdleInstances`, `MaxTotalInstances`, and demand, respecting `AIWorkerDefinition` rate limits).
     - [ ] E4. Implement instance retirement logic within pools based on `InstanceRetirementPolicy`.
     - [ ] E5. Method for `AIWorkerPool` to provide an idle, compatible `AIWorkerInstance` to the Task Dispatcher.

   - | | **F. WorkQueue & WorkItem Implementation (Phase 1 - Stateless Focus, In-Memory) (`pkg/core/ai_wm_queues.go` - new file, `ai_wm.go`)**
     - [ ] F1. Implement `WorkQueueDefinition` CRUD operations in `AIWorkerManager`.
     - [ ] F2. Develop runtime `WorkQueue` struct/logic for managing `WorkItem`s (in-memory for Phase 1).
         - [ ] F2a. Logic for constructing a `WorkItem` using an optional `WorkItemDefinitionName` as a template, and merging/overriding with explicitly provided `WorkItem` fields.
     - [ ] F3. Implement logic for adding `WorkItem`s to a queue and retrieving them (e.g., FIFO initially, consider priority later).
     - [ ] F4. Logic to update `WorkItem` status (`Pending`, `Processing`, `Completed`, `Failed`).
     - [ ] F5. (Future Phase) Implement `WorkItem` persistence if `WorkQueueDefinition.PersistTasks` is true.

   - | | **G. AIWorkerManager Core Logic Updates (`pkg/core/ai_wm.go`)**
     - [ ] G1. Design and implement the **Task Dispatcher** component/logic within `AIWorkerManager`.
       - [ ] G1a. Monitor queues for pending `WorkItem`s.
       - [ ] G1b. Identify compatible `AIWorkerPool`s for a `WorkItem` (using `WorkItem.TargetWorkerCriteria` or `WorkQueueDefinition.AssociatedPoolNames`).
       - [ ] G1c. Request idle instances from pools.
       - [ ] G1d. Assign `WorkItem`s to instances, ensuring full context (DataSources, SAI refs, Tool Permissions) is assembled and passed or made available.
     - [ ] G2. Refactor or adapt `AIWorker.ExecuteStatelessTask` tool and its underlying manager method to potentially use the new queueing system (e.g., as a sync wrapper around async submission).
     - [ ] G3. Ensure `AIWorkerManager` correctly initializes and manages runtime states of pools, queues, and item definitions.

   - | | **H. Context Resolution Logic (`pkg/core/ai_wm.go`, related execution paths)**
     - [ ] H1. Implement logic to resolve **Effective DataSources** for a task (additive merge: `WorkItem` -> `WorkItemDefinition` -> `AIWorkerInstance` -> `AIWorkerPoolDefinition` -> `WorkQueueDefinition` -> `AIWorkerDefinition`).
     - [ ] H2. Implement logic for `AIWorkerInstance` to use its `ResolvedDataSources` during task execution.
     - [ ] H3. Ensure tool permissions (`ToolAllowlist`/`ToolDenylist` from `AIWorkerDefinition`) are correctly applied by the `SecurityLayer` for the current task context.

   - | | **I. NeuroScript Tooling (`pkg/core/ai_wm_tools_*.go` - new/modified files)**
     - [ ] I1. **GlobalDataSource Tools**:
       - [ ] `AIWorkerDataSource.AddGlobal`
       - [ ] `AIWorkerDataSource.GetGlobal`
       - [ ] `AIWorkerDataSource.ListGlobal`
       - [ ] `AIWorkerDataSource.UpdateGlobal`
       - [ ] `AIWorkerDataSource.RemoveGlobal`
     - [ ] I2. **WorkItemDefinition Tools (NEW)**:
       - [ ] `AIWorkerWorkItemDefinition.Add`
       - [ ] `AIWorkerWorkItemDefinition.Get`
       - [ ] `AIWorkerWorkItemDefinition.List`
       - [ ] `AIWorkerWorkItemDefinition.Update`
       - [ ] `AIWorkerWorkItemDefinition.Remove`
     - [ ] I3. **WorkerDefinition Tools**:
       - [ ] Update `AIWorkerDefinition.Update` tool for new reference fields (`DataSourceRefs`, `ToolAllowlist`, `ToolDenylist`, `DefaultSupervisoryAIRef`).
     - [ ] I4. **WorkerPool Tools**:
       - [ ] `AIWorkerPool.Create`
       - [ ] `AIWorkerPool.Get`
       - [ ] `AIWorkerPool.Update`
       - [ ] `AIWorkerPool.Delete`
       - [ ] `AIWorkerPool.List`
       - [ ] `AIWorkerPool.GetInstanceStatus`
     - [ ] I5. **WorkQueue Tools**:
       - [ ] `WorkQueue.Create`
       - [ ] `WorkQueue.Get`
       - [ ] `WorkQueue.Update`
       - [ ] `WorkQueue.Delete`
       - [ ] `WorkQueue.List`
       - [ ] `WorkQueue.GetStatus`
     - [ ] I6. **Task Management Tools**:
       - [ ] Update `WorkQueue.SubmitTask` tool to accept optional `workItemDefinitionName` and allow `payloadOverrideJSON`, `targetWorkerCriteriaOverrideJSON`, etc. to merge with the definition.
       - [ ] `WorkQueue.GetTaskInfo (taskID)`
     - [ ] I7. **File Sync Tools**:
       - [ ] Implement or update `AIWorker.SyncDataSource (dataSourceName)` tool to use `GlobalDataSourceDefinition`.
     - [ ] I8. Register all new tools with the interpreter.

   - | | **J. Security Enhancements (`pkg/core/security.go`, `pkg/core/tools_file_api.go`)**
     - [ ] J1. Enhance `SecurityLayer`'s `ValidateToolCall` and `GetToolDeclarations` to consider the active worker's `AIWorkerDefinition`-specific `ToolAllowlist` and `ToolDenylist`.
     - [ ] J2. Update path resolution logic in `FileAPI` and `SecurityLayer` (e.g., `ResolveAndSecurePath`, `IsPathInSandbox`) to handle `GlobalDataSourceDefinition.LocalPath` correctly, especially for `AllowExternalReadAccess`, using the effective data sources context.
     - [ ] J3. Rigorously ensure all file write tools (`toolWriteFile`, etc.) remain confined to the primary interpreter sandbox, irrespective of `GlobalDataSourceDefinition`s.

   - | | **K. Persistence for New Definitions (`pkg/core/ai_wm.go`)**
     - [ ] K1. Implement persistence for `GlobalDataSourceDefinition`s to/from `ai_global_data_source_definitions.json`.
     - [ ] K2. Implement persistence for `WorkItemDefinition`s to/from `ai_work_item_definitions.json`.
     - [ ] K3. Implement persistence for `AIWorkerPoolDefinition`s to/from `ai_worker_pool_definitions.json`.
     - [ ] K4. Implement persistence for `WorkQueueDefinition`s to/from `ai_work_queue_definitions.json`.

   - | | **L. Supervisory AI (SAI) Integration - Foundational Hooks (Future Focus)**
     - [ ] L1. Define and implement event emission points within `AIWorkerManager`, pools, queues, and instances for key lifecycle events (task submission, start, completion, failure; instance spawn, retire).
     - [ ] L2. Implement logic to resolve **Effective Supervisory AI** based on the hierarchical model (WorkItem -> WorkItemDefinition -> Instance -> Pool -> Queue -> Definition).
     - [ ] L3. Basic mechanism to route emitted events to the resolved SAI (e.g., if SAI is an active worker, could be via an internal "SAI.PostEvent" tool, or direct method call if SAI is a Go component).
     - [ ] L4. Ensure `SupervisorFeedback` in `PerformanceRecord` can be populated by an SAI.

   - | | **M. Testing**
     - [ ] M1. Unit tests for `GlobalDataSourceDefinition` CRUD and validation (including external path security).
     - [ ] M2. Unit tests for `WorkItemDefinition` CRUD and usage in `WorkItem` creation.
     - [ ] M3. Unit tests for `AIWorkerPoolDefinition` and `WorkQueueDefinition` CRUD.
     - [ ] M4. Unit tests for `AIWorkerPool` runtime logic (scaling, instance management, providing idle workers).
     - [ ] M5. Unit tests for `WorkQueue` runtime logic (item add/retrieve, status updates - in-memory initially), including item creation from `WorkItemDefinition`.
     - [ ] M6. Unit tests for Task Dispatcher logic (basic task-to-worker assignment, context assembly including from `WorkItemDefinition`).
     - [ ] M7. Unit tests for Context Resolution (DataSources, Tool Permissions, basic SAI ref, including `WorkItemDefinition` contributions).
     - [ ] M8. Integration tests for task submission (using `WorkItemDefinition`) & execution by a pooled worker (stateless).
     - [ ] M9. Tests for all new/updated NeuroScript tools related to AI WM.
     - [ ] M10. Security tests focusing on `AllowExternalReadAccess` validation and correct application of `ToolAllowlist`/`ToolDenylist`.

 ## gopls Integration for Advanced Go Development (v0.5 Focus) @gopls @core @codegen
   :: description: Enable NeuroScript to leverage gopls for precise Go code diagnostics and semantic understanding, empowering AI to write compiling code reliably.
   :: objective: Allow an AI to write compiling Go code **without fail** as quickly and efficiently as possible using gopls feedback.
   :: dependsOn: docs/gopls_integration_design.md, pkg/core/tools/gopls_client.go (new), pkg/core/tools_gopls.go (new)

   - | | **A. LSP Client Infrastructure (`pkg/core/tools/gopls_client.go` - new file/package)**
     :: description: Establish foundational communication with a gopls server instance.
     - [ ] A1. Design `GoplsClient` interface and basic implementation.
         - [ ] A1a. Method to start and manage a `gopls` subprocess.
         - [ ] A1b. Implement JSON-RPC 2.0 stream handling for LSP messages (consider using `golang.org/x/tools/internal/jsonrpc2` or a similar library if suitable and licensed permissively, otherwise a minimal implementation).
         - [ ] A1c. Basic LSP request/response/notification handling logic.
     - [ ] A2. Implement LSP `initialize` / `initialized` handshake sequence.
         - [ ] A2a. Send `initialize` request with client capabilities (minimal set needed for diagnostics, hover, definition).
         - [ ] A2b. Process `initialize` response from gopls and store server capabilities.
         - [ ] A2c. Send `initialized` notification.
     - [ ] A3. Implement LSP `shutdown` / `exit` sequence for graceful termination.
     - [ ] A4. Basic error handling and logging for LSP communication issues.

   - | | **B. Workspace and Document Synchronization**
     :: description: Ensure gopls has an accurate view of the workspace and file contents.
     - [ ] B1. Implement `GoplsClient` methods for LSP `textDocument/didOpen` notification.
         - [ ] B1a. Tool `Gopls.NotifyDidOpen (filePath string, content string)` for NeuroScript to inform gopls.
     - [ ] B2. Implement `GoplsClient` methods for LSP `textDocument/didChange` notification.
         - [ ] B2a. Tool `Gopls.NotifyDidChange (filePath string, newContent string)` (can initially send full content).
         - [ ] B2b. (Future) Support for incremental changes if performance becomes an issue.
     - [ ] B3. Implement `GoplsClient` methods for LSP `textDocument/didSave` notification (if distinct from didChange for gopls).
         - [ ] B3a. Tool `Gopls.NotifyDidSave (filePath string)`
     - [ ] B4. Implement `GoplsClient` methods for LSP `textDocument/didClose` notification.
         - [ ] B4a. Tool `Gopls.NotifyDidClose (filePath string)`
     - [ ] B5. Tool `Gopls.SetWorkspaceRoot (workspacePath string)` to inform gopls of the project root (likely via `initializeParams.rootUri`).

   - | | **C. Diagnostic Tools (`pkg/core/tools_gopls_diags.go` - new file, part of new `tools_gopls.go` toolset)**
     :: description: Provide NeuroScript tools to fetch and interpret diagnostics from gopls.
     - [ ] C1. Define `DiagnosticInfo` struct in Go (mirroring LSP `Diagnostic` fields: SourceFileURI, Range (precise byte offsets), Severity, Code, Message, Source, SemanticReferenceGuess).
     - [ ] C2. Implement `GoplsClient` logic to receive and parse `textDocument/publishDiagnostics` notifications from gopls.
         - [ ] C2a. Store diagnostics per file URI.
     - [ ] C3. NeuroScript Tool: `Gopls.GetDiagnostics (filePath string) (diagnostics []DiagnosticInfo)`
         - [ ] C3a. Retrieves currently stored diagnostics for the given file.
         - [ ] C3b. Ensures file is opened/synced with gopls if not already.
     - [ ] C4. NeuroScript Tool: `Gopls.GetAllProjectDiagnostics () (map[string][]DiagnosticInfo)`
         - [ ] C4a. Returns all diagnostics for all files currently known and managed by the gopls client.

   - | | **D. Semantic Information & Contextual Tools (AI-Focused)** @gopls @semantic
     :: description: Provide tools for the AI to get rich semantic information about Go code elements, particularly those related to diagnostics, using stable and reliable addressing rather than just line/column. This is crucial for the AI to understand errors and generate correct code.
     - [ ] D1. Define NeuroScript internal representation for "SemanticReference" to Go code elements. This should include:
         - [ ] D1a. `ResourceURI` (string).
         - [ ] D1b. `FullyQualifiedName` (string, optional).
         - [ ] D1c. `SymbolKind` (string, optional, from LSP `SymbolKind` or go/ast).
         - [ ] D1d. `DeclarationLocation` (`LSPLocation`, optional, using precise byte offsets).
         - [ ] D1e. `ByteOffsetRange` (`LSPRange`, optional, precise byte offsets of the symbol's identifier/span).
         - [ ] D1f. `Signature` (string, optional, for functions/methods).
         - [ ] D1g. `PackagePath` (string, optional).
     - [ ] D2. Enhance `DiagnosticInfo` (from C1) to include a primary `SemanticReferenceGuess` for the code element most directly associated with the diagnostic.
     - [ ] D3. NeuroScript Tool: `Gopls.GetSymbolInfoAt(filePath string, byteOffset int) (symbolInfo map[string]interface{}, error string)`
         :: description: Combines hover and definition for a precise byte offset (often part of a diagnostic's range) to give the AI immediate context. Returns a map containing `semanticReference`, `hoverContent`, `definitionLocation` (as maps).
         - [ ] D3a. Internally use LSP `textDocument/hover` and `textDocument/definition` at the given precise location.
         - [ ] D3b. Convert the responses into the standardized `SemanticReference` format for locations and `hoverContent` for descriptions.
         - [ ] D3c. Ensure byte offsets are handled consistently (UTF-8, as used by gopls).
     - [ ] D4. NeuroScript Tool: `Gopls.GetSymbolInfoByName(fullyQualifiedName string) (symbolInfo map[string]interface{}, error string)`
         :: description: Allows AI to query information using a known symbol name.
         - [ ] D4a. Investigate using gopls `workspace/symbol` or a sequence of LSP calls to resolve FQN to a location, then enrich with hover/definition.
         - [ ] D4b. Alternatively, leverage/enhance `pkg/core/tools/gosemantic` (e.g., `toolGoGetDeclarationOfSymbol`) to get an initial location, then use gopls to enrich.
     - [ ] D5. NeuroScript Tool: `Gopls.ListSymbolsInFile(filePath string) (symbols []map[string]interface{}, error string)`
         :: description: Wraps `textDocument/documentSymbol` to list symbols as `SemanticReference` maps.
     - [ ] D6. NeuroScript Tool: `Gopls.FindWorkspaceSymbols(queryString string) (symbols []map[string]interface{}, error string)`
         :: description: Wraps `workspace/symbol` to find symbols project-wide, returning `SemanticReference` maps.
     - [ ] D7. Ensure all tools returning location information primarily use the defined `SemanticReference` format. Raw offset ranges can be part of the `SemanticReference`.

   - | | **E. NeuroScript Toolset Registration (`pkg/core/tools_gopls.go` - new file)**
     - [ ] E1. Create `RegisterGoplsTools(*core.Interpreter)` function.
     - [ ] E2. Instantiate `GoplsClient` (likely as a singleton or per-interpreter service) and manage its lifecycle (start, shutdown).
         - [ ] E2a. `GoplsClient` should be accessible to the gopls tools.
     - [ ] E3. Register all new `Gopls.*` tools with the interpreter.

   - | | **F. Testing**
     - [ ] F1. Unit tests for `GoplsClient` LSP message parsing and basic request/response flow (mocking gopls server or using test doubles for JSON-RPC layer).
     - [ ] F2. Integration tests (requires starting a real `gopls` server):
         - [ ] F2a. Test `initialize` handshake and workspace setup.
         - [ ] F2b. Test `didOpen`, `didChange`, `didSave` notifications trigger `publishDiagnostics`.
         - [ ] F2c. Test `Gopls.GetDiagnostics` for various Go files with known errors and warnings, verify `DiagnosticInfo` structure and `SemanticReferenceGuess`.
         - [ ] F2d. Test `Gopls.GetSymbolInfoAt` and `Gopls.GetSymbolInfoByName` on known symbols, verify `SemanticReference` and other returned info.
         - [ ] F2e. Test `Gopls.ListSymbolsInFile` and `Gopls.FindWorkspaceSymbols`.
     - [ ] F3. Test error handling for gopls communication failures, invalid responses, or if gopls server is not found/started.

 ## Other bits

 - [ ] Matrix tools @lang_features @unprioritized
 - [ ] Named arguments @lang_features @unprioritized
 - [ ] Default arguments @lang_features @unprioritized


 :: title: NeuroScript Road to v0.5.0 Checklist
 :: version: 0.5.0
 :: id: ns-roadmap-v0.5.0
 :: status: draft
 :: description: Prioritized tasks to implement NeuroScript v0.5.0 based on design discussions and Vision (May 10, 2025).
 :: updated: 2025-05-10