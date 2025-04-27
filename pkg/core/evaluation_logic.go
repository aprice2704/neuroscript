// pkg/core/evaluation_logic.go
package core

import (
	// Need errors package
	"fmt"
	"math" // Keep for evaluateFunctionCall
	"reflect"
	// strconv is likely not needed here anymore
)

// --- Evaluation Logic for Operations ---

// NEW Helper: isZeroValue checks if a value is the zero value for its type.
func isZeroValue(val interface{}) bool {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}

// evaluateUnaryOp performs prefix unary operations (not, -, no, some).
func evaluateUnaryOp(op string, operand interface{}) (interface{}, error) {
	switch op {
	case "not":
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
		return nil, fmt.Errorf("%w: unary operator '-' needs number, got %T", ErrInvalidOperandTypeNumeric, operand)
	case "no": // Check if operand IS the zero value for its type
		return isZeroValue(operand), nil
	case "some": // Check if operand is NOT the zero value for its type
		return !isZeroValue(operand), nil
	default:
		return nil, fmt.Errorf("unsupported unary operator '%s'", op)
	}
}

// evaluateBinaryOp performs infix binary operations (dispatching).
func evaluateBinaryOp(left, right interface{}, op string) (interface{}, error) {
	switch op {
	// Logical (Short-circuit handled by caller evaluateExpression)
	case "and":
		fallthrough // Use non-short-circuit helper
	case "or":
		return performLogical(left, right, op) // Pass to helper

	// Comparison (handle nil checks first)
	case "==", "!=":
		leftIsNil := left == nil
		rightIsNil := right == nil
		// Handle comparisons involving nil directly
		if leftIsNil || rightIsNil {
			isEqual := leftIsNil && rightIsNil // True only if both are nil
			if op == "==" {
				return isEqual, nil
			} else {
				return !isEqual, nil
			} // op == "!="
		}
		// If neither is nil, use the helper
		result, err := performComparison(left, right, op)
		if err != nil {
			return nil, fmt.Errorf("compare %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	case "<", ">", "<=", ">=":
		if left == nil || right == nil {
			return nil, fmt.Errorf("operator '%s' needs non-nil operands", op)
		}
		result, err := performComparison(left, right, op)
		if err != nil {
			return nil, fmt.Errorf("compare %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	// Arithmetic (includes '+')
	case "+":
		result, err := performStringConcatOrNumericAdd(left, right) // Handles string vs numeric
		if err != nil {
			return nil, fmt.Errorf("op %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	case "-", "*", "/", "%", "**":
		if left == nil || right == nil {
			return nil, fmt.Errorf("operator '%s' needs non-nil operands", op)
		}
		result, err := performArithmetic(left, right, op)
		if err != nil {
			return nil, fmt.Errorf("arithmetic %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	// Bitwise
	case "&", "|", "^":
		if left == nil || right == nil {
			return nil, fmt.Errorf("operator '%s' needs non-nil operands", op)
		}
		result, err := performBitwise(left, right, op)
		if err != nil {
			return nil, fmt.Errorf("bitwise %T %s %T: %w", left, op, right, err)
		}
		return result, nil // Added missing return

	default:
		return nil, fmt.Errorf("unsupported binary operator '%s'", op)
	}
}

// evaluateFunctionCall handles built-in function calls.
func evaluateFunctionCall(funcName string, args []interface{}) (interface{}, error) {
	// Helper to check arg count and types
	checkArgs := func(expectedCount int, argTypes ...string) error {
		if len(args) != expectedCount {
			return fmt.Errorf("func %s expects %d arg(s), got %d", funcName, expectedCount, len(args))
		}
		for i, expectedType := range argTypes {
			switch expectedType {
			case "float":
				if _, ok := toFloat64(args[i]); !ok {
					return fmt.Errorf("%w: func %s arg %d needs number, got %T", ErrInvalidFunctionArgument, funcName, i+1, args[i])
				}
			case "int":
				if _, ok := toInt64(args[i]); !ok {
					return fmt.Errorf("%w: func %s arg %d needs integer, got %T", ErrInvalidFunctionArgument, funcName, i+1, args[i])
				}
			default:
				return fmt.Errorf("internal: unknown type check '%s' for func %s", expectedType, funcName)
			}
		}
		return nil
	}

	// Built-in math functions
	switch funcName {
	case "LN":
		if err := checkArgs(1, "float"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		if fVal <= 0 {
			return nil, fmt.Errorf("%w: LN needs positive arg, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Log(fVal), nil
	case "LOG":
		if err := checkArgs(1, "float"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		if fVal <= 0 {
			return nil, fmt.Errorf("%w: LOG needs positive arg, got %v", ErrInvalidFunctionArgument, fVal)
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
			return nil, fmt.Errorf("%w: ASIN needs arg between -1 and 1, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Asin(fVal), nil
	case "ACOS":
		if err := checkArgs(1, "float"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		if fVal < -1 || fVal > 1 {
			return nil, fmt.Errorf("%w: ACOS needs arg between -1 and 1, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Acos(fVal), nil
	case "ATAN":
		if err := checkArgs(1, "float"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		return math.Atan(fVal), nil
	// Add cases for askAI, askHuman, askComputer if they are treated as built-in functions here later
	default:
		// If not a known built-in, it might be a user-defined procedure or tool handled elsewhere.
		// For now, assume only built-ins are handled here.
		return nil, fmt.Errorf("unknown built-in function '%s'", funcName)
	}
}
