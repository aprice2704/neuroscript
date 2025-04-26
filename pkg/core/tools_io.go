// filename: pkg/core/tools_io.go
// UPDATED: Add TOOL.Log implementation and registration
package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const maxInputSize = 100 * 1024 // 100 KB limit as per spec v0.2

// --- IO.Input ---
// (toolIOInput implementation unchanged from fetch)
func toolIOInput(i *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: IO.Input internal error: expected 1 argument after validation, got %d", ErrInternalTool, len(args))
	}
	prompt, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: IO.Input internal error: argument 'prompt' not a string after validation, got %T", ErrInternalTool, args[0])
	}

	// IMPORTANT: Enforcement of agent mode restriction needs handling elsewhere.
	fmt.Print(prompt)

	limitedReader := io.LimitedReader{R: os.Stdin, N: maxInputSize}
	reader := bufio.NewReader(&limitedReader)
	input, err := reader.ReadString('\n')
	limitHit := limitedReader.N == 0

	if err != nil && err != io.EOF {
		errMap := map[string]interface{}{"input": nil, "error": fmt.Sprintf("Error reading input: %v", err)}
		return errMap, nil
	}
	if limitHit {
		errMsg := fmt.Sprintf("Input exceeded maximum size limit (%d bytes)", maxInputSize)
		errMap := map[string]interface{}{"input": nil, "error": errMsg}
		return errMap, nil
	}
	if err == io.EOF && input == "" {
		errMap := map[string]interface{}{"input": nil, "error": "EOF encountered reading input"}
		return errMap, nil
	}

	trimmedInput := strings.TrimSpace(input)
	resultMap := map[string]interface{}{"input": trimmedInput, "error": nil}
	return resultMap, nil
}

// --- Log ---

// toolLog implements the Log tool.
func toolLog(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Args: level (string), message (string)
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (level, message), got %d", ErrValidationArgCount, len(args))
	}
	levelStr, okL := args[0].(string)
	message, okM := args[1].(string)
	if !okL || !okM {
		return nil, fmt.Errorf("%w: expected string arguments for level and message", ErrValidationTypeMismatch)
	}

	// Use the interpreter's logger (ensured non-nil by Logger() method)
	logger := interpreter.logger

	// Prepend level to the message (simple approach)
	logPrefix := "[NS-LOG]" // Default prefix
	switch strings.ToLower(levelStr) {
	case "info":
		logPrefix = "[NS-INFO]"
	case "debug":
		logPrefix = "[NS-DEBUG]" // Note: Output depends on Go logger being configured for debug
	case "warn", "warning":
		logPrefix = "[NS-WARN]"
	case "error", "err":
		logPrefix = "[NS-ERROR]"
	default:
		// Log unknown levels as INFO? Or WARN? Let's use WARN.
		logPrefix = "[NS-WARN]"
		logger.Warn("[NS-WARN] Unknown log level '%s' used in TOOL.Log. Original message: %s", levelStr, message)
		// Continue logging with the original message under WARN
	}

	// Log using the interpreter's standard logger
	logger.Debug("%s %s", logPrefix, message)

	return nil, nil // No return value
}

// --- Registration ---

// registerLogTools registers logging-related tools.
func registerLogTools(registry *ToolRegistry) error {
	err := registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "Log",
			Description: "Writes a message to the application's internal log stream at a specified level.",
			Args: []ArgSpec{
				{Name: "level", Type: ArgTypeString, Required: true, Description: "Log level (e.g., 'Info', 'Debug', 'Warn', 'Error'). Case-insensitive."},
				{Name: "message", Type: ArgTypeString, Required: true, Description: "The message to log."},
			},
			ReturnType: ArgTypeAny, // No meaningful return value
		},
		Func: toolLog,
	})
	if err != nil {
		return fmt.Errorf("register Log: %w", err)
	}
	// Register other Log tools here if needed later
	return nil
}

// registerIOTools registers IO-related tools (like IO.Input) with the interpreter.
// UPDATED: Calls registerLogTools
func registerIOTools(registry *ToolRegistry) error {
	// Register IO.Input
	err := registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "IO.Input",
			Description: "Prompts the user for text input via the console. Enforces max input size. Not allowed in agent mode.",
			Args: []ArgSpec{
				{Name: "prompt", Type: ArgTypeString, Required: true, Description: "The text prompt to display to the user."},
			},
			ReturnType: ArgTypeAny, // Returns map {"input": string|null, "error": string|null}
		},
		Func: toolIOInput,
	})
	if err != nil {
		return fmt.Errorf("register IO.Input: %w", err)
	}

	// Register Log tool(s)
	err = registerLogTools(registry) // Call the new registration function
	if err != nil {
		return fmt.Errorf("registering Log tools: %w", err)
	}

	// Register other IO tools here if added later
	return nil
}
