# NeuroScript Development Roadmap

Version: 0.1.0
DependsOn: docs/development checklist.md
HowToUpdate: Review the "Planned Features" in the checklist and synthesize major goals into phases here. Ensure alignment with the overall project goal.

This roadmap outlines the high-level goals for NeuroScript development, driving the priorities in the `development checklist.md`.

## Overall Goal

Achieve **Bootstrapping**: Enable NeuroScript, executed by either the `neurogo` interpreter or an LLM, to manage its own development lifecycle. This includes using NeuroScript procedures (`.ns.txt` files) combined with `CALL LLM` and `CALL TOOL.*` to:
* Find relevant skills (via Vector DB search).
* Generate new NeuroScript code for skills or refactoring.
* Check syntax and format NeuroScript code.
* Manage skills within a Git repository (add, commit, branch).
* Potentially build and test associated Go code (`neurogo` itself or tools).

## Development Phases

### Phase 1: Core Interpreter & Language Foundations (Near-Term Focus)

* **Goal:** Solidify the core NeuroScript language execution capabilities needed for complex, self-managing scripts.
* **Key Targets:**
    * Implement robust LLM context management (`CALL LLM`).
    * Introduce NeuroScript-specific error handling (e.g., TRY/CATCH).
    * Design and implement the `NeuroData` concept for structured data handling within scripts.
    * Refine list/map operations if needed beyond basic access/iteration.

### Phase 2: Tooling for Self-Management & Bootstrapping

* **Goal:** Build and refine the specific `TOOL.*` functions essential for NeuroScript to manage its own ecosystem.
* **Key Targets:**
    * Replace mock Vector DB with a real implementation (`TOOL.VectorUpdate`, `TOOL.SearchSkills`).
    * Enhance Git tooling (`TOOL.GitAdd`, `TOOL.GitCommit`) with branching, status checks, pulling, and auto-indexing capabilities.
    * Create `TOOL.NeuroScriptCheckSyntax` for validating script content programmatically.
    * Develop `nsfmt` (either as a tool or a standard NeuroScript procedure) for code formatting.
    * Finalize versioning conventions (`FILE_VERSION`, `LANG_VERSION`) and potentially automate updates via tooling.
    * Embed core utility NeuroScripts (like `CommitChanges`) into the `neurogo` binary.

### Phase 3: Ecosystem Expansion & General Tooling

* **Goal:** Broaden NeuroScript's applicability by adding more general-purpose tools and integrations.
* **Key Targets:**
    * Add common utility tools (JSON parsing/manipulation, HTTP requests).
    * Implement Markdown reading/writing tools.
    * Develop tools for interacting with structured documents (e.g., hierarchical data).
    * Add table manipulation tools.
    * Explore integration with external services (e.g., Google Sheets, Google Docs).

### Phase 4: Advanced Features & Long-Term Vision

* **Goal:** Explore more advanced language features and capabilities.
* **Key Targets:**
    * Implement self-testing features within NeuroScript.
    * Enhance list manipulation capabilities.
    * Investigate Prolog-style logic programming features.
    * Add support for generating/manipulating other formats (e.g., SVG).
    * Refine LLM interaction (e.g., patch application, armored data passing).
    * Multi-Context Prompting (MCP) support.