// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 20:56:53 PDT // Split file: Go command execution tools
// filename: pkg/tool/gotools/tools_go_execution.go

package gotools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/security"
	"github.com/aprice2704/neuroscript/pkg/tool"
	// No need for 'errors', 'os', 'path/filepath', 'logging', 'modfile' imports here
)

// --- Tool Implementations for Go Command Execution ---

// toolGoBuild implementation
func toolGoBuild(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	buildTarget := "./..." // Default target
	targetArg := ""        // Variable to hold the user-provided target

	// Argument parsing (optional target)
	if len(args) > 0 && args[0] != nil {
		var ok bool
		targetArg, ok = args[0].(string)
		if !ok {
			errMsg := fmt.Sprintf("optional target argument for GoBuild was not a string, got %T", args[0])
			interpreter.Logger().Error("[TOOL-GOBUILD] %s", errMsg)
			return map[string]interface{}{"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false}, nil
		}
	}

	// Use the provided target if non-empty, otherwise stick with default
	if targetArg != "" {
		buildTarget = targetArg
	}

	cmd := "go"
	cmdArgs := []string{"build", buildTarget}

	interpreter.Logger().Debug("[TOOL-GOBUILD] Preparing to execute", "command", cmd, "args", cmdArgs, "target_dir", ".")

	// Use internal helper that runs within the sandbox
	return executeGoCommandHelper(interpreter, ".", cmdArgs...)
}

// toolGoCheck implementation
func toolGoCheck(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	// Validation ensures 1 string argument (handled by interpreter before calling)
	targetPath := args[0].(string)

	cmd := "go"
	cmdArgs := []string{"list", "-e", "-json", targetPath}

	interpreter.Logger().Debug("[TOOL-GOCHECK] Preparing to execute", "command", cmd, "args", cmdArgs, "target_dir", ".")

	// Execute the command using internal helper
	execResultIntf, execCmdErr := executeGoCommandHelper(interpreter, ".", cmdArgs...)
	if execCmdErr != nil {
		return map[string]interface{}{"check_success": false, "error_details": fmt.Sprintf("Internal execution helper error: %v", execCmdErr)}, nil
	}
	execResultMap, ok := execResultIntf.(map[string]interface{})
	if !ok {
		return map[string]interface{}{"check_success": false, "error_details": fmt.Sprintf("Internal error: Execute helper returned unexpected type %T", execResultIntf)}, nil
	}

	// Analyze the result
	checkSuccess := true
	errorDetails := ""
	execSuccess, _ := execResultMap["success"].(bool)
	execStderr, _ := execResultMap["stderr"].(string)
	execStdout, _ := execResultMap["stdout"].(string)

	if !execSuccess {
		checkSuccess = false
		errorDetails = fmt.Sprintf("go list command failed (exit code %v). Stderr: %s", execResultMap["exit_code"], execStderr)
	} else {
		// Decode JSON stream even if stderr has warnings
		decoder := json.NewDecoder(strings.NewReader(execStdout))
		foundError := false
		var firstErrorMsg string
		var firstErrorPkg string

		for decoder.More() {
			var pkgInfo map[string]interface{}
			if err := decoder.Decode(&pkgInfo); err != nil {
				checkSuccess = false
				firstErrorMsg = fmt.Sprintf("Error decoding JSON from 'go list' output: %v. Output: %s", err, execStdout)
				foundError = true
				break // Stop on first decode error
			}
			// Check the "Error" field within the JSON object for this package
			if errField, ok := pkgInfo["Error"]; ok && errField != nil {
				foundError = true
				if errMap, okMap := errField.(map[string]interface{}); okMap {
					if errMsg, okStr := errMap["Err"].(string); okStr && errMsg != "" {
						firstErrorMsg = errMsg // Store the specific error message
						if importPath, okPath := pkgInfo["ImportPath"].(string); okPath {
							firstErrorPkg = importPath
						}
						break // Stop on first reported error
					}
				}
				// Fallback if Error field structure is unexpected
				firstErrorMsg = fmt.Sprintf("Non-nil 'Error' field detected in package JSON.")
				if importPath, okPath := pkgInfo["ImportPath"].(string); okPath {
					firstErrorPkg = importPath
				}
				break // Stop on first reported error
			}
		}

		if foundError {
			checkSuccess = false
			if firstErrorPkg != "" {
				errorDetails = fmt.Sprintf("Error in package %q: %s", firstErrorPkg, firstErrorMsg)
			} else {
				errorDetails = fmt.Sprintf("Error reported by 'go list -e -json': %s", firstErrorMsg)
			}
			// Append stderr if it contains useful info not captured in JSON error
			if execStderr != "" && !strings.Contains(errorDetails, execStderr) {
				errorDetails += "\nStderr: " + execStderr
			}
		} else {
			// No errors found in JSON output
			checkSuccess = true
			errorDetails = ""
			// Still might log stderr if it exists, indicates warnings?
			if execStderr != "" {
				interpreter.Logger().Warn("[TOOL-GOCHECK] Command succeeded but produced stderr", "stderr", execStderr)
			}
		}
	}

	if checkSuccess {
		interpreter.Logger().Debug("[TOOL-GOCHECK] Successful.")
	} else {
		interpreter.Logger().Warn("[TOOL-GOCHECK] Failed.", "details", errorDetails)
	}

	checkResultMap := map[string]interface{}{
		"check_success": checkSuccess,
		"error_details": errorDetails,
	}
	return checkResultMap, nil
}

// toolGoTest implementation
func toolGoTest(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	testTarget := "./..." // Default target
	targetArg := ""       // Variable to hold the user-provided target

	// Argument parsing (optional target)
	if len(args) > 0 && args[0] != nil {
		var ok bool
		targetArg, ok = args[0].(string)
		if !ok {
			errMsg := fmt.Sprintf("optional target argument for GoTest was not a string, got %T", args[0])
			interpreter.Logger().Error("[TOOL-GOTEST] %s", errMsg)
			return map[string]interface{}{"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false}, nil
		}
	}
	if targetArg != "" {
		testTarget = targetArg // Use provided target
	}

	cmd := "go"
	cmdArgs := []string{"test", testTarget}

	interpreter.Logger().Debug("[TOOL-GOTEST] Preparing to execute", "command", cmd, "args", cmdArgs, "target_dir", ".")

	// Use internal helper that runs within the sandbox
	return executeGoCommandHelper(interpreter, ".", cmdArgs...)
}

// toolGoModTidy implementation
func toolGoModTidy(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	cmd := "go"
	cmdArgs := []string{"mod", "tidy"}

	interpreter.Logger().Debug("[TOOL-GOMODTIDY] Preparing to execute", "command", cmd, "args", cmdArgs, "target_dir", ".")

	// Use internal helper that runs within the sandbox
	return executeGoCommandHelper(interpreter, ".", cmdArgs...)
}

// toolGoListPackages implementation
func toolGoListPackages(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	var targetDirRel string = "."             // Default relative dir
	var patterns []string = []string{"./..."} // Default pattern

	// Argument Parsing (using type assertions validated by interpreter)
	if len(args) > 0 && args[0] != nil {
		targetDirRel = args[0].(string)
	}
	if len(args) > 1 && args[1] != nil {
		// Handle potential slice types from script conversion
		switch p := args[1].(type) {
		case []string:
			patterns = p
		case []interface{}:
			strPatterns := make([]string, len(p))
			allStrings := true
			for i, v := range p {
				s, ok := v.(string)
				if !ok {
					allStrings = false
					break
				}
				strPatterns[i] = s
			}
			if allStrings {
				patterns = strPatterns
			} else {
				return nil, fmt.Errorf("%w: patterns argument contained non-string elements", lang.ErrValidationTypeMismatch)
			}
		default:
			// This case should ideally be caught by spec validation
			return nil, fmt.Errorf("%w: internal error: patterns argument type mismatch (%T), expected slice", lang.ErrInternalTool, args[1])
		}
	}

	interpreter.Logger().Debug("[TOOL-GOLIST] Called", "target_dir_arg", targetDirRel, "patterns_arg", patterns)

	// Construct command arguments for 'go'
	cmdArgs := []string{"list", "-json"}
	cmdArgs = append(cmdArgs, patterns...)

	// Execute using helper, specifying the relative directory
	execResultIntf, execCmdErr := executeGoCommandHelper(interpreter, targetDirRel, cmdArgs...)
	if execCmdErr != nil {
		interpreter.Logger().Error("[TOOL-GOLIST] Execution helper failed", "error", execCmdErr)
		// Return empty list and a Go error for interpreter handling
		return []map[string]interface{}{}, fmt.Errorf("%w: execution helper failed: %w", lang.ErrInternalTool, execCmdErr)
	}
	execResultMap, ok := execResultIntf.(map[string]interface{})
	if !ok {
		interpreter.Logger().Error("[TOOL-GOLIST] Execution helper returned unexpected type", "type", fmt.Sprintf("%T", execResultIntf))
		return []map[string]interface{}{}, fmt.Errorf("%w: execution helper returned unexpected type %T", lang.ErrInternalTool, execResultIntf)
	}

	// Check if the 'go list' command itself reported failure via the helper's result map
	if success, _ := execResultMap["success"].(bool); !success {
		stderrStr, _ := execResultMap["stderr"].(string)
		exitCode, _ := execResultMap["exit_code"].(int64)
		errMsg := fmt.Sprintf("'go list' command failed (exit code %d). Stderr: %s", exitCode, stderrStr)
		interpreter.Logger().Error("[TOOL-GOLIST] %s", errMsg)
		// Return empty list, but no Go error (failure indicated by result map)
		// Alternatively, could return a Go error here too if desired.
		return []map[string]interface{}{}, nil
	}

	// Decode the JSON stream from stdout
	stdoutStr, _ := execResultMap["stdout"].(string)
	decoder := json.NewDecoder(strings.NewReader(stdoutStr))
	results := []map[string]interface{}{}

	for decoder.More() {
		var pkgInfo map[string]interface{}
		if decodeErr := decoder.Decode(&pkgInfo); decodeErr != nil {
			errMsg := fmt.Sprintf("failed to decode JSON object from 'go list -json' output: %v.", decodeErr)
			interpreter.Logger().Error("[TOOL-GOLIST] %s", errMsg)
			interpreter.Logger().Debug("[TOOL-GOLIST] Raw stdout causing decode error", "stdout", stdoutStr)
			// Return empty list and indicate error via Go error
			return []map[string]interface{}{}, fmt.Errorf("%w: %s", lang.ErrInternalTool, errMsg)
		}
		results = append(results, pkgInfo)
	}

	interpreter.Logger().Debug("[TOOL-GOLIST] Successfully executed and parsed", "package_count", len(results))
	return results, nil // Return the list of parsed package maps
}

// --- Internal Helper for executing Go commands within Sandbox ---

// executeGoCommandHelper runs a 'go' subcommand with given arguments inside the specified directory (relative to sandbox).
// It handles path validation and command execution, returning a map similar to toolExecuteCommand.
// Returns the result map and a Go-level error ONLY if the helper function itself fails (e.g., bad interpreter state, path resolution internal error).
// Command execution success/failure is indicated *within* the returned map.
func executeGoCommandHelper(interpreter tool.RunTime, targetDirRel string, goArgs ...string) (interface{}, error) {
	if interpreter == nil || interpreter.Logger() == nil {
		// Return nil map and a Go error for internal setup issues
		return nil, fmt.Errorf("executeGoCommandHelper: interpreter or logger is nil")
	}
	logger := interpreter.Logger()
	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		// Return nil map and a Go error for setup issues
		return nil, fmt.Errorf("%w: interpreter sandbox directory not set", lang.ErrInternalSecurity)
	}

	// --- Security: Validate and Resolve Directory ---
	absValidatedDir, pathErr := security.ResolveAndSecurePath(targetDirRel, sandboxRoot)
	if pathErr != nil {
		errMsg := fmt.Sprintf("Path validation failed for directory %q (relative to sandbox %q): %v", targetDirRel, sandboxRoot, pathErr)
		logger.Error("[GO-HELPER] %s", errMsg)
		// Path validation failure is a common case; return failure map, but NO Go-level error from the helper itself.
		return map[string]interface{}{"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false}, nil
	}
	// --- End Security ---

	cmd := "go"                               // Assuming 'go' is in PATH
	fullArgs := append([]string{}, goArgs...) // Copy args

	logger.Debug("[GO-HELPER] Executing command", "command", cmd, "args", fullArgs, "directory", absValidatedDir)

	// Prepare command execution
	cmdExec := exec.Command(cmd, fullArgs...)
	cmdExec.Dir = absValidatedDir // *** Run in the validated absolute directory ***
	var stdout, stderr bytes.Buffer
	cmdExec.Stdout = &stdout
	cmdExec.Stderr = &stderr

	// Run the command
	execErr := cmdExec.Run()
	stdoutStr := stdout.String()
	stderrStr := stderr.String()
	exitCode := 0
	success := true

	// Process execution result
	if execErr != nil {
		success = false
		if exitError, ok := execErr.(*exec.ExitError); ok {
			// Command started but exited non-zero
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			} else {
				exitCode = -1 // Unable to get specific exit code
				stderrStr += fmt.Sprintf("\n(Failed to get exit status: %v)", exitError.Sys())
			}
		} else {
			// Command failed to start (e.g., not found, permission error)
			exitCode = -1
			// Append the Go error to stderr for more context
			if stderrStr != "" && !strings.HasSuffix(stderrStr, "\n") {
				stderrStr += "\n"
			}
			stderrStr += fmt.Sprintf("Execution Error: %v", execErr)
		}
		logger.Warn("[GO-HELPER] Command failed", "command", cmd, "args", fullArgs, "dir", absValidatedDir, "exit_code", exitCode, "stderr", stderrStr)
	} else {
		logger.Debug("[GO-HELPER] Command successful", "command", cmd, "args", fullArgs, "dir", absValidatedDir, "exit_code", 0)
	}

	// Return results in a map
	resultMap := map[string]interface{}{
		"stdout":    stdoutStr,
		"stderr":    stderrStr,
		"exit_code": int64(exitCode),
		"success":   success,
	}
	// Return map, nil Go-level error (execution outcome is within the map)
	return resultMap, nil
}
