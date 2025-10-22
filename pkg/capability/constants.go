// NeuroScript Version: 0.3.0
// File version: 4 // Bumped version
// Purpose: Defines standardized constants for capability resources and verbs, adding CapabilityAllowAll.
// filename: pkg/policy/capability/constants.go
// nlines: 28 // Adjusted line count
// risk_rating: LOW

package capability

// Standard capability resources.
const (
	ResFS      = "fs"
	ResNet     = "net"
	ResEnv     = "env"
	ResModel   = "model"
	ResTool    = "tool"
	ResSecret  = "secret"
	ResBudget  = "budget"
	ResBus     = "bus"
	ResCrypto  = "crypto"
	ResAccount = "account"
	ResCapsule = "capsule"
	ResIPC     = "ipc"
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

// Common Capability Grant Patterns
const (
	// Capability.AllowAll grants all permissions. Use with extreme caution.
	AllowAll = "*:*:*"
)
