# NeuroScript Development Checklist (v4 - Post ANTLR/Debug Flags)

Goal: Reach the "bootstrapping" point where NeuroScript, executed by an LLM or gonsi, can use CALL LLM and TOOLs to find, create, and manage NeuroScript skills stored in Git and indexed in a vector DB.

## A. Capabilities (Existing & Target)

[x] gonsi able to execute basic ns (SET, CALL, RETURN, basic IF/WHILE/FOR headers and block execution) [cite: uploaded:neuroscript/pkg/core/interpreter_steps.go, uploaded:neuroscript/pkg/core/evaluation.go]
[x] ns stored in git (manually, but tools support adding/committing) [cite: uploaded:neuroscript/pkg/core/tools_register.go, uploaded:neuroscript/gonsi/skills/commit_changes.ns.txt]
[x] Basic set of golang tools in gonsi (ReadFile, WriteFile, SanitizeFilename, GitAdd, GitCommit, mock DB/Search, String tools) [cite: uploaded:neuroscript/pkg/core/tools_register.go, uploaded:neuroscript/pkg/core/tools_string.go, uploaded:neuroscript/pkg/core/utils.go]
[ ] LLM able to read ns and execute it (via prompt guidance)
[ ] LLM able to translate simple ns into golang tool
[-] Std lib of foundational ns for LLMs to use (e.g., bootstrapping skills - HandleSkillRequest, CommitChanges partially exist) [cite: uploaded:neuroscript/gonsi/skills/orchestrator.ns.txt, uploaded:neuroscript/gonsi/skills/commit_changes.ns.txt]
[ ] Use git branch for version control within tools
[ ] Markdown tools (r & w)
[ ] Structured document tools (hierarchical info/docs)
[ ] Table tools
[ ] Integration tools (e.g., Google Sheets and Docs)
[ ] Self-test support in ns
[x] In-memory vector DB implemented (mocked, VectorUpdate, SearchSkills) [cite: uploaded:neuroscript/pkg/core/tools_register.go, uploaded:neuroscript/pkg/core/embeddings.go]
[x] gonsi skips loading ns files with errors gracefully (confirmed with parse error handling) [cite: uploaded:neuroscript/gonsi/main.go]
[ ] Consider moving to more typed AST? (Design question)
[-] LLM and gonsi can both check scripts for syntax errors (`gonsi` does via ANTLR, LLM pending) [cite: uploaded:neuroscript/pkg/core/parser_api.go]
[ ] Both can compile golang and get errors back
[ ] Both can run go test ./... etc and see errors
[ ] Both can run gofmt and get correctly formatted file back for line no match
[ ] LLMs can supply git-style patches and have them applied to files
[ ] MCP support
[ ] Ability to pass text from LLM to tool in (BASE64) or some other armored format
[ ] Strong list manipulation (cf lisp)
[ ] Prolog style features
[ ] SVG generation and manipulation


## B. Planned Features (Suggested Order Towards Bootstrapping)

[-] Parser: Implement List ([]) and Map ({}) Literal Parsing (Grammar exists, AST/Eval support pending) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go]
[ ] Interpreter: Add internal support for List/Map types
[ ] Interpreter: Implement FOR EACH List Element Iteration
[ ] Syntax & Interpreter: Define and Implement Native List/Map Element Access (e.g., list[index], map["key"])
[ ] Tools: Implement Real In-Memory Vector DB (VectorUpdate, SearchSkills) (Currently mocked)
[ ] Tools: Enhance Git Workflow (Add Branch support, GitPull?, Auto-index after commit)
[ ] Interpreter: Implement Basic Arithmetic Evaluation
[ ] Interpreter: Implement ELSE Block Execution
[ ] Interpreter: Implement Context Management Strategy for CALL LLM
[ ] Interpreter: Define & Implement FOR EACH Map Iteration (Define Keys or Key/Value)
[ ] LLM Gateway: Make LLM endpoint/model configurable [cite: uploaded:neuroscript/pkg/core/llm.go]
[ ] Tools: Add TOOL.ListDirectory(path) **(NEW)** [cite: uploaded:neuroscript/update_project.ns.txt]
[ ] Tools: Add TOOL.LineCount(string_or_filepath) **(NEW)** [cite: uploaded:neuroscript/update_project.ns.txt]
[ ] Tools: Add more utility tools (JSON, HTTP, etc.)
[ ] Tools: Add Markdown tools (Read/Write)
[ ] Tools: Add Structured Document tools
[ ] Tools: Add Table tools
[ ] Tools: Add Integration tools (Sheets, Docs)
[ ] Feature: Add Self-test support in ns
[ ] Feature: Embed standard utility NeuroScripts (e.g., CommitChanges) into gonsi binary (using Go embed) **(NEW)**


## C. Completed Features (Foundation)

[x] Basic Core Syntax Parsing (DEFINE PROCEDURE, COMMENT:, SET, CALL, RETURN, END) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go]
[x] Structured Docstring Parsing (COMMENT: block content parsed into struct) [cite: uploaded:neuroscript/pkg/core/ast_builder.go, uploaded:neuroscript/pkg/core/utils.go]
[x] Block Header Parsing (IF...THEN, WHILE...DO, FOR EACH...DO) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go]
[x] Line Continuation Parsing (Handled implicitly by ANTLR lexer skipping `\`+newline) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_lexer.go]
[x] Basic Expression Evaluation (String Literals, {{Placeholders}}, Variables, __last_call_result, Parentheses) [cite: uploaded:neuroscript/pkg/core/evaluation.go]
[x] String Concatenation (+) [cite: uploaded:neuroscript/pkg/core/evaluation.go]
[x] Basic Condition Evaluation (==, !=, >, <, >=, <=, true/false strings) [cite: uploaded:neuroscript/pkg/core/evaluation.go]
[x] Basic Interpreter Structure (Interpreter, Scope, RunProcedure) [cite: uploaded:neuroscript/pkg/core/interpreter.go]
[x] CALL LLM Integration (via llm.go) [cite: uploaded:neuroscript/pkg/core/llm.go, uploaded:neuroscript/pkg/core/interpreter_steps.go]
[x] CALL TOOL Mechanism [cite: uploaded:neuroscript/pkg/core/tools.go, uploaded:neuroscript/pkg/core/interpreter_steps.go]
[x] Basic Tools Implemented (ReadFile, WriteFile, SanitizeFilename, GitAdd, GitCommit) [cite: uploaded:neuroscript/pkg/core/tools_register.go, uploaded:neuroscript/pkg/core/utils.go]
[x] String Tools Implemented (StringLength, Substring, ToUpper, ToLower, TrimSpace, SplitString, SplitWords, JoinStrings, ReplaceAll, Contains, HasPrefix, HasSuffix) [cite: uploaded:neuroscript/pkg/core/tools_string.go]
[x] Mock Vector DB Tools (VectorUpdate, SearchSkills) [cite: uploaded:neuroscript/pkg/core/tools_register.go, uploaded:neuroscript/pkg/core/embeddings.go]
[x] Basic CLI Runner (gonsi) [cite: uploaded:neuroscript/gonsi/main.go]
[x] Parser Tests Updated for Blocks/Line Continuation (Assumed from context)
[x] Interpreter Block Execution (Basic IF/WHILE/FOR, RETURN propagation) [cite: uploaded:neuroscript/pkg/core/interpreter_steps.go]
[x] Graceful skipping of files with parse errors in gonsi (Confirmed) [cite: uploaded:neuroscript/gonsi/main.go]
[x] Parsing of Comparison Operators (>, <, >=, <=) (Grammar exists, interpreter handles) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go, uploaded:neuroscript/pkg/core/evaluation.go]
[x] Parsing of Numeric Literals (Grammar exists) [cite: uploaded:neuroscript/pkg/core/generated/neuroscript_parser.go]
[x] Debug Flags and Conditional Logging in gonsi (-debug-tokens, -debug-ast, -debug-on-error) **(NEW)** [cite: uploaded:neuroscript/gonsi/main.go]
[x] Interpreter: Implement FOR EACH String Character Iteration **(Moved from Planned)** [cite: uploaded:neuroscript/pkg/core/interpreter_steps.go]
[x] Bootstrap Skills: Create initial .ns.txt skills (HandleSkillRequest, CommitChanges, etc.) **(Moved from Planned)** [cite: uploaded:neuroscript/gonsi/skills/orchestrator.ns.txt, uploaded:neuroscript/gonsi/skills/commit_changes.ns.txt]
[x] Interpreter: Implement More Conditions (>, <, >=, <=) **(Moved from Planned)** [cite: uploaded:neuroscript/pkg/core/evaluation.go]