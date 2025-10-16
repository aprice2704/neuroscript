// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Defines the canonical HostContext struct. Adds the Actor field to carry identity.
// filename: pkg/interpreter/hostcontext.go
// nlines: 28
// risk_rating: LOW

package interpreter

import (
	"io"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// HostContext holds all host-provided, immutable dependencies for an interpreter.
// This object is created once by the host application using the HostContextBuilder
// and shared by reference among all interpreter instances, ensuring consistent
// access to host capabilities.
type HostContext struct {
	Logger                    interfaces.Logger
	FileAPI                   interfaces.FileAPI
	Emitter                   interfaces.Emitter
	AITranscript              io.Writer
	Stdout                    io.Writer
	Stdin                     io.Reader
	Stderr                    io.Writer
	Actor                     interfaces.Actor // <-- Identity for the execution context
	EmitFunc                  func(lang.Value)
	WhisperFunc               func(handle, data lang.Value)
	EventHandlerErrorCallback func(eventName, source string, err *lang.RuntimeError)
}
