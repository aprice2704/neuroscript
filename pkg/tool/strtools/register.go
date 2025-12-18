// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 6
// :: description: Adds registration for ANSI string tools.
// :: latestChange: Registered stringAnsiToolsToRegister.
// :: filename: pkg/tool/strtools/register.go
// :: serialization: go

package strtools

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the strtools package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	allTools := append(stringToolsToRegister, stringCodecToolsToRegister...)
	allTools = append(allTools, stringRegexToolsToRegister...)
	allTools = append(allTools, stringFormatToolsToRegister...)
	allTools = append(allTools, stringExtraToolsToRegister...)
	allTools = append(allTools, stringAnsiToolsToRegister...) // Added ANSI tools

	tool.AddToolsetRegistration(
		"strtools",
		tool.CreateRegistrationFunc("strtools", allTools),
	)
}
