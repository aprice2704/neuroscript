 # NeuroScript: A script for humans, AIs and computers
 
 Version: 0.3.0
 
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
     ```
 
 * **Call Statement:** Explicitly calls a procedure or tool if its return value is not being assigned (e.g., via `set`). This statement *must* use the `call` keyword.
     ```neuroscript
     call tool.fs.LogMessage("Starting phase 1")
     call MyProcedureWithoutReturnValues()
     ```
     If a procedure or tool returns a value you wish to use, assign it with `set` (e.g., `set my_data = MyFunctionReturningData()`). If you intend to call a procedure or tool for its side effects only and discard any return value, you must use the `call` keyword. An expression like `MyProcedure()` on its own line is not a valid statement.
 
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
 
 * **Must / MustBe Statement:** Asserts a condition must be true. If false, execution halts with an error. `mustbe` uses a built-in check function (like `is_string`).
     ```neuroscript
     must isTruthy(some_variable) # Check truthiness
     must file_handle != nil
     mustbe is_string(input_name) # Check type using built-in
     mustbe not_empty(input_list)
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
 
 * **Ask Statement:** Sends a prompt expression to a configured AI agent (currently implemented via an `LLMClient` interface) and optionally stores the response in a variable. *(Future versions may allow specifying different agent handles, such as human-in-the-loop or other computational agents).*
     ```neuroscript
     ask "Summarize this text: " + document_content # Send prompt, discard result
     ask "Generate code for: " + task into generated_code # Send prompt, store result
     ```
 
 * **Break Statement:** Exits the innermost loop (`while` or `for each`) immediately.
     ```neuroscript
     while true
       # ... some processing ...
       if condition_met
         break
       endif
     endwhile
     ```
 
 * **Continue Statement:** Skips the rest of the current iteration of the innermost loop (`while` or `for each`) and proceeds to the next iteration.
     ```neuroscript
     for each item in my_list
       if item_is_not_relevant
         continue
       endif
       # ... process relevant item ...
     endfor
     ```
 
 ### 2.4 Control Flow Blocks
 
 * **If Statement:** Conditional execution. Requires `endif`. `else` is optional. `NEWLINE` is required after `if condition` and `else`.
     ```neuroscript
     if count > 10
       emit "Count exceeds threshold."
       set status = "High"
     else
       emit "Count is within limits."
       set status = "Normal"
     endif
     ```
 
 * **While Statement:** Loop while a condition is true. Requires `endwhile`. `NEWLINE` is required after `while condition`.
     ```neuroscript
     set i = 0
     while i < 5
       emit "Iteration: " + i
       set i = i + 1
     endwhile
     ```
 
 * **For Each Statement:** Iterates over elements of a list, map (values), or string (characters). Requires `endfor`. `NEWLINE` is required after the `for each` line.
     ```neuroscript
     for each item in my_list
       emit "Processing item: " + item
     endfor
 
     for each char in "hello"
       emit "Char: " + char
     endfor
     ```
 
 * **On Error Statement:** Defines a block to execute if an error occurs within the current procedure. Requires `endon`. The `means` keyword and a `NEWLINE` are required after `on_error`.
     ```neuroscript
     on_error means
       emit "An error occurred: " + system.error_message # Hypothetical error variable
       # clear_error # Optionally clear the error to allow execution to continue
       fail "Procedure failed." # Or explicitly fail
     endon
     ```
 
 ### 2.5 Expressions
 
 Expressions evaluate to a value. NeuroScript supports:
 
 * **Literals:**
     * Strings: `"Hello"` or `'World'`. Standard escape sequences like `\n`, `\"` are supported. String literals can also span multiple physical lines using a backslash (`\`) at the end of a line within the string's content; the backslash and the newline are processed by the interpreter to form a single string value (see Section 2.8).
     * Raw Strings: ```Code block with {{placeholders}}``` (Triple backticks; allows literal content including newlines). Placeholders are **only** evaluated within raw strings during execution. Raw strings do not support escape sequences or backslash-based line continuation in the same way as regular strings; their content is taken literally between the triple backticks.
     * Numbers: `123`, `3.14` (Parsed as int64 or float64).
     * Booleans: `true`, `false`.
     * Lists: `[1, "apple", true, another_list]` (Ordered sequence, heterogeneous types).
     * Maps: `{"key1": "value", "num": 123, "active": true}` (Key-value pairs, keys must be string literals, values can be any expression).
 * **Variables:** `my_variable` (Accesses the value stored in a variable).
 * **`last` Keyword:** Accesses the result of the most recent procedure or tool call within the current scope.
 * **Placeholders:** `{{variable_name}}` or `{{LAST}}`. These are **only substituted** within raw strings (```...```) or when explicitly evaluated using `eval()`. In normal strings or other contexts, they are treated literally.
 * **Operators (Standard Precedence):**
     * Unary: `-` (negation), `not` (logical NOT), `no` (is zero value), `some` (is not zero value), `~` (bitwise NOT).
     * Power: `**` (exponentiation).
     * Multiplicative: `*`, `/`, `%`.
     * Additive: `+` (also string concatenation), `-`.
     * Bitwise Shift (Not currently in G4 v0.3.0): `<<`, `>>`.
     * Relational: `>`, `<`, `>=`, `<=`.
     * Equality: `==`, `!=`.
     * Bitwise AND: `&`.
     * Bitwise XOR: `^`.
     * Bitwise OR: `|`.
     * Logical AND: `and`.
     * Logical OR: `or`.
 * **Element Access:** `my_list[index]` or `my_map["key"]`. Index/key must be an expression.
 * **Calls (`callable_expr`):**
     * User Procedures: `MyProcedure(arg1, optional=arg2)`
     * Tools: `tool.fs.ReadFile("path/to/file")` or `tool.SimpleTool()`
     * Built-ins: `ln(number)`, `log(number)`, `sin(rad)`, `cos(rad)`, `tan(rad)`, `asin(val)`, `acos(val)`, `atan(val)`.
 * **`eval()`:** Explicitly evaluates placeholders within a string expression.
     ```neuroscript
     set template = "User: {{user_name}}, ID: {{user_id}}" # Normal string, placeholders literal
     set user_name = "Alice"
     set user_id = 123
     set resolved_string = eval(template) # Result: "User: Alice, ID: 123"
 
     set raw_template = ```Data for {{target}}```
     set target = "systemA"
     emit raw_template # Output depends on execution context of 'emit'
                      # If emit evaluates its arg like eval(), output: "Data for systemA"
                      # If emit treats its arg literally, output: "Data for {{target}}"
                      # Current emit likely evaluates like eval()
     emit eval(raw_template) # Explicit evaluation: "Data for systemA"
     ```
 * **Parentheses:** `(1 + 2) * 3` for grouping.
 
 ### 2.6 Comments
 
 Lines starting with `#` or `--` are ignored by the parser. Comments can also appear after a statement on the same line.
 ```neuroscript
 # This is a full-line comment
 set x = 1 -- This is an end-of-line comment
 set y = 2 # So is this.
 set z = 10 + \ # Comment on a line with line continuation
           20   # Comment on the continued line itself
 ```
 
 ### 2.7 Metadata
 
 Lines starting with `:: key: value` define metadata. See `docs/metadata.md` for standard keys and placement guidelines (file-level, procedure-level). File-level metadata must appear at the start of the file. Procedure-level metadata appears after `func ... means NEWLINE` and before the first statement.
 Metadata values can span multiple lines using a backslash (`\`) at the end of a line (see Section 2.8).
 
 ### 2.8 Line Continuation
 
 NeuroScript supports line continuation to improve readability for long lines of code, string literals, or metadata values. This feature was stabilized around grammar version 0.5.0.
 
 * **General Line Continuation:** A backslash (`\`) placed at the very end of a physical line signals that the statement or expression continues on the next physical line. The backslash and the immediately following newline character are consumed by the lexer and are effectively invisible to the parser. This allows for breaking long lines of code without altering their meaning.
     ```neuroscript
     set complex_calculation = my_value1 + my_value2 + my_value3 \
                             + another_value4 - yet_another_value5 \
                             + final_value_on_new_line
 
     if some_very_long_condition_part1 and \
        another_very_long_condition_part2 or \
        condition_part3
       emit "Condition met."
     endif
     ```
 
 * **Continuation in String Literals:** Regular string literals (enclosed in `"` or `'`) can span multiple physical lines by placing a backslash (`\`) at the end of the line *within the string's content*. In this case, the backslash and the newline character become part of the raw text of the string token. The AST builder (specifically, the string unescaping logic) is then responsible for processing this sequence, typically by removing the backslash and the newline to join the string parts. Any leading whitespace on the subsequent physical line inside the script will become part of the string value.
     ```neuroscript
     set multi_line_str = "This is the first part of a long string \
                           that continues on the second line, \
                           and finally on the third line."
     # The resulting string will have the parts joined,
     # including the leading spaces from the indented continuation lines.
     # For example: "This is the first part of a long string that continues on the second line, and finally on the third line."
     # (Exact spacing depends on unescaper's trimming of leading whitespace on continued lines, typically preserved)
     ```
 
 * **Continuation in Metadata Values:** Similar to string literals, the values in `:: key: value` metadata lines can also be continued onto subsequent lines using a backslash (`\`) at the end of the physical line. The backslash and the newline become part of the raw metadata value string.
     ```neuroscript
     :: description: This is a very detailed description of a procedure \
     ::              that needs to span multiple lines for clarity and \
     ::              thoroughness.
     ```
 * **Raw Strings and Line Continuation:** Raw strings (```...```) inherently support multi-line content by their nature; they do not use or require backslash-based line continuation for their content to span multiple lines. The backslash character within a raw string is treated as a literal backslash.
 
 ---
 
 ## 3. Tools (`tool.[group.]<Name>`)
 
 External capabilities are exposed via tools, prefixed with `tool.`. Tool names can be simple identifiers (e.g., `tool.MyTool`) or qualified identifiers composed of multiple parts separated by dots, allowing for logical grouping (e.g., `tool.filesystem.ReadFile`, `tool.network.HTTPGet`).
 
 ```neuroscript
 set content = tool.filesystem.ReadFile("config.json")
 call tool.filesystem.WriteFile("output.log", log_data)
 set sum = tool.math.Add(5, 3)
 call tool.utils.LogMessage("An informational message.")
 ```
 Tools are registered within the interpreter and handle interactions with the file system, network, external processes, etc. Standard tools (availability may vary) include `ReadFile`, `WriteFile` (often grouped, e.g., under `filesystem` or `fs`), `ListFiles`, `ExecuteCommand`, `GoBuild`, `VectorIndex`, `VectorSearch`, `GitAdd`, `GitCommit`, etc. The exact naming and grouping (e.g. `tool.ReadFile` vs `tool.fs.ReadFile`) depends on how they are registered in the interpreter.
 
 ---
 
 ## 4. Core Semantics
 
 ### 4.1 Explicit Evaluation Model (`eval`, `last`, Placeholders)
 
 * **Raw Values Default:** Standard expression evaluation returns *raw* values. String literals containing `{{...}}` are returned *with* the placeholders intact. Variables holding such strings also return the raw string.
 * **`eval(string_expression)`:** This is the **explicit trigger** for placeholder substitution. It evaluates its string argument, finds `{{VAR}}` or `{{LAST}}` placeholders within it, looks up `VAR` or the `last` result, substitutes their string representation, and returns the final resolved string.
 * **`last` Keyword:** Directly accesses the *raw* value returned by the most recent `func` or `tool` call in the current scope.
 * **`{{LAST}}` Placeholder:** Used *within* `eval()` or raw strings to substitute the string representation of the `last` result.
 * **Raw Strings (```...```):** Allow literal content but *also* undergo placeholder substitution during execution steps like `set`, `emit`, `return`, `ask`, or when passed to `eval()`. This differs from normal strings (`"..."`, `'...'`) which *never* have placeholders substituted unless explicitly passed to `eval()`.
 
 ### 4.2 Scope
 
 * Variables set within a procedure are local to that procedure's execution call stack.
 * There is currently no global scope across procedure calls, other than potentially through external state modified by tools.
 
 ### 4.3 Error Handling (`on_error`, `fail`, `must`, `mustbe`)
 
 * Runtime errors (invalid operations, tool failures) trigger the `on_error` block if defined.
 * Inside `on_error`, execution continues sequentially. `clear_error` resets the error state. If the block finishes without `clear_error` or `fail`, the error propagates up.
 * `fail` immediately stops execution and signals an error.
 * `must`/`mustbe` check conditions and trigger a specific `ErrMustConditionFailed` if false.
 
 ---
 
 ## 5. Examples (Illustrative - using updated syntax)
 
 ```neuroscript
 :: lang_version: neuroscript@0.5.0
 :: file_version: 1.0.0
 :: description: Example NeuroScript file demonstrating basic features.
 
 func GreetUser(needs name) means
   :: description: Creates a greeting string.
   :: param:name: The name of the user.
   :: return:0: The formatted greeting.
 
   set message = "Hello, " + name + "!"
   # Note: No eval() needed here as '+' operates on raw strings
   return message
 endfunc
 
 func ReadAndGreet(needs file_path) returns status means
   :: description: Reads a name from a file and emits a greeting. \
   ::              This description is continued on the next line.
   :: param:file_path: Path to the file containing the name.
   :: return:status: Success or error message.
   :: requires_ai: false # Example metadata
 
   set file_content = tool.fs.ReadFile(file_path) # Assuming ReadFile is under 'fs' group
   must file_content != nil # Basic check if read succeeded
 
   set greeting = GreetUser(needs=file_content) # Call user function
   emit greeting
   return "Success"
 
   on_error means
     # Error occurred during ReadFile or GreetUser
     emit "Failed to greet from file: " + file_path
     # Error propagates implicitly as clear_error/fail is not used
   endon
 endfunc
 ```
 
 ---
 
 ## 6. AI Integration (`ask`)
 
 The `ask` statement provides direct access to configured AI agents (currently via an LLM client interface).
 
 ```neuroscript
 func GenerateSummary(needs text_content) returns summary means
   :: description: Asks an AI to summarize text.
   :: requires_ai: true
 
   set prompt = "Please summarize the following text:\n" + \
                text_content
   ask prompt into llm_response
   return llm_response
 endfunc
 ```
 The interpreter manages the connection details and interaction protocol with the configured AI service. *(Future: potentially dispatch to different agent types based on handles).*
 
 ---
 
 ## 7. Tooling Ecosystem (Conceptual)
 
 ### 7.1 Interpreter/Executor
 
 * Parses `.ns` files based on the G4 grammar.
 * Builds an Abstract Syntax Tree (AST).
 * Executes the AST, managing variables, call stacks, and tool interactions.
 * Handles the `ask` statement interaction with a configured AI client.
 
 ### 7.2 Skill Storage & Discovery
 
 * Procedures are stored as individual `.ns` files.
 * **Discovery Mechanism:** Likely involves:
     * Indexing procedure metadata (`description`, `purpose`, `tags`) and potentially code content into a Vector Database.
     * Using `tool.VectorSearch` (or similar) with natural language queries to find relevant skill files.
 * **Version Control:** Git is the assumed underlying version control system. Metadata should facilitate tracking.
 
 ### 7.3 Language Server Protocol (LSP) - Future
 
 * An LSP server would provide IDE features like syntax highlighting, diagnostics (linting), code completion (keywords, variables, procedure names), hover information (displaying metadata), and go-to-definition.
 
 ### 7.4 Version Control Tools
 
 * Basic `tool.GitAdd`, `tool.GitCommit` are available. More Git operations could be added as tools.
 
 ---
 
 ## 8. Summary and Future Directions
 
 * **NeuroScript** provides **structured, procedural code** for explicit AI reasoning and skill accumulation.
 * **Metadata** (`:: key: value`) is crucial. File-level metadata (e.g., `lang_version`, `file_version`) must be at the start of `.ns` files. Procedure documentation uses keys like `purpose`, `param`, `return`, `algorithm`. Metadata values can also span multiple lines using `\`.
 * **Line Continuation:** The backslash (`\`) character is used for line continuation. Globally, it joins lines of code for the parser. Within string literals and metadata values, it allows content to span multiple physical lines, with the `\` and newline becoming part of the raw token text for AST processing.
 * **Explicit Evaluation**: `eval()` is required for substituting placeholders (`{{...}}`) in normal strings. Raw strings (```...```) have placeholders evaluated implicitly in most execution contexts. `last` keyword accesses the prior call result.
 * **Syntax:** Uses `func`/`endfunc`, explicit `call` for standalone calls (calls not part of `set` or other expressions), optional `needs`/`optional`/`returns` clauses (parens optional), `means`, `if`/`endif`, `while`/`endwhile`, `for each`/`endfor`, `on_error`/`endon`, `break`, `continue`. Calls (`MyProc()`, `tool.group.MyTool()`) are expressions; if used as a statement on their own, they require the `call` keyword.
 * **Store/Discover/Retrieve** via external tools connected to Git/Vector DB is key (conceptual).
 * **AI Integration** via `ask ... into ...` is central. *(Future: Agent handles)*.
 * **Current Implementation:** Core parsing and execution working for G4 syntax (v0.5.0 features like line continuation are now included). Basic tools available. Metadata parsing/storage implemented. `ask` statement implemented. Qualified tool names supported. Explicit `call` keyword for standalone calls is enforced.
 * **Next Steps:** Implement real Vector DB/Git integration tools; Add more built-in functions/tools (HTTP, JSON, etc.); Refine AI integration (agent routing, context passing, configuration); Consider LSP server implementation.
 
 ---
 ## 9. Fuzzy Logic Support 🧠 Fuzzy Logic Extension Specification for NeuroScript

To be added for 0.4.0

 #### §F.1 – Type: `fuzzy`
 - A `fuzzy` value is a real number in the closed interval `[0.0, 1.0]`.
 - Semantically represents degrees of truth, confidence, similarity, or match strength.
 - Values outside `[0.0, 1.0]` MUST raise a runtime or validation error.

 #### §F.2 – Fuzzy Value Literals
 - `true` and `false` may be implicitly coerced to `1.0` and `0.0` respectively in fuzzy expressions.
 - Readability aliases MAY be defined:

 | Alias       | Value |
 |-------------|-------|
 | `definitely`| `1.0` |
 | `likely`    | `0.75`|
 | `maybe`     | `0.5` |
 | `unlikely`  | `0.25`|
 | `never`     | `0.0` |

 #### §F.3 – Fuzzy Logical Operators
 Let `a` and `b` be values of type `fuzzy`.

 | Operator | Description             | Semantics               |
 |----------|-------------------------|--------------------------|
 | `!a`     | NOT                     | `1 - a`                  |
 | `a & b`  | AND                     | `min(a, b)`              |
 | `a | b`  | OR                      | `max(a, b)`              |
 | `a ⇒ b`  | Implication             | `max(1 - a, b)` (Gödel)  |
 | `a ↔ b`  | Biconditional (IFF)     | `1 - abs(a - b)`         |

 > Implementations MAY allow ASCII alternatives: `->` for `⇒`, `<->` or `<=>` for `↔`.

 #### §F.4 – Comparison & Coercion Rules
 - Comparisons (e.g. `a > 0.6`, `a == b`) between `fuzzy` values yield a `bool`.
 - Fuzzy values MAY be explicitly coerced to `bool` via thresholding:

 ```neuroscript
 if confidence > 0.8:
   take_action()
 endif
 ```

 - Expressions used in control flow statements (`if`, `while`, `until`) MUST evaluate to `bool`.

 #### §F.5 – Fuzzy-Aware Expressions
 - Arithmetic on `fuzzy` values (e.g. `a + b`, `a / 2.0`) is permitted and returns a `fuzzy`.
 - Division by zero MUST be an error, even in fuzzy contexts.

 #### §F.6 – Fuzzy Value Semantics in Context
 Fuzzy values MAY be used to express:

 | Use Case                  | Example                        |
 |---------------------------|--------------------------------|
 | Confidence from LLM       | `similarity = 0.76`            |
 | Partial checklist         | `step.completion = 0.4`        |
 | Graded rule match         | `if pattern_score > 0.65:`     |
 | Agent belief state        | `belief = rule_a ⇒ rule_b`     |

 #### §F.7 – Fuzzy Aggregation Patterns
 NS scripts MAY implement aggregate fuzzy logic using idioms such as:

 ```neuroscript
 total = 0.0
 for s in checklist:
   total += s.completion
 endfor
 average = total / checklist.length
 if average > 0.75:
   emit "Checklist likely complete"
 endif
 ```

 #### §F.8 – Implementation Notes
 - Fuzzy and `bool` types MUST NOT be implicitly interchangeable.
 - Runtime environments MAY use single-precision floats for storage; however, all comparisons MUST be numerically stable across `[0.0, 1.0]`.
 - Scripts relying on fuzzy logic SHOULD explicitly threshold values before using them in control branches.
