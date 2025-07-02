// filename: pkg/parser/ast_builder_if_else_test.go
// NeuroScript Version: 0.5.2
// File version: 0.1.6
// Purpose: Updated empty block tests to be syntactically valid by adding a neutral statement.

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

// Helper function to extract an if ast.Step from parsed procedure body.
// Assumes the IF statement is the first statement in the procedure.
func getIfStepFromTestProc(t *testing.T, scriptContent string) *ast.Step {
	t.Helper()
	bodyNodes := parseStringToProcedureBodyNodes(t, scriptContent, "TestProc") // Returns []ast.Step
	if len(bodyNodes) == 0 {
		t.Fatalf("TestProc body is empty, expected an IF step")
	}
	ifStep := &bodyNodes[0]
	if ifStep.Type != "if" {
		t.Fatalf("Expected first step to be of type 'if', got type %s", ifStep.Type)
	}
	return ifStep
}

// expectEmitStepWithString checks for a specific Emit step with string content.
func expectEmitStepWithString(t *testing.T, stmts []ast.Step, index int, expectedContent string, blockName string) {
	t.Helper()
	if index >= len(stmts) {
		t.Errorf("[%s] Expected at least %d steps, got %d", blockName, index+1, len(stmts))
		return
	}
	emitStep := stmts[index]
	if emitStep.Type != "emit" {
		t.Errorf("[%s] Expected step at index %d to be of type 'emit', got '%s'", blockName, index, emitStep.Type)
		return
	}

	if emitStep.Values == nil || len(emitStep.Values) == 0 {
		t.Errorf("[%s] Emit step at index %d has nil or empty Values slice", blockName, index)
		return
	}
	if emitStep.Values[0] == nil {
		t.Errorf("[%s] Emit step at index %d has a nil expression in its Values slice", blockName, index)
		return
	}

	strLit, ok := emitStep.Values[0].(*ast.StringLiteralNode)
	if !ok {
		t.Errorf("[%s] Expected Emit step's expression to be *ast.StringLiteralNode, got %T", blockName, emitStep.Values[0])
		return
	}
	if strLit.Value != expectedContent {
		t.Errorf("[%s] Emit step content mismatch. Expected: '%s', Got: '%s'", blockName, expectedContent, strLit.Value)
	}
}

func TestIfThenElse_SimpleIfThen(t *testing.T) {
	script := `
func TestProc() means
	if true
		emit "Inside THEN"
	endif
endfunc
`
	ifStep := getIfStepFromTestProc(t, script)

	if len(ifStep.Body) != 1 {
		t.Errorf("SimpleIfThen: Expected 1 step in Body, got %d", len(ifStep.Body))
	} else {
		expectEmitStepWithString(t, ifStep.Body, 0, "Inside THEN", "Body")
	}

	if ifStep.ElseBody != nil && len(ifStep.ElseBody) != 0 {
		t.Errorf("SimpleIfThen: Expected 0 Else steps or nil, got %d", len(ifStep.ElseBody))
	}
}

func TestIfThenElse_SimpleIfThenElse(t *testing.T) {
	script := `
func TestProc() means
	if true
		emit "In THEN block"
	else
		emit "In ELSE block"
	endif
endfunc
`
	ifStep := getIfStepFromTestProc(t, script)

	if len(ifStep.Body) != 1 {
		t.Errorf("SimpleIfThenElse (THEN/Body): Expected 1 statement, got %d.", len(ifStep.Body))
	} else {
		expectEmitStepWithString(t, ifStep.Body, 0, "In THEN block", "Body")
	}
	for _, stmt := range ifStep.Body {
		if stmt.Type == "emit" {
			if len(stmt.Values) > 0 {
				if strLit, okStr := stmt.Values[0].(*ast.StringLiteralNode); okStr && strLit.Value == "In ELSE block" {
					t.Errorf("SimpleIfThenElse (BUG CHECK): ELSE statement (emit \"In ELSE block\") found in Body")
				}
			}
		}
	}

	if ifStep.ElseBody == nil || len(ifStep.ElseBody) != 1 {
		t.Errorf("SimpleIfThenElse (ELSE/Else): Expected 1 statement, got %d (or nil).", len(ifStep.ElseBody))
	} else {
		expectEmitStepWithString(t, ifStep.ElseBody, 0, "In ELSE block", "Else")
	}
}

func TestIfThenElse_IfElseIfElse(t *testing.T) {
	script := `
func TestProc() means
	if false
		emit "In initial THEN"
	else
		if true
			emit "In ELSEIF block"
		else
			emit "In final ELSE block"
		endif
	endif
endfunc
`
	ifStep := getIfStepFromTestProc(t, script)

	if len(ifStep.Body) != 1 {
		t.Errorf("IfElseIfElse (THEN/Body): Expected 1 statement, got %d", len(ifStep.Body))
	} else {
		expectEmitStepWithString(t, ifStep.Body, 0, "In initial THEN", "Body")
	}
	for _, stmt := range ifStep.Body {
		if stmt.Type == "emit" {
			if len(stmt.Values) > 0 {
				if strLit, okStr := stmt.Values[0].(*ast.StringLiteralNode); okStr {
					if strLit.Value == "In ELSEIF block" {
						t.Errorf("IfElseIfElse (BUG CHECK): ELSEIF statement found in Body")
					}
					if strLit.Value == "In final ELSE block" {
						t.Errorf("IfElseIfElse (BUG CHECK): Final ELSE statement found in Body")
					}
				}
			}
		}
	}

	if ifStep.ElseBody == nil || len(ifStep.ElseBody) != 1 {
		t.Fatalf("IfElseIfElse: Expected 1 step in outer Else (for elseif/else structure), got %v", ifStep.ElseBody)
	}

	elseIfStepAsIf := ifStep.ElseBody[0]
	if elseIfStepAsIf.Type != "if" {
		t.Fatalf("IfElseIfElse: Expected elseif block to be represented as an 'if' step, got type %s", elseIfStepAsIf.Type)
	}
	if len(elseIfStepAsIf.Body) != 1 {
		t.Errorf("IfElseIfElse (ELSEIF/Body): Expected 1 statement, got %d.", len(elseIfStepAsIf.Body))
	} else {
		expectEmitStepWithString(t, elseIfStepAsIf.Body, 0, "In ELSEIF block", "ElseIf.Body")
	}

	if elseIfStepAsIf.ElseBody == nil || len(elseIfStepAsIf.ElseBody) != 1 {
		t.Errorf("IfElseIfElse (final ELSE/Else): Expected 1 statement, got %d (or nil).", len(elseIfStepAsIf.ElseBody))
	} else {
		expectEmitStepWithString(t, elseIfStepAsIf.ElseBody, 0, "In final ELSE block", "FinalElse.Else")
	}
}

func TestIfThenElse_IfElseIfOnly(t *testing.T) {
	script := `
func TestProc() means
	if false
		emit "First THEN"
	else
		if true
			emit "The ELSEIF"
		endif
	endif
endfunc
`
	ifStep := getIfStepFromTestProc(t, script)

	if len(ifStep.Body) != 1 {
		t.Errorf("IfElseIfOnly (THEN/Body): Expected 1 statement, got %d", len(ifStep.Body))
	} else {
		expectEmitStepWithString(t, ifStep.Body, 0, "First THEN", "Body")
	}
	for _, stmt := range ifStep.Body {
		if stmt.Type == "emit" {
			if len(stmt.Values) > 0 {
				if strLit, okStr := stmt.Values[0].(*ast.StringLiteralNode); okStr && strLit.Value == "The ELSEIF" {
					t.Errorf("IfElseIfOnly (BUG CHECK): ELSEIF statement found in Body")
				}
			}
		}
	}

	if ifStep.ElseBody == nil || len(ifStep.ElseBody) != 1 {
		t.Fatalf("IfElseIfOnly: Expected 1 step in Else for the elseif, got %v", ifStep.ElseBody)
	}
	elseIfStepAsIf := ifStep.ElseBody[0]
	if elseIfStepAsIf.Type != "if" {
		t.Fatalf("IfElseIfOnly: Expected elseif to be an 'if' step, got %s", elseIfStepAsIf.Type)
	}
	if len(elseIfStepAsIf.Body) != 1 {
		t.Errorf("IfElseIfOnly (ELSEIF/Body): Expected 1 statement, got %d", len(elseIfStepAsIf.Body))
	} else {
		expectEmitStepWithString(t, elseIfStepAsIf.Body, 0, "The ELSEIF", "ElseIf.Body")
	}

	if elseIfStepAsIf.ElseBody != nil && len(elseIfStepAsIf.ElseBody) != 0 {
		t.Errorf("IfElseIfOnly: Expected 0 Else steps or nil for the elseif, got %d", len(elseIfStepAsIf.ElseBody))
	}
}

func TestIfThenElse_MultipleElseIfs(t *testing.T) {
	script := `
func TestProc() means
	if false
		emit "Then"
	else
		if false
			emit "ElseIf 1"
		else
			if true
				emit "ElseIf 2 (executes)"
			else
				if false
					emit "ElseIf 3"
				else
					emit "Else"
				endif
			endif
		endif
	endif
endfunc
`
	ifStep := getIfStepFromTestProc(t, script)

	if len(ifStep.Body) != 1 {
		t.Errorf("MultipleElseIfs (THEN/Body): Expected 1, got %d", len(ifStep.Body))
	} else {
		expectEmitStepWithString(t, ifStep.Body, 0, "Then", "Body")
	}

	ifOuterElse := ifStep.ElseBody
	if ifOuterElse == nil || len(ifOuterElse) != 1 {
		t.Fatalf("MultipleElseIfs: Expected outer else for first elseif, got %v", ifOuterElse)
	}
	firstElseIfStep := ifOuterElse[0]
	if firstElseIfStep.Type != "if" {
		t.Fatalf("MultipleElseIfs: Expected first elseif to be if step, got type %s", firstElseIfStep.Type)
	}
	expectEmitStepWithString(t, firstElseIfStep.Body, 0, "ElseIf 1", "ElseIfs[0].Body")

	ifFirstElseIfElse := firstElseIfStep.ElseBody
	if ifFirstElseIfElse == nil || len(ifFirstElseIfElse) != 1 {
		t.Fatalf("MultipleElseIfs: Expected else for second elseif, got %v", ifFirstElseIfElse)
	}
	secondElseIfStep := ifFirstElseIfElse[0]
	if secondElseIfStep.Type != "if" {
		t.Fatalf("MultipleElseIfs: Expected second elseif to be if step, got type %s", secondElseIfStep.Type)
	}
	expectEmitStepWithString(t, secondElseIfStep.Body, 0, "ElseIf 2 (executes)", "ElseIfs[1].Body")

	ifSecondElseIfElse := secondElseIfStep.ElseBody
	if ifSecondElseIfElse == nil || len(ifSecondElseIfElse) != 1 {
		t.Fatalf("MultipleElseIfs: Expected else for third elseif, got %v", ifSecondElseIfElse)
	}
	thirdElseIfStep := ifSecondElseIfElse[0]
	if thirdElseIfStep.Type != "if" {
		t.Fatalf("MultipleElseIfs: Expected third elseif to be if step, got type %s", thirdElseIfStep.Type)
	}
	expectEmitStepWithString(t, thirdElseIfStep.Body, 0, "ElseIf 3", "ElseIfs[2].Body")

	finalElseSteps := thirdElseIfStep.ElseBody
	if finalElseSteps == nil || len(finalElseSteps) != 1 {
		t.Errorf("MultipleElseIfs (final ELSE/Else): Expected 1, got %v (or nil)", finalElseSteps)
	} else {
		expectEmitStepWithString(t, finalElseSteps, 0, "Else", "FinalElse.Else")
	}
}

func TestIfThenElse_NestedIfInThen(t *testing.T) {
	script := `
func TestProc() means
	if true
		emit "Outer THEN start"
		if false
			emit "Inner THEN"
		else
			emit "Inner ELSE"
		endif
		emit "Outer THEN end"
	else
		emit "Outer ELSE"
	endif
endfunc
`
	ifNodeOuter := getIfStepFromTestProc(t, script)

	if len(ifNodeOuter.Body) != 3 {
		t.Fatalf("NestedIfInThen (Outer THEN/Body): Expected 3 statements, got %d.", len(ifNodeOuter.Body))
	} else {
		expectEmitStepWithString(t, ifNodeOuter.Body, 0, "Outer THEN start", "OuterBody")
		innerIfStep := ifNodeOuter.Body[1]
		if innerIfStep.Type != "if" {
			t.Fatalf("NestedIfInThen: Expected second statement in Outer Body to be an 'if' step, got %s", innerIfStep.Type)
		}
		expectEmitStepWithString(t, ifNodeOuter.Body, 2, "Outer THEN end", "OuterBody")

		if len(innerIfStep.Body) != 1 {
			t.Errorf("NestedIfInThen (Inner THEN/Body): Expected 1 statement, got %d", len(innerIfStep.Body))
		} else {
			expectEmitStepWithString(t, innerIfStep.Body, 0, "Inner THEN", "InnerBody")
		}
		if innerIfStep.ElseBody == nil || len(innerIfStep.ElseBody) != 1 {
			t.Errorf("NestedIfInThen (Inner ELSE/Else): Expected 1 statement, got %d (or nil).", len(innerIfStep.ElseBody))
		} else {
			expectEmitStepWithString(t, innerIfStep.ElseBody, 0, "Inner ELSE", "InnerElse")
		}
	}

	if ifNodeOuter.ElseBody == nil || len(ifNodeOuter.ElseBody) != 1 {
		t.Errorf("NestedIfInThen (Outer ELSE/Else): Expected 1 statement, got %d (or nil)", len(ifNodeOuter.ElseBody))
	} else {
		expectEmitStepWithString(t, ifNodeOuter.ElseBody, 0, "Outer ELSE", "OuterElse")
	}
}

func TestIfThenElse_NestedIfInElse(t *testing.T) {
	script := `
func TestProc() means
	if false
		emit "Outer THEN"
	else
		emit "Outer ELSE start"
		if true
			emit "Inner THEN in ELSE"
		else
			emit "Inner ELSE in ELSE"
		endif
		emit "Outer ELSE end"
	endif
endfunc
`
	ifNodeOuter := getIfStepFromTestProc(t, script)

	if len(ifNodeOuter.Body) != 1 {
		t.Errorf("NestedIfInElse (Outer THEN/Body): Expected 1 statement, got %d", len(ifNodeOuter.Body))
	} else {
		expectEmitStepWithString(t, ifNodeOuter.Body, 0, "Outer THEN", "OuterBody")
	}

	if ifNodeOuter.ElseBody == nil || len(ifNodeOuter.ElseBody) != 3 {
		t.Fatalf("NestedIfInElse (Outer ELSE/Else): Expected 3 statements, got %d (or nil).", len(ifNodeOuter.ElseBody))
	} else {
		expectEmitStepWithString(t, ifNodeOuter.ElseBody, 0, "Outer ELSE start", "OuterElse")
		innerIfStep := ifNodeOuter.ElseBody[1]
		if innerIfStep.Type != "if" {
			t.Fatalf("NestedIfInElse: Expected second statement in Outer ELSE to be an 'if' step, got %s", innerIfStep.Type)
		}
		expectEmitStepWithString(t, ifNodeOuter.ElseBody, 2, "Outer ELSE end", "OuterElse")

		if len(innerIfStep.Body) != 1 {
			t.Errorf("NestedIfInElse (Inner THEN in ELSE/Body): Expected 1 statement, got %d", len(innerIfStep.Body))
		} else {
			expectEmitStepWithString(t, innerIfStep.Body, 0, "Inner THEN in ELSE", "InnerThenInElseBody")
		}
		if innerIfStep.ElseBody == nil || len(innerIfStep.ElseBody) != 1 {
			t.Errorf("NestedIfInElse (Inner ELSE in ELSE/Else): Expected 1 statement, got %d (or nil)", len(innerIfStep.ElseBody))
		} else {
			expectEmitStepWithString(t, innerIfStep.ElseBody, 0, "Inner ELSE in ELSE", "InnerElseInElse")
		}
	}
}

func TestIfThenElse_EmptyThenBlock(t *testing.T) {
	script := `
# Procedure definition
func TestProc() means
	if true
		set _ = nil
	else
		emit "In ELSE"
	endif
endfunc
`
	ifStep := getIfStepFromTestProc(t, script)

	if len(ifStep.Body) == 0 {
		t.Errorf("EmptyThenBlock: Expected at least 1 statement in Body, got 0.")
	}

	if ifStep.ElseBody == nil || len(ifStep.ElseBody) != 1 {
		t.Errorf("EmptyThenBlock (ELSE/Else): Expected 1 statement, got %d (or nil)", len(ifStep.ElseBody))
	} else {
		expectEmitStepWithString(t, ifStep.ElseBody, 0, "In ELSE", "Else")
	}
}

func TestIfThenElse_EmptyElseBlock(t *testing.T) {
	script := `
func TestProc() means
	if false
		emit "In THEN"
	else
		set _ = nil
	endif
endfunc
`
	ifStep := getIfStepFromTestProc(t, script)

	if len(ifStep.Body) != 1 {
		t.Errorf("EmptyElseBlock (THEN/Body): Expected 1 statement, got %d", len(ifStep.Body))
	} else {
		expectEmitStepWithString(t, ifStep.Body, 0, "In THEN", "Body")
	}

	if ifStep.ElseBody == nil || len(ifStep.ElseBody) == 0 {
		t.Errorf("EmptyElseBlock: Expected at least 1 Else step, got 0 or nil")
	}
}

func TestIfThenElse_EmptyThenAndElseBlocks(t *testing.T) {
	script := `
func TestProc() means
	if some_condition == true
		set _ = nil
	else
		set _ = nil
	endif
endfunc
`
	ifStep := getIfStepFromTestProc(t, script)

	if len(ifStep.Body) == 0 {
		t.Errorf("EmptyThenAndElseBlocks (THEN/Body): Expected at least 1 statement, got %d", len(ifStep.Body))
	}
	if ifStep.ElseBody == nil || len(ifStep.ElseBody) == 0 {
		t.Errorf("EmptyThenAndElseBlocks (ELSE/Else): Expected at least 1 Else step, got 0 or nil")
	}
}

func TestIfThenElse_MultipleStatementsInBlocks(t *testing.T) {
	script := `
func TestProc() means
	if some_var > 10
		emit "Then Line 1"
		set x = 100
		emit "Then Line 2"
	else
		emit "Else Line 1"
		set y = 200
	endif
endfunc
`
	ifStep := getIfStepFromTestProc(t, script)

	if len(ifStep.Body) != 3 {
		t.Errorf("MultipleStatements (THEN/Body): Expected 3 statements, got %d.", len(ifStep.Body))
	} else {
		expectEmitStepWithString(t, ifStep.Body, 0, "Then Line 1", "Body")
		setStep := ifStep.Body[1]
		if setStep.Type != "set" {
			t.Errorf("MultipleStatements (THEN/Body): Expected statement 2 to be 'set' step, got type %s", setStep.Type)
		}
		expectEmitStepWithString(t, ifStep.Body, 2, "Then Line 2", "Body")
	}
	for _, stmt := range ifStep.Body {
		if stmt.Type == "emit" {
			if len(stmt.Values) > 0 {
				if strLit, okStr := stmt.Values[0].(*ast.StringLiteralNode); okStr && strLit.Value == "Else Line 1" {
					t.Errorf("MultipleStatements (BUG CHECK): 'Else Line 1' found in Body")
					break
				}
			}
		}
	}

	if ifStep.ElseBody == nil || len(ifStep.ElseBody) != 2 {
		t.Errorf("MultipleStatements (ELSE/Else): Expected 2 statements, got %d (or nil).", len(ifStep.ElseBody))
	} else {
		expectEmitStepWithString(t, ifStep.ElseBody, 0, "Else Line 1", "Else")
		setStep := ifStep.ElseBody[1]
		if setStep.Type != "set" {
			t.Errorf("MultipleStatements (ELSE/Else): Expected statement 2 to be 'set' step, got type %s", setStep.Type)
		}
	}
}
