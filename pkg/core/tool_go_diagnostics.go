// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 21:00:18 PDT
// filename: pkg/core/tools_go_diagnostics.go

package core

import (
	"fmt"
	// "bytes" // Not needed directly if executeGoCommandHelper handles it
	// "encoding/json" // Not needed directly
	// "os/exec" // Not needed directly
	// "strings" // Not needed directly
	// "syscall" // Not needed directly
)

// --- Tool Implementations for Go Diagnostic Commands ---

// toolGoVet implementation
func toolGoVet(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	vetTarget := "./..." // Default target
	targetArg := ""      // Variable to hold the user-provided target

	// Argument parsing (optional target)
	if len(args) > 0 && args[0] != nil {
		var ok bool
		targetArg, ok = args[0].(string)
		if !ok {
			errMsg := fmt.Sprintf("optional target argument for GoVet was not a string, got %T", args[0])
			interpreter.Logger().Error("[TOOL-GOVET] %s", errMsg)
			return map[string]interface{}{"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false}, nil
		}
	}

	// Use the provided target if non-empty, otherwise stick with default
	if targetArg != "" {
		vetTarget = targetArg
	}

	cmd := "go"
	cmdArgs := []string{"vet", vetTarget}

	interpreter.Logger().Debug("[TOOL-GOVET] Preparing to execute", "command", cmd, "args", cmdArgs, "target_dir", ".")

	// Use internal helper that runs within the sandbox
	// Assuming executeGoCommandHelper is available in this package scope (or adjust imports if needed)
	return executeGoCommandHelper(interpreter, ".", cmdArgs...)
}

// toolStaticcheck implementation
// NOTE: Assumes 'staticcheck' executable is available in the PATH where 'ng' runs.
// need: go install honnef.co/go/tools/cmd/staticcheck@latest
func toolStaticcheck(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	checkTarget := "./..." // Default target
	targetArg := ""        // Variable to hold the user-provided target

	// Argument parsing (optional target)
	if len(args) > 0 && args[0] != nil {
		var ok bool
		targetArg, ok = args[0].(string)
		if !ok {
			errMsg := fmt.Sprintf("optional target argument for Staticcheck was not a string, got %T", args[0])
			interpreter.Logger().Error("[TOOL-STATICCHECK] %s", errMsg)
			return map[string]interface{}{"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false}, nil
		}
	}

	// Use the provided target if non-empty, otherwise stick with default
	if targetArg != "" {
		checkTarget = targetArg
	}

	cmd := "staticcheck" // Use staticcheck command directly
	cmdArgs := []string{checkTarget}

	interpreter.Logger().Debug("[TOOL-STATICCHECK] Preparing to execute", "command", cmd, "args", cmdArgs, "target_dir", ".")

	// Staticcheck isn't a 'go' subcommand, so use the general toolExecuteCommand
	// Assemble args for toolExecuteCommand: commandPath, commandArgs, targetDirRel
	execArgs := []interface{}{
		cmd,     // commandPath
		cmdArgs, // commandArgs ([]string compatible with toolExecuteCommand's []interface{} handling)
		".",     // targetDirRel (execute in sandbox root)
	}
	return toolExecuteCommand(interpreter, execArgs)
}
