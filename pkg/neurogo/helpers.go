package neurogo

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// --- Helper Functions (loadToolListFromFile, executeAgentTool, formatToolResult, formatErrorResponse) ---
// Ensure these are implemented correctly as provided in previous steps.
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
func executeAgentTool(toolName string, args map[string]interface{}, interp *core.Interpreter) (interface{}, error) {
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
	interp.Logger().Printf("[AGENT TOOL] Calling %s with ordered args: %v", toolName, orderedArgs)
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
