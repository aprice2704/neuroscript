// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Implements self-registration for the strtools toolset, including codecs and regex.
// filename: pkg/tool/strtools/register.go
package strtools

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the strtools package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	allTools := append(stringToolsToRegister, stringCodecToolsToRegister...)
	allTools = append(allTools, stringRegexToolsToRegister...)
	tool.AddToolsetRegistration(
		"strtools",
		tool.CreateRegistrationFunc("strtools", allTools),
	)
}
