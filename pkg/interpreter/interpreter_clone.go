// NeuroScript Version: 0.7.2
// File version: 20
// Purpose: FIX: Corrects a critical bug by ensuring the adminCapsuleRegistry is propagated to cloned interpreters. ADDED EXTENSIVE DEBUGGING.
// filename: pkg/interpreter/interpreter_clone.go

package interpreter

import (
	"fmt"

	"github.com/google/uuid"
)

// clone creates a new interpreter instance for sandboxing.
// It shares persistent state (stores, tools, config) by reference,
// but creates an isolated variable scope and a context-aware tool registry view.
func (i *Interpreter) clone() *Interpreter {
	// --- MORE DEBUGGING ---
	fmt.Printf("\n[DEBUG] >>> CLONING INTERPRETER (Parent ID: %s) <<<\n", i.id)
	if i.adminCapsuleRegistry != nil {
		fmt.Printf("[DEBUG] PARENT %s: Has a NON-NIL admin registry at clone time.\n", i.id)
	} else {
		fmt.Printf("[DEBUG] PARENT %s: Has a NIL admin registry at clone time. THIS IS LIKELY THE CAUSE OF THE BUG.\n", i.id)
	}
	// --- END DEBUGGING ---

	clone := &Interpreter{
		// Assign a new unique ID to the clone
		id: fmt.Sprintf("interp-%s", uuid.NewString()[:8]),

		// Share core state by reference for persistence
		logger:       i.logger,
		fileAPI:      i.fileAPI,
		eventManager: i.eventManager,
		aiWorker:     i.aiWorker,
		stdout:       i.stdout,
		stdin:        i.stdin,
		stderr:       i.stderr,
		// --- BUG FIX ---
		// The admin capsule registry must be copied from the parent to the clone.
		// Without this, privileged tools like 'tool.capsule.Add' will fail
		// when executed in a sandboxed procedure or command block.
		adminCapsuleRegistry:      i.adminCapsuleRegistry,
		maxLoopIterations:         i.maxLoopIterations,
		bufferManager:             i.bufferManager,
		objectCache:               i.objectCache,
		llmclient:                 i.llmclient,
		skipStdTools:              i.skipStdTools,
		modelStore:                i.modelStore,
		accountStore:              i.accountStore,
		capsuleStore:              i.capsuleStore,
		ExecPolicy:                i.ExecPolicy,
		root:                      i.rootInterpreter(),
		aiTranscript:              i.aiTranscript,
		transientPrivateKey:       i.transientPrivateKey,
		turnCtx:                   i.turnCtx,
		customEmitFunc:            i.customEmitFunc,
		customWhisperFunc:         i.customWhisperFunc,
		eventHandlerErrorCallback: i.eventHandlerErrorCallback,
		emitter:                   i.emitter,
	}

	// --- MORE DEBUGGING ---
	if clone.adminCapsuleRegistry != nil {
		fmt.Printf("[DEBUG] CLONE %s: Has a NON-NIL admin registry after assignment.\n", clone.id)
	} else {
		fmt.Printf("[DEBUG] CLONE %s: Has a NIL admin registry after assignment.\n", clone.id)
	}
	fmt.Printf("[DEBUG] >>> CLONING COMPLETE (Clone ID: %s) <<<\n\n", clone.id)
	// --- END DEBUGGING ---

	if i.tools != nil {
		clone.tools = i.tools.NewViewForInterpreter(clone)
	}

	clone.state = newInterpreterState()
	clone.state.sandboxDir = i.state.sandboxDir

	clone.state.providers = i.state.providers
	clone.state.knownProcedures = i.state.knownProcedures

	root := clone.rootInterpreter()

	// FAIL-FAST: A properly constructed interpreter always has a root.
	if root == nil {
		panic(fmt.Sprintf(
			"FATAL: Interpreter (ID: %s) has a nil root. This should not be possible.",
			clone.id,
		))
	}

	// Add the clone to the root's registry.
	root.cloneRegistryMu.Lock()
	defer root.cloneRegistryMu.Unlock()

	// FAIL-FAST: If the root's registry is nil, it means the root interpreter
	// was not created via NewInterpreter(). This is a critical state corruption.
	if root.cloneRegistry == nil {
		panic(fmt.Sprintf(
			"FATAL: The root interpreter (ID: %s, clone's parent) has a nil cloneRegistry. It was likely created incorrectly without using NewInterpreter().",
			root.id,
		))
	}
	root.cloneRegistry = append(root.cloneRegistry, clone)

	// Copy global variables from the root's state.
	root.state.variablesMu.RLock()
	defer root.state.variablesMu.RUnlock()
	for name, value := range root.state.variables {
		if _, isGlobal := root.state.globalVarNames[name]; isGlobal {
			clone.state.variables[name] = value
			clone.state.globalVarNames[name] = true
		}
	}

	clone.evaluate = &evaluation{i: clone}

	if clone.customWhisperFunc == nil {
		clone.customWhisperFunc = clone.defaultWhisperFunc
	}

	return clone
}
