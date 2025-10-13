// NeuroScript Version: 0.8.0
// File version: 119
// Purpose: Removed temporary debug Fprintf statements after fixing the event handler panic.
// filename: pkg/parser/ast_builder_main.go
// nlines: 95
// risk_rating: LOW

package parser

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

/* ────────────────────────────── Builder plumbing ─────────────────────────── */

type ASTBuilder struct {
	logger                       interfaces.Logger
	debugAST                     bool
	postListenerCreationTestHook func(*neuroScriptListenerImpl)
	eventHandlerCallback         func(decl *ast.OnEventDecl)
}

func NewASTBuilder(l interfaces.Logger) *ASTBuilder {
	if l == nil {
		l = logging.NewNoOpLogger()
	}
	return &ASTBuilder{logger: l, debugAST: true}
}

func (b *ASTBuilder) Build(tree antlr.Tree) (*ast.Program, map[string]string, error) {
	return b.BuildFromParseResult(tree, nil)
}

func (b *ASTBuilder) BuildFromParseResult(tree antlr.Tree, ts antlr.TokenStream) (*ast.Program, map[string]string, error) {
	if tree == nil {
		return nil, nil, fmt.Errorf("cannot build AST from nil parse tree")
	}

	l := newNeuroScriptListener(b.logger, b.debugAST, ts, b.eventHandlerCallback)
	if b.postListenerCreationTestHook != nil {
		b.postListenerCreationTestHook(l)
	}
	antlr.NewParseTreeWalker().Walk(l, tree)

	prog := l.program
	meta := l.GetFileMetadata()

	if ts != nil && prog != nil {
		AttachCommentsSimple(prog, ts)
	}

	if len(l.errors) > 0 {
		uniq := map[string]bool{}
		var msgs []string
		for _, e := range l.errors {
			if e != nil && !uniq[e.Error()] {
				msgs = append(msgs, e.Error())
				uniq[e.Error()] = true
			}
		}
		if len(msgs) > 0 {
			return prog, meta, fmt.Errorf("AST build failed: %s", strings.Join(msgs, "; "))
		}
	}
	if len(l.ValueStack) > 0 {
		return nil, meta, fmt.Errorf("internal AST builder error: value stack not empty")
	}
	return prog, meta, nil
}

// SetEventHandlerCallback sets a callback to be invoked for each event handler declaration.
func (b *ASTBuilder) SetEventHandlerCallback(cb func(decl *ast.OnEventDecl)) {
	b.eventHandlerCallback = cb
}

/* ───────────────────────────── helper functions ─────────────────────────── */

func assignCommentsToNode(n ast.Node, cs []*ast.Comment) {
	if n == nil || len(cs) == 0 {
		return
	}
	switch v := n.(type) {
	case *ast.Program:
		v.Comments = append(v.Comments, cs...)
	case *ast.Procedure:
		v.Comments = append(v.Comments, cs...)
	case *ast.Step:
		v.Comments = append(v.Comments, cs...)
	case *ast.OnEventDecl:
		v.Comments = append(v.Comments, cs...)
	case *ast.CommandNode:
		v.Comments = append(v.Comments, cs...)
	}
}

func setBlankLinesOnNode(n ast.Node, c int) {
	// This function is now a no-op as blank line counting has been removed.
	// It is kept for now to prevent breaking changes in other files, but can be removed later.
}
