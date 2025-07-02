// NeuroScript Version: 0.4.0
// File version: 4
// Purpose: Re-implemented toolGoImports here with a robust, self-contained implementation to fix build errors.
// filename: pkg/tool/gotools/tool_go_diagnostics.go
// nlines: 105
// risk_rating: MEDIUM

package gotools

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"golang.org/x/tools/imports"
)

// --- Tool Implementations for Go Diagnostic Commands ---

// toolGoVet implementation
func toolGoVet(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	vetTarget := "./..." // Default target
	if len(args) > 0 {
		if targetArg, ok := args[0].(string); ok && targetArg != "" {
			vetTarget = targetArg
		}
	}
	cmdArgs := []string{"vet", vetTarget}
	return executeGoCommandHelper(interpreter, ".", cmdArgs...)
}

// toolStaticcheck implementation
// NOTE: Assumes 'staticcheck' executable is available in the PATH.
func toolStaticcheck(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	checkTarget := "./..." // Default target
	if len(args) > 0 {
		if targetArg, ok := args[0].(string); ok && targetArg != "" {
			checkTarget = targetArg
		}
	}
	execArgs := []interface{}{"staticcheck", []string{checkTarget}, "."}
	return shell.toolExecuteCommand(interpreter, execArgs)
}

// toolGoImports formats a Go source string and manages imports using golang.org/x/tools/imports.
func toolGoImports(i tool.RunTime, args []interface{}) (interface{}, error) {
	errorResult := func(errMsg string) map[string]interface{} {
		return map[string]interface{}{"success": false, "error": errMsg}
	}

	if len(args) != 1 {
		errMsg := "Go.Imports expects exactly 1 argument: source_code"
		return errorResult(errMsg), lang.NewRuntimeError(lang.ErrorCodeArgMismatch, errMsg, lang.ErrArgumentMismatch)
	}

	source, ok := args[0].(string)
	if !ok {
		errMsg := fmt.Sprintf("invalid argument: expected source code string, got %T", args[0])
		return errorResult(errMsg), lang.NewRuntimeError(lang.ErrorCodeType, errMsg, lang.ErrInvalidArgument)
	}

	if source == "" {
		return "", nil // Nothing to format, success.
	}

	// Process imports and format the source code.
	// The filename argument can be empty as we are processing a byte slice directly.
	processedContent, err := imports.Process("", []byte(source), nil)
	if err != nil {
		errMsg := fmt.Sprintf("failed to process Go imports: %v", err)
		return errorResult(errMsg), lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, errMsg, lang.ErrToolExecutionFailed)
	}

	return string(processedContent), nil
}
