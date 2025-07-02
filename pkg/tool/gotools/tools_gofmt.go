// NeuroScript Version: 0.4.0
// File version: 5
// Purpose: Corrected toolGoFmt to return the expected map structure on error, fixing test failures.
// filename: pkg/tool/gotools/tools_gofmt.go
// nlines: 40
// risk_rating: MEDIUM
package gotools

import (
	"fmt"
	"go/format"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// toolGoFmt formats a Go source string using go/format.
func toolGoFmt(i *Interpreter, args []interface{}) (interface{}, error) {
	errorResult := func(errMsg string) map[string]interface{} {
		return map[string]interface{}{"success": false, "error": errMsg}
	}

	if len(args) != 1 {
		errMsg := "Go.Fmt expects exactly 1 argument: source_code"
		return errorResult(errMsg), lang.NewRuntimeError(ErrorCodeArgMismatch, errMsg, ErrArgumentMismatch)
	}

	source, ok := args[0].(string)
	if !ok {
		errMsg := fmt.Sprintf("invalid argument: expected source code string, got %T", args[0])
		return errorResult(errMsg), lang.NewRuntimeError(ErrorCodeType, errMsg, ErrInvalidArgument)
	}

	if source == "" {
		return "", nil	// Nothing to format, success.
	}

	// Format the source code
	formattedContent, err := format.Source([]byte(source))
	if err != nil {
		errMsg := fmt.Sprintf("failed to format Go source: %v", err)
		return errorResult(errMsg), lang.NewRuntimeError(ErrorCodeToolExecutionFailed, errMsg, ErrToolExecutionFailed)
	}

	return string(formattedContent), nil
}