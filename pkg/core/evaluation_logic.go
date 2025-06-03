// NeuroScript Version: 0.3.1 // Assuming current project version
// File version: 0.1.1 // Added TypeOf method for the 'typeof' operator.
// Purpose: Defines evaluation logic for NeuroScript operations, built-in functions, and the typeof operator.
// filename: pkg/core/evaluation_logic.go
// nlines: 200 // Approximate
// risk_rating: LOW

package core

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	// Assuming errors like ErrNilOperand, ErrInvalidOperandType*, ErrUnsupportedOperator,
	// ErrIncorrectArgCount, ErrInvalidFunctionArgument, ErrUnknownFunction
	// are defined in errors.go
)

// TypeOf determines the NeuroScript type string for a given Go value.
// It aims to return one of the predefined NeuroScriptType constants from type_names.go.
func (i *Interpreter) TypeOf(value interface{}) string {
	if value == nil {
		return string(TypeNil) // From core/type_names.go
	}

	// Handle specific NeuroScript constructs first
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
	case reflect.Func: // Raw Go funcs (after Procedure/ToolImplementation checks)
		return string(TypeFunction)
	default:
		// Optional: Log unhandled kinds if necessary for debugging,
		// but ensure logger is accessible (e.g., i.Logger().Debugf(...))
		// For now, just return TypeUnknown as per previous design.
		return string(TypeUnknown)
	}
}

// --- Evaluation Logic for Operations ---

// isZeroValue checks if a value is the zero value for its type. (Unchanged)
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

// evaluateUnaryOp performs prefix unary operations (not, -, no, some, ~). (Unchanged)
func evaluateUnaryOp(op string, operand interface{}) (interface{}, error) {
	if operand == nil && (op == "-" || op == "~") {
		return nil, fmt.Errorf("%w: unary operator '%s'", ErrNilOperand, op)
	}
	switch strings.ToLower(op) {
	case "not":
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

// evaluateBinaryOp performs infix binary operations (dispatching). (Unchanged)
func evaluateBinaryOp(left, right interface{}, op string) (interface{}, error) {
	opLower := strings.ToLower(op)
	switch opLower {
	case "and":
		if !isTruthy(left) {
			return false, nil
		}
		return isTruthy(right), nil
	case "or":
		if isTruthy(left) {
			return true, nil
		}
		return isTruthy(right), nil
	case "==", "!=", "<", ">", "<=", ">=":
		result, err := performComparison(left, right, op)
		if err != nil {
			return nil, fmt.Errorf("compare %T %s %T: %w", left, op, right, err)
		}
		return result, nil
	case "+":
		result, err := performStringConcatOrNumericAdd(left, right)
		if err != nil {
			return nil, fmt.Errorf("op %T + %T: %w", left, right, err)
		}
		return result, nil
	case "-", "*", "/", "%", "**":
		result, err := performArithmetic(left, right, op)
		if err != nil {
			return nil, fmt.Errorf("arithmetic %T %s %T: %w", left, op, right, err)
		}
		return result, nil
	case "&", "|", "^":
		result, err := performBitwise(left, right, op)
		if err != nil {
			return nil, fmt.Errorf("bitwise %T %s %T: %w", left, op, right, err)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("%w: '%s'", ErrUnsupportedOperator, op)
	}
}

// --- ADDED isBuiltInFunction helper ---
// isBuiltInFunction checks if a name corresponds to a known built-in function.
func isBuiltInFunction(name string) bool {
	switch strings.ToLower(name) { // Use lower case for check
	case "ln", "log", "sin", "cos", "tan", "asin", "acos", "atan",
		"is_string", "is_number", "is_int", "is_float", "is_bool", "is_list", "is_map", "not_empty":
		return true
	default:
		return false
	}
}

// --- RESTORED evaluateBuiltInFunction ---
// evaluateBuiltInFunction handles built-in function calls.
func evaluateBuiltInFunction(funcName string, args []interface{}) (interface{}, error) {
	// Helper to check arg count
	checkArgCount := func(expectedCount int) error {
		if len(args) != expectedCount {
			return fmt.Errorf("%w: func %s expects %d arg(s), got %d", ErrIncorrectArgCount, funcName, expectedCount, len(args))
		}
		return nil
	}

	funcLower := strings.ToLower(funcName) // Convert func name for case-insensitive matching

	switch funcLower {
	// --- Type Check Functions (for mustBe) ---
	case "is_string":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(string)
		return ok, nil
	case "is_number":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		if args[0] == nil {
			return false, nil
		}
		k := reflect.ValueOf(args[0]).Kind()
		isNum := (k >= reflect.Int && k <= reflect.Int64) ||
			(k >= reflect.Uint && k <= reflect.Uint64) || // Exclude Uintptr
			k == reflect.Float32 || k == reflect.Float64
		return isNum, nil
	case "is_int":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		if args[0] == nil {
			return false, nil
		}
		k := reflect.ValueOf(args[0]).Kind()
		isInt := (k >= reflect.Int && k <= reflect.Int64) ||
			(k >= reflect.Uint && k <= reflect.Uint64) // Exclude Uintptr
		return isInt, nil
	case "is_float":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		if args[0] == nil {
			return false, nil
		}
		k := reflect.ValueOf(args[0]).Kind()
		isFloat := k == reflect.Float32 || k == reflect.Float64
		return isFloat, nil
	case "is_bool":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(bool)
		return ok, nil
	case "is_list":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		if args[0] == nil {
			return false, nil
		}
		k := reflect.ValueOf(args[0]).Kind()
		isList := k == reflect.Slice || k == reflect.Array
		return isList, nil
	case "is_map":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		if args[0] == nil {
			return false, nil
		}
		k := reflect.ValueOf(args[0]).Kind()
		isMap := k == reflect.Map
		return isMap, nil
	case "not_empty":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		if args[0] == nil {
			return false, nil
		}
		v := reflect.ValueOf(args[0])
		k := v.Kind()
		if k == reflect.Slice || k == reflect.Map || k == reflect.String || k == reflect.Array {
			return v.Len() > 0, nil
		}
		return nil, fmt.Errorf("%w: func %s expects list, map, or string argument, got %T", ErrInvalidFunctionArgument, funcName, args[0])

	// --- Built-in Math Functions ---
	case "ln":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: LN needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		if fVal <= 0 {
			return nil, fmt.Errorf("%w: LN needs positive arg, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Log(fVal), nil
	case "log":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: LOG needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		if fVal <= 0 {
			return nil, fmt.Errorf("%w: LOG needs positive arg, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Log10(fVal), nil
	case "sin":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: SIN needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		return math.Sin(fVal), nil
	case "cos":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: COS needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		return math.Cos(fVal), nil
	case "tan":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: TAN needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		return math.Tan(fVal), nil
	case "asin":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: ASIN needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		if fVal < -1 || fVal > 1 {
			return nil, fmt.Errorf("%w: ASIN needs arg between -1 and 1, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Asin(fVal), nil
	case "acos":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: ACOS needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		if fVal < -1 || fVal > 1 {
			return nil, fmt.Errorf("%w: ACOS needs arg between -1 and 1, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Acos(fVal), nil
	case "atan":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: ATAN needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		return math.Atan(fVal), nil

	default:
		// Return specific error for unknown functions
		// This case should technically not be reached if isBuiltInFunction is checked first by the caller,
		// but handle it defensively.
		return nil, fmt.Errorf("%w: '%s'", ErrUnknownFunction, funcName)
	}
}

// --- Placeholders for other evaluation logic functions ---
// func performComparison(left, right interface{}, op string) (bool, error) { ... }
// func performStringConcatOrNumericAdd(left, right interface{}) (interface{}, error) { ... }
// func performArithmetic(left, right interface{}, op string) (interface{}, error) { ... }
// func performBitwise(left, right interface{}, op string) (interface{}, error) { ... }
// func toInt64(v interface{}) (int64, bool) { ... }
// func toFloat64(v interface{}) (float64, bool) { ... }
// func isTruthy(value interface{}) bool { ... }
