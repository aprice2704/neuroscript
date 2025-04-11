# NeuroScript: A Toolkit for AI Communication

## Foundation

The NeuroScript project (NS) aims to allow Humans, AIs and computers to communicate in clear, reliable, repeatable ways by providing more structured means than natural language alone.

<p align="center"><img src="docs/sparking_AI_med.jpg" alt="humans uplift machines" width="320" height="200"></p>

NeuroScript includes:

1. A script language (neuroscript) designed for humans, AI and computers to pass each other procedural knowledge that they may execute together

2. A set of data formats (NeuroData) for communicating passive data in a clear way with agreed rules for manipulation

3. A client program (neurogo) that can take execute neuroscript, communicate with humans, AIs and computers, and run tools for itself, or its correspondents

## Principles

1. Readability: all users must be able to read, and in principle change, NS formats of all kinds without having to resort to documentation for simple changes.
2. Executability: similarly, eveyone should be able to follow the intent of all scripts so that anyone could, in principle, audit and execute NS files.
3. Clarity: The preeminent focus of all NS files should be clarity.

**Embedded Metadata**: Whereever practical, ns files should include within them their version, what files they depend on, and how to update them when those dependencies change.

## neuroscript

The neuroscript script language (ns) is a structured, human-readable language that provides a *procedural scaffolding* for execution. It is designed to store, discover, and reuse **"skills"** (procedures) with clear docstrings and robust metadata, enabling everyone to build up a library of **reusable, well-documented knowledge**. It is intended to be primarily READ by humans, WRITTEN and EXECUTED by AIs and EXECUTED by computers.

NeuroScript interpreters, such as neurogo, are intended to execute NeuroScript scripts on conventional (von Neumann) computers, but are expected to make heavy use of AI abilities via API.

neurodata formats are intended as easy ways for humans, AIs and computers to store, share and edit smallish amounts of data to deal with everyday issues. neurodata provides ways to **template** data items as well as specifying how should be rendered and manipulated.


## (remainder of read, needs fixing)

Version: 0.2.0  
DependsOn: docs/neuroscript overview.md  
HowToUpdate: Review dependancies, update appropriately for README sections, preserve current content  

Authors:  Andrew Price (www.eggstremestructures.com),  
          Gemini 2.5 Pro (Experimental) (gemini.google.com)

**STATUS: EARLY DEVELOPMENT**

Under massive and constant updates, do not use yet.


## Table of Contents

- [NeuroScript: A Toolkit for AI Communication](#neuroscript-a-toolkit-for-ai-communication)
  - [Foundation](#foundation)
  - [Principles](#principles)
  - [neuroscript](#neuroscript)
  - [(remainder of read, needs fixing)](#remainder-of-read-needs-fixing)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Why NeuroScript?](#why-neuroscript)
  - [Core Concepts](#core-concepts)
  - [Example Usage](#example-usage)
  - [Installation \& Setup (neurogo CLI)](#installation--setup-neurogo-cli)
  - [FAQ](#faq)
  - [Contributing](#contributing)
  - [License](#license)

---

## Features

- **Structured Pseudocode for AI/Human/Computer**: Write procedures combining mechanical steps (assignments, loops, conditions) and external calls.
- **Explicit Reasoning Flow**: Makes AI or complex logic explicit, reviewable, and repeatable.
- **Self-Documenting Procedures**: Mandatory `COMMENT:` block includes purpose, inputs, outputs, algorithm, language version, caveats, and examples.
- **Tool Integration**: `CALL TOOL.FunctionName(...)` integrates external capabilities (Filesystem, Git, String manipulation, Shell commands, Go tooling, Vector DB operations).
- **LLM Integration**: `CALL LLM(prompt)` delegates tasks requiring natural language understanding or generation.
- **Rich Data Handling**: Supports string, number, boolean literals, plus list (`[]`) and map (`{}`) literals and element access (`list[idx]`, `map["key"]`).
- **Basic Control Flow**: `IF/THEN/ELSE/ENDBLOCK`, `WHILE/DO/ENDBLOCK`, `FOR EACH/IN/DO/ENDBLOCK` (iterates lists, maps, strings).
- **CLI Interpreter (`neurogo`)**: A Go-based interpreter parses and executes `.ns.txt` files [main.go](neurogo/main.go).
- **VS Code Extension**: Provides syntax highlighting for `.ns.txt` files [package.json](vscode-neuroscript/package.json).

---

## Why NeuroScript?

Most AI models rely on hidden chain-of-thought or ad hoc patterns. **NeuroScript** aims to make reasoning **explicit**, **reusable**, and **collaborative**:

1.  **Modular**: Define small, focused procedures (`SummarizeText`, `CommitChanges`), then orchestrate them for complex tasks (`UpdateProjectDocs`).
2.  **Documented**: Standardized docstrings make skills discoverable, reviewable, and maintainable by humans and AIs.
3.  **Hybrid Execution**: Combine precise procedural logic (executable by `neurogo`) with flexible LLM reasoning (`CALL LLM`) and powerful external tools (`CALL TOOL.*`).
4.  **Scaffold for Complex Workflows**: Provides a clear structure for large or critical AI workflows, guiding execution and facilitating debugging.

---

## Core Concepts

1.  **Procedures**: Defined with `DEFINE PROCEDURE Name(Arguments)`, includes a required `COMMENT:` block with metadata like `PURPOSE`, `INPUTS`, `OUTPUT`, `ALGORITHM`, `LANG_VERSION` ["script spec.md"](docs/script%20spec.md). Ends with `END`.
2.  **Statements**:
    - `SET variable = expression`: Assigns the *raw* result of an expression.
    - `CALL target(args...)`: Invokes Procedures, `LLM`, or `TOOL.Function`. Result accessible via `LAST`.
    - `LAST`: Keyword evaluating to the raw result of the most recent `CALL`.
    - `EVAL(string_expression)`: *Explicitly* resolves `{{placeholders}}` within the string result of `string_expression`. Placeholders are *not* resolved automatically elsewhere.
    - `RETURN expression`: Exits procedure, returning the raw evaluated expression value (or nil).
    - `EMIT expression`: Prints the string representation of the raw evaluated expression value.
    - Control Flow: `IF/THEN/ELSE/ENDBLOCK`, `WHILE/DO/ENDBLOCK`, `FOR EACH/IN/DO/ENDBLOCK`. Blocks require `ENDBLOCK`.
3.  **Expressions**: Combine literals, variables, `LAST`, `EVAL()`, arithmetic (`+`, `-`, `*`, `/`, `%`, `**`), comparisons (`==`, `!=`, `>`, `<`, `>=`, `<=`), logical (`AND`, `OR`, `NOT`), bitwise (`&`, `|`, `^`), function calls (`LN`, `LOG`, `SIN`, etc.), and element access (`[]`).
4.  **Literals**: Strings (`"..."`, `'...'`), numbers (`123`, `4.5`), booleans (`true`, `false`), lists (`[expr1, expr2]`), maps (`{"key": expr1, "val": expr2}`).
5.  **Docstrings**: Ensure procedures are self-documenting via the `COMMENT:` block ["script spec.md"](docs/script%20spec.md).
6.  **Skill Library**: Procedures (`.ns.txt` files) are intended to be stored (e.g., in Git) and discoverable (e.g., via vector search on docstrings - mock implemented) [tools_vector.go](pkg/core/tools_vector.go).
7.  **Versioning**: Files should include `Version:` metadata comment. Procedures can include `LANG_VERSION:` in docstrings. `FILE_VERSION "..."` declaration is also supported but may be deprecated [conventions.md](docs/conventions.md).

---

## Example Usage

Here’s an example demonstrating current syntax features:

```neuroscript
-- FILE_VERSION "1.1.0" # Optional older-style declaration

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
SET counter = 0

FOR EACH item IN items_list DO
    # Simple string concatenation, no EVAL needed here
    SET report_string = report_string + "- " + item + "\n"
    # SET counter = counter + 1 # Requires arithmetic support if we tracked count
ENDBLOCK # End FOR EACH

RETURN report_string

END
```

## Installation & Setup (neurogo CLI)

1.  **Prerequisites**: Go programming language environment (e.g., Go 1.20+). Git command line tool.
2.  **Build `neurogo`**: Navigate to the `neuroscript` directory in your terminal and run:
    ```bash
    go build -o neurogo ./neurogo
    ```
    This creates the `neurogo` executable in the `neuroscript` directory.
3.  **LLM Connection (Optional)**:
    * Set the `GEMINI_API_KEY` environment variable with your API key if you intend to use `CALL LLM`.
    * The default model is `gemini-1.5-flash-latest` [llm.go](pkg/core/llm.go). (Future: Make configurable).
4.  **Run `neurogo`**:
    ```bash
    # Example: Run the TestListAndMapAccess procedure in the skills dir
    ./neurogo neurogo/skills TestListAndMapAccess "MyPrefix"

    # Example: Run with debug logging for the interpreter
    ./neurogo -debug-interpreter neurogo/skills AskCapitalCity
    ```
    * Usage: `./neurogo [flags] <skills_directory> <ProcedureToRun> [args...]`
    * Flags: `-debug-ast`, `-debug-interpreter`, `-no-preload-skills` [main.go](neurogo/main.go).

5.  **(Optional) Database / Skill Registry**:
    * Vector search/update tools (`TOOL.SearchSkills`, `TOOL.VectorUpdate`) are currently mocked in-memory [tools_vector.go](pkg/core/tools_vector.go). No external DB setup required for the mock.

---

## FAQ

**Q1: Is NeuroScript a full programming language?**
A: It’s more of a *structured pseudocode* or *orchestration language*—focused on providing procedural scaffolding, managing state (`SET`), and coordinating calls to external logic (LLMs, TOOLs, other Procedures). Complex computation is typically delegated.

**Q2: Can I integrate external tools besides LLMs?**
A: Yes—define Go functions and register them using the `ToolRegistry` [tools.go](pkg/core/tools.go). They become available via `CALL TOOL.YourFunctionName(...)`. Numerous filesystem, string, Git, and Go tools are already included [tools_register.go](pkg/core/tools_register.go).

**Q3: How do I version-control procedures?**
A: Store `.ns.txt` files in a Git repository. Use `TOOL.GitAdd` and `TOOL.GitCommit` (or external Git commands) to manage changes. Add `Version:` metadata comments to files and `LANG_VERSION:` in procedure docstrings [conventions.md](docs/conventions.md).

---

## Contributing

We will welcome contributions! But **NOT YET** :P

See the roadmap [RoadMap.md](docs/RoadMap.md) and development checklist ["development checklist.md"](docs/development%20checklist.md) for ideas. Key areas include:

* **Interpreter Enhancements**: LLM Context Management, Error Handling (TRY/CATCH?), NeuroData support.
* **Tooling**: Real Vector DB integration, enhanced Git workflow, Syntax Checking (`TOOL.NeuroScriptCheckSyntax`), Formatter (`nsfmt`), JSON/HTTP tools.
* **Language Features**: Self-testing support, advanced list/map manipulation.
* **Documentation**: More examples, tutorials, refining specifications.
* **VS Code Extension**: Adding features beyond syntax highlighting (e.g., linting, diagnostics).

Please open an issue or submit a pull request.

---

## License

This project is licensed under the **MIT License**

---