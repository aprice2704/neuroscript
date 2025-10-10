// NeuroScript Version: 0.8.0
// File version: 31
// Purpose: FIX: Clones the EventManager correctly instead of accessing the removed state.eventHandlers field.
// filename: pkg/interpreter/interpreter_clone.go
// nlines: 79
// risk_rating: HIGH

package interpreter

import (
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/google/uuid"
)

// clone creates a new interpreter instance for sandboxing.
// It shares the RunnerParcel and SharedCatalogs by reference,
// but creates an isolated execution state (variables, stack, etc.).
func (i *Interpreter) clone() *Interpreter {
	fmt.Fprintf(os.Stderr, "\n[CLONE START] Parent ID: %s\n", i.id)

	clone := &Interpreter{
		id: fmt.Sprintf("interp-%s", uuid.NewString()[:8]),

		// --- AX Contracts ---
		// Pass parcel and catalogs by reference (pointer copy).
		parcel:   i.parcel,
		catalogs: i.catalogs,

		// --- Share core state by reference ---
		fileAPI:                   i.fileAPI,
		aiWorker:                  i.aiWorker,
		stdout:                    i.stdout,
		stdin:                     i.stdin,
		stderr:                    i.stderr,
		bufferManager:             i.bufferManager,
		objectCache:               i.objectCache,
		llmclient:                 i.llmclient,
		root:                      i.rootInterpreter(),
		aiTranscript:              i.aiTranscript,
		transientPrivateKey:       i.transientPrivateKey,
		emitter:                   i.emitter,
		customEmitFunc:            i.customEmitFunc,
		customWhisperFunc:         i.customWhisperFunc,
		eventHandlerErrorCallback: i.eventHandlerErrorCallback,
		eventManager:              newEventManager(), // Initialize a new manager for the clone
	}

	// The runtime context for a clone must be the clone itself.
	clone.runtime = clone

	// Create a fresh execution state.
	clone.state = newInterpreterState()
	clone.state.sandboxDir = i.state.sandboxDir
	clone.state.providers = i.state.providers
	clone.state.knownProcedures = i.state.knownProcedures

	// Event handlers are stateful to the interpreter instance, not shared globally.
	// We must deep-copy the handlers from the parent.
	if i.eventManager != nil {
		for eventName, handlers := range i.eventManager.eventHandlers {
			handlersCopy := make([]*ast.OnEventDecl, len(handlers))
			copy(handlersCopy, handlers)
			clone.eventManager.eventHandlers[eventName] = handlersCopy
		}
	}

	root := clone.rootInterpreter()
	if root == nil {
		panic(fmt.Sprintf("FATAL: Interpreter (ID: %s) has a nil root.", clone.id))
	}

	root.cloneRegistryMu.Lock()
	if root.cloneRegistry == nil {
		panic(fmt.Sprintf("FATAL: Root interpreter (ID: %s) has a nil cloneRegistry.", root.id))
	}
	root.cloneRegistry = append(root.cloneRegistry, clone)
	root.cloneRegistryMu.Unlock()

	clone.evaluate = &evaluation{i: clone}

	if clone.customWhisperFunc == nil {
		clone.customWhisperFunc = clone.defaultWhisperFunc
	}

	fmt.Fprintf(os.Stderr, "[CLONE END] Parent ID: %s -> New Clone ID: %s registered with root.\n\n", i.id, clone.id)

	return clone
}
