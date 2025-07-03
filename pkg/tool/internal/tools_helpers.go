// filename: pkg/tool/tools_helpers.go
package internal

import (
	"bytes"
	"fmt"
	"os/exec" // Added regexp
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func MakeArgs(vals ...interface{}) []interface{} {
	if vals == nil {
		return []interface{}{}
	}
	return vals
}

// *** ADDED toolExec function definition ***
// toolExec executes an external command and returns combined stdout/stderr as a string,
// or an error if the command fails to run or exits non-zero.
// This is intended as an *internal* helper for other tools like Git tools.
func toolExec(interpreter tool.Runtime, cmdAndArgs ...string) (string, error) {
	if len(cmdAndArgs) == 0 {
		return "", fmt.Errorf("toolExec requires at least a command")
	}
	commandPath := cmdAndArgs[0]
	commandArgs := cmdAndArgs[1:]

	// Basic security check (can be enhanced)
	if strings.Contains(commandPath, "..") || strings.ContainsAny(commandPath, "|;&$><`\\") {
		errMsg := fmt.Sprintf("toolExec blocked suspicious command path: %q", commandPath)
		if interpreter.GetLogger != nil {
			interpreter.GetLogger().Error("[toolExec] %s", errMsg)
		}
		// Return error message and a wrapped ErrInternalTool or a specific execution error
		return errMsg, fmt.Errorf("%w: %s", lang.ErrInternalTool, errMsg)
	}

	if interpreter.GetLogger != nil {
		logArgs := make([]string, len(commandArgs))
		for i, arg := range commandArgs {
			if strings.Contains(arg, " ") {
				logArgs[i] = fmt.Sprintf("%q", arg) // Quote args with spaces
			} else {
				logArgs[i] = arg
			}
		}
		interpreter.GetLogger().Debug("[toolExec] Executing: %s %s", commandPath, strings.Join(logArgs, " "))
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
		if interpreter.GetLogger != nil {
			interpreter.GetLogger().Error("[toolExec] %s", errMsg)
		}
		// Return the combined output along with the error
		return combinedOutput, fmt.Errorf("%w: %s", lang.ErrInternalTool, errMsg)
	}

	// Command succeeded
	if interpreter.GetLogger != nil {
		interpreter.GetLogger().Debug("[toolExec] Command successful. Output:\n%s", combinedOutput)
	}
	return combinedOutput, nil
}

// --- END ADDED toolExec function definition ---

// getStringArg retrieves a required string argument from the args map.
func getStringArg(args map[string]interface{}, key string) (string, error) {
	val, ok := args[key]
	if !ok {
		return "", fmt.Errorf("missing required argument '%s'", key)
	}
	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("invalid type for argument '%s': expected string, got %T", key, val)
	}
	return strVal, nil
}

// makeArgMap is a convenience function to create a map[string]interface{}
// from key-value pairs, useful for constructing tool arguments programmatically.
// Example: makeArgMap("prompt", "hello", "count", 5)
func makeArgMap(kvPairs ...interface{}) (map[string]interface{}, error) {
	if len(kvPairs)%2 != 0 {
		return nil, fmt.Errorf("makeArgMap requires an even number of arguments (key-value pairs)")
	}
	args := make(map[string]interface{})
	for i := 0; i < len(kvPairs); i += 2 {
		key, ok := kvPairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("makeArgMap requires string keys, got %T at index %d", kvPairs[i], i)
		}
		args[key] = kvPairs[i+1]
	}
	return args, nil
}
