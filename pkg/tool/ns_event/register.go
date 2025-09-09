// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Registers the 'ns_event' toolset with the NeuroScript engine.
// filename: pkg/tool/ns_event/register.go
// nlines: 15
// risk_rating: LOW
package ns_event

import "github.com/aprice2704/neuroscript/pkg/tool"

// init() runs once when the ns_event package is imported. It injects this
// toolset's registration function into the global bootstrap list kept
// in the parent tool package.
func init() {
	tool.AddToolsetRegistration(
		"ns_event",
		tool.CreateRegistrationFunc("ns_event", EventToolsToRegister),
	)
}
