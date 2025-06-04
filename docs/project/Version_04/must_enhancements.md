 :: title: NeuroScript 'must' Enhancements & Structural Validation
 :: version: 0.1.1 
 :: status: proposal
 :: description: Design for enhancing the 'must' keyword and introducing structural validation to reduce post-call clutter and improve script robustness.
 :: date: 2025-06-03
 :: author: Gemini
 
 ## 1. Introduction
 
 To simplify NeuroScript scripting, especially for AI-generated code, and to reduce defensive "post-call clutter," this document proposes enhancements to the `must` keyword and introduces a mechanism for structural validation. These changes build upon the decision to standardize error returns from tools.
 
 ## 2. Standardized Error Returns from Tools
 
 * **Convention:** All tools, when encountering a handled operational error (e.g., file not found, invalid input), **must** return a specific NeuroScript `map` structure, conceptually referred to as an `error` value by scripts.
 * **Structure:** This `error` map will contain at least the following keys:
     * `"code"` (required): A standardized error code (string or integer, ideally from `core.ErrorCode`).
     * `"message"` (required): A human-readable string describing the error.
     * `"details"` (optional): A `map` or `string` for additional context.
 * **Tool `ReturnType`:** A tool's `ToolSpec.ReturnType` will continue to define its *successful* return type. The ability for any tool to return an `error` map instead is a language-level convention.
 * **Tool Implementation:** Go functions implementing tools should return `(nsErrorMapValue, nil)` for handled operational errors, and `(nil, goError)` for unexpected internal Go errors.
 
 ## 3. `must` Keyword Enhancements
 
 The `must` keyword is NeuroScript's primary mechanism for asserting critical conditions and failing loudly if they are not met. Its behavior will be expanded and clarified as follows:
 
 ### 3.1. `must <boolean_expression>` (Statement)
 
 * **Behavior:** This remains the fundamental assertion.
     * Evaluates the `<boolean_expression>`.
     * If the expression result is `false`, the script halts immediately by triggering a runtime error (panic).
     * If the evaluation of the expression itself causes an unhandled runtime error, the script also halts.
 * **Purpose:** Asserting script invariants.
 * **Example:** `must user_count > 0`
 
 ### 3.2. `set <variable> = must <tool_call_expression>` (Assignment)
 
 * **Behavior:** This provides a concise way to call a tool when success is mandatory for the script's logic to proceed.
     1.  The `<tool_call_expression>` is evaluated.
     2.  The returned NeuroScript `Value` is inspected:
         * If the tool returns a standard `error` map (as defined in section 2), the `must` condition fails, and the script halts (panics).
         * If the tool returns any other `Value` (presumed to be the success value), it is assigned to `<variable>`.
         * If the tool call itself results in an unhandled Go-level error from the interpreter, the script also halts.
 * **Purpose:** Simplify tool calls where success is critical, eliminating boilerplate error-checking for these cases.
 * **Example:**
     ```neuroscript
     set config_text = must tool.FS.Read("settings.json")
     set config_map = must tool.JSON.Parse(config_text)
     ```
 
 ### 3.3. `set <variable> = must <map_variable>["<key_name>"] as <expected_type>` (Single Map Key & Type Assertion)
 
 * **Behavior:** Provides compact and safe access to a single map key with type validation.
     1.  Evaluates `<map_variable>`. If it's an `error` map or `nil`, the `must` fails (panics).
     2.  Accesses `<map_variable>["<key_name>"]`.
     3.  **Key Existence:** If `<key_name>` does not exist, `must` fails (panics).
     4.  **Value Type Check:** If the key exists, the retrieved value's type is checked against `<expected_type>` (`string`, `int`, `float`, `bool`, `list`, `map`, `error`). If the type mismatches, or if the value is an `error` map when `<expected_type>` is not `error`, `must` fails (panics).
     5.  If all checks pass, the validated value is assigned to `<variable>`.
 * **Purpose:** Ensure a required map key is present and its value is of the correct basic type, failing hard otherwise.
 * **Example:**
     ```neuroscript
     set user_data = {"name": "Gemini", "id": 123, "active": true}
     set user_id = must user_data["id"] as int
     ```
 
 ### 3.4. `set <var1>, <var2>, ... = must <map_var>[<key1_str>, <key2_str>, ...] as <type1>, <type2>, ...` (Multiple Map Keys & Types Assertion)
 
 * **Behavior:** Extends map key assertion to multiple keys and types simultaneously in a single statement.
     1.  Evaluates `<map_var>`. If it's an `error` map or `nil`, the `must` fails (panics).
     2.  **Count Consistency:** Verifies that the number of variables, keys, and types provided are identical. If not, `must` fails (panics).
     3.  **Atomic Validation:** For each corresponding key-type pair (`<key_i_str>` and `<type_i>`):
         a.  **Key Existence:** Checks if `<key_i_str>` exists in `<map_var>`.
         b.  **Value Type Check:** Retrieves the value and checks if its type matches `<type_i>`. Also checks if the value is an unexpected `error` map (unless `<type_i>` is `error`).
         c.  If *any* key is missing, or *any* type mismatches, or *any* value is an unexpected `error` map, the entire `must` statement fails (panics). Error messages should strive to be specific about the first point of failure.
     4.  **Assignment:** If all validations for all key-type pairs pass, the retrieved and validated values are assigned to `<var1>`, `<var2>`, etc., in order.
 * **Purpose:** Provide a highly compact and robust way to extract and validate multiple required fields from a map.
 * **Example:**
     ```neuroscript
     set user_profile = {"name": "Castor", "id": 456, "role": "admin", "enabled": true}
     set user_name, user_id, is_enabled = must user_profile["name", "id", "enabled"] as string, int, bool
     # user_name is "Castor", user_id is 456, is_enabled is true
     # If "role" was requested and its type was 'string', it would also be included.
