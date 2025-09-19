// NeuroScript Version: 0.7.2
// File version: 22
// Purpose: [DEBUG] Adds extensive logging to the clone method to trace the propagation of custom I/O functions and the admin registry.
// filename: pkg/interpreter/interpreter_clone.go
// nlines: 105
// risk_rating: HIGH

package interpreter

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

// clone creates a new interpreter instance for sandboxing.
// It shares persistent state (stores, tools, config) by reference,
// but creates an isolated variable scope and a context-aware tool registry view.
func (i *Interpreter) clone() *Interpreter {
	// --- VERBOSE DEBUGGING ---
	fmt.Fprintf(os.Stderr, "\n[CLONE START] Parent ID: %s | customEmitFunc is nil: %t\n", i.id, i.customEmitFunc == nil)

	clone := &Interpreter{
		// Assign a new unique ID to the clone
		id: fmt.Sprintf("interp-%s", uuid.NewString()[:8]),

		// Share core state by reference for persistence
		logger:                    i.logger,
		fileAPI:                   i.fileAPI,
		eventManager:              i.eventManager,
		aiWorker:                  i.aiWorker,
		stdout:                    i.stdout,
		stdin:                     i.stdin,
		stderr:                    i.stderr,
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
		eventHandlerErrorCallback: i.eventHandlerErrorCallback,
		emitter:                   i.emitter,

		// --- BUG FIX & VERIFICATION ---
		// Propagate the custom I/O functions from the parent.
		customEmitFunc:    i.customEmitFunc,
		customWhisperFunc: i.customWhisperFunc,
	}

	fmt.Fprintf(os.Stderr, "[CLONE MID]   New Clone ID: %s | customEmitFunc is nil after copy: %t\n", clone.id, clone.customEmitFunc == nil)

	if i.tools != nil {
		clone.tools = i.tools.NewViewForInterpreter(clone)
	}

	clone.state = newInterpreterState()
	clone.state.sandboxDir = i.state.sandboxDir

	clone.state.providers = i.state.providers
	clone.state.knownProcedures = i.state.knownProcedures

	root := clone.rootInterpreter()

	if root == nil {
		panic(fmt.Sprintf(
			"FATAL: Interpreter (ID: %s) has a nil root. This should not be possible.",
			clone.id,
		))
	}

	root.cloneRegistryMu.Lock()
	defer root.cloneRegistryMu.Unlock()

	if root.cloneRegistry == nil {
		panic(fmt.Sprintf(
			"FATAL: The root interpreter (ID: %s, clone's parent) has a nil cloneRegistry. It was likely created incorrectly without using NewInterpreter().",
			root.id,
		))
	}
	root.cloneRegistry = append(root.cloneRegistry, clone)

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

	fmt.Fprintf(os.Stderr, "[CLONE END]   Parent ID: %s -> New Clone ID: %s successfully registered with root.\n\n", i.id, clone.id)

	return clone
}
