:: title: NeuroScript Road to v0.3.0 Checklist
:: version: 0.3.9
:: id: ns-roadmap-v0.3.0
:: status: draft
:: description: Prioritized tasks to implement NeuroScript v0.3.0 based on design discussions and Vision (May 5, 2025).

# NeuroScript v0.3.0 Development Tasks

## Vision (Summary)

Achieve an AI-driven development loop where a Supervising AI (SAI) interacts with NeuroScript (ng) and a File API to:
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
    - [ ] Review error variable definitions (errors.go) for consistency/completeness.
    - [ ] Investigate potential for parallel step execution? (Future/low priority)
    - [ ] Configuration: Mechanism for SAI to configure ng sandbox/tools?

- [-] 3. Core Go Tooling (Go-Specific)
    - [x] Go Module Awareness: GoGetModuleInfo, FindAndParseGoMod helper. (DONE)
    - [x] Formatting/Imports: GoFmt, GoImports. (DONE)
    - [x] Execution: GoTest, GoBuild. (DONE)
    - [x] Basic Listing: GoListPackages, GoCheck. (DONE)
    - [x] Diagnostics: Implement GoVet, GoLint (or similar static analysis) tools.
    - [x] Code Indexing & Search:
        - [x] Implement GoIndexCode tool (using go/packages?) to build a semantic index.
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

- [ ] 4. Core Generic Tree Tooling
    - [x] Define GenericTree/GenericTreeNode. (DONE)
    - [x] Implement TreeLoadJSON. (DONE)
    - [x] Implement TreeGetNode. (DONE)
    - [x] Implement TreeGetChildren. (DONE)
    - [x] Implement TreeGetParent. (DONE)
    - [x] Implement TreeFormatJSON. (DONE)
    - [x] Implement TreeRenderText. (DONE)
    - [x] Implement TreeFindNodes (querying based on type/value/attributes).
    - [x] Implement Tree Modification Tools:
        - [x] TreeModifyNode (basic Value modify)
        - [x] TreeSetAttribute
        - [x] TreeRemoveAttribute
        - [x] TreeAddNode
        - [x] TreeRemoveNode
        - [ ] Array modification tools? (TreeAppendChild, TreeRemoveChild, etc.) (Optional/Lower priority for now)

- [x] 5. Filesystem / OS / Version Control Tools
    - [x] Basic FS: Read/Write/Stat/Delete/ListDir/WalkDir/MkdirAll (FS.*). (Existing)
    - [x] Hashing: FS.Hash. (Existing)
    - [x] Command Execution: Shell.ExecuteCommand. (Existing)
    - [x] Basic Git: Git.Status. (Existing)
    - [ ] Enhanced Git: Add Git.Branch, Git.Checkout, Git.Commit, Git.Push, Git.Diff.
    - [x] Diff/Patching:
        - [x] Implement/refine NSPatch.GeneratePatch tool (line-based diff).
        - [x] Implement NSPatch.ApplyPatch tool (validates/applies patches).
        - [x] Define .ndpatch format (JSON list of operations used by tools).
        - [x] Add tests for various patch scenarios (add, delete, replace, empty files).
        - [ ] Consider FS.DiffFiles tool (or is NSPatch sufficient?).
    - [ ] FileAPI Review: Ensure consistency/no overlap between FS.* tools and direct FileAPI.
    - [ ] ng -> FileAPI Sync: Design and implement mechanism/tools.
    - [ ] Build Artifacts: Review GoBuild output handling; add tools if needed (e.g., FS.Copy, retrieve artifacts).

- [ ] 6. Tooling & Ecosystem
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

- [ ] 7. Example App -- language flashcards (New)
    - [ ] Define core features (add card, review, save/load).
    - [ ] Design data structure (simple list/map, maybe JSON file).
    - [ ] Implement basic TUI or script interaction logic.

- [x] 8. Language / Interpreter Polish (Internal / Done)
    - [x] core.ToIntE undefined error fixed by adding core.ConvertToInt64E. (DONE)
    - [x] Handle non-deterministic map iteration in tests. (DONE)