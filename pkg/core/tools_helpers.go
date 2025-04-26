// filename: pkg/core/tools_helpers.go
package core

import (
	"bytes"
	"fmt"
	"os/exec" // Added regexp
	"strings"
)

func MakeArgs(vals ...interface{}) []interface{} {
	if vals == nil {
		return []interface{}{}
	}
	return vals
}

// runGitCommand executes a git command with the given arguments.
func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		quotedArgs := make([]string, len(args))
		for i, arg := range args {
			if strings.Contains(arg, " ") {
				quotedArgs[i] = fmt.Sprintf("%q", arg)
			} else {
				quotedArgs[i] = arg
			}
		}
		return fmt.Errorf("git command 'git %s' failed: %v\nStderr: %s", strings.Join(quotedArgs, " "), err, stderr.String())
	}
	return nil
}

// *** ADDED toolExec function definition ***
// toolExec executes an external command and returns combined stdout/stderr as a string,
// or an error if the command fails to run or exits non-zero.
// This is intended as an *internal* helper for other tools like Git tools.
func toolExec(interpreter *Interpreter, cmdAndArgs ...string) (string, error) {
	if len(cmdAndArgs) == 0 {
		return "", fmt.Errorf("toolExec requires at least a command")
	}
	commandPath := cmdAndArgs[0]
	commandArgs := cmdAndArgs[1:]

	// Basic security check (can be enhanced)
	if strings.Contains(commandPath, "..") || strings.ContainsAny(commandPath, "|;&$><`\\") {
		errMsg := fmt.Sprintf("toolExec blocked suspicious command path: %q", commandPath)
		if interpreter.logger != nil {
			interpreter.logger.Error("[toolExec] %s", errMsg)
		}
		// Return error message and a wrapped ErrInternalTool or a specific execution error
		return errMsg, fmt.Errorf("%w: %s", ErrInternalTool, errMsg)
	}

	if interpreter.logger != nil {
		logArgs := make([]string, len(commandArgs))
		for i, arg := range commandArgs {
			if strings.Contains(arg, " ") {
				logArgs[i] = fmt.Sprintf("%q", arg) // Quote args with spaces
			} else {
				logArgs[i] = arg
			}
		}
		interpreter.logger.Debug("[toolExec] Executing: %s %s", commandPath, strings.Join(logArgs, " "))
	}

	cmd := exec.Command(commandPath, commandArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	execErr := cmd.Run()

	stdoutStr := stdout.String()
	stderrStr := stderr.String()
	combinedOutput := stdoutStr + stderrStr // Combine outputs

	if execErr != nil {
		// Command failed (non-zero exit or execution error)
		errMsg := fmt.Sprintf("command '%s %s' failed with exit error: %v. Output:\n%s",
			commandPath, strings.Join(commandArgs, " "), execErr, combinedOutput)
		if interpreter.logger != nil {
			interpreter.logger.Error("[toolExec] %s", errMsg)
		}
		// Return the combined output along with the error
		return combinedOutput, fmt.Errorf("%w: %s", ErrInternalTool, errMsg)
	}

	// Command succeeded
	if interpreter.logger != nil {
		interpreter.logger.Debug("[toolExec] Command successful. Output:\n%s", combinedOutput)
	}
	return combinedOutput, nil
}

// --- END ADDED toolExec function definition ---
