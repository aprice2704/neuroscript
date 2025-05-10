:: title: NeuroScript Road to v0.3.0 Checklist
:: version: 0.3.11
:: id: ns-roadmap-v0.3.0
:: status: draft
:: description: Prioritized tasks to implement NeuroScript v0.3.0 based on design discussions and Vision (May 7, 2025).
:: updated: 2025-05-07

# NeuroScript v0.3.0 Development Tasks

## Vision (Summary)

Achieve an AI-driven development loop where a Supervising AI (SAI) interacts with NeuroScript (ng) and a File API to:
- Apply a prompt to each file in a file tree 
- Manage code in branches (Git).
- Read, write, and understand code structure (FS.*, Go*, Tree tools, Indexing).
- Run diagnostics and tests (GoCheck, GoVet?, GoTest).
- Generate and apply changes (Diff/Patch tools).
- Coordinate tasks, potentially using worker AIs.
- Track progress via shared checklists (Checklist tools).

## Checklist

- [x] 1. Language Features
    - [x] break and continue flow control words

- [ ] 2. Core Interpreter / Runtime
    - [ ] Error Handling: Review on_error behavior, especially interaction with return.
    - [ ] Handle Management: Add explicit ReleaseHandle tool? Review potential memory leaks from unreleased handles.
    - [ ] Performance: Basic profiling pass (if major slowdowns observed).
    - [x] Review error variable definitions (errors.go) for consistency/completeness. (Implicitly improved through recent work)
    - [ ] Investigate potential for parallel step execution? (Future/low priority)
    - [ ] Configuration: Mechanism for SAI to configure ng sandbox/tools?
    - [ ] Apply prompt to each file in a tree (file filter, sys prompt etc.)
    - [ ] Update ng "modes" etc.

- [x] 3. Core Go Tooling (Go-Specific)
    - [x] Go Module Awareness: GoGetModuleInfo, FindAndParseGoMod helper.
    - [x] Formatting/Imports: GoFmt, GoImports.
    - [x] Execution: GoTest, GoBuild.
    - [x] Basic Listing: GoListPackages, GoCheck.
    - [x] Diagnostics: Implement GoVet, Staticcheck tools.
    - [x] Code Indexing & Search:
        - [x] Implement GoIndexCode tool (using go/packages) to build a semantic index.
        - [x] Enhance/replace GoFindIdentifiers to use index for faster/better search. (Superseded by semantic tools)
        - [x] Implement GoFindDeclarations tool (using index/AST).
        - [x] Implement GoFindUsages tool (using index/AST).
        - [x] Handle package aliases properly in all find tools. (Tests exist)
    - [x] AST Modification:
        - [x] Basic Structure: Change package, Add/Remove/Replace imports, Replace pkg.Symbol. (Existing GoModifyAST)
        - [x] Renaming: RenameLocalVariable, RenameParameter, RenameFunction (cross-file?). (Covered by GoRenameSymbol)
        - [ ] Refactoring: Extract Function, Extract Variable.
    - [x] AST Navigation:
        - [x] Get node info at position. (DONE: GoGetNodeInfo)
        - [ ] Get parent node info. (GoGetNodeParent - NEXT?)
        - [ ] Get child node info list. (GoGetNodeChildren - NEXT?)
        - [ ] Get node siblings. (GoGetNodeSiblings?)

- [x] 4. Core Generic Tree Tooling
    - [x] Define GenericTree/GenericTreeNode.
    - [x] Implement TreeLoadJSON.
    - [x] Implement TreeGetNode.
    - [x] Implement TreeGetChildren.
    - [x] Implement TreeGetParent.
    - [x] Implement TreeFormatJSON.
    - [x] Implement TreeRenderText.
    - [x] Implement TreeFindNodes (querying based on type/value/attributes).
    - [x] Implement Tree Modification Tools:
        - [x] TreeModifyNode (basic Value modify)
        - [x] TreeSetAttribute
        - [x] TreeRemoveAttribute
        - [x] TreeAddNode
        - [x] TreeRemoveNode
        - [ ] Array modification tools? (TreeAppendChild, TreeRemoveChild, etc.) (Optional/Lower priority for now)

- [x] 5. Filesystem / OS / Version Control Tools
    - [x] Basic FS: Read/Write/Stat/Delete/ListDir/WalkDir/MkdirAll (FS.*).
    - [x] Hashing: FS.Hash.
    - [x] Command Execution: Shell.ExecuteCommand.
    - [x] Basic Git: Git.Status.
    - [x] Enhanced Git: Add Git.Branch, Git.Checkout, Git.Commit, Git.Push, Git.Diff. (Marking as done since tools are registered)
    - [x] Diff/Patching:
        - [x] Implement/refine NSPatch.GeneratePatch tool (line-based diff).
        - [x] Implement NSPatch.ApplyPatch tool (validates/applies patches).
        - [x] Define .ndpatch format (JSON list of operations used by tools).
        - [x] Add tests for various patch scenarios (add, delete, replace, empty files).
        - [ ] Consider FS.DiffFiles tool (or is NSPatch sufficient?).
    - [ ] FileAPI Review: Ensure consistency/no overlap between FS.* tools and direct FileAPI.
    - [ ] ng -> FileAPI Sync: Design and implement mechanism/tools.
    - [ ] Build Artifacts: Review GoBuild output handling; add tools if needed (e.g., FS.Copy, retrieve artifacts).

- [x] 6. Tooling & Ecosystem
    - [ ] Documentation: Update tool docs, checklists to reflect new tools/status.
    - [ ] Formatting: Begin development of nsfmt formatting tool for .ns files.
    - [ ] Workflow Test: Create end-to-end test script simulating SAI interaction (upload, read, modify, build, test).
    - [x] Checklist Tooling (using Tree Tools):
        - [x] Implement ParseChecklistFromString tool (Provides initial data).
        - [x] Define Checklist <-> GenericTree representation mapping.
        - [x] Implement Checklist <-> GenericTree adapter logic/tool(s).
        - [x] Define/Implement Checklist formatting/serialization logic (from Tree representation to Markdown).
        - [x] Define/Implement in-memory Checklist update logic using Tree tools (FindNodes, ModifyNode, etc.).
        - [x] Ensure update logic correctly recomputes status of automatic ('| |') items based on tree structure/dependencies.
        - [x] Define/Implement Checklist tool(s) for updates via Tree (depends on Tree find/modify tools & adapter).
    - |x| AI Worker Management System
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
        - [ ] Design for stateful worker interaction and task lifecycle.
        - [ ] Tooling for SAI to assign/monitor tasks on workers.
        - [ ] Agent mode permitted, allow & deny lists

- [ ] 7. Example App -- language flashcards (New)
    - [ ] Define core features (add card, review, save/load).
    - [ ] Design data structure (simple list/map, maybe JSON file).
    - [ ] Implement basic TUI or script interaction logic.

- [x] 8. Language / Interpreter Polish (Internal / Done)
    - [x] core.ToIntE undefined error fixed by adding core.ConvertToInt64E.
    - [x] Handle non-deterministic map iteration in tests.
