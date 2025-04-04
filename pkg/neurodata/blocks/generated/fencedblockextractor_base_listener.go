// Code generated from FencedBlockExtractor.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // FencedBlockExtractor
import "github.com/antlr4-go/antlr/v4"

// BaseFencedBlockExtractorListener is a complete listener for a parse tree produced by FencedBlockExtractorParser.
type BaseFencedBlockExtractorListener struct{}

var _ FencedBlockExtractorListener = &BaseFencedBlockExtractorListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseFencedBlockExtractorListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseFencedBlockExtractorListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseFencedBlockExtractorListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseFencedBlockExtractorListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterDocument is called when production document is entered.
func (s *BaseFencedBlockExtractorListener) EnterDocument(ctx *DocumentContext) {}

// ExitDocument is called when production document is exited.
func (s *BaseFencedBlockExtractorListener) ExitDocument(ctx *DocumentContext) {}

// EnterToken is called when production token is entered.
func (s *BaseFencedBlockExtractorListener) EnterToken(ctx *TokenContext) {}

// ExitToken is called when production token is exited.
func (s *BaseFencedBlockExtractorListener) ExitToken(ctx *TokenContext) {}
