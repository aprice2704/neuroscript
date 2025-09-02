// NeuroScript Version: 0.6.0
// File version: 2
// Purpose: Defines common types for the public API to decouple from internal AST structures. FIX: Removed redundant 'Tree' definition.
// filename: pkg/api/types.go
// nlines: 25
// risk_rating: LOW

package api

import (
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// Value is an alias for lang.Value for the public API.
type Value = lang.Value

// LoadedUnit represents a fully parsed and verified script, ready for loading.
// type LoadedUnit struct {
// 	Tree *Tree
// 	// Other fields like diagnostics, etc., would go here.
// }

// Logger is an alias for the logger interface for the public API.
// type Logger = interfaces.Logger

// // AIProvider is an alias for the provider interface for the public API.
// type AIProvider = interfaces.AIProvider
