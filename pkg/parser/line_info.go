// filename: pkg/parser/line_info.go
// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Defines the LineInfo structure for a more robust comment and blank line association algorithm.
// nlines: 50
// risk_rating: MEDIUM

package parser

// LineKind represents the type of content on a source code line.
type LineKind int

const (
	LineKindUnknown LineKind = iota
	LineKindBlank
	LineKindComment
	LineKindMetadata
	LineKindCode // Any line with a non-comment, non-metadata token
)

// LineInfo holds structured information about a single line of source code.
// This is the core of the new, more robust "line picture" algorithm.
// type LineInfo struct {
// 	LineNumber int
// 	Kind       LineKind
// 	Content    string         // The raw text of the line
// 	Node       ast.Node       // The primary AST node starting on this line
// 	Comments   []*ast.Comment // Comments found on this line
// }

func (lk LineKind) String() string {
	switch lk {
	case LineKindBlank:
		return "Blank"
	case LineKindComment:
		return "Comment"
	case LineKindMetadata:
		return "Metadata"
	case LineKindCode:
		return "Code"
	default:
		return "Unknown"
	}
}
