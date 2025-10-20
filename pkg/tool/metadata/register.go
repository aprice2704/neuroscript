// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Updated to pass the local RegisterTools function instead of using CreateRegistrationFunc.
// filename: pkg/tool/metadata/register.go
// nlines: 15
// risk_rating: LOW
package metadata

import "github.com/aprice2704/neuroscript/pkg/tool"

// init() runs once when the package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	tool.AddToolsetRegistration(
		"metadata",
		RegisterTools, // FIX: Pass the concrete RegisterTools function
	)
}
