// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: FIX: Corrected the tools adapter struct name to axToolsAdapter and initialized it with the tool registry.
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
	// FIX: Use the correct struct name 'axToolsAdapter' and initialize it
	// with the tool registry from the interpreter.
	return &axToolsAdapter{tr: s.root.internal.ToolRegistry()}
}

func (s *sharedCatalogs) Capsules() *capsule.Store {
	// The interface requires the Store, not the Registry.
	return s.root.internal.CapsuleStore()
}
