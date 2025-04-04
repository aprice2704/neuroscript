// Code generated from FencedBlockExtractor.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // FencedBlockExtractor
import "github.com/antlr4-go/antlr/v4"

type BaseFencedBlockExtractorVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseFencedBlockExtractorVisitor) VisitDocument(ctx *DocumentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseFencedBlockExtractorVisitor) VisitToken(ctx *TokenContext) interface{} {
	return v.VisitChildren(ctx)
}
