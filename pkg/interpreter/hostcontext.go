// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Defines the HostContext struct to hold all host-provided, immutable dependencies.
// filename: pkg/interpreter/hostcontext.go
// nlines: 25
// risk_rating: LOW

package interpreter

import (
	"io"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// HostContext holds all host-provided, immutable dependencies for an interpreter.
// This object is created once by the host application and shared by reference
// among all interpreter instances, ensuring consistent access to host capabilities.
type HostContext struct {
	Logger                    interfaces.Logger
	FileAPI                   interfaces.FileAPI
	Emitter                   interfaces.Emitter
	AITranscript              io.Writer
	Stdout                    io.Writer
	Stdin                     io.Reader
	Stderr                    io.Writer
	EmitFunc                  func(lang.Value)
	WhisperFunc               func(handle, data lang.Value)
	EventHandlerErrorCallback func(eventName, source string, err *lang.RuntimeError)
}
