// NeuroScript Version: 0.6.0
// File version: 3
// Purpose: Defines additional error codes and adds a sentinel error for policy violations.
// filename: pkg/lang/errors_more.go
// nlines: 20
// risk_rating: LOW

package lang

import "errors"

const (
	// ErrorCodeConfig indicates an error related to configuration.
	ErrorCodeConfig ErrorCode = 1001
	// ErrorCodeExternal indicates an error from an external service (e.g., an AI provider).
	ErrorCodeExternal ErrorCode = 1002

	// ErrorCodePolicy indicates a call was rejected by a security or execution policy.
	ErrorCodePolicy ErrorCode = 1003
	// ErrorCodeProviderNotFound indicates a configured AI provider could not be found.
	ErrorCodeProviderNotFound ErrorCode = 1004
)

var (
	// ErrPolicyViolation is returned when a call is rejected by a security or execution policy.
	ErrPolicyViolation = errors.New("policy violation")
)
