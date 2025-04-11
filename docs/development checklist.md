:: version: 0.1.6
:: dependsOn: docs/neuroscript overview.md, pkg/core/, pkg/neurodata/, pkg/neurogo/app_script.go, pkg/neurogo/app_agent.go
:: howToUpdate: Review checklist against current codebase state (core interpreter features, tools, neurodata parsers, neurogo app structure) and project goals. Mark completed items, add new tasks, adjust priorities. Increment patch version.

# NeuroScript Development Checklist (v7 - Updated based on source review)

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

Core Language / Interpreter Refinements:
[ ] Interpreter: Add NeuroScript-specific Error Handling (e.g., TRY/CATCH or specific error types?)
[ ] NeuroData files, template and instance (Design & Implement)
    [x] checklist (Parser exists, Tooling/Integration TBD) [cite: uploaded:neuroscript/pkg/neurodata/checklist/scanner_parser.go, uploaded:neuroscript/pkg/neurodata/checklist/checklist_tool.go]
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

Tooling & Integration (Supporting Self-Management):
[x] Tools: Implement Real In-Memory Vector DB (VectorUpdate, SearchSkills) (Currently mocked, but functional mock exists) [cite: uploaded:neuroscript/pkg/core/tools_vector.go, uploaded:neuroscript/pkg/core/embeddings.go]
[ ] Tools: Enhance Git Workflow (Add Branch support, GitPull?, Auto-index after commit)
[ ] Tools: TOOL.NeuroScriptCheckSyntax(content) - Formal syntax check tool using the parser.
[ ] Feature: nsfmt - A dedicated formatting tool/procedure.
[ ] Feature: Embed standard utility NeuroScripts (e.g., CommitChanges) into neurogo binary (using Go embed)
[x] LLM Gateway: Make LLM endpoint/model configurable [cite: uploaded:neuroscript/pkg/core/llm.go] (Client supports SetModel, default model constant exists)

Tooling (General Purpose):
[ ] Tools: Add more utility tools (JSON, HTTP, etc.)
[ ] Tools: Add Markdown tools (Read/Write)
[ ] Tools: Add Structured Document tools
[ ] Tools: Add Table tools
[ ] Tools: Add Integration tools (Sheets, Docs)
[ ] Tools: Add data encoding/hardening tools (e.g., Base32, Base64, potentially zip/unzip) for reliable data transfer. (NEW)
[ ] Tools: grep/egrep/agrep

Longer Term / Advanced:
[ ] Feature: Add Self-test support in ns
[ ] Interpreter: Implement Context Management Strategy for CALL LLM -- defer
[ ] Block and file level prior version preservation
[x] Restricted mode for running untrusted scripts (Design exists, implementation pending?) [cite: uploaded:neuroscript/docs/restricted_mode.md] (Marked as design exists)

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
[x] More tests for securefile root [cite: uploaded:neuroscript/pkg/core/tools_fs_read_test.go, uploaded:neuroscript/pkg/core/tools_fs_write_test.go, uploaded:neuroscript/pkg/core/tools_fs_list_test.go] (Tests added in fs tests)
[x] Allow LLM to use local tools back (Implemented via Agent Mode) [cite: uploaded:neuroscript/pkg/neurogo/app_agent.go, uploaded:neuroscript/pkg/core/security.go]
[ ] Nice example on website/readme
[ ] Logo
[ ] Eval tool for arith etc (Arithmetic implemented, but not isolated EVAL tool)
[x] Files LLM allowed to see (Handled by Sandbox + Allowlist) [cite: uploaded:neuroscript/pkg/core/security.go, uploaded:neuroscript/pkg/neurogo/config.go]
[x] LLM selection (Via SetModel) [cite: uploaded:neuroscript/pkg/core/llm.go]
[ ] More self building and maint
[ ] keep prior versions meta tag
[x] neurogo as local agent for LLM [cite: uploaded:neuroscript/pkg/neurogo/app_agent.go]
[ ] neurogo plugin for vscode allows direct file edits
[ ] Review existing .ns skills (e.g., CommitChanges) for suitability/ease of conversion to built-in Go TOOLs. (NEW)
[ ] allow line continuation in ns (Supported via \ in SET prompt example, but not fully general) [cite: uploaded:neuroscript/library/modify_and_build.ns.txt] (Partially done)

## C. Found work and things to go back to

[ ] Tools/Known Issue: TOOL.GoBuild and TOOL.GoCheck error reporting for single files needs improvement. [cite: uploaded:neuroscript/docs/development checklist.md]
[ ] Add ns file icon [cite: uploaded:neuroscript/docs/development checklist.md]
[x] Review versioning: Move language version into docstring block (LANG_VERSION:) and clarify FILE_VERSION usage/automation. [cite: uploaded:neuroscript/docs/script spec.md, uploaded:neuroscript/docs/metadata.md]

## D. Completed Features

[x] neurogo able to execute basic ns (SET, CALL, RETURN, basic IF/WHILE/FOR headers and block execution) [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go, uploaded:neuroscript/pkg/core/interpreter_simple_steps.go]
[x] Basic Arithmetic Evaluation (+, -, *, /, %, **, unary -) [cite: uploaded:neuroscript/pkg/core/evaluation_logic.go, uploaded:neuroscript/pkg/core/evaluation_operators.go]
[x] Basic Condition Evaluation (==, !=, >, &lt;, >=, &lt;=, NOT, AND, OR, truthiness) [cite: uploaded:neuroscript/pkg/core/evaluation_comparison.go, uploaded:neuroscript/pkg/core/evaluation_logic.go]
[x] List ([]) and Map ({}) Literal Parsing & Evaluation [cite: uploaded:neuroscript/pkg/core/ast_builder_collections.go, uploaded:neuroscript/pkg/core/evaluation_main.go]
[x] List/Map Element Access (e.g., list[index], map["key"]) [cite: uploaded:neuroscript/pkg/core/ast_builder_terminators.go, uploaded:neuroscript/pkg/core/evaluation_access.go]
[x] FOR EACH List Element Iteration [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
[x] FOR EACH Map Key Iteration [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
[x] FOR EACH String Character/Comma Iteration [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
[x] ELSE Block Execution [cite: uploaded:neuroscript/pkg/core/ast_builder_blocks.go, uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
[x] Basic set of golang tools in neurogo (FS, Git, Mock Vector, Strings, Shell, Go fmt/build/test/check/mod) [cite: uploaded:neuroscript/pkg/core/tools_register.go, uploaded:neuroscript/pkg/core/tools_string.go, uploaded:neuroscript/pkg/core/tools_fs.go, uploaded:neuroscript/pkg/core/tools_git.go, uploaded:neuroscript/pkg/core/tools_shell.go, uploaded:neuroscript/pkg/core/tools_math.go, uploaded:neuroscript/pkg/core/tools_metadata.go]
[x] In-memory vector DB implemented (mocked, VectorUpdate, SearchSkills) [cite: uploaded:neuroscript/pkg/core/tools_vector.go, uploaded:neuroscript/pkg/core/embeddings.go]
[x] LLM Integration via CALL LLM (Gemini) [cite: uploaded:neuroscript/pkg/core/llm.go, uploaded:neuroscript/pkg/core/interpreter_simple_steps.go]
[x] Basic CLI Runner (neurogo) with debug flags [cite: uploaded:neuroscript/cmd/neurogo/main.go, uploaded:neuroscript/pkg/neurogo/app.go]
[x] neurogo skips loading ns files with errors gracefully [cite: uploaded:neuroscript/pkg/neurogo/app_script.go, uploaded:neuroscript/pkg/core/parser_api.go]
[x] ns stored in git (manually, but tools support adding/committing) [cite: uploaded:neuroscript/pkg/core/tools_git.go, uploaded:neuroscript/library/orchestrator.ns.txt]
[x] Bootstrap Skills: Create initial .ns.txt skills (HandleSkillRequest, CommitChanges, UpdateNsSyntax, etc.) [cite: uploaded:neuroscript/library/orchestrator.ns.txt, uploaded:neuroscript/library/UpdateNsSyntax.ns.txt] (CommitChanges seems replaced by TOOL.GitCommit)
[x] Basic Core Syntax Parsing (DEFINE PROCEDURE, COMMENT:, SET, CALL, RETURN, END) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go]
[x] Structured Docstring Parsing (COMMENT: block content parsed into struct, includes LANG_VERSION) [cite: uploaded:neuroscript/pkg/core/ast_builder_procedures.go, uploaded:neuroscript/pkg/core/utils.go]
[x] Block Header Parsing (IF...THEN, WHILE...DO, FOR EACH...DO) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go]
[x] Block Termination Parsing (ENDBLOCK) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go]
[x] Line Continuation Parsing (Handled implicitly by ANTLR lexer) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_lexer.go] (Note: Only specifically tested in one script example)
[x] Basic Expression Evaluation (String/Num/Bool Literals, Variables, LAST, Parentheses) [cite: uploaded:neuroscript/pkg/core/evaluation_main.go, uploaded:neuroscript/pkg/core/evaluation_test.go]
[x] EVAL() Function Parsing & Evaluation (Explicit Placeholder Resolution) [cite: uploaded:neuroscript/pkg/core/evaluation_main.go, uploaded:neuroscript/pkg/core/evaluation_resolve.go]
[x] List ([]) and Map ({}) Literal Parsing & Evaluation [cite: uploaded:neuroscript/pkg/core/ast_builder_collections.go, uploaded:neuroscript/pkg/core/evaluation_main.go]
[x] List/Map Element Access (list[index], map["key"]) [cite: uploaded:neuroscript/pkg/core/ast_builder_terminators.go, uploaded:neuroscript/pkg/core/evaluation_access.go]
[x] Basic Arithmetic Evaluation (+, -, *, /, %, *, unary -) [cite: uploaded:neuroscript/pkg/core/evaluation_logic.go, uploaded:neuroscript/pkg/core/evaluation_operators.go]
[x] String Concatenation (+) [cite: uploaded:neuroscript/pkg/core/evaluation_operators.go]
[x] Basic Condition Evaluation (==, !=, >, &lt;, >=, &lt;=, truthiness) [cite: uploaded:neuroscript/pkg/core/evaluation_comparison.go, uploaded:neuroscript/pkg/core/evaluation_logic.go]
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
[x] CALL TOOL Mechanism & Argument Validation/Conversion [cite: uploaded:neuroscript/pkg/core/tools_validation.go, uploaded:neuroscript/pkg/core/interpreter_simple_steps.go]
[x] File Version Declaration Parsing (FILE_VERSION) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go, uploaded:neuroscript/pkg/core/ast_builder_main.go]
[x] Basic Tools Implemented (ReadFile, WriteFile, SanitizeFilename, GitAdd, GitCommit, ListDirectory, LineCount) [cite: uploaded:neuroscript/pkg/core/tools_fs.go, uploaded:neuroscript/pkg/core/tools_git.go, uploaded:neuroscript/pkg/core/tools_register.go] (*LineCount split into LineCountFile/LineCountString)
[x] String Tools Implemented (StringLength, Substring, ToUpper, ToLower, TrimSpace, SplitString, SplitWords, JoinStrings, ReplaceAll, Contains, HasPrefix, HasSuffix, LineCountString) [cite: uploaded:neuroscript/pkg/core/tools_string.go]
[x] Shell/Go Tools Implemented (ExecuteCommand, GoBuild, GoCheck, GoTest, GoFmt, GoModTidy) [cite: uploaded:neuroscript/pkg/core/tools_shell.go]
[x] Mock Vector DB Tools (VectorUpdate, SearchSkills) [cite: uploaded:neuroscript/pkg/core/tools_vector.go, uploaded:neuroscript/pkg/core/embeddings.go]
[x] Basic CLI Runner (neurogo) [cite: uploaded:neuroscript/cmd/neurogo/main.go, uploaded:neuroscript/pkg/neurogo/app.go]
[x] Debug Flags and Conditional Logging in neurogo [cite: uploaded:neuroscript/pkg/neurogo/app.go, uploaded:neuroscript/pkg/neurogo/config.go]
[x] Graceful skipping of files with parse errors in neurogo [cite: uploaded:neuroscript/pkg/neurogo/app_script.go, uploaded:neuroscript/pkg/core/parser_api.go]
[x] Fenced Code Block Extraction (including metadata) [cite: uploaded:neuroscript/pkg/neurodata/blocks/blocks_extractor.go, uploaded:neuroscript/pkg/neurodata/blocks/blocks_tool.go]
[x] Updated neurogo CLI args (-lib flag, proc/file target, agent mode flags) [cite: uploaded:neuroscript/pkg/neurogo/config.go]
[x] Metadata Extraction Tool (TOOL.ExtractMetadata) [cite: uploaded:neuroscript/pkg/core/tools_metadata.go]
[x] Checklist Parser Tool (TOOL.ParseChecklistFromString) [cite: uploaded:neuroscript/pkg/neurodata/checklist/checklist_tool.go]
[x] NeuroData Block Parsing (Blocks tool) [cite: uploaded:neuroscript/pkg/neurodata/blocks/blocks_tool.go]
[x] Agent Mode Framework (Conversation, Security Layer Stubs, LLM Client, Tool Declarations) [cite: uploaded:neuroscript/pkg/neurogo/app_agent.go, uploaded:neuroscript/pkg/core/conversation.go, uploaded:neuroscript/pkg/core/security.go, uploaded:neuroscript/pkg/core/llm.go, uploaded:neuroscript/pkg/core/llm_tools.go]
[x] Allowlist/Denylist Loading and Enforcement in Agent Mode [cite: uploaded:neuroscript/pkg/neurogo/app_agent.go, uploaded:neuroscript/pkg/core/security.go]
[x] Basic Path Sandboxing (via SecureFilePath) [cite: uploaded:neuroscript/pkg/core/tools_helpers.go, uploaded:neuroscript/pkg/core/security.go]

