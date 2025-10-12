// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Corrects the TestHarness to properly initialize the ASTBuilder with the interpreter's event handler registration callback, fixing the long-standing parser panic.
// filename: pkg/interpreter/testing_helpers_test.go
// nlines: 55
// risk_rating: LOW

package interpreter_test

import (
	"bytes"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
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
// It initializes an interpreter with a default HostContext (including logger and I/O),
// a parser, and an AST builder. CRUCIALLY, it now wires the interpreter's
// event registration method to the ASTBuilder's callback to prevent panics.
func NewTestHarness(t *testing.T) *TestHarness {
	t.Helper()

	logger := logging.NewTestLogger(t)

	hostCtx := &interpreter.HostContext{
		Logger: logger,
		Stdout: &bytes.Buffer{},
		Stdin:  &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}

	interp := interpreter.NewInterpreter(interpreter.WithHostContext(hostCtx))
	p := parser.NewParserAPI(logger)
	b := parser.NewASTBuilder(logger)

	// THIS IS THE CRITICAL FIX: Wire the interpreter's event handler
	// registration function to the AST builder's callback.
	// This requires both methods to be exported.
	b.SetEventHandlerCallback(func(decl *ast.OnEventDecl) {
		interp.RegisterEventHandler(decl)
	})

	return &TestHarness{
		T:           t,
		Interpreter: interp,
		Parser:      p,
		ASTBuilder:  b,
		HostContext: hostCtx,
		Logger:      logger,
	}
}
