// NeuroScript Version: 0.3.1
// File version: 0.0.9
// Purpose: Defines the syntax analysis tool. Updated to return a list of error maps.
// filename: pkg/tool/syntax/tooldefs_syntax.go
// nlines: 45 // Approximate
// risk_rating: LOW

package syntax

import "fmt"

// analyzeSyntax is the function implementing the tool's logic.
// It matches the expected ToolFunc signature (args as []interface{}).
var analyzeSyntax ToolFunc = func(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("analyzeNSSyntax: expected 1 argument (nsScriptContent), got %d: %w", len(args), ErrArgumentMismatch)	//
	}

	content, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("analyzeNSSyntax: nsScriptContent argument must be a string, got %T: %w", args[0], ErrInvalidArgument)	//
	}

	// AnalyzeNSSyntaxInternal (the Go implementation) will be updated to return []map[string]interface{}
	return AnalyzeNSSyntaxInternal(interpreter, content)
}

// syntaxToolsToRegister defines the ToolImplementation structs for syntax-related tools.
// This variable is used by zz_core_tools_registrar.go to register the tools.
var syntaxToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:		"analyzeNSSyntax",
			Description:	"Analyzes a NeuroScript string for syntax errors. Returns a list of maps, where each map details an error. Returns an empty list if no errors are found.",
			Category:	"Syntax Utilities",
			Args: []ArgSpec{
				{Name: "nsScriptContent", Type: ArgTypeString, Description: "The NeuroScript content to analyze.", Required: true},
			},
			ReturnType:	ArgTypeSliceMap,	// Changed from ArgTypeString
			ReturnHelp: "Returns a list (slice) of maps. Each map represents a syntax error and contains the following keys:\n" +
				"- `Line`: number (1-based) - The line number of the error.\n" +
				"- `Column`: number (0-based) - The character lang.Position in the line where the error occurred.\n" +
				"- `Msg`: string - The error message.\n" +
				"- `OffendingSymbol`: string - The text of the token that caused the error (may be empty).\n" +
				"- `SourceName`: string - Identifier for the source (e.g., 'nsSyntaxAnalysisToolInput').\n" +
				"An empty list is returned if no syntax errors are found.",
			Example: "set script_to_check = `func myFunc means\n  set x = \nendfunc`\n" +
				"set error_list = tool.analyzeNSSyntax(script_to_check)\n" +
				"if tool.List.IsEmpty(error_list) == false\n" +
				"  set first_error = tool.List.Get(error_list, 0)\n" +
				"  emit \"First error on line \" + first_error[\"Line\"] + \": \" + first_error[\"Msg\"]\n" +
				"endif",
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is supplied. " +	//
				"Returns `ErrInvalidArgument` if `nsScriptContent` is not a string, or if the interpreter instance is nil. " +	//
				"The underlying call to `AnalyzeNSSyntaxInternal` might return an error (e.g. `ErrInternal`) if there's an unexpected issue during its processing, though it aims to return an error list.",
		},
		Func:	analyzeSyntax,
	},
}