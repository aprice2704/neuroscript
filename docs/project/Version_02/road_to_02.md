:: title: NeuroScript Consolidated Road to v0.2.0 Checklist
:: version: 0.2.0-beta
:: id: ns-roadmap-v0.2.0-consolidated-beta
:: status: draft
:: description: Combined and prioritized tasks for NeuroScript v0.2.0 and beyond, merging items from road_to_02.md, development_checklist.ndcl.md, near-list.ndcl.md, and near-term.md as of Apr 26, 2025. Reformatted to .ndcl syntax with partial rollup.
:: dependsOn: project/Version_02/road_to_02.md, project/Version_02/development_checklist.ndcl.md, project/Version_02/near-list.ndcl.md, project/Version_02/near-term.md, checklist.md

# NeuroScript Consolidated Development Tasks (Targeting v0.2.0+)

## Overall High-Level Goals

-- Reach "Bootstrapping" Point: NeuroScript project can maintain itself (update docs/scripts, run tests, update/fix source code).
-- Full `neurogo` Capabilities: Vector DB for scripts, full/secure LLM comms, conversational mode.

## 1. Core v0.2.0 Syntax Changes (Parser/Lexer)
- |-| Core v0.2.0 Syntax Implementation // Contains non-[x] items
  - [ ] Convert all keywords to lowercase.
  - [ ] Implement `func <name> [needs...] [optional...] [returns...] means ... endfunc` structure.
  - [ ] Implement specific block terminators (`end if`, `end for`, `end while`).
  - [ ] Implement triple-backtick string parsing.
  - [ ] Implement default `{{placeholder}}` evaluation within triple-backtick strings.
  - [ ] Implement `:: keyword: value` metadata parsing (header & inline).
  - [ ] Remove old `comment:` block parsing.
  - [ ] Implement `no`/`some` keyword parsing in expressions.
  - [ ] Implement `must`/`mustBe` statement/expression parsing.
  - [ ] Implement basic `try`/`catch`/`finally` structure parsing.

## 2. Core v0.2.0 Semantics (Interpreter)
- |-| Core v0.2.0 Semantics Implementation // Contains non-[x] items
  - [ ] Implement `askAI`/`askHuman`/`askComputer` execution logic.
  - [ ] Implement initial `handle` mechanism for `ask...` functions (e.g., string identifiers).
  - [ ] Implement direct assignment from `ask...` and returning `call tool...` (e.g., `set result = askAI(...)`).
  - [ ] Implement multiple return value handling (`returns` clause).
  - [ ] Implement `no`/`some` keyword evaluation logic (runtime type zero-value checks).
  - [ ] Ensure standard comparison and arithmetic operators function correctly.
  - [P] Interpreter: Add NeuroScript-specific Error Handling (e.g., TRY/CATCH or specific error types?) // Basic Go error handling (sentinels, wrapping) is implemented, but no NS-level TRY/CATCH.
  - [ ] Design simple `try/catch` mechanism and semantics (implementation might extend beyond 0.2.0).

## 3. Foundational Robustness Features (Phase 1 / Near-Term)
- |-| Robustness Features Implementation // Contains non-[x] items
  - [ ] Define and implement `must`/`mustBe` failure semantics (halt vs error?).
  - [ ] Create essential built-in check functions for `mustBe`.

## 4. Metadata Handling
- |-| Metadata Handling Implementation // Contains non-[x] items
  - [ ] Move metadata in markdown to eof, not sof
  - [ ] Implement storage/access for `::` metadata.
  - [ ] Define initial vocabulary for standard metadata keys and inline annotations.

## 5. Tooling & Ecosystem
- |-| Tooling & Ecosystem Development // Contains non-[x] items
  - |-| Standardization & Formatting // Contains non-[x] items
    - [ ] Standardize internal tool naming to use `tool.` prefix consistently.
    - [ ] Update checklists/docs to reflect standardized tool names.
    - [ ] Begin development of `nsfmt` formatting tool.
    - [ ] Tools: TOOL.NeuroScriptCheckSyntax(content) - Formal syntax check tool using the parser // Parser exists, but not exposed as a tool.
  - |-| Filesystem Operations // Contains non-[x] items
    - [x] tool.ReadFile(path)
    - [x] tool.WriteFile(path, content)
    - [x] tool.ListDirectory(path, [recursive], [pattern]) // Recursive implemented
      - [ ] Add pattern filtering implementation
    - [x] tool.Mkdir(path)
    - [x] tool.DeleteFile(path)
    - [x] tool.MoveFile(source, destination) // Implemented
    - [x] tool.LineCountFile(path) // Implied
    - [x] tool.SanitizeFilename(name) // Implied
    - [x] tool.WalkDir(path) // Implied
    - [x] tool.FileHash(path) // Implied
  - |-| Go Code Analysis & Manipulation (AST Tools) // Contains non-[x] items
    - [x] tool.GoParseFile(path or content)
    - [x] tool.GoFindIdentifiers(ast_handle, pkg_name, identifier)
    - [x] tool.GoModifyAST(ast_handle, modifications) // Marked Exists
      - [x] - Change Package Declaration
      - [x] - Add/Remove/Replace Import Paths
      - [x] - Replace Qualified Identifiers
      - [ ] Sub-op: Change Package Declaration (Check status difference?)
      - [ ] Sub-op: Add/Remove/Replace Import Paths (Check status difference?)
      - [ ] Sub-op: Replace Qualified Identifiers (Check status difference?)
    - [x] tool.GoFormatASTNode(ast_handle)
    - [x] tool.GoUpdateImportsForMovedPackage(...) // Marked Specced, Impl [x] -> Let's assume complete
      - [x] Go Implementation (Marking parent as done based on this)
  - |-| Build & Verification Tools // Contains non-[x] items
    - [x] tool.GoBuild([target])
    - [x] tool.GoTest()
    - [x] tool.GoCheck([target])
    - [x] tool.GoModTidy()
    - [ ] Tools/Known Issue: TOOL.GoBuild and TOOL.GoCheck error reporting for single files needs improvement.
  - |-| Version Control Tools (Git) // Contains non-[x] items
    - [x] tool.GitAdd(path)
    - [x] tool.GitCommit(message)
    - [x] tool.GitNewBranch(branch_name)
    - [x] tool.GitCheckout(branch_name)
    - [x] tool.GitStatus()
    - [x] tool.GitPull()
    - [x] tool.GitPush()
    - [x] tool.GitDiff()
    - [x] tool.GitRm(path)
    - [P] Tools: Enhance Git Workflow (Add Branch support?, GitPull?, Auto-index after commit?) // Basic tools exist, needs enhancements like auto-indexing.
    - [ ] Use git branch for version control within tools
  - |-| List Manipulation // Contains non-[x] items
    - [x] tool.ListAppend(list, element) -> list
    - [x] tool.ListPrepend(list, element) -> list
    - [x] tool.ListGet(list, index, [default]) -> any
    - [x] tool.ListSlice(list, start, end) -> list
    - [x] tool.ListContains(list, element) -> bool
    - [x] tool.ListReverse(list) -> list
    - [x] tool.ListSort(list) -> list // Handles numeric/string, review scope.
    - [ ] tool.ListHead(list) -> any
    - [ ] tool.ListRest(list) -> list // (cdr equivalent)
    - [ ] tool.ListTail(list, count) -> list // (last N elements)
  - |x| User Interaction / Control // All children [x]
    - [x] IO.Input(prompt)
  - |-| General Utility Tools // Contains non-[x] items
    - [ ] Tools: Add more utility tools (JSON, HTTP, etc.)
    - [ ] Tools: Add Markdown tools (Read/Write)
    - [ ] Tools: Add Structured Document tools (hierarchical info/docs)
    - [ ] Tools: Add Table tools
    - [ ] Tools: Add Integration tools (e.g., Google Sheets and Docs)
    - [ ] Tools: Add data encoding/hardening tools (e.g., Base32, Base64, zip/unzip)
    - [ ] Tools: grep/egrep/agrep
    - [ ] fuzzy logic tools

## 6. NeuroData Types and Tools
- |-| NeuroData Implementation // Contains non-[x] items
  - [P] NeuroData files, template and instance (Design & Implement) // Core block/metadata/checklist impl exists.
    - [P] graph // Spec exists
    - [P] table // Spec exists
    - |-| tree // Spec exists // Contains non-[x] items
      - [ ] Design internal Go representation for `tree` handle type.
      - [ ] Implement `tool.listDirectoryTree` (or `tool.walkDir`) returning `tree` handle.
      - [ ] Implement basic tree manipulation tools (e.g., `tool.getNodeValue`, `tool.getNodeChildren`).
    - [P] decision_table // Spec exists
    - [P] form // Spec exists
    - [ ] invoice // Spec not found
    - [ ] statement_of_account // Spec not found
    - [ ] receipt // Spec not found
    - [ ] payment // Spec not found
    - [ ] packing_list // Spec not found
    - [ ] request_for_quote or estimate // Spec not found
    - [ ] quote or estimate // Spec not found
    - [ ] purchase_order // Spec not found
    - [ ] work_order // Spec not found
    - [ ] markdown_doc // Spec not found
    - [P] composite_doc // Spec exists, block extractor impl
    - [ ] bug_report // Spec not found
    - [ ] ns_tool_list // Spec not found
    - [P] enum // Spec exists
    - [ ] roles_list // Spec not found
    - [P] object // Design mentioned, uses map literal parsing
    - [ ] ishikawa // Spec not found
    - [ ] kanban // Spec not found
    - [ ] memo // Spec not found
    - [P] list // Spec exists
    - [x] checklist // Spec exists, Impl exists
  - [ ] block read to support block references
  - [ ] also read from string block references

## 7. Agent Architecture & Core Enhancements
- |-| Agent Architecture Enhancements // Contains non-[x] items
  - [ ] Agent Startup Script (`agent_startup.ns`): Replace flags with scriptable config.
    - [ ] Requires new `TOOL.Agent*` config tools (AgentSetSandbox, AgentPinFile, AgentSetModel, etc.).
  - [ ] `AgentContext` Object (`pkg/neurogo`): Central struct for agent state.
  - [ ] Typed Handles (`category::uuid`): Implement prefix system for runtime handle type safety.
  - |-| Dual Context Management Strategy // Contains non-[x] items
    - [ ] Pinning (`TOOL.AgentPinFile`) + Temp Request (`TOOL.RequestFileContext`) implementation.
    - [ ] AI Forgetting (`TOOL.Forget`/`tool.ForgetAll`) implementation.
  - [ ] Interpreter: Implement Context Management Strategy for CALL LLM -- defer // Basic call exists, needs advanced management.

## 8. File Synchronization Tools (Gemini File API)
- |-| File Synchronization Tools // Contains non-[x] items
  - [x] tool.SyncFiles(direction, localDir, [filterPattern]) // Marked complete in near-list
  - [x] tool.UploadFile(localPath, [displayName]) // Marked complete in near-list
  - [x] tool.ListAPIFiles() // Marked complete in near-list
  - [x] tool.DeleteAPIFile(apiFileName) // Marked complete in near-list
  - [ ] Tools: UploadToFilesAPI (Duplicate?)
  - [ ] Tools: SyncToFilesAPI (Duplicate?)

## 9. Longer Term / Advanced / Miscellaneous
- |-| Longer Term / Advanced / Miscellaneous // Contains non-[x] items
  - [P] Implement Restricted mode for running untrusted scripts (Design exists, implementation needed)
  - [ ] Feature: Add Self-test support in ns
  - [ ] Feature: Embed standard utility NeuroScripts (e.g., CommitChanges) into neurogo binary (using Go embed)
  - [ ] Block and file level prior version preservation
  - [ ] LLM able to read ns and execute it (via prompt guidance)
  - [ ] LLM able to translate simple ns into golang tool
  - [ ] LLMs can supply git-style patches and have them applied to files
  - [ ] MCP support
  - [ ] Ability to pass text from LLM to tool in (BASE64) or some other armored format
  - [ ] Prolog style features
  - [ ] SVG generation and manipulation
  - [ ] Nice example on website/readme
  - [ ] Logo
  - [ ] More self building and maint
  - [ ] keep prior versions meta tag
  - [ ] neurogo plugin for vscode allows direct file edits
  - [ ] Review existing .ns skills (e.g., CommitChanges) for suitability/ease of conversion to built-in Go TOOLs.
  - [ ] Select worker from roles list
  - [ ] Add ns file icon
  - [ ] make update project turn things into proper links
  - [ ] better words than CALL LLM (consult? more general)
  - [ ] Create documentation for all NeuroScript built-in tools in `docs/ns/tools/` following `tool_spec_structure.md`.