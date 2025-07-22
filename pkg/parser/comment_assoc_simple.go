// filename: pkg/parser/comment_assoc_simple.go
// NeuroScript Version: 0.6.0
// File version: 113
// Purpose: A simple, deterministic implementation of the "last-code" comment attachment rule.

package parser

import (
	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
)

// AttachCommentsSimple attaches every stand‑alone comment to the *preceding*
// code node, starting with the Program node at line 1.
func AttachCommentsSimple(program *ast.Program, ts antlr.TokenStream) {
	// 1.  Build start‑line → node map for quick lookup.
	start := map[int]ast.Node{}
	var walk func(ast.Node)
	walk = func(n ast.Node) {
		if n == nil {
			return
		}
		switch v := n.(type) {
		case *ast.Program, *ast.Procedure, *ast.Step, *ast.CommandNode, *ast.OnEventDecl:
			if pos := getNodePos(v); pos != nil && pos.Line > 0 && start[pos.Line] == nil {
				start[pos.Line] = n
			}
		}
		switch v := n.(type) {
		case *ast.Program:
			for _, p := range v.Procedures {
				walk(p)
			}
			for _, e := range v.Events {
				walk(e)
			}
			for _, c := range v.Commands {
				walk(c)
			}
		case *ast.Procedure:
			for i := range v.Steps {
				walk(&v.Steps[i])
			}
		case *ast.CommandNode:
			for i := range v.Body {
				walk(&v.Body[i])
			}
		case *ast.OnEventDecl:
			for i := range v.Body {
				walk(&v.Body[i])
			}
		case *ast.Step:
			for i := range v.Body {
				walk(&v.Body[i])
			}
			for i := range v.ElseBody {
				walk(&v.ElseBody[i])
			}
		}
	}
	walk(program)
	if start[1] == nil {
		start[1] = program
	}

	// 2.  Iterate over the token stream in source order.
	lastCodeNode := ast.Node(program)
	allTokens := ts.(*antlr.CommonTokenStream).GetAllTokens()

	for _, tok := range allTokens {
		// Update lastCodeNode if a new node starts on this token's line
		if tok.GetChannel() == antlr.TokenDefaultChannel {
			if node, ok := start[tok.GetLine()]; ok {
				lastCodeNode = node
			}
		}
		// If it's a comment, attach it to the last seen code node.
		if tok.GetTokenType() == gen.NeuroScriptLexerLINE_COMMENT {
			c := &ast.Comment{Text: tok.GetText()}
			newNode(c, tok, 0)
			assignCommentsToNode(lastCodeNode, []*ast.Comment{c})
		}
	}
}
