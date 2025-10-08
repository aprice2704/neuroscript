// NeuroScript Version: 0.7.4
// File version: 6
// Purpose: Reverted to a stdlib-only package by replacing internal types with pure Go interfaces for readers and admins.
// filename: pkg/ax/runenv.go
// nlines: 39
// risk_rating: LOW

package ax

// Read-only views
type AccountsReader interface {
	Get(name string) (map[string]any, bool)
}
type AgentModelsReader interface {
	Get(name string) (map[string]any, bool)
}

// Admin surfaces (write)
type AccountsAdmin interface {
	Register(name string, cfg map[string]any) error
}
type AgentModelsAdmin interface {
	Register(name string, cfg map[string]any) error
}
type CapsulesAdmin interface {
	Install(name string, content []byte, meta map[string]any) error
}

// Tool registry (host-provided)
type Tools interface {
	Register(name string, impl any) error
	Lookup(name string) (any, bool)
}

// Shared environment bound to a factory
type RunEnv interface {
	AccountsReader() AccountsReader
	AccountsAdmin() AccountsAdmin
	AgentModelsReader() AgentModelsReader
	AgentModelsAdmin() AgentModelsAdmin
	CapsulesAdmin() CapsulesAdmin
	Tools() Tools
}

type EnvCap interface {
	Env() RunEnv
}
