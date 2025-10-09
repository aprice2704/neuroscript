// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: FIX: Updated to use the new Store.Registry(index) method to correctly retrieve the capsule registry.
// filename: pkg/api/ax_catalogs_impl.go
// nlines: 42
// risk_rating: LOW

package api

import (
	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/ax/contract"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// sharedCatalogs is the concrete implementation of the contract.SharedCatalogs interface.
type sharedCatalogs struct {
	root *Interpreter
}

// NewSharedCatalogs creates a new catalog facade pointing to the given root interpreter.
func NewSharedCatalogs(root *Interpreter) contract.SharedCatalogs {
	return &sharedCatalogs{root: root}
}

var _ contract.SharedCatalogs = (*sharedCatalogs)(nil)

func (s *sharedCatalogs) Accounts() interfaces.AccountReader {
	return s.root.internal.Accounts()
}

func (s *sharedCatalogs) AgentModels() interfaces.AgentModelReader {
	return s.root.internal.AgentModels()
}

func (s *sharedCatalogs) Tools() ax.Tools {
	return &axTools{itp: s.root}
}

func (s *sharedCatalogs) Capsules() *capsule.Registry {
	// The root store has multiple layers; layer 0 is the base system registry.
	// We use the newly added Registry() method to access it.
	return s.root.CapsuleStore().Registry(0)
}
