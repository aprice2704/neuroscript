// NeuroScript Version: 0.6.0
// File version: 76
// Purpose: REVERT & SIMPLIFY: Implements the simplest possible deterministic comment association: all comments belong to the preceding code node.
// filename: pkg/parser/ast_builder_main.go
// nlines: 200
// risk_rating: HIGH

package parser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/logging"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// ASTBuilder encapsulates the logic for building the NeuroScript AST.
type ASTBuilder struct {
	logger                       interfaces.Logger
	debugAST                     bool
	postListenerCreationTestHook func(*neuroScriptListenerImpl) // Test hook
}

// NewASTBuilder creates a new ASTBuilder instance.
func NewASTBuilder(logger interfaces.Logger) *ASTBuilder {
	if logger == nil {
		logger = logging.NewNoOpLogger()
	}
	return &ASTBuilder{
		logger:   logger,
		debugAST: true, // Forcing debug for now
	}
}

// Build is a backward-compatible wrapper.
func (b *ASTBuilder) Build(tree antlr.Tree) (*ast.Program, map[string]string, error) {
	b.logger.Debug("--- AST Builder: Build Start (Legacy) ---")
	return b.BuildFromParseResult(tree, nil)
}

// BuildFromParseResult is the new primary build method that can process comments.
func (b *ASTBuilder) BuildFromParseResult(tree antlr.Tree, tokenStream antlr.TokenStream) (*ast.Program, map[string]string, error) {
	if tree == nil {
		b.logger.Error("AST Builder FATAL: Cannot build AST from nil parse tree.")
		return nil, nil, fmt.Errorf("cannot build AST from nil parse tree")
	}

	listener := newNeuroScriptListener(b.logger, b.debugAST)

	if tokenStream != nil {
		if commonTokenStream, ok := tokenStream.(*antlr.CommonTokenStream); ok {
			allTokens := commonTokenStream.GetAllTokens()
			for _, token := range allTokens {
				if token.GetChannel() == antlr.TokenHiddenChannel && token.GetTokenType() == gen.NeuroScriptLexerLINE_COMMENT {
					comment := &ast.Comment{Text: token.GetText()}
					newNode(comment, token, 0)
					listener.allComments = append(listener.allComments, comment)
				}
			}
		} else {
			b.logger.Warn("Could not cast token stream to CommonTokenStream; comments will not be processed.")
		}
	} else {
		b.logger.Warn("Token stream was nil; comments will not be processed.")
	}

	if b.postListenerCreationTestHook != nil {
		b.postListenerCreationTestHook(listener)
	}

	b.logger.Debug("AST Builder: Starting ANTLR walk...")
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)
	b.logger.Debug("AST Builder: ANTLR walk finished.")

	programAST := listener.program
	fileMetadata := listener.GetFileMetadata()

	b.associateAllComments(programAST, listener.allComments)

	if len(listener.errors) > 0 {
		var errorMessages []string
		uniqueErrors := make(map[string]bool)
		for _, err := range listener.errors {
			if err != nil {
				msg := err.Error()
				if !uniqueErrors[msg] {
					errorMessages = append(errorMessages, msg)
					uniqueErrors[msg] = true
				}
			}
		}
		if len(errorMessages) > 0 {
			combinedError := fmt.Errorf("AST build failed with %d error(s): %s", len(errorMessages), strings.Join(errorMessages, "; "))
			b.logger.Error("AST Builder: Failing build", "error", combinedError)
			return programAST, fileMetadata, combinedError
		}
	}
	if len(listener.ValueStack) > 0 {
		return nil, fileMetadata, fmt.Errorf("internal AST builder error: value stack size is %d at end of program", len(listener.ValueStack))
	}

	b.logger.Debug("--- AST Builder: Build process completed successfully. ---")
	return programAST, fileMetadata, nil
}

// associateAllComments implements the simplest deterministic algorithm: all comments belong to the preceding node.
func (b *ASTBuilder) associateAllComments(program *ast.Program, comments []*ast.Comment) {
	if len(comments) == 0 {
		return
	}

	type timelineItem struct {
		isComment bool
		node      ast.Node
		comment   *ast.Comment
	}

	var timeline []timelineItem
	timeline = append(timeline, timelineItem{isComment: false, node: program})

	for _, proc := range program.Procedures {
		timeline = append(timeline, timelineItem{isComment: false, node: proc})
		for i := range proc.Steps {
			timeline = append(timeline, timelineItem{isComment: false, node: &proc.Steps[i]})
		}
	}
	for _, comment := range comments {
		timeline = append(timeline, timelineItem{isComment: true, comment: comment})
	}

	sort.SliceStable(timeline, func(i, j int) bool {
		var posI, posJ *types.Position
		if timeline[i].isComment {
			posI = getNodePos(timeline[i].comment)
		} else {
			posI = getNodePos(timeline[i].node)
		}
		if timeline[j].isComment {
			posJ = getNodePos(timeline[j].comment)
		} else {
			posJ = getNodePos(timeline[j].node)
		}
		if posI == nil || posJ == nil {
			return false
		}
		if posI.Line != posJ.Line {
			return posI.Line < posJ.Line
		}
		return posI.Column < posJ.Column
	})

	fmt.Println("\n--- DEBUG: CENTRALIZED associateAllComments (ULTRA-SIMPLE ALGO) ---")
	var lastCodeNode ast.Node = nil
	var pendingComments []*ast.Comment

	for _, item := range timeline {
		if item.isComment {
			pendingComments = append(pendingComments, item.comment)
		} else {
			currentNode := item.node
			if len(pendingComments) > 0 {
				assignCommentsToNode(lastCodeNode, pendingComments)
				pendingComments = []*ast.Comment{}
			}
			lastCodeNode = currentNode
		}
	}

	if len(pendingComments) > 0 {
		assignCommentsToNode(lastCodeNode, pendingComments)
	}
	fmt.Println("--- FINISHED CENTRALIZED associateAllComments (ULTRA-SIMPLE ALGO) ---")
}

// assignCommentsToNode is a helper to append comments to the correct field in a node.
func assignCommentsToNode(node ast.Node, comments []*ast.Comment) {
	if node == nil || len(comments) == 0 {
		if node == nil && len(comments) > 0 {
			fmt.Println("   [ASSIGNING DEBUG] WARNING: Attempted to assign comments to a nil node.")
		}
		return
	}
	fmt.Printf("   [ASSIGNING DEBUG] Assigning %d comments to Node %T at Line %d\n", len(comments), node, getNodePos(node).Line)
	for _, c := range comments {
		fmt.Printf("     - Comment: '%s' (Line %d)\n", strings.TrimSpace(c.Text), getNodePos(c).Line)
	}

	switch n := node.(type) {
	case *ast.Program:
		n.Comments = append(n.Comments, comments...)
	case *ast.Procedure:
		if n.Comments == nil {
			n.Comments = make([]*ast.Comment, 0)
		}
		n.Comments = append(n.Comments, comments...)
	case *ast.Step:
		if n.Comments == nil {
			n.Comments = make([]*ast.Comment, 0)
		}
		n.Comments = append(n.Comments, comments...)
	}
}
