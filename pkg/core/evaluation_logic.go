// pkg/core/evaluation_logic.go
package core

import (
	"fmt"  // Keep for error checks potentially
	"math" // Keep for evaluateFunctionCall
	// strconv is likely not needed here anymore
)

// --- Evaluation Logic for Operations ---

// evaluateUnaryOp performs prefix unary operations (NOT, -).
// (Keeping this function here for now)
func evaluateUnaryOp(op string, operand interface{}) (interface{}, error) {
	switch op {
	case "NOT":
		return !isTruthy(operand), nil
	case "-": // Numeric negation
		iVal, isInt := toInt64(operand)
		if isInt {
			return -iVal, nil
		}
		fVal, isFloat := toFloat64(operand)
		if isFloat {
			return -fVal, nil
		}
		return nil, fmt.Errorf("unary operator '-' requires a numeric operand, got %T", operand)
	default:
		return nil, fmt.Errorf("unsupported unary operator '%s'", op)
	}
}

// evaluateBinaryOp performs infix binary operations by dispatching to helper functions.
func evaluateBinaryOp(left, right interface{}, op string) (interface{}, error) {
	// --- Dispatch based on operator type ---

	switch op {
	// Logical (Handle short-circuiting here)
	case "AND":
		leftBool := isTruthy(left)
		if !leftBool {
			return false, nil // Short-circuit
		}
		// Evaluate right only if left is true
		return isTruthy(right), nil
	case "OR":
		leftBool := isTruthy(left)
		if leftBool {
			return true, nil // Short-circuit
		}
		// Evaluate right only if left is false
		return isTruthy(right), nil

	// Nil handling before dispatching to typed operators
	case "==", "!=":
		leftIsNil := left == nil
		rightIsNil := right == nil
		if leftIsNil || rightIsNil {
			if op == "==" {
				return leftIsNil && rightIsNil, nil
			} else { // op == "!="
				return !(leftIsNil && rightIsNil), nil
			}
		}
		// If neither is nil, fall through to comparison helper
		result, err := performComparison(left, right, op)
		if err != nil {
			return nil, fmt.Errorf("failed comparison for %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	// Comparison (non-nil handled by helper)
	case "<", ">", "<=", ">=":
		// Check for nil *before* calling comparison helper for these ops
		if left == nil || right == nil {
			return nil, fmt.Errorf("operator '%s' cannot be applied to nil operand", op)
		}
		result, err := performComparison(left, right, op)
		if err != nil {
			return nil, fmt.Errorf("failed comparison for %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	// Arithmetic (includes '+')
	case "+":
		// Special '+' handler handles strings vs numeric automatically
		result, err := performStringConcatOrNumericAdd(left, right)
		if err != nil {
			// Add context if error occurs
			return nil, fmt.Errorf("failed operation for %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	case "-", "*", "/", "%", "**":
		// Check for nil before calling arithmetic helper
		if left == nil || right == nil {
			return nil, fmt.Errorf("operator '%s' cannot be applied to nil operand", op)
		}
		result, err := performArithmetic(left, right, op)
		if err != nil {
			// Add context if error occurs
			return nil, fmt.Errorf("failed arithmetic for %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	// Bitwise
	case "&", "|", "^":
		// Check for nil before calling bitwise helper
		if left == nil || right == nil {
			return nil, fmt.Errorf("operator '%s' cannot be applied to nil operand", op)
		}
		result, err := performBitwise(left, right, op)
		if err != nil {
			// Add context if error occurs
			return nil, fmt.Errorf("failed bitwise operation for %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported binary operator '%s'", op)
	}
}

// evaluateFunctionCall handles built-in function calls.
// (Keeping this function here for now)
func evaluateFunctionCall(funcName string, args []interface{}) (interface{}, error) {
	// Helper to check arg count and types (using helpers from evaluation_helpers.go)
	checkArgs := func(expectedCount int, argTypes ...string) error {
		if len(args) != expectedCount {
			return fmt.Errorf("function %s expects %d argument(s), got %d", funcName, expectedCount, len(args))
		}
		for i, expectedType := range argTypes {
			switch expectedType {
			case "float":
				if _, ok := toFloat64(args[i]); !ok {
					return fmt.Errorf("function %s argument %d expects a number, got %T", funcName, i+1, args[i])
				}
			case "int":
				if _, ok := toInt64(args[i]); !ok {
					return fmt.Errorf("function %s argument %d expects an integer, got %T", funcName, i+1, args[i])
				}
			// Add "string", "bool" etc. if needed for future functions
			default:
				return fmt.Errorf("internal error: unknown type check '%s' for function %s", expectedType, funcName)
			}
		}
		return nil
	}

	// Math functions mostly expect float64
	switch funcName {
	case "LN":
		if err := checkArgs(1, "float"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		if fVal <= 0 {
			return nil, fmt.Errorf("LN requires positive arg, got %v", fVal)
		}
		return math.Log(fVal), nil
	case "LOG": // Assume base 10
		if err := checkArgs(1, "float"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		if fVal <= 0 {
			return nil, fmt.Errorf("LOG requires positive arg, got %v", fVal)
		}
		return math.Log10(fVal), nil
	case "SIN":
		if err := checkArgs(1, "float"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		return math.Sin(fVal), nil
	case "COS":
		if err := checkArgs(1, "float"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		return math.Cos(fVal), nil
	case "TAN":
		if err := checkArgs(1, "float"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		return math.Tan(fVal), nil
	case "ASIN":
		if err := checkArgs(1, "float"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		if fVal < -1 || fVal > 1 {
			return nil, fmt.Errorf("ASIN requires arg between -1 and 1, got %v", fVal)
		}
		return math.Asin(fVal), nil
	case "ACOS":
		if err := checkArgs(1, "float"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		if fVal < -1 || fVal > 1 {
			return nil, fmt.Errorf("ACOS requires arg between -1 and 1, got %v", fVal)
		}
		return math.Acos(fVal), nil
	case "ATAN":
		if err := checkArgs(1, "float"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		return math.Atan(fVal), nil
	// Add other functions here...
	default:
		return nil, fmt.Errorf("unknown built-in function '%s'", funcName)
	}
}
