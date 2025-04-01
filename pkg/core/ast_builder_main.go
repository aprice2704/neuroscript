// pkg/core/ast_builder_main.go
package core

import (
	// Import fmt for debug logging
	"io"
	"log"

	// "strconv" // Not needed directly here
	// "strings" // Not needed directly here

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// neuroScriptListenerImpl builds the AST using a stack for expression nodes.
type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	procedures     []Procedure
	currentProc    *Procedure
	currentSteps   *[]Step            // Pointer to the current list of steps being built (could be procedure body or block body)
	blockStepStack []*[]Step          // Stack for managing nested block bodies ([][]Step)
	valueStack     []interface{}      // Stack holds expression AST nodes (VariableNode, LiteralNode, ConcatenationNode, etc.)
	currentMapKey  *StringLiteralNode // Temp storage for map key node during entry parsing
	logger         *log.Logger
	debugAST       bool
}

// newNeuroScriptListener creates a new listener instance.
func newNeuroScriptListener(logger *log.Logger, debugAST bool) *neuroScriptListenerImpl {
	if logger == nil {
		logger = log.New(io.Discard, "", 0) // Default to discarding logs if none provided
	}
	return &neuroScriptListenerImpl{
		procedures:     make([]Procedure, 0),
		blockStepStack: make([]*[]Step, 0),
		valueStack:     make([]interface{}, 0, 10), // Initialize with some capacity
		logger:         logger,
		debugAST:       debugAST,
	}
}

// --- Stack Helper Methods ---

// pushValue pushes an AST node onto the value stack.
func (l *neuroScriptListenerImpl) pushValue(v interface{}) {
	l.valueStack = append(l.valueStack, v)
	l.logDebugAST("    Pushed Value: %T %+v (Stack size: %d)", v, v, len(l.valueStack))
}

// popValue pops the top AST node from the value stack.
func (l *neuroScriptListenerImpl) popValue() (interface{}, bool) {
	if len(l.valueStack) == 0 {
		l.logger.Println("[ERROR] AST Builder: Attempted to pop from empty value stack!")
		return nil, false
	}
	index := len(l.valueStack) - 1
	value := l.valueStack[index]
	l.valueStack = l.valueStack[:index] // Pop
	l.logDebugAST("    Popped Value: %T %+v (Stack size: %d)", value, value, len(l.valueStack))
	return value, true
}

// popNValues pops N values, returning them in the order they were pushed.
func (l *neuroScriptListenerImpl) popNValues(n int) ([]interface{}, bool) {
	if len(l.valueStack) < n {
		l.logger.Printf("[ERROR] AST Builder: Stack underflow. Tried to pop %d values, only have %d.", n, len(l.valueStack))
		return nil, false
	}
	startIndex := len(l.valueStack) - n
	values := make([]interface{}, n)
	copy(values, l.valueStack[startIndex:])  // Copy the required slice
	l.valueStack = l.valueStack[:startIndex] // Truncate the stack
	l.logDebugAST("    Popped %d Values (Stack size: %d)", n, len(l.valueStack))
	return values, true // Return in original push order
}

// --- Core Listener Methods ---

func (l *neuroScriptListenerImpl) GetResult() []Procedure {
	return l.procedures
}

func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Printf(format, v...)
	}
}

// Enter/Exit Program remain the same
func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) {
	l.logDebugAST(">>> Enter Program")
	l.procedures = make([]Procedure, 0)
}
func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
	l.logDebugAST("<<< Exit Program")
}
