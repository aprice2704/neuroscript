// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements self-registration for the tree toolset.
// filename: pkg/tool/tree/register.go
package tree

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the tree package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	tool.AddToolsetRegistration(
		"tree",
		tool.CreateRegistrationFunc("tree", treeToolsToRegister),
	)
}
