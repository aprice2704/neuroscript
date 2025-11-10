// NeuroScript Version: 0.8.0
// File version: 26
// Purpose: Corrected a bug where converting a VariableNode to an LValueNode copied the wrong NodeKind.
// filename: pkg/parser/ast_builder_nslistener.go
// nlines: 165
// risk_rating: MEDIUM

package parser

import (
	"fmt"
	"reflect"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/types"
)

type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	program                   *ast.Program
	pendingMetadata           map[string]string
	lastPendingMetadataToken  antlr.Token
	procedures                []*ast.Procedure
	events                    []*ast.OnEventDecl
	commands                  []*ast.CommandNode
	currentProc               *ast.Procedure
	currentCommand            *ast.CommandNode
	currentEvent              *ast.OnEventDecl
	ValueStack                []interface{}
	blockStack                []*blockContext
	logger                    interfaces.Logger
	debugAST                  bool
	errors                    []error
	loopDepth                 int
	blockValueDepth           []int
	allComments               []*ast.Comment
	lastProcessedCommentIndex int
	tokenStream               antlr.TokenStream
	eventHandlerCallback      func(decl *ast.OnEventDecl)
}

// --- Standard Listener Implementation ---

func (l *neuroScriptListenerImpl) EnterEveryRule(ctx antlr.ParserRuleContext) {}
func (l *neuroScriptListenerImpl) ExitEveryRule(ctx antlr.ParserRuleContext)  {}
func (l *neuroScriptListenerImpl) VisitErrorNode(node antlr.ErrorNode)        {}
func (l *neuroScriptListenerImpl) VisitTerminal(node antlr.TerminalNode)      {}

var _ gen.NeuroScriptListener = (*neuroScriptListenerImpl)(nil)

func newNeuroScriptListener(logger interfaces.Logger, debugAST bool, tokenStream antlr.TokenStream, cb func(decl *ast.OnEventDecl)) *neuroScriptListenerImpl {
	listener := &neuroScriptListenerImpl{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		program:                 ast.NewProgram(),
		pendingMetadata:         make(map[string]string),
		procedures:              make([]*ast.Procedure, 0),
		events:                  make([]*ast.OnEventDecl, 0),
		commands:                make([]*ast.CommandNode, 0),
		ValueStack:              make([]interface{}, 0),
		blockStack:              make([]*blockContext, 0),
		logger:                  logger,
		debugAST:                debugAST,
		tokenStream:             tokenStream,
		eventHandlerCallback:    cb,
	}

	listener.program.BaseNode.StartPos = &types.Position{Line: 1, Column: 1, File: "<source>"}
	listener.program.BaseNode.NodeKind = types.KindProgram

	return listener
}

// getNodePos uses reflection to safely get the position from a node's BaseNode.
func getNodePos(node ast.Node) *types.Position {
	v := reflect.ValueOf(node)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	baseNodeField := v.FieldByName("BaseNode")
	if !baseNodeField.IsValid() {
		return nil
	}
	if baseNodeField.CanAddr() {
		baseNode, ok := baseNodeField.Addr().Interface().(*ast.BaseNode)
		if ok && baseNode.StartPos != nil {
			return baseNode.StartPos
		}
	}
	return nil
}

func (l *neuroScriptListenerImpl) associateCommentsToNode(node ast.Node) []*ast.Comment {
	// This logic is now handled by the LineInfo algorithm in the builder.
	return nil
}

func (l *neuroScriptListenerImpl) GetFileMetadata() map[string]string {
	return l.program.Metadata
}

func (l *neuroScriptListenerImpl) GetResult() []*ast.Procedure {
	return l.procedures
}

func (l *neuroScriptListenerImpl) addError(ctx antlr.ParserRuleContext, format string, args ...interface{}) {
	token := ctx.GetStart()
	l.addErrorf(token, format, args...)
}

func (l *neuroScriptListenerImpl) addErrorf(token antlr.Token, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	pos := tokenToPosition(token)
	l.errors = append(l.errors, fmt.Errorf("AST build error near %s: %s", pos.String(), msg))
	l.logger.Error("AST Builder Error", "pos", pos.String(), "message", msg)
}

func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Debug(fmt.Sprintf(format, v...))
	}
}

func (l *neuroScriptListenerImpl) isInsideLoop() bool {
	return l.loopDepth > 0
}

func (l *neuroScriptListenerImpl) EnterProgram(c *gen.ProgramContext) {
	l.logDebugAST("EnterProgram")
}

func (l *neuroScriptListenerImpl) ExitProgram(c *gen.ProgramContext) {
	l.logDebugAST("ExitProgram: Finalizing program...")
	// Any metadata left over at the end of the file is considered file-level metadata.
	if len(l.pendingMetadata) > 0 {
		l.assignPendingMetadata(nil, nil) // Pass nil to force assignment to file.
	}

	SetEndPos(l.program, c.GetStop())

	if len(l.ValueStack) > 0 {
		l.addErrorf(c.GetStart(), "internal AST builder error: value stack size is %d at end of program", len(l.ValueStack))
	}
	if len(l.blockStack) != 0 {
		l.addErrorf(c.GetStart(), "internal AST builder error: block stack size is %d at end of program", len(l.blockStack))
	}
}

func (l *neuroScriptListenerImpl) EnterLvalue_list(ctx *gen.Lvalue_listContext) {
	l.logDebugAST("EnterLvalue_list")
}

func (l *neuroScriptListenerImpl) ExitLvalue_list(ctx *gen.Lvalue_listContext) {
	l.logDebugAST("ExitLvalue_list: Collecting assignable targets.")
	numLValues := len(ctx.AllLvalue())
	if numLValues == 0 {
		l.push([]*ast.LValueNode{})
		return
	}
	values, ok := l.popN(numLValues)
	if !ok {
		l.addError(ctx, "internal error: failed to pop values for lvalue_list")
		l.push([]*ast.LValueNode{})
		return
	}

	lValues := make([]*ast.LValueNode, numLValues)
	for i, v := range values {
		var lval *ast.LValueNode
		switch node := v.(type) {
		case *ast.LValueNode:
			lval = node
		case *ast.VariableNode:
			lval = &ast.LValueNode{
				Identifier: node.Name,
				Accessors:  []*ast.AccessorNode{},
				BaseNode:   node.BaseNode, // Copy StartPos/StopPos
			}
			// FIX: The copied BaseNode has KindVariable; it must be corrected.
			lval.BaseNode.NodeKind = types.KindLValue
		default:
			l.addError(ctx, "internal error: value for lvalue is not an ast.LValueNode or ast.VariableNode, got %T", v)
			l.push([]*ast.LValueNode{})
			return
		}
		lValues[i] = lval
	}
	l.push(lValues)
}
