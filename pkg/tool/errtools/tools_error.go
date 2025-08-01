// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Implements the Go function for the 'Error.New' tool.
// filename: pkg/tool/errtools/tools_error.go
// nlines: 30
// risk_rating: LOW

package errtools

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolErrorNew implements the "Error.New" tool function.
func toolErrorNew(i tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("Error.New() expects 2 arguments (code, message), got %d", len(args))
	}

	codeArg := args[0]
	messageArg, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("Error.New() expects a string for the 'message' argument, got %T", args[1])
	}

	var codeValue lang.Value
	if num, isNum := lang.ToFloat64(codeArg); isNum {
		codeValue = lang.NumberValue{Value: num}
	} else if str, isStr := codeArg.(string); isStr {
		codeValue = lang.StringValue{Value: str}
	} else {
		return nil, fmt.Errorf("Error.New() expects a string or number for the 'code' argument, got %T", codeArg)
	}

	errorMap := map[string]lang.Value{
		"code":    codeValue,
		"message": lang.StringValue{Value: messageArg},
	}

	return lang.ErrorValue{Value: errorMap}, nil
}
