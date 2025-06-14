// NeuroScript Version: 0.3.5
// File version: 9
// Purpose: Reverted tests to expect native bool types.
// filename: pkg/core/evaluation_comparison_test.go

package core

import (
	"errors"
	"testing"
)

func TestEvaluateCondition(t *testing.T) {
	vars := map[string]interface{}{
		"trueVar":  BoolValue{Value: true},
		"falseVar": BoolValue{Value: false},
		"num10":    NumberValue{Value: 10},
		"strTrue":  StringValue{Value: "true"},
		"strFalse": StringValue{Value: "false"},
		"strOther": StringValue{Value: "hello"},
		"nilVar":   NilValue{},
	}

	testCases := []struct {
		name            string
		node            Expression
		vars            map[string]interface{}
		last            interface{}
		want            bool
		wantErr         bool
		expectedErrorIs error
	}{
		{name: "Boolean Literal True", node: &BooleanLiteralNode{Value: true}, want: true},
		{name: "Var Boolean True", node: &VariableNode{Name: "trueVar"}, vars: vars, want: true},
		{name: "String Literal True", node: &StringLiteralNode{Value: "true"}, want: true},
		{name: "String Literal Other", node: &StringLiteralNode{Value: "yes"}, want: false},
		{name: "Var String False", node: &VariableNode{Name: "strFalse"}, vars: vars, want: false},
		{
			name:            "Comp Numeric Error Types",
			node:            &BinaryOpNode{Left: &StringLiteralNode{Value: "a"}, Operator: ">", Right: &NumberLiteralNode{Value: 5}},
			wantErr:         true,
			expectedErrorIs: ErrInvalidOperandType,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interpreter, _ := NewTestInterpreter(t, tc.vars, tc.last)
			got, err := interpreter.evaluateCondition(tc.node)

			if (err != nil) != tc.wantErr {
				t.Errorf("evaluateCondition(%s): Error expectation mismatch. got err = %v, wantErr %t", tc.name, err, tc.wantErr)
			} else if tc.wantErr && tc.expectedErrorIs != nil {
				if !errors.Is(err, tc.expectedErrorIs) {
					t.Errorf("evaluateCondition(%s): Expected error to wrap [%v], but got [%v]", tc.name, tc.expectedErrorIs, err)
				}
			}

			if err == nil && got != tc.want {
				t.Errorf("evaluateCondition(%s)\n    Node:       %s\n    Got bool:   %v\n    Want bool:  %v", tc.name, tc.node, got, tc.want)
			}
		})
	}
}
