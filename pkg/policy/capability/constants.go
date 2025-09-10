// NeuroScript Version: 0.3.0
// File version: 3
// Purpose: Defines standardized constants for capability resources and verbs, adding ResCrypto and VerbSign.
// filename: pkg/policy/capability/constants.go
// nlines: 24
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
	ResBus    = "bus"
	ResCrypto = "crypto"
)

// Standard capability verbs.
const (
	VerbRead  = "read"
	VerbWrite = "write"
	VerbAdmin = "admin"
	VerbUse   = "use"
	VerbExec  = "exec"
	VerbSign  = "sign"
)
