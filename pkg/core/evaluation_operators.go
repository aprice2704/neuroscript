// NeuroScript Version: 0.4.2
// File version: 5
// Purpose: Corrected operator evaluation logic to be aware of wrapped Value types.
// filename: pkg/core/evaluation_operators.go
// nlines: 201
// risk_rating: HIGH

package core

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
	// Use the helper to get NumberValues; it will fail for non-numeric types.
	leftNum, leftOk := ToNumeric(left)
	rightNum, rightOk := ToNumeric(right)

	if !leftOk || !rightOk {
		return nil, fmt.Errorf("op '%s' needs numerics, but got %s and %s: %w", op, TypeOf(left), TypeOf(right), ErrInvalidOperandTypeNumeric)
	}

	leftF := leftNum.Value
	rightF := rightNum.Value

	if op == "%" {
		// Modulo requires integer semantics. Check if the floats are whole numbers.
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

	// This path should ideally not be reached if called from evaluateBinaryOp
	return nil, fmt.Errorf("unknown arithmetic op '%s': %w", op, ErrUnsupportedOperator)
}

// performStringConcatOrNumericAdd handles the '+' operator.
func performStringConcatOrNumericAdd(left, right interface{}) (Value, error) {
	// Highest precedence: Numeric addition.
	// If both can be converted to numbers, perform addition.
	leftNum, isLeftNum := ToNumeric(left)
	rightNum, isRightNum := ToNumeric(right)
	if isLeftNum && isRightNum {
		return NumberValue{Value: leftNum.Value + rightNum.Value}, nil
	}

	// Fallback: String concatenation.
	// The toString helper is now robust enough to handle any Value type.
	leftStr, _ := toString(left)
	rightStr, _ := toString(right)

	return StringValue{Value: leftStr + rightStr}, nil
}

// areValuesEqual performs a robust equality check between any two values (raw or wrapped).
func areValuesEqual(left, right interface{}) bool {
	// Unwrap values to their native Go representation first.
	// This allows comparing, for example, a NumberValue with a raw int.
	leftNative := unwrapValue(left)
	rightNative := unwrapValue(right)

	if leftNative == nil || rightNative == nil {
		return leftNative == rightNative
	}

	// Attempt numeric comparison first
	leftF, lOk := toFloat64(leftNative)
	rightF, rOk := toFloat64(rightNative)
	if lOk && rOk {
		return leftF == rightF
	}

	// Use reflection for a deep comparison as a fallback.
	// This correctly handles lists, maps, etc.
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

	// Handle Timedate comparisons
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

	// Handle comparisons against FuzzyValue (always on the left)
	if lVal, ok := left.(FuzzyValue); ok {
		rVal, rOk := toFloat64(right) // Compare fuzzy's μ against a float
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

	// Default to numeric comparison for everything else
	leftF, leftOk := toFloat64(left)
	rightF, rightOk := toFloat64(right)
	if !leftOk || !rightOk {
		return nil, fmt.Errorf("comparison op '%s' needs comparable types, got %s and %s", op, TypeOf(left), TypeOf(right))
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

	// Use the robust helper which can handle NumberValue
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
