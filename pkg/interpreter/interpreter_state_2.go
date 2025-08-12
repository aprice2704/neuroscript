// NeuroScript Version: 0.6.0
// File version: 5.0.0
// Purpose: Consolidated AgentModel state directly into the main interpreterState struct, resolving missing field errors.
// filename: pkg/interpreter/interpreter_state_2.go
// nlines: 70
// risk_rating: MEDIUM

package interpreter

import (
	"sync"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/types"
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

	// --- AgentModel State ---
	// FIX: The agentModels map is now correctly defined here.
	agentModels   map[types.AgentModelName]AgentModel
	agentModelsMu sync.RWMutex

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
		variables:       make(map[string]lang.Value),
		knownProcedures: make(map[string]*ast.Procedure),
		commands:        []*ast.CommandNode{},
		stackFrames:     []string{},
		globalVarNames:  make(map[string]bool),
		agentModels:     make(map[types.AgentModelName]AgentModel), // Initialize agentModels map
		providers:       make(map[string]provider.AIProvider),
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
