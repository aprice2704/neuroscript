// pkg/core/tools_io.go
package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const maxInputSize = 100 * 1024 // 100 KB limit as per spec v0.2

// toolIOInput implements the IO.Input tool.
// Spec: docs/ns/tools/io_input.md (or similar)
// REFACTORED: Accepts []interface{} and returns (interface{}, error)
//
//	Relies on prior validation by ValidateAndConvertArgs.
func toolIOInput(i *Interpreter, args []interface{}) (interface{}, error) {
	// --- Validation (Type Assertion) ---
	// ValidateAndConvertArgs (called before this) should have ensured:
	// - Exactly 1 argument was provided.
	// - The argument (args[0]) was converted to a string.
	// We perform type assertions here as a safeguard and to get the concrete type.

	if len(args) != 1 {
		// This indicates an internal error if ValidateAndConvertArgs worked correctly.
		return nil, fmt.Errorf("%w: IO.Input internal error: expected 1 argument after validation, got %d", ErrInternalTool, len(args))
	}

	prompt, ok := args[0].(string)
	if !ok {
		// This indicates an internal error if ValidateAndConvertArgs worked correctly.
		return nil, fmt.Errorf("%w: IO.Input internal error: argument 'prompt' not a string after validation, got %T", ErrInternalTool, args[0])
	}

	// --- Security Note ---
	// IMPORTANT: Enforcement of agent mode restriction still needs to be handled
	// in the Interpreter's tool dispatch logic *before* calling this function.
	// Example check (would go in interpreter logic):
	// if i.agentMode { return nil, errors.New("IO.Input tool not allowed in agent mode") }

	// --- Core Logic ---
	// Use EMIT equivalent (interpreter's stdout) for the prompt
	fmt.Print(prompt) // Display prompt without newline

	// Use a LimitedReader to enforce the size limit
	limitedReader := io.LimitedReader{R: os.Stdin, N: maxInputSize}
	reader := bufio.NewReader(&limitedReader) // New reader for each call

	input, err := reader.ReadString('\n')
	limitHit := limitedReader.N == 0 // Check if limit was hit *after* reading

	// Handle errors and limit condition
	if err != nil && err != io.EOF { // Handle non-EOF errors first
		errMap := map[string]interface{}{ // Use map[string]interface{}
			"input": nil, // Use Go nil
			"error": fmt.Sprintf("Error reading input: %v", err),
		}
		// Should this return a Go error or just the map?
		// The spec seems to imply it always returns the map structure.
		return errMap, nil // Return error map
	}

	if limitHit { // Check if the limit was reached
		errMsg := fmt.Sprintf("Input exceeded maximum size limit (%d bytes)", maxInputSize)
		// Try to drain the rest of the line from the *original* stdin buffer
		// This is tricky and might not be perfectly reliable.
		// go func() {
		//	bufio.NewReader(os.Stdin).ReadString('\n')
		// }()
		errMap := map[string]interface{}{ // Use map[string]interface{}
			"input": nil, // Use Go nil
			"error": errMsg,
		}
		return errMap, nil // Return error map
	}

	if err == io.EOF && input == "" { // Genuine EOF before any input (e.g., Ctrl+D on empty line)
		errMap := map[string]interface{}{ // Use map[string]interface{}
			"input": nil, // Use Go nil
			"error": "EOF encountered reading input",
		}
		return errMap, nil // Return error map
	}

	// --- Success Case ---
	trimmedInput := strings.TrimSpace(input)

	// Return success map using map[string]interface{}
	resultMap := map[string]interface{}{
		"input": trimmedInput, // Return string
		"error": nil,          // Use Go nil for no error
	}
	return resultMap, nil
}

// --- Registration ---

// registerIOTools registers IO-related tools (like IO.Input) with the interpreter.
// REFACTORED: Added implementation and return type specification.
func registerIOTools(registry *ToolRegistry) error {
	err := registry.RegisterTool(ToolImplementation{ // Capture potential error
		Spec: ToolSpec{
			Name:        "IO.Input",
			Description: "Prompts the user for text input via the console. Enforces max input size. Not allowed in agent mode.",
			Args: []ArgSpec{
				{Name: "prompt", Type: ArgTypeString, Required: true, Description: "The text prompt to display to the user."},
			},
			// Returns a map {"input": string|null, "error": string|null}
			ReturnType: ArgTypeAny,
		},
		Func: toolIOInput,
	})
	if err != nil {
		return fmt.Errorf("register IO.Input: %w", err)
	}
	// Register other IO tools here if added later
	return nil
}
