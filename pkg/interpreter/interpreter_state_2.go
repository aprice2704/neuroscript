// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Contains internal state and event management structs for the Interpreter.
// filename: pkg/interpreter/state.go
// nlines: 54
// risk_rating: MEDIUM

package interpreter

import (
	"sync"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// interpreterState holds the non-exported state of the interpreter.
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
	globalVarNames    map[string]bool
}

// EventManager handles event subscriptions and emissions.
type EventManager struct {
	eventHandlers   map[string][]*ast.OnEventDecl
	eventHandlersMu sync.RWMutex
}

func newInterpreterState() *interpreterState {
	return &interpreterState{
		variables:       make(map[string]lang.Value),
		knownProcedures: make(map[string]*ast.Procedure),
		commands:        []*ast.CommandNode{},
		stackFrames:     []string{},
		globalVarNames:  make(map[string]bool),
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

	eventName, err := i.evaluate.Expression(decl.EventNameExpr)
	if err != nil {
		return lang.WrapErrorWithPosition(err, decl.EventNameExpr.GetPos(), "evaluating event name expression")
	}

	eventNameStr, _ := lang.ToString(eventName)
	em.eventHandlers[eventNameStr] = append(em.eventHandlers[eventNameStr], decl)
	return nil
}
