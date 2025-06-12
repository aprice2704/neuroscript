// NeuroScript Version: 0.4.2
// File version: 4
// Purpose: Corrected operator evaluation logic to be aware of wrapped Value types.
// filename: pkg/core/evaluation_operators.go

package core

import (
	"fmt"
	"math"
	"reflect"
)

const (
	// fuzzyEpsilon is used for float equality checks with FuzzyValue.
	fuzzyEpsilon = 1e-9
)

func isIntegerType(v interface{}) bool {
	if v == nil {
		return false
	}
	k := reflect.TypeOf(v).Kind()
	return k >= reflect.Int && k <= reflect.Uint64
}

func typeErrorForOp(op string, left, right interface{}) error {
	return fmt.Errorf("operator '%s' cannot be applied to types %T and %T: %w", op, left, right, ErrInvalidOperandType)
}

func performArithmetic(left, right interface{}, op string) (interface{}, error) {
	// Reject non-numeric Value types.
	switch left.(type) {
	case TimedateValue, ErrorValue, EventValue, FuzzyValue, StringValue, BoolValue, ListValue, MapValue, NilValue:
		return nil, typeErrorForOp(op, left, right)
	}
	switch right.(type) {
	case TimedateValue, ErrorValue, EventValue, FuzzyValue, StringValue, BoolValue, ListValue, MapValue, NilValue:
		return nil, typeErrorForOp(op, left, right)
	}

	// FIX: Directly check for NumberValue instead of relying on fragile helpers.
	leftNum, isLeftNum := left.(NumberValue)
	rightNum, isRightNum := right.(NumberValue)

	if !isLeftNum || !isRightNum {
		return nil, fmt.Errorf("op '%s' needs numerics: %w", op, ErrInvalidOperandTypeNumeric)
	}

	leftF := leftNum.Value
	rightF := rightNum.Value

	if op == "%" {
		// Modulo requires integer semantics. Check if the floats are whole numbers.
		if leftF != math.Trunc(leftF) || rightF != math.Trunc(rightF) {
			return nil, fmt.Errorf("op '%%' needs integers: %w", ErrInvalidOperandTypeInteger)
		}
		if rightF == 0 {
			return nil, fmt.Errorf("modulo by zero: %w", ErrDivisionByZero)
		}
		// Perform modulo on the integer parts.
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

func performStringConcatOrNumericAdd(left, right interface{}) (interface{}, error) {
	if _, ok := left.(FuzzyValue); ok {
		return nil, fmt.Errorf("cannot apply op '+' to fuzzy values")
	}
	if _, ok := right.(FuzzyValue); ok {
		return nil, fmt.Errorf("cannot apply op '+' to fuzzy values")
	}

	leftNum, isLeftNum := left.(NumberValue)
	rightNum, isRightNum := right.(NumberValue)

	// FIX: Prioritize numeric addition by directly checking for NumberValue types.
	if isLeftNum && isRightNum {
		return NumberValue{Value: leftNum.Value + rightNum.Value}, nil
	}

	// Fallback to string concatenation. Assumes all Value types have a .String() method.
	leftStr, lOk := toString(left)
	rightStr, rOk := toString(right)

	if !lOk || !rOk {
		return nil, fmt.Errorf("could not convert operands to string for concatenation: %T and %T", left, right)
	}

	return StringValue{Value: leftStr + rightStr}, nil
}

func performComparison(left, right interface{}, op string) (interface{}, error) {
	if op == "==" || op == "!=" {
		isEqual := reflect.DeepEqual(left, right)
		if op == "==" {
			return BoolValue{Value: isEqual}, nil
		}
		return BoolValue{Value: !isEqual}, nil
	}
	switch lVal := left.(type) {
	case TimedateValue:
		rVal, ok := right.(TimedateValue)
		if !ok {
			return nil, typeErrorForOp(op, left, right)
		}
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
	case FuzzyValue:
		rVal, ok := toFloat64(right)
		if !ok {
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
	case ErrorValue, EventValue:
		return nil, fmt.Errorf("inequality op (%s) not supported for %T", op, left)
	}
	leftF, leftOk := toFloat64(left)
	rightF, rightOk := toFloat64(right)
	if !leftOk || !rightOk {
		return nil, fmt.Errorf("comparison op '%s' needs comparable types, got %T and %T", op, left, right)
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

func performBitwise(left, right interface{}, op string) (interface{}, error) {
	if left == nil || right == nil {
		return nil, fmt.Errorf("bitwise op '%s' needs non-nil: %w", op, ErrNilOperand)
	}
	leftI, lOk := toInt64(left)
	rightI, rOk := toInt64(right)
	if !lOk || !rOk {
		return nil, fmt.Errorf("bitwise op '%s' needs integers: %w", op, ErrInvalidOperandTypeInteger)
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

// isZeroValue is a helper function that should be available to this package.
func isZeroValue(val interface{}) bool {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr, reflect.UnsafePointer:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}
