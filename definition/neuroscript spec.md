# NeuroScript: A Pseudocode Framework for AI Reasoning

Version: 2025-03-28

NeuroScript is a structured, human-readable language that provides a *procedural scaffolding* for large language models (LLMs). It is designed to store, discover, and reuse **"skills"** (procedures) with clear docstrings and robust metadata, enabling AI systems to build up a library of **reusable, well-documented knowledge**.

New aspect: we should focus on clarity and simplicity of executing ns for the using LLM, the writing LLM can be expected to work harder to compose ns.

## 1. Goals and Principles

1. **Explicit Reasoning**: Rather than relying on hidden chain-of-thought, NeuroScript encourages step-by-step logic in a code-like format that is both *executable* and *self-documenting*.
2. **Reusable Skills**: Each procedure is stored and can be retrieved via a standard interface. LLMs or humans can then call, refine, or extend these procedures without re-implementing from scratch.
3. **Self-Documenting**: NeuroScript procedures must include docstrings that clarify *purpose*, *inputs*, *outputs*, *algorithmic rationale*, and *edge cases*—mirroring how humans comment code.
4. **LLM Integration**: NeuroScript natively supports calling LLMs for tasks that benefit from free-form generation, pattern matching, or advanced “human-like” reasoning.
5. **Structured Data**: Supports basic list (`[]`) and map (`{}`) literals inspired by JSON/Python for handling structured data.
6. **Multi-Modal Reasoning [TODO]**: The language aims to incorporate constructs for deductive logic (assertions), inductive inference (via LLM calls), reflection, and more.

---

## 2. Language Constructs

### 2.1 High-Level Structure

A NeuroScript file (or “skill” definition) typically contains:

1. **DEFINE PROCEDURE** *Name*( *Arguments* )
2. **COMMENT:** block (Docstring) ending with **END**
3. **Statements** (the pseudocode body, potentially including nested blocks)
4. **END** to close the procedure definition

**Example (Illustrating Lists/Maps and Blocks):**

```neuroscript
DEFINE PROCEDURE ProcessChecklist(checklist_items)
COMMENT:
    PURPOSE: Processes a checklist, potentially logging items.
    INPUTS: - checklist_items: A list of map objects, where each map has "task" (string) and "status" (string). Example: `[{"task":"Review", "status":"done"}, {"task":"Implement", "status":"pending"}]`
    OUTPUT: A summary message (string).
    ALGORITHM:
        1. Initialize counters.
        2. Iterate through the checklist items using FOR EACH.
        3. Access task and status from each item map. [TODO: Implement map access]
        4. Log or process based on status.
        5. Return summary.
    CAVEATS: Assumes input is a valid list of maps. Map access syntax/tools TBD.
    EXAMPLES: ProcessChecklist('[{"task":"A", "status":"done"}]') => "Processed 1 items."
END

SET done_count = 0
SET pending_count = 0
SET item_summary = "" # Placeholder for accessed item info

FOR EACH item IN checklist_items DO # [Interpreter TODO: Iterate native list]
    # TODO: Access map elements, e.g.:
    # SET current_task = item["task"]
    # SET current_status = item["status"]
    SET item_summary = "Processing item..." # Placeholder

    CALL Log({{item_summary}}) # Assuming CALL Log exists or is added

    # TODO: Implement IF current_status == "done" THEN ... logic
    SET done_count = done_count + 1 # [TODO: Requires arithmetic]

END # End FOR EACH block

SET summary = "Processed items. Done: " + done_count + ", Pending: " + pending_count # [TODO: Requires arithmetic/string conversion]
RETURN {{summary}}

END # End DEFINE PROCEDURE
```

### 2.2 Statements & Syntax

Statements are processed line by line. A line ending with `\` continues onto the next line. Comments (`#` or `--`) are ignored.

* **`DEFINE PROCEDURE Name(arg1, arg2, ...)`**
  * Starts a procedure definition.

* **`COMMENT:`**
  * Starts a multi-line docstring block, terminated by `END`. (See Section 2.3).

* **`SET variable = expression`**
  * Assigns the result of an `expression` to a `variable`.
  * `expression` can be:
    * A literal (string `""`/`''`, list `[]`, map `{}`).
    * A number **[TODO: Implement numeric types/evaluation]**.
    * A variable name (e.g., `my_var`).
    * A placeholder (e.g., `{{my_placeholder}}`).
    * A concatenation of terms using `+` (primarily for strings). **[TODO: Define semantics for list/map concatenation?]**
    * The special variable `__last_call_result`.

* **`CALL target(arg1, arg2, ...)`**
  * Invokes another procedure, an LLM, or an external tool.
  * `CALL ProcedureName(...)`
  * `CALL LLM("prompt expression")`
  * `CALL TOOL.FunctionName(...)`
    * **Implemented Tools:** `TOOL.ReadFile`, `TOOL.WriteFile`, `TOOL.SanitizeFilename`, `TOOL.VectorUpdate` (mock), `TOOL.GitAdd`, `TOOL.GitCommit`, `TOOL.SearchSkills` (mock).
    * **[TODO: Implement real Vector DB tools]**
    * **[TODO: Add more tools, e.g., List/Map tools, String tools, HTTP tools?]**

* **`RETURN expression`**
  * Returns the evaluated `expression`.

* **`IF condition THEN`**
  * Starts a conditional block, terminated by `END`.
  * `condition`: `expr1 == expr1`, `expr1 != expr1`, `true`, `false`, variable resolving to "true"/"false". **[TODO: Numeric comparisons]**.
  * Body follows on subsequent lines.

* **`ELSE` [TODO: Block support pending]**
  * Not currently supported for execution.

* **`WHILE condition DO`**
  * Starts a loop block, terminated by `END`.
  * `condition` syntax same as `IF`.
  * Body follows on subsequent lines.

* **`FOR EACH variable IN collection DO`**
  * Starts a loop block, terminated by `END`.
  * `collection` is an expression evaluating to a list, map, string, or comma-separated string.
  * **Iteration Behavior:**
    * If `collection` is a **list**: Iterates over elements. `variable` gets each element. **[Interpreter TODO]**
    * If `collection` is a **map**: Iterates over keys? key-value pairs? **[TODO: Define map iteration]**
    * If `collection` is a **string**: Iterates over characters (runes). `variable` gets each character. **[Interpreter TODO]**
    * Otherwise: Collection is converted to string, split by commas. `variable` gets each part.
  * Body follows on subsequent lines.

* **`END`**
  * Terminates `COMMENT:`, `IF`, `WHILE`, `FOR EACH`, or `DEFINE PROCEDURE` blocks. Must be on its own line.

* **Line Continuation `\`**
  * Joins line with the next.

* **Comments (`#`, `--`)**
  * Ignored to end of line.

### 2.3 Literals

* **String:** `"..."` or `'...'` (standard escapes `\"`, `\'`, `\\`).
* **List:** `[` `]` containing comma-separated expressions. Example: `["a", 1, {{var}}, ["nested"]]`. **[Parser/Interpreter TODO]**
* **Map:** `{` `}` containing comma-separated `string_key : expression` pairs. Example: `{"name": "Thing", "value": 10, "tags": ["A", "B"]}`. **[Parser/Interpreter TODO]**
* **Number:** **[TODO: Define syntax and type]**
* **Boolean:** `true`, `false` **[TODO: Define as distinct type?]**

### 2.4 Docstrings (Structured Comments)

*(No change from previous version - structure remains the same)*
Requires `COMMENT:` block immediately following `DEFINE PROCEDURE`, terminated by `END`. Recommended sections:

* **`PURPOSE:`** (Required)
* **`INPUTS:`** (Required) Use `- name: description` format or `INPUTS: None`.
* **`OUTPUT:`** (Required) Use `OUTPUT: None` if applicable.
* **`ALGORITHM:`** (Required)
* **`CAVEATS:`** (Optional)
* **`EXAMPLES:`** (Optional)

---

Markdown

---

## 3. Storing and Discovering Procedures

### 3.1 Skill Registry Schema

You need a repository or database where each NeuroScript procedure (“skill”) is stored, typically with:

* **name** (unique identifier)
* **docstring** (text metadata)
* **neuroscript_code** (the body of the pseudocode)
* **version** or **timestamp**
* Possibly **embeddings** (for semantic search) **[TODO: Implement real indexing]**

### 3.2 Retrieval & Discovery

* **Vector Search [TODO: Implement real search]**: The docstring (and possibly the code) can be embedded and stored. `TOOL.SearchSkills` provides a mock interface.
* **Keyword/Full-Text Search**: Standard text search on name/docstring/code.

### 3.3 API or Functions

*(Conceptual, depends on system architecture)*

* `search_procedures(query) -> list of matches`
* `get_procedure(name) -> returns docstring + code`
* `save_procedure(name, docstring, code) -> updates or creates skill`

---

## 4. Interfacing with LLMs

### 4.1 `CALL LLM("prompt")`

A built-in statement that:

1. Takes a string prompt expression.
2. Sends the evaluated prompt to an LLM gateway or API endpoint (currently hardcoded for Gemini).
3. Returns the raw text response, storing it in `__last_call_result`.

**Example 1: Simple Text Analysis**

```neuroscript
SET text_to_analyze = "NeuroScript seems promising."
SET analysis_prompt = "Analyze the following text for sentiment (positive/negative/neutral): {{text_to_analyze}}"
CALL LLM({{analysis_prompt}})
SET analysis_result = __last_call_result

# Condition checking might need refinement based on LLM output format
IF analysis_result == "positive" THEN # [TODO: Requires string comparison refinement/tools?]
    RETURN "Positive Sentiment Detected"
END
RETURN "Sentiment: " + analysis_result
Example 2: Using Structured Data (Conceptual)

Code snippet

# Assume checklist_item is a map: {"task": "Write tests", "status": "pending", "assignee": null}
# TODO: Need a way to represent/pass structured data to LLM (e.g., JSON stringify tool)
# SET item_json = CALL TOOL.ToJSON({{checklist_item}}) # TOOL.ToJSON is hypothetical
SET item_json = '{"task": "Write tests", "status": "pending", "assignee": null}' # Placeholder string

SET review_prompt = "Review this checklist item JSON and suggest an assignee based on task type: " + {{item_json}}
CALL LLM({{review_prompt}})
SET suggested_assignee = __last_call_result

# TODO: Add logic to update the original checklist_item map with the suggestion
# e.g., CALL TOOL.MapSet({{checklist_item}}, "assignee", {{suggested_assignee}})

RETURN "Suggested assignee: " + suggested_assignee
4.2 Variation: Provide Context or Additional Instructions [TODO]
NeuroScript might support CALL LLM_WITH_CONTEXT(contextData, "prompt"). The interpreter could embed contextData into the LLM prompt. Requires defining contextData structure and interpreter support.

5. Built-In Reasoning Constructs [TODO]
NeuroScript aims to support multiple forms of reasoning, akin to human cognition:

Deductive – ASSERT, VERIFY, or explicit logic checks. [TODO]
Inductive/Abductive – Typically requires free-form pattern recognition via CALL LLM(...).
Heuristic – Fallback solutions or simple procedures.
Reflection – Meta-analysis via REFLECT block or specific CALL LLM. [TODO]
6. Example Workflow
(Conceptual flow)

User / System: Needs a skill to “Classify text by emotion.”

LLM (Orchestrator Script):

Search the registry (e.g., CALL TOOL.SearchSkills("emotion classification")). [TODO: Requires real search]
If found, CALL the existing procedure.
If not found, generate a new one using CALL LLM with specific instructions based on this spec. Example (Conceptual):
Code snippet

# --- Inside an orchestrator skill ---
SET request = "Create skill for emotion classification"
SET prompt = "Generate NeuroScript procedure for task: {{request}}. Follow spec rules..." # Use detailed prompt
CALL LLM({{prompt}})
SET generated_code = __last_call_result
# ... Sanitize filename, WriteFile, GitAdd, GitCommit, VectorUpdate ...
Store the new procedure using tools (TOOL.WriteFile, TOOL.GitAdd, etc.).
Runtime executes the procedure (either found or newly generated).

Feedback / Revision: If inaccurate, an LLM or user could update the procedure (requires TOOL.ReadFile, generation, TOOL.WriteFile, etc.).


---

## 7. Implementation and Architecture

### 7.1 NeuroScript Interpreter (Go Implementation)

* **Parsing**: Handles core syntax, line continuation, block headers (`IF`/`WHILE`/`FOR EACH`). **[Parser TODO: Implement list `[]` and map `{}` literal parsing]**.
* **Execution**: Handles `SET`, `CALL`, `RETURN`, basic conditions (`==`, `!=`), string concatenation (`+`). Executes block bodies recursively. **[Interpreter TODO: Implement `FOR EACH` iteration for lists, maps, strings]**. **[TODO: Implement list/map element access]**. **[TODO: Implement arithmetic, more conditions, error handling (TRY/ASSERT)]**.
* **Error Handling**: Propagates Go errors.

### 7.2 Database / Store

* **[TODO: Implement]** Mocked currently.

### 7.3 LLM Gateway

* **[TODO: Make configurable]**.

### 7.4 Version Control

* **[TODO: More robust Git management needed]**.

---

## 8. Summary and Future Directions

*(Updated summary)*

* **NeuroScript** provides **structured pseudocode** for explicit AI reasoning and skill accumulation.
* **Docstrings** are crucial.
* **Store/Discover/Retrieve** via external tools connected to Git/Vector DB is key. **[TODO: Implement fully]**
* **LLM Integration** via `CALL LLM` is central.
* **Current Implementation:** Basic parsing (including blocks, line continuation), basic execution (SET, CALL, RETURN, simple IF/WHILE, FOR EACH header), mock/basic tools.
* **Next Steps:** Implement list/map literal parsing; Implement interpreter block execution, list/map/string iteration, and element access; Implement real DB/Git integration; Add arithmetic/complex conditions, error handling, concurrency.

---

```
