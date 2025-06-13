// NeuroScript Version: 0.4.1
// File version: 11
// Purpose: Implements short-circuiting for boolean AND/OR and delegates other ops.
// filename: pkg/core/evaluation_logic.go
// nlines: 130
// risk_rating: HIGH

package core

import (
	"fmt"
	"math"
	"strings"
)

// evaluateUnaryOp handles unary operations like NOT, -, and ~.
func (i *Interpreter) evaluateUnaryOp(op string, operand interface{}) (Value, error) {
	if operand == nil && (op == "-" || op == "~") {
		return nil, fmt.Errorf("%w: unary operator '%s'", ErrNilOperand, op)
	}
	switch strings.ToLower(op) {
	case "not":
		// 'not' inverts the truthiness of its operand.
		return BoolValue{Value: !isTruthy(operand)}, nil
	case "-":
		num, ok := ToNumeric(operand)
		if !ok {
			return nil, fmt.Errorf("%w: unary operator '-' needs number, got %s", ErrInvalidOperandTypeNumeric, TypeOf(operand))
		}
		return NumberValue{Value: -num.Value}, nil
	case "~":
		iVal, isInt := toInt64(operand)
		if !isInt {
			return nil, fmt.Errorf("%w: unary operator '~' needs integer, got %s", ErrInvalidOperandTypeInteger, TypeOf(operand))
		}
		return NumberValue{Value: float64(^iVal)}, nil
	case "no":
		return BoolValue{Value: isZeroValue(operand)}, nil
	case "some":
		return BoolValue{Value: !isZeroValue(operand)}, nil
	default:
		return nil, fmt.Errorf("%w: '%s'", ErrUnsupportedOperator, op)
	}
}

// toFuzzy attempts to coerce a value to a FuzzyValue.
// It succeeds for existing FuzzyValues and booleans.
func toFuzzy(v interface{}) (FuzzyValue, bool) {
	if f, ok := v.(FuzzyValue); ok {
		return f, true
	}
	if b, ok := v.(BoolValue); ok {
		if b.Value {
			return NewFuzzyValue(1.0), true
		}
		return NewFuzzyValue(0.0), true
	}
	return FuzzyValue{}, false
}

// evaluateBinaryOp performs infix binary operations.
// It now handles boolean short-circuiting for 'and'/'or' before delegating.
func (i *Interpreter) evaluateBinaryOp(left, right interface{}, op string) (Value, error) {
	opLower := strings.ToLower(op)

	// --- SHORT-CIRCUITING LOGIC ---
	// The evaluator calls this function for all binary ops. We must handle
	// boolean short-circuiting here before attempting other logic.
	switch opLower {
	case "and":
		if !isTruthy(left) {
			return BoolValue{Value: false}, nil
		}
		// If left is truthy, the result is the truthiness of the right side.
		return BoolValue{Value: isTruthy(right)}, nil
	case "or":
		if isTruthy(left) {
			return BoolValue{Value: true}, nil
		}
		// If left is falsy, the result is the truthiness of the right side.
		return BoolValue{Value: isTruthy(right)}, nil
	}

	// --- FUZZY LOGIC (for AND/OR if not short-circuited by boolean logic) ---
	// This part is now less likely to be hit for standard booleans but remains for FuzzyValue types.
	if opLower == "and" || opLower == "or" {
		leftF, leftIsFuzzy := toFuzzy(left)
		rightF, rightIsFuzzy := toFuzzy(right)

		if leftIsFuzzy && rightIsFuzzy {
			if opLower == "and" {
				return NewFuzzyValue(math.Min(leftF.μ, rightF.μ)), nil
			}
			return NewFuzzyValue(math.Max(leftF.μ, rightF.μ)), nil
		}
	}

	// --- STANDARD OPERATOR DELEGATION ---
	switch opLower {
	case "==", "!=", "<", ">", "<=", ">=":
		// performComparison now returns a Value
		return performComparison(left, right, op)
	case "+":
		// performStringConcatOrNumericAdd now returns a Value
		return performStringConcatOrNumericAdd(left, right)
	case "-", "*", "/", "%", "**":
		// performArithmetic now returns a Value
		return performArithmetic(left, right, op)
	case "&", "|", "^":
		// performBitwise now returns a Value
		return performBitwise(left, right, op)
	default:
		// This should not be reached if the parser is correct.
		return nil, fmt.Errorf("%w: '%s'", ErrUnsupportedOperator, op)
	}
}
