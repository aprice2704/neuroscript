// NeuroScript Version: 0.3.1
// File version: 1
// Purpose: Contains the value stack manipulation methods for the AST builder listener.
// filename: pkg/core/ast_builder_stack.go
// nlines: 75
// risk_rating: MEDIUM

package core

import (
	"errors"
	"fmt"
)

// pushValue pushes a value onto the listener's value stack.
func (l *neuroScriptListenerImpl) pushValue(v interface{}) {
	if l.debugAST {
		l.logger.Debug("[DEBUG-AST-STACK] --> PUSH", "value_type", fmt.Sprintf("%T", v), "new_stack_size", len(l.valueStack)+1)
	}
	l.valueStack = append(l.valueStack, v)
}

// popValue pops a single value from the listener's value stack.
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

// popNValues pops N values from the listener's value stack.
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
