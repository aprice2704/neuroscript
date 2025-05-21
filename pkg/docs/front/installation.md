:: type: NSproject  
:: subtype: documentation  
:: version: 0.1.0  
:: id: installation-v0.1  
:: status: draft  
:: dependsOn: cmd/neurogo/main.go, pkg/neurogo/config.go, docs/build.md  
:: howToUpdate: Update Go version, build steps, or CLI flags/examples as the project evolves.  

# Installation & Setup (`neurogo` CLI)

**STATUS: EARLY DEVELOPMENT**

Under massive and constant updates, do not use yet. This guide provides instructions for building and running the `neurogo` command-line tool from source.

## 1. Prerequisites

* **Go Environment:** You need a working Go installation. Version 1.21 or later is recommended.
* **Git:** The Git command-line tool is required for version control interaction (used by some `TOOL.Git*` functions) and potentially for fetching dependencies.
* **(Optional) Java & ANTLR:** If you need to *regenerate* the NeuroScript parser code from the `.g4` grammar file (`pkg/core/NeuroScript.g4`), you will need Java and the ANTLR tool itself. However, the generated Go parser files are included in the repository, so **ANTLR is NOT required just to build and run `neurogo`**. See [docs/build.md](../build.md) for parser generation details.

## 2. Building `neurogo`

1.  **Clone/Download Source:** Obtain the `neuroscript` project source code (e.g., via `git clone`).
2.  **Navigate to Root:** Open your terminal and change directory to the project's root folder (the one containing `go.mod`).
3.  **Build:** Run the standard Go build command:
    ```bash
    go build -o neurogo ./cmd/neurogo
    ```
4.  **Result:** This command compiles the code and creates the `neurogo` executable file in the current directory (the project root). Go will automatically handle downloading necessary dependencies defined in `go.mod`.

## 3. Configuration

### LLM Connection (Optional)

* If you plan to use NeuroScript features that interact with Large Language Models (`CALL LLM`), you need to provide API credentials.
* **Primary Method (Recommended):** Set the `GEMINI_API_KEY` environment variable:
    ```bash
    export GEMINI_API_KEY="YOUR_API_KEY_HERE"
    ```
* **Alternative Method:** Use the `-apikey` command-line flag when running `neurogo`:
    ```bash
    ./neurogo -apikey "YOUR_API_KEY_HERE" ...
    ```
* **Default Model:** The system currently defaults to using Google's `gemini-1.5-pro-latest` model.
* **Changing Model (Agent Mode):** When running in Agent mode, you can specify a different Gemini model using the `-model` flag:
    ```bash
    ./neurogo -agent -model models/gemini-1.5-flash-latest ...
    ```

## 4. Running `neurogo` (Script Execution Mode)

This is the primary mode for executing `.ns.txt` scripts directly.

* **Basic Syntax:**
    ```bash
    ./neurogo [flags] <Target> [ProcedureArguments...]
    ```
* **`<Target>`:** Can be either:
    * The path to a `.ns.txt` file (e.g., `./library/examples/example.ns.txt`). `neurogo` will execute the *first* procedure defined in that file.
    * The name of a specific procedure (e.g., `TestListAndMapAccess`). `neurogo` will search for this procedure in files within the library path(s).
* **`[ProcedureArguments...]`:** Any arguments to pass to the NeuroScript procedure being run, provided as separate strings.
* **Common Flags:**
    * `-lib <path>`: Specifies a directory containing `.ns.txt` files to be treated as a library. Can be used multiple times. Procedures in library files can be called by `<Target>` name. Example: `-lib ./library`.
    * `-debug-ast`: Prints the Abstract Syntax Tree after parsing the script.
    * `-debug-interpreter`: Enables verbose logging of the interpreter's execution steps.
* **Examples:**
    ```bash
    # Run the first procedure in examples/example.ns.txt, using ./library for CALLs
    ./neurogo -lib ./library ./library/examples/example.ns.txt

    # Run the specific procedure 'AskCapitalCity' found in the library, with debug output
    ./neurogo -debug-interpreter -lib ./library AskCapitalCity "France"

    # Run a procedure from a specific file, passing arguments
    ./neurogo -lib ./library ./library/test_listmap.ns.txt TestListAndMapAccess "MyPrefix" "Arg2Value"
    ```

## 5. Running `neurogo` (Agent Mode - Experimental)

This mode allows `neurogo` to act as a secure backend for an LLM.

* **Basic Syntax:**
    ```bash
    ./neurogo -agent [security_flags...] [other_flags...]
    ```
* **Required Security Flags:**
    * `-agent`: Enables agent mode.
    * `-allowlist <file>`: Path to a text file listing `TOOL.FunctionName`s the LLM is *allowed* to call (one per line).
    * `-sandbox <dir>`: Path to a directory that acts as the root for all filesystem operations requested by the LLM via tools. **Crucial for security.**
* **Optional Security Flags:**
    * `-denylist <file>`: Path to a text file listing `TOOL.FunctionName`s the LLM is explicitly *forbidden* from calling (overrides allowlist). Recommended for disabling dangerous tools like `TOOL.ExecuteCommand`.
* **Other Agent Flags:**
    * `-model <model_name>`: Specify which Gemini model to use (e.g., `models/gemini-1.5-flash-latest`).
    * `-apikey <key>`: Provide API key via flag (alternative to environment variable).
* **Example:**
    ```bash
    # Start agent mode, requiring GEMINI_API_KEY env var
    # Allow tools listed in agent_allowlist.txt
    # Confine file operations to ./agent_sandbox directory
    ./neurogo -agent \
        -allowlist ./cmd/neurogo/agent_allowlist.txt \
        -sandbox ./cmd/neurogo/agent_sandbox \
        -denylist ./cmd/neurogo/agent_denylist.txt # Optional: Explicitly deny certain tools
    ```

## 6. Optional Setup (Vector Database)

* The tools related to finding NeuroScript skills (`TOOL.SearchSkills`, `TOOL.VectorUpdate`) currently use a simple **in-memory mock implementation**.
* No external vector database setup is required to use the current version of `neurogo`. A real implementation is planned for the future.

---
