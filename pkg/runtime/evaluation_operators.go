// NeuroScript Version: 0.4.2
// File version: 12
// Purpose: Reverted '+' operator to be lenient, performing string coercion for mixed types as required by the test suite.
// filename: pkg/core/evaluation_operators.go
// nlines: 201
// risk_rating: MEDIUM

package runtime

import (
	"fmt"
	"math"
	"reflect"
)

// typeErrorForOp creates a standardized error for invalid operations.
func typeErrorForOp(op string, left, right interface{}) error {
	return fmt.Errorf("operator '%s' cannot be applied to types %s and %s: %w", op, TypeOf(left), TypeOf(right), ErrInvalidOperandType)
}

// performArithmetic handles operators: -, *, /, %, **
func performArithmetic(left, right interface{}, op string) (Value, error) {
	leftNum, leftOk := ToNumeric(left)
	rightNum, rightOk := ToNumeric(right)

	if !leftOk || !rightOk {
		return nil, fmt.Errorf("op '%s' needs numerics, but got %s and %s: %w", op, TypeOf(left), TypeOf(right), ErrInvalidOperandTypeNumeric)
	}

	leftF := leftNum.Value
	rightF := rightNum.Value

	if op == "%" {
		if leftF != math.Trunc(leftF) || rightF != math.Trunc(rightF) {
			return nil, fmt.Errorf("op '%%' needs integers, but got %s and %s: %w", TypeOf(left), TypeOf(right), ErrInvalidOperandTypeInteger)
		}
		if rightF == 0 {
			return nil, fmt.Errorf("modulo by zero: %w", ErrDivisionByZero)
		}
		return NumberValue{Value: float64(int64(leftF) % int64(rightF))}, nil
	}

	switch op {
	case "-":
		return NumberValue{Value: leftF - rightF}, nil
	case "*":
		return NumberValue{Value: leftF * rightF}, nil
	case "/":
		if rightF == 0.0 {
			return nil, fmt.Errorf("division by zero: %w", ErrDivisionByZero)
		}
		return NumberValue{Value: leftF / rightF}, nil
	case "**":
		return NumberValue{Value: math.Pow(leftF, rightF)}, nil
	}

	return nil, fmt.Errorf("unknown arithmetic op '%s': %w", op, ErrUnsupportedOperator)
}

// performStringConcatOrNumericAdd handles the '+' operator.
func performStringConcatOrNumericAdd(left, right interface{}) (Value, error) {
	// FINAL, LENIENT LOGIC: The desired behavior is to add if both
	// operands are numeric, and otherwise coerce to string and concatenate.
	leftNum, isLeftNum := ToNumeric(left)
	rightNum, isRightNum := ToNumeric(right)
	if isLeftNum && isRightNum {
		return NumberValue{Value: leftNum.Value + rightNum.Value}, nil
	}

	// If not both numbers, perform string concatenation.
	leftStr, _ := toString(left)
	rightStr, _ := toString(right)

	return StringValue{Value: leftStr + rightStr}, nil
}

// areValuesEqual performs a robust equality check between any two values (raw or wrapped).
func areValuesEqual(left, right interface{}) bool {
	leftNative := unwrapValue(left)
	rightNative := unwrapValue(right)

	if leftNative == nil || rightNative == nil {
		return leftNative == rightNative
	}

	leftF, lOk := toFloat64(leftNative)
	rightF, rOk := toFloat64(rightNative)
	if lOk && rOk {
		return leftF == rightF
	}

	return reflect.DeepEqual(leftNative, rightNative)
}

// performComparison handles operators: ==, !=, <, >, <=, >=
func performComparison(left, right interface{}, op string) (Value, error) {
	if op == "==" {
		return BoolValue{Value: areValuesEqual(left, right)}, nil
	}
	if op == "!=" {
		return BoolValue{Value: !areValuesEqual(left, right)}, nil
	}

	if lVal, ok := left.(TimedateValue); ok {
		if rVal, rOk := right.(TimedateValue); rOk {
			switch op {
			case "<":
				return BoolValue{Value: lVal.Value.Before(rVal.Value)}, nil
			case ">":
				return BoolValue{Value: lVal.Value.After(rVal.Value)}, nil
			case "<=":
				return BoolValue{Value: !lVal.Value.After(rVal.Value)}, nil
			case ">=":
				return BoolValue{Value: !lVal.Value.Before(rVal.Value)}, nil
			}
		}
		return nil, typeErrorForOp(op, left, right)
	}

	if lVal, ok := left.(FuzzyValue); ok {
		rVal, rOk := toFloat64(right)
		if !rOk {
			return nil, typeErrorForOp(op, left, right)
		}
		switch op {
		case "<":
			return BoolValue{Value: lVal.μ < rVal}, nil
		case ">":
			return BoolValue{Value: lVal.μ > rVal}, nil
		case "<=":
			return BoolValue{Value: lVal.μ <= rVal}, nil
		case ">=":
			return BoolValue{Value: lVal.μ >= rVal}, nil
		}
	}

	leftF, leftOk := toFloat64(left)
	rightF, rightOk := toFloat64(right)
	if !leftOk || !rightOk {
		return nil, typeErrorForOp(op, left, right)
	}

	switch op {
	case "<":
		return BoolValue{Value: leftF < rightF}, nil
	case ">":
		return BoolValue{Value: leftF > rightF}, nil
	case "<=":
		return BoolValue{Value: leftF <= rightF}, nil
	case ">=":
		return BoolValue{Value: leftF >= rightF}, nil
	}

	return nil, fmt.Errorf("unknown comparison op '%s'", op)
}

// performBitwise handles operators: &, |, ^
func performBitwise(left, right interface{}, op string) (Value, error) {
	if left == nil || right == nil {
		return nil, fmt.Errorf("bitwise op '%s' needs non-nil operands: %w", op, ErrNilOperand)
	}

	leftI, lOk := toInt64(left)
	rightI, rOk := toInt64(right)
	if !lOk || !rOk {
		return nil, fmt.Errorf("bitwise op '%s' needs integers, but got %s and %s: %w", op, TypeOf(left), TypeOf(right), ErrInvalidOperandTypeInteger)
	}

	switch op {
	case "&":
		return NumberValue{Value: float64(leftI & rightI)}, nil
	case "|":
		return NumberValue{Value: float64(leftI | rightI)}, nil
	case "^":
		return NumberValue{Value: float64(leftI ^ rightI)}, nil
	}

	return nil, fmt.Errorf("unknown bitwise op '%s'", op)
}
