## D. Completed Features

[x] neurogo able to execute basic ns (SET, CALL, RETURN, basic IF/WHILE/FOR headers and block execution) [interpreter_simple_steps.go](./pkg/core/interpreter_control_flow.go, pkg/core/interpreter_simple_steps.go)
[x] Basic Arithmetic Evaluation (+, -, *, /, %, **, unary -) [evaluation_operators.go](./pkg/core/evaluation_logic.go, pkg/core/evaluation_operators.go)
[x] Basic Condition Evaluation (==, !=, >, <, >=, <=, NOT, AND, OR, truthiness) [evaluation_logic.go](./pkg/core/evaluation_comparison.go, pkg/core/evaluation_logic.go)
[x] List ([]) and Map ({}) Literal Parsing & Evaluation [evaluation_main.go](./pkg/core/ast_builder_collections.go, pkg/core/evaluation_main.go)
[x] List/Map Element Access (e.g., list[index], map["key"]) [evaluation_access.go](./pkg/core/ast_builder_terminators.go, pkg/core/evaluation_access.go)
[x] FOR EACH List Element Iteration [interpreter_control_flow.go](./pkg/core/interpreter_control_flow.go)
[x] FOR EACH Map Key Iteration [interpreter_control_flow.go](./pkg/core/interpreter_control_flow.go)
[x] FOR EACH String Character/Comma Iteration [interpreter_control_flow.go](./pkg/core/interpreter_control_flow.go)
[x] ELSE Block Execution [interpreter_control_flow.go](./pkg/core/ast_builder_blocks.go, pkg/core/interpreter_control_flow.go)
[x] Basic set of golang tools in neurogo (FS, Git, Mock Vector, Strings, Shell, Go fmt/build/test/check/mod, Metadata, Checklist, Blocks, Math, Lists) [tools_list_register.go](./pkg/core/tools_register.go, pkg/core/tools_string.go, pkg/core/tools_fs.go, pkg/core/tools_git.go, pkg/core/tools_shell.go, pkg/core/tools_math.go, pkg/core/tools_metadata.go, pkg/neurodata/checklist/checklist_tool.go, pkg/neurodata/blocks/blocks_tool.go, pkg/core/tools_list_register.go)
[x] In-memory vector DB implemented (mocked, VectorUpdate, SearchSkills) [embeddings.go](./pkg/core/tools_vector.go, pkg/core/embeddings.go)
[x] LLM Integration via CALL LLM (Gemini) [interpreter_simple_steps.go](./pkg/core/llm.go, pkg/core/interpreter_simple_steps.go)
[x] Basic CLI Runner (neurogo) with debug flags [config.go](./cmd/neurogo/main.go, pkg/neurogo/app.go, pkg/neurogo/config.go)
[x] neurogo skips loading ns files with errors gracefully [parser_api.go](./pkg/neurogo/app_script.go, pkg/core/parser_api.go)
[x] ns stored in git (manually, but tools support adding/committing) [orchestrator.ns.txt](./pkg/core/tools_git.go, library/orchestrator.ns.txt)
[x] Bootstrap Skills: Create initial .ns.txt skills (HandleSkillRequest, CommitChanges, UpdateNsSyntax, etc.) [UpdateNsSyntax.ns.txt](./library/orchestrator.ns.txt, library/UpdateNsSyntax.ns.txt) (Note: CommitChanges likely superseded by TOOL.GitCommit)
[x] Basic Core Syntax Parsing (DEFINE PROCEDURE, COMMENT:, SET, CALL, RETURN, END) [neuroscript_parser.go](./pkg/core/generated/neuroscript_parser.go)
[x] Structured Docstring Parsing (COMMENT: block content parsed into struct, includes LANG_VERSION) [utils.go](./pkg/core/ast_builder_procedures.go, pkg/core/utils.go)
[x] Block Header Parsing (IF...THEN, WHILE...DO, FOR EACH...DO) [neuroscript_parser.go](./pkg/core/generated/neuroscript_parser.go)
[x] Block Termination Parsing (ENDBLOCK) [neuroscript_parser.go](./pkg/core/generated/neuroscript_parser.go)
[x] Line Continuation Parsing (Supported via \ in SET prompt example) [modify_and_build.ns.txt](./library/modify_and_build.ns.txt) (Partially done)
[x] Basic Expression Evaluation (String/Num/Bool Literals, Variables, LAST, Parentheses) [evaluation_test.go](./pkg/core/evaluation_main.go, pkg/core/evaluation_test.go)
[x] EVAL() Function Parsing & Evaluation (Explicit Placeholder Resolution) [evaluation_resolve.go](./pkg/core/evaluation_main.go, pkg/core/evaluation_resolve.go)
[x] List ([]) and Map ({}) Literal Parsing & Evaluation [evaluation_main.go](./pkg/core/ast_builder_collections.go, pkg/core/evaluation_main.go)
[x] List/Map Element Access (list[index], map["key"]) [evaluation_access.go](./pkg/core/ast_builder_terminators.go, pkg/core/evaluation_access.go)
[x] Basic Arithmetic Evaluation (+, -, *, /, %, **, unary -) [evaluation_operators.go](./pkg/core/evaluation_logic.go, pkg/core/evaluation_operators.go)
[x] String Concatenation (+) [evaluation_operators.go](./pkg/core/evaluation_operators.go)
[x] Logical Operators (AND, OR, NOT - includes short-circuiting) [evaluation_logical_bitwise_test.go](./pkg/core/evaluation_logic.go, pkg/core/evaluation_logical_bitwise_test.go)
[x] Bitwise Operators (&, |, ^, ~) [evaluation_operators.go](./pkg/core/evaluation_logic.go, pkg/core/evaluation_operators.go, pkg/core/evaluation_logical_bitwise_test.go)
[x] Built-in Math Functions (LN, LOG, SIN, COS, TAN, ASIN, ACOS, ATAN) [evaluations_functions_test.go](./pkg/core/evaluation_logic.go, pkg/core/evaluations_functions_test.go)
[x] Operator Precedence Handling (via Grammar/AST) [ast_builder_operators.go](./pkg/core/generated/neuroscript_parser.go, pkg/core/ast_builder_operators.go)
[x] Basic Interpreter Structure (Interpreter, Scope, RunProcedure) [interpreter.go](./pkg/core/interpreter.go)
[x] Interpreter Block Execution (IF/ELSE/WHILE/FOR, RETURN propagation) [interpreter_control_flow.go](./pkg/core/interpreter_control_flow.go)
[x] Interpreter: Implement FOR EACH List Element Iteration [interpreter_control_flow.go](./pkg/core/interpreter_control_flow.go)
[x] Interpreter: Implement FOR EACH Map Key Iteration [interpreter_control_flow.go](./pkg/core/interpreter_control_flow.go)
[x] Interpreter: Implement FOR EACH String Character/Comma Iteration [interpreter_control_flow.go](./pkg/core/interpreter_control_flow.go)
[x] CALL LLM Integration (via llm.go) [interpreter_simple_steps.go](./pkg/core/llm.go, pkg/core/interpreter_simple_steps.go)
[x] CALL TOOL Mechanism & Argument Validation/Conversion [interpreter_simple_steps.go](./pkg/core/tools_validation.go, pkg/core/interpreter_simple_steps.go)
[x] CALL TOOL requires `TOOL.` prefix for built-ins [interpreter_test.go](./pkg/core/interpreter_simple_steps.go, pkg/core/interpreter_test.go)
[x] File Version Declaration Parsing (FILE_VERSION) [ast_builder_main.go](./pkg/core/generated/neuroscript_parser.go, pkg/core/ast_builder_main.go)
[x] Basic FS Tools Implemented (ReadFile, WriteFile, ListDirectory, LineCountFile, SanitizeFilename) [tools_fs.go](./pkg/core/tools_fs.go)
[x] Basic Git Tools Implemented (GitAdd, GitCommit) [tools_git.go](./pkg/core/tools_git.go)
[x] String Tools Implemented (StringLength, Substring, ToUpper, ToLower, TrimSpace, SplitString, SplitWords, JoinStrings, ReplaceAll, Contains, HasPrefix, HasSuffix, LineCountString) [tools_string.go](./pkg/core/tools_string.go)
[x] Shell/Go Tools Implemented (ExecuteCommand, GoBuild, GoCheck, GoTest, GoFmt, GoModTidy) [tools_shell.go](./pkg/core/tools_shell.go)
[x] Mock Vector DB Tools (VectorUpdate, SearchSkills) [embeddings.go](./pkg/core/tools_vector.go, pkg/core/embeddings.go)
[x] Basic CLI Runner (neurogo) [app.go](./cmd/neurogo/main.go, pkg/neurogo/app.go)
[x] Debug Flags and Conditional Logging in neurogo [config.go](./pkg/neurogo/app.go, pkg/neurogo/config.go)
[x] Graceful skipping of files with parse errors in neurogo [parser_api.go](./pkg/neurogo/app_script.go, pkg/core/parser_api.go)
[x] Fenced Code Block Extraction (including metadata) [blocks_tool.go](./pkg/neurodata/blocks/blocks_extractor.go, pkg/neurodata/blocks/blocks_tool.go)
[x] Updated neurogo CLI args (-lib flag, proc/file target, agent mode flags) [config.go](./pkg/neurogo/config.go)
[x] Metadata Extraction Tool (TOOL.ExtractMetadata) [tools_metadata.go](./pkg/core/tools_metadata.go)
[x] Checklist Parser (String manipulation based) [scanner_parser.go](./pkg/neurodata/checklist/scanner_parser.go)
[x] Checklist Parser Tool (TOOL.ParseChecklistFromString) [checklist_tool.go](./pkg/neurodata/checklist/checklist_tool.go)
[x] NeuroData Block Parsing (Blocks tool) [blocks_tool.go](./pkg/neurodata/blocks/blocks_tool.go)
[x] Agent Mode Framework (Conversation, Security Layer, LLM Client, Tool Declarations) [llm_tools.go](./pkg/neurogo/app_agent.go, pkg/core/conversation.go, pkg/core/security.go, pkg/core/llm.go, pkg/core/llm_tools.go)
[x] Allowlist/Denylist Loading and Enforcement in Agent Mode [security.go](./pkg/neurogo/app_agent.go, pkg/core/security.go)
[x] Basic Path Sandboxing (via SecureFilePath) [security.go](./pkg/core/tools_helpers.go, pkg/core/security.go)
[x] Basic List Tools (ListLength, ListIsEmpty) [tools_list_impl.go](./pkg/core/tools_list_impl.go)
[x] Review versioning: Moved language version into docstring block (`LANG_VERSION:`) and clarified `FILE_VERSION` vs `:: version:` usage. [metadata.md](./docs/script spec.md, docs/metadata.md)
[x] Only load skills when requested (Implicitly true as they are loaded from files on demand or via lib path scan)
[x] More tests for securefile root [tools_fs_list_test.go](./pkg/core/tools_fs_read_test.go, pkg/core/tools_fs_write_test.go, pkg/core/tools_fs_list_test.go)
[x] Allow LLM to use local tools back (Implemented via Agent Mode) [security.go](./pkg/neurogo/app_agent.go, pkg/core/security.go)
[x] Eval tool for arith etc (Arithmetic implemented as operators, not standalone EVAL tool for *just* math)
[x] Files LLM allowed to see (Handled by Sandbox + Allowlist) [config.go](./pkg/core/security.go, pkg/neurogo/config.go)
[x] LLM selection (Via SetModel method, default model constant) [llm.go](./pkg/core/llm.go)
[x] neurogo as local agent for LLM [app_agent.go](./pkg/neurogo/app_agent.go)
[x] Restricted mode design exists [restricted_mode.md](./docs/restricted_mode.md) (Marked as design exists)
[x] Tools: Implement Real In-Memory Vector DB (Functional mock exists) [embeddings.go](./pkg/core/tools_vector.go, pkg/core/embeddings.go)
[x] LLM Gateway: Make LLM endpoint/model configurable (Client supports SetModel, default constant exists) [llm.go](./pkg/core/llm.go)
[x] Basic Math Tools (TOOL.Add, Subtract, Multiply, Divide, Modulo) [tools_math.go](./pkg/core/tools_math.go)
[x] NeuroData Checklist Parser (String Manipulation) [scanner_parser.go](./pkg/neurodata/checklist/scanner_parser.go)
[x] Defined Errors for Checklist Parser (ErrMalformedItem, ErrNoContent) [defined_errors.go](./pkg/neurodata/checklist/defined_errors.go)
[x] Tool for Checklist Parsing (TOOL.ParseChecklistFromString) [checklist_tool.go](./pkg/neurodata/checklist/checklist_tool.go)
[x] Updated test error checking to use `errors.Is` vs `errorContains` appropriately [interpreter_test.go](./pkg/core/testing_helpers_test.go, pkg/core/interpreter_test.go)