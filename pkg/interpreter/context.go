// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Defines the ContextKey type and centralizes the definition of context keys used for passing AEIOU v3 session data.
// filename: pkg/interpreter/context.go
// nlines: 18
// risk_rating: LOW

package interpreter

// ContextKey is the type for keys used to store and retrieve values from a context.Context.
// This prevents collisions with keys defined in other packages.
type ContextKey string

// Context keys for passing AEIOU session information through the interpreter context.
// These are defined here to avoid circular dependencies and redeclarations.
var (
	aeiouSessionIDKey = ContextKey("aeiou.sessionID")
	aeiouTurnIndexKey = ContextKey("aeiou.turnIndex")
	aeiouTurnNonceKey = ContextKey("aeiou.turnNonce")
)
