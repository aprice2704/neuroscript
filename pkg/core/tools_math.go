// filename: pkg/core/tools_math.go
package core

import (
	"fmt"
)

// registerMathTools adds math-related tools.
func registerMathTools(registry *ToolRegistry) {
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "TOOL.Add", // Use TOOL. prefix for consistency
			Description: "Calculates the sum of two numbers (integers or decimals). Strings convertible to numbers are accepted.",
			Args: []ArgSpec{
				// *** CHANGED Type from ArgTypeAny to ArgTypeFloat ***
				{Name: "num1", Type: ArgTypeFloat, Required: true, Description: "The first number (or numeric string) to add."},
				{Name: "num2", Type: ArgTypeFloat, Required: true, Description: "The second number (or numeric string) to add."},
			},
			// Return type changed to ArgTypeFloat for consistency
			ReturnType: ArgTypeFloat,
		},
		Func: toolAdd,
	})
	// Add other math tools here later if needed
}

// toolAdd performs addition on two validated numeric arguments.
func toolAdd(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation layer (ValidateAndConvertArgs) now ensures args are float64
	// due to the updated ToolSpec using ArgTypeFloat.
	num1, ok1 := args[0].(float64)
	num2, ok2 := args[1].(float64)

	if !ok1 || !ok2 {
		// This check is now mainly a safeguard against internal errors if validation logic failed.
		return nil, fmt.Errorf("TOOL.Add internal error: arguments were not converted to float64. Got %T and %T", args[0], args[1])
	}

	result := num1 + num2 // Simple float64 addition

	if interpreter.logger != nil {
		// Log the float values used in the calculation
		interpreter.logger.Printf("[TOOL Add] Calculated %v + %v = %v", num1, num2, result)
	}
	return result, nil
}
