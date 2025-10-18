// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Adds more comprehensive error condition tests for binary/unary operators.
// filename: pkg/lang/operators_lang_test.go
// nlines: 200
// risk_rating: LOW

package lang

import (
	"errors"
	"reflect"
	"testing"
)

func TestPerformBinaryOperation(t *testing.T) {
	testCases := []struct {
		name     string
		op       string
		left     Value
		right    Value
		expected Value
		wantErr  bool
		errIs    error
	}{
		// --- Equality ---
		{name: `string == string (true)`, op: "==", left: StringValue{"a"}, right: StringValue{"a"}, expected: BoolValue{true}},
		{name: `string == string (false)`, op: "==", left: StringValue{"a"}, right: StringValue{"b"}, expected: BoolValue{false}},
		{name: `number == number (true)`, op: "==", left: NumberValue{5}, right: NumberValue{5.0}, expected: BoolValue{true}},
		{name: `number == string (true)`, op: "==", left: NumberValue{5}, right: StringValue{"5"}, expected: BoolValue{true}},
		{name: `string == number (true)`, op: "==", left: StringValue{"5.0"}, right: NumberValue{5}, expected: BoolValue{true}},
		{name: `string == number (false)`, op: "==", left: StringValue{"5.1"}, right: NumberValue{5}, expected: BoolValue{false}},
		{name: `nil == nil`, op: "==", left: NilValue{}, right: NilValue{}, expected: BoolValue{true}},
		{name: `nil == *nil`, op: "==", left: NilValue{}, right: &NilValue{}, expected: BoolValue{true}},
		{name: `*nil == *nil`, op: "==", left: &NilValue{}, right: &NilValue{}, expected: BoolValue{true}},
		{name: `string == nil (false)`, op: "==", left: StringValue{""}, right: NilValue{}, expected: BoolValue{false}},
		{name: `nil == string (false)`, op: "==", left: NilValue{}, right: StringValue{""}, expected: BoolValue{false}},
		{name: `bool == bool (true)`, op: "==", left: BoolValue{true}, right: BoolValue{true}, expected: BoolValue{true}},
		{name: `bool == nil (false)`, op: "==", left: BoolValue{false}, right: NilValue{}, expected: BoolValue{false}},

		// --- Inequality ---
		{name: `string != string (true)`, op: "!=", left: StringValue{"a"}, right: StringValue{"b"}, expected: BoolValue{true}},
		{name: `number != number (false)`, op: "!=", left: NumberValue{5}, right: NumberValue{5}, expected: BoolValue{false}},
		{name: `nil != nil (false)`, op: "!=", left: NilValue{}, right: NilValue{}, expected: BoolValue{false}},
		{name: `string != nil (true)`, op: "!=", left: StringValue{""}, right: NilValue{}, expected: BoolValue{true}},

		// --- Comparison ---
		{name: `number > number`, op: ">", left: NumberValue{10}, right: NumberValue{5}, expected: BoolValue{true}},
		{name: `number <= number`, op: "<=", left: NumberValue{5}, right: NumberValue{5}, expected: BoolValue{true}},

		// --- Arithmetic ---
		{name: `number + number`, op: "+", left: NumberValue{2}, right: NumberValue{3}, expected: NumberValue{5}},
		{name: `string + number`, op: "+", left: StringValue{"a"}, right: NumberValue{3}, expected: StringValue{"a3"}},
		{name: `string + nil`, op: "+", left: StringValue{"a"}, right: NilValue{}, expected: StringValue{"a"}},
		{name: `nil + number`, op: "+", left: NilValue{}, right: NumberValue{3}, expected: StringValue{"3"}},
		{name: `number - number`, op: "-", left: NumberValue{10}, right: NumberValue{3}, expected: NumberValue{7}},
		{name: `number * number`, op: "*", left: NumberValue{4}, right: NumberValue{5}, expected: NumberValue{20}},
		{name: `number / number`, op: "/", left: NumberValue{20}, right: NumberValue{4}, expected: NumberValue{5}},
		{name: `string * number (repetition)`, op: "*", left: StringValue{"a"}, right: NumberValue{3}, expected: StringValue{"aaa"}},

		// --- Logical ---
		{name: `true and false`, op: "and", left: BoolValue{true}, right: BoolValue{false}, expected: BoolValue{false}},
		{name: `true or false`, op: "or", left: BoolValue{true}, right: BoolValue{false}, expected: BoolValue{true}},
		{name: `0 or "a"`, op: "or", left: NumberValue{0}, right: StringValue{"a"}, expected: BoolValue{true}},

		// --- Error Cases ---
		{name: `division by zero`, op: "/", left: NumberValue{1}, right: NumberValue{0}, wantErr: true, errIs: ErrDivisionByZero},
		{name: `invalid operand type for - (string, num)`, op: "-", left: StringValue{"a"}, right: NumberValue{1}, wantErr: true, errIs: ErrInvalidOperandType},
		{name: `invalid operand type for - (num, string)`, op: "-", left: NumberValue{1}, right: StringValue{"a"}, wantErr: true, errIs: ErrInvalidOperandType},
		{name: `invalid operand type for / (string, num)`, op: "/", left: StringValue{"a"}, right: NumberValue{1}, wantErr: true, errIs: ErrInvalidOperandType},
		{name: `invalid operand type for * (string, string)`, op: "*", left: StringValue{"a"}, right: StringValue{"b"}, wantErr: true, errIs: ErrInvalidOperandType},
		{name: `invalid operand type for ** (string, num)`, op: "**", left: StringValue{"a"}, right: NumberValue{2}, wantErr: true, errIs: ErrInvalidOperandType},
		{name: `invalid comparison (string < bool)`, op: "<", left: StringValue{"a"}, right: BoolValue{false}, wantErr: true, errIs: ErrInvalidOperandType},
		{name: `invalid comparison (num > bool)`, op: ">", left: NumberValue{1}, right: BoolValue{true}, wantErr: true, errIs: ErrInvalidOperandType},
		{name: `invalid comparison (nil <= num)`, op: "<=", left: NilValue{}, right: NumberValue{1}, wantErr: true, errIs: ErrInvalidOperandType},
		{name: `invalid bitwise (num & string)`, op: "&", left: NumberValue{1}, right: StringValue{"a"}, wantErr: true, errIs: ErrInvalidOperandTypeInteger},
		{name: `invalid bitwise (float | int)`, op: "|", left: NumberValue{1.5}, right: NumberValue{1}, wantErr: true, errIs: ErrInvalidOperandTypeInteger},
		{name: `invalid bitwise (nil ^ int)`, op: "^", left: NilValue{}, right: NumberValue{1}, wantErr: true, errIs: ErrInvalidOperandTypeInteger},
		{name: `unsupported operator`, op: "??", left: NumberValue{1}, right: NumberValue{2}, wantErr: true, errIs: ErrUnsupportedOperator},
		{name: `string repetition negative`, op: "*", left: StringValue{"a"}, right: NumberValue{-1}, wantErr: true, errIs: ErrInvalidOperandType},
		{name: `string repetition float`, op: "*", left: NumberValue{2.5}, right: StringValue{"a"}, wantErr: true, errIs: ErrInvalidOperandType},
		{name: `modulo by zero`, op: "%", left: NumberValue{5}, right: NumberValue{0}, wantErr: true, errIs: ErrDivisionByZero},
		{name: `modulo with float`, op: "%", left: NumberValue{5.5}, right: NumberValue{2}, wantErr: true, errIs: ErrInvalidOperandTypeInteger},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// DEBUG
			t.Logf("--- RUNNING TEST: %s ---", tc.name)
			t.Logf("DEBUG: Left: %#v, Op: %s, Right: %#v", tc.left, tc.op, tc.right)
			result, err := PerformBinaryOperation(tc.op, tc.left, tc.right)
			t.Logf("DEBUG: Result: %#v, Err: %v", result, err)

			if (err != nil) != tc.wantErr {
				t.Fatalf("PerformBinaryOperation() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr {
				if tc.errIs != nil && !errors.Is(err, tc.errIs) {
					t.Errorf("Expected error to be '%v', but got '%v'", tc.errIs, err)
				}
				return
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected result %T(%#v), but got %T(%#v)", tc.expected, tc.expected, result, result)
			}
		})
	}
}

func TestPerformUnaryOperation(t *testing.T) {
	testCases := []struct {
		name     string
		op       string
		operand  Value
		expected Value
		wantErr  bool
		errIs    error // ADDED
	}{
		{name: "not true", op: "not", operand: BoolValue{true}, expected: BoolValue{false}},
		{name: "not 0", op: "not", operand: NumberValue{0}, expected: BoolValue{true}},
		{name: "not non-empty string", op: "not", operand: StringValue{"a"}, expected: BoolValue{false}},
		{name: "negate number", op: "-", operand: NumberValue{5}, expected: NumberValue{-5}},
		{name: "some non-empty string", op: "some", operand: StringValue{"a"}, expected: BoolValue{true}},
		{name: "some empty string", op: "some", operand: StringValue{""}, expected: BoolValue{false}},
		{name: "no empty string", op: "no", operand: StringValue{""}, expected: BoolValue{true}},
		{name: "no non-empty string", op: "no", operand: StringValue{"a"}, expected: BoolValue{false}},
		{name: "no nil", op: "no", operand: NilValue{}, expected: BoolValue{true}},
		{name: "no *nil", op: "no", operand: &NilValue{}, expected: BoolValue{true}},

		// --- Error Cases ---
		{name: "negate string", op: "-", operand: StringValue{"a"}, wantErr: true, errIs: ErrInvalidOperandTypeNumeric},
		{name: "negate nil", op: "-", operand: NilValue{}, wantErr: true, errIs: ErrInvalidOperandTypeNumeric},
		{name: "negate bool", op: "-", operand: BoolValue{true}, wantErr: true, errIs: ErrInvalidOperandTypeNumeric},
		{name: "bitwise not string", op: "~", operand: StringValue{"a"}, wantErr: true, errIs: ErrInvalidOperandTypeInteger},
		{name: "bitwise not float", op: "~", operand: NumberValue{1.5}, wantErr: true, errIs: ErrInvalidOperandTypeInteger},
		{name: "bitwise not nil", op: "~", operand: NilValue{}, wantErr: true, errIs: ErrInvalidOperandTypeInteger},
		{name: "unknown operator", op: "!", operand: BoolValue{true}, wantErr: true, errIs: ErrUnsupportedOperator},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// DEBUG
			t.Logf("--- RUNNING TEST: %s ---", tc.name)
			t.Logf("DEBUG: Op: %s, Operand: %#v", tc.op, tc.operand)
			result, err := PerformUnaryOperation(tc.op, tc.operand)
			t.Logf("DEBUG: Result: %#v, Err: %v", result, err)

			if (err != nil) != tc.wantErr {
				t.Fatalf("PerformUnaryOperation() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr {
				if tc.errIs != nil && !errors.Is(err, tc.errIs) {
					t.Errorf("Expected error to be '%v', but got '%v'", tc.errIs, err)
				}
				return
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected result %T(%#v), but got %T(%#v)", tc.expected, tc.expected, result, result)
			}
		})
	}
}
