:: type: NSproject  
:: subtype: documentation  
:: version: 0.1.0  
:: id: faq-v0.1  
:: status: draft  
:: dependsOn: docs/front/concepts.md, docs/script spec.md, docs/neurodata_and_composite_file_spec.md, docs/metadata.md, pkg/core/tools_register.go  
:: howToUpdate: Add new questions as they arise, update answers based on project changes, ensure links remain valid.  

# Frequently Asked Questions (FAQ)

## General / What is NeuroScript?

**Q: What problem does NeuroScript aim to solve?** A: NeuroScript tackles the communication friction inherent in complex systems where humans, AI agents (like LLMs), and traditional computer programs need to collaborate. It aims to provide clearer, more reliable, and repeatable ways to exchange procedural knowledge ("skills") and structured data than using natural language or complex code interfaces alone. See [Why NeuroScript?](why-ns.md) for more motivation.

**Q: Who is NeuroScript for?** A: It's designed for developers building hybrid systems, AI engineers creating agentic workflows, technical teams needing clear process documentation, and potentially anyone looking for a structured way to define and share procedures that can be understood and executed by different types of actors (human, AI, computer).

**Q: What are the core principles?** A: Readability, Executability, Clarity, and Embedded Metadata. The goal is formats that are self-describing, auditable, and prioritize clarity over concision. See [Principles](concepts.md#principles).

**Q: Is NeuroScript production-ready?** A: **No.** As stated clearly in the main [README.md](../../README.md), NeuroScript is in **EARLY DEVELOPMENT** and undergoing massive, constant updates. It should not be used in production environments at this stage.

## NeuroScript Language (`.ns.txt`)

**Q: Is NeuroScript a full programming language?** A: It’s more of a *structured pseudocode* or *orchestration language*. It's focused on providing procedural scaffolding, managing state (`SET`), and coordinating calls to external logic (LLMs via `CALL LLM`, external tools via `CALL TOOL.*`, other NeuroScript Procedures via `CALL ProcedureName`). Complex computation is typically delegated to tools or LLMs. See the [Language Specification](../script%20spec.md).

**Q: What's the `COMMENT:` block for? Why is it mandatory?** A: The `COMMENT:` block serves as a structured docstring for each `DEFINE PROCEDURE`. It's mandatory to enforce the principle of self-documenting skills. It includes standardized sections like `PURPOSE`, `INPUTS`, `OUTPUT`, `ALGORITHM`, `LANG_VERSION`, `CAVEATS`, and `EXAMPLES`, making procedures understandable and discoverable by both humans and AI. See the [Language Specification](../script%20spec.md#24-docstrings-comment-block).

**Q: How does variable substitution work? What's `EVAL()` for?** A: NeuroScript uses explicit evaluation for placeholders. Standard expressions (like in `SET variable = "Hello " + name` or `EMIT message`) evaluate variables/literals directly to their raw values. Placeholders like `{{variable}}` or `{{LAST}}` are *only* substituted when processed by the `EVAL(string_expression)` function. `EVAL` first evaluates its argument to get a string, then scans that string for placeholders and replaces them with current variable values. See [Core Concepts in concepts.md](concepts.md#core-concepts) and the [Language Specification](../script%20spec.md#23-expressions-literals-and-evaluation).

**Q: What does `LAST` do?** A: The `LAST` keyword evaluates to the raw value returned by the most recently executed `CALL` statement (whether calling another procedure, `LLM`, or a `TOOL.*`). See the [Language Specification](../script%20spec.md#23-expressions-literals-and-evaluation).

**Q: How are NeuroScript procedures/skills stored and found?** A: Procedures are defined in `.ns.txt` files. These files are intended to be stored in a library structure (e.g., a directory specified via the `-lib` flag in `neurogo`, potentially managed by Git). Discovery is planned via tools like `TOOL.SearchSkills` (currently mocked) which would likely use vector embeddings generated from the procedure docstrings.

**Q: How does versioning work?** A: There are two main levels:
    * **File Content Version:** Use `:: version: <semver>` metadata at the top of any file (`.ns.txt`, `.nd*`, `.md`, etc.) to track changes to that specific file's content. See [Metadata Specification](../metadata.md). The older `FILE_VERSION "..."` directive in `.ns.txt` is supported but deprecated in favor of `:: version:`.
    * **Language Compatibility:** Use `LANG_VERSION: <semver>` inside a procedure's `COMMENT:` block to indicate which version of the NeuroScript language specification the procedure targets. See the [Language Specification](../script%20spec.md#25-versioning-conventions-new-section).

## NeuroData Formats (`.nd*`)

**Q: What is NeuroData?** A: NeuroData is a collection of simple, plain-text, human-readable formats designed for representing structured data like checklists (`.ndcl`), tables (`.ndtable`), graphs (`.ndgraph`), trees (`.ndtree`), schemas (`.ndmap_schema`), forms (`.ndform`), etc., within the NeuroScript ecosystem. See the [NeuroData Overview](../neurodata_and_composite_file_spec.md).

**Q: What is a "composite file"?** A: A file (typically Markdown `.md`) that contains multiple fenced code/data blocks, potentially including NeuroScript code, NeuroData formats, or other language snippets, interspersed with explanatory text. NeuroScript provides tools (`TOOL.BlocksExtractAll`) for parsing these files and extracting the structured blocks based on their fence tags and `:: id:` metadata. See the [NeuroData Overview](../neurodata_and_composite_file_spec.md).

**Q: How do references like `[ref:...]` work?** A: They provide a standard way to link to other resources (files or specific blocks within files) within the project. The syntax is `[ref:<location>]` for files or `[ref:<location>#<block_id>]` for blocks, where `<location>` is usually `this` or a relative path (using `/`). Tools resolving these references must use security mechanisms like `SecureFilePath`. See the [References Specification](../NeuroData/references.md).

## `neurogo` Interpreter/Agent

**Q: How do I run a NeuroScript procedure?** A: Use the `neurogo` command-line tool. Specify the library path (`-lib`), the file containing the procedure, and the procedure name, followed by any arguments. Example: `./neurogo -lib ./library ./library/ask_llm.ns.txt AskCapitalCity "Canada"`. See [Installation & Setup](installation.md).

**Q: What are the debug flags?** A: `-debug-ast` prints the Abstract Syntax Tree after parsing. `-debug-interpreter` provides step-by-step logging of the interpreter's execution flow. See [Installation & Setup](installation.md).

**Q: What is Agent Mode (`-agent`)?** A: An experimental mode where `neurogo` acts as a secure backend for an LLM (like Gemini). Instead of executing a script directly, it listens for function call requests from the LLM, validates them against allow/deny lists and security rules, executes permitted `TOOL.*` functions within a sandbox, and returns the results to the LLM. See the [Agent Facilities Design](../llm_agent_facilities.md) and [Installation & Setup](installation.md).

## Tools (`TOOL.*`)

**Q: Can I integrate external tools besides LLMs?** A: Yes—this is a core feature. You can define Go functions and register them using the `ToolRegistry` (`pkg/core/tools_register.go`). They become available via `CALL TOOL.YourFunctionName(...)`. Numerous filesystem, string, Git, Go, NeuroData, and Metadata tools are already included.

**Q: Is `TOOL.ExecuteCommand` safe?** A: **Potentially dangerous.** Executing arbitrary shell commands carries inherent security risks. While `neurogo` provides some basic safeguards (like attempting path validation if arguments look like paths), it's highly recommended to **disable** `TOOL.ExecuteCommand` completely when running in Agent Mode or executing untrusted scripts, using the `-denylist` flag or similar security configurations. See the [Agent Facilities Design](../llm_agent_facilities.md#section-4-critical-security-design).

**Q: What is `SecureFilePath`?** A: It's a security mechanism used internally by filesystem-related tools (`TOOL.ReadFile`, `TOOL.WriteFile`, `TOOL.ListDirectory`, etc.). When `neurogo` is run with a sandbox directory (`-sandbox`), `SecureFilePath` ensures that any file paths manipulated by tools resolve safely *within* that designated directory, preventing access to files outside the sandbox (e.g., via `../` traversal). This is crucial for agent security. See `pkg/core/security.go` and related tool implementations.

## Contributing / Future

**Q: How do I version-control procedures?** A: Store `.ns.txt` files (and related `.nd*` files) in a Git repository. Use `TOOL.GitAdd` and `TOOL.GitCommit` (or external Git commands) to manage changes. Use `:: version:` metadata in files and `LANG_VERSION:` in procedure docstrings to track content versions.

**Q: How can I contribute?** A: Contributions are planned but the project is currently in very early, rapid development ("NOT YET :P"). When open, contributions will likely involve adding tools, NeuroData formats, enhancing the interpreter, improving documentation, or adding tests. See [Contributing](contributing.md), the [Roadmap](../RoadMap.md), and the [Development Checklist](../development%20checklist.md) for potential areas.
