// :: product: NS
// :: majorVersion: 1
// :: fileVersion: 3
// :: description: Re-exports tool.ArgType constants for the public API facade.
// :: latestChange: Added NodeID, EntityID, List, Blob, and Embedding for FDM compatibility.
// :: filename: pkg/api/reexport_type_strings.go
// :: serialization: go

package api

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// These constants re-export the values from pkg/tool/tool_types.go
// for external API consumers, ensuring they can use api.ArgType... constants
// without importing the internal 'tool' package.
const (
	// Primitives
	ArgTypeAny       = tool.ArgTypeAny
	ArgTypeString    = tool.ArgTypeString
	ArgTypeInt       = tool.ArgTypeInt
	ArgTypeFloat     = tool.ArgTypeFloat
	ArgTypeBool      = tool.ArgTypeBool
	ArgTypeNil       = tool.ArgTypeNil
	ArgTypeHandle    = tool.ArgTypeHandle
	ArgTypeNodeID    = tool.ArgTypeNodeID
	ArgTypeEntityID  = tool.ArgTypeEntityID
	ArgTypeBlob      = tool.ArgTypeBlob
	ArgTypeEmbedding = tool.ArgTypeEmbedding

	// Generic Collections
	ArgTypeMap   = tool.ArgTypeMap
	ArgTypeSlice = tool.ArgTypeSlice
	ArgTypeList  = tool.ArgTypeList // Alias for Slice/List compatibility

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
