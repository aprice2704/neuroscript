package lang

import "fmt"

// Position represents a location in the source code.
type Position struct {
	Line   int
	Column int
	File   string
}

func (p *Position) String() string {
	if p == nil {
		return "<nil position>"
	}
	return fmt.Sprintf("line %d, col %d", p.Line, p.Column)
}
