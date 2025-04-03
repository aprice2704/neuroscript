// pkg/core/ast_builder_main.go
package core

import (
	"io"
	"log"

	// "strconv" // Not needed directly here
	// "strings" // Not needed directly here
	"github.com/antlr4-go/antlr/v4" // Import antlr
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// neuroScriptListenerImpl builds the AST.
type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	procedures     []Procedure
	currentProc    *Procedure
	currentSteps   *[]Step            // Pointer to the current list of steps being built
	blockStepStack []*[]Step          // Stack for managing nested block contexts (stores pointers to parent step lists)
	valueStack     []interface{}      // Stack holds expression AST nodes
	currentMapKey  *StringLiteralNode // Temp storage for map key node

	// *** NEW: Map to store collected steps for specific block contexts ***
	blockSteps map[antlr.ParserRuleContext][]Step

	logger   *log.Logger
	debugAST bool
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
		// *** Initialize the new map ***
		blockSteps: make(map[antlr.ParserRuleContext][]Step),
	}
}

// --- Stack Helper Methods (pushValue, popValue, popNValues) remain the same ---
func (l *neuroScriptListenerImpl) pushValue(v interface{}) {
	l.valueStack = append(l.valueStack, v)
	l.logDebugAST("    Pushed Value: %T %+v (Stack size: %d)", v, v, len(l.valueStack))
}
func (l *neuroScriptListenerImpl) popValue() (interface{}, bool) {
	if len(l.valueStack) == 0 {
		l.logger.Println("[ERROR] AST Builder: Pop from empty value stack!")
		return nil, false
	}
	index := len(l.valueStack) - 1
	value := l.valueStack[index]
	l.valueStack = l.valueStack[:index]
	l.logDebugAST("    Popped Value: %T %+v (Stack size: %d)", value, value, len(l.valueStack))
	return value, true
}
func (l *neuroScriptListenerImpl) popNValues(n int) ([]interface{}, bool) {
	if len(l.valueStack) < n {
		l.logger.Printf("[ERROR] AST Builder: Stack underflow pop %d, have %d.", n, len(l.valueStack))
		return nil, false
	}
	startIndex := len(l.valueStack) - n
	values := make([]interface{}, n)
	copy(values, l.valueStack[startIndex:])
	l.valueStack = l.valueStack[:startIndex]
	l.logDebugAST("    Popped %d Values (Stack size: %d)", n, len(l.valueStack))
	return values, true
}

// --- Core Listener Methods ---
func (l *neuroScriptListenerImpl) GetResult() []Procedure { return l.procedures }
func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Printf(format, v...)
	}
}
func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) {
	l.logDebugAST(">>> Enter Program")
	l.procedures = make([]Procedure, 0)
}
func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
	l.logDebugAST("<<< Exit Program")
}
