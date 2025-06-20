// NeuroScript Version: 0.3.0
// File version: 1.0.2
// Purpose: Reverted test script to use the '+' operator for concatenation, aligning with the non-strict (lenient) operator model.
// filename: pkg/core/interpreter_param_passing_test.go
// nlines: 275+
// risk_rating: MEDIUM
package core

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
)

// The script is reverted to its original form, using '+' for concatenation.
// This is now valid because the operator is non-strict.
const paramPassingTestScriptEnhanced = `
:: Name: Parameter Passing Test Script (Enhanced)
:: Version: 1.2.0

func mainEntry(needs strArg, intArg, boolArg, floatArg returns result) means
  :: description: Simulates entry point receiving CLI-like arguments with varied types.
  emit "mainEntry_recvd:" + strArg + "," + intArg + "," + boolArg + "," + floatArg
  
  call helperProc(strArg + "_to_helper", intArg * 2, not boolArg, floatArg / 2.0)
  
  call recursiveProc(strArg, intArg, boolArg, 3)
  
  set result = "mainEntry_completed_with_" + strArg
  return result
endfunc

func helperProc(needs pStr, pInt, pBool, pFloat returns result) means
  :: description: Helper procedure to check argument passing with varied types.
  emit "helperProc_recvd:" + pStr + "," + pInt + "," + pBool + "," + pFloat
  set result = "helperProc_completed"
  return result
endfunc

func recursiveProc(needs rStrArg, rIntArg, rBoolArg, rDepth returns result) means
  :: description: Recursive procedure to check argument passing and scope with varied types.
  emit "recursiveProc_recvd:" + rStrArg + "," + rIntArg + "," + rBoolArg + "," + rDepth
  
  if rDepth > 1
    set newDepth = rDepth - 1
    call recursiveProc(rStrArg, rIntArg + rDepth, not rBoolArg, newDepth)
  endif
  
  set result = "recursiveProc_completed_depth_" + rDepth
  return result
endfunc
`

// compareOutputLineWithSpecialFloatHandling compares actual and expected lines.
// It performs numeric comparison for floats in lines starting with specific prefixes.
func compareOutputLineWithSpecialFloatHandling(t *testing.T, iteration, lineIndex int, actualLine, expectedLine string) {
	t.Helper()

	const helperProcPrefix = "helperProc_recvd:"
	const mainEntryRecvdPrefix = "mainEntry_recvd:"
	const epsilon = 0.00001 // Tolerance for float comparison

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

			if errExp != nil {
				t.Errorf("Iteration %d: Output line %d: Error parsing EXPECTED float value '%s': %v",
					iteration, lineIndex, expectedFloatStr, errExp)
				return
			}
			if errAct != nil {
				t.Errorf("Iteration %d: Output line %d: Error parsing ACTUAL float value '%s': %v",
					iteration, lineIndex, actualFloatStr, errAct)
				return
			}

			if math.Abs(actualFloat-expectedFloat) > epsilon {
				t.Errorf("Iteration %d: Output line %d float part numeric mismatch.\nGot:      '%s' (parsed: %f)\nExpected: '%s' (parsed: %f)",
					iteration, lineIndex, actualLine, actualFloat, expectedLine, expectedFloat)
			}
			return
		}
	}

	normalizedGot := strings.Join(strings.Fields(actualLine), "")
	normalizedExpected := strings.Join(strings.Fields(expectedLine), "")
	if normalizedGot != normalizedExpected {
		t.Errorf("Iteration %d: Output line %d string mismatch.\nGot:      '%s' (Normalized: '%s')\nExpected: '%s' (Normalized: '%s')",
			iteration, lineIndex, actualLine, normalizedGot, expectedLine, normalizedExpected)
	}
}

func TestInterpreter_ParameterPassingFuzz(t *testing.T) {
	const numTestIterations = 10

	baseLogger := NewTestLogger(t)

	parser := NewParserAPI(baseLogger)
	parseTree, parseErr := parser.Parse(paramPassingTestScriptEnhanced)
	if parseErr != nil {
		t.Fatalf("Failed to parse script: %v", parseErr)
	}
	astBuilder := NewASTBuilder(baseLogger)
	program, _, buildErr := astBuilder.Build(parseTree)
	if buildErr != nil {
		t.Fatalf("Failed to build AST: %v", buildErr)
	}
	if program == nil || len(program.Procedures) == 0 {
		t.Fatalf("AST build resulted in no procedures")
	}

	for i := 0; i < numTestIterations; i++ {
		iteration := i
		t.Run(fmt.Sprintf("Iteration%d", iteration), func(t *testing.T) {
			t.Parallel()

			var capturedOutput bytes.Buffer
			iterLogger := NewTestLogger(t)

			interp, err := NewInterpreter(iterLogger, nil, ".", nil, nil)
			if err != nil {
				t.Fatalf("Iteration %d: Failed to create interpreter: %v", iteration, err)
			}
			interp.SetStdout(&capturedOutput)

			for name, procDef := range program.Procedures {
				procCopy := *procDef
				if errP := interp.AddProcedure(procCopy); errP != nil {
					t.Fatalf("Iteration %d: Failed to add procedure %s: %v", iteration, name, errP)
				}
			}

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

			wrappedArgs := make([]Value, len(simulatedCLIArgs))
			for i, arg := range simulatedCLIArgs {
				wrapped, err := Wrap(arg)
				if err != nil {
					t.Fatalf("Iteration %d: Failed to wrap argument #%d (%v): %v", iteration, i, arg, err)
				}
				wrappedArgs[i] = wrapped
			}

			targetProcName := "mainEntry"

			_, runErr := interp.RunProcedure(targetProcName, wrappedArgs...)
			if runErr != nil {
				t.Errorf("Iteration %d: Error executing procedure '%s': %v. Output so far: %s", iteration, targetProcName, runErr, capturedOutput.String())
				return
			}

			outputStr := capturedOutput.String()
			outputLines := strings.Split(strings.TrimSpace(outputStr), "\n")

			expectedEmits := []string{
				fmt.Sprintf("mainEntry_recvd:%s,%d,%t,%.1f", strVal, intVal, boolVal, floatVal),
				fmt.Sprintf("helperProc_recvd:%s_to_helper,%d,%t,%.2f", strVal, intVal*2, !boolVal, floatVal/2.0),
				fmt.Sprintf("recursiveProc_recvd:%s,%d,%t,3", strVal, intVal, boolVal),
				fmt.Sprintf("recursiveProc_recvd:%s,%d,%t,2", strVal, intVal+3, !boolVal),
				fmt.Sprintf("recursiveProc_recvd:%s,%d,%t,1", strVal, (intVal+3)+2, boolVal),
			}

			if len(outputLines) != len(expectedEmits) {
				t.Errorf("Iteration %d: Unexpected number of output lines. Got %d, want %d", iteration, len(outputLines), len(expectedEmits))
				t.Log("Got lines:")
				for i, line := range outputLines {
					t.Logf("  [%d] %s", i, line)
				}
				t.Log("Expected lines:")
				for i, line := range expectedEmits {
					t.Logf("  [%d] %s", i, line)
				}
				return
			}

			for j, expected := range expectedEmits {
				if j >= len(outputLines) {
					t.Errorf("Iteration %d: Missing expected output line %d: %s", iteration, j, expected)
					continue
				}
				actual := strings.TrimSpace(outputLines[j])
				compareOutputLineWithSpecialFloatHandling(t, iteration, j, actual, expected)
			}
		})
	}
}
