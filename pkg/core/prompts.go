// pkg/core/prompts.go
package core

const (
	// PromptDevelop provides strict rules for an LLM generating NeuroScript code.
	PromptDevelop = `You are generating NeuroScript code.
Adhere strictly to the following rules. Generate ONLY the raw code, with no explanations or markdown fences (using three backticks).
**NeuroScript Syntax Rules (v1.1.0 - Explicit EVAL Model):**

1.  **File Structure:** Start with optional '# NeuroScript Syntax: v...' comment, then 'DEFINE PROCEDURE Name(args)', then 'COMMENT:' block, then steps, then 'END'.
2.  **'COMMENT:' Block:** Required after DEFINE. Must include PURPOSE:, INPUTS:, OUTPUT:, ALGORITHM:. Use '- arg: Desc' for INPUTS. End with 'ENDCOMMENT'. (Note: COMMENT_BLOCK token includes start/end keywords).
3.  **Assignment ('SET')**: Use 'SET variable = expression'. Direct assignment is INVALID.
4.  **Invocations ('CALL')**: Use 'CALL Target(args...)'. Target is IDENTIFIER, TOOL.IDENTIFIER, or LLM. Store results using 'SET result = LAST'. Direct invocation is INVALID.
5.  **'LAST' Keyword**: Represents the result of the most recent CALL. Use directly in expressions (e.g., SET x = LAST).
6.  **'EVAL(expr)' Function**: Use 'EVAL(expression)' explicitly to resolve {{placeholders}} within the string result of the expression. Placeholders are NOT resolved automatically in string literals or concatenation.
7.  **Block Structure ('IF', 'WHILE', 'FOR EACH'):**
    * Headers MUST end with 'THEN' (IF) or 'DO' (WHILE/FOR), followed by NEWLINE.
    * Body contains statements, each ending with NEWLINE.
    * Blocks MUST end with 'ENDBLOCK' followed by NEWLINE.
    * 'ELSE' is supported: 'ELSE NEWLINE statement_list'.
8.  **Looping ('FOR EACH')**: ONLY 'FOR EACH var IN collection DO ... ENDBLOCK'.
    * Collection can be list '[]', map '{}', or string. Iterates elements, keys, or characters respectively. String fallback splits by comma.
9.  **Literals**: Strings '""' or '''' (raw, no auto-eval). Lists '[expr, ...]'. Maps '{"key": expr, ...}' (keys are string literals). Numbers '123', '4.5'. Booleans 'true', 'false'.
10. **Element Access**: Use 'list[index_expr]' and 'map[key_expr]'. Index is 0-based integer. Key is evaluated expression converted to string. Map access returns nil (not error) if key not found.
11. **No Built-in Functions (except EVAL)**: Use TOOLs for length, substring, etc.
12. **Available 'TOOL's:** TOOL.ReadFile, TOOL.WriteFile, TOOL.SanitizeFilename, TOOL.VectorUpdate, TOOL.GitAdd, TOOL.GitCommit, TOOL.SearchSkills, TOOL.ListDirectory, TOOL.LineCount, TOOL.ExecuteCommand, TOOL.GoBuild, TOOL.GoCheck, TOOL.GoTest, TOOL.GoFmt, TOOL.GoModTidy, and string tools (StringLength, Substring, ToUpper, ToLower, TrimSpace, SplitString, SplitWords, JoinStrings, ReplaceAll, Contains, HasPrefix, HasSuffix). Do NOT invent tools.
13. **Variables & Placeholders**: '{{varname}}' is syntax used within strings passed to EVAL(). Use 'varname' directly elsewhere.
14. **Comments**: Use '#' or '--' for single-line comments outside the COMMENT_BLOCK.
15. **Output Format**: Generate ONLY raw code. Start with DEFINE PROCEDURE. End with END and a final newline. Do NOT include markdown fences (using three backticks).`

	// PromptExecute provides guidance for an LLM executing NeuroScript code.
	PromptExecute = `You are executing the provided NeuroScript procedure step-by-step. Track variable state.
Key execution points (v1.1.0 - Explicit EVAL model):

* **'SET var = expr'**: Evaluate 'expr' (getting raw value), store result in 'var'. String literals like "Hi {{x}}" remain raw.
* **'CALL Target(...)'**: Evaluate args (raw). Execute Procedure, LLM, or TOOL.FunctionName. Store single return value in internal 'LAST' state.
* **'LAST'**: Keyword evaluates to the raw value returned by the most recent 'CALL'.
* **'EVAL(expr)'**: Evaluate 'expr' to get a raw value (usually a string). Resolve any '{{placeholder}}' syntax within that string using current variable/LAST values recursively (up to iteration limit). Returns the resolved string.
* **'IF cond THEN ... [ELSE ...] ENDBLOCK'**: Evaluate 'cond' (comparisons '==', '!=', '>', '<', '>=', '<='; or single expression). Non-zero numbers, boolean 'true', string '"true"'/'"1"' are truthy. Execute THEN or ELSE block.
* **'WHILE cond DO ... ENDBLOCK'**: Repeat block while 'cond' is true.
* **'FOR EACH var IN coll DO ... ENDBLOCK'**: Evaluate 'coll'. Iterate over list elements, map keys, or string characters, assigning current item/key/char to 'var' in each iteration. Run block.
* **List/Map Literals**: '[...]' evaluates to a list containing the raw evaluated elements. '{ "key": expr, ... }' evaluates to a map containing raw evaluated values.
* **Element Access**: 'list[index]' accesses list element (0-based int index). 'map[key]' accesses map value (key expression converted to string). Map access returns nil if key not found.
* **Concatenation '+'**: Evaluate operands raw, convert to string, join strings. No placeholder resolution.
* **'RETURN expr'**: Evaluate 'expr' (raw), stop procedure, return the value.
* **'EMIT expr'**: Evaluate 'expr' (raw), print its string representation.
* **'END' / 'ENDBLOCK'**: Terminates blocks or procedure.
* **Comments**: '#', '--', 'COMMENT:' blocks are ignored for execution flow.

Execute step-by-step, maintain state, determine final 'RETURN' value or outcome.`
)
