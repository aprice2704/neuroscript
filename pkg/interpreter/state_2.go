// NeuroScript Version: 0.8.0
// File version: 10
// Purpose: Adds globalConstants map to interpreterState, fixing compile errors.
// filename: pkg/interpreter/state_2.go
// nlines: 71
// risk_rating: MEDIUM

package interpreter

import (
	"sync"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/eval"
	"github.com/aprice2704/neuroscript/pkg/lang"
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
	globalVarNames    map[string]bool
	globalConstants   map[string]lang.Value // ADDED: For tool-defined global constants

	// --- Provider State (Root Only) ---
	// REMOVED: providers map and providersMu
	// This is now handled by the root-level, injected provider.Registry.
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
		globalConstants: make(map[string]lang.Value), // ADDED: Initialize the map
		// REMOVED: providers map initialization
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

// setGlobalVariable sets a variable and marks it as global in a thread-safe manner.
func (s *interpreterState) setGlobalVariable(name string, value lang.Value) {
	s.variablesMu.Lock()
	defer s.variablesMu.Unlock()
	if s.variables == nil {
		s.variables = make(map[string]lang.Value)
	}
	if s.globalVarNames == nil {
		s.globalVarNames = make(map[string]bool)
	}
	s.variables[name] = value
	s.globalVarNames[name] = true
}

func newEventManager() *EventManager {
	return &EventManager{
		eventHandlers: make(map[string][]*ast.OnEventDecl),
	}
}

func (em *EventManager) register(decl *ast.OnEventDecl, i *Interpreter) error {
	em.eventHandlersMu.Lock()
	defer em.eventHandlersMu.Unlock()

	eventName, err := eval.Expression(i, decl.EventNameExpr)
	if err != nil {
		return lang.WrapErrorWithPosition(err, decl.EventNameExpr.GetPos(), "evaluating event name expression")
	}

	eventNameStr, _ := lang.ToString(eventName)
	em.eventHandlers[eventNameStr] = append(em.eventHandlers[eventNameStr], decl)
	return nil
}
