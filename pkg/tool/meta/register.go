// filename: pkg/tool/meta/register.go
package meta

import "github.com/aprice2704/neuroscript/pkg/tool"

// init() runs once when the meta package is imported. It injects this
// toolset's registration function into the global bootstrap list kept
// in the parent tool package.
func init() {
	tool.AddToolsetRegistration(
		"meta",
		tool.CreateRegistrationFunc("meta", metaToolsToRegister),
	)
}
