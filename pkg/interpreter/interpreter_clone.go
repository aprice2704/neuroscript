// NeuroScript Version: 0.7.2
// File version: 15
// Purpose: Corrects a bug where the LLM telemetry emitter was not being propagated to cloned interpreters.
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
		emitter:                   i.emitter, // FIX: Propagate the emitter to the clone.
	}

	if i.tools != nil {
		clone.tools = i.tools.NewViewForInterpreter(clone)
	}

	clone.state = newInterpreterState()
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
