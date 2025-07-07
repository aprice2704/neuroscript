// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements self-registration for the syntax toolset.
// filename: pkg/tool/syntax/register.go
package syntax

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the syntax package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	tool.AddToolsetRegistration(
		"syntax",
		tool.CreateRegistrationFunc("syntax", syntaxToolsToRegister),
	)
}
