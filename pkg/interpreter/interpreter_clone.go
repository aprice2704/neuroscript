// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Centralizes the interpreter cloning logic for sandboxed execution.
// filename: pkg/interpreter/interpreter_clone.go
// nlines: 60
// risk_rating: HIGH

package interpreter

// clone creates a new interpreter instance for sandboxing.
// It shares persistent state (stores, tools, config) by reference,
// but creates an isolated variable scope.
func (i *Interpreter) clone() *Interpreter {
	clone := &Interpreter{
		// Share core state by reference for persistence
		logger:              i.logger,
		fileAPI:             i.fileAPI,
		tools:               i.tools,
		eventManager:        i.eventManager,
		aiWorker:            i.aiWorker,
		stdout:              i.stdout,
		stdin:               i.stdin,
		stderr:              i.stderr,
		maxLoopIterations:   i.maxLoopIterations,
		bufferManager:       i.bufferManager,
		objectCache:         i.objectCache,
		llmclient:           i.llmclient,
		skipStdTools:        i.skipStdTools,
		modelStore:          i.modelStore,
		accountStore:        i.accountStore,
		ExecPolicy:          i.ExecPolicy,
		root:                i, // Link back to the parent
		aiTranscript:        i.aiTranscript,
		transientPrivateKey: i.transientPrivateKey,
		customEmitFunc:      i.customEmitFunc,
		customWhisperFunc:   i.customWhisperFunc,
		turnCtx:             i.turnCtx,
	}

	// Create isolated execution state
	clone.state = newInterpreterState()
	clone.state.providers = i.state.providers             // Share provider registry
	clone.state.knownProcedures = i.state.knownProcedures // Share procedure definitions

	// Find ultimate root to copy globals from.
	root := i
	for root.root != nil {
		root = root.root
	}

	// Copy global variables from the root interpreter
	root.state.variablesMu.RLock()
	defer root.state.variablesMu.RUnlock()
	for name, value := range root.state.variables {
		if _, isGlobal := root.state.globalVarNames[name]; isGlobal {
			clone.state.variables[name] = value
			clone.state.globalVarNames[name] = true
		}
	}

	clone.evaluate = &evaluation{i: clone}

	// Ensure the clone has a whisper function. If the parent's is custom,
	// it was already copied. If not, set the default.
	if clone.customWhisperFunc == nil {
		clone.customWhisperFunc = clone.defaultWhisperFunc
	}

	return clone
}
