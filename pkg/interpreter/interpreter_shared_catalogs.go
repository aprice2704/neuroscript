// NeuroScript Version: 0.8.0
// File version: 9
// Purpose: FIX: Corrects the axToolsAdapter to use the right method names (RegisterTool, GetTool) and the concrete *tool.ToolRegistryImpl type.
// filename: pkg/interpreter/interpreter_shared_catalogs.go
// nlines: 87
// risk_rating: MEDIUM

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/ax/contract"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// --- Adapter to satisfy the ax.Tools interface ---
type axToolsAdapter struct{ reg *tool.ToolRegistryImpl }

func (a *axToolsAdapter) Register(name string, impl any) error {
	toolImpl, ok := impl.(tool.ToolImplementation)
	if !ok {
		return fmt.Errorf("axToolsAdapter: expected tool.ToolImplementation, got %T", impl)
	}
	_, err := a.reg.RegisterTool(toolImpl)
	return err
}

func (a *axToolsAdapter) Lookup(name string) (any, bool) {
	return a.reg.GetTool(types.FullName(name))
}

func (a *axToolsAdapter) ListTools() []any {
	tools := a.reg.ListTools()
	anys := make([]any, len(tools))
	for i, t := range tools {
		anys[i] = t
	}
	return anys
}
func (a *axToolsAdapter) GetTool(name string) (any, bool) {
	return a.reg.GetTool(types.FullName(name))
}

// sharedCatalogs is the concrete implementation of the SharedCatalogs interface.
type sharedCatalogs struct {
	accounts    *account.Store
	agentModels *agentmodel.AgentModelStore
	tools       ax.Tools
	capsules    *capsule.Store
}

// Compile-time check to ensure sharedCatalogs satisfies the interface.
var _ contract.SharedCatalogs = (*sharedCatalogs)(nil)

func (sc *sharedCatalogs) Accounts() interfaces.AccountReader {
	return account.NewReader(sc.accounts)
}
func (sc *sharedCatalogs) AgentModels() interfaces.AgentModelReader {
	return agentmodel.NewReader(sc.agentModels)
}
func (sc *sharedCatalogs) Tools() ax.Tools          { return sc.tools }
func (sc *sharedCatalogs) Capsules() *capsule.Store { return sc.capsules }

// newSharedCatalogs creates and initializes the shared catalogs for a new root interpreter.
func newSharedCatalogs(i *Interpreter, isRoot bool, skipStdTools bool) contract.SharedCatalogs {
	if !isRoot {
		panic("newSharedCatalogs should only be called for a root interpreter")
	}

	internalToolRegistry := tool.NewToolRegistry(i)

	sc := &sharedCatalogs{
		accounts:    account.NewStore(),
		agentModels: agentmodel.NewAgentModelStore(),
		tools:       &axToolsAdapter{reg: internalToolRegistry},
		capsules:    capsule.NewStore(capsule.DefaultRegistry()),
	}

	if !skipStdTools {
		if err := tool.RegisterGlobalToolsets(internalToolRegistry); err != nil {
			panic(fmt.Sprintf("FATAL: Failed to register global toolsets: %v", err))
		}
	}

	if err := registerDebugTools(internalToolRegistry); err != nil {
		panic(fmt.Sprintf("FATAL: Failed to register debug tools: %v", err))
	}

	return sc
}
