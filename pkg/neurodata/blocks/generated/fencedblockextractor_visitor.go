// Code generated from FencedBlockExtractor.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // FencedBlockExtractor
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by FencedBlockExtractorParser.
type FencedBlockExtractorVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by FencedBlockExtractorParser#document.
	VisitDocument(ctx *DocumentContext) interface{}

	// Visit a parse tree produced by FencedBlockExtractorParser#token.
	VisitToken(ctx *TokenContext) interface{}
}
