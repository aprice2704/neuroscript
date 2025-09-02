// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines the common structure and codes for AEIOU v3 lints.
// filename: aeiou/lints.go
// nlines: 20
// risk_rating: LOW

package aeiou

import "fmt"

// Lint represents a non-fatal warning about the envelope or output.
type Lint struct {
	Code    string
	Message string
}

// String provides a standard format for displaying lints.
func (l Lint) String() string {
	return fmt.Sprintf("[%s] %s", l.Code, l.Message)
}

// Standard Lint Codes
const (
	LintCodeDuplicateSection = "LINT_DUP_SECTION_IGNORED"
	LintCodePostTokenText    = "LINT_POST_TOKEN_TEXT"
)
