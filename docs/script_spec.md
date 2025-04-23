# NeuroScript: A script for humans, AIs and computers

Version: 0.1.0

NeuroScript is a structured, human-readable language that provides a *procedural scaffolding* for large language models (LLMs). It is designed to store, discover, and reuse **"skills"** (procedures) with clear docstrings and robust metadata, enabling AI systems to build up a library of **reusable, well-documented knowledge**.

## 1. Goals and Principles

1.  **Explicit Reasoning**: Rather than relying on hidden chain-of-thought, NeuroScript encourages step-by-step logic in a code-like format that is both *executable* and *self-documenting*.
2.  **Reusable Skills**: Each procedure is stored and can be retrieved via a standard interface. LLMs or humans can then call, refine, or extend these procedures without re-implementing from scratch.
3.  **Self-Documenting**: NeuroScript procedures should include docstrings that clarify *purpose*, *inputs*, *outputs*, *algorithmic rationale*, *language version*, and *caveats*—mirroring how humans comment code.
4.  **LLM Integration**: NeuroScript natively supports calling LLMs for tasks that benefit from free-form generation, pattern matching, or advanced “human-like” reasoning.
5.  **Structured Data**: Supports basic list (`[]`) and map (`{}`) literals inspired by JSON/Python for handling structured data.
6.  **Versioning**: Supports file-level and procedure-level version tracking.

---

## 2. Language Constructs

### 2.1 High-Level Structure

A NeuroScript file (`.ns.txt`) typically contains:

1.  Optional **`FILE_VERSION`** declaration (See 2.5)
2.  Zero or more **`DEFINE PROCEDURE`** blocks.
3.  Comments (`#` or `--`) and blank lines can appear between elements.

Each procedure definition follows this structure:

1.  **`DEFINE PROCEDURE`** *Name*( *Arguments* ) `NEWLINE`
2.  Optional **`COMMENT:`** block (Docstring) ending with `ENDCOMMENT` (lexed as `COMMENT_BLOCK`)
3.  **Statements** (the pseudocode body, including nested blocks)
4.  **`END`** `NEWLINE`? to close the procedure definition

**Example (Illustrating Current Syntax):**

```neuroscript
# Optional File Version Declaration
FILE_VERSION "1.0.0"

# Example Procedure
DEFINE PROCEDURE ProcessChecklist(checklist_items)
COMMENT:
    PURPOSE: Processes a checklist, potentially logging items.
    INPUTS: - checklist_items: A list of map objects, where each map has "task" (string) and "status" (string). Example: `[{"task":"Review", "status":"done"}, {"task":"Implement", "status":"pending"}]`
    OUTPUT: A summary message (string).
    LANG_VERSION: 1.1.0
    ALGORITHM:
        1. Initialize counters.
        2. Iterate through the checklist items using FOR EACH.
        3. Access task and status from each item map using `item["key"]` syntax.
        4. Log or process based on status using IF/THEN/ELSE/ENDBLOCK.
        5. Return summary using string concatenation.
    CAVEATS: Assumes input is a valid list of maps. Requires arithmetic tools/support for counts.
    EXAMPLES: ProcessChecklist('[{"task":"A", "status":"done"}]') => "Processed 1 items. Done: 1, Pending: 0"
ENDCOMMENT # Lexer requires ENDCOMMENT here

SET done_count = 0
SET pending_count = 0
SET summary = "" # Final summary

EMIT "Processing checklist..."

FOR EACH item IN checklist_items DO # Iterates list elements
    SET current_task = item["task"]   # Map access
    SET current_status = item["status"] # Map access

    EMIT "Processing Task: " + current_task # Direct concatenation

    IF current_status == "done" THEN
        # SET done_count = done_count + 1 # Requires arithmetic support
        EMIT "  Status: Done"
    ELSE
        # SET pending_count = pending_count + 1 # Requires arithmetic support
        EMIT "  Status: Pending"
    ENDBLOCK # End IF block

ENDBLOCK # End FOR EACH block

# SET summary = "Processed items. Done: " + done_count + ", Pending: " + pending_count # Requires arithmetic + conversion
SET summary = "Finished processing checklist." # Placeholder
RETURN summary

END # End DEFINE PROCEDURE
```

### 2.2 Statements & Syntax (v1.1.0 - Explicit EVAL Model)

Statements are processed line by line. Comments (`#` or `--`) are ignored. Blank lines are allowed.

* **`DEFINE PROCEDURE Name(arg1, arg2, ...)`** `NEWLINE`
    * Starts a procedure definition. Arguments are optional.

* **`COMMENT:`** ... **`ENDCOMMENT`**
    * Defines a multi-line docstring block (lexed as `COMMENT_BLOCK`). See Section 2.4.

* **`SET variable = expression`** `NEWLINE`
    * Assigns the *raw result* of an `expression` to a `variable`. Expression evaluation does *not* automatically resolve `{{placeholders}}`. Use `EVAL()` for that.

* **`CALL target(arg1, arg2, ...)`** `NEWLINE`
    * Invokes another procedure, an LLM, or an external tool. Arguments passed are the *raw evaluated* results of the argument expressions.
    * `CALL ProcedureName(...)`
    * `CALL LLM(prompt_expression)`
    * `CALL TOOL.FunctionName(...)`
        * **Implemented Tools:** `ReadFile`, `WriteFile`, `ListDirectory`, `LineCount`, `SanitizeFilename`, `GitAdd`, `GitCommit`, `VectorUpdate` (mock), `SearchSkills` (mock), `ExecuteCommand`, `GoBuild`, `GoCheck`, `GoTest`, `GoFmt`, `GoModTidy`, `StringLength`, `Substring`, `ToUpper`, `ToLower`, `TrimSpace`, `SplitString`, `SplitWords`, `JoinStrings`, `ReplaceAll`, `Contains`, `HasPrefix`, `HasSuffix`.
    * The *raw result* of the `CALL` is stored internally, accessible via the `LAST` keyword.

* **`RETURN expression`** `NEWLINE`
    * Returns the *raw evaluated* result of the optional `expression`. Terminates the current procedure.

* **`EMIT expression`** `NEWLINE`
    * Evaluates the `expression` to its raw value and prints its string representation to standard output, prefixed with `[EMIT] `.

* **`IF condition THEN`** `NEWLINE` ... **`ELSE`** `NEWLINE` ... **`ENDBLOCK`** `NEWLINE`
    * Starts a conditional block. `ELSE` part is optional.
    * `condition`: `expr1 op expr2` (where `op` is `==`, `!=`, `>`, `<`, `>=`, `<=`) or a single `expression`.
    * Boolean evaluation: `true`, non-zero numbers, strings `"true"`/`"1"` are true. `false`, `0`, other strings, `nil`, lists, maps are false. Comparison follows rules in `evaluation_comparison.go`.
    * Terminated by `ENDBLOCK`.

* **`WHILE condition DO`** `NEWLINE` ... **`ENDBLOCK`** `NEWLINE`
    * Starts a loop block, terminated by `ENDBLOCK`.
    * `condition` syntax same as `IF`.

* **`FOR EACH variable IN collection DO`** `NEWLINE` ... **`ENDBLOCK`** `NEWLINE`
    * Starts a loop block, terminated by `ENDBLOCK`.
    * `collection` is an expression evaluating to a list, map, or string.
    * **Iteration Behavior:**
        * List (`[]interface{}` or `[]string`): Iterates elements. `variable` gets each element.
        * Map (`map[string]interface{}`): Iterates keys (sorted alphabetically). `variable` gets each key (string).
        * String: Iterates characters (runes). `variable` gets each character (string).
        * Nil/Other: 0 iterations.

* **`END`** `NEWLINE`?
    * Terminates `DEFINE PROCEDURE` block.

### 2.3 Expressions, Literals, and Evaluation

* **Literals**:
    * String: `"..."` or `'...'` (standard escapes `\"`, `\'`, `\\`, `\n`, `\r`, `\t`). Represents the raw string content.
    * List: `[` `]` containing comma-separated expressions. Example: `["a", 1, true, ["nested"]]`. Evaluates to `[]interface{}` containing raw evaluated elements.
    * Map: `{` `}` containing comma-separated `string_key : expression` pairs. Keys *must* be string literals. Example: `{"name": "Thing", "value": 10, "tags": ["A", "B"]}`. Evaluates to `map[string]interface{}` containing raw evaluated values.
    * Number: `123`, `4.5`. Parsed as `int64` or `float64`.
    * Boolean: `true`, `false`. Parsed as `bool`.

* **Variables**: `variable_name` (e.g., `my_var`). Evaluates to the raw value stored in the variable.

* **`LAST` Keyword**: Evaluates to the raw value returned by the most recent `CALL` statement in the current scope.

* **Placeholders**: `{{variable_name}}` or `{{LAST}}`. This syntax is *only* processed when used inside a string passed to the `EVAL()` function. In all other contexts, it's treated as part of a raw string or causes a parse error if used standalone where an expression is expected.

* **`EVAL(expression)`**: Evaluates the inner `expression` to get a raw value (typically a string). Then, recursively resolves any `{{placeholder}}` syntax within that resulting string using current variable/`LAST` values. Returns the final resolved string. This is the *only* mechanism for placeholder expansion.

* **Concatenation (`+`)**: Evaluates operands to raw values, converts them to their string representation (`fmt.Sprintf("%v", val)`), and joins the strings. *Does not* resolve placeholders.

* **Element Access**: `collection_expr[accessor_expr]`
    * Evaluates `collection_expr` (must be list or map).
    * Evaluates `accessor_expr`.
    * For lists (`[]interface{}`), accessor must evaluate to an integer (int64) index (0-based). Returns element or error if out of bounds/wrong type.
    * For maps (`map[string]interface{}`), accessor is converted to string key. Returns value or error if key not found.

### 2.4 Docstrings (`COMMENT:` Block)

* Optional block immediately following `DEFINE PROCEDURE`, starting with `COMMENT:` and terminated by the lexer rule using `ENDCOMMENT`.
* Content is skipped by the parser but parsed semantically by the Go `parseDocstring` function.
* Recommended sections (parsed by `parseDocstring`):
    * **`PURPOSE:`** (Required)
    * **`INPUTS:`** (Required) Use `- name: description` format or `INPUTS: None`.
    * **`OUTPUT:`** (Required) Use `OUTPUT: None` if applicable.
    * **`LANG_VERSION:`** (Optional, New) Semantic version string (e.g., `1.1.0`) indicating the NeuroScript version targeted.
    * **`ALGORITHM:`** (Required)
    * **`CAVEATS:`** (Optional)
    * **`EXAMPLES:`** (Optional)

### 2.5 Versioning Conventions (New Section)

* **File Version:**
    * Syntax: `FILE_VERSION "semver_string"` (e.g., `FILE_VERSION "1.0.0"`)
    * Placement: Optional, must appear before any `DEFINE PROCEDURE` lines, usually at the top of the `.ns.txt` file after initial comments/newlines.
    * Purpose: Indicates the version of the *content* within the specific file.
    * Convention: Tooling (like an editor extension or a dedicated script) *should* ideally increment the patch number (the last part) automatically whenever the file is saved with changes. This is a tooling convention, not enforced by the parser.
* **Language Version:**
    * Syntax: `LANG_VERSION: semver_string` (e.g., `LANG_VERSION: 1.1.0`)
    * Placement: Optional, within the `COMMENT:` block of a specific procedure definition.
    * Purpose: Indicates the version of the NeuroScript language specification that the procedure was written for or tested against. Helps manage compatibility as the language evolves.

---

## 3. Storing and Discovering Procedures

*(No change from previous version - structure remains the same)*

### 3.1 Skill Registry Schema

Requires a repository/database storing: name, docstring, code, version/timestamp, embeddings. **[TODO: Implement real indexing]**

### 3.2 Retrieval & Discovery

Vector Search (`TOOL.SearchSkills` mock) or Keyword Search. **[TODO: Implement real search]**

### 3.3 API or Functions

Conceptual: `search_procedures`, `get_procedure`, `save_procedure`.

---

## 4. Interfacing with LLMs

### 4.1 `CALL LLM(prompt_expression)`

1.  Evaluates `prompt_expression` to its raw value (string).
2.  Sends the prompt string to the LLM gateway (Gemini).
3.  Returns the raw text response, storing it internally, accessible via `LAST`.

**Example:**

```neuroscript
SET text_to_analyze = "NeuroScript seems promising."
# Construct prompt using direct concatenation
SET analysis_prompt = "Analyze the following text for sentiment (positive/negative/neutral): " + text_to_analyze
# Call LLM, result stored internally
CALL LLM(analysis_prompt)
# Assign result using LAST
SET analysis_result = LAST

# Compare using direct variable
IF analysis_result == "positive" THEN
    RETURN "Positive Sentiment Detected"
ENDBLOCK
RETURN "Sentiment: " + analysis_result
```

### 4.2 Variation: Provide Context or Additional Instructions [TODO]

Future: `CALL LLM_WITH_CONTEXT(contextData, "prompt")`.

---

## 5. Built-In Reasoning Constructs [TODO]

Future: `ASSERT`, `VERIFY`, `REFLECT`.

---

## 6. Example Workflow

*(No significant change, relies on future tools)*

---

## 7. Implementation and Architecture

### 7.1 NeuroScript Interpreter (Go Implementation)

* **Parsing**: Handles core syntax including `FILE_VERSION`, blocks (`IF/THEN/ELSE/ENDBLOCK`, `WHILE/DO/ENDBLOCK`, `FOR/EACH/IN/DO/ENDBLOCK`), list `[]` and map `{}` literals, `LAST` keyword, `EVAL()` syntax. Uses ANTLR4.
* **Execution**: Handles `SET`, `CALL`, `RETURN`, `EMIT`. Evaluates expressions (concatenation `+`, literals, variables, `LAST`, `EVAL`). Implements list/map element access (`[]`). Executes blocks correctly, including `FOR EACH` iteration over lists (elements), maps (keys), and strings (chars). Implements conditions (`==`, `!=`, `>`, `<`, `>=`, `<=`) based on `evaluation_comparison.go`.
* **Error Handling**: Propagates Go errors from tools/evaluation; basic runtime error reporting. **[TODO: Add specific NeuroScript error types/handling (TRY/CATCH?)]**
* **Docstring Parsing**: Parses `COMMENT:` block content semantically using `parseDocstring` in Go, extracting standard sections and `LANG_VERSION`.

### 7.2 Database / Store

* **[TODO: Implement]** Mocked currently (`TOOL.SearchSkills`, `TOOL.VectorUpdate`).

### 7.3 LLM Gateway

* Currently targets Gemini API. **[TODO: Make configurable]**.

### 7.4 Version Control

* Basic `TOOL.GitAdd`, `TOOL.GitCommit` implemented. **[TODO: More robust Git management needed (branching, status, pull?)]**.

---

## 8. Summary and Future Directions

* **NeuroScript** provides **structured pseudocode** for explicit AI reasoning and skill accumulation.
* **Docstrings** (`COMMENT:` block) are crucial, with conventions like `LANG_VERSION`. File-level versioning via `FILE_VERSION`.
* **Explicit Evaluation**: `EVAL()` is required for placeholder resolution; standard evaluation returns raw values. `LAST` keyword accesses prior `CALL` result.
* **Store/Discover/Retrieve** via external tools connected to Git/Vector DB is key. **[TODO: Implement fully]**
* **LLM Integration** via `CALL LLM` is central.
* **Current Implementation:** Core parsing and execution working for most defined syntax (including blocks, lists, maps, access, `LAST`, `EVAL`). Basic tools available.
* **Next Steps:** Implement real Vector DB/Git integration; Add arithmetic; Add NeuroScript-specific error handling; Add more tools (HTTP, JSON, etc.); Refine LLM integration (context passing, configuration). Consider LSP server implementation.

---
```

This updated specification incorporates the `FILE_VERSION` and `LANG_VERSION` conventions, clarifies the explicit `EVAL` / `LAST` / placeholder model, updates syntax based on the G4 grammar (e.g., `THEN`, `ENDBLOCK`), lists the currently implemented tools, and provides a more accurate summary of the current implementation state.