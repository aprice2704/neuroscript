// NeuroScript Version: 0.3.0
// filename: pkg/core/prompts/prompts.go
package prompts

// --- Core Rules & Preamble ---

const (
	// PreambleDevelop outlines the AI's role in code generation.
	PreambleDevelop = "You are generating NeuroScript code based on the NeuroScript.g4 grammar (reflecting v0.5.0 features like line continuation).\n" +
		"Adhere strictly to the following rules. Generate ONLY the raw code, with no explanations or markdown fences (using three backticks).\n" +
		"**NeuroScript Syntax Rules (Reflecting NeuroScript.g4 v0.5.0 & Language Spec v0.3.0):**\n\n"

	// PreambleExecute outlines the AI's role in code execution.
	PreambleExecute = "You are executing the provided NeuroScript procedure step-by-step based on the NeuroScript.g4 grammar (v0.5.0 features). Track variable state precisely.\n" +
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
	RuleSignaturePart = "3.  **Signature Part:** After 'ProcedureName', optionally include clauses 'needs param1, param2', 'optional opt1', 'returns ret1, ret2'. Parentheses '()' around these clauses are optional, for grouping only. If no clauses, nothing is needed between the name and 'means'.\n"

	// RuleMetadata defines usage and placement of metadata.
	RuleMetadata = "4.  **Metadata ('::')**: Procedure-level metadata (e.g., ':: description:', ':: param:<name>:', ':: return:<name>:') MUST be immediately after 'func ... means NEWLINE' and before the first statement. Step-level metadata immediately precedes the step. Use ':: key: value' format. Values can span lines using '\\' at line end. See docs/metadata.md for standard keys.\n"

	// RuleSetStatement describes variable assignment.
	RuleSetStatement = "5.  **Assignment ('set')**: Use 'set variable = expression'. Variable must be a valid identifier.\n"

	// RuleCallStatementAndExpressions describes how procedures and tools are invoked.
	RuleCallStatementAndExpressions = "6.  **Calls**: Procedure and tool calls are expressions. Use in assignments: 'set result = MyProcedure(arg)', 'set data = tool.ReadFile(\"path\")'. To call for side effects without assigning, MUST use the 'call' statement: 'call tool.LogMessage(\"Done\")'. An expression like 'MyProcedure()' on its own line is NOT valid.\n"

	// RuleLastKeyword explains the 'last' keyword.
	RuleLastKeyword = "7.  **'last' Keyword**: Use 'last' keyword directly in an expression to refer to the result of the *most recent* successful procedure or tool call expression that produced a value.\n"

	// RuleEvalFunction describes the 'eval()' function.
	RuleEvalFunction = "8.  **'eval(expr)' Function**: Use 'eval(expression)' explicitly to resolve '{{placeholder}}' syntax within the string *result* of the expression. Essential for resolving placeholders in standard quoted strings.\n"

	// RulePlaceholders explains placeholder syntax.
	RulePlaceholders = "9.  **Placeholders ('{{...}}')**: Syntax '{{varname}}' or '{{LAST}}' is auto-resolved within raw strings (```...```) during execution. Use 'eval()' to resolve them within standard quoted strings. Use bare 'varname' or 'last' directly in most other expression contexts.\n"

	// RuleBlockStructure defines control flow block syntax.
	RuleBlockStructure = "10. **Block Structure ('if', 'while', 'for each', 'on_error'):**\n" +
		"    * Headers: 'if condition NEWLINE', 'while condition NEWLINE', 'for each var in collection NEWLINE', 'on_error means NEWLINE'. Note required newline.\n" +
		"    * Body: One or more 'statement NEWLINE' or just 'newline'.\n" +
		"    * Termination: Use 'endif', 'endwhile', 'endfor', 'endon' respectively.\n" +
		"    * 'else': Optional 'else NEWLINE statement_list' within 'if'.\n"

	// RuleLooping describes 'while' and 'for each' loops.
	RuleLooping = "11. **Looping ('while', 'for each')**:\n" +
		"    * 'while condition ... endwhile': Executes body while condition is truthy.\n" +
		"    * 'for each var in collection ... endfor': 'collection' expression must evaluate to list, map (iterates values), or string (iterates characters).\n" +
		"    * 'break': Immediately exits the innermost loop.\n" +
		"    * 'continue': Immediately skips to the next iteration of the innermost loop.\n"

	// RuleLiterals lists available literal types.
	RuleLiterals = "12. **Literals**:\n" +
		"    * Strings: '\"...\"' or \"'...'\" (support escapes like \\n, \\t; can span lines with '\\' at EOL).\n" +
		"    * Raw Strings: '```...```' (Triple backticks; literal content including newlines, '{{...}}' placeholders evaluated on execution).\n" +
		"    * Lists: '[expr, ...]' (elements are evaluated expressions).\n" +
		"    * Maps: '{\"key\": expr, ...}' (keys MUST be string literals).\n" +
		"    * Numbers: '123', '4.5' (parsed as int64 or float64).\n" +
		"    * Booleans: 'true', 'false'.\n" +
		"    * Nil: 'nil'.\n"

	// RuleElementAccess describes accessing elements of collections.
	RuleElementAccess = "13. **Element Access**: Use 'collection_expr[accessor_expr]'.\n"

	// RuleOperatorsAndPrecedence outlines operator behavior.
	RuleOperatorsAndPrecedence = "14. **Operators**: Standard precedence (Power '**' -> Unary '- not no some ~ typeof' -> Mul/Div/Mod '* / %' -> Add/Sub '+ -' -> Relational '> < >= <=' -> Equality '== !=' -> Bitwise '& ^ |' -> Logical 'and or'). Use '()' for grouping. '+' concatenates strings or adds numbers.\n"

	// RuleBuiltinFunctions notes the availability of built-in math functions.
	RuleBuiltinFunctions = "15. **Built-in Functions**: Use math functions like 'ln(expr)', 'sin(expr)', etc., directly within expressions. Also 'typeof(expr)'.\n"

	// RuleStatementSummary lists all valid statement types.
	RuleStatementSummary = "16. **Statements**: Valid statements are 'set', 'call', 'return', 'emit', 'must', 'mustbe', 'fail', 'clear_error', 'ask', 'if', 'while', 'for each', 'on_error', 'break', 'continue'.\n"

	// RuleAskStatement describes the 'ask' statement for AI interaction.
	RuleAskStatement = "17. **'ask' Statement**: Use 'ask prompt_expr' or 'ask prompt_expr into variable'. Interacts with a configured AI agent.\n"

	// RuleToolUsage provides guidelines for using tools.
	RuleToolUsage = "18. **Available 'tool's:** (Refer to 'tooldefs_*.go' or 'tool.Meta.ListTools()' for actual list & signatures) Examples: tool.FS.Read, tool.FS.Write, tool.AIWorker.ExecuteStatelessTask, tool.Git.Commit, tool.List.Length. Tool names can be qualified (e.g., tool.FS.Read). Do NOT invent tools or their arguments/return types.\n"

	// RuleComments describes how to write comments.
	RuleComments = "19. **Comments**: Use '#' or '--' for single-line comments (skipped by parser). Use '::' metadata for documentation.\n"

	// RuleOutputFormatDevelop specifies the expected output format for code generation.
	RuleOutputFormatDevelop = "20. **Output Format**: Generate ONLY raw NeuroScript code. Start with optional file-level ':: metadata', then 'func' definitions. End each procedure with 'endfunc'. Ensure a final newline if content exists.\n"

	// RuleToolErrorHandlingAndReturnValues details behavior of tool calls.
	RuleToolErrorHandlingAndReturnValues = "21. **Tool Error Handling & Return Values (for Generation & Execution understanding):**\n" +
		"    * **Primary Error Handling:** Interpreter-level errors during tool execution (Go panics, critical issues) typically trigger an 'on_error' block or halt the script. The direct result assigned in NeuroScript may be 'nil'.\n" +
		"    * **Interpreting Successful Returns (Consult Tool Spec - e.g., tool.Meta.ListTools()):**\n" +
		"        * Data-returning tools (e.g., 'tool.FS.Read', 'tool.Go.GetModuleInfo'): Assign result to a variable. A 'nil' result post-call (without 'on_error') might indicate an issue caught by the tool itself (e.g., file not found but not a panic). Otherwise, variable holds data (string, map, list, etc.).\n" +
		"        * Side-effect tools (e.g., 'tool.FS.Write', 'tool.Git.Commit'): Often return a success message string (e.g., \"OK\", \"Commit successful...\") or nil on success. Presence of a non-nil string implies Go-level success. Do not assume specific string content unless documented for THAT tool.\n" +
		"        * Map-returning tools (e.g., 'tool.AIWorker.ExecuteStatelessTask', 'tool.Shell.Execute'): If call succeeds (no interpreter error), inspect documented keys within the map for specific results ('output', 'error', 'status', 'stdout', 'stderr', etc.).\n" +
		"        * Nil-returning tools (e.g., 'tool.AIWorkerDefinition.Remove'): If specified to return 'nil' on success, script continuation without error implies success.\n" +
		"    * **Best Practices (Generation):** Use 'on_error ... endon' for unexpected failures. Use 'must' or 'if' checks on tool results (e.g., `must result != nil`, `if result_map[\"error\"] != nil`).\n"

	// RuleLineContinuation explains how to continue lines.
	RuleLineContinuation = "22. **Line Continuation ('\\'):**\n" +
		"    * **Code:** A '\\' at the very end of a line followed by a newline joins it with the next line for the parser (e.g., for long expressions or `if` conditions).\n" +
		"    * **String Literals & Metadata Values:** A '\\' at the end of a line *within* a standard string ('\"...\"', \"'...'\"') or a metadata value allows it to span multiple physical lines; the '\\' and newline become part of the raw token, typically processed by the interpreter to join parts.\n" +
		"    * **Raw Strings ('```...```'):** Do not use '\\' for line continuation; they inherently support multi-line content.\n"
)

// --- Execution Semantics (Primarily for PromptExecute) ---
const (
	ExecIntro            = "**Execution Semantics (Key Points):**\n"
	ExecSetStatement     = "* **'set var = expr'**: Evaluate 'expr' (raw value: string, int64, float64, bool, list, map, or nil). Store in 'var'. Placeholders '{{...}}' in standard strings remain literal unless 'eval()'. Raw strings ('```...```') with '{{...}}' ARE evaluated on use/assignment.\n"
	ExecCalls            = "* **Calls (Expressions & 'call' statement)**: Evaluate args (raw), execute Procedure/TOOL.Function. Return value (raw) available for expression, also stored in 'last'. 'call MyProc()' discards return (still populates 'last').\n"
	ExecLastKeyword      = "* **'last'**: Keyword evaluates to raw value from most recent successful call (procedure/tool) or 'call' statement.\n"
	ExecEvalFunction     = "* **'eval(expr)'**: Evaluate 'expr' (must be string). Recursively resolve '{{placeholder}}' within that string using current var/'last' values. Returns final resolved string.\n"
	ExecPlaceholders     = "* **Placeholders ('{{...}}')**: Resolved via 'eval()' or implicitly in raw strings. Otherwise, likely literal.\n"
	ExecIfStatement      = "* **'if cond ... [else ...] endif'**: Evaluate 'cond'. Truthiness: true, non-zero numbers, \"true\"/\"1\" are true; false, 0, other strings, nil, empty collections are false. Execute relevant block.\n"
	ExecWhileLoop        = "* **'while cond ... endwhile'**: Evaluate 'cond'. Repeat block if truthy. 'break' exits, 'continue' skips to next iteration.\n"
	ExecForEachLoop      = "* **'for each var in coll ... endfor'**: Evaluate 'coll'. Iterate list elements, map values, or string chars. Assign item to 'var'. 'break' exits, 'continue' skips.\n"
	ExecOnErrorBlock     = "* **'on_error means ... endon'**: Jumps here on runtime error. 'clear_error' resets error state. Otherwise, error propagates post-block.\n"
	ExecLiterals         = "* **List/Map Literals**: '[...]' -> []interface{} (raw elements). '{ \"key\": expr, ... }' -> map[string]interface{} (raw values, literal keys).\n"
	ExecElementAccess    = "* **Element Access**: 'list[index_expr]' (index_expr to int64). 'map[key_expr]' (key_expr to string). Error if OOB, key not found, or wrong type.\n"
	ExecOperators        = "* **Operators**: Standard precedence. '+' concatenates if any operand is string (converts non-strings), else adds. Others: arithmetic, comparison, bitwise, logical ('and', 'or', 'not', 'no', 'some', '~'), power ('**'), 'typeof'.\n"
	ExecBuiltinFunctions = "* **Built-in Functions**: 'ln(num)', 'sin(num)', 'typeof(val)' etc. - Evaluate arg(s), call corresponding function.\n"
	ExecReturnStatement  = "* **'return expr?'**: Evaluate 'expr' (raw), stop procedure, return value (or nil).\n"
	ExecEmitStatement    = "* **'emit expr'**: Evaluate 'expr' (placeholders resolved as if via 'eval'), print string representation.\n"
	ExecMustFail         = "* **'must expr' / 'mustbe check()'**: Evaluate condition. Halt with error if false.\n* **'fail expr?'**: Evaluate message, halt with error.\n"
	ExecAskStatement     = "* **'ask prompt [into var]'**: Evaluate prompt (placeholders resolved). Send to AI. If 'into var', store text response. Result in 'last'.\n"
	ExecMetadataComments = "* **Metadata/Comments**: '::', '#', '--' ignored for execution flow.\n"
	ExecGoal             = "Execute step-by-step, maintain variable state, handle 'last', determine final 'RETURN' value or error outcome.\n"
)

const globalConstants = `
**Global Constants** The following constants are injected into the global variable scope:

1.  Development and Execution Prompts:
    * NEUROSCRIPT_DEVELOP_PROMPT: (string) Contains the detailed PromptDevelop
        text defined in this file. This is useful if a script needs to instruct an
        AI to generate further NeuroScript code adhering to the same set of rules.
    * NEUROSCRIPT_EXECUTE_PROMPT: (string) Contains the detailed PromptExecute
        text. This might be used for advanced scenarios where a script reasons about
        or simulates execution.

2.  Standardized Type Strings:
    These constants correspond to the string representation of types returned by
    the typeof() built-in function. Using these constants for type comparisons
    is more robust than using literal strings, as it protects against typos and
    centralizes the type string definitions.

    * TYPE_STRING:   (string) Represents the string "string".
        Example: if typeof(my_var) == TYPE_STRING
    * TYPE_NUMBER:   (string) Represents the string "number" (for both integers and floats).
    * TYPE_BOOLEAN:  (string) Represents the string "boolean".
    * TYPE_LIST:     (string) Represents the string "list".
    * TYPE_MAP:      (string) Represents the string "map".
    * TYPE_NIL:      (string) Represents the string "nil".
    * TYPE_FUNCTION: (string) Represents the string "function" (for user-defined procedures).
    * TYPE_TOOL:     (string) Represents the string "tool" (for registered tools).
    * TYPE_ERROR:    (string) Represents the string "error" (for error objects/values).
    * TYPE_UNKNOWN:  (string) Represents the string "unknown" (for types not otherwise classified).

Usage Example in NeuroScript:

func CheckVariableType(needs some_variable) means
  :: description: Checks the type of a variable and emits it.
  set var_type = typeof(some_variable)
  emit "The variable is of type: " + var_type

  if var_type == TYPE_LIST
    emit "It's a list! Processing each item..."
    for each item in some_variable
      emit "Item: " + item
    endfor
  else
    if var_type == TYPE_STRING
      emit "It's a string with length: " + tool.Length(some_variable)
    endif
  endif
endfunc
`

// --- Illustrative Examples (For both Prompts) ---
const (
	ExampleFileHeaderAndSimpleFunc = `
:: lang_version: neuroscript@0.5.0
:: file_version: 1.0.1
:: purpose: Basic NeuroScript file and function example.

func GreetUser(needs user_name returns greeting_message) means
  :: description: Creates a personalized greeting.
  :: param:user_name: The name of the user.
  :: return:greeting_message: The composed greeting string.

  set greeting_message = "Hello, " + user_name + "! Welcome to NeuroScript."
  return greeting_message
endfunc
`

	ExampleControlFlowIfElse = `
func CheckValue(needs input_value) means
  :: description: Demonstrates if/else if/else logic.
  if input_value > 100
    emit "Value is greater than 100."
  else
    if input_value > 50
      emit "Value is greater than 50 but not over 100."
    else
      emit "Value is 50 or less."
    endif
  endif
endfunc
`
	ExampleLoopListAndToolCall = `
func ProcessFiles(needs directory, file_suffix) means
  :: description: Lists files, filters by suffix, and reads content.
  :: param:directory: The directory to scan.
  :: param:file_suffix: The suffix to filter files by (e.g., ".txt").

  on_error means
    emit "An error occurred during file processing: " + last # 'last' might hold error info from tool
    fail "File processing failed."
  endon

  set all_entries = tool.FS.List(directory)
  must typeof(all_entries) == "list" # Ensure FS.List returned a list

  set matching_files_contents = []
  for each entry_map in all_entries
    set entry_name = entry_map["name"]
    set is_dir = entry_map["isDir"]

    if not is_dir and tool.HasSuffix(entry_name, file_suffix)
      set file_path = directory + "/" + entry_name
      emit "Reading file: " + file_path
      set content = tool.FS.Read(file_path)
      if content != nil and typeof(content) == "string"
        set matching_files_contents = tool.List.Append(matching_files_contents, content)
      else
        emit "[WARN] Could not read or content invalid for: " + file_path
      endif
    endif
  endfor

  emit "Total matching files processed: " + tool.List.Length(matching_files_contents)
  # Further processing of matching_files_contents could happen here
  return matching_files_contents
endfunc
`
	ExampleRawStringEvalAndAsk = `
func QueryAIAboutUser(needs user_id, user_name, task_description, ai_worker_def) returns ai_response means
  :: description: Uses raw strings, eval, and 'ask' for an AI query.
  :: param:user_id: The ID of the user.
  :: param:user_name: The name of the user.
  :: param:task_description: The task for the AI.
  :: param:ai_worker_def: Name of the AI worker definition to use.

  set context_info_raw = ` + TripleBacktick + `User ID is {{user_id}}. User name is {{user_name}}.` + TripleBacktick + `
  # 'context_info_raw' now holds "User ID is <actual_user_id>. User name is <actual_user_name>."

  set prompt_template_std = "Based on the context: {{CONTEXT}}, please {{TASK}}."
  set intermediate_prompt = tool.StrReplaceAll(prompt_template_std, "{{CONTEXT}}", context_info_raw)
  set final_prompt = eval(tool.StrReplaceAll(intermediate_prompt, "{{TASK}}", task_description))
  # 'final_prompt' has all placeholders resolved.

  emit "Sending to AI (" + ai_worker_def + "): " + final_prompt
  ask final_prompt into ai_response

  if ai_response == nil or ai_response == ""
    emit "[WARN] AI returned no response or an empty response."
    return "AI query failed to yield a response."
  endif
  return ai_response
endfunc
`
	ExampleLineContinuation = `
:: file_version: 1.2.0
:: purpose: Demonstrate line continuation for code and strings.

func CalculateComplexValue(needs val1, val2, val3, val4, val5) returns final_result means
  :: description: Shows line continuation in a complex expression.
  set intermediate_sum = val1 + val2 + \
                         val3 - val4

  if intermediate_sum > 100 and \
     val5 < 50 or \
     val1 == val2
    set final_result = intermediate_sum * val5
  else
    set final_result = intermediate_sum + val5
  endif

  set long_message = "The calculation for values (" + val1 + ", " + val2 + ", " + val3 + ", " + val4 + ", " + val5 + ") \
has been completed and the intermediate sum was: " + intermediate_sum + ". \
The final result is: " + final_result + "."
  emit long_message
  return final_result
endfunc
`
	ExampleHandlingToolMapReturn = `
func ExecuteRemoteTask(needs worker_name, command_prompt) returns status_summary means
  :: description: Example of calling a tool that returns a map and checking its fields.
  :: param:worker_name: The AI worker definition for the task.
  :: param:command_prompt: The prompt to send to the worker.

  set task_result_map = tool.AIWorker.ExecuteStatelessTask(worker_name, command_prompt, nil)

  if task_result_map == nil or typeof(task_result_map) != "map"
    emit "[ERROR] AIWorker task for '" + worker_name + "' returned nil or non-map."
    fail "AI task did not return a valid map."
  endif

  set output_content = task_result_map["output"] # Expected key for successful output
  set error_message = task_result_map["error"]   # Expected key for error message

  if error_message != nil and typeof(error_message) == "string" and error_message != ""
    emit "[ERROR] AIWorker task '" + worker_name + "' reported error: " + error_message
    set status_summary = "Task failed: " + error_message
  else
    if output_content != nil and typeof(output_content) == "string"
      emit "[INFO] AIWorker task '" + worker_name + "' successful. Output preview: " + tool.Substring(output_content, 0, 50)
      # Assuming 'output_content' is the primary successful result.
      # Potentially write 'output_content' to a file or process further.
      set status_summary = "Task Succeeded. Output length: " + tool.Length(output_content)
    else
      emit "[WARN] AIWorker task '" + worker_name + "' had no error message, but output was nil or not a string. Type: " + typeof(output_content)
      set status_summary = "Task completed with unexpected output format."
    endif
  endif
  return status_summary
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
		RuleCallStatementAndExpressions +
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
		RuleToolErrorHandlingAndReturnValues +
		RuleLineContinuation +
		globalConstants +
		"\n**Illustrative Examples:**\n" +
		ExampleFileHeaderAndSimpleFunc + "\n" +
		ExampleControlFlowIfElse + "\n" +
		ExampleLoopListAndToolCall + "\n" +
		ExampleRawStringEvalAndAsk + "\n" +
		ExampleLineContinuation + "\n" +
		ExampleHandlingToolMapReturn

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
		"\n**Consider these examples during execution analysis:**\n" + // Slightly different intro for examples in execute prompt
		ExampleFileHeaderAndSimpleFunc + "\n" +
		ExampleControlFlowIfElse + "\n" +
		ExampleLoopListAndToolCall + "\n" +
		ExampleRawStringEvalAndAsk + "\n" +
		ExampleLineContinuation + "\n" +
		ExampleHandlingToolMapReturn
)
