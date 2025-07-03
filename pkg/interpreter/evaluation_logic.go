// NeuroScript Version: 0.5.2
// File version: 17
// Purpose: Exported EvaluateBinaryOp to make it accessible to external test packages.
// filename: pkg/interpreter/evaluation_logic.go
// nlines: 120
// risk_rating: HIGH

package interpreter

import (
	"fmt"
	"math"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// evaluateUnaryOp handles unary operations like NOT, -, and ~.
func (i *Interpreter) EvaluateUnaryOp(op string, operand lang.Value) (lang.Value, error) {
	if operand == nil && (op == "-" || op == "~") {
		return nil, fmt.Errorf("%w: unary operator '%s'", lang.ErrNilOperand, op)
	}

	switch strings.ToLower(op) {
	case "not":
		if fv, ok := operand.(lang.FuzzyValue); ok {
			return lang.NewFuzzyValue(1.0 - fv.GetValue()), nil
		}
		return lang.BoolValue{Value: !lang.IsTruthy(operand)}, nil

	case "-":
		num, ok := lang.ToNumeric(operand)
		if !ok {
			return nil, fmt.Errorf("%w: unary operator '-' needs number, got %s", lang.ErrInvalidOperandTypeNumeric, lang.TypeOf(operand))
		}
		return lang.NumberValue{Value: -num.Value}, nil

	case "~":
		iVal, isInt := lang.ToInt64(operand)
		if !isInt {
			return nil, fmt.Errorf("%w: unary operator '~' needs integer, got %s", lang.ErrInvalidOperandTypeInteger, lang.TypeOf(operand))
		}
		return lang.NumberValue{Value: float64(^iVal)}, nil

	case "no":
		return lang.BoolValue{Value: lang.IsZeroValue(operand)}, nil
	case "some":
		return lang.BoolValue{Value: !lang.IsZeroValue(operand)}, nil
	default:
		return nil, fmt.Errorf("%w: '%s'", lang.ErrUnsupportedOperator, op)
	}
}

// toFuzzy attempts to coerce a value to a FuzzyValue.
func toFuzzy(v lang.Value) (lang.FuzzyValue, bool) {
	if f, ok := v.(lang.FuzzyValue); ok {
		return f, true
	}
	if b, ok := v.(lang.BoolValue); ok {
		if b.Value {
			return lang.NewFuzzyValue(1.0), true
		}
		return lang.NewFuzzyValue(0.0), true
	}
	return lang.FuzzyValue{}, false
}

// EvaluateBinaryOp performs infix binary operations.
// FIX: Capitalized the function name to export it.
func (i *Interpreter) EvaluateBinaryOp(left, right lang.Value, op string) (lang.Value, error) {
	opLower := strings.ToLower(op)

	_, leftIsFuzzy := left.(lang.FuzzyValue)
	_, rightIsFuzzy := right.(lang.FuzzyValue)

	if (opLower == "and" || opLower == "or") && (leftIsFuzzy || rightIsFuzzy) {
		leftF, canConvertToFuzzyLeft := toFuzzy(left)
		rightF, canConvertToFuzzyRight := toFuzzy(right)

		if canConvertToFuzzyLeft && canConvertToFuzzyRight {
			if opLower == "and" {
				return lang.NewFuzzyValue(math.Min(leftF.GetValue(), rightF.GetValue())), nil
			}
			return lang.NewFuzzyValue(math.Max(leftF.GetValue(), rightF.GetValue())), nil
		}
	}

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
