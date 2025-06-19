// NeuroScript Version: 0.3.0
// File version: 0.1.5
// Purpose: Updated test scripts to use the new 'on error do' syntax.
// filename: pkg/core/ast_builder_stack_test.go
// nlines: 179
// risk_rating: LOW

package core_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
)

type astTestCase struct {
	name          string
	scriptContent string
	expectProc    bool
}

func TestASTBuilderScenarios(t *testing.T) {
	logger := adapters.NewNoOpLogger()

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
    # Empty then-branch
  endif

  if false
    # This won't run
  else
    # Empty else-branch
  endif

  for each item in []
    # Empty for-each loop
  endfor

  set y = 0
  while y > 10 # Condition initially false
    # Empty while loop
  endwhile

  on error do
    # Empty on_error handler
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
    set outer_tracker = outer_tracker + "o" + x
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
      set counter = counter + 1 # Increments for (x=2,y=a), (x=3,y=a)
    endfor
    
    if x == 3
      emit "outer_break_for_x_3"
      break # Break outer loop
    endif
    emit "end_outer_iteration_for_x_" + x
  endfor
  return counter # x=1 (skipped), x=2 (y=a adds 1), x=3 (y=a adds 1, then outer break). Expected: 1+1 = 2
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

:: Key: Value
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
			parserAPI := core.NewParserAPI(logger)
			scriptNameForTest := tc.name + ".ns"

			antlrTree, antlrErr := parserAPI.Parse(tc.scriptContent)
			if antlrErr != nil {
				t.Fatalf("parserAPI.Parse() returned an ANTLR error for '%s': %v", scriptNameForTest, antlrErr)
			}
			if antlrTree == nil {
				t.Fatalf("parserAPI.Parse() returned a nil ANTLR tree without error for '%s'", scriptNameForTest)
			}

			astBuilder := core.NewASTBuilder(logger)
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

func getProcNames(procs map[string]*core.Procedure) []string {
	names := make([]string, 0, len(procs))
	for name := range procs {
		names = append(names, name)
	}
	return names
}
