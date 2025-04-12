:: version: 0.1.8
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
[ ] Interpreter: Add NeuroScript-specific Error Handling (e.g., TRY/CATCH or specific error types?)
[ ] NeuroData files, template and instance (Design & Implement)
    [ ] graph
    [ ] table
    [ ] tree
    [ ] decision_table
    [ ] form
    [ ] invoice
    [ ] statement_of_account
    [ ] receipt
    [ ] payment
    [ ] packing_list
    [ ] request_for_quote or estimate
    [ ] quote or estimate
    [ ] purchase_order
    [ ] work_order
    [ ] markdown_doc
    [ ] composite_doc
    [ ] bug_report
    [ ] ns_tool_list
    [ ] enum
    [ ] roles_list

**Tooling & Integration (Supporting Self-Management):**
[ ] Tools: Enhance Git Workflow (Add Branch support, GitPull?, Auto-index after commit)
[ ] Tools: TOOL.NeuroScriptCheckSyntax(content) - Formal syntax check tool using the parser.
[ ] Feature: nsfmt - A dedicated formatting tool/procedure.
[ ] Feature: Embed standard utility NeuroScripts (e.g., CommitChanges) into neurogo binary (using Go embed)

**Tooling (General Purpose):**
[ ] Tools: List Manipulation (Expand on existing capabilities?)
    [ ] TOOL.ListAppend(list, element) -> list
    [ ] TOOL.ListPrepend(list, element) -> list
    [ ] TOOL.ListGet(list, index, [default]) -> any
    [ ] TOOL.ListSlice(list, start, end) -> list
    [ ] TOOL.ListContains(list, element) -> bool
    [ ] TOOL.ListReverse(list) -> list
    [ ] TOOL.ListSort(list) -> list (Current implementation handles numeric/string, review scope)
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

**Longer Term / Advanced:**
[ ] Feature: Add Self-test support in ns
[ ] Interpreter: Implement Context Management Strategy for CALL LLM -- defer
[ ] Block and file level prior version preservation
[ ] Implement Restricted mode for running untrusted scripts (Design exists [cite: uploaded:neuroscript/docs/restricted_mode.md], implementation needed)

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

[ ] Tools/Known Issue: TOOL.GoBuild and TOOL.GoCheck error reporting for single files needs improvement. [cite: uploaded:neuroscript/docs/development checklist.md]
[ ] Add ns file icon [cite: uploaded:neuroscript/docs/development checklist.md]
[ ] make update project turn things into proper links
[ ] better words than CALL LLM (consult? more general)

