// NeuroScript Version: 0.8.0
// File version: 5
// Purpose: Corrected the fuzzy logic precedence for AND/OR operators to ensure boolean operations return booleans.
// filename: pkg/lang/operators_lang.go
// nlines: 240
// risk_rating: MEDIUM

package lang

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

// typeErrorForOp creates a standardized error for invalid operations.
func typeErrorForOp(op string, left, right Value) error {
	// FIX: Wrap the more specific ErrInvalidOperandType sentinel error.
	return fmt.Errorf("operator '%s' cannot be applied to types %s and %s: %w", op, TypeOf(left), TypeOf(right), ErrInvalidOperandType)
}

// PerformBinaryOperation performs infix binary operations.
func PerformBinaryOperation(op string, left, right Value) (Value, error) {
	opLower := strings.ToLower(op)

	// Fuzzy logic for AND/OR should only apply if one operand is actually fuzzy.
	_, leftIsFuzzy := left.(FuzzyValue)
	_, rightIsFuzzy := right.(FuzzyValue)

	if (opLower == "and" || opLower == "or") && (leftIsFuzzy || rightIsFuzzy) {
		leftF, _ := toFuzzy(left)
		rightF, _ := toFuzzy(right)
		if opLower == "and" {
			return NewFuzzyValue(math.Min(leftF.GetValue(), rightF.GetValue())), nil
		}
		return NewFuzzyValue(math.Max(leftF.GetValue(), rightF.GetValue())), nil
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
	// GATE: First, check for special non-numeric cases like string repetition.
	if op == "*" {
		// Case 1: string * int
		if s, ok := left.(StringValue); ok {
			if n, ok := right.(NumberValue); ok {
				if n.Value == math.Trunc(n.Value) && n.Value >= 0 {
					return StringValue{Value: strings.Repeat(s.Value, int(n.Value))}, nil
				}
			}
		}
		// Case 2: int * string
		if s, ok := right.(StringValue); ok {
			if n, ok := left.(NumberValue); ok {
				if n.Value == math.Trunc(n.Value) && n.Value >= 0 {
					return StringValue{Value: strings.Repeat(s.Value, int(n.Value))}, nil
				}
			}
		}
	}

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

	stringify := func(v Value) string {
		if v == nil {
			return ""
		}
		if _, ok := v.(*NilValue); ok {
			return ""
		}
		s, _ := ToString(v)
		return s
	}

	leftStr := stringify(left)
	rightStr := stringify(right)
	return StringValue{Value: leftStr + rightStr}, nil
}

// areValuesEqual provides a more robust, type-aware equality check.
func areValuesEqual(left, right Value) bool {
	// Nil is only equal to nil.
	_, leftIsNil := left.(NilValue)
	_, leftIsNilPtr := left.(*NilValue)
	_, rightIsNil := right.(NilValue)
	_, rightIsNilPtr := right.(*NilValue)
	if (leftIsNil || leftIsNilPtr) || (rightIsNil || rightIsNilPtr) {
		return (leftIsNil || leftIsNilPtr) && (rightIsNil || rightIsNilPtr)
	}

	// Handle core types directly for clarity and correctness.
	switch lVal := left.(type) {
	case StringValue:
		if rVal, ok := right.(StringValue); ok {
			return lVal.Value == rVal.Value
		}
		// Try numeric conversion for cross-type comparison, e.g., "5" == 5
		if _, ok := right.(NumberValue); ok {
			lNum, lOk := ToFloat64(lVal)
			rNum, rOk := ToFloat64(right)
			return lOk && rOk && lNum == rNum
		}
		return false

	case NumberValue:
		// Allow comparison with strings that can be numbers
		lNum, lOk := ToFloat64(lVal)
		rNum, rOk := ToFloat64(right)
		return lOk && rOk && lNum == rNum

	case BoolValue:
		if rVal, ok := right.(BoolValue); ok {
			return lVal.Value == rVal.Value
		}
		return false

	default:
		// Fallback for complex types like List, Map, Timedate, etc.
		return reflect.DeepEqual(Unwrap(left), Unwrap(right))
	}
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
