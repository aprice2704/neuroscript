// NeuroScript Version: 0.8.0
// File version: 25
// Purpose: Simplifies the clone method by removing the erroneous aiWorker assignment.
// filename: pkg/interpreter/clone.go
// nlines: 50
// risk_rating: HIGH

package interpreter

import (
	"fmt"

	"github.com/google/uuid"
)

// fork creates a new interpreter instance for sandboxing procedure calls or event handlers.
// It shares the immutable HostContext and a reference to the root interpreter.
// It creates an isolated variable scope and a context-aware tool registry view.
func (i *Interpreter) fork() *Interpreter {
	clone := &Interpreter{
		// Assign a new unique ID to the clone
		id: fmt.Sprintf("interp-%s", uuid.NewString()[:8]),

		// Share immutable or root-managed state by reference
		hostContext:         i.hostContext,
		root:                i.rootInterpreter(),
		eventManager:        i.eventManager,
		bufferManager:       i.bufferManager,
		objectCache:         i.objectCache,
		transientPrivateKey: i.transientPrivateKey,
		turnCtx:             i.turnCtx,
	}

	if i.tools != nil {
		clone.tools = i.tools.NewViewForInterpreter(clone)
	}

	// Create a new, isolated state for variables
	clone.state = newInterpreterState()
	clone.state.sandboxDir = i.state.sandboxDir
	clone.state.knownProcedures = i.state.knownProcedures

	// Copy global variables from the root into the new clone's scope
	root := clone.rootInterpreter()
	root.state.variablesMu.RLock()
	for name, value := range root.state.variables {
		if _, isGlobal := root.state.globalVarNames[name]; isGlobal {
			clone.state.variables[name] = value
			clone.state.globalVarNames[name] = true
		}
	}
	root.state.variablesMu.RUnlock()

	// Register the clone with the root for debugging purposes.
	root.cloneRegistryMu.Lock()
	root.cloneRegistry = append(root.cloneRegistry, clone)
	root.cloneRegistryMu.Unlock()

	return clone
}
