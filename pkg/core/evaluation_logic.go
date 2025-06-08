// NeuroScript Version: 0.4.1
// File version: 10
// Purpose: Final correction to evaluateBinaryOp to remove boolean fallback for and/or.
// filename: core/evaluation_logic.go
// nlines: 125
// risk_rating: MEDIUM

package core

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

// ... (TypeOf, evaluateUnaryOp, toFuzzy functions are unchanged) ...
func (i *Interpreter) TypeOf(value interface{}) string {
	if value == nil {
		return string(TypeNil)
	}
	if v, ok := value.(Value); ok {
		return string(v.Type())
	}
	switch value.(type) {
	case Procedure:
		return string(TypeFunction)
	case ToolImplementation:
		return string(TypeTool)
	}
	val := reflect.ValueOf(value)
	kind := val.Kind()
	if kind == reflect.Interface {
		if val.IsNil() {
			return string(TypeNil)
		}
		val = val.Elem()
		kind = val.Kind()
	}
	for kind == reflect.Ptr {
		if val.IsNil() {
			return string(TypeNil)
		}
		val = val.Elem()
		kind = val.Kind()
	}
	switch kind {
	case reflect.String:
		return string(TypeString)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return string(TypeNumber)
	case reflect.Bool:
		return string(TypeBoolean)
	case reflect.Slice, reflect.Array:
		return string(TypeList)
	case reflect.Map:
		return string(TypeMap)
	case reflect.Func:
		return string(TypeFunction)
	default:
		return string(TypeUnknown)
	}
}

func evaluateUnaryOp(op string, operand interface{}) (interface{}, error) {
	if operand == nil && (op == "-" || op == "~") {
		return nil, fmt.Errorf("%w: unary operator '%s'", ErrNilOperand, op)
	}
	switch strings.ToLower(op) {
	case "not":
		if f, ok := operand.(FuzzyValue); ok {
			return NewFuzzyValue(1.0 - f.μ), nil
		}
		return !isTruthy(operand), nil
	case "-":
		iVal, isInt := toInt64(operand)
		if isInt {
			return -iVal, nil
		}
		fVal, isFloat := toFloat64(operand)
		if isFloat {
			return -fVal, nil
		}
		return nil, fmt.Errorf("%w: unary operator '-' needs number, got %T", ErrInvalidOperandTypeNumeric, operand)
	case "~":
		iVal, isInt := toInt64(operand)
		if isInt {
			return ^iVal, nil
		}
		return nil, fmt.Errorf("%w: unary operator '~' needs integer, got %T", ErrInvalidOperandTypeInteger, operand)
	case "no":
		return isZeroValue(operand), nil
	case "some":
		return !isZeroValue(operand), nil
	default:
		return nil, fmt.Errorf("%w: '%s'", ErrUnsupportedOperator, op)
	}
}

func toFuzzy(v interface{}) (FuzzyValue, bool) {
	if f, ok := v.(FuzzyValue); ok {
		return f, true
	}
	if b, ok := v.(bool); ok {
		if b {
			return NewFuzzyValue(1.0), true
		}
		return NewFuzzyValue(0.0), true
	}
	return FuzzyValue{}, false
}

// evaluateBinaryOp performs infix binary operations.
func evaluateBinaryOp(left, right interface{}, op string) (interface{}, error) {
	opLower := strings.ToLower(op)

	// The boolean short-circuit for `and` and `or` is handled in `evaluateExpression`.
	// If this function is called with `and` or `or`, it MUST be a fuzzy logic operation.
	if opLower == "and" || opLower == "or" {
		leftF, leftIsFuzzyCoercible := toFuzzy(left)
		rightF, rightIsFuzzyCoercible := toFuzzy(right)

		// If both can be treated as fuzzy, perform the operation.
		if leftIsFuzzyCoercible && rightIsFuzzyCoercible {
			if opLower == "and" {
				return NewFuzzyValue(math.Min(leftF.μ, rightF.μ)), nil
			}
			return NewFuzzyValue(math.Max(leftF.μ, rightF.μ)), nil
		}
		// If we are here for an 'and'/'or' but it's not a valid fuzzy op, it's a type error.
		return nil, typeErrorForOp(op, left, right)
	}

	// Handle all other non-and/or operators.
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
