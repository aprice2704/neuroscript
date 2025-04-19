:: type: NSproject  
:: subtype: documentation  
:: version: 0.1.0  
:: id: concepts-v0.1  
:: status: draft  
:: dependsOn: docs/script spec.md, docs/metadata.md, docs/neurodata_and_composite_file_spec.md, pkg/core/tools_register.go, docs/ns/tools/index.md, docs/ns/tools/move_file.md, docs/ns/tools/query_table.md, docs/ns/tools/go_update_imports_for_moved_package.md  
:: howToUpdate: Review against core project goals and implemented features. Update feature list and links as functionality evolves.  

# Core Concepts & Features of NeuroScript

This document outlines the fundamental principles driving NeuroScript's design and lists its key features.

## Principles

NeuroScript development adheres to these core principles to ensure it effectively facilitates communication between humans, AI, and computers:

1.  **Readability:** All users (human, AI, or computer executing parsing logic) must be able to easily read and understand the intent of NeuroScript files (`.ns.txt`) and NeuroData (`.nd*`) formats. Simple edits should ideally be possible without constantly referring to documentation. NS formats prioritize being self-describing.
    [Diagram Suggestion: Simple icons representing Human, AI, Computer all looking at a readable text document.]

2.  **Executability:** The procedural steps defined in NeuroScript should be clear enough that any participant could, in principle, follow the logic and perform the actions described, whether manually or through automated interpretation. This supports auditing and understanding workflow.

3.  **Clarity:** The primary focus is always on clear communication. Features, syntax, and data structures favor explicitness and obviousness over achieving maximum concision or supporting highly complex, obscure constructs. The "mile wide, inch deep" philosophy applies here.

4.  **Embedded Metadata:** Wherever practical, NeuroScript files should contain standard metadata (`:: key: value`) indicating their version, dependencies, and potentially instructions for maintenance. This promotes better organization and understanding of interrelationships within a project. See the [Metadata Specification](../metadata.md).

## Key Features

NeuroScript achieves its goals through the following key features:

[Diagram Suggestion: High-level block diagram showing 'neurogo' interacting with '.ns.txt Scripts (Skills)', '.nd* NeuroData', 'External TOOLs (Go Code)', and 'LLM API'.]

* **Structured Pseudocode for AI/Human/Computer**: Provides a way to write procedures (`DEFINE PROCEDURE`) that combine simple, imperative steps (like `SET`, `EMIT`, `RETURN`) with loops (`FOR EACH`, `WHILE`), conditions (`IF/THEN/ELSE`), and calls to more complex logic (other procedures, tools, LLMs). See the [Language Specification](../script%20spec.md).
* **Explicit Reasoning Flow**: Moves complex workflows out of ambiguous natural language or hidden model "thoughts" into a reviewable, step-by-step script format.
* **Self-Documenting Procedures**: The mandatory `COMMENT:` block within each procedure requires defining `PURPOSE`, `INPUTS`, `OUTPUT`, and `ALGORITHM`, ensuring that the "skill" captured is understandable and reusable. See the [Language Specification](../script%20spec.md#24-docstrings-comment-block).
* **Tool Integration**: A core concept is extending capabilities via `CALL TOOL.FunctionName(...)`. This integrates external Go functions for specific tasks. Numerous tools are built-in, covering areas like Filesystem ([e.g., TOOL.MoveFile spec](../ns/tools/move_file.md)), Git, String manipulation, Shell execution (use with caution!), Go build/test ([e.g., TOOL.GoUpdateImports spec](../ns/tools/go_update_imports_for_moved_package.md)), Vector DB operations (Mock), Metadata Extraction, NeuroData Checklist Parsing, NeuroData Block Extraction, NeuroData Table Querying ([TOOL.QueryTable spec](../ns/tools/query_table.md)), Math operations, List operations, and more. See the [Tool Specification Index](../ns/tools/index.md) for available detailed specs or [tools_register.go](../../pkg/core/tools_register.go) for the source list.
* **LLM Integration**: `CALL LLM(prompt)` provides a straightforward way to delegate tasks suited to Large Language Models, like text generation, summarization, or complex analysis. See [llm.go](../../pkg/core/llm.go).
* **Rich Data Handling**: Supports basic literals (string, number, boolean) and composite types like lists (`[...]`) and maps (`{...}`) directly in the syntax, including element access (`list[index]`, `map["key"]`). See [Language Specification](../script%20spec.md#23-expressions-literals-and-evaluation).
* **Basic Control Flow**: Standard `IF/THEN/ELSE/ENDBLOCK`, `WHILE/DO/ENDBLOCK`, and `FOR EACH/IN/DO/ENDBLOCK` constructs allow for essential procedural logic. `FOR EACH` supports iteration over lists, map keys, and string characters. See [interpreter_control_flow.go](../../pkg/core/interpreter_control_flow.go).
* **CLI Interpreter (`neurogo`)**: The reference implementation (`neurogo`) is a command-line tool written in Go that parses and executes `.ns.txt` files. It supports loading libraries of procedures, debug flags, and different execution modes. See [Installation & Setup](installation.md) and [neurogo source](../../cmd/neurogo/).
* **Agent Mode (Experimental)**: `neurogo` can operate as a secure backend agent, allowing an external LLM (like Gemini) to request the execution of allowlisted `TOOL.*` functions via its Function Calling API. This enables AI-driven interaction with the local environment under strict security controls. See the [Agent Facilities Design](../llm_agent_facilities.md).
    [Diagram Suggestion: Flowchart illustrating Agent Mode: User -> LLM -> NeuroGo Agent (Security Layer -> Tool Executor) -> Local Env -> NeuroGo Agent -> LLM -> User.]
* **VS Code Extension**: Basic syntax highlighting support for `.ns.txt` files is available to improve the editing experience ([vscode-neuroscript](../../vscode-neuroscript/)).
* **NeuroData Parsing**: Built-in tools facilitate working with specific NeuroData formats like Checklists (`.ndcl`) and extracting data from composite files using fenced blocks. See [NeuroData tools](../../pkg/neurodata/).
