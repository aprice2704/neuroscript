// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Corrects tests to provide a mandatory HostContext during interpreter creation, resolving a panic.
// filename: pkg/api/exec_test.go
// nlines: 65
// risk_rating: MEDIUM

package api_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// TestExecWithInterpreter_NilSafety directly tests Rule 9: Bail Out On Nil.
// It ensures that passing a nil interpreter or a nil tree results in a
// non-panicking, clean error.
func TestExecWithInterpreter_NilSafety(t *testing.T) {
	hc, err := api.NewHostContextBuilder().
		WithLogger(logging.NewNoOpLogger()).
		WithStdout(io.Discard).
		WithStdin(os.Stdin).
		WithStderr(io.Discard).
		Build()
	if err != nil {
		t.Fatalf("Failed to build HostContext for test: %v", err)
	}
	interp := api.New(api.WithHostContext(hc))
	tree := &api.Tree{Root: &ast.Program{}} // A valid, empty tree

	// Test case 1: Nil interpreter
	_, err = api.ExecWithInterpreter(context.Background(), nil, tree)
	if err == nil {
		t.Error("Expected an error when passing a nil interpreter, but got nil")
	}

	// Test case 2: Nil tree
	_, err = api.ExecWithInterpreter(context.Background(), interp, nil)
	if err == nil {
		t.Error("Expected an error when passing a nil tree, but got nil")
	}
}

// TestExecWithInterpreter_InvalidRootNode ensures the type assertion guard
// for the program's root node is working as expected.
func TestExecWithInterpreter_InvalidRootNode(t *testing.T) {
	// A tree with a root that is not a runnable *ast.Program
	type NotAProgram struct {
		interfaces.Node
	}
	tree := &api.Tree{Root: &NotAProgram{}}
	hc, err := api.NewHostContextBuilder().
		WithLogger(logging.NewNoOpLogger()).
		WithStdout(io.Discard).
		WithStdin(os.Stdin).
		WithStderr(io.Discard).
		Build()
	if err != nil {
		t.Fatalf("Failed to build HostContext for test: %v", err)
	}
	interp := api.New(api.WithHostContext(hc))

	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)
	if err == nil {
		t.Fatal("Expected an error for an invalid root node type, but got nil")
	}
	if !strings.Contains(err.Error(), "is not a runnable *ast.Program") {
		t.Errorf("Expected error message to contain specific text, but got: %v", err)
	}
}

// TestExecInNewInterpreter_ParseError ensures that syntax errors from the
// source code are correctly propagated up from the one-shot executor.
func TestExecInNewInterpreter_ParseError(t *testing.T) {
	invalidSrc := `func main() { this is not valid neuroscript }`
	_, err := api.ExecInNewInterpreter(context.Background(), invalidSrc)

	if err == nil {
		t.Fatal("Expected a parsing error, but got nil")
	}
	// This checks if the error is wrapped correctly, adhering to Rule 7.
	if !strings.Contains(err.Error(), "parsing failed") {
		t.Errorf("Expected error to indicate a parse failure, but got: %v", err)
	}
}
