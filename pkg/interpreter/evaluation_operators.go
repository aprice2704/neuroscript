// NeuroScript Version: 0.5.2
// File version: 15
// Purpose: Corrected type handling in operator functions to properly manage the Value wrapper boundary, resolving interface assignment errors.
// filename: pkg/interpreter/evaluation_operators.go
// nlines: 205
// risk_rating: HIGH

package interpreter

import (
	"fmt"
	"math"
	"reflect"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// typeErrorForOp creates a standardized error for invalid operations.
func typeErrorForOp(op string, left, right lang.Value) error {
	return fmt.Errorf("operator '%s' cannot be applied to types %s and %s: %w", op, lang.TypeOf(left), lang.TypeOf(right), lang.ErrInvalidOperandType)
}

// performArithmetic handles operators: -, *, /, %, **
func performArithmetic(left, right lang.Value, op string) (lang.Value, error) {
	leftNum, leftOk := lang.ToNumeric(left)
	rightNum, rightOk := lang.ToNumeric(right)

	if !leftOk || !rightOk {
		return nil, typeErrorForOp(op, left, right)
	}

	leftF := leftNum.Value
	rightF := rightNum.Value

	if op == "%" {
		if leftF != math.Trunc(leftF) || rightF != math.Trunc(rightF) {
			return nil, fmt.Errorf("op '%%' needs integers, but got %s and %s: %w", lang.TypeOf(left), lang.TypeOf(right), lang.ErrInvalidOperandTypeInteger)
		}
		if rightF == 0 {
			return nil, lang.NewRuntimeError(lang.ErrorCodeDivisionByZero, "modulo by zero", lang.ErrDivisionByZero)
		}
		return lang.NumberValue{Value: float64(int64(leftF) % int64(rightF))}, nil
	}

	switch op {
	case "-":
		return lang.NumberValue{Value: leftF - rightF}, nil
	case "*":
		return lang.NumberValue{Value: leftF * rightF}, nil
	case "/":
		if rightF == 0.0 {
			return nil, lang.NewRuntimeError(lang.ErrorCodeDivisionByZero, "division by zero", lang.ErrDivisionByZero)
		}
		return lang.NumberValue{Value: leftF / rightF}, nil
	case "**":
		return lang.NumberValue{Value: math.Pow(leftF, rightF)}, nil
	}

	return nil, fmt.Errorf("unknown arithmetic op '%s': %w", op, lang.ErrUnsupportedOperator)
}

// performStringConcatOrNumericAdd handles the '+' operator.
func performStringConcatOrNumericAdd(left, right lang.Value) (lang.Value, error) {
	leftNum, isLeftNum := lang.ToNumeric(left)
	rightNum, isRightNum := lang.ToNumeric(right)
	if isLeftNum && isRightNum {
		return lang.NumberValue{Value: leftNum.Value + rightNum.Value}, nil
	}

	leftStr, _ := lang.ToString(left)
	rightStr, _ := lang.ToString(right)

	return lang.StringValue{Value: leftStr + rightStr}, nil
}

// areValuesEqual performs a robust equality check between any two values.
func areValuesEqual(left, right lang.Value) bool {
	// FIX: Pass the values directly to Unwrap.
	leftNative := lang.Unwrap(left)
	rightNative := lang.Unwrap(right)

	if leftNative == nil || rightNative == nil {
		return leftNative == rightNative
	}

	// Handle numeric comparison separately to equate int-like floats.
	leftF, lOk := lang.ToFloat64(leftNative)
	rightF, rOk := lang.ToFloat64(rightNative)
	if lOk && rOk {
		return leftF == rightF
	}

	return reflect.DeepEqual(leftNative, rightNative)
}

// performComparison handles operators: ==, !=, <, >, <=, >=
func performComparison(left, right lang.Value, op string) (lang.Value, error) {
	if op == "==" {
		return lang.BoolValue{Value: areValuesEqual(left, right)}, nil
	}
	if op == "!=" {
		return lang.BoolValue{Value: !areValuesEqual(left, right)}, nil
	}

	// Type-specific comparisons for Timedate and Fuzzy
	if lVal, ok := left.(lang.TimedateValue); ok {
		if rVal, rOk := right.(lang.TimedateValue); rOk {
			switch op {
			case "<":
				return lang.BoolValue{Value: lVal.Value.Before(rVal.Value)}, nil
			case ">":
				return lang.BoolValue{Value: lVal.Value.After(rVal.Value)}, nil
			case "<=":
				return lang.BoolValue{Value: !lVal.Value.After(rVal.Value)}, nil
			case ">=":
				return lang.BoolValue{Value: !lVal.Value.Before(rVal.Value)}, nil
			}
		}
		return nil, typeErrorForOp(op, left, right)
	}

	if lVal, ok := left.(lang.FuzzyValue); ok {
		rVal, rOk := lang.ToNumeric(right)	// Compare against any numeric type
		if !rOk {
			return nil, typeErrorForOp(op, left, right)
		}
		switch op {
		case "<":
			return lang.BoolValue{Value: lVal.GetValue() < rVal.Value}, nil
		case ">":
			return lang.BoolValue{Value: lVal.GetValue() > rVal.Value}, nil
		case "<=":
			return lang.BoolValue{Value: lVal.GetValue() <= rVal.Value}, nil
		case ">=":
			return lang.BoolValue{Value: lVal.GetValue() >= rVal.Value}, nil
		}
	}

	// Fallback to numeric comparison for everything else
	leftF, leftOk := lang.ToNumeric(left)
	rightF, rightOk := lang.ToNumeric(right)
	if !leftOk || !rightOk {
		return nil, typeErrorForOp(op, left, right)
	}

	switch op {
	case "<":
		return lang.BoolValue{Value: leftF.Value < rightF.Value}, nil
	case ">":
		return lang.BoolValue{Value: leftF.Value > rightF.Value}, nil
	case "<=":
		return lang.BoolValue{Value: leftF.Value <= rightF.Value}, nil
	case ">=":
		return lang.BoolValue{Value: leftF.Value >= rightF.Value}, nil
	}

	return nil, fmt.Errorf("unknown comparison op '%s'", op)
}

// performBitwise handles operators: &, |, ^
func performBitwise(left, right lang.Value, op string) (lang.Value, error) {
	if left == nil || right == nil {
		return nil, fmt.Errorf("bitwise op '%s' needs non-nil operands: %w", op, lang.ErrNilOperand)
	}

	leftI, lOk := lang.ToInt64(left)
	rightI, rOk := lang.ToInt64(right)
	if !lOk || !rOk {
		return nil, fmt.Errorf("bitwise op '%s' needs integers, but got %s and %s: %w", op, lang.TypeOf(left), lang.TypeOf(right), lang.ErrInvalidOperandTypeInteger)
	}

	switch op {
	case "&":
		return lang.NumberValue{Value: float64(leftI & rightI)}, nil
	case "|":
		return lang.NumberValue{Value: float64(leftI | rightI)}, nil
	case "^":
		return lang.NumberValue{Value: float64(leftI ^ rightI)}, nil
	}

	return nil, fmt.Errorf("unknown bitwise op '%s'", op)
}