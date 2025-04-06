// Code generated from FencedBlockExtractor.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // FencedBlockExtractor
import "github.com/antlr4-go/antlr/v4"

// FencedBlockExtractorListener is a complete listener for a parse tree produced by FencedBlockExtractorParser.
type FencedBlockExtractorListener interface {
	antlr.ParseTreeListener

	// EnterDocument is called when entering the document production.
	EnterDocument(c *DocumentContext)

	// EnterToken is called when entering the token production.
	EnterToken(c *TokenContext)

	// ExitDocument is called when exiting the document production.
	ExitDocument(c *DocumentContext)

	// ExitToken is called when exiting the token production.
	ExitToken(c *TokenContext)
}
