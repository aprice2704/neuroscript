package api

// Tree wraps the root node and a comment table for fmt round-trip.
type Tree struct {
	Root     Node
	Comments []Comment
}

type Comment struct {
	Pos  Position
	Text string // original // or /* */ text
}
