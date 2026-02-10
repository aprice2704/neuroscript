// NeuroScript Version: 0.8.0
// File version: 24
// Purpose: Replaced non-thread-safe bytes.Buffer with a thread-safe implementation to fix the Dirty Buffer anti-pattern.
// filename: pkg/interpreter/testing_helpers_test.go
// nlines: 110
// risk_rating: LOW

package interpreter_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// ThreadSafeBuffer is a simple wrapper around bytes.Buffer to satisfy Law 15.12.
type ThreadSafeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *ThreadSafeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *ThreadSafeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func (b *ThreadSafeBuffer) Read(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Read(p)
}

// TestHarness provides a consistent, fully-initialized set of components for interpreter testing.
type TestHarness struct {
	T                *testing.T
	Interpreter      *interpreter.Interpreter
	Parser           *parser.ParserAPI
	ASTBuilder       *parser.ASTBuilder
	HostContext      *interpreter.HostContext
	Logger           interfaces.Logger
	ProviderRegistry *provider.Registry
}

// NewTestHarness creates a new, fully configured test harness.
func NewTestHarness(t *testing.T) *TestHarness {
	t.Helper()

	logger := logging.NewTestLogger(t)

	// Create a HostContext with thread-safe buffers instead of bytes.Buffer
	// to prevent "Dirty Buffer" data races during async tests.
	hostCtx := &interpreter.HostContext{
		Logger:      logger,
		Stdout:      &ThreadSafeBuffer{},
		Stdin:       &ThreadSafeBuffer{},
		Stderr:      &ThreadSafeBuffer{},
		EmitFunc:    func(v lang.Value) {},    // Default no-op
		WhisperFunc: func(h, d lang.Value) {}, // Default no-op
	}

	privilegedPolicy := policy.NewBuilder(policy.ContextConfig).
		Allow("*").
		Grant("model:admin:*").
		Grant("model:read:*").
		Grant("account:admin:*").
		Grant("env:read:*").
		Grant("bus:write:*").
		Grant("net:read:*").
		Grant("net:write:*").
		Grant("tool:exec:*").
		Build()

	providerRegistry := provider.NewRegistry()

	interp := interpreter.NewInterpreter(
		interpreter.WithHostContext(hostCtx),
		interpreter.WithExecPolicy(privilegedPolicy),
		interpreter.WithProviderRegistry(providerRegistry),
	)

	h := &TestHarness{
		T:                t,
		Interpreter:      interp,
		Parser:           interp.Parser(),
		ASTBuilder:       interp.ASTBuilder(),
		HostContext:      hostCtx,
		Logger:           logger,
		ProviderRegistry: providerRegistry,
	}

	return h
}
