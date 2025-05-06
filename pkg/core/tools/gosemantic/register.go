// NeuroScript Version: 0.3.1
// File version: 0.0.5 // Keep GoFindDeclarations registered per user request.
// Registration function for gosemantic tools.
// filename: pkg/core/tools/gosemantic/register.go

package gosemantic

import (
	"github.com/aprice2704/neuroscript/pkg/core" // Import core
)

// init registers the tools defined in this package using the central mechanism.
func init() {
	core.AddToolImplementations(
		toolGoIndexCodeImpl,              // From semantic_index.go
		toolGoFindDeclarationsImpl,       // From find_declarations_lc.go (Kept as requested)
		toolGoGetDeclarationOfSymbolImpl, // From find_declarations_query.go
		toolGoFindUsagesImpl,             // From find_usages.go (Query-based)
		// Add future gosemantic tools here
	)
}
