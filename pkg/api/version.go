// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Exposes the program and grammar version constants for the public API.
// filename: pkg/api/version.go
// nlines: 12
// risk_rating: LOW

package api

import "github.com/aprice2704/neuroscript/pkg/lang"

const (
	// ProgramVersion is the semantic version of the NeuroScript program.
	ProgramVersion = "0.6.0"
)

// GrammarVersion is the version of the underlying NeuroScript grammar.
var GrammarVersion = lang.GrammarVersion
