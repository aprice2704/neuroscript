// NeuroScript Version: 0.4.1
// File version: 13
// Purpose: Corrected binary 'and'/'or' operators to properly handle mixed FuzzyValue and BoolValue operands, preventing incorrect boolean short-circuiting.
// filename: pkg/interpreter/internal/eval/evaluation_logic.go

package eval

import (
	"fmt"
	"math"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// evaluateUnaryOp handles unary operations like NOT, -, and ~.
// It now correctly handles the 'not' operator for FuzzyValue types.
func (i *Interpreter) evaluateUnaryOp(op string, operand Value) (Value, error) {
	if operand == nil && (op == "-" || op == "~") {
		return nil, fmt.Errorf("%w: unary operator '%s'", lang.ErrNilOperand, op)
	}

	switch strings.ToLower(op) {
	case "not":
		// The 'not' operator has special behavior for FuzzyValue.
		if fv, ok := operand.(FuzzyValue); ok {
			// The logical NOT of a fuzzy value is 1 minus its membership degree.
			return lang.NewFuzzyValue(1.0 - fv.μ), nil
		}
		// For all other types, 'not' inverts their standard truthiness.
		return lang.BoolValue{Value: !lang.IsTruthy(operand)}, nil

	case "-":
		num, ok := lang.ToNumeric(operand)
		if !ok {
			return nil, fmt.Errorf("%w: unary operator '-' needs number, got %s", lang.ErrInvalidOperandTypeNumeric, lang.TypeOf(operand))
		}
		return lang.NumberValue{Value: -num.Value}, nil

	case "~":
		iVal, isInt := lang.toInt64(operand)
		if !isInt {
			return nil, fmt.Errorf("%w: unary operator '~' needs integer, got %s", lang.ErrInvalidOperandTypeInteger, lang.TypeOf(operand))
		}
		return lang.NumberValue{Value: float64(^iVal)}, nil

	case "no":
		return lang.BoolValue{Value: lang.isZeroValue(operand)}, nil
	case "some":
		return lang.BoolValue{Value: !lang.isZeroValue(operand)}, nil
	default:
		return nil, fmt.Errorf("%w: '%s'", lang.ErrUnsupportedOperator, op)
	}
}

// toFuzzy attempts to coerce a value to a FuzzyValue.
func toFuzzy(v lang.Value) (lang.FuzzyValue, bool) {
	if f, ok := v.(FuzzyValue); ok {
		return f, true
	}
	if b, ok := v.(BoolValue); ok {
		if b.Value {
			return lang.NewFuzzyValue(1.0), true
		}
		return lang.NewFuzzyValue(0.0), true
	}
	return lang.FuzzyValue{}, false
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
				return lang.NewFuzzyValue(math.Min(leftF.μ, rightF.μ)), nil
			}
			// Fuzzy OR is the maximum of the two values.
			return lang.NewFuzzyValue(math.Max(leftF.μ, rightF.μ)), nil
		}
	}

	// --- BOOLEAN SHORT-CIRCUITING LOGIC ---
	// This will now only be reached for 'and'/'or' if neither operand is fuzzy.
	switch opLower {
	case "and":
		if !lang.IsTruthy(left) {
			return lang.BoolValue{Value: false}, nil
		}
		return lang.BoolValue{Value: lang.IsTruthy(right)}, nil
	case "or":
		if lang.IsTruthy(left) {
			return lang.BoolValue{Value: true}, nil
		}
		return lang.BoolValue{Value: lang.IsTruthy(right)}, nil
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
		return nil, fmt.Errorf("%w: '%s'", lang.ErrUnsupportedOperator, op)
	}
}