// NeuroScript Version: 0.8.0
// File version: 16
// Purpose: Defines the public configuration options. Hardens WithHostContext. Removed WithProviderRegistry (which must be defined in the internal pkg).
// filename: pkg/api/options.go
// nlines: 70
// risk_rating: HIGH

package api

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// WithHostContext provides the interpreter with its essential host-provided dependencies.
// This is the primary and mandatory option for creating a new interpreter.
func WithHostContext(hc *HostContext) Option {
	// This function acts as a bridge, converting the public, opaque api.HostContext
	// into the internal interpreter.HostContext that the core logic requires.
	return func(i *interpreter.Interpreter) {
		// FIX: Enforce mandatory HostContext fields.
		// This aligns with Rule 9 (Fail Fast) and the panic-on-config-error
		// pattern established in config.go, as this Option func cannot
		// return an error.
		if hc.Logger == nil {
			panic("api.WithHostContext: provided HostContext must have a non-nil Logger")
		}
		if hc.Stdout == nil {
			panic("api.WithHostContext: provided HostContext must have a non-nil Stdout")
		}
		if hc.Stdin == nil {
			panic("api.WithHostContext: provided HostContext must have a non-nil Stdin")
		}
		if hc.Stderr == nil {
			panic("api.WithHostContext: provided HostContext must have a non-nil Stderr")
		}

		internalHC := &interpreter.HostContext{
			Logger:       hc.Logger, // Now guaranteed non-nil
			Emitter:      hc.Emitter,
			AITranscript: hc.AITranscript,
			Stdout:       hc.Stdout, // Now guaranteed non-nil
			Stdin:        hc.Stdin,  // Now guaranteed non-nil
			Stderr:       hc.Stderr, // Now guaranteed non-nil
			Actor:        hc.Actor,
			EmitFunc: func(v lang.Value) {
				if hc.EmitFunc != nil {
					hc.EmitFunc(v)
					return
				}
				// Stdout is guaranteed non-nil by the check above.
				unwrapped := lang.Unwrap(v)
				fmt.Fprintln(hc.Stdout, unwrapped)
			},
			WhisperFunc: func(handle, data lang.Value) {
				if hc.WhisperFunc != nil {
					hc.WhisperFunc(handle, data)
				}
			},
			EventHandlerErrorCallback: func(eventName, source string, err *lang.RuntimeError) {
				if hc.EventHandlerErrorCallback != nil {
					hc.EventHandlerErrorCallback(eventName, source, err)
				}
			},
		}
		// Use the internal WithHostContext option to apply the constructed context.
		interpreter.WithHostContext(internalHC)(i)
	}
}

// FIX: Removed WithProviderRegistry. It must be defined in the internal
// 'interpreter' package and re-exported via api/reexport.go,
// just like WithAccountStore and WithAgentModelStore.
