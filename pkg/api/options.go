// NeuroScript Version: 0.8.0
// File version: 7
// Purpose: Defines the public configuration options for the NeuroScript interpreter API.
// filename: pkg/api/options.go
// nlines: 40
// risk_rating: MEDIUM

package api

import (
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
				if hc.EmitFunc != nil {
					hc.EmitFunc(v)
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
