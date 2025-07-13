// NeuroScript Version: 0.5.2
// File version: 4
// Purpose: Implements the public parsing entrypoint with full error handling.
// filename: pkg/api/parse.go
// nlines: 31
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
func Parse(src []byte, mode ParseMode) (*Tree, error) {
	// A full implementation would pass the mode to the parser to handle comments.
	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	antlrTree, err := parserAPI.Parse(string(src))
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	builder := parser.NewASTBuilder(logging.NewNoOpLogger())
	program, _, err := builder.Build(antlrTree)
	if err != nil {
		return nil, fmt.Errorf("AST construction failed: %w", err)
	}

	// The api.Tree is an alias for the foundational tree type from interfaces.
	return &Tree{Root: program}, nil
}
