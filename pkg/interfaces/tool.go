// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Cleaned interface to depend only on the 'types' package.
// filename: pkg/interfaces/tool.go
// nlines: 12
// risk_rating: LOW

package interfaces

import "github.com/aprice2704/neuroscript/pkg/types"

// Tool represents the minimal interface for a registered tool.
type Tool interface {
	IsTool()
	Name() types.FullName
}
