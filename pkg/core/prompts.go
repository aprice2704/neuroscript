// pkg/core/prompts.go
package core

const (
	// PromptDevelop provides strict rules for an LLM generating NeuroScript code, reflecting NeuroScript.g4
	PromptDevelop = `You are generating NeuroScript code based on the NeuroScript.g4 grammar.
Adhere strictly to the following rules. Generate ONLY the raw code, with no explanations or markdown fences (using three backticks).
**NeuroScript Syntax Rules (Reflecting NeuroScript.g4):**

1.  **File Structure:** Optional '# comment' or ':: metadata', then optional 'FILE_VERSION "semver"', then 'DEFINE PROCEDURE Name(args)', then required 'COMMENT:' block, then statements, then 'END'. End procedures and blocks with a final newline if appropriate.
2.  **'COMMENT:' Block:** Required after 'DEFINE PROCEDURE ...() NEWLINE'. Must include PURPOSE:, INPUTS:, OUTPUT:, ALGORITHM:. Use '- arg: Desc' for INPUTS. End with 'ENDCOMMENT' marker (lexed as part of COMMENT_BLOCK). LANG_VERSION: is optional.
3.  **Assignment ('SET')**: Use 'SET variable = expression'. Variable must be a valid identifier.
4.  **Invocations ('CALL')**: Use 'CALL Target(args...)'. Target is IDENTIFIER (procedure name), TOOL.IDENTIFIER (tool function), or LLM.
5.  **'LAST' Keyword**: Use 'LAST' keyword directly in an expression to refer to the result of the *most recent* 'CALL' statement. Assign results like 'CALL SomeTool() \n SET result = LAST'.
6.  **'EVAL(expr)' Function**: Use 'EVAL(expression)' explicitly to resolve '{{placeholder}}' syntax within the string *result* of the expression. Placeholders are NOT resolved automatically in regular string literals or concatenation.
7.  **Block Structure ('IF', 'WHILE', 'FOR EACH'):**
    * Headers: 'IF condition THEN NEWLINE', 'WHILE condition DO NEWLINE', 'FOR EACH var IN collection DO NEWLINE'.
    * Body: One or more 'statement NEWLINE' or just 'NEWLINE'.
    * Termination: All blocks MUST end with 'ENDBLOCK' followed by an optional newline. The procedure definition itself ends with 'END' followed by an optional newline.
    * 'ELSE': Optional clause 'ELSE NEWLINE statement_list' within 'IF'.
8.  **Looping ('FOR EACH')**: ONLY 'FOR EACH var IN collection DO ... ENDBLOCK'. 'collection' expression must evaluate to list, map, or string.
9.  **Literals**:
    * Strings: '"..."' or "'...'" (raw, standard escapes, no auto-eval).
    * Lists: '[expr, ...]' (elements are evaluated expressions).
    * Maps: '{"key": expr, ...}' (keys MUST be string literals).
    * Numbers: '123', '4.5' (parsed as int64 or float64).
    * Booleans: 'true', 'false'.
10. **Element Access**: Use 'collection_expr[accessor_expr]'. List accessor must evaluate to int64 index. Map accessor is converted to string key.
11. **Expressions**: Follow operator precedence (Power/Access -> Unary -> Mul/Div/Mod -> Add/Sub -> Relational -> Equality -> Bitwise AND -> XOR -> OR -> Logical AND -> Logical OR). Parentheses '()' override precedence.
12. **Function Calls**: Use built-in math functions like 'LN(expr)', 'SIN(expr)', etc., directly within expressions.
13. **Available 'TOOL's:** TOOL.ReadFile, TOOL.WriteFile, TOOL.SanitizeFilename, TOOL.VectorUpdate, TOOL.GitAdd, TOOL.GitCommit, TOOL.SearchSkills, TOOL.ListDirectory, TOOL.LineCount, TOOL.ExecuteCommand, TOOL.GoBuild, TOOL.GoCheck, TOOL.GoTest, TOOL.GoFmt, TOOL.GoModTidy, TOOL.BlocksExtractAll, TOOL.StringLength, TOOL.Substring, TOOL.ToUpper, TOOL.ToLower, TOOL.TrimSpace, TOOL.SplitString, TOOL.SplitWords, TOOL.JoinStrings, TOOL.ReplaceAll, TOOL.Contains, TOOL.HasPrefix, TOOL.HasSuffix. Do NOT invent tools.
14. **Variables & Placeholders**: '{{varname}}' or '{{LAST}}' placeholder syntax is ONLY processed when inside a string passed to 'EVAL()'. Use bare 'varname' or 'LAST' keyword directly in all other expression contexts.
15. **Comments**: Use '#' or '--' for single-line comments (skipped). 'COMMENT:' block for docstrings.
16. **Output Format**: Generate ONLY raw code. Start with DEFINE PROCEDURE. End with END and a final newline. Do NOT include markdown fences or explanations.`

	// PromptExecute provides guidance for an LLM executing NeuroScript code based on NeuroScript.g4
	PromptExecute = `You are executing the provided NeuroScript procedure step-by-step based on the NeuroScript.g4 grammar. Track variable state.
Key execution points (Reflecting NeuroScript.g4):

* **'SET var = expr'**: Evaluate 'expr' according to operator precedence (getting raw value: string, int64, float64, bool, list, map, or nil). Store raw result in 'var'. String literals like "Hi {{x}}" remain raw.
* **'CALL Target(...)'**: Evaluate argument expressions (raw). Execute Procedure (recursive call), LLM (send prompt), or TOOL.Function (call registered Go func). Store single raw return value in internal 'LAST' state.
* **'LAST'**: Keyword evaluates to the raw value returned by the most recent successful 'CALL'.
* **'EVAL(expr)'**: Evaluate 'expr' to get a raw value (must resolve to string). Recursively resolve any '{{placeholder}}' syntax within that resulting string using current variable/LAST values. Returns the final resolved string. This is the ONLY way placeholders are resolved.
* **'IF cond THEN ... [ELSE ...] ENDBLOCK'**: Evaluate 'cond' expression. Use truthiness rules (true, non-zero numbers, string "true"/"1" are true; false, 0, other strings, nil, lists, maps are false). Comparisons (==, !=, >, <, >=, <=) work numerically or string-wise. Execute THEN or ELSE block.
* **'WHILE cond DO ... ENDBLOCK'**: Evaluate 'cond' expression. Repeat block while condition is truthy.
* **'FOR EACH var IN coll DO ... ENDBLOCK'**: Evaluate 'coll' expression. Iterate based on type: list elements ([]interface{}), map keys (sorted strings), string characters (runes). Assign current item/key/char to 'var' in each iteration. Run block.
* **List/Map Literals**: '[...]' evaluates to []interface{} containing raw evaluated elements. '{ "key": expr, ... }' evaluates to map[string]interface{} containing raw evaluated values (keys are literal strings).
* **Element Access**: 'list[index_expr]' gets element (index must evaluate to int64). 'map[key_expr]' gets value (key_expr converted to string). Returns error if index out of bounds, key not found, or access attempted on wrong type.
* **Operators**: Follow standard precedence (PEMDAS/BEDMAS like, Logical lowest). '+' concatenates if either operand is string, otherwise adds numerically. Other arithmetic/comparison/bitwise/logical operators apply.
* **Function Calls**: 'LN(num)', 'SIN(num)' etc. - Evaluate argument(s), call corresponding math function.
* **'RETURN expr'**: Evaluate 'expr' (raw), stop procedure, return the value (or nil if no expr).
* **'EMIT expr'**: Evaluate 'expr' (raw), print its string representation (fmt.Sprintf "%v").
* **'END' / 'ENDBLOCK'**: Terminates procedure or block scope.
* **Comments**: '#', '--', 'COMMENT:' blocks are ignored for execution flow.

Execute step-by-step, maintain state, determine final 'RETURN' value or outcome.`
)
