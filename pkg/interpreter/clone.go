// NeuroScript Version: 0.8.0
// File version: 39
// Purpose: Ensures the root providerRegistry is correctly propagated to forks.
// filename: pkg/interpreter/clone.go
// nlines: 75
// risk_rating: HIGH

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
	// fmt.Printf("[DEBUG] fork: Creating fork from interpreter %s (root: %s). Parent turnCtx is %p.\n", i.id, root.id, i.GetTurnContext())

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
		providerRegistry:     i.providerRegistry, // <-- ADDED (Task p1-clone)
		ExecPolicy:           i.ExecPolicy,
		accountStore:         i.accountStore,
		capsuleStore:         i.capsuleStore,
		adminCapsuleRegistry: i.adminCapsuleRegistry,
		parser:               i.parser,
		astBuilder:           i.astBuilder,
		aiWorker:             i.aiWorker,
		// THE FIX: STEP 3a - Ensure the clone inherits the back-reference.
		PublicAPI: i.PublicAPI,
	}
	// DEBUG: Print the memory address of the newly created clone.
	// fmt.Printf("[DEBUG] fork: New clone %s created at pointer %p\n", clone.id, clone)

	// THE FIX: STEP 3b - Create a tool registry VIEW bound to the PUBLIC API WRAPPER, not the internal clone.
	// This guarantees that tools called from a sandbox still receive the identity-aware runtime.
	if clone.PublicAPI != nil {
		clone.tools = i.tools.NewViewForInterpreter(clone.PublicAPI)
		// fmt.Printf("[DEBUG] fork: Created new tool registry view for clone %s, bound to PublicAPI wrapper.\n", clone.id)
	} else {
		// Fallback for safety, though this path should not be taken in normal operation.
		clone.tools = i.tools.NewViewForInterpreter(clone)
		// fmt.Printf("[DEBUG] fork: (Fallback) Created new tool registry view for clone %s, bound to internal clone.\n", clone.id)
	}

	// Create a new, isolated state, but inherit key properties.
	clone.state = newInterpreterState()
	clone.state.sandboxDir = i.state.sandboxDir
	clone.state.knownProcedures = i.state.knownProcedures

	// CRITICAL: The clone must inherit the parent's context.
	clone.SetTurnContext(i.GetTurnContext())
	// fmt.Printf("[DEBUG] fork: Cloned interpreter %s created. Clone turnCtx is %p.\n", clone.id, clone.GetTurnContext())

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
