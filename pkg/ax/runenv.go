// NeuroScript Version: 0.8.0
// File version: 7
// Purpose: FEAT: Added ListTools and GetTool methods for host introspection, as required by tools like the LSP.
// filename: pkg/ax/runenv.go
// nlines: 48
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

	// ListTools returns all registered tool implementations.
	// The return type is []any to avoid a dependency on internal types;
	// consumers should type-assert to []api.ToolImplementation.
	ListTools() []any

	// GetTool returns a single tool implementation by its fully qualified name.
	// The return type is any to avoid a dependency on internal types;
	// consumers should type-assert to api.ToolImplementation.
	GetTool(name string) (any, bool)
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
