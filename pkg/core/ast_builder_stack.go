// ast_builder_stack.go - Stack manipulation helpers for the AST builder.
//
// These functions operate on the two runtime stacks maintained by
// neuroScriptListenerImpl while the ANTLR listener walks the parse tree:
//
//   - valueStack         – []interface{} acting as a LIFO for every
//     value‑producing grammar construct
//     (expressions, []Step blocks, literals, etc.).
//
//   - blockStepStack     – []*[]Step used to collect the Steps belonging to the
//     body of each nested block (func, on event, if, loop…).
//
// Invariants enforced (see ast.go for full specification):
//
//  1. Every pushValue must eventually be matched by a popValue (or
//     equivalent popNValues) before the listener returns to the top level.
//  2. popNValues(n) removes exactly n elements or records an internal error
//     if the stack has fewer than n entries.
//  3. These helpers never touch blockStepStack — block entry/exit helpers
//     manage it separately.
//
// Not safe for concurrent use; the AST builder is single‑threaded.
//
// file version: 1
package core

import (
	"errors"
	"fmt"
)

// pushValue appends v to the valueStack.
func (l *neuroScriptListenerImpl) pushValue(v interface{}) {
	if l.debugAST {
		l.logger.Debug("[DEBUG-AST-STACK] --> PUSH", "value_type", fmt.Sprintf("%T", v), "new_stack_size", len(l.valueStack)+1)
	}
	l.valueStack = append(l.valueStack, v)
}

// popValue pops and returns the top element of valueStack.
func (l *neuroScriptListenerImpl) popValue() (interface{}, bool) {
	if len(l.valueStack) == 0 {
		l.logger.Error("AST Builder: Pop from empty value stack!")
		l.errors = append(l.errors, errors.New("AST builder internal error: attempted pop from empty value stack"))
		return nil, false
	}
	index := len(l.valueStack) - 1
	value := l.valueStack[index]
	l.valueStack = l.valueStack[:index]
	if l.debugAST {
		l.logger.Debug("[DEBUG-AST-STACK] <-- POP", "value_type", fmt.Sprintf("%T", value), "new_stack_size", len(l.valueStack))
	}
	return value, true
}

// popNValues pops n elements and returns them in the same order they were on the stack.
func (l *neuroScriptListenerImpl) popNValues(n int) ([]interface{}, bool) {
	if n < 0 {
		l.addError(nil, "internal AST builder error: popNValues called with negative count %d", n)
		return nil, false
	}
	if n == 0 {
		return []interface{}{}, true
	}
	if len(l.valueStack) < n {
		l.addError(nil, "internal AST builder error: stack underflow, needed %d values, only have %d", n, len(l.valueStack))
		return nil, false
	}

	startIndex := len(l.valueStack) - n
	values := make([]interface{}, n)
	copy(values, l.valueStack[startIndex:])
	l.valueStack = l.valueStack[:startIndex]
	if l.debugAST {
		l.logger.Debug("[DEBUG-AST-STACK] <-- POP N", "count", n, "new_stack_size", len(l.valueStack))
	}
	return values, true
}
