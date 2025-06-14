// NeuroScript Version: 0.3.0
// File version: 2
// Purpose: Fixes a compiler error by aligning legacy pop/popN wrappers to a single return value.
// filename: pkg/core/ast_builder_stack.go
// nlines: 55
// risk_rating: LOW

package core

// --- Value Stack Helpers ---

// pushValue is the core implementation for pushing a value onto the stack.
func (l *neuroScriptListenerImpl) pushValue(v interface{}) {
	if v == nil {
		l.addErrorf(nil, "internal error: attempt to push nil value onto AST value stack")
		return
	}
	l.valueStack = append(l.valueStack, v)
	l.logDebugAST("pushed %T onto value stack (size: %d)", v, len(l.valueStack))
}

// popValue is the core implementation for popping a value from the stack.
// It returns the value and a boolean indicating if the pop was successful.
func (l *neuroScriptListenerImpl) popValue() (interface{}, bool) {
	if len(l.valueStack) < 1 {
		l.logger.Error("AST Builder FATAL: value stack underflow on pop")
		return nil, false
	}
	popped := l.valueStack[len(l.valueStack)-1]
	l.valueStack = l.valueStack[:len(l.valueStack)-1]
	l.logDebugAST("popped %T from value stack (new size: %d)", popped, len(l.valueStack))
	return popped, true
}

// popNValues is the core implementation for popping N values.
func (l *neuroScriptListenerImpl) popNValues(n int) ([]interface{}, bool) {
	if len(l.valueStack) < n {
		l.logger.Error("AST Builder FATAL: value stack underflow on popN", "requested", n, "available", len(l.valueStack))
		return nil, false
	}
	index := len(l.valueStack) - n
	popped := l.valueStack[index:]
	l.valueStack = l.valueStack[:index]
	l.logDebugAST("popped %d values from value stack (new size: %d)", n, len(l.valueStack))
	return popped, true
}

// --- Wrapper Methods for Legacy Naming ---

// push is a wrapper for pushValue.
func (l *neuroScriptListenerImpl) push(v interface{}) {
	l.pushValue(v)
}

// pop is a wrapper for popValue that discards the success flag for legacy callers.
func (l *neuroScriptListenerImpl) pop() interface{} {
	v, _ := l.popValue()
	return v
}

// popN is a wrapper for popNValues that discards the success flag for legacy callers.
func (l *neuroScriptListenerImpl) popN(n int) []interface{} {
	v, _ := l.popNValues(n)
	return v
}
