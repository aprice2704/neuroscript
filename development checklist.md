# NeuroScript Development Checklist

Goal: Reach the "bootstrapping" point where NeuroScript, executed by an LLM or gonsi, can use CALL LLM and TOOLs to find, create, and manage NeuroScript skills stored in Git and indexed in a vector DB.

Completed Features (Foundation)
[x] Basic Core Syntax Parsing (DEFINE PROCEDURE, COMMENT:, SET, CALL, RETURN, END)
[x] Structured Docstring Parsing (COMMENT: block with sections)
[x] Block Header Parsing (IF...THEN, WHILE...DO, FOR EACH...DO)
[x] Line Continuation Parsing (\)
[x] Basic Expression Evaluation (String Literals, {{Placeholders}}, Variables, __last_call_result)
[x] String Concatenation (+)
[x] Basic Condition Evaluation (==, !=, true/false strings)
[x] Basic Interpreter Structure (Interpreter, Scope, RunProcedure)
[x] CALL LLM Integration (via llm.go)
[x] CALL TOOL Mechanism
[x] Basic Tools Implemented (ReadFile, WriteFile, SanitizeFilename, GitAdd, GitCommit)
[x] Mock Vector DB Tools (VectorUpdate, SearchSkills)
[x] Basic CLI Runner (gonsi)
[x] Parser Tests Updated for Blocks/Line Continuation

Planned Features (Suggested Order Towards Bootstrapping)
[ ] Interpreter: Implement Block Execution (Execute []Step in Value for IF/WHILE/FOR EACH)
[ ] Interpreter: Implement FOR EACH String Character Iteration
[ ] Parser: Implement List ([]) and Map ({}) Literal Parsing
[ ] Interpreter: Add internal support for List/Map types
[ ] Interpreter: Implement FOR EACH List Element Iteration
[ ] Tools & Syntax: Define and Implement List/Map Element Access (e.g., TOOL.ListGet, TOOL.MapGet, TOOL.MapSet, or native list[idx]/map["key"] syntax)
[ ] Tools: Implement Real Vector DB Integration (VectorUpdate, SearchSkills)
[ ] Tools: Enhance Git Workflow (GitPull?, Conflict checks?, Auto-index after commit?)
[ ] Bootstrap Skills: Create initial .ns skills (SearchSkills, GetSkillCode, WriteNewSkill, ImproveSkill?) using implemented features.
[ ] Interpreter: Implement Basic Arithmetic Evaluation
[ ] Interpreter: Implement More Conditions (>, <, >=, <=)
[ ] Interpreter: Implement ELSE Block Execution
[ ] Interpreter: Define & Implement NeuroScript Error Handling (ASSERT / TRY?)
[ ] Interpreter: Implement Context Management Strategy for CALL LLM
[ ] Interpreter: Define & Implement FOR EACH Map Iteration (Keys? Key-Value pairs?)
[ ] Tools: Add more utility tools (String manipulation, JSON, HTTP, etc.)
[ ] LLM Gateway: Make LLM endpoint/model configurable
[ ] Advanced: Concurrency, Reflection (REFLECT?)