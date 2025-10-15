// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Registers the meta toolset with the global tool registry bootstrap list.
// filename: pkg/tool/meta/register.go
// nlines: 15
// risk_rating: LOW

package meta

import "github.com/aprice2704/neuroscript/pkg/tool"

// init() runs once when the meta package is imported, adding the
// registration function for this toolset to the global bootstrap list.
func init() {
	tool.AddToolsetRegistration("meta", RegisterTools)
}
