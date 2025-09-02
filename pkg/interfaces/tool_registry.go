// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Cleaned interface to depend only on 'types' and other 'interfaces' files.
// filename: pkg/interfaces/tool_registry.go
// nlines: 18
// risk_rating: MEDIUM

package interfaces

import (
	"github.com/aprice2704/neuroscript/pkg/types"
)

// ToolRegistry defines the interface for a complete tool registry.
// It uses `any` for value types to avoid a dependency on the 'lang' or 'data' package.
type ToolRegistry interface {
	GetTool(name types.FullName) (Tool, bool)
	GetToolShort(group types.ToolGroup, name types.ToolName) (Tool, bool)
	ListTools() []Tool
	NTools() int
	ExecuteTool(toolName types.FullName, args map[string]any) (any, error)
}
