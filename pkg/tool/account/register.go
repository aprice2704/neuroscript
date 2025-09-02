// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Registers the 'account' toolset with the NeuroScript engine.
// filename: pkg/tool/account/register.go
// nlines: 15
// risk_rating: LOW
package account

import "github.com/aprice2704/neuroscript/pkg/tool"

// init() runs once when the account package is imported. It injects this
// toolset's registration function into the global bootstrap list kept
// in the parent tool package.
func init() {
	tool.AddToolsetRegistration(
		"account",
		tool.CreateRegistrationFunc("account", AccountToolsToRegister),
	)
}
