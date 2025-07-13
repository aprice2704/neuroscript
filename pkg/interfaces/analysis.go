// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines the core interfaces and data structures for static analysis.
// filename: pkg/interfaces/analysis.go
// nlines: 30
// risk_rating: MEDIUM

package interfaces

import "github.com/aprice2704/neuroscript/pkg/types"

// // Severity indicates the seriousness of a diagnostic message.
// type Severity int

// const (
// 	SeverityError   Severity = 1
// 	SeverityWarning Severity = 2
// 	SeverityInfo    Severity = 3
// 	SeverityHint    Severity = 4
// )

// // Diag represents a single diagnostic message, such as a linter error or warning.
// // It uses the foundational types.Position.
// type Diag struct {
// 	Severity Severity
// 	Position *types.Position
// 	Message  string
// 	Source   string // The name of the pass that generated the diagnostic
// }

// Pass is the interface that all static analysis passes must implement.
// It uses the foundational interfaces.Tree.
type Pass interface {
	Name() string
	Analyse(tree *Tree) []types.Diag
}
