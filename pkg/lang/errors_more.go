// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Defines additional error codes for the lang package.
// filename: pkg/lang/more_errors.go
// nlines: 10
// risk_rating: LOW

package lang

const (
	// ErrorCodeConfig indicates an error related to configuration.
	ErrorCodeConfig ErrorCode = 1001
	// ErrorCodeExternal indicates an error from an external service (e.g., an AI provider).
	ErrorCodeExternal ErrorCode = 1002
)
