// NeuroScript Version: 0.7.4
// File version: 1
// Purpose: Provides a helper for creating an AEIOU v3 turn context.
// filename: pkg/aeiou/context.go
// nlines: 20
// risk_rating: LOW

package aeiou

import "context"

// contextKey is a private type to prevent key collisions.
type contextKey string

const (
	// SessionIDKey is the context key for the AEIOU session ID.
	SessionIDKey contextKey = "aeiou.sessionID"
	// TurnIndexKey is the context key for the AEIOU turn index.
	TurnIndexKey contextKey = "aeiou.turnIndex"
	// TurnNonceKey is the context key for the AEIOU turn nonce.
	TurnNonceKey contextKey = "aeiou.turnNonce"
)

// ContextWithSessionID creates a new context with the given AEIOU session ID.
func ContextWithSessionID(ctx context.Context, sid string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sid)
}
