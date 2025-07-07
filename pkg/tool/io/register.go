// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements self-registration for the io toolset.
// filename: pkg/tool/io/register.go
package io

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the io package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	tool.AddToolsetRegistration(
		"io",
		tool.CreateRegistrationFunc("io", ioToolsToRegister),
	)
}
