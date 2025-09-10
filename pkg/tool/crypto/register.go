// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Implements self-registration for the crypto toolset, including hashing tools.
// filename: pkg/tool/crypto/register.go
// nlines: 16
// risk_rating: LOW

package crypto

import "github.com/aprice2704/neuroscript/pkg/tool"

// init runs once when the crypto package is imported. It injects this
// toolset's registration function into the global bootstrap list.
func init() {
	allTools := append(cryptoToolsToRegister, cryptoHashToolsToRegister...)
	tool.AddToolsetRegistration(
		"crypto",
		tool.CreateRegistrationFunc("crypto", allTools),
	)
}
