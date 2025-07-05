// NeuroScript Version: 0.3.0
// File version: 0.0.3
// Corrected interpreter type from core.Interpreter to interpreter.Interpreter.
// filename: pkg/neurogo/helpers.go
package neurogo

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interpreter" // Corrected import
)

// --- Helper Functions (loadToolListFromFile, executeAgentTool, formatToolResult, formatErrorResponse) ---
func loadToolListFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", filePath, err)
	}
	defer file.Close()
	var tools []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "--") {
			continue
		}
		tools = append(tools, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}
	return tools, nil
}
func executeAgentTool(toolName string, args map[string]interface{}, interp *interpreter.Interpreter) (interface{}, error) {
	toolImpl, found := interp.ToolRegistry().GetTool(toolName)
	if !found {
		return nil, fmt.Errorf("tool %s not found", toolName)
	}
	orderedArgs := make([]interface{}, len(toolImpl.Spec.Args))
	for i, argSpec := range toolImpl.Spec.Args {
		val, exists := args[argSpec.Name]
		if !exists && argSpec.Required {
			return nil, fmt.Errorf("missing required arg %s for %s", argSpec.Name, toolName)
		}
		orderedArgs[i] = val
	}
	interp.Logger().Debug("[AGENT TOOL] Calling %s with ordered args: %v", toolName, orderedArgs)
	return toolImpl.Func(interp, orderedArgs)
}
func formatToolResult(toolOutput interface{}, execErr error) map[string]interface{} {
	r := make(map[string]interface{})
	if execErr != nil {
		r["success"] = false
		r["error"] = execErr.Error()
		if toolOutput != nil {
			r["partial_output"] = fmt.Sprintf("%v", toolOutput)
		}
	} else {
		r["success"] = true
		r["result"] = toolOutput
	}
	return r
}
func formatErrorResponse(err error) map[string]interface{} {
	s := "unknown error"
	if err != nil {
		s = err.Error()
	}
	return map[string]interface{}{"success": false, "error": s}
}

// base36ToIndex converts a base36 string (0-9, a-z, case-insensitive) to a 0-indexed integer.
func base36ToIndex(s string) (int, error) {
	s = strings.ToLower(s)
	if len(s) == 0 {
		return -1, fmt.Errorf("empty base36 string")
	}
	if len(s) == 1 {
		char := s[0]
		if char >= '0' && char <= '9' {
			return int(char - '0'), nil
		}
		if char >= 'a' && char <= 'z' {
			return int(char-'a') + 10, nil
		}
		return -1, fmt.Errorf("invalid single base36 character: %c", char)
	}
	i, err := strconv.ParseInt(s, 36, 64)
	if err != nil {
		return -1, fmt.Errorf("invalid base36 string: '%s', error: %w", s, err)
	}
	return int(i), nil
}

// indexToBase36 converts a 0-indexed integer to a base36 string (0-z, then 10, 11, ...)
func indexToBase36(idx int) string {
	if idx < 0 {
		return "?"
	}
	if idx < 10 {
		return strconv.Itoa(idx)
	}
	if idx < 36 {
		return string(rune('a' + (idx - 10)))
	}
	return strings.ToLower(strconv.FormatInt(int64(idx), 36))
}
