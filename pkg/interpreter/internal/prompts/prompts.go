// Package prompts contains the core prompts for AI interaction with NeuroScript.
//
// :: file_version: 4
// :: lang_version: neuroscript@0.4.0
// :: description: Defines the core development and execution prompts for NeuroScript AI agents, updated for v0.4 features and corrected syntax examples.
// :: author: Gemini
// :: license: MIT
// :: sdi_spec: core-prompts
//
// sdi:design Centralized rule definitions for AI code generation and execution simulation.
package prompts

// --- Core Rules & Preamble ---

const (
	// PreambleDevelop outlines the AI's role in code generation.
	PreambleDevelop = "You are generating NeuroScript code based on the NeuroScript.g4 grammar (reflecting v0.4.0 features like 'must' enhancements).\n" +
		"Adhere strictly to the following rules. Generate ONLY the raw code, with no explanations or markdown fences (using three backticks).\n" +
		"**NeuroScript Syntax Rules (Reflecting Language Spec v0.4.0):**\n\n"

	// PreambleExecute outlines the AI's role in code execution.
	PreambleExecute = "You are executing the provided NeuroScript procedure step-by-step based on the NeuroScript v0.4.0 language specification. Track variable state precisely.\n" +
		"Key execution points:\n\n"

	// TripleBacktick defines the sequence for NeuroScript raw strings.
	TripleBacktick = "```"
)

// --- Syntax & Structure Rules (Primarily for PromptDevelop) ---

const (
	// RuleFileStructure defines the overall layout of a .ns file.
	RuleFileStructure = "1.  **File Structure:** Optional '# comment' or '-- comment' lines. File-level ':: metadata' (e.g., ':: lang_version:', ':: file_version:') MUST be at the START of the file, before any procedure definitions. Follow with zero or more procedure definitions.\n"

	// RuleProcedureDefinition describes how functions are defined.
	RuleProcedureDefinition = "2.  **Procedure Definition:** Start with 'func ProcedureName'. Follow with the signature part. Follow with 'means' keyword and a newline. End with 'endfunc'.\n"

	// RuleSignaturePart details procedure signatures (parameters, returns).
	RuleSignaturePart = "3.  **Signature Part:** After 'ProcedureName', optionally include clauses 'needs param1, param2', 'optional opt1', 'returns ret1, ret2'. If you use parentheses '()', they MUST enclose the *entire* signature part (all needs, optional, and returns clauses), and the closing parenthesis ')' MUST come immediately before the 'means' keyword. If no signature clauses exist, nothing is needed between the procedure name and 'means'.\n"

	// RuleMetadata defines usage and placement of metadata.
	RuleMetadata = "4.  **Metadata ('::')**: Procedure-level metadata (e.g., ':: description:', ':: param:<name>:', ':: return:<name>:') MUST be immediately after 'func ... means NEWLINE' and before the first statement. ast.Step-level metadata immediately precedes the step. Use ':: key: value' format. Values can span lines using '\\' at line end. See docs/metadata.md for standard keys.\n"

	// RuleSetStatement describes variable assignment.
	RuleSetStatement = "5.  **Assignment ('set')**: Use 'set variable = expression'. Variable must be a valid identifier. For mandatory assignments that must succeed, see the `must` keyword rules.\n"

	// RuleCallStatementAnd.Expressions describes how procedures and tools are invoked.
	RuleCallStatementAnd.Expressions = "6.  **Calls**: Procedure and tool calls are expressions. Use in assignments: 'set result = MyProcedure(arg)', 'set data = tool.ReadFile(\"path\")'. To call for side effects without assigning, MUST use the 'call' statement: 'call tool.LogMessage(\"Done\")'. An expression like 'MyProcedure()' on its own line is NOT valid.\n"

	// RuleMustStatement describes the enhanced 'must' keyword for assertions.
	RuleMustStatement = "7. **'must' Keyword for Assertions & Assignments**: `must` is the primary tool for defensive programming. It halts execution with a runtime error if its condition fails, which can be caught by an `on_error` block.\n" +
		"    * **Boolean Assertion**: `must <boolean_expression>` or `mustbe <check_function>`. Halts if the expression is false. `must file_handle != nil`.\n" +
		"    * **Mandatory Successful Assignment**: `set result = must tool.Call()`. Halts if the tool returns a standard `error` map (see Error Handling rule). If successful, assigns the result to the variable.\n" +
		"    * **Map Key & Type Assertion (Single)**: `set val = must my_map[\"key\"] as type`. Halts if `my_map` is an error, if `\"key\"` does not exist, or if the value is not of the specified `type` (`string`, `int`, `float`, `bool`, `list`, `map`, `error`).\n" +
		"    * **Map Key & Type Assertion (Multiple)**: `set v1, v2 = must my_map[\"k1\", \"k2\"] as type1, type2`. Atomically checks for all keys and validates all corresponding types. Halts on the first failure.\n"

	// RuleLastKeyword explains the 'last' keyword.
	RuleLastKeyword = "8.  **'last' Keyword**: Use 'last' keyword directly in an expression to refer to the result of the *most recent* successful procedure or tool call expression that produced a value.\n"

	// RuleEvalFunction describes the 'eval()' function.
	RuleEvalFunction = "9.  **'eval(expr)' Function**: Use 'eval(expression)' explicitly to resolve '{{placeholder}}' syntax within the string *result* of the expression. Essential for resolving placeholders in standard quoted strings.\n"

	// RulePlaceholders explains placeholder syntax.
	RulePlaceholders = "10. **Placeholders ('{{...}}')**: Syntax '{{varname}}' or '{{LAST}}' is auto-resolved within raw strings (```...```) during execution. Use 'eval()' to resolve them within standard quoted strings. Use bare 'varname' or 'last' directly in most other expression contexts.\n"

	// RuleBlockStructure defines control flow block syntax.
	RuleBlockStructure = "11. **Block Structure ('if', 'while', 'for each', 'on_error'):**\n" +
		"    * Headers: 'if condition NEWLINE', 'while condition NEWLINE', 'for each var in collection NEWLINE', 'on_error means NEWLINE'. Note required newline.\n" +
		"    * Body: One or more 'statement NEWLINE' or just 'newline'.\n" +
		"    * Termination: Use 'endif', 'endwhile', 'endfor', 'endon' respectively.\n" +
		"    * 'else': Optional 'else NEWLINE statement_list' within 'if'.\n"

	// RuleLooping describes 'while' and 'for each' loops.
	RuleLooping = "12. **Looping ('while', 'for each')**:\n" +
		"    * 'while condition ... endwhile': Executes body while condition is truthy.\n" +
		"    * 'for each var in collection ... endfor': 'collection' expression must evaluate to list, map (iterates values), or string (iterates characters).\n" +
		"    * 'break': Immediately exits the innermost loop.\n" +
		"    * 'continue': Immediately skips to the next iteration of the innermost loop.\n"

	// RuleLiterals lists available literal types.
	RuleLiterals = "13. **Literals & New Types**:\n" +
		"    * Strings: '\"...\"' or \"'...'\" (support escapes like \\n, \\t; can span lines with '\\' at EOL).\n" +
		"    * Raw Strings: '```...```' (Triple backticks; literal content including newlines, '{{...}}' placeholders evaluated on execution).\n" +
		"    * Lists: '[expr, ...]' (elements are evaluated expressions).\n" +
		"    * Maps: '{\"key\": expr, ...}' (keys MUST be string literals).\n" +
		"    * Numbers: '123', '4.5' (parsed as int64 or float64).\n" +
		"    * Booleans: 'true', 'false'.\n" +
		"    * Nil: 'nil'.\n" +
		"    * **New Types (v0.4)**: `error`, `timedate`, `event`, and `fuzzy` are first-class types. They are typically created by tools (e.g., `tool.Time.Now()`) or runtime events rather than literals.\n"

	// RuleElementAccess describes accessing elements of collections.
	RuleElementAccess = "14. **Element Access**: Use 'collection_expr[accessor_expr]'.\n"

	// RuleOperatorsAndPrecedence outlines operator behavior.
	RuleOperatorsAndPrecedence = "15. **Operators**: Standard precedence (Power '**' -> Unary '- not no some ~ typeof' -> ...). Use '()' for grouping. '+' concatenates strings or adds numbers. **Fuzzy Logic**: When applied to `fuzzy` values, `and` becomes `min(a,b)`, `or` becomes `max(a,b)`, and `not` becomes `1-a`.\n"

	// RuleBuiltinFunctions notes the availability of built-in math functions.
	RuleBuiltinFunctions = "16. **Built-in Functions**: Use math functions like 'ln(expr)', 'sin(expr)', etc., directly within expressions. Also 'typeof(expr)'.\n"

	// RuleStatementSummary lists all valid statement types.
	RuleStatementSummary = "17. **Statements**: Valid statements are 'set', 'call', 'return', 'emit', 'must', 'mustbe', 'fail', 'clear_error', 'ask', 'if', 'while', 'for each', 'on_error', 'break', 'continue'.\n"

	// RuleAskStatement describes the 'ask' statement for AI interaction.
	RuleAskStatement = "18. **'ask' Statement**: Use 'ask prompt_expr' or 'ask prompt_expr into variable'. Interacts with a configured AI agent.\n"

	// RuleToolUsage provides guidelines for using tools.
	RuleToolUsage = "19. **Available 'tool's:** (Refer to 'tooldefs_*.go' or 'tool.Meta.ListTools()' for actual list & signatures) Examples: tool.FS.Read, tool.FS.Write. Tool names can be qualified (e.g., tool.FS.Read). Do NOT invent tools or their arguments/return types.\n"

	// RuleComments describes how to write comments.
	RuleComments = "20. **Comments**: Use '#' or '--' for single-line comments (skipped by parser). Use '::' metadata for documentation.\n"

	// RuleOutputFormatDevelop specifies the expected output format for code generation.
	RuleOutputFormatDevelop = "21. **Output Format**: Generate ONLY raw NeuroScript code. Start with optional file-level ':: metadata', then 'func' definitions. End each procedure with 'endfunc'. Ensure a final newline if content exists.\n"

	// RuleToolErrorHandlingAndReturnValues details behavior of tool calls.
	RuleToolErrorHandlingAndReturnValues = "22. **Tool Error Handling & The Standard `error` Map (v0.4)**:\n" +
		"    * **Standard `error` Map**: Handled operational errors (e.g., file not found, invalid input) are returned by tools as a standard NeuroScript `map` value. This is a normal return value, NOT a script-halting panic.\n" +
		"    * **Error Map Structure**: The map contains `{\"code\":..., \"message\":..., \"details\":...}`.\n" +
		"    * **Checking for Errors**: Use `set result = must tool.Call()` to automatically check for and halt on these `error` maps. The `must` keyword turns the returned `error` map into a script-halting runtime error.\n" +
		"    * **Critical Failures**: Unexpected Go-level errors or panics within a tool will still halt the script and can be caught by an `on_error` block.\n"

	// RuleLineContinuation explains how to continue lines.
	RuleLineContinuation = "23. **Line Continuation ('\\'):**\n" +
		"    * **Code:** A '\\' at the very end of a line followed by a newline joins it with the next line for the parser (e.g., for long expressions or `if` conditions).\n" +
		"    * **String Literals & Metadata Values:** A '\\' at the end of a line *within* a standard string ('\"...\"', \"'...'\"') or a metadata value allows it to span multiple physical lines.\n" +
		"    * **Raw Strings ('```...```'):** Do not use '\\' for line continuation; they inherently support multi-line content.\n"
)

// --- Execution Semantics (Primarily for PromptExecute) ---
const (
	ExecIntro            = "**Execution Semantics (Key Points):**\n"
	ExecSetStatement     = "* **'set var = expr'**: Evaluate 'expr'. If `expr` begins with `must`, perform success/validation checks first. If checks pass, assign the resulting value. Otherwise, halt with a runtime error. Placeholders '{{...}}' in standard strings remain literal unless 'eval()'. Raw strings ('```...```') with '{{...}}' ARE evaluated on use/assignment.\n"
	ExecCalls            = "* **Calls (ast.Expressions & 'call' statement)**: Evaluate args (raw), execute Procedure/TOOL.Function. Return value (raw) available for expression, also stored in 'last'. A standard `error` map return value is a valid return, not a script-halting error.\n"
	ExecLastKeyword      = "* **'last'**: Keyword evaluates to raw value from most recent successful call (procedure/tool) or 'call' statement.\n"
	ExecEvalFunction     = "* **'eval(expr)'**: Evaluate 'expr' (must be string). Recursively resolve '{{placeholder}}' within that string using current var/'last' values. Returns final resolved string.\n"
	ExecPlaceholders     = "* **Placeholders ('{{...}}')**: Resolved via 'eval()' or implicitly in raw strings. Otherwise, likely literal.\n"
	ExecIfStatement      = "* **'if cond ... [else ...] endif'**: Evaluate 'cond'. Truthiness: true, non-zero numbers, \"true\"/\"1\" are true; false, 0, other strings, nil, empty collections are false. Execute relevant block.\n"
	ExecWhileLoop        = "* **'while cond ... endwhile'**: Evaluate 'cond'. Repeat block if truthy. 'break' exits, 'continue' skips to next iteration.\n"
	ExecForEachLoop      = "* **'for each var in coll ... endfor'**: Evaluate 'coll'. Iterate list elements, map values, or string chars. Assign item to 'var'. 'break' exits, 'continue' skips.\n"
	ExecOnErrorBlock     = "* **'on_error means ... endon'**: Jumps here on runtime error (e.g., from a failed `must` check). 'clear_error' resets error state. Otherwise, error propagates post-block.\n"
	ExecLiterals         = "* **List/Map Literals**: '[...]' -> []interface{} (raw elements). '{ \"key\": expr, ... }' -> map[string]interface{} (raw values, literal keys).\n"
	ExecElementAccess    = "* **Element Access**: 'list[index_expr]' (index_expr to int64). 'map[key_expr]' (key_expr to string). Error if OOB, key not found, or wrong type.\n"
	ExecOperators        = "* **Operators**: Standard precedence. '+' concatenates if any operand is string. For `fuzzy` types, `and`/`or`/`not` use fuzzy logic (min/max/1-x).\n"
	ExecBuiltinFunctions = "* **Built-in Functions**: 'ln(num)', 'sin(num)', 'typeof(val)' etc. - Evaluate arg(s), call corresponding function.\n"
	ExecReturnStatement  = "* **'return expr?'**: Evaluate 'expr' (raw), stop procedure, return value (or nil).\n"
	ExecEmitStatement    = "* **'emit expr'**: Evaluate 'expr' (placeholders resolved as if via 'eval'), print string representation.\n"
	ExecMustFail         = "* **'must expr' / 'mustbe check()'**: Evaluate condition. Halt with error if false.\n* **'set x = must ...'**: Evaluate RHS. If it's a standard `error` map, or a map key/type assertion fails, halt with error.\n* **'fail expr?'**: Evaluate message, halt with error.\n"
	ExecAskStatement     = "* **'ask prompt [into var]'**: Evaluate prompt (placeholders resolved). Send to AI. If 'into var', store text response. Result in 'last'.\n"
	ExecMetadataComments = "* **Metadata/Comments**: '::', '#', '--' ignored for execution flow.\n"
	ExecGoal             = "Execute step-by-step, maintain variable state, handle 'last', determine final 'RETURN' value or error outcome.\n"
)

const globalConstants = `
**Global Constants** The following constants are injected into the global variable scope:

1.  Development and Execution Prompts:
    * NEUROSCRIPT_DEVELOP_PROMPT: (string) Contains the detailed PromptDevelop text.
    * NEUROSCRIPT_EXECUTE_PROMPT: (string) Contains the detailed PromptExecute text.

2.  Standardized Type Strings:
    These constants correspond to the string representation of types returned by
    the typeof() built-in function.

    * TYPE_STRING:   (string) Represents the string "string".
    * TYPE_NUMBER:   (string) Represents the string "number".
    * TYPE_BOOLEAN:  (string) Represents the string "boolean".
    * TYPE_LIST:     (string) Represents the string "list".
    * TYPE_MAP:      (string) Represents the string "map".
    * TYPE_NIL:      (string) Represents the string "nil".
    * TYPE_FUNCTION: (string) Represents the string "function".
    * TYPE_TOOL:     (string) Represents the string "tool".
    * TYPE_ERROR:    (string) Represents the string "error".
    * TYPE_TIMEDATE: (string) Represents the string "timedate".
    * TYPE_EVENT:    (string) Represents the string "event".
    * TYPE_FUZZY:    (string) Represents the string "fuzzy".
    * TYPE_UNKNOWN:  (string) Represents the string "unknown".
`

// --- Illustrative Examples (For both Prompts) ---
const (
	ExampleFileHeaderAndSimpleFunc = `
:: lang_version: neuroscript@0.4.0
:: file_version: 1.0.1
:: purpose: Basic NeuroScript file and function example.

func GreetUser(needs user_name returns greeting_message) means
  :: description: Creates a personalized greeting.
  set greeting_message = "Hello, " + user_name + "! Welcome to NeuroScript."
  return greeting_message
endfunc
`
	ExampleMustAndErrorHandling = `
:: lang_version: neuroscript@0.4.0
:: file_version: 1.1.0
:: purpose: Demonstrates robust loading and parsing using 'must'.

func GetAndProcessConfig(needs path returns status) means
  :: description: Securely loads, parses, and validates a JSON config file.
  
  on_error means
    # 'must' failures will trigger a runtime error and land here.
    emit "Operation failed: " + system.error_message
    return "Failed"
  endon

  # Halt if tool returns an 'error' map (e.g., file not found)
  set config_text = must tool.FS.Read(path)

  # Halt if parsing fails and returns an 'error' map
  set config_map = must tool.JSON.Parse(config_text)

  # Atomically extract and validate required keys and their types from the map.
  # Halts if "port" or "host" are missing, or if their types are not int/string.
  set port, host_name = must config_map["port", "host"] as int, string

  emit "Successfully loaded config for " + host_name + " on port " + port
  return "Success"
endfunc
`
	ExampleLoopListAndToolCall = `
func ProcessFiles(needs directory, file_suffix) means
  :: description: Lists files, filters by suffix, and reads content.
  set all_entries = must tool.FS.List(directory)
  set matching_files_contents = []
  for each entry_map in all_entries
    set entry_name = must entry_map["name"] as string
    set is_dir = must entry_map["isDir"] as bool

    if not is_dir and tool.HasSuffix(entry_name, file_suffix)
      set file_path = directory + "/" + entry_name
      set content = must tool.FS.Read(file_path)
      set matching_files_contents = tool.List.Append(matching_files_contents, content)
    endif
  endfor
  return matching_files_contents
endfunc
`
	ExampleRawStringEvalAndAsk = `
func QueryAIAboutUser(needs user_id, user_name, task_description returns ai_response) means
  :: description: Uses raw strings, eval, and 'ask' for an AI query.
  set context_info_raw = ` + "`" + `User ID is {{user_id}}. User name is {{user_name}}.` + "`" + `
  set final_prompt = eval("Based on the context: {{CONTEXT}}, please {{TASK}}.")
  ask final_prompt into ai_response
  return ai_response
endfunc
`
	ExampleLineContinuation = `
func CalculateComplexValue(needs val1, val2, val3, val4 returns final_result) means
  :: description: Shows line continuation in a complex expression.
  set final_result = val1 + val2 + \
                     val3 - val4
  return final_result
endfunc
`
)

// --- Assembled Prompts ---
const (
	// PromptDevelop provides strict rules for an AI generating NeuroScript code.
	PromptDevelop = PreambleDevelop +
		RuleFileStructure +
		RuleProcedureDefinition +
		RuleSignaturePart +
		RuleMetadata +
		RuleSetStatement +
		RuleCallStatementAnd.Expressions +
		RuleMustStatement +
		RuleToolErrorHandlingAndReturnValues +
		RuleLastKeyword +
		RuleEvalFunction +
		RulePlaceholders +
		RuleBlockStructure +
		RuleLooping +
		RuleLiterals +
		RuleElementAccess +
		RuleOperatorsAndPrecedence +
		RuleBuiltinFunctions +
		RuleStatementSummary +
		RuleAskStatement +
		RuleToolUsage +
		RuleComments +
		RuleOutputFormatDevelop +
		RuleLineContinuation +
		globalConstants +
		"\n**Illustrative Examples:**\n" +
		ExampleFileHeaderAndSimpleFunc + "\n" +
		ExampleMustAndErrorHandling + "\n" +
		ExampleLoopListAndToolCall + "\n" +
		ExampleRawStringEvalAndAsk + "\n" +
		ExampleLineContinuation

	// PromptExecute provides guidance for an AI executing NeuroScript code.
	PromptExecute = PreambleExecute +
		ExecSetStatement +
		ExecCalls +
		ExecLastKeyword +
		ExecEvalFunction +
		ExecPlaceholders +
		ExecIfStatement +
		ExecWhileLoop +
		ExecForEachLoop +
		ExecOnErrorBlock +
		ExecLiterals +
		ExecElementAccess +
		ExecOperators +
		ExecBuiltinFunctions +
		ExecReturnStatement +
		ExecEmitStatement +
		ExecMustFail +
		ExecAskStatement +
		ExecMetadataComments +
		RuleToolErrorHandlingAndReturnValues + // Re-using this rule as it's relevant for execution understanding too
		ExecGoal +
		"\n**Consider these examples during execution analysis:**\n" +
		ExampleFileHeaderAndSimpleFunc + "\n" +
		ExampleMustAndErrorHandling + "\n" +
		ExampleLoopListAndToolCall + "\n" +
		ExampleRawStringEvalAndAsk + "\n" +
		ExampleLineContinuation
)
