// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements the tool for reading environment variables.
// filename: pkg/tool/os/tools_os_env.go
// nlines: 30
// risk_rating: HIGH

package os

import (
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolGetenv implements the OS.Getenv tool.
func toolGetenv(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return "", lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Getenv: expected 1 argument (varName)", lang.ErrArgumentMismatch)
	}
	varName, ok := args[0].(string)
	if !ok {
		return "", lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Getenv: varName argument must be a string, got %T", args[0]), lang.ErrInvalidArgument)
	}
	if varName == "" {
		return "", lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Getenv: varName cannot be empty", lang.ErrInvalidArgument)
	}

	// The policy gate should have already checked for the specific variable in scope.
	// e.g., grants: [{Resource: "env", Verbs: ["read"], Scopes: ["HOME", "USER"]}]
	value := os.Getenv(varName)
	//	interpreter.GetLogger().Debug("Tool: Getenv", "variable", varName, "found", value != "")
	return value, nil
}
