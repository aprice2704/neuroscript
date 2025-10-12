// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Implements a fluent builder for the HostContext to improve API ergonomics and safety.
// filename: pkg/api/hostcontext_builder.go
// nlines: 85
// risk_rating: LOW

package api

import (
	"fmt"
	"io"
)

// HostContextBuilder provides a fluent interface for constructing a HostContext.
type HostContextBuilder struct {
	hc  *HostContext
	err error
}

// NewHostContextBuilder creates a new builder for a HostContext.
func NewHostContextBuilder() *HostContextBuilder {
	return &HostContextBuilder{
		hc: &HostContext{},
	}
}

// WithLogger sets the mandatory structured logger.
func (b *HostContextBuilder) WithLogger(l Logger) *HostContextBuilder {
	b.hc.logger = l
	return b
}

// WithStdout sets the mandatory standard output writer.
func (b *HostContextBuilder) WithStdout(w io.Writer) *HostContextBuilder {
	b.hc.stdout = w
	return b
}

// WithStdin sets the mandatory standard input reader.
func (b *HostContextBuilder) WithStdin(r io.Reader) *HostContextBuilder {
	b.hc.stdin = r
	return b
}

// WithStderr sets the mandatory standard error writer.
func (b *HostContextBuilder) WithStderr(w io.Writer) *HostContextBuilder {
	b.hc.stderr = w
	return b
}

// WithEmitFunc sets the callback for the 'emit' statement.
func (b *HostContextBuilder) WithEmitFunc(f func(Value)) *HostContextBuilder {
	b.hc.emitFunc = f
	return b
}

// WithEmitter sets the LLM telemetry emitter.
func (b *HostContextBuilder) WithEmitter(e Emitter) *HostContextBuilder {
	b.hc.emitter = e
	return b
}

// WithAITranscript sets the writer for AI transcripts.
func (b *HostContextBuilder) WithAITranscript(w io.Writer) *HostContextBuilder {
	b.hc.aiTranscript = w
	return b
}

// WithWhisperFunc sets the callback for the 'whisper' statement.
func (b *HostContextBuilder) WithWhisperFunc(f func(handle, data Value)) *HostContextBuilder {
	b.hc.whisperFunc = f
	return b
}

// WithEventHandlerErrorCallback sets the callback for unhandled errors in event handlers.
func (b *HostContextBuilder) WithEventHandlerErrorCallback(f func(eventName, source string, err *RuntimeError)) *HostContextBuilder {
	b.hc.eventHandlerErrorCallback = f
	return b
}

// Build finalizes the HostContext, validating that all mandatory fields are set.
func (b *HostContextBuilder) Build() (*HostContext, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.hc.logger == nil {
		return nil, fmt.Errorf("validation failed: Logger is a mandatory field")
	}
	if b.hc.stdout == nil {
		return nil, fmt.Errorf("validation failed: Stdout is a mandatory field")
	}
	if b.hc.stdin == nil {
		return nil, fmt.Errorf("validation failed: Stdin is a mandatory field")
	}
	if b.hc.stderr == nil {
		return nil, fmt.Errorf("validation failed: Stderr is a mandatory field")
	}
	return b.hc, nil
}
