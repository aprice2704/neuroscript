// filename: pkg/core/tools_math.go
package core

import (
	"fmt"
)

// registerMathTools adds math-related tools.
// *** MODIFIED: Returns error ***
func registerMathTools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		{
			Spec: ToolSpec{
				Name:        "Add", // Keep TOOL. prefix consistent with NeuroScript CALL syntax? No, use base name for registry key.
				Description: "Calculates the sum of two numbers (integers or decimals). Strings convertible to numbers are accepted.",
				Args: []ArgSpec{
					{Name: "num1", Type: ArgTypeFloat, Required: true, Description: "The first number (or numeric string) to add."},
					{Name: "num2", Type: ArgTypeFloat, Required: true, Description: "The second number (or numeric string) to add."},
				},
				ReturnType: ArgTypeFloat,
			},
			Func: toolAdd,
		},
		{
			Spec: ToolSpec{
				Name:        "Subtract",
				Description: "Calculates the difference between two numbers (num1 - num2). Strings convertible to numbers are accepted.",
				Args: []ArgSpec{
					{Name: "num1", Type: ArgTypeFloat, Required: true, Description: "The number to subtract from."},
					{Name: "num2", Type: ArgTypeFloat, Required: true, Description: "The number to subtract."},
				},
				ReturnType: ArgTypeFloat,
			},
			Func: toolSubtract,
		},
		{
			Spec: ToolSpec{
				Name:        "Multiply",
				Description: "Calculates the product of two numbers. Strings convertible to numbers are accepted.",
				Args: []ArgSpec{
					{Name: "num1", Type: ArgTypeFloat, Required: true, Description: "The first number."},
					{Name: "num2", Type: ArgTypeFloat, Required: true, Description: "The second number."},
				},
				ReturnType: ArgTypeFloat,
			},
			Func: toolMultiply,
		},
		{
			Spec: ToolSpec{
				Name:        "Divide",
				Description: "Calculates the division of two numbers (num1 / num2). Returns float. Handles division by zero.",
				Args: []ArgSpec{
					{Name: "num1", Type: ArgTypeFloat, Required: true, Description: "The dividend."},
					{Name: "num2", Type: ArgTypeFloat, Required: true, Description: "The divisor."},
				},
				ReturnType: ArgTypeFloat,
			},
			Func: toolDivide,
		},
		{
			Spec: ToolSpec{
				Name:        "Modulo",
				Description: "Calculates the modulo (remainder) of two integers (num1 % num2). Handles division by zero.",
				Args: []ArgSpec{
					{Name: "num1", Type: ArgTypeInt, Required: true, Description: "The dividend (must be integer)."},
					{Name: "num2", Type: ArgTypeInt, Required: true, Description: "The divisor (must be integer)."},
				},
				ReturnType: ArgTypeInt,
			},
			Func: toolModulo,
		},
	}
	for _, tool := range tools {
		// *** Check error from RegisterTool ***
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register Math tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil // Success
}

// --- Implementations (Unchanged) ---

func toolAdd(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	num1, ok1 := args[0].(float64)
	num2, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("TOOL.Add internal error: arguments were not converted to float64. Got %T and %T", args[0], args[1])
	}
	result := num1 + num2
	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL Add] Calculated %v + %v = %v", num1, num2, result)
	}
	return result, nil
}

func toolSubtract(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	num1, ok1 := args[0].(float64)
	num2, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("TOOL.Subtract internal error: arguments not float64. Got %T and %T", args[0], args[1])
	}
	result := num1 - num2
	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL Subtract] Calculated %v - %v = %v", num1, num2, result)
	}
	return result, nil
}

func toolMultiply(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	num1, ok1 := args[0].(float64)
	num2, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("TOOL.Multiply internal error: arguments not float64. Got %T and %T", args[0], args[1])
	}
	result := num1 * num2
	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL Multiply] Calculated %v * %v = %v", num1, num2, result)
	}
	return result, nil
}

func toolDivide(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	num1, ok1 := args[0].(float64)
	num2, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("TOOL.Divide internal error: arguments not float64. Got %T and %T", args[0], args[1])
	}
	if num2 == 0.0 {
		return nil, fmt.Errorf("%w: division by zero in TOOL.Divide", ErrInternalTool) // Wrap internal tool error
	}
	result := num1 / num2
	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL Divide] Calculated %v / %v = %v", num1, num2, result)
	}
	return result, nil
}

func toolModulo(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	num1, ok1 := args[0].(int64)
	num2, ok2 := args[1].(int64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("TOOL.Modulo internal error: arguments not int64. Got %T and %T", args[0], args[1])
	}
	if num2 == 0 {
		return nil, fmt.Errorf("%w: division by zero in TOOL.Modulo", ErrInternalTool) // Wrap internal tool error
	}
	result := num1 % num2
	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL Modulo] Calculated %v %% %v = %v", num1, num2, result)
	}
	return result, nil
}
