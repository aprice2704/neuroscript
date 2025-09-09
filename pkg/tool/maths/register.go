// filename: pkg/tool/maths/register.go
package maths

import "github.com/aprice2704/neuroscript/pkg/tool"

// init() runs once when the math package is imported (e.g. by the CLI bundle
// or by tests).  It injects this tool-set’s registration function into the
// global bootstrap list kept in pkg/tool/register.go.
//
// At interpreter start-up, tool.RegisterGlobalToolsets() will call that
// registration func, which, in turn, adds every ToolImplementation in
// mathToolsToRegister to the live registry.
func init() {
	tool.AddToolsetRegistration(
		"math",
		tool.CreateRegistrationFunc("math", mathToolsToRegister),
	)
}
