// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 19:37:06 PDT
// filename: pkg/core/tools_shell.go
package core

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	// Removed encoding/json, go/format, os, path/filepath
)

// registerShellTools adds general shell execution tools to the registry.
// Go-specific tools are registered in registerGoTools.
func registerShellTools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		{
			Spec: ToolSpec{
				Name:        "ExecuteCommand",
				Description: "Executes an arbitrary shell command. WARNING: Use with extreme caution due to security risks. Command path validation is basic. Consider using specific tools (e.g., GoBuild, GitAdd) instead.",
				Args: []ArgSpec{
					{Name: "command", Type: ArgTypeString, Required: true, Description: "The command or executable path."},
					// Changed args_list to ArgTypeSliceString for better type safety if possible, or keep SliceAny if mixed types needed
					{Name: "args_list", Type: ArgTypeSliceString, Required: false, Description: "A list of string arguments for the command."},
					{Name: "directory", Type: ArgTypeString, Required: false, Description: "Optional directory (relative to sandbox) to execute the command in. Defaults to sandbox root."},
				},
				ReturnType: ArgTypeAny, // Returns map {stdout, stderr, exit_code, success}
			},
			Func: toolExecuteCommand,
		},
		// Removed GoBuild, GoCheck, GoTest, GoFmt, GoModTidy registrations
	}
	for _, tool := range tools {
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register Shell tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil // Success
}

// toolExecuteCommand executes an external command securely within the sandbox.
func toolExecuteCommand(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	commandPath := args[0].(string)
	var commandArgs []string
	var targetDirRel string = "." // Default directory relative to sandbox

	// Parse args_list (optional, index 1)
	if len(args) > 1 && args[1] != nil {
		// Expect []string based on updated spec
		strArgs, ok := args[1].([]string)
		if !ok {
			// Fallback: Try converting from []interface{}
			intfArgs, okFallback := args[1].([]interface{})
			if !okFallback {
				errMsg := fmt.Sprintf("ExecuteCommand: args_list type mismatch, expected []string or []interface{}, got %T", args[1])
				interpreter.Logger().Error("[TOOL-EXEC] %s", errMsg)
				return map[string]interface{}{"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false}, nil
			}
			commandArgs = make([]string, len(intfArgs))
			for i, v := range intfArgs {
				commandArgs[i] = fmt.Sprintf("%v", v)
			}
		} else {
			commandArgs = strArgs
		}
	} else {
		commandArgs = []string{} // Empty args list if nil or not provided
	}

	// Parse directory (optional, index 2)
	if len(args) > 2 && args[2] != nil {
		dirStr, ok := args[2].(string)
		if !ok {
			errMsg := fmt.Sprintf("ExecuteCommand: directory argument type mismatch, expected string, got %T", args[2])
			interpreter.Logger().Error("[TOOL-EXEC] %s", errMsg)
			return map[string]interface{}{"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false}, nil
		}
		targetDirRel = dirStr
	}

	// Basic security check on command path itself (preventing injection via command name)
	if strings.Contains(commandPath, "..") || strings.ContainsAny(commandPath, "|;&$><`\\") {
		errMsg := fmt.Sprintf("ExecuteCommand blocked suspicious command path: %q", commandPath)
		interpreter.Logger().Error("[TOOL-EXEC] %s", errMsg)
		return map[string]interface{}{"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false}, nil
	}
	// We don't necessarily need to resolve commandPath here if it's expected to be in PATH

	// --- Security: Validate and Resolve Directory ---
	sandboxRoot := interpreter.SandboxDir()
	absValidatedDir, pathErr := ResolveAndSecurePath(targetDirRel, sandboxRoot)
	if pathErr != nil {
		errMsg := fmt.Sprintf("ExecuteCommand path validation failed for directory %q (relative to sandbox %q): %v", targetDirRel, sandboxRoot, pathErr)
		interpreter.Logger().Error("[TOOL-EXEC] %s", errMsg)
		return map[string]interface{}{"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false}, nil
	}
	// --- End Security ---

	interpreter.Logger().Debug("[TOOL-EXEC] Preparing command", "command", commandPath, "args", commandArgs, "directory", absValidatedDir)

	cmd := exec.Command(commandPath, commandArgs...)
	cmd.Dir = absValidatedDir // *** Run in the validated absolute directory ***

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	execErr := cmd.Run()
	stdoutStr := stdout.String()
	stderrStr := stderr.String()
	exitCode := 0
	success := true

	if execErr != nil {
		success = false
		if exitError, ok := execErr.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			} else {
				exitCode = -1
			}
		} else {
			exitCode = -1
			if stderrStr != "" && !strings.HasSuffix(stderrStr, "\n") {
				stderrStr += "\n"
			}
			stderrStr += fmt.Sprintf("Execution Error: %v", execErr)
		}
		interpreter.Logger().Warn("[TOOL-EXEC] Command failed", "command", commandPath, "args", commandArgs, "dir", absValidatedDir, "exit_code", exitCode, "stderr", stderrStr)
	} else {
		interpreter.Logger().Debug("[TOOL-EXEC] Command successful", "command", commandPath, "args", commandArgs, "dir", absValidatedDir, "exit_code", 0)
	}

	resultMap := map[string]interface{}{
		"stdout":    stdoutStr,
		"stderr":    stderrStr,
		"exit_code": int64(exitCode),
		"success":   success,
	}
	return resultMap, nil // Return map, nil Go-level error
}
