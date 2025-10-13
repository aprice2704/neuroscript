// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Refactored MustParse helper to accept a TestHarness, ensuring it uses the harness's correctly configured ASTBuilder and eliminating a rogue builder instance.
// filename: pkg/interpreter/core_test.go
// nlines: 70
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

const testScript = `
:: name: Core Forking Test
func main() means
    set local_var = "I am local"
    emit global_var // Should emit the initial global value
    call Subroutine()
    emit local_var // Should still be "I am local"
endfunc

func Subroutine() means
    set local_var = "I am in a subroutine" // Should not affect main's local_var
    emit local_var
endfunc
`

// MustParse is a test helper that panics if parsing fails.
// It now accepts the harness to ensure it uses a correctly configured builder.
func MustParse(t *testing.T, h *TestHarness, script string) *ast.Program {
	t.Helper()
	tree, err := h.Parser.Parse(script)
	if err != nil {
		t.Fatalf("Failed to parse script: %v", err)
	}
	prog, _, err := h.ASTBuilder.Build(tree)
	if err != nil {
		t.Fatalf("Failed to build AST: %v", err)
	}
	return prog
}

func TestForkingAndContext(t *testing.T) {
	var output strings.Builder
	h := NewTestHarness(t) // Create harness once.

	// 1. Setup HostContext's EmitFunc on the existing harness
	h.HostContext.EmitFunc = func(v lang.Value) {
		fmt.Fprintln(&output, "EMIT:", v.String())
	}

	// 2. Create a new Interpreter instance using the harness's configured components
	interp := interpreter.NewInterpreter(
		interpreter.WithHostContext(h.HostContext),
		interpreter.WithParser(h.Parser),
		interpreter.WithASTBuilder(h.ASTBuilder),
		interpreter.WithGlobals(map[string]any{"global_var": "I am global"}),
	)

	// 3. Load and Run
	program := MustParse(t, h, testScript)
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if _, err := interp.Run("main"); err != nil {
		t.Fatalf("RunProcedure() failed: %v", err)
	}

	// 4. Assert
	expected := "EMIT: I am global\nEMIT: I am in a subroutine\nEMIT: I am local\n"
	if got := output.String(); got != expected {
		t.Errorf("Unexpected output:\ngot:\n%s\nwant:\n%s", got, expected)
	}
}
