// NeuroScript Version: 0.4.1
// File version: 2
// Purpose: Defines operator evaluation logic, including comparisons and arithmetic.
// filename: core/evaluation_operators.go
// nlines: 200
// risk_rating: MEDIUM

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
	switch left.(type) {
	case TimedateValue, ErrorValue, EventValue:
		return nil, typeErrorForOp(op, left, right)
	case FuzzyValue:
		return nil, fmt.Errorf("cannot apply arithmetic operator '%s' to fuzzy values", op)
	}
	switch right.(type) {
	case TimedateValue, ErrorValue, EventValue, FuzzyValue:
		return nil, typeErrorForOp(op, left, right)
	}
	if left == nil || right == nil {
		return nil, fmt.Errorf("%w: op '%s' requires non-nil", ErrNilOperand, op)
	}
	if op == "%" {
		if !isIntegerType(left) || !isIntegerType(right) {
			return nil, fmt.Errorf("op '%%' needs integers: %w", ErrInvalidOperandTypeInteger)
		}
		leftI, _ := toInt64(left)
		rightI, _ := toInt64(right)
		if rightI == 0 {
			return nil, fmt.Errorf("modulo by zero: %w", ErrDivisionByZero)
		}
		return leftI % rightI, nil
	}
	_, leftIsNum := ToNumeric(left)
	_, rightIsNum := ToNumeric(right)
	if !leftIsNum || !rightIsNum {
		return nil, fmt.Errorf("op '%s' needs numerics: %w", op, ErrInvalidOperandTypeNumeric)
	}
	leftF, lfo := toFloat64(left)
	rightF, rfo := toFloat64(right)
	leftI, lio := toInt64(left)
	rightI, rio := toInt64(right)
	useFloat := !lio || !rio
	switch op {
	case "-":
		if useFloat {
			return leftF - rightF, nil
		}
		return leftI - rightI, nil
	case "*":
		if useFloat {
			return leftF * rightF, nil
		}
		return leftI * rightI, nil
	case "/":
		if rfo && rightF == 0.0 {
			return nil, fmt.Errorf("division by zero: %w", ErrDivisionByZero)
		}
		if useFloat || (rio && leftI%rightI != 0) {
			if !lfo {
				return nil, fmt.Errorf("internal float conversion error")
			}
			return leftF / rightF, nil
		}
		return leftI / rightI, nil
	case "**":
		if !lfo || !rfo {
			return nil, fmt.Errorf("op '**' needs floats: %w", ErrInvalidOperandTypeNumeric)
		}
		return math.Pow(leftF, rightF), nil
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
	_, leftIsNum := ToNumeric(left)
	_, rightIsNum := ToNumeric(right)
	if leftIsNum && rightIsNum {
		leftF, _ := toFloat64(left)
		rightF, _ := toFloat64(right)
		leftI, lio := toInt64(left)
		rightI, rio := toInt64(right)
		if lio && rio {
			return leftI + rightI, nil
		}
		return leftF + rightF, nil
	}
	leftStr, _ := toString(left)
	rightStr, _ := toString(right)
	return leftStr + rightStr, nil
}

func performComparison(left, right interface{}, op string) (bool, error) {
	if op == "==" || op == "!=" {
		isEqual := reflect.DeepEqual(left, right)
		if op == "==" {
			return isEqual, nil
		}
		return !isEqual, nil
	}
	switch lVal := left.(type) {
	case TimedateValue:
		rVal, ok := right.(TimedateValue)
		if !ok {
			return false, typeErrorForOp(op, left, right)
		}
		switch op {
		case "<":
			return lVal.Value.Before(rVal.Value), nil
		case ">":
			return lVal.Value.After(rVal.Value), nil
		case "<=":
			return !lVal.Value.After(rVal.Value), nil
		case ">=":
			return !lVal.Value.Before(rVal.Value), nil
		}
	case FuzzyValue:
		rVal, ok := toFloat64(right)
		if !ok {
			return false, typeErrorForOp(op, left, right)
		}
		switch op {
		case "<":
			return lVal.μ < rVal, nil
		case ">":
			return lVal.μ > rVal, nil
		case "<=":
			return lVal.μ <= rVal, nil
		case ">=":
			return lVal.μ >= rVal, nil
		}
	case ErrorValue, EventValue:
		return false, fmt.Errorf("inequality op (%s) not supported for %T", op, left)
	}
	leftF, leftOk := toFloat64(left)
	rightF, rightOk := toFloat64(right)
	if !leftOk || !rightOk {
		return false, fmt.Errorf("comparison op '%s' needs comparable types, got %T and %T", op, left, right)
	}
	switch op {
	case "<":
		return leftF < rightF, nil
	case ">":
		return leftF > rightF, nil
	case "<=":
		return leftF <= rightF, nil
	case ">=":
		return leftF >= rightF, nil
	}
	return false, fmt.Errorf("unknown comparison op '%s'", op)
}

func performBitwise(left, right interface{}, op string) (int64, error) {
	if left == nil || right == nil {
		return 0, fmt.Errorf("bitwise op '%s' needs non-nil: %w", op, ErrNilOperand)
	}
	if !isIntegerType(left) || !isIntegerType(right) {
		return 0, fmt.Errorf("bitwise op '%s' needs integers: %w", op, ErrInvalidOperandTypeInteger)
	}
	leftI, _ := toInt64(left)
	rightI, _ := toInt64(right)
	switch op {
	case "&":
		return leftI & rightI, nil
	case "|":
		return leftI | rightI, nil
	case "^":
		return leftI ^ rightI, nil
	}
	return 0, fmt.Errorf("unknown bitwise op '%s'", op)
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
