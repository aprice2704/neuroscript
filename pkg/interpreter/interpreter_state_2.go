// NeuroScript Version: 0.8.0
// File version: 8.0.1
// Purpose: FIX: Uses the public EvaluateExpression method instead of the private evaluate field.
// filename: pkg/interpreter/interpreter_state_2.go
// nlines: 55
// risk_rating: LOW

package interpreter

import (
	"sync"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// interpreterState holds all the non-exported, mutable state of the interpreter.
type interpreterState struct {
	variables         map[string]lang.Value
	variablesMu       sync.RWMutex
	knownProcedures   map[string]*ast.Procedure
	commands          []*ast.CommandNode
	stackFrames       []string
	currentProcName   string
	errorHandlerStack [][]*ast.Step
	sandboxDir        string
	vectorIndex       map[string][]float32
	maxLoopIterations int // Moved from Interpreter struct

	// --- Provider State ---
	providers   map[string]provider.AIProvider
	providersMu sync.RWMutex
}

// EventManager handles event subscriptions and emissions.
type EventManager struct {
	eventHandlers   map[string][]*ast.OnEventDecl
	eventHandlersMu sync.RWMutex
}

func newInterpreterState() *interpreterState {
	return &interpreterState{
		variables:         make(map[string]lang.Value),
		knownProcedures:   make(map[string]*ast.Procedure),
		commands:          []*ast.CommandNode{},
		stackFrames:       []string{},
		providers:         make(map[string]provider.AIProvider),
		maxLoopIterations: 1000, // Default value
	}
}

func (s *interpreterState) setVariable(name string, value lang.Value) {
	s.variablesMu.Lock()
	defer s.variablesMu.Unlock()
	if s.variables == nil {
		s.variables = make(map[string]lang.Value)
	}
	s.variables[name] = value
}

func newEventManager() *EventManager {
	return &EventManager{
		eventHandlers: make(map[string][]*ast.OnEventDecl),
	}
}

func (em *EventManager) register(decl *ast.OnEventDecl, i *Interpreter) error {
	em.eventHandlersMu.Lock()
	defer em.eventHandlersMu.Unlock()

	eventName, err := i.EvaluateExpression(decl.EventNameExpr)
	if err != nil {
		return lang.WrapErrorWithPosition(err, decl.EventNameExpr.GetPos(), "evaluating event name expression")
	}

	eventNameStr, _ := lang.ToString(eventName)
	em.eventHandlers[eventNameStr] = append(em.eventHandlers[eventNameStr], decl)
	return nil
}
