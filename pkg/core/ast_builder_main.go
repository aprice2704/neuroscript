// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-29 // Updated for string literal un-escaping and literal handling
package core

import (
	"errors"
	"fmt" // Ensure fmt is imported

	// Added for number parsing
	"strings"

	"github.com/antlr4-go/antlr/v4"                            // Corrected ANTLR import
	gen "github.com/aprice2704/neuroscript/pkg/core/generated" // Corrected path
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// Add this method to your neuroScriptListenerImpl in pkg/core/ast_builder_main.go

// ExitLvalue is called when the lvalue rule is exited by the parser.
// It constructs an LValueNode and pushes it onto the listener's value stack.
func (l *neuroScriptListenerImpl) ExitLvalue(ctx *gen.LvalueContext) {
	l.logDebugAST("ExitLvalue: %s", ctx.GetText())

	baseIdentifierToken := ctx.IDENTIFIER(0) // Rule: IDENTIFIER ( LBRACK ... | DOT IDENTIFIER )*
	if baseIdentifierToken == nil {
		l.addErrorf(ctx.GetStart(), "AST Builder: Malformed lvalue, missing base identifier.")
		l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Malformed lvalue: missing base identifier"})
		return
	}
	baseIdentifierName := baseIdentifierToken.GetText()
	basePos := tokenToPosition(baseIdentifierToken.GetSymbol())

	lValueNode := &LValueNode{
		Pos:        basePos,
		Identifier: baseIdentifierName,
		Accessors:  make([]AccessorNode, 0),
	}

	// Expressions for bracket accessors are pushed onto the valueStack by their Exit rules.
	// We need to pop them in the reverse order of their appearance in the lvalue.
	numBracketExpressions := len(ctx.AllExpression())
	bracketExprAsts := make([]Expression, numBracketExpressions)

	// Pop expressions for bracket accessors.
	// Based on your popNValues: "Reverse to get them in parsing order".
	// So if source is a[expr1][expr2], stack top is expr2, then expr1.
	// popNValues(2) would return [expr1_node, expr2_node].
	if numBracketExpressions > 0 {
		rawExprs, ok := l.popNValues(numBracketExpressions)
		if !ok {
			// popNValues already logs an error and potentially adds to l.errors
			// Ensure an ErrorNode is pushed if the contract is to always push something.
			l.addErrorf(ctx.GetStart(), "AST Builder: Stack underflow or error popping %d expressions for lvalue '%s'", numBracketExpressions, baseIdentifierName)
			l.pushValue(&ErrorNode{Pos: basePos, Message: "Lvalue stack error: issue popping bracket expressions"})
			return
		}
		for i := 0; i < numBracketExpressions; i++ {
			expr, castOk := rawExprs[i].(Expression)
			if !castOk {
				// This error should ideally be caught if popNValues returns an error or if an ErrorNode was pushed by a failing expression rule.
				l.addErrorf(ctx.GetStart(), "AST Builder: Expected Expression on stack for lvalue '%s', got %T at index %d of popped values", baseIdentifierName, rawExprs[i], i)
				l.pushValue(&ErrorNode{Pos: basePos, Message: "Lvalue stack error: invalid bracket expression type from popNValues"})
				return
			}
			bracketExprAsts[i] = expr // Stored in source order
		}
	}

	// Iterate through the grammar elements that form accessors.
	// The lvalue rule structure from ANTLR: IDENTIFIER (LBRACK expression RBRACK | DOT IDENTIFIER)*
	// We need to walk through the accessor chain. ctx.children can be used, but ANTLR also provides
	// specific accessors like ctx.AllLBRACK(), ctx.AllDOT(), ctx.AllIDENTIFIER(), ctx.AllExpression().

	// Counters for elements we've used
	bracketExprUsed := 0
	dotIdentUsed := 0 // How many of ctx.IDENTIFIER(i>0) we've used

	// We determine the type of each accessor segment based on the order of LBRACK and DOT tokens.
	// This assumes that ANTLR provides these tokens in sequence corresponding to the source.
	// The children of LvalueContext will be the base IDENTIFIER, then a sequence of tokens/contexts
	// representing the accessors. E.g., for `a[e1].f[e2]`:
	// IDENTIFIER(a), LBRACK, Expression(e1), RBRACK, DOT, IDENTIFIER(f), LBRACK, Expression(e2), RBRACK

	// Iterate based on the number of LBRACKs and DOTs
	numLBracks := len(ctx.AllLBRACK())
	numDots := len(ctx.AllDOT())
	totalAccessors := numLBracks + numDots

	// We need to reconstruct the original order of accessors.
	// We can iterate through the children of the LvalueContext after the base IDENTIFIER.
	accessorChildren := ctx.GetChildren()[1:] // Skip the base IDENTIFIER

	currentChildPtr := 0
	for len(lValueNode.Accessors) < totalAccessors {
		if currentChildPtr >= len(accessorChildren) {
			break // Should have found all accessors
		}
		child := accessorChildren[currentChildPtr]

		if term, ok := child.(antlr.TerminalNode); ok {
			tokenType := term.GetSymbol().GetTokenType()
			accessor := AccessorNode{Pos: tokenToPosition(term.GetSymbol())}

			if tokenType == gen.NeuroScriptLexerLBRACK {
				accessor.Type = BracketAccess
				if bracketExprUsed < len(bracketExprAsts) {
					accessor.IndexOrKey = bracketExprAsts[bracketExprUsed]
					bracketExprUsed++
					lValueNode.Accessors = append(lValueNode.Accessors, accessor)
					currentChildPtr += 3 // Skip LBRACK, expression_rule_placeholder, RBRACK
					// The expression_rule_placeholder isn't directly a child TerminalNode here.
					// We've already popped the expression.
				} else {
					l.addErrorf(term.GetSymbol(), "AST Builder: Mismatch: Found LBRACK but no corresponding expression for lvalue '%s'", baseIdentifierName)
					l.pushValue(&ErrorNode{Pos: basePos, Message: "Lvalue error: LBRACK without expression"})
					return
				}
			} else if tokenType == gen.NeuroScriptLexerDOT {
				accessor.Type = DotAccess
				currentChildPtr++ // Move past DOT to the IDENTIFIER
				if currentChildPtr < len(accessorChildren) {
					fieldIdentTerm, identOk := accessorChildren[currentChildPtr].(antlr.TerminalNode)
					if identOk && fieldIdentTerm.GetSymbol().GetTokenType() == gen.NeuroScriptLexerIDENTIFIER {
						accessor.FieldName = fieldIdentTerm.GetText()
						// Optionally, update accessor.Pos to fieldIdentTerm.GetSymbol() if more precise
						lValueNode.Accessors = append(lValueNode.Accessors, accessor)
						dotIdentUsed++    // This counter isn't strictly necessary with child iteration
						currentChildPtr++ // Skip IDENTIFIER
					} else {
						l.addErrorf(term.GetSymbol(), "AST Builder: Expected IDENTIFIER after DOT in lvalue for '%s'", baseIdentifierName)
						l.pushValue(&ErrorNode{Pos: basePos, Message: "Lvalue error: DOT not followed by IDENTIFIER"})
						return
					}
				} else {
					l.addErrorf(term.GetSymbol(), "AST Builder: DOT at end of lvalue for '%s'", baseIdentifierName)
					l.pushValue(&ErrorNode{Pos: basePos, Message: "Lvalue error: DOT at end"})
					return
				}
			} else {
				// This might be an RBRACK or an unexpected token. RBRACKs are part of the LBRACK sequence.
				// If it's not LBRACK or DOT, we might just advance.
				if tokenType != gen.NeuroScriptLexerRBRACK { // RBRACKs are expected and skipped as part of LBRACK processing
					l.addErrorf(term.GetSymbol(), "AST Builder: Unexpected token '%s' while parsing lvalue accessors for '%s'", term.GetText(), baseIdentifierName)
				}
				currentChildPtr++
			}
		} else {
			// If child is not a TerminalNode, it might be an ExpressionContext (already handled by popping)
			// or an error node. For this simplified child iteration, we primarily expect tokens.
			// The expression part of `LBRACK expression RBRACK` is handled by popping from stack.
			// If an ExpressionContext itself is a child, it means the grammar is structured differently than assumed.
			// For `( A | B )*`, ANTLR makes direct context accessors for A and B, e.g. `ctx.A(i)` and `ctx.B(i)`.
			// The `children` based walk is an alternative if specific accessors are tricky.
			// Given the expression popping logic, we primarily care about LBRACK/DOT/IDENTIFIER tokens here.
			currentChildPtr++
		}
	}

	if len(lValueNode.Accessors) != totalAccessors {
		l.addErrorf(ctx.GetStart(), "AST Builder: Could not parse all accessors for lvalue '%s'. Expected %d, got %d.", baseIdentifierName, totalAccessors, len(lValueNode.Accessors))
		// Fallback or push error node
	}

	l.pushValue(lValueNode)
}

// --- Position Helper ---

// tokenToPosition converts an ANTLR token to a core.Position.
// It sets the exported fields Line, Column, and File.
func tokenToPosition(token antlr.Token) *Position {
	if token == nil {
		// *** CORRECTED LINE BELOW: Removed unexported 'token' field ***
		return &Position{Line: 0, Column: 0, File: "<nil token>"} // Return a default invalid position
	}
	// Handle potential nil InputStream or SourceName gracefully
	sourceName := "<unknown>"
	if token.GetInputStream() != nil {
		sourceName = token.GetInputStream().GetSourceName()
		if sourceName == "<INVALID>" { // Use a more descriptive name if ANTLR provides one
			sourceName = "<input stream>"
		}
	}
	// *** CORRECTED LINE BELOW: Removed unexported 'token' field ***
	return &Position{
		Line:   token.GetLine(),
		Column: token.GetColumn() + 1, // ANTLR columns are 0-based, prefer 1-based
		File:   sourceName,
		// Length: len(token.GetText()), // Add if needed by Position struct consumers
	}
}

// --- ASTBuilder (Exported Constructor and Build Method) ---

// ASTBuilder encapsulates the logic for building the NeuroScript AST using a listener.
type ASTBuilder struct {
	logger   logging.Logger
	debugAST bool // Option to enable detailed AST construction logging
}

// NewASTBuilder creates a new ASTBuilder instance.
func NewASTBuilder(logger logging.Logger) *ASTBuilder {
	if logger == nil {
		// Use the existing coreNoOpLogger from this package (defined in helpers.go)
		logger = &coreNoOpLogger{}
	}
	// Forcing debugAST true here for continued debugging.
	return &ASTBuilder{
		logger:   logger,
		debugAST: true,
	}
}

// Build takes an ANTLR parse tree and constructs the NeuroScript Program AST (*core.Program).
// It now returns the Program, the collected file metadata, and any error.
func (b *ASTBuilder) Build(tree antlr.Tree) (*Program, map[string]string, error) {
	if tree == nil {
		b.logger.Error("AST Builder: Cannot build AST from nil parse tree.")
		return nil, nil, fmt.Errorf("cannot build AST from nil parse tree")
	}
	b.logger.Debug("AST Builder: Starting AST build process using Listener.")

	// Create the listener instance.
	listener := newNeuroScriptListener(b.logger, b.debugAST) // Pass debug flag

	// Walk the parse tree with the listener.
	b.logger.Debug("AST Builder: Starting ANTLR walk...")
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)
	b.logger.Debug("AST Builder: ANTLR walk finished.")

	// Get metadata *after* the walk, before returning on error.
	fileMetadata := listener.GetFileMetadata()
	if fileMetadata == nil {
		fileMetadata = make(map[string]string)
		b.logger.Warn("AST Builder: Listener returned nil metadata map, initialized empty map.")
	}

	// Check for errors collected during the walk
	if len(listener.errors) > 0 {
		errorMessages := make([]string, 0, len(listener.errors))
		for _, err := range listener.errors {
			if err != nil {
				errorMessages = append(errorMessages, err.Error())
			} else {
				errorMessages = append(errorMessages, "<nil error recorded>")
				b.logger.Error("AST Builder: Listener recorded a nil error object.")
			}
		}
		combinedError := fmt.Errorf("AST build failed with %d error(s): %s", len(errorMessages), strings.Join(errorMessages, "; "))
		b.logger.Error("AST Builder: Errors detected during ANTLR walk.", "error", combinedError)
		return listener.program, fileMetadata, combinedError
	}

	// Get the assembled program from the listener.
	programAST := listener.program

	if programAST == nil {
		b.logger.Error("AST Builder: Build completed without explicit errors, but resulted in a nil program AST")
		return nil, fileMetadata, errors.New("AST builder internal error: program AST is unexpectedly nil after successful walk")
	}

	// Final assembly: Populate the Program's map from the listener's temporary slice.
	if programAST.Procedures == nil {
		b.logger.Warn("AST Builder: Program AST Procedures map was nil, initializing.")
		programAST.Procedures = make(map[string]*Procedure)
	}

	// --- Debug Logging Added Here ---
	b.logger.Debug("AST Builder: Assembling procedures found by listener.", "count", len(listener.procedures))
	// --- End Debug Logging ---

	duplicateProcs := false
	var firstDuplicateName string

	for i, proc := range listener.procedures { // Iterate listener's temporary slice

		// --- Debug Logging Added Here ---
		b.logger.Debug("AST Builder: Checking procedure from listener list.", "index", i, "proc_pointer", fmt.Sprintf("%p", proc))
		// --- End Debug Logging ---

		if proc != nil {
			// Check for duplicates *before* assigning.
			if existingProc, exists := programAST.Procedures[proc.Name]; exists {
				if !duplicateProcs { // Record the first duplicate found
					firstDuplicateName = proc.Name
				}
				duplicateProcs = true
				// Use the String() method of the Position struct
				errMsg := fmt.Sprintf("duplicate procedure definition: '%s' found at %s conflicts with existing at %s",
					proc.Name, proc.Pos.String(), existingProc.Pos.String())

				b.logger.Error("AST Builder: Duplicate procedure.", "name", proc.Name, "new_pos", proc.Pos.String(), "existing_pos", existingProc.Pos.String())
				listener.errors = append(listener.errors, errors.New(errMsg))
				continue // Skip adding this duplicate procedure
			}
			// Assign the non-nil proc pointer to the map only if it doesn't exist yet
			programAST.Procedures[proc.Name] = proc

		} else {
			b.logger.Warn("AST Builder encountered a nil procedure pointer in the temporary list.", "index", i)
			// Potentially make this fatal:
			// return nil, fileMetadata, fmt.Errorf("internal AST builder error: found nil procedure pointer at index %d", i)
		}
	}

	// If duplicates were found OR other errors exist, aggregate errors and return.
	if duplicateProcs || len(listener.errors) > 0 { // Check listener.errors again as we might have added new ones
		errorMessages := make([]string, 0, len(listener.errors))
		hasNonDuplicateError := false
		uniqueErrors := make(map[string]struct{}) // Avoid adding same error msg multiple times

		currentErrors := listener.errors
		for _, err := range currentErrors {
			if err != nil {
				msg := err.Error()
				if _, seen := uniqueErrors[msg]; !seen {
					errorMessages = append(errorMessages, msg)
					uniqueErrors[msg] = struct{}{}
					if !strings.Contains(msg, "duplicate procedure definition") {
						hasNonDuplicateError = true
					}
				}
			}
		}
		errorPrefix := "AST build failed"
		// Adjust error prefix logic based on whether there are any messages to join
		if len(errorMessages) > 0 {
			if duplicateProcs && !hasNonDuplicateError && strings.Contains(errorMessages[0], "duplicate procedure definition") {
				// If the first (and possibly only) error is the duplicate one
				errorPrefix = fmt.Sprintf("AST build failed due to duplicate procedure definition(s) (first: '%s')", firstDuplicateName)
			} else if duplicateProcs {
				errorPrefix = fmt.Sprintf("AST build failed due to duplicate procedure definition(s) (first: '%s') and other errors", firstDuplicateName)
			}
		} else if duplicateProcs { // Only duplicates, but errorMessages might be empty if they were all identical.
			errorPrefix = fmt.Sprintf("AST build failed due to duplicate procedure definition(s) (first: '%s')", firstDuplicateName)
			// Ensure there's at least one message if we indicate duplicates
			if len(errorMessages) == 0 {
				errorMessages = append(errorMessages, fmt.Sprintf("duplicate procedure: %s", firstDuplicateName))
			}
		}

		combinedError := fmt.Errorf("%s: %s", errorPrefix, strings.Join(errorMessages, "; "))
		b.logger.Error("AST Builder: Build failed.", "error", combinedError)
		return programAST, fileMetadata, combinedError
	}

	b.logger.Debug("AST Builder: Build process completed successfully.")
	return programAST, fileMetadata, nil // Return nil error on success
}

// --- neuroScriptListenerImpl (Internal Listener Implementation) ---
type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	program        *Program
	fileMetadata   map[string]string // Points to program.Metadata
	procedures     []*Procedure      // Temporary list of procedures built
	currentProc    *Procedure
	currentSteps   *[]Step
	blockStepStack []*[]Step
	valueStack     []interface{} // Stack for AST nodes (Expression, Step, etc.)
	currentMapKey  *StringLiteralNode
	logger         logging.Logger
	debugAST       bool
	errors         []error // For collecting parse/build errors
	loopDepth      int
}

// newNeuroScriptListener creates a new listener instance.
func newNeuroScriptListener(logger logging.Logger, debugAST bool) *neuroScriptListenerImpl {
	prog := &Program{
		Metadata:   make(map[string]string),
		Procedures: make(map[string]*Procedure),
		Pos:        nil, // Position set later in EnterProgram
	}
	return &neuroScriptListenerImpl{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		program:                 prog,
		fileMetadata:            prog.Metadata, // Directly use program's metadata map
		procedures:              make([]*Procedure, 0, 10),
		blockStepStack:          make([]*[]Step, 0, 5),
		valueStack:              make([]interface{}, 0, 20),
		logger:                  logger,
		debugAST:                debugAST,
		errors:                  make([]error, 0),
		loopDepth:               0,
	}
}

// --- Listener Error Handling ---
func (l *neuroScriptListenerImpl) addError(ctx antlr.ParserRuleContext, format string, args ...interface{}) {
	var startToken antlr.Token
	if ctx != nil {
		startToken = ctx.GetStart()
	}
	pos := tokenToPosition(startToken) // tokenToPosition handles nil token
	errMsg := fmt.Sprintf(format, args...)
	// Create a ParseError type if you have one, or just a standard error.
	// Assuming ParseError is not the standard error type for listener.errors
	err := fmt.Errorf("AST build error at %s: %s", pos.String(), errMsg)
	isDuplicate := false
	for _, existingErr := range l.errors {
		if existingErr.Error() == err.Error() {
			isDuplicate = true
			break
		}
	}
	if !isDuplicate {
		l.errors = append(l.errors, err)
		l.logger.Error("AST Build Error", "position", pos.String(), "message", errMsg)
	} else {
		l.logger.Debug("AST Build Error (Duplicate)", "position", pos.String(), "message", errMsg)
	}
}
func (l *neuroScriptListenerImpl) addErrorf(token antlr.Token, format string, args ...interface{}) {
	pos := tokenToPosition(token) // tokenToPosition handles nil token
	errMsg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("AST build error near %s: %s", pos.String(), errMsg)
	isDuplicate := false
	for _, existingErr := range l.errors {
		if existingErr.Error() == err.Error() {
			isDuplicate = true
			break
		}
	}
	if !isDuplicate {
		l.errors = append(l.errors, err)
		l.logger.Error("AST Build Error", "position", pos.String(), "message", errMsg)
	} else {
		l.logger.Debug("AST Build Error (Duplicate)", "position", pos.String(), "message", errMsg)
	}
}

// --- Listener Getters ---
func (l *neuroScriptListenerImpl) GetFileMetadata() map[string]string {
	if l.program != nil && l.program.Metadata != nil {
		return l.program.Metadata
	}
	// This should not be reached if program and its metadata are initialized in newNeuroScriptListener
	l.logger.Warn("GetFileMetadata called when listener.program.Metadata is nil.")
	return make(map[string]string) // Return empty map to avoid nil issues
}
func (l *neuroScriptListenerImpl) GetResult() []*Procedure { // This seems to be for the Build method's final assembly
	l.logger.Warn("GetResult called on listener; this returns the temporary slice for final assembly, not the final program map.")
	return l.procedures
}

// --- Listener Stack Helpers ---
func (l *neuroScriptListenerImpl) pushValue(v interface{}) { // v should be an AST Node that implements GetPos()
	if l.debugAST {
		valueStr := fmt.Sprintf("%+v", v)
		if len(valueStr) > 100 {
			valueStr = valueStr[:100] + "..."
		}
		posStr := "<unknown_pos>"
		if pnode, ok := v.(interface{ GetPos() *Position }); ok {
			if p := pnode.GetPos(); p != nil {
				posStr = p.String()
			}
		}
		l.logger.Debug("[DEBUG-AST-STACK] --> PUSH", "value_type", fmt.Sprintf("%T", v), "pos", posStr, "value_preview", valueStr, "new_stack_size", len(l.valueStack)+1)
	}
	l.valueStack = append(l.valueStack, v)
}

func (l *neuroScriptListenerImpl) popValue() (interface{}, bool) { // should return an AST Node
	if len(l.valueStack) == 0 {
		l.logger.Error("AST Builder: Pop from empty value stack!")
		l.errors = append(l.errors, errors.New("AST builder internal error: attempted pop from empty value stack"))
		// Return an ErrorNode or a distinctly identifiable error value instead of nil, if possible
		return &ErrorNode{Pos: tokenToPosition(nil), Message: "Pop from empty stack"}, false
	}
	index := len(l.valueStack) - 1
	value := l.valueStack[index]
	l.valueStack = l.valueStack[:index]
	if l.debugAST {
		valueStr := fmt.Sprintf("%+v", value)
		if len(valueStr) > 100 {
			valueStr = valueStr[:100] + "..."
		}
		posStr := "<unknown_pos>"
		if pnode, ok := value.(interface{ GetPos() *Position }); ok {
			if p := pnode.GetPos(); p != nil {
				posStr = p.String()
			}
		}
		l.logger.Debug("[DEBUG-AST-STACK] <-- POP", "value_type", fmt.Sprintf("%T", value), "pos", posStr, "value_preview", valueStr, "new_stack_size", len(l.valueStack))
	}
	return value, true
}

func (l *neuroScriptListenerImpl) popNValues(n int) ([]interface{}, bool) { // Should return []ASTNode
	if n < 0 {
		l.logger.Error("AST Builder: popNValues called with negative count", "n", n)
		l.errors = append(l.errors, fmt.Errorf("AST builder internal error: popNValues called with negative count %d", n))
		return nil, false
	}
	if n == 0 {
		return []interface{}{}, true
	}
	if len(l.valueStack) < n {
		l.logger.Error("AST Builder: Stack underflow", "needed", n, "available", len(l.valueStack))
		l.errors = append(l.errors, fmt.Errorf("AST builder internal error: stack underflow, needed %d values, only have %d", n, len(l.valueStack)))
		// To avoid panic, return what's available and indicate failure
		actualN := len(l.valueStack)
		values := make([]interface{}, actualN) // Allocate for actualN items initially
		errorNodes := 0
		if actualN > 0 {
			copy(values, l.valueStack[len(l.valueStack)-actualN:])
			l.valueStack = l.valueStack[:len(l.valueStack)-actualN]
		}
		// If more were requested than available, fill the rest with ErrorNode
		if n > actualN {
			// Ensure `values` slice has capacity for `n` items
			if cap(values) < n {
				newValues := make([]interface{}, n)
				copy(newValues, values) // Copy existing valid values
				values = newValues
			} else {
				values = values[:n] // Extend slice to n if capacity allows
			}

			for i := actualN; i < n; i++ {
				values[i] = &ErrorNode{Pos: tokenToPosition(nil), Message: "Missing value due to stack underflow"}
				errorNodes++
			}
		}

		if l.debugAST {
			l.logger.Debug("[DEBUG-AST-STACK] <-- POP N (underflow/adjusted)", "requested_n", n, "actual_n", actualN, "returned_len", len(values), "error_nodes_added", errorNodes, "new_stack_size", len(l.valueStack))
		}
		return values, false // Indicate failure because not all requested items were validly popped
	}

	startIndex := len(l.valueStack) - n
	values := make([]interface{}, n)
	copy(values, l.valueStack[startIndex:])
	l.valueStack = l.valueStack[:startIndex]
	if l.debugAST {
		l.logger.Debug("[DEBUG-AST-STACK] <-- POP N", "count", n, "new_stack_size", len(l.valueStack))
	}
	return values, true
}

// --- Listener Logging Helper ---
func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Debug(fmt.Sprintf(format, v...))
	}
}

// --- ADDED: Loop Context Helper ---
func (l *neuroScriptListenerImpl) isInsideLoop() bool {
	return l.loopDepth > 0
}

// --- Listener ANTLR Method Implementations ---
func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) {
	l.logDebugAST(">>> Enter Program")
	// Initialize ProgramNode or set its token
	if l.program == nil { // Should be initialized by newNeuroScriptListener
		l.program = &Program{
			Metadata:   make(map[string]string),
			Procedures: make(map[string]*Procedure),
		}
		l.logger.Warn("neuroScriptListenerImpl.program was nil in EnterProgram, re-initialized.")
	}
	l.program.Pos = tokenToPosition(ctx.GetStart())
	l.fileMetadata = l.program.Metadata // Ensure fileMetadata points to the program's map

	// Reset fields for potential re-use (though usually a new listener is made per build)
	l.procedures = make([]*Procedure, 0, 10)
	l.errors = make([]error, 0)
	l.valueStack = make([]interface{}, 0, 20)
	l.blockStepStack = make([]*[]Step, 0, 5)
	l.currentProc = nil
	l.currentSteps = nil
	l.currentMapKey = nil
	l.loopDepth = 0
}

func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
	finalProcCount := 0
	if l.program != nil && l.program.Procedures != nil {
		finalProcCount = len(l.program.Procedures)
	} else if l.program == nil {
		l.logger.Error("ExitProgram: l.program is nil!")
	} else {
		l.logger.Error("ExitProgram: l.program.Procedures is nil!")
	}
	metaCount := 0
	if l.fileMetadata != nil {
		metaCount = len(l.fileMetadata)
	} else {
		l.logger.Error("ExitProgram: l.fileMetadata is nil!")
	}
	l.logDebugAST("<<< Exit Program (Metadata Count: %d, Final Procedure Count: %d, Listener Errors: %d, Final Stack Size: %d)",
		metaCount, finalProcCount, len(l.errors), len(l.valueStack))

	if len(l.valueStack) != 0 {
		errMsg := fmt.Sprintf("internal AST builder error: value stack size is %d at end of program", len(l.valueStack))
		l.logger.Error("ExitProgram: Value stack not empty!", "size", len(l.valueStack), "top_value_type", fmt.Sprintf("%T", l.valueStack[len(l.valueStack)-1]))
		l.errors = append(l.errors, errors.New(errMsg))
	}
	if len(l.blockStepStack) != 0 {
		errMsg := fmt.Sprintf("internal AST builder error: block step stack size is %d at end of program", len(l.blockStepStack))
		l.logger.Error("ExitProgram: Block step stack not empty!", "size", len(l.blockStepStack))
		l.errors = append(l.errors, errors.New(errMsg))
	}
}

func (l *neuroScriptListenerImpl) EnterFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("   >> Enter File Header")
	if l.program == nil || l.program.Metadata == nil {
		l.logger.Error("EnterFile_header called with nil program or metadata map! This should have been initialized.")
		// Attempt to recover if program is nil, which is highly problematic
		if l.program == nil {
			l.program = &Program{Metadata: make(map[string]string), Procedures: make(map[string]*Procedure)}
		} else if l.program.Metadata == nil { // If only metadata is nil
			l.program.Metadata = make(map[string]string)
		}
		l.fileMetadata = l.program.Metadata // Re-assign
		l.errors = append(l.errors, errors.New("internal AST builder error: program/metadata nil in EnterFile_header"))
		// No return here, try to process metadata anyway
	}
	for _, metaLineNode := range ctx.AllMETADATA_LINE() {
		lineText := metaLineNode.GetText()
		token := metaLineNode.GetSymbol()
		l.logDebugAST("   - Processing File Metadata Line: %s", lineText)
		// The lexer rule for METADATA_LINE: [\t ]* '::' [ \t]+ ~[\r\n]* ;
		idx := strings.Index(lineText, "::")
		if idx == -1 { // Should not happen if lexer rule is correct
			l.addErrorf(token, "METADATA_LINE token did not contain '::' separator as expected: '%s'", lineText)
			continue
		}

		contentAfterDoubleColon := lineText[idx+2:] // Content after "::"
		trimmedContent := strings.TrimSpace(contentAfterDoubleColon)

		parts := strings.SplitN(trimmedContent, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				if _, exists := l.program.Metadata[key]; exists {
					l.logDebugAST("     Overwriting File Metadata: '%s'", key)
				}
				l.program.Metadata[key] = value
				l.logDebugAST("     Stored File Metadata: '%s' = '%s'", key, value)
			} else {
				l.addErrorf(token, "Ignoring file metadata line with empty key: '%s'", lineText)
			}
		} else { // Only one part, means no ':' or key is empty if content was just ':'
			keyOnly := strings.TrimSpace(parts[0])
			if keyOnly != "" {
				if _, exists := l.program.Metadata[keyOnly]; exists {
					l.logDebugAST("     Overwriting File Metadata (key only): '%s' to empty value", keyOnly)
				}
				l.program.Metadata[keyOnly] = "" // Store key with empty value
				l.logDebugAST("     Stored File Metadata (key only): '%s' = ''", keyOnly)

			} else {
				l.addErrorf(token, "Ignoring malformed file metadata line (empty key or content after '::'): '%s'", lineText)
			}
		}
	}
}

func (l *neuroScriptListenerImpl) ExitFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("   << Exit File Header")
}

func MapKeysListener(m map[string]string) []string { // Renamed to avoid conflict if core.MapKeys exists
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Note: Implementations for other Enter/Exit methods (e.g., for expressions, statements,
// list_literal, map_literal, map_entry) are expected to be in this file or other ast_builder_*.go files.
// They will need to correctly use l.pushValue and l.popValue (or l.popNValues)
// to manage the AST node stack.
