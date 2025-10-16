// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Implements a fluent builder for the canonical HostContext. Adds WithActor.
// filename: pkg/interpreter/hostcontext_builder.go
// nlines: 87
// risk_rating: LOW

package interpreter

import (
	"fmt"
	"io"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// HostContextBuilder provides a fluent API for safely constructing a HostContext.
type HostContextBuilder struct {
	hc   *HostContext
	errs []string
}

// NewHostContextBuilder creates a new builder instance.
func NewHostContextBuilder() *HostContextBuilder {
	return &HostContextBuilder{hc: &HostContext{}}
}

// WithLogger sets the mandatory structured logger.
func (b *HostContextBuilder) WithLogger(l interfaces.Logger) *HostContextBuilder {
	b.hc.Logger = l
	return b
}

// WithActor sets the actor identity for the execution context.
func (b *HostContextBuilder) WithActor(actor interfaces.Actor) *HostContextBuilder {
	b.hc.Actor = actor
	return b
}

// WithStdout sets the mandatory standard output writer.
func (b *HostContextBuilder) WithStdout(w io.Writer) *HostContextBuilder {
	b.hc.Stdout = w
	return b
}

// WithStdin sets the mandatory standard input reader.
func (b *HostContextBuilder) WithStdin(r io.Reader) *HostContextBuilder {
	b.hc.Stdin = r
	return b
}

// WithStderr sets the mandatory standard error writer.
func (b *HostContextBuilder) WithStderr(w io.Writer) *HostContextBuilder {
	b.hc.Stderr = w
	return b
}

// WithEmitFunc sets the callback for the 'emit' statement.
func (b *HostContextBuilder) WithEmitFunc(f func(lang.Value)) *HostContextBuilder {
	b.hc.EmitFunc = f
	return b
}

// WithWhisperFunc sets the callback for the 'whisper' statement.
func (b *HostContextBuilder) WithWhisperFunc(f func(handle, data lang.Value)) *HostContextBuilder {
	b.hc.WhisperFunc = f
	return b
}

// WithEventHandlerErrorCallback sets the callback for unhandled errors in event handlers.
func (b *HostContextBuilder) WithEventHandlerErrorCallback(f func(eventName, source string, err *lang.RuntimeError)) *HostContextBuilder {
	b.hc.EventHandlerErrorCallback = f
	return b
}

// Build validates the constructed HostContext and returns it, or an error if mandatory fields are missing.
func (b *HostContextBuilder) Build() (*HostContext, error) {
	if b.hc.Logger == nil {
		b.errs = append(b.errs, "Logger is mandatory")
	}
	if b.hc.Stdout == nil {
		b.errs = append(b.errs, "Stdout is mandatory")
	}
	if b.hc.Stdin == nil {
		b.errs = append(b.errs, "Stdin is mandatory")
	}
	if b.hc.Stderr == nil {
		b.errs = append(b.errs, "Stderr is mandatory")
	}

	if len(b.errs) > 0 {
		return nil, fmt.Errorf("failed to build HostContext: %s", strings.Join(b.errs, ", "))
	}
	return b.hc, nil
}
