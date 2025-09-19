// NeuroScript Version: 0.7.2
// File version: 1
// Purpose: Registers the 'metadata' toolset with the NeuroScript engine.
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
		tool.CreateRegistrationFunc("metadata", MetadataToolsToRegister),
	)
}
