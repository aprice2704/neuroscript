// NeuroScript Version: 0.6.0
// File version: 8
// Purpose: Implements the public parsing entrypoint with full error handling.
// filename: pkg/api/parse.go
// nlines: 34
// risk_rating: MEDIUM

package api

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// ParseMode controls parsing behavior, like comment handling.
type ParseMode uint8

const (
	ParsePreserveComments ParseMode = 1 << iota
	ParseSkipComments
)

// Parse converts a byte slice of NeuroScript source into an AST.
// It now uses the full parsing pipeline to ensure formatting is preserved.
func Parse(src []byte, mode ParseMode) (*Tree, error) {
	logger := logging.NewNoOpLogger()
	parserAPI := parser.NewParserAPI(logger)

	// Note: ParseAndGetStream is used to get the token stream, which is
	// now required for the builder's automatic comment association.
	antlrTree, tokenStream, err := parserAPI.ParseAndGetStream("source.ns", string(src))
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	builder := parser.NewASTBuilder(logger)

	// FIX: The call signature for BuildFromParseResult was corrected.
	// The new comment association runs automatically if the tokenStream is not nil.
	program, _, err := builder.BuildFromParseResult(antlrTree, tokenStream)
	if err != nil {
		return nil, fmt.Errorf("AST construction failed: %w", err)
	}

	return &Tree{Root: program}, nil
}
