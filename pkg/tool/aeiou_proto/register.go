// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Implements self-registration for the aeiou toolset.
// filename: pkg/tool/aeiou_proto/register.go
// nlines: 12
// risk_rating: LOW

package aeiou_proto

import "github.com/aprice2704/neuroscript/pkg/tool"

func init() {
	tool.AddToolsetRegistration(
		"aeiou",
		tool.CreateRegistrationFunc("aeiou", aeiouToolsToRegister),
	)
}
