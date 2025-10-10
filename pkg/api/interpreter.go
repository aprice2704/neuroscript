// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: FIX: Removed incorrect type assertion loop. The functional options pattern correctly handles overriding the default policy.
// filename: pkg/api/interpreter.go
// nlines: 35
// risk_rating: MEDIUM

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
	// Prepend our secure, "deny-all" default policy. If the caller provides
	// their own WithExecPolicy option, it will run after this one and
	// correctly override the default.
	defaultPolicy := policy.NewBuilder(policy.ContextNormal).Build()
	finalOpts := append([]Option{WithExecPolicy(defaultPolicy)}, opts...)

	i := interpreter.NewInterpreter(finalOpts...)

	googleProvider := google.New()
	i.RegisterProvider("google", googleProvider)

	return &Interpreter{internal: i}
}

func (i *Interpreter) InternalRuntime() tool.Runtime {
	return i.internal
}
