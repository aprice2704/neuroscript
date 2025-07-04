// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Use security.ResolveAndSecurePath, add os.Stat check for directory.
// nlines: 115 // Approximate
// risk_rating: HIGH // Due to shell execution capabilities
// filename: pkg/tool/shell/tools_shell.go

package shell

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/security"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolExecuteCommand executes an external command securely within the sandbox.
// Corresponds to ToolSpec "Shell.Execute".
func ToolExecuteCommand(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	toolName := "Shell.Execute"

	// Expected args: command (string), args_list ([]string, optional), directory (string, optional)
	if len(args) < 1 || len(args) > 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 1 to 3 arguments, got %d", toolName, len(args)), lang.ErrArgumentMismatch)
	}

	commandPath, okCmd := args[0].(string)
	if !okCmd {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: command argument must be a string, got %T", toolName, args[0]), lang.ErrInvalidArgument)
	}

	var commandArgs []string
	var targetDirRel string = "." // Default directory relative to sandbox

	// Parse args_list (optional, index 1)
	if len(args) > 1 && args[1] != nil {
		if intfArgs, okIntf := args[1].([]interface{}); okIntf {
			commandArgs = make([]string, len(intfArgs))
			for i, v := range intfArgs {
				if s, okConv := v.(string); okConv {
					commandArgs[i] = s
				} else {
					commandArgs[i] = fmt.Sprintf("%v", v)
				}
			}
		} else if strArgs, okStr := args[1].([]string); okStr {
			commandArgs = strArgs
		} else {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: args_list argument must be a list (slice), got %T", toolName, args[1]), lang.ErrInvalidArgument)
		}
	} else {
		commandArgs = []string{}
	}

	// Parse directory (optional, index 2)
	if len(args) > 2 && args[2] != nil {
		dirStr, okDir := args[2].(string)
		if !okDir {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: directory argument must be a string or null, got %T", toolName, args[2]), lang.ErrInvalidArgument)
		}
		// Allow empty string to mean sandbox root (effectively same as ".")
		if dirStr != "" {
			targetDirRel = dirStr
		}
	}

	// Basic security check on command path itself
	if !IsValidCommandPath(commandPath) {
		errMsg := fmt.Sprintf("%s blocked suspicious command path: %q", toolName, commandPath)
		interpreter.GetLogger().Error(errMsg)
		return nil, lang.NewRuntimeError(lang.ErrorCodeSecurity, errMsg, lang.ErrSecurityViolation)
	}

	// Validate and Resolve Directory
	absValidatedDir, pathErr := security.ResolveAndSecurePath(targetDirRel, interpreter.SandboxDir())
	if pathErr != nil {
		// ResolvePath returns RuntimeError already
		errMsg := fmt.Sprintf("%s: invalid execution directory %q: %v", toolName, targetDirRel, pathErr)
		interpreter.GetLogger().Error(errMsg)
		return nil, pathErr
	}

	// Check if the resolved path exists and is a directory
	dirInfo, statErr := os.Stat(absValidatedDir)
	if statErr != nil {
		sentinel := lang.ErrIOFailed
		ec := lang.ErrorCodeIOFailed
		if os.IsNotExist(statErr) {
			sentinel = lang.ErrNotFound
			ec = lang.ErrorCodeFileNotFound // Use specific code for not found
		} else if os.IsPermission(statErr) {
			sentinel = lang.ErrPermissionDenied
			ec = lang.ErrorCodePermissionDenied
		}
		errMsg := fmt.Sprintf("%s: cannot stat execution directory %q: %v", toolName, targetDirRel, statErr)
		interpreter.GetLogger().Error(errMsg, "absolute_path", absValidatedDir)
		return nil, lang.NewRuntimeError(ec, errMsg, errors.Join(sentinel, statErr))
	}
	if !dirInfo.IsDir() {
		errMsg := fmt.Sprintf("%s: execution path %q is not a directory", toolName, targetDirRel)
		interpreter.GetLogger().Error(errMsg, "absolute_path", absValidatedDir)
		return nil, lang.NewRuntimeError(lang.ErrorCodePathTypeMismatch, errMsg, lang.ErrPathNotDirectory) // Use specific sentinel
	}

	interpreter.GetLogger().Debug(fmt.Sprintf("[%s] Preparing command", toolName), "command", commandPath, "args", commandArgs, "directory", absValidatedDir)

	cmd := exec.Command(commandPath, commandArgs...)
	cmd.Dir = absValidatedDir

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
			stderrStr += fmt.Sprintf("[NeuroScript Execution Error: %v]", execErr)
		}
		interpreter.GetLogger().Warn(fmt.Sprintf("[%s] Command failed", toolName), "command", commandPath, "exit_code", exitCode, "stderr", stderrStr)
	} else {
		interpreter.GetLogger().Debug(fmt.Sprintf("[%s] Command successful", toolName), "command", commandPath, "exit_code", 0)
	}

	resultMap := map[string]interface{}{
		"stdout":    stdoutStr,
		"stderr":    stderrStr,
		"exit_code": int64(exitCode),
		"success":   success,
	}
	return resultMap, nil
}

// IsValidCommandPath performs basic checks
func IsValidCommandPath(command string) bool {
	if command == "" {
		return false
	}
	// Basic check: prevent directory traversal in the command itself.
	// Allow simple paths like "go" or "python" etc. assumed to be in PATH.
	// Disallow absolute paths or paths containing separators to force reliance on PATH
	// unless a more sophisticated allowlist/validation mechanism is implemented.
	if strings.ContainsAny(command, "/\\") {
		return false
	}
	// Disallow common shell metacharacters
	if strings.ContainsAny(command, "|;&$><`") {
		return false
	}
	return true
}
