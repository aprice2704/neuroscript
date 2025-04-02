// pkg/core/tools_shell.go
package core

import (
	"bytes"
	"fmt"
	"go/format" // Use go/format

	// "io" // No longer needed
	"os"
	"os/exec"
	"strings"
	"syscall"
	// "time" // No longer needed
	// "github.com/bluekeyes/go-gitdiff/gitdiff" // No longer needed
)

// --- toolExecuteCommand, toolGoBuild, toolGoTest ---
func toolExecuteCommand(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... implementation ... */
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.ExecuteCommand internal error: expected 2 args (command, args_list), got %d", len(args))
	}
	commandPath, ok := args[0].(string)
	if !ok || commandPath == "" {
		return nil, fmt.Errorf("TOOL.ExecuteCommand internal error: command path must be a non-empty string, got %T", args[0])
	}
	var commandArgs []string
	switch argList := args[1].(type) {
	case []string:
		commandArgs = argList
	case []interface{}:
		strArgs := make([]string, len(argList))
		for i, v := range argList {
			strArgs[i] = fmt.Sprintf("%v", v)
		}
		commandArgs = strArgs
	default:
		return nil, fmt.Errorf("TOOL.ExecuteCommand internal error: arguments must be a slice (slice_string or slice_any), got %T", args[1])
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
	if execErr != nil {
		if exitError, ok := execErr.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			} else {
				exitCode = -1
			}
		} else {
			exitCode = -1
			if stderrStr != "" {
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
	resultMap := map[string]interface{}{"stdout": stdoutStr, "stderr": stderrStr, "exit_code": int64(exitCode), "success": exitCode == 0}
	return resultMap, nil
}
func toolGoBuild(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... implementation ... */
	if len(args) != 0 {
		return nil, fmt.Errorf("TOOL.GoBuild internal error: expected 0 arguments, got %d", len(args))
	}
	cmd := "go"
	cmdArgs := []interface{}{"build", "./..."}
	executeArgs := []interface{}{cmd, cmdArgs}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GoBuild (executing: %s %s)", cmd, strings.Join([]string{"build", "./..."}, " "))
	}
	return toolExecuteCommand(interpreter, executeArgs)
}
func toolGoTest(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... implementation ... */
	if len(args) != 0 {
		return nil, fmt.Errorf("TOOL.GoTest internal error: expected 0 arguments, got %d", len(args))
	}
	cmd := "go"
	cmdArgs := []interface{}{"test", "./..."}
	executeArgs := []interface{}{cmd, cmdArgs}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GoTest (executing: %s %s)", cmd, strings.Join([]string{"test", "./..."}, " "))
	}
	return toolExecuteCommand(interpreter, executeArgs)
}

// --- Apply Patch Tool (REMOVED) ---
// func toolApplyPatch(interpreter *Interpreter, args []interface{}) (interface{}, error) { ... }

// --- Go Format Tool (Using go/format library) ---
func toolGoFmt(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.GoFmt internal error: expected 1 arg (filepath), got %d", len(args))
	}
	filePath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.GoFmt internal error: filepath must be a string, got %T", args[0])
	}

	cwd, errWd := os.Getwd()
	if errWd != nil {
		return nil, fmt.Errorf("GoFmt failed to get working directory: %w", errWd)
	}
	// secureFilePath requires relative path now
	absPath, secErr := secureFilePath(filePath, cwd)
	if secErr != nil {
		return map[string]interface{}{"formatted_content": "", "error": fmt.Sprintf("GoFmt path error: %s", secErr.Error()), "success": false}, nil
	}

	srcBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		return map[string]interface{}{"formatted_content": string(srcBytes), "error": fmt.Sprintf("GoFmt read error: %s", readErr.Error()), "success": false}, nil
	}

	formattedBytes, fmtErr := format.Source(srcBytes) // Use Go library

	success := fmtErr == nil
	errorString := ""
	formattedContent := ""
	if success {
		formattedContent = string(formattedBytes)
		if interpreter.logger != nil {
			interpreter.logger.Printf("[DEBUG-INTERP]      GoFmt successful for %s", filePath)
		}
	} else {
		errorString = fmtErr.Error()
		formattedContent = string(srcBytes) // Return original content on format error
		if interpreter.logger != nil {
			interpreter.logger.Printf("[DEBUG-INTERP]      GoFmt failed for %s. Error: %q", filePath, errorString)
		}
	}
	resultMap := map[string]interface{}{"formatted_content": formattedContent, "error": errorString, "success": success}
	return resultMap, nil
}

// --- Go Mod Tidy Tool ---
func toolGoModTidy(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("TOOL.GoModTidy internal error: expected 0 arguments, got %d", len(args))
	}
	cmd := "go"
	cmdArgs := []interface{}{"mod", "tidy"}
	executeArgs := []interface{}{cmd, cmdArgs}
	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GoModTidy (executing: go mod tidy)")
	}
	// Run relative to CWD implicitly via toolExecuteCommand
	return toolExecuteCommand(interpreter, executeArgs)
}
