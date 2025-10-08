// NeuroScript Version: 0.7.4
// File version: 1
// Purpose: Defines the core Interpreter facade struct and its constructor.
// filename: pkg/api/interpreter.go
// nlines: 36
// risk_rating: LOW

package api

import (
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider/google"
)

// Interpreter is a facade over the internal interpreter, providing a stable,
// high-level API for embedding NeuroScript.
type Interpreter struct {
	internal *interpreter.Interpreter
	runtime  Runtime // Stored on the facade for the ax Identity() method.
}

// New creates a new, persistent NeuroScript interpreter instance.
func New(opts ...Option) *Interpreter {
	i := interpreter.NewInterpreter(opts...)
	if i.ExecPolicy == nil {
		i.ExecPolicy = &policy.ExecPolicy{
			Context: policy.ContextNormal,
			Allow:   []string{},
		}
	}

	googleProvider := google.New()
	i.RegisterProvider("google", googleProvider)

	return &Interpreter{internal: i}
}
