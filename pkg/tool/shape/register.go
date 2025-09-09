// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements self-registration for the shape toolset.
// filename: pkg/tool/shape/register.go
// nlines: 12
// risk_rating: LOW

package shape

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the shape package is imported.
func init() {
	tool.AddToolsetRegistration(
		group,
		tool.CreateRegistrationFunc(group, shapeToolsToRegister),
	)
}
