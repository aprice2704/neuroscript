// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Defines context keys for passing AEIOU v3 session data.
// filename: pkg/interpreter/context.go
// nlines: 18
// risk_rating: LOW

package interpreter

// ContextKey is the type for keys used to store and retrieve values from a context.Context.
type ContextKey string

// Context keys for passing AEIOU session information through the interpreter context.
var (
	aeiouSessionIDKey = ContextKey("aeiou.sessionID")
	aeiouTurnIndexKey = ContextKey("aeiou.turnIndex")
	aeiouTurnNonceKey = ContextKey("aeiou.turnNonce")
)
