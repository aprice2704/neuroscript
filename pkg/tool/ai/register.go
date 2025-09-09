// NeuroScript Version: 0.3.0
// File version: 2
// Purpose: Registers the AI toolset. Updated to use init-based registration.
// filename: pkg/tool/ai/register.go
// nlines: 14
// risk_rating: LOW

package ai

import "github.com/aprice2704/neuroscript/pkg/tool"

// init() runs once when the ai package is imported. It injects this tool-set’s
// registration function into the global bootstrap list.
//
// At interpreter start-up, tool.RegisterGlobalToolsets() will call that
// registration func, which, in turn, adds every ToolImplementation in
// aiToolsToRegister to the live registry.
func init() {
	tool.AddToolsetRegistration(
		"ai",
		tool.CreateRegistrationFunc("ai", aiToolsToRegister),
	)
}
