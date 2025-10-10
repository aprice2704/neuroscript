// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: FIX: Removed the public ExecPolicy() method to enforce policy management via the ax factory.
// filename: pkg/api/interpreter.go
// nlines: 32
// risk_rating: LOW

package api

import (
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider/google"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// Interpreter is a facade over the internal interpreter.
type Interpreter struct {
	internal *interpreter.Interpreter
	runtime  Runtime
}

// New creates a new, persistent NeuroScript interpreter instance.
func New(opts ...Option) *Interpreter {
	i := interpreter.NewInterpreter(opts...)
	if i.ExecPolicy == nil {
		// Default to a deny-by-default policy if none is provided.
		i.ExecPolicy = policy.NewBuilder(policy.ContextNormal).Build()
	}

	googleProvider := google.New()
	i.RegisterProvider("google", googleProvider)

	return &Interpreter{internal: i}
}

func (i *Interpreter) InternalRuntime() tool.Runtime {
	return i.runtime
}
