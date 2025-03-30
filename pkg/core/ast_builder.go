// pkg/core/ast_builder.go
package core

import (
	"fmt"
	"strings" // For processing comment block text

	// Use alias 'gen' for the generated package
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// neuroScriptListenerImpl implements the ANTLR listener interface
// to build our core.Procedure AST.
type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener // Embed the base listener using alias

	procedures     []Procedure // Stores the final list of procedures
	currentProc    *Procedure  // Pointer to the procedure currently being built
	currentSteps   *[]Step     // Pointer to the slice where new steps should be added
	blockStepStack []*[]Step   // Stack to manage nested block step targets (*[]Step)
}

// newNeuroScriptListener creates a new listener instance.
func newNeuroScriptListener() *neuroScriptListenerImpl {
	return &neuroScriptListenerImpl{
		procedures:     make([]Procedure, 0),
		blockStepStack: make([]*[]Step, 0), // Initialize the stack
	}
}

// GetResult returns the built AST.
func (l *neuroScriptListenerImpl) GetResult() []Procedure {
	return l.procedures
}

// --- Listener Method Implementations ---

// EnterProgram is called when parsing starts.
func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) { // Use alias
	l.procedures = make([]Procedure, 0) // Initialize
}

// EnterProcedure_definition is called when entering a procedure definition.
func (l *neuroScriptListenerImpl) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) { // Use alias
	procName := ""
	if ctx.IDENTIFIER() != nil {
		procName = ctx.IDENTIFIER().GetText()
	}

	params := []string{}
	// Use alias for sub-context types as well
	if ctx.Param_list_opt() != nil && ctx.Param_list_opt().Param_list() != nil {
		for _, id := range ctx.Param_list_opt().Param_list().AllIDENTIFIER() {
			params = append(params, id.GetText())
		}
	}

	l.currentProc = &Procedure{
		Name:   procName,
		Params: params,
		Steps:  make([]Step, 0), // Initialize steps slice
		// Docstring will be filled by EnterComment_block
	}
	// Set the current steps target to the top-level steps of this procedure
	l.currentSteps = &l.currentProc.Steps
}

// ExitProcedure_definition is called when exiting a procedure definition.
func (l *neuroScriptListenerImpl) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) { // Use alias
	if l.currentProc != nil {
		l.procedures = append(l.procedures, *l.currentProc)
	}
	l.currentProc = nil  // Reset for the next procedure
	l.currentSteps = nil // Reset step target outside any procedure
}

// EnterComment_block is called when entering a comment block.
func (l *neuroScriptListenerImpl) EnterComment_block(ctx *gen.Comment_blockContext) { // Use alias
	if l.currentProc == nil || ctx.COMMENT_BLOCK() == nil {
		return // Should not happen if grammar is correct
	}
	// Get the full text of the COMMENT_BLOCK token
	fullCommentText := ctx.COMMENT_BLOCK().GetText()

	// Basic content extraction (Remove delimiters)
	// TODO: Implement more robust parsing via parseDocstring helper
	content := strings.TrimPrefix(fullCommentText, "COMMENT:")
	content = strings.TrimSuffix(content, "END")
	content = strings.TrimSpace(content) // Trim surrounding whitespace/newlines

	// Use your existing parseDocstring or implement it here
	l.currentProc.Docstring = parseDocstring(content) // Assuming parseDocstring is defined elsewhere
}

// EnterSet_statement is called for SET statements.
func (l *neuroScriptListenerImpl) EnterSet_statement(ctx *gen.Set_statementContext) { // Use alias
	if l.currentSteps == nil {
		fmt.Println("[WARN] EnterSet_statement called outside a valid step context")
		return
	}
	varName := ""
	if ctx.IDENTIFIER() != nil {
		varName = ctx.IDENTIFIER().GetText()
	}
	// Expression is captured as raw text for now, interpreter handles evaluation
	exprText := ""
	if ctx.Expression() != nil {
		exprText = ctx.Expression().GetText()
	}

	step := newStep("SET", varName, "", exprText, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}

// EnterCall_statement is called for CALL statements.
func (l *neuroScriptListenerImpl) EnterCall_statement(ctx *gen.Call_statementContext) { // Use alias
	if l.currentSteps == nil {
		fmt.Println("[WARN] EnterCall_statement called outside a valid step context")
		return
	}
	target := ""
	if ctx.Call_target() != nil {
		target = ctx.Call_target().GetText() // Includes TOOL./LLM if present
	}

	args := []string{}
	if ctx.Expression_list_opt() != nil && ctx.Expression_list_opt().Expression_list() != nil {
		for _, expr := range ctx.Expression_list_opt().Expression_list().AllExpression() {
			// Store arguments as raw text for now
			args = append(args, expr.GetText())
		}
	}

	step := newStep("CALL", target, "", nil, args)
	*l.currentSteps = append(*l.currentSteps, step)
}

// EnterReturn_statement is called for RETURN statements
func (l *neuroScriptListenerImpl) EnterReturn_statement(ctx *gen.Return_statementContext) { // Use alias
	if l.currentSteps == nil {
		fmt.Println("[WARN] EnterReturn_statement called outside a valid step context")
		return
	}
	exprText := "" // Default for RETURN without value
	if ctx.Expression() != nil {
		exprText = ctx.Expression().GetText()
	}
	step := newStep("RETURN", "", "", exprText, nil)
	*l.currentSteps = append(*l.currentSteps, step)
}

// --- Block Handling ---

// Helper to manage entering a block statement (IF, WHILE, FOR)
func (l *neuroScriptListenerImpl) enterBlock(blockType string, condOrLoopVar string, collectionExpr string) {
	if l.currentSteps == nil {
		fmt.Printf("[WARN] Enter%s called outside a valid step context\n", blockType)
		return
	}
	blockBody := make([]Step, 0)
	step := newStep(blockType, condOrLoopVar, collectionExpr, blockBody, nil)
	*l.currentSteps = append(*l.currentSteps, step)
	l.blockStepStack = append(l.blockStepStack, l.currentSteps)

	// WORKAROUND APPROACH (See previous explanation - requires care):
	// Get pointer to the slice *within* the step we just added.
	lastStepIndex := len(*l.currentSteps) - 1
	if lastStepIndex < 0 {
		fmt.Printf("[ERROR] Could not find the %s step just added\n", blockType)
		l.currentSteps = nil // Indicate error state
		return
	}
	// We need to modify the []Step stored in the interface{} Value field
	// A safer way is needed, but this attempts to set currentSteps to point to that slice's location.
	// This is conceptually what we want, but Go makes modifying slices within interface{} tricky.
	// Re-assigning Value on exit is likely necessary.
	(*l.currentSteps)[lastStepIndex].Value = blockBody // Ensure the empty slice is in the step
	l.currentSteps = &blockBody                        // Point to local slice; will need update on exitBlock

	fmt.Printf("[DEBUG] Enter %s: Stack depth %d, currentSteps target potentially set (needs exit update)\n", blockType, len(l.blockStepStack))

}

// Helper to manage exiting a block statement
func (l *neuroScriptListenerImpl) exitBlock(blockType string) {
	if len(l.blockStepStack) == 0 {
		fmt.Printf("[ERROR] Exit%s called with empty stack!\n", blockType)
		l.currentSteps = nil
		return
	}

	// The steps for the block we just finished are in the slice currently pointed to by l.currentSteps
	finishedBlockSteps := *l.currentSteps

	// Pop the stack to get the parent step list target
	lastIndex := len(l.blockStepStack) - 1
	parentSteps := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex]

	// Find the block step in the parent list (it's the last one added before we entered the block)
	// and update its Value field with the steps we collected.
	if parentSteps != nil {
		parentStepIndex := len(*parentSteps) - 1
		if parentStepIndex >= 0 {
			// Replace the initial empty slice with the populated one
			(*parentSteps)[parentStepIndex].Value = finishedBlockSteps
			fmt.Printf("[DEBUG] Exit %s: Updated parent step %d's Value. Stack depth %d\n", blockType, parentStepIndex, len(l.blockStepStack))
		} else {
			fmt.Printf("[ERROR] Exit%s: Parent step list was empty?\n", blockType)
		}
	} else {
		fmt.Printf("[ERROR] Exit%s: Parent step list pointer was nil\n", blockType)
	}

	// Restore the currentSteps target to the parent list
	l.currentSteps = parentSteps
}

// --- Implementing Block Listeners ---

func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) { // Use alias
	condText := ""
	if ctx.Condition() != nil {
		condText = ctx.Condition().GetText()
	}
	l.enterBlock("IF", "", condText)
}

func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) { // Use alias
	l.exitBlock("IF")
}

func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) { // Use alias
	condText := ""
	if ctx.Condition() != nil {
		condText = ctx.Condition().GetText()
	}
	l.enterBlock("WHILE", "", condText)
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) { // Use alias
	l.exitBlock("WHILE")
}

func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) { // Use alias
	loopVar := ""
	if ctx.IDENTIFIER() != nil {
		loopVar = ctx.IDENTIFIER().GetText()
	}
	collectionExpr := ""
	if ctx.Expression() != nil {
		collectionExpr = ctx.Expression().GetText()
	}
	// Use Target field for loop variable, Cond field for collection expression
	l.enterBlock("FOR", loopVar, collectionExpr)
}

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) { // Use alias
	l.exitBlock("FOR")
}

// TODO: Implement methods for other statement types (ASSERT, TRY etc. if added)
// TODO: Implement methods for specific expression/term types if needed for AST later
