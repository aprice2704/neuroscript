// NeuroScript Version: 0.4.1
// File version: 2
// Purpose: Defines string constants for NeuroScript type names, including new error, event, timedate, and fuzzy types.
// filename: core/type_names.go
// nlines: 27
// risk_rating: LOW

package core

// NeuroScriptType represents the string name of a NeuroScript data type.
type NeuroScriptType string

const (
	TypeString   NeuroScriptType = "string"
	TypeNumber   NeuroScriptType = "number" // TypeNumber represents the 'number' type (for both integers and floats).
	TypeBoolean  NeuroScriptType = "boolean"
	TypeList     NeuroScriptType = "list"
	TypeMap      NeuroScriptType = "map"
	TypeNil      NeuroScriptType = "nil"
	TypeFunction NeuroScriptType = "function"
	TypeTool     NeuroScriptType = "tool"
	TypeError    NeuroScriptType = "error"
	TypeEvent    NeuroScriptType = "event"
	TypeTimedate NeuroScriptType = "timedate"
	TypeFuzzy    NeuroScriptType = "fuzzy"
	TypeUnknown  NeuroScriptType = "unknown" // TypeUnknown represents an unknown or indeterminate type.

)
