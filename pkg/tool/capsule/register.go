// NeuroScript Version: 0.7.2
// File version: 2
// Purpose: Registers the 'capsule' toolset with the NeuroScript engine.
// filename: pkg/tool/capsule/register.go
// nlines: 15
// risk_rating: LOW
package capsule

import "github.com/aprice2704/neuroscript/pkg/tool"

// init() runs once when the capsule package is imported. It injects this
// toolset's registration function into the global bootstrap list kept
// in the parent tool package.
func init() {
	tool.AddToolsetRegistration(
		"capsule",
		tool.CreateRegistrationFunc("capsule", CapsuleToolsToRegister),
	)
}
