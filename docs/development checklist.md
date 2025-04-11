:: version: 0.1.5
:: dependsOn: docs/neuroscript overview.md, pkg/core/*, pkg/neurodata/*
:: howToUpdate: Review checklist against current codebase state (core interpreter features, tools, neurodata parsers) and project goals. Mark completed items, add new tasks, adjust priorities. Increment patch version.

# NeuroScript Development Checklist (v6 - Added Built-in Review Item)

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
    [x] checklist (Parser exists, Tooling/Integration TBD) [cite: uploaded:neuroscript/pkg/neurodata/checklist/checklist_parser.go]
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

**Tooling & Integration (Supporting Self-Management):**
[ ] Tools: Implement Real In-Memory Vector DB (VectorUpdate, SearchSkills) (Currently mocked)
[ ] Tools: Enhance Git Workflow (Add Branch support, GitPull?, Auto-index after commit)
[ ] Tools: TOOL.NeuroScriptCheckSyntax(content) - Formal syntax check tool using the parser.
[ ] Feature: nsfmt - A dedicated formatting tool/procedure.
[ ] Feature: Embed standard utility NeuroScripts (e.g., CommitChanges) into neurogo binary (using Go embed)
[ ] LLM Gateway: Make LLM endpoint/model configurable [cite: uploaded:neuroscript/pkg/core/llm.go]

**Tooling (General Purpose):**
[ ] Tools: Add more utility tools (JSON, HTTP, etc.)
[ ] Tools: Add Markdown tools (Read/Write)
[ ] Tools: Add Structured Document tools
[ ] Tools: Add Table tools
[ ] Tools: Add Integration tools (Sheets, Docs)
[ ] Tools: Add data encoding/hardening tools (e.g., Base32, Base64, potentially zip/unzip) for reliable data transfer. **(NEW)**
[ ] Tools: grep/egrep/agrep

**Longer Term / Advanced:**
[ ] Feature: Add Self-test support in ns
[ ] Interpreter: Implement Context Management Strategy for CALL LLM -- defer
[ ] Block and file level prior version preservation
[ ] Restricted mode for running untrusted scripts

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
[ ] Strong list manipulation (cf lisp)
[ ] Prolog style features
[ ] SVG generation and manipulation
[ ] Only load skills when requested
[ ] More tests for securefile root
[ ] Allow LLM to use local tools back
[ ] Nice example on website/readme
[ ] Logo
[ ] Eval tool for arith etc
[ ] Files LLM allowed to see
[ ] LLM selection
[ ] More self building and maint
[ ] keep prior versions meta tag
[ ] neurogo as local agent for LLM
[ ] neurogo plugin for vscode allows direct file edits
[ ] **Review existing .ns skills (e.g., CommitChanges) for suitability/ease of conversion to built-in Go TOOLs.** **(NEW)**
[ ] allow line continuation in ns

## C. Found work and things to go back to

[ ] Tools/Known Issue: `TOOL.GoBuild` and `TOOL.GoCheck` error reporting for single files needs improvement. [cite: uploaded:neuroscript/docs/development checklist.md]
[ ] Add ns file icon [cite: uploaded:neuroscript/docs/development checklist.md]
[ ] Review versioning: Move language version into docstring block (`LANG_VERSION:`) and clarify `FILE_VERSION` usage/automation. [cite: uploaded:neuroscript/docs/script spec.md, uploaded:neuroscript/docs/development checklist.md]

## D. Completed Features

[x] neurogo able to execute basic ns (SET, CALL, RETURN, basic IF/WHILE/FOR headers and block execution) [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go, uploaded:neuroscript/pkg/core/interpreter_simple_steps.go]
[x] Basic Arithmetic Evaluation (+, -, *, /, %, **, unary -) [cite: uploaded:neuroscript/pkg/core/evaluation_logic.go, uploaded:neuroscript/pkg/core/evaluation_operators.go]
[x] Basic Condition Evaluation (==, !=, >, <, >=, <=, NOT, AND, OR, truthiness) [cite: uploaded:neuroscript/pkg/core/evaluation_comparison.go, uploaded:neuroscript/pkg/core/evaluation_logic.go]
[x] List ([]) and Map ({}) Literal Parsing & Evaluation [cite: uploaded:neuroscript/pkg/core/ast_builder_collections.go, uploaded:neuroscript/pkg/core/evaluation_main.go]
[x] List/Map Element Access (e.g., list[index], map["key"]) [cite: uploaded:neuroscript/pkg/core/ast_builder_terminators.go, uploaded:neuroscript/pkg/core/evaluation_access.go]
[x] FOR EACH List Element Iteration [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
[x] FOR EACH Map Key Iteration [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
[x] FOR EACH String Character/Comma Iteration [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
[x] ELSE Block Execution [cite: uploaded:neuroscript/pkg/core/ast_builder_blocks.go, uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
[x] Basic set of golang tools in neurogo (FS, Git, Mock Vector, Strings, Shell, Go fmt/build/test/check/mod) [cite: uploaded:neuroscript/pkg/core/tools_register.go, uploaded:neuroscript/pkg/core/tools_string.go, uploaded:neuroscript/pkg/core/tools_fs.go, uploaded:neuroscript/pkg/core/tools_git.go, uploaded:neuroscript/pkg/core/tools_shell.go]
[x] In-memory vector DB implemented (mocked, VectorUpdate, SearchSkills) [cite: uploaded:neuroscript/pkg/core/tools_vector.go, uploaded:neuroscript/pkg/core/embeddings.go]
[x] LLM Integration via CALL LLM (Gemini) [cite: uploaded:neuroscript/pkg/core/llm.go, uploaded:neuroscript/pkg/core/interpreter_simple_steps.go]
[x] Basic CLI Runner (neurogo) with debug flags [cite: uploaded:neuroscript/neurogo/main.go]
[x] neurogo skips loading ns files with errors gracefully [cite: uploaded:neuroscript/neurogo/main.go, uploaded:neuroscript/pkg/core/parser_api.go]
[x] ns stored in git (manually, but tools support adding/committing) [cite: uploaded:neuroscript/pkg/core/tools_git.go, uploaded:neuroscript/neurogo/skills/commit_changes.ns.txt]
[x] Bootstrap Skills: Create initial .ns.txt skills (HandleSkillRequest, CommitChanges, UpdateNsSyntax, etc.) [cite: uploaded:neuroscript/neurogo/skills/orchestrator.ns.txt, uploaded:neuroscript/neurogo/skills/commit_changes.ns.txt, uploaded:neuroscript/neurogo/skills/UpdateNsSyntax.ns.txt]
[x] Basic Core Syntax Parsing (DEFINE PROCEDURE, COMMENT:, SET, CALL, RETURN, END) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go]
[x] Structured Docstring Parsing (COMMENT: block content parsed into struct, includes LANG_VERSION) [cite: uploaded:neuroscript/pkg/core/ast_builder_procedures.go, uploaded:neuroscript/pkg/core/utils.go]
[x] Block Header Parsing (IF...THEN, WHILE...DO, FOR EACH...DO) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go]
[x] Block Termination Parsing (ENDBLOCK) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go]
[x] Line Continuation Parsing (Handled implicitly by ANTLR lexer) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_lexer.go]
[x] Basic Expression Evaluation (String/Num/Bool Literals, Variables, LAST, Parentheses) [cite: uploaded:neuroscript/pkg/core/evaluation_main.go, uploaded:neuroscript/pkg/core/evaluation_test.go]
[x] EVAL() Function Parsing & Evaluation (Explicit Placeholder Resolution) [cite: uploaded:neuroscript/pkg/core/evaluation_main.go, uploaded:neuroscript/pkg/core/evaluation_resolve.go]
[x] List ([]) and Map ({}) Literal Parsing & Evaluation [cite: uploaded:neuroscript/pkg/core/ast_builder_collections.go, uploaded:neuroscript/pkg/core/evaluation_main.go]
[x] List/Map Element Access (`list[index]`, `map["key"]`) [cite: uploaded:neuroscript/pkg/core/ast_builder_terminators.go, uploaded:neuroscript/pkg/core/evaluation_access.go]
[x] Basic Arithmetic Evaluation (+, -, *, /, %, **, unary -) [cite: uploaded:neuroscript/pkg/core/evaluation_logic.go, uploaded:neuroscript/pkg/core/evaluation_operators.go]
[x] String Concatenation (+) [cite: uploaded:neuroscript/pkg/core/evaluation_operators.go]
[x] Basic Condition Evaluation (==, !=, >, <, >=, <=, truthiness) [cite: uploaded:neuroscript/pkg/core/evaluation_comparison.go]
[x] Logical Operators (AND, OR, NOT - includes short-circuiting) [cite: uploaded:neuroscript/pkg/core/evaluation_logic.go, uploaded:neuroscript/pkg/core/evaluation_logical_bitwise_test.go]
[x] Bitwise Operators (&, |, ^) [cite: uploaded:neuroscript/pkg/core/evaluation_logic.go, uploaded:neuroscript/pkg/core/evaluation_operators.go]
[x] Built-in Math Functions (LN, LOG, SIN, COS, TAN, ASIN, ACOS, ATAN) [cite: uploaded:neuroscript/pkg/core/evaluation_logic.go, uploaded:neuroscript/pkg/core/evaluations_functions_test.go]
[x] Operator Precedence Handling (via Grammar/AST) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go, uploaded:neuroscript/pkg/core/ast_builder_operators.go]
[x] Basic Interpreter Structure (Interpreter, Scope, RunProcedure) [cite: uploaded:neuroscript/pkg/core/interpreter.go]
[x] Interpreter Block Execution (IF/ELSE/WHILE/FOR, RETURN propagation) [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
[x] Interpreter: Implement FOR EACH List Element Iteration [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
[x] Interpreter: Implement FOR EACH Map Key Iteration [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
[x] Interpreter: Implement FOR EACH String Character/Comma Iteration [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
[x] CALL LLM Integration (via llm.go) [cite: uploaded:neuroscript/pkg/core/llm.go, uploaded:neuroscript/pkg/core/interpreter_simple_steps.go]
[x] CALL TOOL Mechanism & Argument Validation/Conversion [cite: uploaded:neuroscript/pkg/core/tools.go, uploaded:neuroscript/pkg/core/interpreter_simple_steps.go]
[x] File Version Declaration Parsing (`FILE_VERSION`) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go, uploaded:neuroscript/pkg/core/ast_builder_main.go]
[x] Basic Tools Implemented (ReadFile, WriteFile, SanitizeFilename, GitAdd, GitCommit, ListDirectory, LineCount) [cite: uploaded:neuroscript/pkg/core/tools_fs.go, uploaded:neuroscript/pkg/core/tools_git.go, uploaded:neuroscript/pkg/core/tools_register.go]
[x] String Tools Implemented (StringLength, Substring, ToUpper, ToLower, TrimSpace, SplitString, SplitWords, JoinStrings, ReplaceAll, Contains, HasPrefix, HasSuffix) [cite: uploaded:neuroscript/pkg/core/tools_string.go]
[x] Shell/Go Tools Implemented (ExecuteCommand, GoBuild, GoCheck, GoTest, GoFmt, GoModTidy) [cite: uploaded:neuroscript/pkg/core/tools_shell.go]
[x] Mock Vector DB Tools (VectorUpdate, SearchSkills) [cite: uploaded:neuroscript/pkg/core/tools_vector.go, uploaded:neuroscript/pkg/core/embeddings.go]
[x] Basic CLI Runner (neurogo) [cite: uploaded:neuroscript/neurogo/main.go]
[x] Debug Flags and Conditional Logging in neurogo [cite: uploaded:neuroscript/neurogo/main.go]
[x] Graceful skipping of files with parse errors in neurogo [cite: uploaded:neuroscript/neurogo/main.go, uploaded:neuroscript/pkg/core/parser_api.go]
[x] Bootstrap Skills: Create initial .ns.txt skills (HandleSkillRequest, CommitChanges, UpdateNsSyntax, etc.) [cite: uploaded:neuroscript/neurogo/skills/orchestrator.ns.txt, uploaded:neuroscript/neurogo/skills/commit_changes.ns.txt, uploaded:neuroscript/neurogo/skills/UpdateNsSyntax.ns.txt]
[x] Fenced Code Block Extraction (including metadata) [cite: uploaded:neuroscript/pkg/neurodata/blocks/blocks_extractor.go, uploaded:neuroscript/pkg/neurodata/blocks/blocks_tool.go, uploaded:neuroscript/pkg/neurodata/blocks/blocks_complex_test.go]
[x] Updated neurogo CLI args (-lib flag, proc/file target) [cite: uploaded:neuroscript/neurogo/main.go]