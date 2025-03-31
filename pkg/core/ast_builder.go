// pkg/core/ast_builder.go
package core

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// neuroScriptListenerImpl ... (struct definition remains the same) ...
type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	procedures     []Procedure
	currentProc    *Procedure
	currentSteps   *[]Step
	blockStepStack []*[]Step
}

// newNeuroScriptListener ... (remains the same) ...
func newNeuroScriptListener() *neuroScriptListenerImpl {
	return &neuroScriptListenerImpl{
		procedures:     make([]Procedure, 0),
		blockStepStack: make([]*[]Step, 0),
	}
}

// GetResult ... (remains the same) ...
func (l *neuroScriptListenerImpl) GetResult() []Procedure {
	return l.procedures
}

// --- Listener Method Implementations with Logging ---

// EnterEveryRule / ExitEveryRule ... (logging code remains the same) ...
func (l *neuroScriptListenerImpl) EnterEveryRule(ctx antlr.ParserRuleContext) {}
func (l *neuroScriptListenerImpl) ExitEveryRule(ctx antlr.ParserRuleContext)  {}

func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) {
	fmt.Println(">>> Enter Program")
	l.procedures = make([]Procedure, 0)
}

func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
	fmt.Println("<<< Exit Program")
}

func (l *neuroScriptListenerImpl) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	// ... (logging and setup code remains the same) ...
	procName := ""
	if ctx.IDENTIFIER() != nil {
		procName = ctx.IDENTIFIER().GetText()
	}
	fmt.Printf(">>> Enter Procedure_definition: %s\n", procName)
	params := []string{}
	if ctx.Param_list_opt() != nil && ctx.Param_list_opt().Param_list() != nil {
		for _, id := range ctx.Param_list_opt().Param_list().AllIDENTIFIER() {
			params = append(params, id.GetText())
		}
	}
	l.currentProc = &Procedure{Name: procName, Params: params, Steps: make([]Step, 0), Docstring: Docstring{InputLines: make([]string, 0), Inputs: make(map[string]string)}}
	l.currentSteps = &l.currentProc.Steps
}

func (l *neuroScriptListenerImpl) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ""
	if l.currentProc != nil {
		procName = l.currentProc.Name
	}
	fmt.Printf("<<< Exit Procedure_definition: %s\n", procName)

	// --- TEMPORARY DEBUGGING CODE ---
	// Force adding a dummy procedure to see if the slice is returned correctly
	dummyProc := Procedure{Name: fmt.Sprintf("DUMMY_FROM_%s", procName)} // Give it a unique name
	l.procedures = append(l.procedures, dummyProc)
	fmt.Printf("    DEBUG: Added dummy procedure %s. Procedures count: %d\n", dummyProc.Name, len(l.procedures))
	// --- END TEMPORARY DEBUGGING CODE ---

	if l.currentProc != nil {
		fmt.Printf("    DEBUG: Appending actual procedure: %s (Steps: %d)\n", l.currentProc.Name, len(l.currentProc.Steps))
		l.procedures = append(l.procedures, *l.currentProc) // Appends the actual built procedure
		fmt.Printf("    DEBUG: Procedures count after append: %d\n", len(l.procedures))
	} else {
		fmt.Printf("    DEBUG: l.currentProc was nil, cannot append actual procedure.\n")
	}
	l.currentProc = nil
	l.currentSteps = nil
}

// --- Docstring Line Handlers REMOVED ---

// --- Statement Handlers (logging included) ---
// ... (Keep Enter/ExitStatement, Enter/ExitSet_statement, etc. with logging from V16) ...
func (l *neuroScriptListenerImpl) EnterStatement(ctx *gen.StatementContext) {
	fmt.Printf(">>> Enter Statement: Text: %q\n", ctx.GetText())
}
func (l *neuroScriptListenerImpl) ExitStatement(ctx *gen.StatementContext) {
	fmt.Printf("<<< Exit Statement\n")
}
func (l *neuroScriptListenerImpl) EnterSet_statement(ctx *gen.Set_statementContext) {
	fmt.Printf(">>> Enter Set_statement: %q\n", ctx.GetText())
	if l.currentSteps == nil {
		return
	}
	varName := ""
	if ctx.IDENTIFIER() != nil {
		varName = ctx.IDENTIFIER().GetText()
	}
	exprText := ""
	if ctx.Expression() != nil {
		exprText = ctx.Expression().GetText()
	}
	step := newStep("SET", varName, "", exprText, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}
func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	fmt.Printf("<<< Exit Set_statement\n")
}
func (l *neuroScriptListenerImpl) EnterCall_statement(ctx *gen.Call_statementContext) {
	fmt.Printf(">>> Enter Call_statement: %q\n", ctx.GetText())
	if l.currentSteps == nil {
		return
	}
	target := ""
	if ctx.Call_target() != nil {
		target = ctx.Call_target().GetText()
	}
	args := []string{}
	if ctx.Expression_list_opt() != nil && ctx.Expression_list_opt().Expression_list() != nil {
		for _, expr := range ctx.Expression_list_opt().Expression_list().AllExpression() {
			args = append(args, expr.GetText())
		}
	}
	step := newStep("CALL", target, "", nil, args)
	*l.currentSteps = append(*l.currentSteps, step)
}
func (l *neuroScriptListenerImpl) ExitCall_statement(ctx *gen.Call_statementContext) {
	fmt.Printf("<<< Exit Call_statement\n")
}
func (l *neuroScriptListenerImpl) EnterReturn_statement(ctx *gen.Return_statementContext) {
	fmt.Printf(">>> Enter Return_statement: %q\n", ctx.GetText())
	if l.currentSteps == nil {
		return
	}
	exprText := ""
	if ctx.Expression() != nil {
		exprText = ctx.Expression().GetText()
	}
	step := newStep("RETURN", "", "", exprText, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}
func (l *neuroScriptListenerImpl) ExitReturn_statement(ctx *gen.Return_statementContext) {
	fmt.Printf("<<< Exit Return_statement\n")
}
func (l *neuroScriptListenerImpl) EnterEmit_statement(ctx antlr.ParserRuleContext) {
	fmt.Printf(">>> Enter Emit_statement: %q\n", ctx.GetText())
	if l.currentSteps == nil {
		return
	}
	exprText := ""
	if ctx.GetChildCount() > 1 {
		exprCtx := ctx.GetChild(1)
		if exprCtx != nil {
			exprText = exprCtx.(antlr.ParseTree).GetText()
		}
	}
	step := newStep("EMIT", "", "", exprText, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}
func (l *neuroScriptListenerImpl) ExitEmit_statement(ctx antlr.ParserRuleContext) {
	fmt.Printf("<<< Exit Emit_statement\n")
}

// --- Block Handling Helpers with Logging (Keep from V16) ---
// ... (Keep enterBlock and exitBlock with logging from V16) ...
func (l *neuroScriptListenerImpl) enterBlock(blockType string, targetVarOrCond string, collectionExpr string) {
	fmt.Printf(">>> enterBlock: %s (Target/Cond: %q, Collection: %q)\n", blockType, targetVarOrCond, collectionExpr)
	if l.currentSteps == nil {
		fmt.Printf("[WARN] Enter%s called outside a valid step context\n", blockType)
		return
	}
	blockBody := make([]Step, 0)
	step := newStep(blockType, targetVarOrCond, collectionExpr, blockBody, nil)
	*l.currentSteps = append(*l.currentSteps, step)
	l.blockStepStack = append(l.blockStepStack, l.currentSteps)
	tempBlockBody := make([]Step, 0)
	l.currentSteps = &tempBlockBody
}
func (l *neuroScriptListenerImpl) exitBlock(blockType string) {
	stackSize := len(l.blockStepStack)
	fmt.Printf("<<< exitBlock: %s (Stack size before pop: %d)\n", blockType, stackSize)
	if stackSize == 0 {
		fmt.Printf("[ERROR] Exit%s called with empty stack!\n", blockType)
		l.currentSteps = nil
		return
	}
	finishedBlockSteps := *l.currentSteps
	lastIndex := stackSize - 1
	parentStepsPtr := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex]
	if parentStepsPtr != nil {
		parentStepIndex := len(*parentStepsPtr) - 1
		if parentStepIndex >= 0 {
			blockStepToUpdate := &(*parentStepsPtr)[parentStepIndex]
			blockStepToUpdate.Value = finishedBlockSteps
			fmt.Printf("    Assigned %d steps to %s block step at parent index %d\n", len(finishedBlockSteps), blockType, parentStepIndex)
		} else {
			fmt.Printf("[ERROR] Exit%s: Parent step list pointer was valid but list was empty?\n", blockType)
		}
		l.currentSteps = parentStepsPtr
	} else {
		fmt.Printf("[ERROR] Exit%s: Parent step list pointer from stack was nil\n", blockType)
		l.currentSteps = nil
	}
}

// --- Block Statement Enter/Exit using Helpers (Keep from V16) ---
// ... (Keep Enter/ExitIf_statement, Enter/ExitWhile_statement, Enter/ExitFor_each_statement with logging from V16) ...
func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	fmt.Printf(">>> Enter If_statement\n")
	condText := ""
	if ctx.Condition() != nil {
		condText = ctx.Condition().GetText()
	}
	l.enterBlock("IF", "", condText)
}
func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.exitBlock("IF")
	fmt.Printf("<<< Exit If_statement\n")
}
func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	fmt.Printf(">>> Enter While_statement\n")
	condText := ""
	if ctx.Condition() != nil {
		condText = ctx.Condition().GetText()
	}
	l.enterBlock("WHILE", "", condText)
}
func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.exitBlock("WHILE")
	fmt.Printf("<<< Exit While_statement\n")
}
func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	fmt.Printf(">>> Enter For_each_statement\n")
	loopVar := ""
	if ctx.IDENTIFIER() != nil {
		loopVar = ctx.IDENTIFIER().GetText()
	}
	collectionExpr := ""
	if ctx.Expression() != nil {
		collectionExpr = ctx.Expression().GetText()
	}
	l.enterBlock("FOR", loopVar, collectionExpr)
}
func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	l.exitBlock("FOR")
	fmt.Printf("<<< Exit For_each_statement\n")
}
