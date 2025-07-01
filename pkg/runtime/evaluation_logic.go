// NeuroScript Version: 0.4.1
// File version: 13
// Purpose: Corrected binary 'and'/'or' operators to properly handle mixed FuzzyValue and BoolValue operands, preventing incorrect boolean short-circuiting.
// filename: pkg/core/evaluation_logic.go

package runtime

import (
	"fmt"
	"math"
	"strings"
)

// evaluateUnaryOp handles unary operations like NOT, -, and ~.
// It now correctly handles the 'not' operator for FuzzyValue types.
func (i *Interpreter) evaluateUnaryOp(op string, operand Value) (Value, error) {
	if operand == nil && (op == "-" || op == "~") {
		return nil, fmt.Errorf("%w: unary operator '%s'", ErrNilOperand, op)
	}

	switch strings.ToLower(op) {
	case "not":
		// The 'not' operator has special behavior for FuzzyValue.
		if fv, ok := operand.(FuzzyValue); ok {
			// The logical NOT of a fuzzy value is 1 minus its membership degree.
			return NewFuzzyValue(1.0 - fv.μ), nil
		}
		// For all other types, 'not' inverts their standard truthiness.
		return BoolValue{Value: !IsTruthy(operand)}, nil

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
func toFuzzy(v Value) (FuzzyValue, bool) {
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
func (i *Interpreter) evaluateBinaryOp(left, right Value, op string) (Value, error) {
	opLower := strings.ToLower(op)

	// --- FUZZY LOGIC PRIORITY CHECK ---
	// If either operand is a FuzzyValue and the operator is 'and' or 'or',
	// we must use fuzzy logic, not boolean short-circuiting.
	_, leftIsFuzzy := left.(FuzzyValue)
	_, rightIsFuzzy := right.(FuzzyValue)

	if (opLower == "and" || opLower == "or") && (leftIsFuzzy || rightIsFuzzy) {
		leftF, canConvertToFuzzyLeft := toFuzzy(left)
		rightF, canConvertToFuzzyRight := toFuzzy(right)

		if canConvertToFuzzyLeft && canConvertToFuzzyRight {
			if opLower == "and" {
				// Fuzzy AND is the minimum of the two values.
				return NewFuzzyValue(math.Min(leftF.μ, rightF.μ)), nil
			}
			// Fuzzy OR is the maximum of the two values.
			return NewFuzzyValue(math.Max(leftF.μ, rightF.μ)), nil
		}
	}

	// --- BOOLEAN SHORT-CIRCUITING LOGIC ---
	// This will now only be reached for 'and'/'or' if neither operand is fuzzy.
	switch opLower {
	case "and":
		if !IsTruthy(left) {
			return BoolValue{Value: false}, nil
		}
		return BoolValue{Value: IsTruthy(right)}, nil
	case "or":
		if IsTruthy(left) {
			return BoolValue{Value: true}, nil
		}
		return BoolValue{Value: IsTruthy(right)}, nil
	}

	// --- STANDARD OPERATOR DELEGATION ---
	switch opLower {
	case "==", "!=", "<", ">", "<=", ">=":
		return performComparison(left, right, op)
	case "+":
		return performStringConcatOrNumericAdd(left, right)
	case "-", "*", "/", "%", "**":
		return performArithmetic(left, right, op)
	case "&", "|", "^":
		return performBitwise(left, right, op)
	default:
		return nil, fmt.Errorf("%w: '%s'", ErrUnsupportedOperator, op)
	}
}
