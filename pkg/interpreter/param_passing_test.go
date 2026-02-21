// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 7
// :: description: Refactored to use the centralized TestHarness, test optional parameters, and comprehensively test typeof operator.
// :: latestChange: Added TestInterpreter_TypeOf and corresponding test script.
// :: filename: pkg/interpreter/interpreter_param_passing_test.go
// :: serialization: go

package interpreter_test

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

const paramPassingTestScriptEnhanced = `
:: Name: Parameter Passing Test Script (Enhanced)
:: Version: 1.3.0

func mainEntry(needs strArg, intArg, boolArg, floatArg returns result) means
  emit "mainEntry_recvd:" + strArg + "," + intArg + "," + boolArg + "," + floatArg
  call helperProc(strArg + "_to_helper", intArg * 2, not boolArg, floatArg / 2.0)
  call recursiveProc(strArg, intArg, boolArg, 3)
  call optionalProc("reqVal", "optVal")
  call optionalProc("reqValOnly")
  set result = "mainEntry_completed_with_" + strArg
  return result
endfunc

func helperProc(needs pStr, pInt, pBool, pFloat returns result) means
  emit "helperProc_recvd:" + pStr + "," + pInt + "," + pBool + "," + pFloat
  set result = "helperProc_completed"
  return result
endfunc

func recursiveProc(needs rStrArg, rIntArg, rBoolArg, rDepth returns result) means
  emit "recursiveProc_recvd:" + rStrArg + "," + rIntArg + "," + rBoolArg + "," + rDepth
  if rDepth > 1
    set newDepth = rDepth - 1
    call recursiveProc(rStrArg, rIntArg + rDepth, not rBoolArg, newDepth)
  endif
  set result = "recursiveProc_completed_depth_" + rDepth
  return result
endfunc

func optionalProc(needs reqParam optional optParam) means
  emit "optionalProc_recvd:" + reqParam + "," + typeof(optParam)
endfunc
`

const typeOfTestScript = `
:: Name: TypeOf Operator Test Script
:: Version: 1.0.0

func testTypeOf() means
  set s = "hello"
  set n = 42.5
  set b = true
  set l = [1, 2, 3]
  set m = {"key": "value"}
  set x = nil
  
  emit "typeof_s:" + typeof(s)
  emit "typeof_n:" + typeof(n)
  emit "typeof_b:" + typeof(b)
  emit "typeof_l:" + typeof(l)
  emit "typeof_m:" + typeof(m)
  emit "typeof_x:" + typeof(x)
  emit "typeof_literal_num:" + typeof(100)
  emit "typeof_literal_str:" + typeof("direct")
endfunc
`

func compareOutputLineWithSpecialFloatHandling(t *testing.T, iteration, lineIndex int, actualLine, expectedLine string) {
	t.Helper()

	const helperProcPrefix = "helperProc_recvd:"
	const mainEntryRecvdPrefix = "mainEntry_recvd:"
	const epsilon = 0.00001

	isHelperLine := strings.HasPrefix(expectedLine, helperProcPrefix) && strings.HasPrefix(actualLine, helperProcPrefix)
	isMainEntryLine := strings.HasPrefix(expectedLine, mainEntryRecvdPrefix) && strings.HasPrefix(actualLine, mainEntryRecvdPrefix)

	if isHelperLine || isMainEntryLine {
		partsExpected := strings.Split(expectedLine, ",")
		partsActual := strings.Split(actualLine, ",")

		if len(partsExpected) == 4 && len(partsActual) == 4 {
			for k := 0; k < 3; k++ {
				if strings.TrimSpace(partsActual[k]) != strings.TrimSpace(partsExpected[k]) {
					t.Errorf("Iteration %d: Output line %d, part %d string mismatch.\nGot:      '%s'\nExpected: '%s'",
						iteration, lineIndex, k+1, actualLine, expectedLine)
					return
				}
			}

			expectedFloatStr := strings.TrimSpace(partsExpected[3])
			actualFloatStr := strings.TrimSpace(partsActual[3])

			expectedFloat, errExp := strconv.ParseFloat(expectedFloatStr, 64)
			actualFloat, errAct := strconv.ParseFloat(actualFloatStr, 64)

			if errExp != nil || errAct != nil {
				t.Errorf("Iteration %d: Output line %d: Error parsing float values", iteration, lineIndex)
				return
			}

			if math.Abs(actualFloat-expectedFloat) > epsilon {
				t.Errorf("Iteration %d: Output line %d float part numeric mismatch.\nGot:      '%s'\nExpected: '%s'",
					iteration, lineIndex, actualLine, expectedLine)
			}
			return
		}
	}

	if strings.TrimSpace(actualLine) != strings.TrimSpace(expectedLine) {
		t.Errorf("Iteration %d: Output line %d string mismatch.\nGot:      '%s'\nExpected: '%s'",
			iteration, lineIndex, actualLine, expectedLine)
	}
}

func TestInterpreter_ParameterPassingFuzz(t *testing.T) {
	const numTestIterations = 10
	t.Logf("[DEBUG] Turn 1: Starting parameter passing fuzz test for %d iterations.", numTestIterations)

	h := NewTestHarness(t)
	tree, parseErr := h.Parser.Parse(paramPassingTestScriptEnhanced)
	if parseErr != nil {
		t.Fatalf("Failed to parse script: %v", parseErr)
	}
	program, _, buildErr := h.ASTBuilder.Build(tree)
	if buildErr != nil {
		t.Fatalf("Failed to build AST: %v", buildErr)
	}
	t.Logf("[DEBUG] Turn 2: Shared test script parsed and AST built successfully.")

	for i := 0; i < numTestIterations; i++ {
		iteration := i
		t.Run(fmt.Sprintf("Iteration%d", iteration), func(t *testing.T) {
			t.Parallel()

			var capturedOutput bytes.Buffer
			iterHarness := NewTestHarness(t) // Create a fresh harness for each parallel run
			iterHarness.HostContext.EmitFunc = func(v lang.Value) {
				fmt.Fprintln(&capturedOutput, v.String())
			}
			interp := iterHarness.Interpreter
			interp.Load(&interfaces.Tree{Root: program}) // Load the shared program

			strVal := fmt.Sprintf("cliStr%d", iteration)
			intVal := 100 + iteration
			boolVal := iteration%2 == 0
			floatVal := 10.5 + float64(iteration)*0.1

			simulatedCLIArgs := []interface{}{
				strVal,
				int64(intVal),
				boolVal,
				floatVal,
			}

			wrappedArgs := make([]lang.Value, len(simulatedCLIArgs))
			for i, arg := range simulatedCLIArgs {
				wrapped, err := lang.Wrap(arg)
				if err != nil {
					t.Fatalf("Iteration %d: Failed to wrap argument #%d: %v", iteration, i, err)
				}
				wrappedArgs[i] = wrapped
			}

			t.Logf("[DEBUG] Turn 3 (Iter %d): Starting execution of 'mainEntry'.", iteration)
			_, runErr := interp.Run("mainEntry", wrappedArgs...)
			if runErr != nil {
				t.Errorf("Iteration %d: Error executing procedure: %v. Output so far: %s", iteration, runErr, capturedOutput.String())
				return
			}
			t.Logf("[DEBUG] Turn 4 (Iter %d): Execution finished.", iteration)

			outputLines := strings.Split(strings.TrimSpace(capturedOutput.String()), "\n")

			expectedEmits := []string{
				fmt.Sprintf("mainEntry_recvd:%s,%d,%t,%.1f", strVal, intVal, boolVal, floatVal),
				fmt.Sprintf("helperProc_recvd:%s_to_helper,%d,%t,%.2f", strVal, intVal*2, !boolVal, floatVal/2.0),
				fmt.Sprintf("recursiveProc_recvd:%s,%d,%t,3", strVal, intVal, boolVal),
				fmt.Sprintf("recursiveProc_recvd:%s,%d,%t,2", strVal, intVal+3, !boolVal),
				fmt.Sprintf("recursiveProc_recvd:%s,%d,%t,1", strVal, (intVal+3)+2, boolVal),
				"optionalProc_recvd:reqVal,string",
				"optionalProc_recvd:reqValOnly,nil",
			}

			if len(outputLines) != len(expectedEmits) {
				t.Errorf("Iteration %d: Unexpected number of output lines. Got %d, want %d", iteration, len(outputLines), len(expectedEmits))
				return
			}

			for j, expected := range expectedEmits {
				if j >= len(outputLines) {
					continue
				}
				compareOutputLineWithSpecialFloatHandling(t, iteration, j, outputLines[j], expected)
			}
			t.Logf("[DEBUG] Turn 5 (Iter %d): Assertions complete.", iteration)
		})
	}
}

func TestInterpreter_TypeOf(t *testing.T) {
	h := NewTestHarness(t)
	tree, parseErr := h.Parser.Parse(typeOfTestScript)
	if parseErr != nil {
		t.Fatalf("Failed to parse script: %v", parseErr)
	}
	program, _, buildErr := h.ASTBuilder.Build(tree)
	if buildErr != nil {
		t.Fatalf("Failed to build AST: %v", buildErr)
	}

	var capturedOutput bytes.Buffer
	h.HostContext.EmitFunc = func(v lang.Value) {
		fmt.Fprintln(&capturedOutput, v.String())
	}

	interp := h.Interpreter
	interp.Load(&interfaces.Tree{Root: program})

	_, runErr := interp.Run("testTypeOf")
	if runErr != nil {
		t.Fatalf("Error executing procedure: %v. Output: %s", runErr, capturedOutput.String())
	}

	output := capturedOutput.String()
	expectedEmits := []string{
		"typeof_s:string",
		"typeof_n:number",
		"typeof_b:boolean",
		"typeof_l:list",
		"typeof_m:map",
		"typeof_x:nil",
		"typeof_literal_num:number",
		"typeof_literal_str:string",
	}

	for _, expected := range expectedEmits {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got:\n%s", expected, output)
		}
	}
}
