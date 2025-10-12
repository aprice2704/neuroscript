// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Centralizes logic for performing binary and unary operations on lang.Value types.
// filename: pkg/lang/operators.go
// nlines: 200
// risk_rating: HIGH

package lang

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

// typeErrorForOp creates a standardized error for invalid operations.
func typeErrorForOp(op string, left, right Value) error {
	return fmt.Errorf("operator '%s' cannot be applied to types %s and %s: %w", op, TypeOf(left), TypeOf(right), ErrInvalidOperation)
}

// PerformBinaryOperation performs infix binary operations.
func PerformBinaryOperation(op string, left, right Value) (Value, error) {
	opLower := strings.ToLower(op)

	// Fuzzy logic for AND/OR
	if opLower == "and" || opLower == "or" {
		leftF, leftIsFuzzy := toFuzzy(left)
		rightF, rightIsFuzzy := toFuzzy(right)
		if leftIsFuzzy && rightIsFuzzy {
			if opLower == "and" {
				return NewFuzzyValue(math.Min(leftF.GetValue(), rightF.GetValue())), nil
			}
			return NewFuzzyValue(math.Max(leftF.GetValue(), rightF.GetValue())), nil
		}
	}

	// Standard boolean logic
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

	// Comparisons, arithmetic, and bitwise operations
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

// PerformUnaryOperation handles unary operations like NOT, -, and ~.
func PerformUnaryOperation(op string, operand Value) (Value, error) {
	if operand == nil && (op == "-" || op == "~") {
		return nil, fmt.Errorf("%w: unary operator '%s'", ErrNilOperand, op)
	}

	switch strings.ToLower(op) {
	case "not":
		if fv, ok := operand.(FuzzyValue); ok {
			return NewFuzzyValue(1.0 - fv.GetValue()), nil
		}
		return BoolValue{Value: !IsTruthy(operand)}, nil

	case "-":
		num, ok := ToNumeric(operand)
		if !ok {
			return nil, fmt.Errorf("%w: unary operator '-' needs number, got %s", ErrInvalidOperandTypeNumeric, TypeOf(operand))
		}
		return NumberValue{Value: -num.Value}, nil
	case "~":
		iVal, isInt := ToInt64(operand)
		if !isInt {
			return nil, fmt.Errorf("%w: unary operator '~' needs integer, got %s", ErrInvalidOperandTypeInteger, TypeOf(operand))
		}
		return NumberValue{Value: float64(^iVal)}, nil
	case "no":
		return BoolValue{Value: IsZeroValue(operand)}, nil
	case "some":
		return BoolValue{Value: !IsZeroValue(operand)}, nil
	default:
		return nil, fmt.Errorf("%w: '%s'", ErrUnsupportedOperator, op)
	}
}

// --- Private Helpers ---

func performArithmetic(left, right Value, op string) (Value, error) {
	leftNum, leftOk := ToNumeric(left)
	rightNum, rightOk := ToNumeric(right)

	if !leftOk || !rightOk {
		return nil, typeErrorForOp(op, left, right)
	}
	leftF, rightF := leftNum.Value, rightNum.Value

	if op == "%" {
		if leftF != math.Trunc(leftF) || rightF != math.Trunc(rightF) {
			return nil, fmt.Errorf("op '%%' needs integers, but got %s and %s: %w", TypeOf(left), TypeOf(right), ErrInvalidOperandTypeInteger)
		}
		if rightF == 0 {
			return nil, NewRuntimeError(ErrorCodeDivisionByZero, "modulo by zero", ErrDivisionByZero)
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
			return nil, NewRuntimeError(ErrorCodeDivisionByZero, "division by zero", ErrDivisionByZero)
		}
		return NumberValue{Value: leftF / rightF}, nil
	case "**":
		return NumberValue{Value: math.Pow(leftF, rightF)}, nil
	}
	return nil, fmt.Errorf("unknown arithmetic op '%s': %w", op, ErrUnsupportedOperator)
}

func performStringConcatOrNumericAdd(left, right Value) (Value, error) {
	leftNum, isLeftNum := ToNumeric(left)
	rightNum, isRightNum := ToNumeric(right)
	if isLeftNum && isRightNum {
		return NumberValue{Value: leftNum.Value + rightNum.Value}, nil
	}
	var leftStr, rightStr string
	if _, isNil := left.(*NilValue); isNil {
		leftStr = ""
	} else {
		leftStr, _ = ToString(left)
	}
	if _, isNil := right.(*NilValue); isNil {
		rightStr = ""
	} else {
		rightStr, _ = ToString(right)
	}
	return StringValue{Value: leftStr + rightStr}, nil
}

func areValuesEqual(left, right Value) bool {
	leftNative, rightNative := Unwrap(left), Unwrap(right)
	if leftNative == nil || rightNative == nil {
		return leftNative == rightNative
	}
	leftF, lOk := ToFloat64(leftNative)
	rightF, rOk := ToFloat64(rightNative)
	if lOk && rOk {
		return leftF == rightF
	}
	return reflect.DeepEqual(leftNative, rightNative)
}

func performComparison(left, right Value, op string) (Value, error) {
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
		rVal, rOk := ToNumeric(right)
		if !rOk {
			return nil, typeErrorForOp(op, left, right)
		}
		switch op {
		case "<":
			return BoolValue{Value: lVal.GetValue() < rVal.Value}, nil
		case ">":
			return BoolValue{Value: lVal.GetValue() > rVal.Value}, nil
		case "<=":
			return BoolValue{Value: lVal.GetValue() <= rVal.Value}, nil
		case ">=":
			return BoolValue{Value: lVal.GetValue() >= rVal.Value}, nil
		}
	}
	leftF, leftOk := ToNumeric(left)
	rightF, rightOk := ToNumeric(right)
	if !leftOk || !rightOk {
		return nil, typeErrorForOp(op, left, right)
	}
	switch op {
	case "<":
		return BoolValue{Value: leftF.Value < rightF.Value}, nil
	case ">":
		return BoolValue{Value: leftF.Value > rightF.Value}, nil
	case "<=":
		return BoolValue{Value: leftF.Value <= rightF.Value}, nil
	case ">=":
		return BoolValue{Value: leftF.Value >= rightF.Value}, nil
	}
	return nil, fmt.Errorf("unknown comparison op '%s'", op)
}

func performBitwise(left, right Value, op string) (Value, error) {
	if left == nil || right == nil {
		return nil, fmt.Errorf("bitwise op '%s' needs non-nil operands: %w", op, ErrNilOperand)
	}
	leftI, lOk := ToInt64(left)
	rightI, rOk := ToInt64(right)
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
