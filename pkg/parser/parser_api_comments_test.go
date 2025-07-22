// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Adds a dedicated test to verify that the parser's token stream includes comments from the hidden channel.
// filename: pkg/parser/parser_api_comments_test.go
// nlines: 45
// risk_rating: LOW

package parser

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

func TestParse_IncludesCommentsInStream(t *testing.T) {
	scriptWithComments := `
# File header comment
:: key: value
func main() means # Trailing comment
    -- Statement comment
    set x = 1
endfunc
`
	logger := logging.NewTestLogger(t)
	parserAPI := NewParserAPI(logger)

	// We need to inspect the token stream, so we use the internal parse method.
	_, stream, errs := parserAPI.parseInternal("comments_test.ns", scriptWithComments)
	if len(errs) > 0 {
		t.Fatalf("parseInternal failed unexpectedly: %v", errs)
	}

	commonTokenStream, ok := stream.(*antlr.CommonTokenStream)
	if !ok {
		t.Fatalf("Expected a *antlr.CommonTokenStream, but got %T", stream)
	}

	allTokens := commonTokenStream.GetAllTokens()
	commentCount := 0
	for _, token := range allTokens {
		if token.GetTokenType() == gen.NeuroScriptLexerLINE_COMMENT {
			commentCount++
		}
	}

	// There are 3 comments in the script.
	if commentCount < 3 {
		t.Errorf("Expected at least 3 comment tokens in the stream, but found %d", commentCount)
	}
}
