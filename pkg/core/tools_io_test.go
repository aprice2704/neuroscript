// filename: pkg/core/tools_io_test.go
package core

import (
	"errors" // Import errors for Is
	"testing"
	// Testing stdin requires more complex setup (like redirecting os.Stdin)
	// or manual interaction. This test focuses solely on argument validation.
)

// Use the common ValidationTestCase struct (assuming it's defined in testing_helpers_test.go or locally)
// type ValidationTestCase struct {
//	Name          string
//	InputArgs     []interface{}
//	ExpectedError error // Expected error from ValidateAndConvertArgs
// }

func TestToolIOInputValidation(t *testing.T) {
	// Create a default test interpreter to access the tool registry and spec
	// Ignore the sandbox path as it's not needed for validation testing.
	interp, _ := newDefaultTestInterpreter(t)

	// --- Define Test Cases ---
	// Use standard Go types and the makeArgs helper
	testCases := []ValidationTestCase{
		{
			Name:          "Valid prompt (string)",
			InputArgs:     makeArgs("Enter name: "), // Use makeArgs and string literal
			ExpectedError: nil,                      // Validation should pass
		},
		{
			Name:          "No arguments",
			InputArgs:     makeArgs(), // Use makeArgs for empty slice
			ExpectedError: ErrValidationArgCount,
		},
		{
			Name:          "Too many arguments",
			InputArgs:     makeArgs("prompt", "extra"), // Use makeArgs
			ExpectedError: ErrValidationArgCount,
		},
		{
			Name:          "Incorrect argument type (number)",
			InputArgs:     makeArgs(int64(123)), // Use makeArgs and int64
			ExpectedError: ErrValidationTypeMismatch,
		},
		{
			Name:          "Incorrect argument type (bool)",
			InputArgs:     makeArgs(true), // Use makeArgs and bool
			ExpectedError: ErrValidationTypeMismatch,
		},
		{
			// Although nil is not the *correct* type (string expected),
			// ValidateAndConvertArgs should catch it as a required arg being nil first.
			Name:          "Incorrect argument type (nil)",
			InputArgs:     makeArgs(nil), // Use makeArgs and nil
			ExpectedError: ErrValidationRequiredArgNil,
		},
	}

	// --- Get Tool Spec ---
	// Use the correct tool name as registered (assuming "IO.Input")
	toolName := "IO.Input"
	toolImpl, found := interp.ToolRegistry().GetTool(toolName)
	if !found {
		t.Fatalf("FATAL: Tool %q not found in registry during test setup.", toolName)
	}
	spec := toolImpl.Spec // Get the ToolSpec

	// --- Run Validation Tests ---
	for _, tc := range testCases {
		// Use tc.Name directly from the struct field
		t.Run(tc.Name, func(t *testing.T) {
			// Call ValidateAndConvertArgs using the spec and the raw input args
			// We don't need the convertedArgs, just the validation error.
			_, err := ValidateAndConvertArgs(spec, tc.InputArgs)

			// Use tc.ExpectedError directly from the struct field
			if tc.ExpectedError != nil {
				// Expecting a specific validation error
				if err == nil {
					t.Errorf("ValidateAndConvertArgs() expected error [%v], but got nil", tc.ExpectedError)
				} else if !errors.Is(err, tc.ExpectedError) {
					// Check if the returned error wraps the expected error type
					t.Errorf("ValidateAndConvertArgs() expected error type [%T], but got type [%T]: %v", tc.ExpectedError, err, err)
				}
			} else {
				// Expecting validation success (nil error)
				if err != nil {
					t.Errorf("ValidateAndConvertArgs() unexpected validation error: %v", err)
				}
			}
		})
	}
}

// --- Manual Test Guidance (Remains the same) ---
// To test the full functionality including stdin reading, size limits, and EOF:
// 1. Build the `neurogo` executable.
// 2. Create a simple NeuroScript file (e.g., test_input.ns.txt):
//    DEFINE PROCEDURE Main()
//    COMMENT:
//    PURPOSE: Test IO.Input
//    INPUTS: none
//    OUTPUT: none
//    ALGORITHM:
//    - Prompt user
//    - Emit result
//    ENDCOMMENT
//    SET result_map = CALL IO.Input("Enter something: ")
//    EMIT "Input Result Map: ", result_map
//    END
// 3. Run: `neurogo run ./test_input.ns.txt`
// 4. Interact with the prompt:
//    - Type text and press Enter. Check output map.
//    - Press Ctrl+D (for EOF). Check output map (should show EOF error).
//    - Paste >100KB of text. Check output map (should show size limit error).
