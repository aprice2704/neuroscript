// filename: pkg/tool/fs/register.go
package fs

import "github.com/aprice2704/neuroscript/pkg/tool"

// init() runs once when the fs package is imported. It injects this
// toolset's registration function into the global bootstrap list kept
// in the parent tool package.
func init() {
	tool.AddToolsetRegistration(
		"fs",
		tool.CreateRegistrationFunc("fs", FsToolsToRegister),
	)
}
