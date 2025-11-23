// NeuroScript Version: 0.4.1
// File version: 4
// Purpose: Defines string constants for NeuroScript type names, and imports the foundational NeuroScriptType.
// Latest change: Imported NeuroScriptType from interfaces to avoid redefinition. Added TypeHandle constant.
// filename: pkg/lang/type_names.go
// nlines: 32
// risk_rating: LOW

package lang

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// NeuroScriptType represents the string name of a NeuroScript data type.
// It is aliased from the interfaces package to avoid redefinition and import cycles.
type NeuroScriptType = interfaces.NeuroScriptType

const (
	TypeString   NeuroScriptType = "string"
	TypeNumber   NeuroScriptType = "number" // TypeNumber represents the 'number' type (for both integers and floats).
	TypeBoolean  NeuroScriptType = "boolean"
	TypeBytes    NeuroScriptType = "bytes"
	TypeList     NeuroScriptType = "list"
	TypeMap      NeuroScriptType = "map"
	TypeNil      NeuroScriptType = "nil"
	TypeFunction NeuroScriptType = "function"
	TypeTool     NeuroScriptType = "tool"
	TypeError    NeuroScriptType = "error"
	TypeEvent    NeuroScriptType = "event"
	TypeTimedate NeuroScriptType = "timedate"
	TypeFuzzy    NeuroScriptType = "fuzzy"
	TypeHandle   NeuroScriptType = "handle"  // Added handle type
	TypeUnknown  NeuroScriptType = "unknown" // TypeUnknown represents an unknown or indeterminate type.

)
