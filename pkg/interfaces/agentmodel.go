// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Purified interface of pkg/types imports to break import cycles.
// filename: pkg/interfaces/agentmodel.go
// nlines: 21
// risk_rating: LOW

package interfaces

// AgentModelReader provides read-only access to registered AgentModels.
type AgentModelReader interface {
	List() []string
	Get(name string) (any, bool)
}

// AgentModelAdmin provides administrative (read/write) access to AgentModels.
type AgentModelAdmin interface {
	AgentModelReader
	Register(name string, cfg map[string]any) error
	RegisterFromModel(model any) error // New method for host injection
	Update(name string, updates map[string]any) error
	Delete(name string) bool
}
