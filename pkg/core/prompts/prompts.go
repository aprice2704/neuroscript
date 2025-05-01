// pkg/core/prompts/prompts.go
package prompts

const (
	// PromptDevelop provides strict rules for an AI generating NeuroScript code, reflecting NeuroScript.g4
	PromptDevelop = "You are generating NeuroScript code based on the NeuroScript.g4 grammar.\n" +
		"Adhere strictly to the following rules. Generate ONLY the raw code, with no explanations or markdown fences (using three backticks).\n" +
		"**NeuroScript Syntax Rules (Reflecting NeuroScript.g4 v0.2.0-alpha-func-returns-2):**\n\n" +
		"1.  **File Structure:** Optional '# comment' or ':: metadata' lines at the START of the file. File-level ':: metadata' (like ':: lang_version:', ':: file_version:') should conventionally be placed at the START. Follow with zero or more procedure definitions.\n" + // Corrected Placement
		"2.  **Procedure Definition:** Start with 'func ProcedureName'. Follow with the signature part (parameters/returns). Follow with 'means' keyword and a newline. End with 'endfunc'.\n" +
		"3.  **Signature Part:** After 'ProcedureName', optionally include clauses 'needs param1, param2', 'optional opt1', 'returns ret1, ret2'. Parentheses '()' around the clauses are optional for grouping only. If no clauses, nothing is needed between the name and 'means'.\n" +
		"4.  **Metadata ('::')**: Procedure-level metadata (like ':: description:', ':: purpose:', ':: param:<name>:', ':: return:<name>:', ':: algorithm:', ':: caveats:') MUST be placed immediately after 'func ... means NEWLINE' and before the first statement. Step-level metadata immediately precedes the step. Use ':: key: value' format. See docs/metadata.md for standard keys.\n" +
		"5.  **Assignment ('set')**: Use 'set variable = expression'. Variable must be a valid identifier.\n" +
		"6.  **Calls (Expressions)**: Procedure and tool calls are expressions. Use them directly where a value is needed, typically with 'set': 'set result = MyProcedure(arg)', 'set data = tool.ReadFile(\"path\")'. Calls can stand alone as statements if the return value is ignored: 'tool.LogMessage(\"Done\")'.\n" +
		"7.  **'last' Keyword**: Use 'last' keyword directly in an expression to refer to the result of the *most recent* successful procedure or tool call expression.\n" +
		"8.  **'eval(expr)' Function**: Use 'eval(expression)' explicitly to resolve '{{placeholder}}' syntax within the string *result* of the expression.\n" +
		"9.  **Placeholders ('{{...}}')**: Placeholder syntax '{{varname}}' or '{{LAST}}' is ONLY processed within raw strings (```...```) during execution OR when explicitly passed to 'eval()'. In normal strings (\"...\", '...') or other expressions, they are treated literally. Use bare 'varname' or 'last' keyword directly in most expression contexts.\n" +
		"10. **Block Structure ('if', 'while', 'for each', 'on_error'):**\n" +
		"    * Headers: 'if condition NEWLINE', 'while condition NEWLINE', 'for each var in collection NEWLINE', 'on_error means NEWLINE'. Note the required newline.\n" +
		"    * Body: One or more 'statement NEWLINE' or just 'newline'.\n" +
		"    * Termination: Use 'endif', 'endwhile', 'endfor', 'endon' respectively.\n" +
		"    * 'else': Optional clause 'else NEWLINE statement_list' within 'if'.\n" +
		"11. **Looping ('for each')**: ONLY 'for each var in collection'. 'collection' expression must evaluate to list, map (iterates values), or string (iterates characters).\n" +
		"12. **Literals**:\n" +
		"    * Strings: '\"...\"' or \"'...'\".\n" + // Escaped quote for example
		"    * Raw Strings: '```...```' (Triple backticks; literal content including newlines, placeholders evaluated on execution/eval).\n" + // Triple backticks included literally
		"    * Lists: '[expr, ...]' (elements are evaluated expressions).\n" +
		"    * Maps: '{\"key\": expr, ...}' (keys MUST be string literals).\n" +
		"    * Numbers: '123', '4.5'.\n" +
		"    * Booleans: 'true', 'false'.\n" +
		"13. **Element Access**: Use 'collection_expr[accessor_expr]'.\n" +
		"14. **Operators**: Follow standard precedence (Power '**' -> Unary '- not no some ~' -> Mul/Div/Mod '* / %' -> Add/Sub '+ -' -> Relational '> < >= <=' -> Equality '== !=' -> Bitwise '& ^ |' -> Logical 'and or'). Use '()' for grouping.\n" +
		"15. **Built-in Functions**: Use math functions like 'ln(expr)', 'sin(expr)', etc., directly within expressions.\n" +
		"16. **Statements**: Valid statements are 'set', 'return', 'emit', 'must', 'mustbe', 'fail', 'clear_error', 'ask', 'if', 'while', 'for each', 'on_error', and call expressions used standalone.\n" +
		"17. **'ask' Statement**: Use 'ask prompt_expr' or 'ask prompt_expr into variable'.\n" +
		"18. **Available 'tool's:** (List may be incomplete, use available tools) tool.ReadFile, tool.WriteFile, tool.ListFiles, tool.ExecuteCommand, tool.GoBuild, tool.GoCheck, tool.GoFmt, tool.GitAdd, tool.GitCommit, tool.VectorIndex, tool.VectorSearch, tool.StrEndsWith, tool.StrContains, tool.StrReplaceAll, etc. Do NOT invent tools.\n" +
		"19. **Comments**: Use '#' or '--' for single-line comments (skipped). Use '::' metadata for documentation.\n" +
		"20. **Output Format**: Generate ONLY raw code. Start with optional ':: metadata', then 'func'. End with 'endfunc' and a final newline. Do NOT include markdown fences or explanations." // Updated Rule 20 to mention optional start metadata

	// PromptExecute provides guidance for an AI executing NeuroScript code based on NeuroScript.g4
	PromptExecute = "You are executing the provided NeuroScript procedure step-by-step based on the NeuroScript.g4 grammar (v0.2.0-alpha-func-returns-2). Track variable state precisely.\n" +
		"Key execution points:\n\n" +
		"* **'set var = expr'**: Evaluate 'expr' according to operator precedence (getting raw value: string, int64, float64, bool, list, map, or nil). Store raw result in 'var'. Normal strings (\"...\", '...') containing '{{...}}' remain raw. Raw strings ('```...```') containing '{{...}}' ARE evaluated upon assignment or use.\n" + // Triple backticks included literally
		"* **Calls (Expressions)**: Evaluate argument expressions (raw). Execute Procedure (recursive call) or TOOL.Function (call registered Go func). Store single raw return value in internal 'last' state.\n" +
		"* **'last'**: Keyword evaluates to the raw value returned by the most recent successful call expression (procedure or tool).\n" +
		"* **'eval(expr)'**: Evaluate 'expr' to get a raw value (must resolve to string). Recursively resolve any '{{placeholder}}' syntax within that resulting string using current variable/'last' values. Returns the final resolved string. This is the primary way placeholders in normal strings are resolved.\n" +
		"* **Placeholders ('{{...}}')**: Primarily resolved via 'eval()' or implicitly within raw strings ('```...```').\n" + // Triple backticks included literally
		"* **'if cond ... [else ...] endif'**: Evaluate 'cond' expression. Use truthiness rules (true, non-zero numbers, string \"true\"/\"1\" are true; false, 0, other strings, nil, empty lists/maps/strings are false). Comparisons (==, !=, >, <, >=, <=) work numerically or string-wise. Execute first or else block. Requires 'endif'.\n" +
		"* **'while cond ... endwhile'**: Evaluate 'cond' expression. Repeat block while condition is truthy. Requires 'endwhile'.\n" +
		"* **'for each var in coll ... endfor'**: Evaluate 'coll' expression. Iterate based on type: list elements ([]interface{}), map values (map[string]interface{}), string characters (runes as strings). Assign current item/value/char to 'var' in each iteration. Run block. Requires 'endfor'.\n" +
		"* **'on_error means ... endon'**: Defines an error handling block. If a runtime error occurs, execution jumps here. 'clear_error' resets the error. Otherwise, error propagates after the block. Requires 'endon'.\n" +
		"* **List/Map Literals**: '[...]' evaluates to []interface{} containing raw evaluated elements. '{ \"key\": expr, ... }' evaluates to map[string]interface{} containing raw evaluated values (keys are literal strings).\n" +
		"* **Element Access**: 'list[index_expr]' gets element (index must evaluate to int64). 'map[key_expr]' gets value (key_expr converted to string). Returns error if index out of bounds, key not found, or access attempted on wrong type.\n" +
		"* **Operators**: Follow standard precedence (PEMDAS/BEDMAS like, Logical lowest). '+' concatenates if either operand is string (converts non-strings), otherwise adds numerically. Other arithmetic/comparison/bitwise/logical ('and', 'or', 'not', 'no', 'some', '~') operators apply. '**' is power.\n" +
		"* **Built-in Functions**: 'ln(num)', 'sin(num)' etc. - Evaluate argument(s), call corresponding math function.\n" +
		"* **'return expr?'**: Evaluate 'expr' (raw), stop procedure, return the value (or nil if no expr).\n" +
		"* **'emit expr'**: Evaluate 'expr' (likely resolving placeholders as if via `eval`), print its string representation (fmt.Sprintf \"%v\").\n" +
		"* **'must expr' / 'mustbe check(args)'**: Evaluate condition/check. Halt with error if false.\n" +
		"* **'fail expr?'**: Evaluate optional message, halt with error.\n" +
		"* **'ask prompt [into var]'**: Evaluate prompt (resolving placeholders). Send to AI client. If 'into var', store text response in 'var'. Result stored in 'last'.\n" +
		"* **Metadata/Comments**: '::', '#', '--' are ignored for execution flow.\n\n" +
		"Execute step-by-step, maintain variable state, handle 'last', determine final 'RETURN' value or error outcome."
)
