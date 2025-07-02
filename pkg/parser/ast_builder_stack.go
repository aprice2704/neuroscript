// NeuroScript Version: 0.3.0
// File version: 2
// Purpose: Fixes a compiler error by aligning legacy pop/popN wrappers to a single return value.
// filename: pkg/parser/ast_builder_stack.go
// nlines: 55
// risk_rating: LOW

package parser

// --- Value Stack Helpers ---

// push is the core implementation for pushing a value onto the stack.
func (l *neuroScriptListenerImpl) push(v interface{}) {
	if v == nil {
		l.addErrorf(nil, "internal error: attempt to push nil value onto AST value stack")
		return
	}
	l.ValueStack = append(l.ValueStack, v)
	l.logDebugAST("pushed %T onto value stack (size: %d)", v, len(l.ValueStack))
}

// pop is the core implementation for popping a value from the stack.
// It returns the value and a boolean indicating if the pop was successful.
func (l *neuroScriptListenerImpl) pop() (interface{}, bool) {
	if len(l.ValueStack) < 1 {
		l.logger.Error("AST Builder FATAL: value stack underflow on pop")
		return nil, false
	}
	popped := l.ValueStack[len(l.ValueStack)-1]
	l.ValueStack = l.ValueStack[:len(l.ValueStack)-1]
	l.logDebugAST("popped %T from value stack (new size: %d)", popped, len(l.ValueStack))
	return popped, true
}

// popN is the core implementation for popping N values.
func (l *neuroScriptListenerImpl) popN(n int) ([]interface{}, bool) {
	if len(l.ValueStack) < n {
		l.logger.Error("AST Builder FATAL: value stack underflow on popN", "requested", n, "available", len(l.ValueStack))
		return nil, false
	}
	index := len(l.ValueStack) - n
	popped := l.ValueStack[index:]
	l.ValueStack = l.ValueStack[:index]
	l.logDebugAST("popped %d values from value stack (new size: %d)", n, len(l.ValueStack))
	return popped, true
}
