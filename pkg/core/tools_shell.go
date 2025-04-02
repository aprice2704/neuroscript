// pkg/core/tools_shell.go
package core

import (
	"bytes"
	"fmt"
	"go/format" // Use go/format for GoFmt
	"os"
	"os/exec"
	"path/filepath" // Needed for secureFilePath
	"strings"
	"syscall"
)

// registerShellTools adds shell execution and Go-related tools to the registry.
func registerShellTools(registry *ToolRegistry) {
	// Execute arbitrary commands (Use with caution!)
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "ExecuteCommand",
			Description: "Executes an arbitrary shell command with arguments. Returns map {stdout, stderr, exit_code, success}.",
			Args: []ArgSpec{
				{Name: "command", Type: ArgTypeString, Required: true, Description: "The command or path to execute."},
				{Name: "args_list", Type: ArgTypeSliceAny, Required: true, Description: "A list of arguments (converted to strings)."},
			},
			ReturnType: ArgTypeAny, // Returns a map
		},
		Func: toolExecuteCommand,
	})

	// *** MODIFIED GoBuild SPECIFICATION ***
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GoBuild",
			Description: "Runs 'go build [target]' in the current directory. Defaults to './...' if no target provided. Returns map {stdout, stderr, exit_code, success}.",
			Args: []ArgSpec{
				// Added optional target argument
				{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional build target (e.g., relative path like './pkg/core' or a file like 'test_files/simple_test.go'). Defaults to './...'"},
			},
			ReturnType: ArgTypeAny, // Returns a map
		},
		Func: toolGoBuild, // Use updated function
	})
	// *** END MODIFICATION ***

	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GoTest",
			Description: "Runs 'go test ./...' in the current directory. Returns map {stdout, stderr, exit_code, success}.",
			Args:        []ArgSpec{}, // No arguments
			ReturnType:  ArgTypeAny,  // Returns a map
		},
		Func: toolGoTest,
	})

	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GoFmt",
			Description: "Formats Go source code provided as a string using go/format. Returns map {formatted_content, error, success}.",
			Args: []ArgSpec{
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The Go source code content as a string."},
			},
			ReturnType: ArgTypeAny, // Returns a map
		},
		Func: toolGoFmt,
	})

	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GoModTidy",
			Description: "Runs 'go mod tidy' in the current directory. Returns map {stdout, stderr, exit_code, success}.",
			Args:        []ArgSpec{}, // No arguments
			ReturnType:  ArgTypeAny,  // Returns a map
		},
		Func: toolGoModTidy,
	})

}

// toolExecuteCommand implementation (no changes needed here)
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
		return map[string]interface{}{
			"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false,
		}, nil
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

	resultMap := map[string]interface{}{
		"stdout": stdoutStr, "stderr": stderrStr, "exit_code": int64(exitCode), "success": success,
	}
	return resultMap, nil
}

// *** MODIFIED toolGoBuild IMPLEMENTATION ***
// toolGoBuild runs 'go build [target]' where target is an optional argument.
func toolGoBuild(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	buildTarget := "./..." // Default target

	// Check if the optional argument was provided
	if len(args) > 0 {
		targetArg, ok := args[0].(string)
		if !ok {
			// This should ideally be caught by validation, but double-check
			return nil, fmt.Errorf("TOOL.GoBuild internal error: optional target argument was not a string, got %T", args[0])
		}

		// Validate the provided path using secureFilePath relative to CWD
		cwd, errWd := os.Getwd()
		if errWd != nil {
			return nil, fmt.Errorf("GoBuild failed to get working directory: %w", errWd)
		}
		// Allow '.' as a valid target (meaning current directory)
		if targetArg != "." {
			// We just need to check it's *within* the CWD, secureFilePath gives absolute
			// But 'go build' often works best with relative paths. Let's secure it first
			// then try to make it relative again for the command.
			_, secErr := secureFilePath(targetArg, cwd)
			if secErr != nil {
				// Return path error in the result map
				errMsg := fmt.Sprintf("GoBuild path error for target '%s': %s", targetArg, secErr.Error())
				return map[string]interface{}{
					"stdout": "", "stderr": errMsg, "exit_code": int64(-1), "success": false,
				}, nil
			}
			// If secure, use the validated relative path provided by the user
			buildTarget = filepath.Clean(targetArg)
		} else {
			buildTarget = "." // Explicitly use "." if provided
		}
	}

	cmd := "go"
	cmdArgs := []interface{}{"build", buildTarget} // Args for ExecuteCommand must be []interface{}
	executeArgs := []interface{}{cmd, cmdArgs}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GoBuild (executing: go build %s)", buildTarget)
	}
	// Execute the command via the ExecuteCommand tool implementation
	return toolExecuteCommand(interpreter, executeArgs)
}

// *** END MODIFICATION ***

// toolGoTest implementation (no changes)
func toolGoTest(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	cmd := "go"
	cmdArgs := []interface{}{"test", "./..."}
	executeArgs := []interface{}{cmd, cmdArgs}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GoTest (executing: go test ./...)")
	}
	return toolExecuteCommand(interpreter, executeArgs)
}

// toolGoFmt implementation (no changes)
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

// toolGoModTidy implementation (no changes)
func toolGoModTidy(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	cmd := "go"
	cmdArgs := []interface{}{"mod", "tidy"}
	executeArgs := []interface{}{cmd, cmdArgs}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GoModTidy (executing: go mod tidy)")
	}
	return toolExecuteCommand(interpreter, executeArgs)
}
