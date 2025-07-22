// NeuroScript Version: 0.6.0
// File version: 2
// Purpose: Corrected constructor call for newNeuroScriptListener.
// filename: pkg/parser/ast_builder_stack_explicit_test.go
// nlines: 55
// risk_rating: LOW

package parser

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
)

// TestStackPopNOrder explicitly verifies the LIFO (Last-In, First-Out)
// behavior of the popN function and the correctness of reversing its output
// to restore the original source-code order. This test is crucial to
// prevent future regressions or "flip-flopping" on this logic.
func TestStackPopNOrder(t *testing.T) {
	// Setup: Push items in source order.
	listener := newNeuroScriptListener(logging.NewNoOpLogger(), false, nil)
	listener.push("first")  // Pushed first
	listener.push("second") // Pushed second
	listener.push("third")  // Pushed third (last)

	// Action: Pop all three items.
	popped, ok := listener.popN(3)
	if !ok {
		t.Fatal("popN(3) failed unexpectedly")
	}

	// Verification 1: Check the raw popped order (should be LIFO).
	expectedPoppedOrder := []interface{}{"first", "second", "third"}
	if !reflect.DeepEqual(popped, expectedPoppedOrder) {
		t.Errorf("popN did not return items in the expected raw stack order.\n- want: %v\n-  got: %v", expectedPoppedOrder, popped)
	}

	// Verification 2: Check the source-restored order (after reversing).
	// This is how consumers like 'return' should process the result.
	sourceOrder := make([]interface{}, len(popped))
	for i := 0; i < len(popped); i++ {
		sourceOrder[i] = popped[len(popped)-1-i]
	}

	expectedSourceOrder := []interface{}{"third", "second", "first"}
	if !reflect.DeepEqual(sourceOrder, expectedSourceOrder) {
		t.Errorf("Reversing the popped slice did not restore source order correctly.\n- want: %v\n-  got: %v", expectedSourceOrder, sourceOrder)
	}
}
