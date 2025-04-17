:: version: 0.1.8 // Increment when changes are made
:: dependsOn: docs/neuroscript overview.md, pkg/core/, pkg/neurodata/, pkg/neurogo/app_script.go, pkg/neurogo/app_agent.go
:: howToUpdate: Review checklist against current codebase state (core interpreter features, tools, neurodata parsers, neurogo app structure) and project goals. Mark completed items, add new tasks, adjust priorities. Increment patch version. Move completed items to completed.md

# NeuroScript Development Checklist

## Goal: Reach "bootstrapping" point

--  NeuroScript project can maintain itself:
    1.  update docs based on progress
    2.  update scripts based on changes in syntax
    3.  run tests and recompile on change
    4.  update source code based on prompt
    5.  fix source code based on tests

-- neurogo provides full ns capabilities:
    1. vector db of scripts & formats with retrieval
    2. full neurogo/LLM comms with basic security
    3. conversational mode (with human)

## A. Planned Features (Reordered for Bootstrapping/Dependencies)

**Core Language / Interpreter Refinements:**
[P] Interpreter: Add NeuroScript-specific Error Handling (e.g., TRY/CATCH or specific error types?) // Basic Go error handling (sentinels, wrapping) is implemented, but no NS-level TRY/CATCH.
[P] NeuroData files, template and instance (Design & Implement) // Specs exist for many, core block/metadata/checklist impl exists, but specific format parsers/tools mostly missing.
    [P] graph // Spec exists
    [P] table // Spec exists
    [P] tree // Spec exists
    [P] decision_table // Spec exists
    [P] form // Spec exists
    [ ] invoice // Spec not found
    [ ] statement_of_account // Spec not found
    [ ] receipt // Spec not found
    [ ] payment // Spec not found
    [ ] packing_list // Spec not found
    [ ] request_for_quote or estimate // Spec not found
    [ ] quote or estimate // Spec not found
    [ ] purchase_order // Spec not found
    [ ] work_order // Spec not found
    [ ] markdown_doc // Spec not found
    [P] composite_doc // Spec exists (composite_file_spec.md), block extractor impl
    [ ] bug_report // Spec not found
    [ ] ns_tool_list // Spec not found
    [P] enum // Spec exists
    [ ] roles_list // Spec not found
    [P] object // Design mentioned in form.md, uses map literal parsing
    [ ] ishikawa // Spec not found
    [ ] kanban // Spec not found
    [ ] memo // Spec not found
    [P] list // Spec exists
    [X] checklist // Spec exists, Impl exists

**Tooling & Integration (Supporting Self-Management):**
[P] Tools: Enhance Git Workflow (Add Branch support, GitPull?, Auto-index after commit) // Basic Git tools exist (Commit, Add, Status), but enhancements like branch support are not implemented.
[ ] Tools: TOOL.NeuroScriptCheckSyntax(content) - Formal syntax check tool using the parser. // Parser exists, but not exposed as a tool.
[ ] Feature: nsfmt - A dedicated formatting tool/procedure.
[ ] Feature: Embed standard utility NeuroScripts (e.g., CommitChanges) into neurogo binary (using Go embed) // Likely outside pkg/core, needs check elsewhere.

**Tooling (General Purpose):**
[P] Tools: List Manipulation (Expand on existing capabilities?) // Many list tools implemented, but some missing (Head, Rest, Tail).
    [X] TOOL.ListAppend(list, element) -> list
    [X] TOOL.ListPrepend(list, element) -> list
    [X] TOOL.ListGet(list, index, [default]) -> any
    [X] TOOL.ListSlice(list, start, end) -> list
    [X] TOOL.ListContains(list, element) -> bool
    [X] TOOL.ListReverse(list) -> list
    [X] TOOL.ListSort(list) -> list (Current implementation handles numeric/string, review scope)
    [ ] TOOL.ListHead(list) -> any
    [ ] TOOL.ListRest(list) -> list (cdr equivalent)
    [ ] TOOL.ListTail(list, count) -> list (last N elements)
[ ] Tools: Add more utility tools (JSON, HTTP, etc.)
[ ] Tools: Add Markdown tools (Read/Write)
[ ] Tools: Add Structured Document tools
[ ] Tools: Add Table tools
[ ] Tools: Add Integration tools (Sheets, Docs)
[ ] Tools: Add data encoding/hardening tools (e.g., Base32, Base64, potentially zip/unzip) for reliable data transfer.
[ ] Tools: grep/egrep/agrep
[ ] Tools: UploadToFilesAPI
[ ] Tools: SyncToFilesAPI

**Longer Term / Advanced:**
[ ] Feature: Add Self-test support in ns
[ ] Interpreter: Implement Context Management Strategy for CALL LLM -- defer // Basic LLM call exists, but no advanced context management.
[ ] Block and file level prior version preservation
[P] Implement Restricted mode for running untrusted scripts (Design exists, implementation needed) // Security foundations exist, but full restricted mode not implemented.

## B. Various "do soon" things

[ ] LLM able to read ns and execute it (via prompt guidance)
[ ] LLM able to translate simple ns into golang tool
[ ] Use git branch for version control within tools
[ ] Markdown tools (r & w)
[ ] Structured document tools (hierarchical info/docs)
[ ] Table tools
[ ] Integration tools (e.g., Google Sheets and Docs)
[ ] Self-test support in ns
[ ] LLMs can supply git-style patches and have them applied to files
[ ] MCP support
[ ] Ability to pass text from LLM to tool in (BASE64) or some other armored format
[ ] Prolog style features
[ ] SVG generation and manipulation
[ ] fuzzy logic tools
[ ] Nice example on website/readme
[ ] Logo
[ ] More self building and maint
[ ] keep prior versions meta tag
[ ] neurogo plugin for vscode allows direct file edits
[ ] Review existing .ns skills (e.g., CommitChanges) for suitability/ease of conversion to built-in Go TOOLs.
[ ] Select worker from roles list
[ ] block read to support block references
[ ] also read from string block references

## C. Found work and things to go back to

[ ] Tools/Known Issue: TOOL.GoBuild and TOOL.GoCheck error reporting for single files needs improvement. // No GoBuild/GoCheck tools found in pkg/core, likely in neurogo or scripts? Status unchanged.
[ ] Add ns file icon
[ ] make update project turn things into proper links
[ ] better words than CALL LLM (consult? more general)
[ ] Create documentation for all NeuroScript built-in tools in `docs/ns/tools/` following `tool_spec_structure.md`.