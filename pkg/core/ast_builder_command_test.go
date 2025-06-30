// filename: pkg/core/ast_builder_command_test.go
// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Corrected expected parser error messages to match actual output.

package core_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
)

// testParseAndBuild is a helper that runs the full parsing and AST build pipeline.
func testParseAndBuild(t *testing.T, script string) *core.Program {
	t.Helper()
	logger := adapters.NewNoOpLogger()
	parserAPI := core.NewParserAPI(logger)
	tree, err := parserAPI.Parse(script)
	if err != nil {
		t.Fatalf("Parser failed: %v", err)
	}
	builder := core.NewASTBuilder(logger)
	prog, _, err := builder.Build(tree)
	if err != nil {
		t.Fatalf("AST builder failed: %v", err)
	}
	return prog
}

// testForParserError is a helper that asserts a script fails at the parsing stage.
func testForParserError(t *testing.T, script string, expectedError string) {
	t.Helper()
	logger := adapters.NewNoOpLogger()
	parserAPI := core.NewParserAPI(logger)
	_, err := parserAPI.Parse(script)
	if err == nil {
		t.Fatalf("Expected a parser error containing '%s', but parsing succeeded.", expectedError)
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', but got: %v", expectedError, err)
	}
}

func TestCommandBlockParsing(t *testing.T) {
	t.Run("Empty command block fails", func(t *testing.T) {
		script := `
			command
			endcommand
		`
		testForParserError(t, script, "extraneous input 'endcommand' expecting")
	})

	t.Run("Mixing command and func fails", func(t *testing.T) {
		script := `
			command
				set x = 1
			endcommand

			func myFunc() means
				set y = 2
			endfunc
		`
		testForParserError(t, script, "mismatched input 'func' expecting <EOF>")
	})

	t.Run("Return statement is a syntax error", func(t *testing.T) {
		script := `
			command
				return 42
			endcommand
		`
		testForParserError(t, script, "mismatched input 'return' expecting")
	})

	t.Run("On event statement is a syntax error", func(t *testing.T) {
		script := `
			command
				on event "foo" do
					set x = 1
				endon
			endcommand
		`
		testForParserError(t, script, "mismatched input 'event' expecting 'error'")
	})

	t.Run("Func definition is a syntax error", func(t *testing.T) {
		script := `
			command
				func inner() means
					set x = 1
				endfunc
			endcommand
		`
		testForParserError(t, script, "mismatched input 'func' expecting")
	})

	t.Run("On error handler is allowed", func(t *testing.T) {
		script := `
			command
				set x = 1
				on error do
					emit "error"
				endon
				set y = 2
			endcommand
		`
		prog := testParseAndBuild(t, script)
		if len(prog.Commands) != 1 {
			t.Fatalf("Expected 1 command block, got %d", len(prog.Commands))
		}
		cmd := prog.Commands[0]
		if len(cmd.ErrorHandlers) != 1 {
			t.Fatalf("Expected 1 error handler, got %d", len(cmd.ErrorHandlers))
		}
		if cmd.ErrorHandlers[0].Type != "on_error" {
			t.Errorf("Expected handler type 'on_error', got '%s'", cmd.ErrorHandlers[0].Type)
		}
		if len(cmd.Body) != 2 {
			t.Errorf("Expected 2 regular statements in the body, got %d", len(cmd.Body))
		}
	})

	t.Run("Valid statements are parsed correctly", func(t *testing.T) {
		script := `
			command
				set x = (1 + 2)
				if x > 2
					emit "x is greater than 2"
				else
					fail "logic error"
				endif
				call someFunction(x)
			endcommand
		`
		prog := testParseAndBuild(t, script)
		if len(prog.Commands) != 1 {
			t.Fatalf("Expected 1 command block, got %d", len(prog.Commands))
		}
		cmd := prog.Commands[0]
		if len(cmd.Body) != 3 {
			t.Fatalf("Expected 3 statements in command body, got %d", len(cmd.Body))
		}
	})

	t.Run("Sequential command blocks are parsed", func(t *testing.T) {
		script := `
			command
				:: name: first
				set x = 1
			endcommand
			
			command
				:: name: second
				set y = 2
			endcommand
		`
		prog := testParseAndBuild(t, script)
		if len(prog.Commands) != 2 {
			t.Fatalf("Expected 2 command blocks, got %d", len(prog.Commands))
		}
		cmd1 := prog.Commands[0]
		cmd2 := prog.Commands[1]

		if name := cmd1.Metadata["name"]; name != "first" {
			t.Errorf("Expected first command metadata name to be 'first', got '%s'", name)
		}
		if name := cmd2.Metadata["name"]; name != "second" {
			t.Errorf("Expected second command metadata name to be 'second', got '%s'", name)
		}
	})
}
