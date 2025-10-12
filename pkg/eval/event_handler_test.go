// NeuroScript Version: 0.8.0
// File version: 16
// Purpose: Refactored to be self-contained within the eval package for isolated testing.
// filename: pkg/eval/event_handler_test.go
// nlines: 50
// risk_rating: LOW

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestEventHandlerDynamicName(t *testing.T) {
	// This test is now simplified. The goal is to ensure that an identifier
	// used as an event name can be evaluated correctly. The registration
	// logic itself is part of the interpreter, not the evaluator.
	vars := map[string]lang.Value{
		"my_event": lang.StringValue{Value: "some_event"},
	}

	tc := localEvalTestCase{
		Name:        "Event name can be a variable",
		InputNode:   &ast.VariableNode{Name: "my_event"},
		InitialVars: vars,
		Expected:    lang.StringValue{Value: "some_event"},
	}

	runLocalExpressionTest(t, tc)
}
