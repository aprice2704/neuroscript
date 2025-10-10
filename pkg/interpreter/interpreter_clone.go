// NeuroScript Version: 0.8.0
// File version: 29
// Purpose: Refactored to pass the RunnerParcel by reference and removed the obsolete globals-to-locals copying logic.
// filename: pkg/interpreter/interpreter_clone.go
// nlines: 100
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
	fmt.Fprintf(os.Stderr, "[CLONE RUNTIME DEBUG] Parent ID: %s, Parent Addr: %p, Parent Runtime Addr: %p\n", i.id, i, i.runtime)

	clone := &Interpreter{
		// Assign a new unique ID to the clone
		id: fmt.Sprintf("interp-%s", uuid.NewString()[:8]),

		// --- Runner Parcel ---
		// The parcel is passed by reference (pointer copy).
		parcel: i.parcel,

		// Share core state by reference for persistence
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
		root:                      i.rootInterpreter(),
		aiTranscript:              i.aiTranscript,
		transientPrivateKey:       i.transientPrivateKey,
		eventHandlerErrorCallback: i.eventHandlerErrorCallback,
		emitter:                   i.emitter,
		// The runtime is set below after the clone is fully constructed.

		// Propagate the turn context from the parent. This is critical for AEIOU.
		turnCtx: i.turnCtx,

		// Propagate the custom I/O functions from the parent.
		customEmitFunc:    i.customEmitFunc,
		customWhisperFunc: i.customWhisperFunc,
	}

	// The runtime context for a clone must be the clone itself.
	clone.runtime = clone

	fmt.Fprintf(os.Stderr, "[CLONE RUNTIME DEBUG] Clone ID: %s,  Clone Addr: %p,  Clone Runtime Addr: %p\n", clone.id, clone, clone.runtime)
	fmt.Fprintf(os.Stderr, "[CLONE MID]   New Clone ID: %s | customEmitFunc is nil after copy: %t\n", clone.id, clone.customEmitFunc == nil)

	if i.tools != nil {
		clone.tools = i.tools.NewViewForInterpreter(clone)
	}

	// Create a fresh execution state.
	clone.state = newInterpreterState()
	clone.state.sandboxDir = i.state.sandboxDir
	clone.state.providers = i.state.providers
	clone.state.knownProcedures = i.state.knownProcedures

	// The old logic for copying globals into the local variable map has been removed.
	// Globals are now accessed read-only via `i.parcel.Globals()`.

	root := clone.rootInterpreter()
	if root == nil {
		panic(fmt.Sprintf("FATAL: Interpreter (ID: %s) has a nil root.", clone.id))
	}

	root.cloneRegistryMu.Lock()
	defer root.cloneRegistryMu.Unlock()
	if root.cloneRegistry == nil {
		panic(fmt.Sprintf("FATAL: Root interpreter (ID: %s) has a nil cloneRegistry.", root.id))
	}
	root.cloneRegistry = append(root.cloneRegistry, clone)

	clone.evaluate = &evaluation{i: clone}

	if clone.customWhisperFunc == nil {
		clone.customWhisperFunc = clone.defaultWhisperFunc
	}

	fmt.Fprintf(os.Stderr, "[CLONE END]   Parent ID: %s -> New Clone ID: %s successfully registered with root.\n\n", i.id, clone.id)

	return clone
}
