// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Corrected MustParse helper to return the correct ast.Program type.
// filename: pkg/interpreter/core_test.go
// nlines: 65
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
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
func MustParse(t *testing.T, script string) *ast.Program {
	t.Helper()
	pAPI := parser.NewParserAPI(logging.NewTestLogger(t))
	tree, err := pAPI.Parse(script)
	if err != nil {
		t.Fatalf("Failed to parse script: %v", err)
	}
	prog, _, err := parser.NewASTBuilder(logging.NewTestLogger(t)).Build(tree)
	if err != nil {
		t.Fatalf("Failed to build AST: %v", err)
	}
	return prog
}

func TestForkingAndContext(t *testing.T) {
	var output strings.Builder

	// 1. Setup HostContext
	hostCtx := &interpreter.HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: &output,
		EmitFunc: func(v lang.Value) {
			fmt.Fprintln(&output, "EMIT:", v.String())
		},
	}

	// 2. Create Root Interpreter with a global
	interp := interpreter.NewInterpreter(
		interpreter.WithHostContext(hostCtx),
		interpreter.WithGlobals(map[string]any{"global_var": "I am global"}),
	)

	// 3. Load and Run
	program := MustParse(t, testScript)
	if err := interp.Load(&ast.Tree{Root: program}); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if _, err := interp.RunProcedure("main"); err != nil {
		t.Fatalf("RunProcedure() failed: %v", err)
	}

	// 4. Assert
	expected := "EMIT: I am global\nEMIT: I am in a subroutine\nEMIT: I am local\n"
	if got := output.String(); got != expected {
		t.Errorf("Unexpected output:\ngot:\n%s\nwant:\n%s", got, expected)
	}
}
