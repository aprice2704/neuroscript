// filename: pkg/core/tools_math.go
package core

import (
	"fmt"
)

// --- Implementations (Unchanged) ---

func toolAdd(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	num1, ok1 := args[0].(float64)
	num2, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("TOOL.Add internal error: arguments were not converted to float64. Got %T and %T", args[0], args[1])
	}
	result := num1 + num2
	if interpreter.logger != nil {
		interpreter.logger.Info("Tool: Add] Calculated %v + %v = %v", num1, num2, result)
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
		interpreter.logger.Info("Tool: Subtract] Calculated %v - %v = %v", num1, num2, result)
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
		interpreter.logger.Info("Tool: Multiply] Calculated %v * %v = %v", num1, num2, result)
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
		interpreter.logger.Info("Tool: Divide] Calculated %v / %v = %v", num1, num2, result)
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
		interpreter.logger.Info("Tool: Modulo] Calculated %v %% %v = %v", num1, num2, result)
	}
	return result, nil
}
