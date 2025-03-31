// pkg/core/ast_builder.go
package core

import (
	// "fmt" // No longer needed for direct printing
	"io"  // Needed for io.Discard
	"log" // Import log package
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	procedures     []Procedure
	currentProc    *Procedure
	currentSteps   *[]Step
	blockStepStack []*[]Step
	logger         *log.Logger // Add logger
	debugAST       bool        // Add debug flag
}

// Updated constructor to accept logger and flag
func newNeuroScriptListener(logger *log.Logger, debugAST bool) *neuroScriptListenerImpl {
	// Ensure logger is non-nil
	if logger == nil {
		logger = log.New(io.Discard, "", 0) // Default to discard if nil
	}
	return &neuroScriptListenerImpl{
		procedures:     make([]Procedure, 0),
		blockStepStack: make([]*[]Step, 0),
		logger:         logger,   // Store logger
		debugAST:       debugAST, // Store flag
	}
}

func (l *neuroScriptListenerImpl) GetResult() []Procedure {
	return l.procedures
}

// Helper for conditional logging
func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Printf(format, v...)
	}
}

// --- Listener Method Implementations with Conditional Logging ---

func (l *neuroScriptListenerImpl) EnterEveryRule(ctx antlr.ParserRuleContext) {
	// l.logDebugAST(" -> Enter %s", parser.RuleNames[ctx.GetRuleIndex()]) // Requires parser instance
}
func (l *neuroScriptListenerImpl) ExitEveryRule(ctx antlr.ParserRuleContext) {
	// l.logDebugAST(" <- Exit %s", parser.RuleNames[ctx.GetRuleIndex()]) // Requires parser instance
}

func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) {
	l.logDebugAST(">>> Enter Program")
	l.procedures = make([]Procedure, 0)
}

func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
	l.logDebugAST("<<< Exit Program")
}

func (l *neuroScriptListenerImpl) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ""
	if ctx.IDENTIFIER() != nil {
		procName = ctx.IDENTIFIER().GetText()
	}
	l.logDebugAST(">>> Enter Procedure_definition: %s", procName) // Use helper
	params := []string{}
	if ctx.Param_list_opt() != nil && ctx.Param_list_opt().Param_list() != nil {
		for _, id := range ctx.Param_list_opt().Param_list().AllIDENTIFIER() {
			params = append(params, id.GetText())
		}
	}
	// Parse COMMENT_BLOCK content into Docstring struct
	docstring := Docstring{} // Default empty
	if ctx.COMMENT_BLOCK() != nil {
		commentContent := ctx.COMMENT_BLOCK().GetText()
		// Trim COMMENT: and END markers (this might need refinement based on lexer token content)
		content := strings.TrimPrefix(commentContent, "COMMENT:")
		content = strings.TrimSuffix(content, "END") // Assuming END is part of the token? If not, adjust.
		docstring = parseDocstring(strings.TrimSpace(content))
	}

	l.currentProc = &Procedure{
		Name:      procName,
		Params:    params,
		Steps:     make([]Step, 0),
		Docstring: docstring, // Store parsed docstring
	}
	l.currentSteps = &l.currentProc.Steps
}

func (l *neuroScriptListenerImpl) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ""
	if l.currentProc != nil {
		procName = l.currentProc.Name
	}
	l.logDebugAST("<<< Exit Procedure_definition: %s", procName) // Use helper

	if l.currentProc != nil {
		l.logDebugAST("    Appending actual procedure: %s (Steps: %d)", l.currentProc.Name, len(l.currentProc.Steps))
		l.procedures = append(l.procedures, *l.currentProc)
		l.logDebugAST("    Procedures count after append: %d", len(l.procedures))
	} else {
		l.logDebugAST("    l.currentProc was nil, cannot append actual procedure.")
	}
	l.currentProc = nil
	l.currentSteps = nil
}

// --- Statement Handlers (using conditional logging) ---
func (l *neuroScriptListenerImpl) EnterStatement(ctx *gen.StatementContext) {
	l.logDebugAST(">>> Enter Statement: Text: %q", ctx.GetText())
}
func (l *neuroScriptListenerImpl) ExitStatement(ctx *gen.StatementContext) {
	l.logDebugAST("<<< Exit Statement")
}
func (l *neuroScriptListenerImpl) EnterSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST(">>> Enter Set_statement: %q", ctx.GetText())
	if l.currentSteps == nil {
		l.logger.Println("[WARN] Set_statement entered with nil currentSteps")
		return
	}
	varName := ""
	if ctx.IDENTIFIER() != nil {
		varName = ctx.IDENTIFIER().GetText()
	}
	exprText := ""
	if ctx.Expression() != nil {
		exprText = ctx.Expression().GetText() // Get raw text for AST
	}
	step := newStep("SET", varName, "", exprText, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}
func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST("<<< Exit Set_statement")
}
func (l *neuroScriptListenerImpl) EnterCall_statement(ctx *gen.Call_statementContext) {
	l.logDebugAST(">>> Enter Call_statement: %q", ctx.GetText())
	if l.currentSteps == nil {
		l.logger.Println("[WARN] Call_statement entered with nil currentSteps")
		return
	}
	target := ""
	if ctx.Call_target() != nil {
		target = ctx.Call_target().GetText()
	}
	args := []string{}
	if ctx.Expression_list_opt() != nil && ctx.Expression_list_opt().Expression_list() != nil {
		for _, expr := range ctx.Expression_list_opt().Expression_list().AllExpression() {
			args = append(args, expr.GetText()) // Store raw expression text
		}
	}
	step := newStep("CALL", target, "", nil, args)
	*l.currentSteps = append(*l.currentSteps, step)
}
func (l *neuroScriptListenerImpl) ExitCall_statement(ctx *gen.Call_statementContext) {
	l.logDebugAST("<<< Exit Call_statement")
}
func (l *neuroScriptListenerImpl) EnterReturn_statement(ctx *gen.Return_statementContext) {
	l.logDebugAST(">>> Enter Return_statement: %q", ctx.GetText())
	if l.currentSteps == nil {
		l.logger.Println("[WARN] Return_statement entered with nil currentSteps")
		return
	}
	exprText := ""
	if ctx.Expression() != nil {
		exprText = ctx.Expression().GetText() // Store raw expression text
	}
	step := newStep("RETURN", "", "", exprText, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}
func (l *neuroScriptListenerImpl) ExitReturn_statement(ctx *gen.Return_statementContext) {
	l.logDebugAST("<<< Exit Return_statement")
}

// --- Block Handling Helpers with Logging ---
func (l *neuroScriptListenerImpl) enterBlock(blockType string, targetVarOrCond string, collectionExpr string) {
	l.logDebugAST(">>> enterBlock: %s (Target/Cond: %q, Collection: %q)", blockType, targetVarOrCond, collectionExpr)
	if l.currentSteps == nil {
		l.logger.Printf("[WARN] Enter%s called outside a valid step context", blockType)
		return
	}
	blockBody := make([]Step, 0) // Initialize empty block body
	// For block steps, Value will hold the []Step body eventually. Store condition/collection in Cond field.
	step := newStep(blockType, targetVarOrCond, collectionExpr, blockBody, nil)
	*l.currentSteps = append(*l.currentSteps, step)

	// Push the parent step list pointer and make the new block body the current target
	l.blockStepStack = append(l.blockStepStack, l.currentSteps)
	newBlockSteps := make([]Step, 0) // Create a new slice for the block's steps
	l.currentSteps = &newBlockSteps  // Point currentSteps to the new slice
}

func (l *neuroScriptListenerImpl) exitBlock(blockType string) {
	stackSize := len(l.blockStepStack)
	l.logDebugAST("<<< exitBlock: %s (Stack size before pop: %d)", blockType, stackSize)
	if stackSize == 0 {
		l.logger.Printf("[ERROR] Exit%s called with empty stack!", blockType)
		l.currentSteps = nil // Avoid nil pointer dereference later
		return
	}

	finishedBlockSteps := *l.currentSteps // Get the steps added to the block body

	// Pop the stack to get the pointer to the parent's step list
	lastIndex := stackSize - 1
	parentStepsPtr := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex] // Perform pop

	if parentStepsPtr == nil {
		l.logger.Printf("[ERROR] Exit%s: Parent step list pointer from stack was nil", blockType)
		l.currentSteps = nil
		return
	}

	// Find the block step added in enterBlock (should be the last one in parent)
	parentStepIndex := len(*parentStepsPtr) - 1
	if parentStepIndex < 0 {
		l.logger.Printf("[ERROR] Exit%s: Parent step list pointer was valid but list was empty?", blockType)
		l.currentSteps = parentStepsPtr // Restore parent, but something's wrong
		return
	}

	// Update the Value field of the parent's block step with the finished body
	blockStepToUpdate := &(*parentStepsPtr)[parentStepIndex]
	blockStepToUpdate.Value = finishedBlockSteps
	l.logDebugAST("    Assigned %d steps to %s block step at parent index %d", len(finishedBlockSteps), blockType, parentStepIndex)

	// Restore currentSteps to point back to the parent's list
	l.currentSteps = parentStepsPtr
}

// --- Block Statement Enter/Exit using Helpers ---
func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST(">>> Enter If_statement")
	condText := ""
	if ctx.Condition() != nil {
		condText = ctx.Condition().GetText() // Get raw condition text
	}
	l.enterBlock("IF", "", condText) // Pass condition text
}
func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.exitBlock("IF")
	l.logDebugAST("<<< Exit If_statement")
}

func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST(">>> Enter While_statement")
	condText := ""
	if ctx.Condition() != nil {
		condText = ctx.Condition().GetText() // Get raw condition text
	}
	l.enterBlock("WHILE", "", condText) // Pass condition text
}
func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.exitBlock("WHILE")
	l.logDebugAST("<<< Exit While_statement")
}

func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST(">>> Enter For_each_statement")
	loopVar := ""
	if ctx.IDENTIFIER() != nil {
		loopVar = ctx.IDENTIFIER().GetText()
	}
	collectionExpr := ""
	if ctx.Expression() != nil {
		collectionExpr = ctx.Expression().GetText() // Get raw expression text
	}
	// Store loop var in Target, collection expr in Cond for FOR steps
	l.enterBlock("FOR", loopVar, collectionExpr)
}
func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	l.exitBlock("FOR")
	l.logDebugAST("<<< Exit For_each_statement")
}

// --- Enter/Exit Emit Statement (Example, if needed) ---
func (l *neuroScriptListenerImpl) EnterEmit_statement(ctx *gen.Emit_statementContext) {
	l.logDebugAST(">>> Enter Emit_statement: %q", ctx.GetText())
	if l.currentSteps == nil {
		l.logger.Println("[WARN] Emit_statement entered with nil currentSteps")
		return
	}
	exprText := ""
	if ctx.Expression() != nil {
		exprText = ctx.Expression().GetText() // Store raw expression text
	}
	step := newStep("EMIT", "", "", exprText, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}

func (l *neuroScriptListenerImpl) ExitEmit_statement(ctx *gen.Emit_statementContext) {
	l.logDebugAST("<<< Exit Emit_statement")
}
