// filename: pkg/parser/comment_assoc_simple.go
// NeuroScript Version: 0.8.0
// File version: 116
// Purpose: A new, robust comment attachment algorithm that assigns floating comments to the *next* node they precede.
// nlines: 115

package parser

import (
	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types" // Import types package
)

// AttachCommentsSimple attaches comments to the AST nodes they are
// associated with, based on source position.
//
// Logic:
//  1. Standalone comments (on their own lines) are "floating" and attach
//     to the *next* code node that appears.
//  2. Trailing comments (on the same line as code) attach to the
//     *current* code node.
//  3. File-header comments (before any code) attach to the Program.
func AttachCommentsSimple(program *ast.Program, ts antlr.TokenStream) {
	allTokens := ts.(*antlr.CommonTokenStream).GetAllTokens()

	var floatingComments []*ast.Comment
	var lastCodeLine int = -1
	var lastNode ast.Node = program // Start with the program node

	// 1. Build a map of line -> node for all code nodes
	startLineToNode := map[int]ast.Node{}
	var walk func(ast.Node)
	walk = func(n ast.Node) {
		if n == nil {
			return
		}
		// We only map "code" nodes
		switch v := n.(type) {
		case *ast.Program, *ast.Procedure, *ast.Step, *ast.CommandNode, *ast.OnEventDecl:
			if pos := getNodePos(v); pos != nil && pos.Line > 0 {
				startLineToNode[pos.Line] = n
			}
		}

		// Recurse
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
			for _, eh := range v.ErrorHandlers {
				walk(eh)
			}
		case *ast.CommandNode:
			for i := range v.Body {
				walk(&v.Body[i])
			}
			for _, eh := range v.ErrorHandlers {
				walk(eh)
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
	startLineToNode[1] = program // Ensure line 1 always maps to Program

	// 2. Iterate the token stream and assign comments
	for _, tok := range allTokens {
		tokenLine := tok.GetLine()

		// Is this a code token?
		if tok.GetChannel() == antlr.TokenDefaultChannel {
			lastCodeLine = tokenLine
			// Is this the *start* of a new node?
			if node, ok := startLineToNode[tokenLine]; ok {
				// Assign all pending floating comments to this new node
				if len(floatingComments) > 0 {
					assignCommentsToNode(node, floatingComments)
					floatingComments = nil // Clear the buffer
				}
				lastNode = node
			}
		}

		// Is this a comment token?
		if tok.GetTokenType() == gen.NeuroScriptLexerLINE_COMMENT {
			comment := &ast.Comment{Text: tok.GetText()}
			// FIX: The NodeKind was being set to 0 (KindUnknown).
			// It must be set to the correct type for the canon package.
			newNode(comment, tok, types.KindComment)

			if tokenLine == lastCodeLine {
				// This is a trailing comment. Attach to the current node.
				assignCommentsToNode(lastNode, []*ast.Comment{comment})
			} else {
				// This is a floating comment. Add to buffer for the *next* node.
				floatingComments = append(floatingComments, comment)
			}
		}
	}

	// Any remaining floating comments at the end of the file
	// attach to the last node seen (which is often the Program).
	if len(floatingComments) > 0 {
		assignCommentsToNode(lastNode, floatingComments)
	}
}
