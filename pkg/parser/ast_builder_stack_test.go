// filename: pkg/parser/ast_builder_stack_test.go
// NeuroScript Version: 0.5.2
// File version: 5
// Purpose: Corrected the popN function to return elements in the correct LIFO order.

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// --- Stack Implementation (copied for testing) ---

type valueStack struct {
	data []interface{}
}

func newValueStack() *valueStack {
	return &valueStack{data: make([]interface{}, 0)}
}

func (s *valueStack) push(v interface{}) {
	s.data = append(s.data, v)
}

func (s *valueStack) pop() (interface{}, bool) {
	if s.len() == 0 {
		return nil, false
	}
	index := len(s.data) - 1
	val := s.data[index]
	s.data = s.data[:index]
	return val, true
}

func (s *valueStack) popN(n int) ([]interface{}, bool) {
	if s.len() < n {
		return nil, false
	}
	index := len(s.data) - n
	vals := s.data[index:]
	s.data = s.data[:index]

	// Reverse the slice to ensure LIFO order for the popped items.
	for i, j := 0, len(vals)-1; i < j; i, j = i+1, j-1 {
		vals[i], vals[j] = vals[j], vals[i]
	}

	return vals, true
}

func (s *valueStack) peek() (interface{}, bool) {
	if s.len() == 0 {
		return nil, false
	}
	return s.data[len(s.data)-1], true
}

func (s *valueStack) len() int {
	return len(s.data)
}

func popAs[T any](s *valueStack) (T, bool) {
	var zero T
	raw, ok := s.pop()
	if !ok {
		return zero, false
	}
	typed, ok := raw.(T)
	if !ok {
		s.push(raw) // Push it back on type mismatch
		return zero, false
	}
	return typed, true
}

// --- Tests ---

type astTestCase struct {
	name          string
	scriptContent string
	expectProc    bool
}

func TestASTBuilderScenarios(t *testing.T) {
	logger := logging.NewNoOpLogger()

	testCases := []astTestCase{
		{
			name:       "MinimalStackTestFromPrevious",
			expectProc: true,
			scriptContent: `
func MinimalStackTest(returns result) means
  :: description: A test function to check for stack corruption in AST builder.
  set counter = 0
  for each i in [1, 2]
    if i == 1
      set counter = counter + 1
    endif
    emit "in_loop"
  endfor
  if counter > 0
    set counter = counter + 10
    emit "after_loop_if"
  endif
  set final_val = counter
  return final_val
  on error do
    emit "error_occurred_in_MinimalStackTest"
    return "ERROR_STATE"
  endon
endfunc
`,
		},
		{
			name:       "DeeplyNestedBlocks",
			expectProc: true,
			scriptContent: `
func DeepNesting() means
  set status = "init"
  if true
    emit "outer_if_true"
    for each x in [1]
      emit "for_x_start"
      if x == 1
        emit "for_x_if_true"
        set status = "in_level_3_if"
        while status == "in_level_3_if" # Loop once then change status
          emit "in_while"
          set status = "exiting_while"
        endwhile
        emit "after_while"
      else
        emit "for_x_if_else" # Should not happen
      endif
      emit "for_x_end"
    endfor
    emit "after_for_x"
  else
    emit "outer_if_else" # Should not happen
  endif
  return status
endfunc
`,
		},
		{
			name:       "EmptyAndMinimalBlocks",
			expectProc: true,
			scriptContent: `
func EmptyAndMinimalBlocksTest() means
  if true
    set _ = nil
  endif

  if false
    set _ = nil
  else
    set _ = nil
  endif

  for each item in []
    set _ = nil
  endfor

  set y = 0
  while y > 10 # Condition initially false
    set _ = nil
  endwhile

  on error do
    set _ = nil
  endon
  return "done_empty_blocks"
endfunc
`,
		},
		{
			name:       "SequentialBlocksOfDifferentTypes",
			expectProc: true,
			scriptContent: `
func SequentialBlocksTest() means
  emit "start"
  if true
    emit "first_if"
  endif

  set x = 0
  for each i in [1,2]
    set x = x + i
  endfor

  if x == 3
    emit "second_if_after_for"
  else
    emit "error_in_logic_sequential"
  endif

  set z = 5
  while z > 4 # Condition for loop
    emit "in_while_block"
    set z = 0 # exit loop
  endwhile

  return "sequential_blocks_processed"
endfunc
`,
		},
		{
			name:       "LoopControlsWithNesting",
			expectProc: true,
			scriptContent: `
func LoopControlTest() means
  set counter = 0
  set outer_tracker = ""
  for each x in [1,2,3,4]
    set outer_tracker = outer_tracker + "o" + string(x)
    if x == 1
      emit "outer_continue_for_x_1"
      continue # Skip to next x
    endif

    set inner_tracker = ""
    for each y in ["a", "b", "c"]
      set inner_tracker = inner_tracker + "i" + y
      if y == "b"
        emit "inner_break_for_y_b"
        break # Break inner loop
      endif
      set counter = counter + 1
    endfor

    if x == 3
      emit "outer_break_for_x_3"
      break # Break outer loop
    endif
    emit "end_outer_iteration_for_x_" + string(x)
  endfor
  return counter
endfunc
`,
		},
		{
			name:       "ScriptWithOnlyMetadataAndComments",
			expectProc: false,
			scriptContent: `
:: Name: Only Metadata
:: Version: 1.0
# This is a comment.
# Another comment.

:: Key: value
`,
		},
		{
			name:       "OnErrorAtVeryBeginning",
			expectProc: true,
			scriptContent: `
func OnErrorFirstTest() means
  on error do
    emit "handled_early"
  endon
  set a = 1 / 0 # This should trigger on_error
  return a # Should not be reached
endfunc
`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			// Use the consolidated helper from ast_builder_test_helpers.go
			parserAPI := NewParserAPI(logger)
			scriptNameForTest := tc.name + ".ns"

			antlrTree, antlrErr := parserAPI.Parse(tc.scriptContent)
			if antlrErr != nil {
				t.Fatalf("parserAPI.Parse() returned an ANTLR error for '%s': %v", scriptNameForTest, antlrErr)
			}
			if antlrTree == nil {
				t.Fatalf("parserAPI.Parse() returned a nil ANTLR tree without error for '%s'", scriptNameForTest)
			}

			astBuilder := NewASTBuilder(logger)
			programAst, fileMetadata, buildErr := astBuilder.Build(antlrTree)

			if buildErr != nil {
				t.Errorf("astBuilder.Build() returned an error for script '%s':\n%v", scriptNameForTest, buildErr)
			}
			if programAst == nil && buildErr == nil {
				t.Errorf("astBuilder.Build() returned a nil Program AST without errors for script '%s'", scriptNameForTest)
			}

			if buildErr == nil && programAst != nil {
				if tc.expectProc {
					if len(programAst.Procedures) != 1 {
						t.Errorf("Expected 1 procedure in Program AST for script '%s', got %d. FileMetadata: %v. Procedures found: %v",
							scriptNameForTest, len(programAst.Procedures), fileMetadata, getProcNames(programAst.Procedures))
					}
				} else {
					if len(programAst.Procedures) != 0 {
						t.Errorf("Expected 0 procedures for script '%s', but got %d: %v",
							scriptNameForTest, len(programAst.Procedures), getProcNames(programAst.Procedures))
					}
				}
			}
		})
	}
}

func getProcNames(procs map[string]*ast.Procedure) []string {
	names := make([]string, 0, len(procs))
	for name := range procs {
		names = append(names, name)
	}
	return names
}

func TestValueStack(t *testing.T) {
	t.Run("Push and Pop", func(t *testing.T) {
		s := newValueStack()
		s.push(1)
		s.push("hello")

		if s.len() != 2 {
			t.Errorf("Expected stack length of 2, got %d", s.len())
		}

		val, ok := s.pop()
		if !ok || val != "hello" {
			t.Errorf("Expected to pop 'hello', got %v", val)
		}

		val, ok = s.pop()
		if !ok || val != 1 {
			t.Errorf("Expected to pop 1, got %v", val)
		}

		if s.len() != 0 {
			t.Errorf("Expected stack to be empty, got length %d", s.len())
		}
	})

	t.Run("Pop from empty stack", func(t *testing.T) {
		s := newValueStack()
		_, ok := s.pop()
		if ok {
			t.Error("Expected pop from empty stack to fail, but it succeeded")
		}
	})

	t.Run("Peek", func(t *testing.T) {
		s := newValueStack()
		s.push(123)
		val, ok := s.peek()
		if !ok || val != 123 {
			t.Errorf("Expected peek to return 123, got %v", val)
		}
		if s.len() != 1 {
			t.Errorf("Expected stack length to be 1 after peek, got %d", s.len())
		}
	})

	t.Run("Peek from empty stack", func(t *testing.T) {
		s := newValueStack()
		_, ok := s.peek()
		if ok {
			t.Error("Expected peek from empty stack to fail, but it succeeded")
		}
	})

	t.Run("PopN", func(t *testing.T) {
		s := newValueStack()
		s.push(1)
		s.push("two")
		s.push(true)

		vals, ok := s.popN(3)
		if !ok {
			t.Fatal("popN(3) failed")
		}
		if len(vals) != 3 {
			t.Fatalf("Expected 3 values from popN, got %d", len(vals))
		}
		// Expected LIFO order
		if vals[0] != true || vals[1] != "two" || vals[2] != 1 {
			t.Errorf("popN returned incorrect values or order: %v", vals)
		}
	})

	t.Run("PopN more than available", func(t *testing.T) {
		s := newValueStack()
		s.push(1)
		_, ok := s.popN(2)
		if ok {
			t.Error("Expected popN(2) on a stack of size 1 to fail, but it succeeded")
		}
	})

	t.Run("PopAs successful", func(t *testing.T) {
		s := newValueStack()
		expectedNode := &ast.StringLiteralNode{Value: "test"}
		s.push(expectedNode)

		node, ok := popAs[*ast.StringLiteralNode](s)
		if !ok {
			t.Fatal("popAs failed unexpectedly")
		}
		if node != expectedNode {
			t.Errorf("popAs returned wrong node. Expected %v, got %v", expectedNode, node)
		}
	})

	t.Run("PopAs type mismatch", func(t *testing.T) {
		s := newValueStack()
		s.push(123) // Push an int

		_, ok := popAs[*ast.StringLiteralNode](s) // Try to pop as a node
		if ok {
			t.Error("Expected popAs to fail due to type mismatch, but it succeeded")
		}
	})
}
