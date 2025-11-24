// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Registers the 'handle' toolset.
// filename: pkg/tool/handle/register.go
// nlines: 14
// risk_rating: LOW

package handle

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the handle package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	tool.AddToolsetRegistration(
		"handle",
		tool.CreateRegistrationFunc("handle", handleToolsToRegister),
	)
}
