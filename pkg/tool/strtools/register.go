// NeuroScript Version: 0.8.0
// File version: 5
// Purpose: Adds registration for extra string/codec tools.
// filename: pkg/tool/strtools/register.go
package strtools

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the strtools package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	allTools := append(stringToolsToRegister, stringCodecToolsToRegister...)
	allTools = append(allTools, stringRegexToolsToRegister...)
	allTools = append(allTools, stringFormatToolsToRegister...)
	allTools = append(allTools, stringExtraToolsToRegister...) // Added this line

	tool.AddToolsetRegistration(
		"strtools",
		tool.CreateRegistrationFunc("strtools", allTools),
	)
}
