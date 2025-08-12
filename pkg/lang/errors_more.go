// NeuroScript Version: 0.6.0
// File version: 2.0.0
// Purpose: Defines additional error codes for the lang package, including Policy and Provider errors.
// filename: pkg/lang/errors_more.go
// nlines: 15
// risk_rating: LOW

package lang

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
