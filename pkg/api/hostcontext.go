// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Defines the HostContext struct for the public API, with unexported fields to enforce builder usage.
// filename: pkg/api/hostcontext.go
// nlines: 27
// risk_rating: LOW

package api

import (
	"io"
)

// HostContext holds all host-provided, immutable dependencies for an interpreter.
// This struct must be constructed using the HostContextBuilder to ensure all
// mandatory fields are correctly initialized.
type HostContext struct {
	logger                    Logger
	emitter                   Emitter
	aiTranscript              io.Writer
	stdout                    io.Writer
	stdin                     io.Reader
	stderr                    io.Writer
	emitFunc                  func(Value)
	whisperFunc               func(handle, data Value)
	eventHandlerErrorCallback func(eventName, source string, err *RuntimeError)
}
