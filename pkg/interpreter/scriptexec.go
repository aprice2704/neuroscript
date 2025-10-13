// NeuroScript Version: 0.8.0
// File version: 4.0.0
// Purpose: Refactored to use the interpreter's injected parser and AST builder instead of creating local instances.
// filename: pkg/interpreter/scriptexec.go
// nlines: 75
// risk_rating: MEDIUM

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// ExecuteScriptString parses and executes a given string of NeuroScript code.
func (i *Interpreter) ExecuteScriptString(scriptName, scriptContent string, args map[string]interface{}) (result lang.Value, rErr *lang.RuntimeError) {
	if i == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "interpreter instance is nil", nil)
	}
	logger := i.Logger() // This will now correctly pull the logger from the HostContext
	logger.Debugf("ExecuteScriptString called: %s", scriptName)

	originalProcName := i.state.currentProcName
	i.state.currentProcName = scriptName
	defer func() {
		i.state.currentProcName = originalProcName
	}()

	defer func() {
		if r := recover(); r != nil {
			rErr = lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("panic occurred during script '%s': %v", scriptName, r), fmt.Errorf("panic: %v", r))
			logger.Error("Panic recovered during ExecuteScriptString", "script_name", scriptName, "panic_value", r, "error", rErr)
			result = nil
		}
	}()

	antlrTree, antlrParseErr := i.parser.Parse(scriptContent)
	if antlrParseErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, fmt.Sprintf("parsing script '%s' failed: %s", scriptName, antlrParseErr.Error()), antlrParseErr)
	}
	if antlrTree == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal error: parser returned nil ANTLR tree for script '%s'", scriptName), nil)
	}

	programAST, _, buildErr := i.astBuilder.Build(antlrTree)
	if buildErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, fmt.Sprintf("building AST for script '%s' failed: %s", scriptName, buildErr.Error()), buildErr)
	}

	if programAST == nil || programAST.Procedures == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal error: AST builder yielded nil ast.Program/Procedures for script '%s'", scriptName), nil)
	}

	scriptProcedure, ok := programAST.Procedures[scriptName]
	if !ok || scriptProcedure == nil || len(scriptProcedure.Steps) == 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal error: failed to extract steps for script '%s' from AST", scriptName), nil)
	}
	stepsToExecute := scriptProcedure.Steps

	var execErr error
	result, _, _, execErr = i.executeSteps(stepsToExecute, false, nil)

	if execErr != nil {
		if re, ok := execErr.(*lang.RuntimeError); ok {
			return result, re
		}
		return result, lang.NewRuntimeError(lang.ErrorCodeExecutionFailed, fmt.Sprintf("execution of script '%s' failed: %v", scriptName, execErr), execErr)
	}

	i.lastCallResult = result
	return result, nil
}
