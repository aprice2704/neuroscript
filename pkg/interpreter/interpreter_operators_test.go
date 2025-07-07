// NeuroScript Version: 0.5.2
// File version: 1.0.0
// Purpose: Adds direct unit tests for the core operator functions (performArithmetic, performComparison, etc.).
// filename: pkg/interpreter/interpreter_operators_test.go
// nlines: 150
// risk_rating: LOW

package interpreter

import (
	"errors"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestPerformArithmetic(t *testing.T) {
	testCases := []struct {
		name    string
		left    lang.Value
		right   lang.Value
		op      string
		want    lang.Value
		wantErr error
	}{
		{"Subtract", lang.NumberValue{Value: 10}, lang.NumberValue{Value: 4}, "-", lang.NumberValue{Value: 6}, nil},
		{"Multiply", lang.NumberValue{Value: 5}, lang.NumberValue{Value: 3}, "*", lang.NumberValue{Value: 15}, nil},
		{"Divide", lang.NumberValue{Value: 20}, lang.NumberValue{Value: 4}, "/", lang.NumberValue{Value: 5}, nil},
		{"Power", lang.NumberValue{Value: 2}, lang.NumberValue{Value: 3}, "**", lang.NumberValue{Value: 8}, nil},
		{"Modulo", lang.NumberValue{Value: 10}, lang.NumberValue{Value: 3}, "%", lang.NumberValue{Value: 1}, nil},
		{"Division by zero", lang.NumberValue{Value: 10}, lang.NumberValue{Value: 0}, "/", nil, lang.ErrDivisionByZero},
		{"Invalid type", lang.StringValue{Value: "a"}, lang.NumberValue{Value: 1}, "*", nil, lang.ErrInvalidOperandType},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := performArithmetic(tc.left, tc.right, tc.op)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Expected error: %v, got: %v", tc.wantErr, err)
			}
			if tc.wantErr == nil && got != tc.want {
				t.Errorf("Expected result: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestPerformStringConcatOrNumericAdd(t *testing.T) {
	testCases := []struct {
		name  string
		left  lang.Value
		right lang.Value
		want  lang.Value
	}{
		{"Add numbers", lang.NumberValue{Value: 5}, lang.NumberValue{Value: 10}, lang.NumberValue{Value: 15}},
		{"Concat strings", lang.StringValue{Value: "hello "}, lang.StringValue{Value: "world"}, lang.StringValue{Value: "hello world"}},
		{"Concat string and number", lang.StringValue{Value: "age: "}, lang.NumberValue{Value: 30}, lang.StringValue{Value: "age: 30"}},
		{"Concat number and string", lang.NumberValue{Value: 30}, lang.StringValue{Value: " years"}, lang.StringValue{Value: "30 years"}},
		{"Concat with nil", lang.StringValue{Value: "value: "}, &lang.NilValue{}, lang.StringValue{Value: "value: "}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := performStringConcatOrNumericAdd(tc.left, tc.right)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("Expected result: %#v, got: %#v", tc.want, got)
			}
		})
	}
}

func TestPerformComparison(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(time.Second)

	testCases := []struct {
		name    string
		left    lang.Value
		right   lang.Value
		op      string
		want    lang.Value
		wantErr bool
	}{
		{"Equal numbers", lang.NumberValue{Value: 5}, lang.NumberValue{Value: 5}, "==", lang.BoolValue{Value: true}, false},
		{"Not equal strings", lang.StringValue{Value: "a"}, lang.StringValue{Value: "b"}, "!=", lang.BoolValue{Value: true}, false},
		{"Less than", lang.NumberValue{Value: 4}, lang.NumberValue{Value: 5}, "<", lang.BoolValue{Value: true}, false},
		{"Greater than or equal", lang.NumberValue{Value: 5}, lang.NumberValue{Value: 5}, ">=", lang.BoolValue{Value: true}, false},
		{"Time less than", lang.TimedateValue{Value: t1}, lang.TimedateValue{Value: t2}, "<", lang.BoolValue{Value: true}, false},
		{"Time greater than", lang.TimedateValue{Value: t2}, lang.TimedateValue{Value: t1}, ">", lang.BoolValue{Value: true}, false},
		{"Invalid comparison", lang.StringValue{Value: "a"}, lang.NumberValue{Value: 1}, ">", nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := performComparison(tc.left, tc.right, tc.op)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Unexpected error state. Got err: %v, wantErr: %t", err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("Expected result: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestPerformBitwise(t *testing.T) {
	testCases := []struct {
		name    string
		left    lang.Value
		right   lang.Value
		op      string
		want    lang.Value
		wantErr error
	}{
		{"AND", lang.NumberValue{Value: 5}, lang.NumberValue{Value: 3}, "&", lang.NumberValue{Value: 1}, nil}, // 101 & 011 = 001
		{"OR", lang.NumberValue{Value: 5}, lang.NumberValue{Value: 3}, "|", lang.NumberValue{Value: 7}, nil},  // 101 | 011 = 111
		{"XOR", lang.NumberValue{Value: 5}, lang.NumberValue{Value: 3}, "^", lang.NumberValue{Value: 6}, nil}, // 101 ^ 011 = 110
		{"Invalid type (float)", lang.NumberValue{Value: 5.5}, lang.NumberValue{Value: 3}, "&", nil, lang.ErrInvalidOperandTypeInteger},
		{"Invalid type (string)", lang.StringValue{Value: "a"}, lang.NumberValue{Value: 3}, "|", nil, lang.ErrInvalidOperandTypeInteger},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := performBitwise(tc.left, tc.right, tc.op)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Expected error: %v, got: %v", tc.wantErr, err)
			}
			if tc.wantErr == nil && got != tc.want {
				t.Errorf("Expected result: %v, got: %v", tc.want, got)
			}
		})
	}
}
