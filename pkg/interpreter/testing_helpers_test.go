// NeuroScript Version: 0.8.0
// File version: 22
// Purpose: Added the 'model:read:*' grant to the default policy to allow tools like 'tool.agentmodel.Get'.
// filename: pkg/interpreter/testing_helpers_test.go
// nlines: 75
// risk_rating: LOW

package interpreter_test

import (
	"bytes"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

// TestHarness provides a consistent, fully-initialized set of components for interpreter testing.
type TestHarness struct {
	T           *testing.T
	Interpreter *interpreter.Interpreter
	Parser      *parser.ParserAPI
	ASTBuilder  *parser.ASTBuilder
	HostContext *interpreter.HostContext
	Logger      interfaces.Logger
}

// NewTestHarness creates a new, fully configured test harness.
func NewTestHarness(t *testing.T) *TestHarness {
	t.Helper()

	logger := logging.NewTestLogger(t)

	// Create a HostContext with safe, non-nil defaults for I/O functions.
	hostCtx := &interpreter.HostContext{
		Logger:      logger,
		Stdout:      &bytes.Buffer{},
		Stdin:       &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
		EmitFunc:    func(v lang.Value) {},    // Default no-op
		WhisperFunc: func(h, d lang.Value) {}, // Default no-op
	}

	// Create a maximally permissive policy for testing. This runs in a trusted
	// context and grants all administrative capabilities to prevent permission
	// errors in interpreter logic tests.
	privilegedPolicy := policy.NewBuilder(policy.ContextConfig).
		Allow("*").
		Grant("model:admin:*").
		Grant("model:read:*"). // Grant permission to read models.
		Grant("account:admin:*").
		Grant("env:read:*").
		Grant("bus:write:*").
		Grant("net:read:*").
		Grant("net:write:*").
		Grant("tool:exec:*").
		Build()

	// The interpreter creates its own parser and builder internally,
	// so we create it directly and then get its components.
	interp := interpreter.NewInterpreter(
		interpreter.WithHostContext(hostCtx),
		interpreter.WithExecPolicy(privilegedPolicy),
	)

	h := &TestHarness{
		T:           t,
		Interpreter: interp,
		Parser:      interp.Parser(),     // Get the parser FROM the interpreter.
		ASTBuilder:  interp.ASTBuilder(), // Get the builder FROM the interpreter.
		HostContext: hostCtx,
		Logger:      logger,
	}

	t.Logf("[DEBUG] NewTestHarness: RETURNING harness with a fully privileged interpreter.")
	return h
}
