// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Defines the public configuration options for the NeuroScript interpreter API. FIX: The default EmitFunc now correctly prints to Stdout if no custom function is provided.
// filename: pkg/api/options.go
// nlines: 45
// risk_rating: MEDIUM

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
		internalHC := &interpreter.HostContext{
			Logger:       hc.Logger,
			Emitter:      hc.Emitter,
			AITranscript: hc.AITranscript,
			Stdout:       hc.Stdout,
			Stdin:        hc.Stdin,
			Stderr:       hc.Stderr,
			EmitFunc: func(v lang.Value) {
				// If the user provided a custom emit function, use it.
				if hc.EmitFunc != nil {
					hc.EmitFunc(v)
					return
				}
				// Otherwise, default to printing the unwrapped value to Stdout.
				if hc.Stdout != nil {
					unwrapped := lang.Unwrap(v)
					// Use Fprintln to ensure a newline, matching 'emit' behavior.
					fmt.Fprintln(hc.Stdout, unwrapped)
				}
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
