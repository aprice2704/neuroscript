// NeuroScript Version: 0.3.1
// File version: 0.1.1
// Return ErrDivisionByZero sentinel directly.
// nlines: 65
// risk_rating: MEDIUM
// filename: pkg/tool/math/tools_math.go
package math

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- Implementations ---

func toolAdd(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	num1, ok1 := args[0].(float64)
	num2, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		// This indicates a failure in validation/coercion, likely an internal error
		return nil, fmt.Errorf("%w: arguments were not converted to float64. Got %T and %T", lang.ErrInternalTool, args[0], args[1])
	}
	result := num1 + num2
	if interpreter.logger != nil {
		interpreter.logger.Debug("Tool: Add] Calculated %v + %v = %v", num1, num2, result)
	}
	return result, nil
}

func toolSubtract(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	num1, ok1 := args[0].(float64)
	num2, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("%w: arguments not float64. Got %T and %T", lang.ErrInternalTool, args[0], args[1])
	}
	result := num1 - num2
	if interpreter.logger != nil {
		interpreter.logger.Debug("Tool: Subtract] Calculated %v - %v = %v", num1, num2, result)
	}
	return result, nil
}

func toolMultiply(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	num1, ok1 := args[0].(float64)
	num2, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("%w: arguments not float64. Got %T and %T", lang.ErrInternalTool, args[0], args[1])
	}
	result := num1 * num2
	if interpreter.logger != nil {
		interpreter.logger.Debug("Tool: Multiply] Calculated %v * %v = %v", num1, num2, result)
	}
	return result, nil
}

func toolDivide(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	num1, ok1 := args[0].(float64)
	num2, ok2 := args[1].(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("%w: arguments not float64. Got %T and %T", lang.ErrInternalTool, args[0], args[1])
	}
	if num2 == 0.0 {
		// Corrected: Return the specific sentinel error directly
		return nil, lang.ErrDivisionByZero
	}
	result := num1 / num2
	if interpreter.logger != nil {
		interpreter.logger.Debug("Tool: Divide] Calculated %v / %v = %v", num1, num2, result)
	}
	return result, nil
}

func toolModulo(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	num1, ok1 := args[0].(int64)
	num2, ok2 := args[1].(int64)
	if !ok1 || !ok2 {
		// Modulo requires integers, ensure validation handles this.
		// If validation passes non-ints, this is an internal error.
		return nil, fmt.Errorf("%w: arguments not int64. Got %T and %T", lang.ErrInternalTool, args[0], args[1])
	}
	if num2 == 0 {
		// Corrected: Return the specific sentinel error directly
		return nil, lang.ErrDivisionByZero
	}
	result := num1 % num2
	if interpreter.logger != nil {
		interpreter.logger.Debug("Tool: Modulo] Calculated %v %% %v = %v", num1, num2, result)
	}
	// Modulo result should remain int64
	return result, nil
}
