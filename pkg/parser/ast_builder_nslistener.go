// filename: pkg/parser/ast_builder_nslistener.go
package parser

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
	"github.com/aprice2704/neuroscript/pkg/types"
)

type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	program         *ast.Program
	fileMetadata    map[string]string
	procedures      []*ast.Procedure
	events          []*ast.OnEventDecl
	commands        []*ast.CommandNode
	currentProc     *ast.Procedure
	currentCommand  *ast.CommandNode
	ValueStack      []interface{}
	blockStack      []*blockContext
	logger          interfaces.Logger
	debugAST        bool
	errors          []error
	loopDepth       int
	blockValueDepth []int
}

// --- Standard Listener Implementation ---

func (l *neuroScriptListenerImpl) EnterEveryRule(ctx antlr.ParserRuleContext) {}
func (l *neuroScriptListenerImpl) ExitEveryRule(ctx antlr.ParserRuleContext)  {}
func (l *neuroScriptListenerImpl) VisitErrorNode(node antlr.ErrorNode)        {}

func (l *neuroScriptListenerImpl) VisitTerminal(node antlr.TerminalNode) {
	if node.GetSymbol().GetTokenType() == gen.NeuroScriptLexerMETADATA_LINE && l.currentProc == nil {
		// l.processMetadataLine(l.fileMetadata, node.GetSymbol())
	}
}

var _ gen.NeuroScriptListener = (*neuroScriptListenerImpl)(nil)

func newNeuroScriptListener(logger interfaces.Logger, debugAST bool) *neuroScriptListenerImpl {
	return &neuroScriptListenerImpl{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		program:                 ast.NewProgram(),
		fileMetadata:            make(map[string]string),
		procedures:              make([]*ast.Procedure, 0),
		events:                  make([]*ast.OnEventDecl, 0),
		commands:                make([]*ast.CommandNode, 0),
		ValueStack:              make([]interface{}, 0),
		blockStack:              make([]*blockContext, 0),
		logger:                  logger,
		debugAST:                debugAST,
	}
}

func (l *neuroScriptListenerImpl) GetFileMetadata() map[string]string {
	return l.fileMetadata
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
			// Convert the VariableNode to a simple LValueNode for the set statement
			lval = &ast.LValueNode{
				Identifier: node.Name,
				Accessors:  []*ast.AccessorNode{},
				BaseNode:   node.BaseNode,
			}
		default:
			l.addError(ctx, "internal error: value for lvalue is not an ast.LValueNode or ast.VariableNode, got %T", v)
			l.push([]*ast.LValueNode{})
			return
		}
		lValues[i] = lval
	}
	l.push(lValues)
}

func (l *neuroScriptListenerImpl) ExitSet_statement(ctx *gen.Set_statementContext) {
	l.logDebugAST("ExitSet_statement: Building set step.")
	rhsVal, ok := l.pop()
	if !ok {
		l.addError(ctx, "internal error in set_statement: could not pop RHS")
		return
	}
	rhsExpr, ok := rhsVal.(ast.Expression)
	if !ok {
		l.addError(ctx, "internal error in set_statement: RHS value is not an ast.Expression, but %T", rhsVal)
		return
	}
	lhsVal, ok := l.pop()
	if !ok {
		l.addError(ctx, "internal error in set_statement: could not pop LHS")
		return
	}
	lhsExprs, ok := lhsVal.([]*ast.LValueNode)
	if !ok {
		l.addError(ctx, "internal error in set_statement: LHS value is not []*ast.LValueNode, but %T", lhsVal)
		return
	}

	pos := tokenToPosition(ctx.GetStart())
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Position: pos,
		Type:     "set",
		LValues:  lhsExprs,
		Values:   []ast.Expression{rhsExpr},
	}

	l.addStep(step)
}
