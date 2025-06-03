// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Corrected parser and error code usage based on core_index.json
// Purpose: Provides the ExecuteScriptString method for the Interpreter.
// filename: pkg/core/interpreter_scriptexec.go
// nlines: 80 // Approximate
// risk_rating: MEDIUM

package core

import (
	"fmt"
)

// ExecuteScriptString parses and executes a given string of NeuroScript code.
// scriptName is used for context in error messages or debugging.
// scriptContent is the actual NeuroScript code to execute.
// args is a map of arguments that could be made available to the script (currently not implemented for direct injection).
// It returns the result of the script execution (if any) and a *RuntimeError if an error occurs.
func (i *Interpreter) ExecuteScriptString(scriptName, scriptContent string, args map[string]interface{}) (result interface{}, rErr *RuntimeError) {
	if i == nil {
		return nil, NewRuntimeError(ErrorCodeInternal, "interpreter instance is nil", nil)
	}
	logger := i.Logger()
	logger.Debugf("ExecuteScriptString called: %s", scriptName)

	// Set up current procedure name for logging and context
	originalProcName := i.currentProcName
	i.currentProcName = scriptName // Using scriptName as a temporary procedure name
	defer func() {
		i.currentProcName = originalProcName
		logArgsMap := map[string]interface{}{
			"script_name":        scriptName,
			"restored_proc_name": i.currentProcName,
			"result_type":        fmt.Sprintf("%T", result),
			"error":              rErr,
		}
		logger.Debug("Finished ExecuteScriptString.", "details", logArgsMap)
	}()

	// Recover from panics during script execution
	defer func() {
		if r := recover(); r != nil {
			rErr = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("panic occurred during script '%s': %v", scriptName, r), fmt.Errorf("panic: %v", r))
			logger.Error("Panic recovered during ExecuteScriptString", "script_name", scriptName, "panic_value", r, "error", rErr)
			result = nil
		}
	}()

	// 1. Parsing phase
	parserAPI := NewParserAPI(logger) // NewParserAPI is in core_index.json
	// Wrap the script content in a dummy function to make it parsable as a full program unit
	// This assumes ParseScriptString is for executing arbitrary statements, not defining new funcs.
	wrappedScriptContent := fmt.Sprintf("func %s means\n%s\nendfunc", scriptName, scriptContent)

	antlrTree, antlrParseErr := parserAPI.Parse(wrappedScriptContent) // ParserAPI.Parse returns (antlr.Tree, error)
	if antlrParseErr != nil {
		logger.Errorf("Failed to parse wrapped script '%s': %v", scriptName, antlrParseErr.Error())
		return nil, NewRuntimeError(ErrorCodeSyntax, fmt.Sprintf("parsing script '%s' failed: %s", scriptName, antlrParseErr.Error()), antlrParseErr) // Use ErrorCodeSyntax
	}
	if antlrTree == nil {
		logger.Error("ParserAPI.Parse returned nil ANTLR tree without error for script", "script_name", scriptName)
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("internal error: parser returned nil ANTLR tree for script '%s'", scriptName), nil)
	}

	astBuilder := NewASTBuilder(logger)                    // NewASTBuilder is in core_index.json
	programAST, _, buildErr := astBuilder.Build(antlrTree) // ASTBuilder.Build returns (*Program, map[string]string, error)
	if buildErr != nil {
		logger.Errorf("Failed to build AST from parsed script '%s': %v", scriptName, buildErr.Error())
		return nil, NewRuntimeError(ErrorCodeSyntax, fmt.Sprintf("building AST for script '%s' failed: %s", scriptName, buildErr.Error()), buildErr) // Use ErrorCodeSyntax
	}

	if programAST == nil || programAST.Procedures == nil {
		logger.Error("ASTBuilder.Build returned nil Program or nil Procedures map for script", "script_name", scriptName)
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("internal error: AST builder yielded nil Program/Procedures for script '%s'", scriptName), nil)
	}

	scriptProcedure, ok := programAST.Procedures[scriptName]
	if !ok || scriptProcedure == nil || scriptProcedure.Steps == nil {
		logger.Errorf("Could not find dummy procedure '%s' or its steps in the built AST", scriptName)
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("internal error: failed to extract steps for script '%s' from AST", scriptName), nil)
	}
	stepsToExecute := scriptProcedure.Steps

	// Note: The 'args map[string]interface{}' parameter is present for future extension.

	// 2. Execution phase
	var execErr error
	// Assuming i.executeSteps is the method that runs []Step (from interpreter_exec.go)
	result, _, _, execErr = i.executeSteps(stepsToExecute, false, nil)

	if execErr != nil {
		logger.Errorf("Error executing script '%s': %v", scriptName, execErr)
		if re, ok := execErr.(*RuntimeError); ok {
			return result, re
		}
		return result, NewRuntimeError(ErrorCodeExecutionFailed, fmt.Sprintf("execution of script '%s' failed: %v", scriptName, execErr), execErr) // ErrorCodeExecutionFailed is in core_index.json
	}

	i.lastCallResult = result
	return result, nil
}
