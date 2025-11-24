// NeuroScript Version: 0.8.0
// File version: 2 // Bumped version
// Purpose: Re-exports tool.ArgType constants as plain strings for the public API.
// Latest change: Added ArgTypeHandle.
// filename: pkg/api/reexport_type_strings.go
// nlines: 33
// risk_rating: LOW

package api

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// These constants re-export the string values from pkg/tool/tool_types.go
// for external API consumers, avoiding the need for them to import the
// internal 'tool' package directly.
const (
	// Primitives
	ArgTypeAny    = tool.ArgTypeAny
	ArgTypeString = tool.ArgTypeString
	ArgTypeInt    = tool.ArgTypeInt
	ArgTypeFloat  = tool.ArgTypeFloat
	ArgTypeBool   = tool.ArgTypeBool
	ArgTypeNil    = tool.ArgTypeNil
	ArgTypeHandle = tool.ArgTypeHandle // Added

	// Generic Collections
	ArgTypeMap   = tool.ArgTypeMap
	ArgTypeSlice = tool.ArgTypeSlice

	// Specific Slices
	ArgTypeSliceString = tool.ArgTypeSliceString
	ArgTypeSliceInt    = tool.ArgTypeSliceInt
	ArgTypeSliceFloat  = tool.ArgTypeSliceFloat
	ArgTypeSliceBool   = tool.ArgTypeSliceBool
	ArgTypeSliceMap    = tool.ArgTypeSliceMap
	ArgTypeSliceAny    = tool.ArgTypeSliceAny

	// Specific Maps
	ArgTypeMapStringString = tool.ArgTypeMapStringString
	ArgTypeMapStringInt    = tool.ArgTypeMapStringInt
	ArgTypeMapStringAny    = tool.ArgTypeMapStringAny
	ArgTypeMapAnyAny       = tool.ArgTypeMapAnyAny
)
