:: type: NSproject  
:: subtype: documentation  
:: version: 0.1.0  
:: id: architecture-v0.1  
:: status: draft  
:: dependsOn: docs/front/concepts.md, docs/script spec.md, docs/neurodata_and_composite_file_spec.md, docs/llm_agent_facilities.md, pkg/core/interpreter.go, pkg/neurogo/app.go  
:: howToUpdate: Review and update component descriptions and interaction flows as the implementation evolves. Update diagrams as needed.  

# NeuroScript Architecture

This document provides a high-level overview of the main components comprising the NeuroScript ecosystem and how they interact.

## Core Components

The NeuroScript project consists of three primary parts working together:

* **neuroscript (`.ns.txt`):** The scripting language itself. It's a simple, readable, procedural language designed for defining "skills" or procedures. It combines basic control flow (`IF`, `WHILE`, `FOR EACH`), state management (`SET`), and crucially, calls to external logic (`CALL TOOL.*`, `CALL LLM`, `CALL OtherProcedure`). See the [Language Specification](../script%20spec.md).

* **neurodata (`.nd*`):** A suite of simple, plain-text data formats designed for clarity and ease of parsing. Examples include checklists (`.ndcl`), tables (`.ndtable`), graphs (`.ndgraph`), trees (`.ndtree`), map schemas (`.ndmap_schema`), and more. These formats allow structured data exchange between humans, AI, and tools. See the [NeuroData Overview](../neurodata_and_composite_file_spec.md) and specific format specifications in [docs/NeuroData/](../NeuroData/).

* **neurogo:** The reference implementation, written in Go [since you program mostly in golang]. It serves two main roles:
    1.  **Interpreter:** Parses and executes `.ns.txt` script files directly, managing state and calling registered TOOLs or LLMs as instructed by the script.
    2.  **Agent Backend (Experimental):** Acts as a secure execution environment for an external LLM (like Gemini). It receives requests from the LLM (via Function Calling), validates them against security policies (allowlists, sandboxing), executes permitted TOOLs, and returns results to the LLM. See the [Agent Facilities Design](../llm_agent_facilities.md).
    See the [neurogo source code](../../pkg/neurogo/) and [CLI implementation](../../cmd/neurogo/).

## Component Interaction / Workflow

How these components interact depends on the mode `neurogo` is operating in:

### 1. Script Execution Mode

This is the default mode when running `neurogo` with a script file or procedure name.

[Diagram Suggestion: Flowchart for Script Execution Mode: Start -> `neurogo` CLI receives command (script path, args) -> `neurogo` parses `.ns.txt` -> Interpreter loads procedure -> Interpreter executes step -> Evaluation (Expression/Variable?) -> Condition (IF/WHILE/FOR?) -> Action (SET/EMIT/RETURN?) -> Tool Call (-> Go Func) -> LLM Call (-> API) -> Update Interpreter State -> Loop to next step or End.]

* **Initiation:** User executes `neurogo` via the command line, providing the path to a `.ns.txt` file (or a specific procedure within a library) and any necessary arguments or flags (like `-lib`).
* **Parsing:** `neurogo` reads the specified file(s) and uses the NeuroScript parser (built with ANTLR, see [parser_api.go](../../pkg/core/parser_api.go)) to create an Abstract Syntax Tree (AST) representing the procedures.
* **Interpretation:** The `Interpreter` ([interpreter.go](../../pkg/core/interpreter.go)) walks the AST of the target procedure.
* **Execution:** For each statement:
    * `SET`: Evaluates the expression and updates the interpreter's current scope.
    * `IF`/`WHILE`/`FOR EACH`: Evaluates conditions/collections and controls the execution flow.
    * `CALL ProcedureName`: Pushes a new scope and starts interpreting the called procedure.
    * `CALL TOOL.FunctionName`: Looks up the tool in the `ToolRegistry` ([tools_register.go](../../pkg/core/tools_register.go)), validates arguments, executes the corresponding Go function, and stores the result in `LAST`.
    * `CALL LLM`: Formats the prompt, sends it to the configured LLM API ([llm.go](../../pkg/core/llm.go)), and stores the response text in `LAST`.
    * `RETURN`: Evaluates the optional expression and passes the result back to the caller (or exits the script).
    * `EMIT`: Evaluates the expression and prints its string form to standard output.
* **Output:** The primary output is typically generated via `EMIT` statements or the final `RETURN` value of the main procedure.

### 2. Agent Mode

This mode is activated using the `-agent` flag and related security flags (`-allowlist`, `-denylist`, `-sandbox`).

[Diagram Suggestion: Flowchart illustrating Agent Mode: User Box -> Arrow to -> LLM Service Box -> Arrow (Function Call Request) -> NeuroGo Agent Box [contains inner boxes: Security Layer (Allowlist/Validate) -> Tool Executor] -> Arrow (TOOL Execution) -> Local Environment Box -> Arrow (Result) -> NeuroGo Agent Box -> Arrow (Function Response) -> LLM Service Box -> Arrow (Final Answer) -> User Box.]

* **Initiation:** User interacts with an external application (e.g., a chat interface) which communicates with an LLM (like Gemini). `neurogo -agent` runs as a background process or service.
* **LLM Planning:** The LLM receives the user prompt and its list of available "functions" (which correspond to the NeuroGo TOOLs declared via configuration). If it decides a tool is needed, it generates a Function Call request.
* **Request Reception:** The `neurogo` agent ([app_agent.go](../../pkg/neurogo/app_agent.go)) receives the structured Function Call request from the LLM API.
* **Security Validation:** The request (tool name, arguments) is passed to the `SecurityLayer` ([security.go](../../pkg/core/security.go)).
    * Checks if the tool is on the allowlist and not on the denylist.
    * Validates and potentially sanitizes arguments. Crucially, file paths are validated against the sandbox root using `SecureFilePath` logic.
    * Rejects the request if any check fails.
* **Tool Execution:** If validation passes, the agent calls the appropriate Go function for the requested TOOL via the `ToolRegistry`, passing the validated arguments. The tool executes within the security context (e.g., filesystem access confined to the sandbox).
* **Response Formatting:** The result (or error) from the tool execution is formatted into a Function Response structure expected by the LLM API.
* **LLM Continuation:** The agent sends the Function Response back to the LLM API. The LLM uses this result to continue its reasoning and either generate another Function Call or formulate a final text response for the user.
* **Output:** The final LLM text response is relayed back to the user via the initial application.

## Diagrams

[Diagram Suggestion: High-level block diagram showing the 'neurogo' executable as the central component. It reads/writes '.ns.txt Files (Skills)' and '.nd* Files (Data)'. It contains an 'Interpreter/Agent Core' which uses 'Built-in TOOLs (Go Code)' and interacts via API with an external 'LLM Service'.]

*(Placeholder for Overall Architecture Diagram)*

[Diagram Suggestion: Flowchart for Script Execution Mode: Start -> `neurogo` CLI receives command (script path, args) -> `neurogo` parses `.ns.txt` -> Interpreter loads procedure -> Interpreter executes step -> Evaluation (Expression/Variable?) -> Condition (IF/WHILE/FOR?) -> Action (SET/EMIT/RETURN?) -> Tool Call (-> Go Func) -> LLM Call (-> API) -> Update Interpreter State -> Loop to next step or End.]

*(Placeholder for Script Execution Flow Diagram)*

[Diagram Suggestion: Flowchart illustrating Agent Mode: User Box -> Arrow to -> LLM Service Box -> Arrow (Function Call Request) -> NeuroGo Agent Box [contains inner boxes: Security Layer (Allowlist/Validate) -> Tool Executor] -> Arrow (TOOL Execution) -> Local Environment Box -> Arrow (Result) -> NeuroGo Agent Box -> Arrow (Function Response) -> LLM Service Box -> Arrow (Final Answer) -> User Box.]

*(Placeholder for Agent Mode Flow Diagram)*
