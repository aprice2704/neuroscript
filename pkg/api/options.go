// NeuroScript Version: 0.8.0
// File version: 18
// Purpose: Defines public options. Copies ServiceRegistry from public HostContext to internal HostContext.
// filename: pkg/api/options.go
// nlines: 75
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
	// (which is an alias for interpreter.HostContext)
	// into the internal interpreter.HostContext that the core logic requires.
	return func(i *interpreter.Interpreter) {
		// FIX: Enforce mandatory HostContext fields.
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
			Logger:       hc.Logger,
			Emitter:      hc.Emitter,
			AITranscript: hc.AITranscript,
			Stdout:       hc.Stdout,
			Stdin:        hc.Stdin,
			Stderr:       hc.Stderr,
			Actor:        hc.Actor,
			// ADDED: Copy the ServiceRegistry, as required by the AEIOU v2 spec
			// This now works because the public 'hc' type (aliased from
			// interpreter.HostContext) has this field.
			ServiceRegistry: hc.ServiceRegistry,
			EmitFunc: func(v lang.Value) {
				if hc.EmitFunc != nil {
					hc.EmitFunc(v)
					return
				}
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
		// This assumes 'interpreter.WithHostContext' is an unexported option.
		interpreter.WithHostContext(internalHC)(i)
	}
}
