// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Corrected Capability struct instantiation to use Verbs and Scopes slices.
// filename: pkg/types/capability.go
// nlines: 25
// risk_rating: LOW

package types

import "github.com/aprice2704/neuroscript/pkg/capability"

var (
	// CapModelAdmin is the capability required to register, update, or delete AgentModels.
	CapModelAdmin = capability.Capability{
		Resource: capability.ResModel,
		Verbs:    []string{capability.VerbAdmin},
		Scopes:   []string{"*"},
	}
	// CapAccountAdmin is the capability required to manage accounts.
	CapAccountAdmin = capability.Capability{
		Resource: capability.ResAccount,
		Verbs:    []string{capability.VerbAdmin},
		Scopes:   []string{"*"},
	}
	// Add other common capabilities here as the system grows.
)
