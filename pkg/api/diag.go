// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines the core Diagnostic type for static analysis passes.
// filename: pkg/api/diag.go
// nlines: 25
// risk_rating: LOW

package api

import "github.com/aprice2704/neuroscript/pkg/types"

// Severity indicates the seriousness of a diagnostic message.
type Severity int

const (
	SeverityError   Severity = 1
	SeverityWarning Severity = 2
	SeverityInfo    Severity = 3
	SeverityHint    Severity = 4
)

// Diag represents a single diagnostic message, such as a linter error or warning.
type Diag struct {
	Severity Severity
	Position *types.Position
	Message  string
	Source   string // The name of the pass that generated the diagnostic (e.g., "typecheck")
}
