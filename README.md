# NeuroScript

**NeuroScript** is a lightweight pseudocode framework for **explicit AI reasoning**. 
It allows you to define procedures (“skills”) that combine traditional step-by-step logic 
with on-demand calls to large language models (LLMs). Think of it as a **structured thought system** 
for AI: each procedure has a clear docstring, uses human-readable syntax, and can be executed 
(or improved) by humans and LLMs alike.

---

## Table of Contents
1. [Features](#features)
2. [Why NeuroScript?](#why-neuroscript)
3. [Core Concepts](#core-concepts)
4. [Example Usage](#example-usage)
5. [Installation & Setup](#installation--setup)
6. [FAQ](#faq)
7. [Contributing](#contributing)
8. [License](#license)

---

## Features

- **Pseudocode for AI Reasoning**  
  Write procedures that combine mechanical steps and open-ended LLM calls.

- **Self-Documenting Procedures**  
  Each skill has a detailed docstring: purpose, inputs, outputs, caveats, and examples.

- **Easy Integration**  
  Plug in your own LLM backend (e.g., OpenAI API or a local model) and external tools 
  (like system commands, Git operations, etc.).

- **Composable Skill Library**  
  Store, discover, and retrieve procedures from a registry (database or version-controlled repo) 
  to reuse across tasks.

- **Lightweight & Extensible**  
  NeuroScript’s syntax is minimal, letting you extend it with custom statements and domain-specific features.

---

## Why NeuroScript?

Most AI models rely on hidden chain-of-thought or ad hoc patterns. 
**NeuroScript** makes reasoning **explicit** and **reusable**:

1. **Modular**: You can define small, focused procedures (e.g. `SummarizeText`, `AnalyzeData`), 
   then orchestrate them for more complex tasks (`InvestigateMurderScene`).

2. **Documented**: Each procedure includes a docstring that states its purpose, algorithm, 
   inputs, outputs, and examples. That makes it simple for humans and AIs to discover, review, 
   and maintain.

3. **LLM-Ready**: For tasks requiring subjective interpretation, NeuroScript delegates to an 
   LLM via a statement like `CALL LLM("prompt")`, enabling advanced text analysis or generation 
   within a structured workflow.

4. **Scaffold for Complex Projects**: 
   Provides a skeleton for large or critical AI workflows, guiding them via a clear procedural 
   flow with optional reflection or error-handling steps.

---

## Core Concepts

1. **Procedures**  
   Defined with `DEFINE PROCEDURE <Name>(Arguments)`, plus a required docstring block (`COMMENT: ... END`) 
   that outlines **PURPOSE**, **INPUTS**, **OUTPUT**, **ALGORITHM**, **CAVEATS**, and **EXAMPLES**.

2. **Statements**  
   - **SET**: Assign a value to a variable (`SET x = 10`).  
   - **CALL**: Invoke another procedure or an LLM/tool (`CALL MyProcedure(...)` or `CALL LLM("...")`).  
   - **CONTROL FLOW**: `IF/THEN/END`, `FOR/WHILE` loops, etc.  
   - **RETURN**: Return a value from a procedure.

3. **Docstrings**  
   Ensure each procedure is self-documenting. This is crucial for discoverability and future improvements.

4. **Registry / Skill Library**  
   NeuroScript procedures are stored in a database or version-controlled repo, 
   with a standard mechanism to *search* docstrings or code (e.g. vector-based semantic search).

5. **LLM Integration**  
   Use statements like `CALL LLM("prompt")` to delegate open-ended reasoning, summarization, or 
   pattern recognition tasks to a large language model.

---

## Example Usage

Here’s a toy example showing a procedure that calls an LLM to summarize text, 
then integrates a standard “analysis” sub-procedure.

```neuroscript
DEFINE PROCEDURE SummarizeAndAnalyze(InputText)
COMMENT:
    PURPOSE: Summarize the given text via an LLM, then perform a basic analysis step.
    INPUTS: 
      - InputText (string)
    OUTPUT: 
      - resultSummary (string) => The summarized and analyzed result
    ALGORITHM:
      1. Call an LLM to generate a short summary of `InputText`.
      2. Pass that summary to `AnalyzeSummary` (a local procedure).
      3. Combine the analysis and summary into a final string or object.
    EXAMPLES:
      SummarizeAndAnalyze("Some long text...") => "Analysis: ... Summary: ..."
END

SET summary = CALL LLM("Please provide a concise summary of the following text: {{InputText}}")

# Now call a sub-procedure to analyze the summary
SET analysis = CALL AnalyzeSummary(summary)

SET resultSummary = "Analysis: " + analysis + "\nSummary: " + summary
RETURN resultSummary
```

---

## Installation & Setup

1. **Clone this Repo**  
   ```bash
   git clone https://github.com/YourOrg/NeuroScript.git
   cd NeuroScript
   ```

2. **Interpreter / Runtime** (Prototype)  
   - (Optional) We provide a minimal interpreter in Go/Python that can parse and run .ns.txtfiles.  
   - Or integrate NeuroScript statements as a layer on top of your existing AI agent framework.

3. **LLM Connection**  
   - Configure your environment with an API key or local LLM endpoint.  
   - In your config file, specify how `CALL LLM("prompt")` is routed to your model.

4. **(Optional) Database / Skill Registry**  
   - Set up a Postgres or vector DB (e.g., Pinecone, Qdrant, Weaviate) for storing procedures.  
   - The included scripts can help index docstrings and retrieve them by semantic or keyword search.

---

## FAQ

**Q1: Is NeuroScript a full programming language?**  
A: It’s more of a *lightweight pseudocode*—focused on orchestrating AI steps and tool calls. 
Complex data structures or concurrency can be added as extensions.

**Q2: Can I integrate external tools besides LLMs?**  
A: Yes—just define statements like `CALL TOOL.SomeFunction(...)`, linking them to your own 
commands or APIs. 

**Q3: How do I version-control procedures?**  
A: Store them in a Git repo or a versioned DB. NeuroScript can define meta-procedures to 
pull, push, and manage branches automatically.

---

## Contributing

We welcome contributions in the form of:
- **Core Language Enhancements**: Additional statements, error-handling mechanisms, concurrency models.
- **Interpreter Implementations**: Runtimes for other languages (Python, Rust, etc.).
- **Documentation**: More examples, tutorials, or best-practice guides.
- **Tooling**: Scripts or libraries that facilitate indexing, searching, or introspecting NeuroScript code.

Please open an issue or submit a pull request if you have ideas or patches.

---

## License

This project is licensed under the **MIT License** – see the [LICENSE](LICENSE) file for details.

---

### Tips

- Feel free to customize the **Installation** section based on your actual interpreter or environment.  
- If you have *specific instructions* for running your prototype interpreter or hooking it into a particular LLM, add that detail in **Installation** or a separate **docs/** folder.  

This layout should give prospective users/developers a solid overview of **NeuroScript**—what it does, why it exists, and how to start experimenting with it.
