// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: FIX: Added the ExecContext constants to their authoritative source file, including the missing ContextUser.
// filename: pkg/interfaces/policy.go
// nlines: 28
// risk_rating: MEDIUM

package interfaces

import "github.com/aprice2704/neuroscript/pkg/capability"

// ExecContext defines the security context in which a script is running.
type ExecContext int

const (
	ContextConfig ExecContext = iota
	ContextNormal
	ContextTest
	ContextUser
)

// ExecPolicy defines the full set of security rules for an interpreter.
// It is defined in this neutral package to be accessible by both the `api`
// and `contract` packages without causing an import cycle.
type ExecPolicy struct {
	Context  ExecContext
	Allow    []string
	Deny     []string
	Grants   capability.GrantSet
	Parent   *ExecPolicy
	ReadOnly bool
}
