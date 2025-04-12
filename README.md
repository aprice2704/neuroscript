# NeuroScript: A Toolkit for AI Communication

## Foundation

**STATUS: EARLY DEVELOPMENT**

Under massive and constant updates, do not use yet.


The NeuroScript project (NS) aims to allow Humans, AIs and computers to communicate in clear, reliable, repeatable ways by providing more structured means than natural language alone.

<p align="center"><img src="docs/sparking_AI_med.jpg" alt="humans uplift machines" width="320" height="200"></p>

NeuroScript includes:

1. A script language (neuroscript) using which humans, AIs and computers may pass each other procedural knowledge, and the means to build cooperative systems

2. A set of data formats (neurodata) for communicating passive data in a clear way with agreed rules for manipulation

3. A client program (neurogo) that can take execute neuroscript, communicate with humans, AIs and computers, and run tools for itself, or its co-workers

## Principles

1. Readability: all users must be able to read, and in principle edit, NS formats of all kinds without having to resort to documentation for simple changes; thus NS formats should be as self-describing as practical  

2. Executability: similarly, eveyone should be able to follow the intent of all scripts so that anyone could, in principle, audit and execute NS formats

3. Clarity: The preeminent focus of all NS files should be clarity over concision or features

4. Embedded Metadata: Wherever practical, ns files should include within them at least their version, what files they depend on, and how to update them when those dependencies change.

## Overview of the parts of NS

### neuroscript

The neuroscript script language (ns) is a structured, human-readable language that provides a *procedural scaffolding* for execution. It is designed to store, discover, and reuse **"skills"** (procedures) with clear docstrings and robust metadata, enabling everyone to build up a library of **reusable, well-documented knowledge**. It is intended to be primarily READ by humans, WRITTEN and EXECUTED by AIs and EXECUTED by computers.

NeuroScript interpreters, such as neurogo, are intended to execute NeuroScript scripts on conventional (von Neumann) computers, but are expected to make heavy use of AI abilities via API. They allow non-AI computers to participate in a network of NS workers.

### neurodata (nd)

NeuroData provides simple, human-readable formats for tracking tasks, requirements, or states (e.g., checklists `.ndcl`) and extracting structured data blocks from documents. They are designed to be easily parsed and manipulated by tools while remaining in plain text.  

## neuroscript in more detail

- #TODO

### Features

- **Structured Pseudocode for AI/Human/Computer**: Write procedures combining mechanical steps (assignments, loops, conditions) and external calls. [cite: uploaded:neuroscript/pkg/core/ast.go]
- **Explicit Reasoning Flow**: Makes AI or complex logic explicit, reviewable, and repeatable.
- **Self-Documenting Procedures**: Mandatory `COMMENT:` block includes purpose, inputs, outputs, algorithm, language version, caveats, and examples. [cite: uploaded:neuroscript/pkg/core/ast.go, uploaded:neuroscript/pkg/core/utils.go]
- **Tool Integration**: `CALL TOOL.FunctionName(...)` integrates external capabilities (Filesystem, Git, String manipulation, Shell commands, Go tooling, Vector DB operations, Metadata Extraction, Checklist Parsing, Block Extraction). [cite: uploaded:neuroscript/pkg/core/tools_register.go]
- **LLM Integration**: `CALL LLM(prompt)` delegates tasks requiring natural language understanding or generation. [cite: uploaded:neuroscript/pkg/core/llm.go, uploaded:neuroscript/pkg/core/interpreter_simple_steps.go]
- **Rich Data Handling**: Supports string, number, boolean literals, plus list (`[]`) and map (`{}`) literals and element access (`list[idx]`, `map["key"]`). [cite: uploaded:neuroscript/pkg/core/ast_builder_collections.go, uploaded:neuroscript/pkg/core/evaluation_access.go]
- **Basic Control Flow**: `IF/THEN/ELSE/ENDBLOCK`, `WHILE/DO/ENDBLOCK`, `FOR EACH/IN/DO/ENDBLOCK` (iterates lists, maps, strings). [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
- **CLI Interpreter (`neurogo`)**: A Go-based interpreter parses and executes `.ns.txt` files, with library loading and debug flags. [cite: uploaded:neuroscript/cmd/neurogo/main.go, uploaded:neuroscript/pkg/neurogo/app.go, uploaded:neuroscript/pkg/neurogo/config.go]
- **Agent Mode (Experimental)**: Allows `neurogo` to act as a secure backend for an LLM, executing allowlisted tools based on LLM requests via Function Calling. [cite: uploaded:neuroscript/pkg/neurogo/app_agent.go, uploaded:neuroscript/pkg/core/security.go]
- **VS Code Extension**: Provides syntax highlighting for `.ns.txt` files. [cite: uploaded:neuroscript/vscode-neuroscript/package.json]
- **NeuroData Parsing**: Tools for parsing checklists (`TOOL.ParseChecklistFromString`) and extracting fenced blocks with metadata (`TOOL.BlocksExtractAll`). [cite: uploaded:neuroscript/pkg/neurodata/checklist/checklist_tool.go, uploaded:neuroscript/pkg/neurodata/blocks/blocks_tool.go]

### Why NeuroScript?

Most AI models rely on hidden chain-of-thought or ad hoc patterns. **NeuroScript** aims to make reasoning **explicit**, **reusable**, and **collaborative**:

1.  **Modular**: Define small, focused procedures (`SummarizeText`, `CommitChanges`), then orchestrate them for complex tasks (`UpdateProjectDocs`).
2.  **Documented**: Standardized docstrings make skills discoverable, reviewable, and maintainable by humans and AIs.
3.  **Hybrid Execution**: Combine precise procedural logic (executable by `neurogo`) with flexible LLM reasoning (`CALL LLM`) and powerful external tools (`CALL TOOL.*`).
4.  **Scaffold for Complex Workflows**: Provides a clear structure for large or critical AI workflows, guiding execution and facilitating debugging.

### Core Concepts

1.  **Procedures**: Defined with `DEFINE PROCEDURE Name(Arguments)`, includes a required `COMMENT:` block with metadata like `PURPOSE`, `INPUTS`, `OUTPUT`, `ALGORITHM`, `LANG_VERSION`. [cite: uploaded:neuroscript/docs/script spec.md] Ends with `END`.
2.  **Statements**:
    - `SET variable = expression`: Assigns the *raw* result of an expression. [cite: uploaded:neuroscript/pkg/core/ast_builder_statements.go]
    - `CALL target(args...)`: Invokes Procedures, `LLM`, or `TOOL.Function`. Result accessible via `LAST`. [cite: uploaded:neuroscript/pkg/core/ast_builder_statements.go]
    - `LAST`: Keyword evaluating to the raw result of the most recent `CALL`. [cite: uploaded:neuroscript/pkg/core/ast.go]
    - `EVAL(string_expression)`: *Explicitly* resolves `{{placeholders}}` within the string result of `string_expression`. Placeholders are *not* resolved automatically elsewhere. [cite: uploaded:neuroscript/pkg/core/evaluation_resolve.go]
    - `RETURN expression`: Exits procedure, returning the raw evaluated expression value (or nil). [cite: uploaded:neuroscript/pkg/core/ast_builder_statements.go]
    - `EMIT expression`: Prints the string representation of the raw evaluated expression value. [cite: uploaded:neuroscript/pkg/core/ast_builder_statements.go]
    - Control Flow: `IF/THEN/ELSE/ENDBLOCK`, `WHILE/DO/ENDBLOCK`, `FOR EACH/IN/DO/ENDBLOCK`. Blocks require `ENDBLOCK`. [cite: uploaded:neuroscript/pkg/core/interpreter_control_flow.go]
3.  **Expressions**: Combine literals, variables, `LAST`, `EVAL()`, arithmetic (`+`, `-`, `*`, `/`, `%`, `**`), comparisons (`==`, `!=`, `>`, `<`, `>=`, `<=`), logical (`AND`, `OR`, `NOT`), bitwise (`&`, `|`, `^`), function calls (`LN`, `LOG`, `SIN`, etc.), and element access (`[]`). [cite: uploaded:neuroscript/pkg/core/ast.go, uploaded:neuroscript/pkg/core/evaluation_logic.go]
4.  **Literals**: Strings (`"..."`, `'...'`), numbers (`123`, `4.5`), booleans (`true`, `false`), lists (`[expr1, expr2]`), maps (`{"key": expr1, "val": expr2}`). [cite: uploaded:neuroscript/pkg/core/ast_builder_collections.go]
5.  **Docstrings**: Ensure procedures are self-documenting via the `COMMENT:` block. [cite: uploaded:neuroscript/docs/script spec.md]
6.  **Skill Library**: Procedures (`.ns.txt` files) are intended to be stored (e.g., in Git) and discoverable (e.g., via vector search on docstrings - mock implemented). [cite: uploaded:neuroscript/pkg/core/tools_vector.go]
7.  **Versioning**: Files should include `:: version:` metadata comment. Procedures can include `LANG_VERSION:` in docstrings. `FILE_VERSION "..."` declaration is also supported. [cite: uploaded:neuroscript/docs/metadata.md, uploaded:neuroscript/docs/script spec.md]
8.  **Available TOOLs**: `ReadFile`, `WriteFile`, `ListDirectory`, `LineCountFile`, `SanitizeFilename`, `GitAdd`, `GitCommit`, `SearchSkills` (mock), `VectorUpdate` (mock), `ExecuteCommand`, `GoBuild`, `GoCheck`, `GoTest`, `GoFmt`, `GoModTidy`, `ExtractMetadata`, `ParseChecklistFromString`, `BlocksExtractAll`, `StringLength`, `Substring`, `ToUpper`, `ToLower`, `TrimSpace`, `SplitString`, `SplitWords`, `JoinStrings`, `ReplaceAll`, `Contains`, `HasPrefix`, `HasSuffix`, `LineCountString`, `TOOL.Add`. [cite: uploaded:neuroscript/pkg/core/tools_register.go]

### Example Usage

Here’s an example demonstrating current syntax features:

```neuroscript
-- Example using list iteration and string concatenation

DEFINE PROCEDURE GenerateReport(items_list, report_title)
COMMENT:
    PURPOSE: Generates a simple report string from a list of items.
    INPUTS:
      - items_list (list): A list of items (e.g., ["Task A", "Task B"]).
      - report_title (string): The title for the report.
    OUTPUT:
      - report_string (string): The generated report.
    LANG_VERSION: 1.1.0
    ALGORITHM:
      1. Initialize report string with title.
      2. Use FOR EACH to loop through items_list.
      3. Access list item using loop variable.
      4. Concatenate item to report string using '+'.
      5. Return final string.
    EXAMPLES:
      GenerateReport(["A", "B"], "Status") => "Report: Status\n- A\n- B\n"
ENDCOMMENT

SET report_string = "Report: " + report_title + "\n"
SET counter = 0 # Example: Not used, but shows SET

FOR EACH item IN items_list DO
    # Simple string concatenation, no EVAL needed here
    SET report_string = report_string + "- " + item + "\n"
ENDBLOCK # End FOR EACH

RETURN report_string

END
```

## Installation & Setup (neurogo CLI)

1.  **Prerequisites**: Go programming language environment (e.g., Go 1.20+). Git command line tool.
2.  **Build `neurogo`**: Navigate to the `neuroscript` directory in your terminal and run:
    ```bash
    go build -o neurogo ./cmd/neurogo
    ```
    This creates the `neurogo` executable in the `neuroscript` directory.
3.  **LLM Connection (Optional)**:
    * Set the `GEMINI_API_KEY` environment variable with your API key if you intend to use `CALL LLM`.
    * The default model is `gemini-1.5-pro-latest`. [cite: uploaded:neuroscript/pkg/core/llm.go] Use `neurogo -agent -model <model_name>` to specify a different one in agent mode.
4.  **Run `neurogo`**:
    ```bash
    # Example: Run the TestListAndMapAccess procedure from the library
    ./neurogo -lib ./library ./library/test_listmap.ns.txt TestListAndMapAccess "MyPrefix"

    # Example: Run with debug logging for the interpreter
    ./neurogo -debug-interpreter -lib ./library ./library/ask_llm.ns.txt AskCapitalCity

    # Example: Start agent mode (requires GEMINI_API_KEY)
    ./neurogo -agent -allowlist ./cmd/neurogo/agent_allowlist.txt -sandbox ./cmd/neurogo/agent_sandbox
    ```
    * **Script Mode Usage**: `./neurogo [flags] <ProcedureToRun | FileToRun.ns.txt> [proc_args...]`
    * **Agent Mode Usage**: `./neurogo -agent [agent_flags...]`
    * **Flags**: `-lib <path>`, `-debug-ast`, `-debug-interpreter`, `-agent`, `-allowlist <file>`, `-denylist <file>`, `-sandbox <dir>`, `-apikey <key>`. [cite: uploaded:neuroscript/pkg/neurogo/config.go]
5.  **(Optional) Database / Skill Registry**:
    * Vector search/update tools (`TOOL.SearchSkills`, `TOOL.VectorUpdate`) are currently mocked in-memory. [cite: uploaded:neuroscript/pkg/core/tools_vector.go] No external DB setup required for the mock.

---

## FAQ

**Q1: Is NeuroScript a full programming language?**
A: It’s more of a *structured pseudocode* or *orchestration language*—focused on providing procedural scaffolding, managing state (`SET`), and coordinating calls to external logic (LLMs, TOOLs, other Procedures). Complex computation is typically delegated.

**Q2: Can I integrate external tools besides LLMs?**
A: Yes—define Go functions and register them using the `ToolRegistry`. [cite: uploaded:neuroscript/pkg/core/tools_registry.go] They become available via `CALL TOOL.YourFunctionName(...)`. Numerous filesystem, string, Git, Go, NeuroData, and Metadata tools are already included. [cite: uploaded:neuroscript/pkg/core/tools_register.go, uploaded:neuroscript/pkg/neurodata/blocks/blocks_tool.go, uploaded:neuroscript/pkg/neurodata/checklist/checklist_tool.go]

**Q3: How do I version-control procedures?**
A: Store `.ns.txt` files in a Git repository. Use `TOOL.GitAdd` and `TOOL.GitCommit` (or external Git commands) to manage changes. Add `:: version:` metadata comments to files and `LANG_VERSION:` in procedure docstrings. [cite: uploaded:neuroscript/docs/metadata.md]

---

## Contributing

We will welcome contributions! But **NOT YET** :P

See the roadmap ["docs/RoadMap.md"](docs/RoadMap.md) and development checklist ["docs/development checklist.md"](docs/development%20checklist.md) for ideas. Key areas include:

* **Interpreter Enhancements**: Error Handling (TRY/CATCH?), NeuroData support (beyond checklist/blocks).
* **Tooling**: Real Vector DB integration, enhanced Git workflow, Syntax Checking (`TOOL.NeuroScriptCheckSyntax`), Formatter (`nsfmt`), JSON/HTTP/Markdown tools.
* **Language Features**: Self-testing support, advanced list/map manipulation.
* **Documentation**: More examples, tutorials, refining specifications.
* **VS Code Extension**: Adding features beyond syntax highlighting (e.g., linting, diagnostics).
* **Agent Mode**: Enhancing security, capabilities, and context management.

Please open an issue or submit a pull request.

---

## License

This project is licensed under the **MIT License**

---

## Authors

Authors:  Andrew Price (www.eggstremestructures.com),
          Gemini 1.5 Pro (gemini.google.com)

:: version: 0.1.6
:: dependsOn: docs/script spec.md, docs/development checklist.md
:: Authors: Andrew Price, Gemini 1.5