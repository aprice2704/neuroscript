// NeuroScript Version: 0.3.1 // Assuming current project version
// File version: 0.1.0
// Purpose: Defines string constants for NeuroScript type names returned by 'typeof'.
// filename: pkg/core/type_names.go
// nlines: 20
// risk_rating: LOW

package core

// NeuroScriptType represents the string name of a NeuroScript data type.
type NeuroScriptType string

const (
	// TypeString represents the 'string' type.
	TypeString NeuroScriptType = "string"
	// TypeNumber represents the 'number' type (for both integers and floats).
	TypeNumber NeuroScriptType = "number"
	// TypeBoolean represents the 'boolean' type.
	TypeBoolean NeuroScriptType = "boolean"
	// TypeList represents the 'list' type.
	TypeList NeuroScriptType = "list"
	// TypeMap represents the 'map' type.
	TypeMap NeuroScriptType = "map"
	// TypeNil represents the 'nil' type.
	TypeNil NeuroScriptType = "nil"
	// TypeFunction represents the 'function' type.
	TypeFunction NeuroScriptType = "function"
	// TypeTool represents the 'tool' type.
	TypeTool NeuroScriptType = "tool"
	// TypeError represents the 'error' type (if typeof can be used on error values directly).
	TypeError NeuroScriptType = "error" // Or "err" if preferred for brevity
	// TypeUnknown represents an unknown or indeterminate type.
	TypeUnknown NeuroScriptType = "unknown"
)
