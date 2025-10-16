// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Defines and EXPORTS context keys for passing AEIOU v3 session data, making them accessible to other packages.
// filename: pkg/interpreter/context.go
// nlines: 20
// risk_rating: LOW

package interpreter

// ContextKey is the type for keys used to store and retrieve values from a context.Context.
type ContextKey string

// Context keys for passing AEIOU session information through the interpreter context.
// These are exported to be accessible by tools and other packages that need to
// inspect the interpreter's turn-specific context.
var (
	AeiouSessionIDKey = ContextKey("aeiou.sessionID")
	AeiouTurnIndexKey = ContextKey("aeiou.turnIndex")
	AeiouTurnNonceKey = ContextKey("aeiou.turnNonce")
)
