// NeuroScript Version: 0.8.0
// File version: 27
// Purpose: Corrected the fork method to properly share all root-level resources by reference, fixing numerous test failures related to policy, state, and parser configuration.
// filename: pkg/interpreter/clone.go
// nlines: 55
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
	root := i.rootInterpreter()
	clone := &Interpreter{
		// Assign a new unique ID to the clone
		id: fmt.Sprintf("interp-%s", uuid.NewString()[:8]),

		// Share immutable or root-managed state by reference
		hostContext:          i.hostContext,
		root:                 root,
		eventManager:         i.eventManager,
		bufferManager:        i.bufferManager,
		objectCache:          i.objectCache,
		transientPrivateKey:  i.transientPrivateKey,
		turnCtx:              i.turnCtx,
		maxLoopIterations:    i.maxLoopIterations,
		modelStore:           i.modelStore,
		ExecPolicy:           i.ExecPolicy,
		accountStore:         i.accountStore,
		capsuleStore:         i.capsuleStore,
		adminCapsuleRegistry: i.adminCapsuleRegistry,
		parser:               i.parser,
		astBuilder:           i.astBuilder,
		aiWorker:             i.aiWorker,
	}

	if i.tools != nil {
		clone.tools = i.tools.NewViewForInterpreter(clone)
	}

	// Create a new, isolated state for variables
	clone.state = newInterpreterState()
	clone.state.sandboxDir = i.state.sandboxDir
	// FIX: Procedures from the parent/root must be available to the fork.
	clone.state.knownProcedures = i.state.knownProcedures

	// Copy global variables from the root into the new clone's scope
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
