// Tool represents a registered tool.

package interfaces

import "github.com/aprice2704/neuroscript/pkg/types"

type Tool interface {
	IsTool()
	Name() types.FullName // Getter for the tool's name
}
