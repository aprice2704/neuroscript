// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines the foundational Position struct for source code locations.
// filename: pkg/types/position.go
// nlines: 20
// risk_rating: LOW

package types

import "fmt"

// Position represents a specific location in a source file.
type Position struct {
	Line   int    // 1-based line number
	Column int    // 1-based column (character) number
	File   string // The name or path of the source file
}

// String returns a human-readable representation of the position.
func (p Position) String() string {
	if p.File == "" {
		return fmt.Sprintf("%d:%d", p.Line, p.Column)
	}
	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
}
