# NeuroScript: A script for humans, AIs and computers

Version: 0.4.0

NeuroScript is a structured, human-readable language that provides a *procedural scaffolding* for Artificial Intelligence (AI) systems. It is designed to store, discover, and reuse **"skills"** (procedures) with clear docstrings and robust metadata, enabling AI systems to build up a library of **reusable, well-documented knowledge**.

## 1. Goals and Principles

1.  **Explicit Reasoning**: Rather than relying on hidden chain-of-thought, NeuroScript encourages step-by-step logic in a code-like format that is both *executable* and *self-documenting*.
2.  **Reusable Skills**: Each procedure is stored and can be retrieved via a standard interface. AIs or humans can then call, refine, or extend these procedures without re-implementing from scratch.
3.  **Self-Documenting**: NeuroScript procedures should include metadata (`:: key: value`) that clarify *purpose*, *parameters*, *return values*, *algorithmic rationale*, *language version*, and *caveats*—mirroring how humans comment code. See `docs/metadata.md`.
4.  **AI Integration**: NeuroScript natively supports interacting with AI models (and potentially other agents) via the `ask` statement.
5.  **Structured Data**: Supports basic list (`[]`) and map (`{}`) literals inspired by JSON/Python for handling structured data.
6.  **Versioning**: Supports file-level version tracking via `:: lang_version:` and `:: file_version:` metadata. See `docs/metadata.md`.

---

## 2. Language Constructs

### 2.1 High-Level Structure

* **File Extension:** Typically `.ns`.
* **File Structure:** A NeuroScript file (`.ns` or similar type) consists of:
    * Optional file header: This includes file-level metadata lines (`:: key: value`) and/or blank lines. **File-level metadata MUST be placed at the START of the file**, before any procedure definitions.
    * Zero or more procedure definitions, separated by optional blank lines.
* **File-Level Metadata:** Use `:: key: value` format (e.g., `:: lang_version: neuroscript@0.5.0`). Key standard keys include `lang_version`, `file_version`, `author`, `description`. See `docs/metadata.md`. File-level metadata is associated with the `Program` node in the AST. *(For Markdown (`.md`) documentation files that might reference NeuroScript or contain metadata, metadata conventions might differ, such as being placed at the end of the file).*

### 2.2 Procedure Definition

Procedures (or "skills") are the core reusable units.

* **Syntax:**
    ```neuroscript
    func ProcedureName signature_part means
      :: procedure_metadata_key: value
      :: description: Describes what this does.
      :: param:input_arg: Describes this parameter.
      :: return:0: Describes the first return value.
      :: algorithm: Describes the steps.

      statement_1
      statement_2
      ...
    endfunc
    ```

* **Keywords:** Starts with `func`, ends with `endfunc`. The `means` keyword separates the signature/metadata from the body.
* **Signature (`signature_part`):** Defines parameters and return value names using optional `needs`, `optional`, and `returns` clauses. Parentheses `()` around the signature part are optional and serve only as visual grouping; if present, they must enclose the entire signature part (from after the procedure name to before `means`).
    * Example without Parens: `needs req1 optional opt1 returns ret1`
    * Example with Parens: `(needs req1 optional opt1 returns ret1)`
    * Example Empty: (no clauses or parentheses)
* **Clauses:**
    * `needs param1, param2`: Required parameters.
    * `optional opt_param1, opt_param2`: Optional parameters (receive `nil` if not provided).
    * `returns ret_name1, ret_name2`: Names for return values (used for documentation via `:: return:<name>:`, but execution returns an ordered list).
* **Metadata:** `:: key: value` lines placed immediately after `means` and before the first statement define procedure-level metadata. Key standard keys include `description`, `purpose`, `param:<name>`, `return:<name>`, `algorithm`, `caveats`. See `docs/metadata.md`. Procedure-level metadata is associated with the `Procedure` node in the AST. Metadata values can span multiple lines using a backslash (`\`) at the end of a line (see Section 2.8).
* **Body (`statement_list`):** A sequence of statements, one per line (blank lines allowed). Statements can be continued onto the next line using a backslash (`\`) (see Section 2.8).

### 2.3 Statements

Statements define the actions within a procedure body.

* **Set Statement:** Assigns the result of an expression to a variable. Variables are dynamically typed and scoped to the procedure execution.
    ```neuroscript
    set my_variable = expression
    set result = tool.fs.ReadFile("input.txt") # Calls are expressions
    set sum = 1 + 2
    # See 'must' statement for enhanced assignment validation
    ```

* **Call Statement:** Explicitly calls a procedure or tool if its return value is not being assigned (e.g., via `set`). This statement *must* use the `call` keyword.
    ```neuroscript
    call tool.fs.LogMessage("Starting phase 1")
    call MyProcedureWithoutReturnValues()
    ```

* **Return Statement:** Exits the current procedure and returns zero or more values (as a list if more than one).
    ```neuroscript
    return # Returns nil implicitly
    return single_value
    return value1, value2, value3 # Returns a list [value1, value2, value3]
    ```

* **Emit Statement:** Outputs a value (typically a string) to the primary output stream or log. Useful for status updates.
    ```neuroscript
    emit "Processing started..."
    emit "Result: " + result_variable
    ```

* **Must Statement:** (**Updated**) Asserts that a condition is true or that an operation succeeds. If the condition is false or the operation returns a standard `error` map, execution halts with an error. This is NeuroScript's primary mechanism for defensive programming and failing loudly.

    1.  **Boolean Assertion:** The original form, asserts an expression is `true`.
        ```neuroscript
        must file_handle != nil
        mustbe is_string(input_name) # Check type using built-in
        ```
    2.  **Mandatory Successful Assignment:** Ensures a tool call or operation succeeds before assigning its result. If the expression on the right returns a standard `error` map, the script halts. This form is used within a `set` statement.
        ```neuroscript
        # Halts if ReadFile returns an error map, otherwise assigns content
        set file_content = must tool.FS.Read("config.json")
        ```
    3.  **Map Key and Type Assertion:** Safely extracts one or more values from a map, ensuring the keys exist and the values have the correct type. The entire statement is atomic; if any check fails, the script halts and no variables are assigned.
        ```neuroscript
        set user_data = {"name": "Gemini", "id": 123, "active": true}

        # Extract a single key, ensuring it exists and is an integer
        set user_id = must user_data["id"] as int

        # Extract multiple keys and validate their types atomically
        set name, is_active = must user_data["name", "active"] as string, bool
        ```

* **Fail Statement:** Intentionally halts execution with an optional error message expression.
    ```neuroscript
    fail # Halts with a generic failure
    fail "Required input file not found: " + filepath
    ```

* **Clear Error Statement:** Clears the current error state within an `on_error` block, allowing execution to continue.
    ```neuroscript
    clear_error
    ```

* **Ask Statement:** Sends a prompt expression to a configured AI agent.
    ```neuroscript
    ask "Summarize this text: " + document_content # Send prompt, discard result
    ask "Generate code for: " + task into generated_code # Send prompt, store result
    ```

* **Break Statement:** Exits the innermost loop (`while` or `for each`) immediately.

* **Continue Statement:** Skips the rest of the current iteration of the innermost loop and proceeds to the next.

### 2.4 Control Flow Blocks

* **If Statement:** `if ... else ... endif`
* **While Statement:** `while ... endwhile`
* **For Each Statement:** `for each ... endfor`
* **On Error Statement:** Defines a block to execute if an error occurs. `on_error means ... endon`

### 2.5 Expressions

NeuroScript supports standard expressions including literals, variables, operators, and function calls.

### 2.6 Comments

Lines starting with `#` or `--` are comments.

### 2.7 Metadata

Lines starting with `:: key: value` define metadata.

### 2.8 Line Continuation

A backslash (`\`) at the end of a line continues a statement, string, or metadata value to the next line.

---

## 3. Tools (`tool.[group.]<Name>`)

External capabilities are exposed via tools, prefixed with `tool.`. See Section 4.3 for error handling conventions.

---

## 4. Core Semantics

### 4.1 Explicit Evaluation Model (`eval`, `last`, Placeholders)

NeuroScript uses an explicit evaluation model for placeholders in strings, triggered by `eval()`.

### 4.2 Scope

Variables are scoped locally to the procedure's execution call stack.

### 4.3 Error Handling (Updated)

NeuroScript's error handling model is designed to be robust and explicit.

* **Standard Error Map:** Handled operational errors from tools (e.g., file not found, invalid input) are not signaled by a Go-level error but are returned as a standard NeuroScript `map` value. This `error` map has a defined structure:
    * `"code"` (required): A standardized error code.
    * `"message"` (required): A human-readable error description.
    * `"details"` (optional): A map or string with additional context.

* **`must` for Assertions:** The `must` keyword is the primary way to handle mandatory operations. As shown in Section 2.3, it can be used to assert boolean conditions or to wrap assignments. When `set my_var = must tool.call()`, the interpreter checks if the tool returned a standard `error` map. If it did, the script halts immediately.

* **`on_error` Block:** The `on_error` block is used to catch unexpected runtime errors (panics), including those generated by a failed `must` assertion. It allows for graceful cleanup or logging before the script terminates or the error is cleared.

    ```neuroscript
    func ProcessFile(needs path) returns status means
      on_error means
        emit "A critical error occurred while processing " + path
        return "Failed"
      endon

      # If ReadFile returns an error map, must triggers a panic,
      # which is then caught by the on_error block above.
      set content = must tool.FS.Read(path)

      emit "File read successfully."
      return "Success"
    endfunc
    ```

* **`fail` for Intentional Errors:** The `fail` statement is used to manually stop execution and raise an error.

### 4.4 New Value Types

NeuroScript v0.4 introduces several new first-class value types.

* **`error`**: The standard error map described in Section 4.3.
* **`timedate`**: Represents a specific point in time, typically wrapping Go's `time.Time`. Created and manipulated via tools like `tool.Time.Now()`.
* **`event`**: Represents a discrete event within the system, containing fields like `name`, `source`, `timestamp`, and `payload`.
* **`fuzzy`**: Represents a value with a degree of truth or confidence. See Section 9 for details.

---

## 9. Fuzzy Logic Support

NeuroScript integrates fuzzy logic to handle uncertainty.

#### §F.1 – Type: `fuzzy`
 - A `fuzzy` value is a real number in the closed interval `[0.0, 1.0]`.
 - It semantically represents degrees of truth, confidence, or similarity.

#### §F.2 – Fuzzy Value Literals & Coercion
 - Readability aliases MAY be defined: `definitely` (1.0), `likely` (0.75), `maybe` (0.5), `unlikely` (0.25), `never` (0.0).
 - `true` and `false` are coerced to `1.0` and `0.0` in fuzzy expressions.

#### §F.3 – Fuzzy Logical Operators
Let `a` and `b` be fuzzy values.
 - `not a` or `!a`: `1 - a`
 - `a and b` or `a & b`: `min(a, b)`
 - `a or b` or `a | b`: `max(a, b)`

#### §F.4 – Usage in Control Flow
 - Standard comparison operators (`>`, `==`, etc.) on fuzzy values return a strict `bool`.
 - Control flow statements (`if`, `while`) require boolean expressions. Fuzzy values must be explicitly compared against a threshold.
    ```neuroscript
    if fuzzy_confidence > 0.8
      emit "High confidence action"
    endif
    ```

:: language: md  
:: lang_version: neuroscript@0.4.0  
:: file_version: 2  
:: type: NSproject  
:: subtype: spec  
:: author: Gemini  
:: created: 2025-06-16  
:: modified: 2025-06-16  
:: dependsOn: ns_script_spec.md (original), must_enhancements.md, new_types.md  
:: howToUpdate: Review source 'dependsOn' documents and integrate any further changes to language features, especially error handling and types.