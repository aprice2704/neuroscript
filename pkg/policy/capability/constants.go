// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Defines standardized constants for capability resources and verbs.
// filename: pkg/policy/capability/constants.go
// nlines: 20
// risk_rating: LOW

package capability

// Standard capability resources.
const (
	ResFS     = "fs"
	ResNet    = "net"
	ResEnv    = "env"
	ResModel  = "model"
	ResTool   = "tool"
	ResSecret = "secret"
	ResBudget = "budget"
)

// Standard capability verbs.
const (
	VerbRead  = "read"
	VerbWrite = "write"
	VerbAdmin = "admin"
	VerbUse   = "use"
	VerbExec  = "exec"
)
