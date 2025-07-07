// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements self-registration for the strtools toolset.
// filename: pkg/tool/strtools/register.go
package strtools

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the strtools package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	tool.AddToolsetRegistration(
		"strtools",
		tool.CreateRegistrationFunc("strtools", stringToolsToRegister),
	)
}
