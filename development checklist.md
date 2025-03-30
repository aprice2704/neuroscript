# NeuroScript Development Checklist (v3 - Updated based on HandleSkillRequest success)

Goal: Reach the "bootstrapping" point where NeuroScript, executed by an LLM or gonsi, can use CALL LLM and TOOLs to find, create, and manage NeuroScript skills stored in Git and indexed in a vector DB.

## A. Capabilities (Existing & Target)

[x] gonsi able to execute basic ns (SET, CALL, RETURN, basic IF/WHILE/FOR headers and block execution)
[x] ns stored in git (manually, but tools support adding/committing)
[x] Basic set of golang tools in gonsi (ReadFile, WriteFile, SanitizeFilename, GitAdd, GitCommit, mock DB/Search, String tools)
[ ] LLM able to read ns and execute it (via prompt guidance)
[ ] LLM able to translate simple ns into golang tool
[ ] Std lib of foundational ns for LLMs to use (e.g., bootstrapping skills - HandleSkillRequest, CommitChanges)
[ ] Use git branch for version control within tools
[ ] Markdown tools (r & w)
[ ] Structured document tools (hierarchical info/docs)
[ ] Table tools
[ ] Integration tools (e.g., Google Sheets and Docs)
[ ] Self-test support in ns
[x] In-memory vector DB implemented (mocked, VectorUpdate, SearchSkills likely working via HandleSkillRequest)
[x] gonsi skips loading ns files with errors gracefully (needed to load HandleSkillRequest)
[ ] Consider moving to more typed AST? (Design question)
[ ] LLM and gonsi can both check scripts for syntax errors
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

[x] Interpreter: Implement Block Execution (Execute []Step in Value for IF/WHILE/FOR EACH) (Inferred as working)
[ ] Parser: Implement List ([]) and Map ({}) Literal Parsing
[ ] Interpreter: Add internal support for List/Map types
[ ] Interpreter: Implement FOR EACH String Character Iteration
[ ] Interpreter: Implement FOR EACH List Element Iteration
[ ] Syntax & Interpreter: Define and Implement Native List/Map Element Access (e.g., list[index], map["key"])
[ ] Tools: Implement Real In-Memory Vector DB (VectorUpdate, SearchSkills) (Currently mocked)
[ ] Tools: Enhance Git Workflow (Add Branch support, GitPull?, Auto-index after commit)
[x] Bootstrap Skills: Create initial .ns skills (SearchSkills, GetSkillCode, WriteNewSkill, ImproveSkill?, UpdateProjectDocs, CommitChanges) using implemented features. (HandleSkillRequest likely part of this)
[ ] Interpreter: Implement Basic Arithmetic Evaluation
[ ] Interpreter: Implement More Conditions (>, <, >=, <=)
[ ] Interpreter: Implement ELSE Block Execution
[ ] Interpreter: Implement Context Management Strategy for CALL LLM
[ ] Interpreter: Define & Implement FOR EACH Map Iteration (Define Keys or Key/Value)
[ ] LLM Gateway: Make LLM endpoint/model configurable
[ ] Tools: Add TOOL.ListDirectory(path) **(NEW)**
[ ] Tools: Add TOOL.LineCount(string_or_filepath) **(NEW)**
[ ] Tools: Add more utility tools (JSON, HTTP, etc.)
[ ] Tools: Add Markdown tools (Read/Write)
[ ] Tools: Add Structured Document tools
[ ] Tools: Add Table tools
[ ] Tools: Add Integration tools (Sheets, Docs)
[ ] Feature: Add Self-test support in ns
[x] Feature: gonsi skips loading files with parse errors (Moved from Planned to Confirmed)
[ ] Feature: Embed standard utility NeuroScripts (e.g., CommitChanges) into gonsi binary (using Go embed) **(NEW)**


## C. Completed Features (Foundation)

[x] Basic Core Syntax Parsing (DEFINE PROCEDURE, COMMENT:, SET, CALL, RETURN, END)
[x] Structured Docstring Parsing (COMMENT: block with sections)
[x] Block Header Parsing (IF...THEN, WHILE...DO, FOR EACH...DO)
[x] Line Continuation Parsing ()
[x] Basic Expression Evaluation (String Literals, {{Placeholders}}, Variables, __last_call_result, Parentheses)
[x] String Concatenation (+)
[x] Basic Condition Evaluation (==, !=, >, <, >=, <=, true/false strings)
[x] Basic Interpreter Structure (Interpreter, Scope, RunProcedure)
[x] CALL LLM Integration (via llm.go)
[x] CALL TOOL Mechanism
[x] Basic Tools Implemented (ReadFile, WriteFile, SanitizeFilename, GitAdd, GitCommit)
[x] String Tools Implemented (StringLength, Substring, ToUpper, ToLower, TrimSpace, SplitString, SplitWords, JoinStrings, ReplaceAll, Contains, HasPrefix, HasSuffix)
[x] Mock Vector DB Tools (VectorUpdate, SearchSkills)
[x] Basic CLI Runner (gonsi)
[x] Parser Tests Updated for Blocks/Line Continuation
[x] Interpreter Block Execution (Basic IF/WHILE/FOR, RETURN propagation)
[x] Graceful skipping of files with parse errors in gonsi (Confirmed)
[x] Parsing of Comparison Operators (>, <, >=, <=) (Grammar exists)
[x] Parsing of Numeric Literals (Grammar exists)