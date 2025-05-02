:: title: NeuroScript Road to v0.3.0 Checklist
 :: version: 0.3.0-alpha2
 :: id: ns-roadmap-v0.3.0
 :: status: draft
 :: description: Prioritized tasks to implement NeuroScript v0.3.0 based on design discussions and Vision (May 1, 2025).

 # NeuroScript v0.3.0 Development Tasks

 ## Vision (Summary)

 Achieve an AI-driven development loop where a Supervising AI (SAI) interacts with NeuroScript (`ng`) and a File API to:
 - Manage code in branches (`Git`).
 - Read, write, and understand code structure (`FS.*`, `Go*`, Tree tools, Indexing).
 - Run diagnostics and tests (`GoCheck`, `GoVet`?, `GoTest`).
 - Generate and apply changes (Diff/Patch tools).
 - Coordinate tasks, potentially using worker AIs.
 - Track progress via shared checklists (`Checklist` tools).

 ## Checklist

 - [x] 1. Language Features
     - [x] `break` and `continue` flow control words

 - [ ] 2. Core Interpreter / Runtime
     - [ ] Error Handling: Review `on_error` behavior, especially interaction with `return`.
     - [ ] Handle Management: Add explicit `ReleaseHandle` tool? Review potential memory leaks from unreleased handles.
     - [ ] Performance: Basic profiling pass (if major slowdowns observed).
     - [ ] Review error variable definitions (`errors.go`) for consistency/completeness.
     - [ ] Investigate potential for parallel step execution? (Future/low priority)
     - [ ] Configuration: Mechanism for SAI to configure `ng` sandbox/tools?

 - [ ] 3. Core Go Tooling (Go-Specific)
     - [x] Go Module Awareness: `GoGetModuleInfo`, `FindAndParseGoMod` helper. (DONE)
     - [x] Formatting/Imports: `GoFmt`, `GoImports`. (DONE)
     - [x] Execution: `GoTest`, `GoBuild`. (DONE)
     - [x] Basic Listing: `GoListPackages`, `GoCheck`. (DONE)
     - [ ] Diagnostics: Implement `GoVet`, `GoLint` (or similar static analysis) tools. **(NEW - Vision #11)**
     - [ ] Code Indexing & Search:
         - [ ] Implement `GoIndexCode` tool (using `go/packages`?) to build a semantic index. **(NEW - Vision #5, #9)**
         - [ ] Enhance/replace `GoFindIdentifiers` to use index for faster/better search.
         - [ ] Implement `GoFindDeclarations` tool (using index/AST).
         - [ ] Implement `GoFindUsages` tool (using index/AST).
         - [ ] Handle package aliases properly in all find tools.
     - [ ] AST Modification:
         - [x] Basic Structure: Change package, Add/Remove/Replace imports, Replace `pkg.Symbol`. (Existing `GoModifyAST`)
         - [ ] Renaming: `RenameLocalVariable`, `RenameParameter`, `RenameFunction` (cross-file?).
         - [ ] Refactoring: Extract Function, Extract Variable.
     - [ ] AST Navigation:
         - [x] Get node info at position. (DONE: `GoGetNodeInfo`)
         - [ ] Get parent node info. (`GoGetNodeParent` - NEXT?)
         - [ ] Get child node info list. (`GoGetNodeChildren` - NEXT?)
         - [ ] Get node siblings. (`GoGetNodeSiblings`?)

 - [ ] 4. Core Generic Tree Tooling
     - [x] Define `GenericTree`/`GenericTreeNode`. (DONE)
     - [x] Implement `TreeLoadJSON`. (DONE)
     - [x] Implement `TreeGetNode`. (DONE)
     - [x] Implement `TreeGetChildren`. (DONE)
     - [x] Implement `TreeGetParent`. (DONE)
     - [x] Implement `TreeFormatJSON`. (DONE)
     - [x] Implement `TreeRenderText`. (DONE)
     - [ ] Implement `TreeFromGoAST` adapter tool.
     - [ ] Implement `TreeFindNodes` (querying based on type/value/attributes).
     - [ ] Implement Tree Modification Tools (`TreeModifyNode`, `TreeAddNode`, `TreeRemoveNode`).
     - [ ] Implement `TreeLoadXML`?

 - [ ] 5. Filesystem / OS / Version Control Tools
     - [x] Basic FS: Read/Write/Stat/Delete/ListDir/WalkDir/MkdirAll (`FS.*`). (Existing)
     - [x] Hashing: `FS.Hash`. (Existing)
     - [x] Command Execution: `Shell.ExecuteCommand`. (Existing)
     - [x] Basic Git: `Git.Status`. (Existing)
     - [ ] Enhanced Git: Add `Git.Branch`, `Git.Checkout`, `Git.Commit`, `Git.Push`, `Git.Diff`. **(NEW - Vision #3)**
     - [ ] Diff/Patching: Add `FS.DiffFiles`, potentially `NSPatch.Apply` tool. **(NEW - Vision Implied)**
     - [ ] FileAPI Review: Ensure consistency/no overlap between `FS.*` tools and direct `FileAPI`.
     - [ ] `ng` -> FileAPI Sync: Design and implement mechanism/tools. **(NEW - Vision #10)**
     - [ ] Build Artifacts: Review `GoBuild` output handling; add tools if needed (e.g., `FS.Copy`, retrieve artifacts). **(NEW - Vision Implied)**

 - [ ] 6. Tooling & Ecosystem
     - [ ] Documentation: Update tool docs, checklists to reflect new tools/status.
     - [ ] Formatting: Begin development of `nsfmt` formatting tool for `.ns` files.
     - [ ] Workflow Test: Create end-to-end test script simulating SAI interaction (upload, read, modify, build, test). **(Refined - Vision #6)**
     - [ ] Checklist Tooling: Implement `Checklist.Read`, `Checklist.UpdateItem` tools using `pkg/neurodata/checklist`. **(NEW - Vision #13)**

 - [ ] 7. Example App -- language flashcards *(New)*
     - [ ] Define core features (add card, review, save/load).
     - [ ] Design data structure (simple list/map, maybe JSON file).
     - [ ] Implement basic TUI or script interaction logic.

 - [-] 8. Language / Interpreter Polish *(Internal / Done)*
     - [-] `core.ToIntE` undefined error fixed by adding `core.ConvertToIntE`. (DONE)
     - [-] Handle non-deterministic map iteration in tests. (DONE)