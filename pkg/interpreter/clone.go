// NeuroScript Version: 0.8.0
// File version: 36
// Purpose: Corrected a typo in a Printf format string that was causing a build failure.
// filename: pkg/interpreter/clone.go
// nlines: 70
// risk_rating: LOW

package interpreter

import (
	"fmt"

	"github.com/google/uuid"
)

// fork creates a new interpreter instance for sandboxing procedure calls or event handlers.
// It shares the immutable HostContext and a reference to the root interpreter.
// It creates an isolated variable scope and a new tool registry view bound to itself.
func (i *Interpreter) fork() *Interpreter {
	root := i.rootInterpreter()
	// THE FIX: Corrected the argument order for the format string.
	fmt.Printf("[DEBUG] fork: Creating fork from interpreter %s (root: %s). Parent turnCtx is %p.\n", i.id, root.id, i.GetTurnContext())

	clone := &Interpreter{
		id:   fmt.Sprintf("interp-%s", uuid.NewString()[:8]),
		root: root, // Explicitly set the root
		// Copy/share all other fields from parent 'i'
		hostContext:          i.hostContext,
		eventManager:         i.eventManager,
		bufferManager:        i.bufferManager,
		objectCache:          i.objectCache,
		transientPrivateKey:  i.transientPrivateKey,
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

	// Create a new tool registry VIEW bound to the CLONE.
	// Do not share the parent's registry object directly, as its internal
	// 'interpreter' reference would be stale.
	clone.tools = i.tools.NewViewForInterpreter(clone)
	fmt.Printf("[DEBUG] fork: Created new tool registry view for clone %s.\n", clone.id)

	// Create a new, isolated state, but inherit key properties.
	clone.state = newInterpreterState()
	clone.state.sandboxDir = i.state.sandboxDir
	clone.state.knownProcedures = i.state.knownProcedures

	// CRITICAL: The clone must inherit the parent's context.
	clone.SetTurnContext(i.GetTurnContext())
	fmt.Printf("[DEBUG] fork: Cloned interpreter %s created. Clone turnCtx is %p.\n", clone.id, clone.GetTurnContext())

	// Copy global variables from the root.
	root.state.variablesMu.RLock()
	for name, value := range root.state.variables {
		if _, isGlobal := root.state.globalVarNames[name]; isGlobal {
			clone.state.variables[name] = value
			clone.state.globalVarNames[name] = true
		}
	}
	root.state.variablesMu.RUnlock()

	// Register with root for debugging.
	root.cloneRegistryMu.Lock()
	root.cloneRegistry = append(root.cloneRegistry, clone)
	root.cloneRegistryMu.Unlock()

	return clone
}
