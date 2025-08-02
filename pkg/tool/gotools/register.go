// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements self-registration for the gotools toolset.
// filename: pkg/tool/gotools/register.go
package gotools

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the gotools package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	tool.AddToolsetRegistration(
		"gotools",
		tool.CreateRegistrationFunc("gotools", GoToolsToRegister),
	)
}
