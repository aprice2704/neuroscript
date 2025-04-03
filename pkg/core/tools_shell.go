// pkg/core/tools_shell.go
package core

import (
	"bytes" // Import encoding/json
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// registerShellTools adds shell execution and Go-related tools to the registry.
func registerShellTools(registry *ToolRegistry) {
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "ExecuteCommand", Description: "Executes an arbitrary shell command...", Args: []ArgSpec{{Name: "command", Type: ArgTypeString, Required: true}, {Name: "args_list", Type: ArgTypeSliceAny, Required: true}}, ReturnType: ArgTypeAny}, Func: toolExecuteCommand,
	})

	// Keep GoBuild as is for now, but maybe make target required or default to '.'?
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "GoBuild", Description: "Runs 'go build [target]'...", Args: []ArgSpec{{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional build target. Defaults to './...'"}}, ReturnType: ArgTypeAny}, Func: toolGoBuild,
	})

	// *** ADDING GoCheck TOOL ***
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GoCheck",
			Description: "Checks Go code validity using 'go list -e -json <target>'. Returns map {check_success: bool, error_details: string}.",
			Args: []ArgSpec{
				{Name: "target", Type: ArgTypeString, Required: true, Description: "Target Go package path or file path (e.g., './pkg/core', 'test_files/simple_test.go')."},
			},
			ReturnType: ArgTypeAny, // Returns a map
		},
		Func: toolGoCheck, // Use new function
	})
	// *** END ADDITION ***

	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "GoTest", Description: "Runs 'go test ./...'...", Args: []ArgSpec{}, ReturnType: ArgTypeAny}, Func: toolGoTest,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "GoFmt", Description: "Formats Go source code provided as a string...", Args: []ArgSpec{{Name: "content", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeAny}, Func: toolGoFmt,
	})
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{Name: "GoModTidy", Description: "Runs 'go mod tidy'...", Args: []ArgSpec{}, ReturnType: ArgTypeAny}, Func: toolGoModTidy,
	})
}

// toolExecuteCommand implementation (remains the same)
func toolExecuteCommand(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	commandPath := args[0].(string)
	rawCmdArgs := args[1].([]interface{})
	commandArgs := make([]string, len(rawCmdArgs))
	for i, v := range rawCmdArgs {
		commandArgs[i] = fmt.Sprintf("%v", v)
	}

	if strings.Contains(commandPath, "..") || strings.ContainsAny(commandPath, "|;&$><`\\") {
		errMsg := fmt.Sprintf("ExecuteCommand blocked suspicious command path: %q", commandPath)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[ERROR] %s", errMsg)
		}
		return map[string]interface{}{"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false}, nil
	}

	if interpreter.logger != nil {
		logArgs := make([]string, len(commandArgs))
		for i, arg := range commandArgs {
			if strings.Contains(arg, " ") {
				logArgs[i] = fmt.Sprintf("%q", arg)
			} else {
				logArgs[i] = arg
			}
		}
		interpreter.logger.Printf("[DEBUG-INTERP]      Executing Command: %s %s", commandPath, strings.Join(logArgs, " "))
	}

	cmd := exec.Command(commandPath, commandArgs...)
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
		if interpreter.logger != nil {
			interpreter.logger.Printf("[DEBUG-INTERP]        Command failed. Exit Code: %d, Stderr: %q", exitCode, stderrStr)
		}
	} else {
		if interpreter.logger != nil {
			interpreter.logger.Printf("[DEBUG-INTERP]        Command finished successfully. Exit Code: 0, Stdout: %q", stdoutStr)
		}
	}
	resultMap := map[string]interface{}{"stdout": stdoutStr, "stderr": stderrStr, "exit_code": int64(exitCode), "success": success}
	return resultMap, nil
}

// *** NEW toolGoCheck IMPLEMENTATION ***
func toolGoCheck(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures 1 string argument
	targetPath := args[0].(string)

	// Validate the path (ensure it's relative and within bounds)
	cwd, errWd := os.Getwd()
	if errWd != nil {
		return nil, fmt.Errorf("GoCheck failed to get working directory: %w", errWd)
	}
	// Use secureFilePath - allow '.' as target
	cleanTargetPath := "."
	if targetPath != "." {
		_, secErr := secureFilePath(targetPath, cwd)
		if secErr != nil {
			errMsg := fmt.Sprintf("GoCheck path error for target '%s': %s", targetPath, secErr.Error())
			return map[string]interface{}{"check_success": false, "error_details": errMsg}, nil
		}
		cleanTargetPath = filepath.Clean(targetPath) // Use validated relative path
	}

	// Prepare arguments for toolExecuteCommand
	cmd := "go"
	// Use -e to report errors but continue, -json for structured output
	cmdArgs := []interface{}{"list", "-e", "-json", cleanTargetPath}
	executeArgs := []interface{}{cmd, cmdArgs}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GoCheck (executing: go list -e -json %s)", cleanTargetPath)
	}

	// Execute the command
	execResultIntf, execCmdErr := toolExecuteCommand(interpreter, executeArgs)
	if execCmdErr != nil {
		// Should not happen if toolExecuteCommand handles errors properly
		return nil, fmt.Errorf("GoCheck internal error calling ExecuteCommand: %w", execCmdErr)
	}
	execResultMap, ok := execResultIntf.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("GoCheck internal error: ExecuteCommand returned unexpected type %T", execResultIntf)
	}

	// --- Analyze the result ---
	checkSuccess := true // Assume success unless error found
	errorDetails := ""
	execSuccess := execResultMap["success"].(bool)
	execStderr := execResultMap["stderr"].(string)
	execStdout := execResultMap["stdout"].(string)

	// 1. Check if the 'go list' command itself failed unexpectedly
	if !execSuccess {
		checkSuccess = false
		errorDetails = fmt.Sprintf("go list command failed (exit code %v). Stderr: %s", execResultMap["exit_code"], execStderr)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[DEBUG-INTERP]        GoCheck failed: %s", errorDetails)
		}
	} else {
		// 2. If command ran, check stdout for JSON error fields
		// The output might be multiple JSON objects concatenated.
		// Simple check: look for `"Error":` indicating *any* package load error.
		// More robust: Properly decode the JSON stream.
		// Let's start simple.
		if strings.Contains(execStdout, `"Error":`) {
			checkSuccess = false
			// Try to extract a snippet around the error for better detail
			errIdx := strings.Index(execStdout, `"Error":`)
			snippetStart := errIdx - 30
			if snippetStart < 0 {
				snippetStart = 0
			}
			snippetEnd := errIdx + 100
			if snippetEnd > len(execStdout) {
				snippetEnd = len(execStdout)
			}
			errorDetails = fmt.Sprintf("Found errors in 'go list -e -json' output. Snippet near first error: ...%s...", execStdout[snippetStart:snippetEnd])

			// Also include stderr, just in case go list prints warnings there
			if execStderr != "" {
				errorDetails += "\nStderr from go list: " + execStderr
			}

			if interpreter.logger != nil {
				interpreter.logger.Printf("[DEBUG-INTERP]        GoCheck found errors in JSON output.")
			}
		} else {
			// Command succeeded and no "Error": field found in stdout
			checkSuccess = true
			errorDetails = "" // Explicitly empty
			if interpreter.logger != nil {
				interpreter.logger.Printf("[DEBUG-INTERP]        GoCheck successful (no errors found in 'go list' output).")
			}
		}
	}

	// Return the check result map
	checkResultMap := map[string]interface{}{
		"check_success": checkSuccess,
		"error_details": errorDetails,
	}
	return checkResultMap, nil
}

// *** END NEW IMPLEMENTATION ***

// toolGoBuild implementation (remains the same for now)
func toolGoBuild(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	buildTarget := "./..."
	if len(args) > 0 {
		targetArg, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("TOOL.GoBuild internal error: optional target argument was not a string, got %T", args[0])
		}
		cwd, errWd := os.Getwd()
		if errWd != nil {
			return nil, fmt.Errorf("GoBuild failed to get working directory: %w", errWd)
		}
		if targetArg != "." {
			_, secErr := secureFilePath(targetArg, cwd)
			if secErr != nil {
				errMsg := fmt.Sprintf("GoBuild path error for target '%s': %s", targetArg, secErr.Error())
				return map[string]interface{}{"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false}, nil
			}
			buildTarget = filepath.Clean(targetArg)
		} else {
			buildTarget = "."
		}
	}
	cmd := "go"
	cmdArgs := []interface{}{"build", buildTarget}
	executeArgs := []interface{}{cmd, cmdArgs}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GoBuild (executing: go build %s)", buildTarget)
	}
	return toolExecuteCommand(interpreter, executeArgs)
}

// toolGoTest implementation (remains the same)
func toolGoTest(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	cmd := "go"
	cmdArgs := []interface{}{"test", "./..."}
	executeArgs := []interface{}{cmd, cmdArgs}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GoTest (executing: go test ./...)")
	}
	return toolExecuteCommand(interpreter, executeArgs)
}

// toolGoFmt implementation (remains the same)
func toolGoFmt(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	content := args[0].(string)
	srcBytes := []byte(content)
	if interpreter.logger != nil {
		logSnippet := content
		if len(logSnippet) > 100 {
			logSnippet = logSnippet[:100] + "..."
		}
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GoFmt on input content (snippet): %q", logSnippet)
	}
	formattedBytes, fmtErr := format.Source(srcBytes)
	success := fmtErr == nil
	errorString := ""
	formattedContent := ""
	if success {
		formattedContent = string(formattedBytes)
		if interpreter.logger != nil {
			if !bytes.Equal(srcBytes, formattedBytes) {
				interpreter.logger.Printf("[DEBUG-INTERP]        GoFmt successful (content changed).")
			} else {
				interpreter.logger.Printf("[DEBUG-INTERP]        GoFmt successful (no changes needed).")
			}
		}
	} else {
		errorString = fmtErr.Error()
		formattedContent = content
		if interpreter.logger != nil {
			interpreter.logger.Printf("[DEBUG-INTERP]        GoFmt failed. Error: %q", errorString)
		}
	}
	resultMap := map[string]interface{}{"formatted_content": formattedContent, "error": errorString, "success": success}
	return resultMap, nil
}

// toolGoModTidy implementation (remains the same)
func toolGoModTidy(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	cmd := "go"
	cmdArgs := []interface{}{"mod", "tidy"}
	executeArgs := []interface{}{cmd, cmdArgs}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GoModTidy (executing: go mod tidy)")
	}
	return toolExecuteCommand(interpreter, executeArgs)
}
