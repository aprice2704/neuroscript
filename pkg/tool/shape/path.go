// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements the tool.shape.IsValidPath function.
// filename: pkg/tool/shape/path.go
// nlines: 23
// risk_rating: LOW

package shape

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolShapeIsValidPath implements the tool.shape.IsValidPath function.
func toolShapeIsValidPath(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("IsValidPath: expected 1 argument, got %d", len(args)), lang.ErrArgumentMismatch)
	}

	pathStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("IsValidPath: path_string argument must be a string, got %T", args[0]), lang.ErrInvalidArgument)
	}

	_, err := json_lite.ParsePath(pathStr)
	return err == nil, nil
}
