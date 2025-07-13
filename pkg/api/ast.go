package api

// Comment represents a comment in the source code.
type Comment struct {
	Pos  Position
	Text string // original // or /* */ text
}
