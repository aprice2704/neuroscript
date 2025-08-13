// NeuroScript Version: 0.6.0
// File version: 1.0.0
// Purpose: Defines the interfaces for interacting with the AgentModelStore.
// filename: pkg/interfaces/agentmodel.go
// nlines: 20
// risk_rating: LOW

package interfaces

import "github.com/aprice2704/neuroscript/pkg/types"

// AgentModelReader provides read-only access to registered AgentModels.
type AgentModelReader interface {
	List() []types.AgentModelName
	Get(name types.AgentModelName) (any, bool)
}

// AgentModelAdmin provides administrative (read/write) access to AgentModels.
type AgentModelAdmin interface {
	AgentModelReader
	Register(name types.AgentModelName, cfg map[string]any) error
	Update(name types.AgentModelName, updates map[string]any) error
	Delete(name types.AgentModelName) bool
}
