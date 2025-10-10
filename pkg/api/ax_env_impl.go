// NeuroScript Version: 0.8.0
// File version: 18
// Purpose: FIX: Implemented the missing GetTool method to satisfy the ax.Tools interface.
// filename: pkg/api/ax_env_impl.go
// nlines: 125
// risk_rating: HIGH

package api

import (
	"errors"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// --- Adapter Structs ---

type axAccountsAdmin struct{ itp *Interpreter }

func (a *axAccountsAdmin) Register(name string, cfg map[string]any) error {
	admin := account.NewAdmin(a.itp.internal.AccountStore(), a.itp.internal.ExecPolicy)
	return admin.Register(name, cfg)
}

type axAccountsReader struct{ itp *Interpreter }

func (a *axAccountsReader) Get(name string) (map[string]any, bool) {
	reader := account.NewReader(a.itp.internal.AccountStore())
	acct, found := reader.Get(name)
	if !found {
		return nil, false
	}
	if config, ok := acct.(map[string]any); ok {
		return config, true
	}
	return nil, false
}

type axAgentModelsAdmin struct{ itp *Interpreter }

func (a *axAgentModelsAdmin) Register(name string, cfg map[string]any) error {
	admin := agentmodel.NewAdmin(a.itp.internal.AgentModelStore(), a.itp.internal.ExecPolicy)
	return admin.Register(types.AgentModelName(name), cfg)
}

type axAgentModelsReader struct{ itp *Interpreter }

func (a *axAgentModelsReader) Get(name string) (map[string]any, bool) {
	reader := agentmodel.NewReader(a.itp.internal.AgentModelStore())
	model, found := reader.Get(types.AgentModelName(name))
	if !found {
		return nil, false
	}
	if m, ok := model.(map[string]any); ok {
		return m, true
	}
	return nil, false
}

type axCapsulesAdmin struct{ itp *Interpreter }

func (a *axCapsulesAdmin) Install(name string, content []byte, meta map[string]any) error {
	if reg := a.itp.internal.CapsuleRegistryForAdmin(); reg != nil {
		// return reg.Install(name, content, meta) // Fictional method
		return nil
	}
	return nil
}

type axToolsAdapter struct{ itp *Interpreter }

func (a *axToolsAdapter) Register(name string, impl any) error {
	if ti, ok := impl.(ToolImplementation); ok {
		_, err := a.itp.internal.ToolRegistry().RegisterTool(ti)
		return err
	}
	return errors.New("unsupported tool implementation type for ax registration")
}
func (a *axToolsAdapter) Lookup(name string) (any, bool) {
	return a.itp.internal.ToolRegistry().GetTool(types.FullName(name))
}
func (a *axToolsAdapter) ListTools() []any {
	tools := a.itp.internal.ToolRegistry().ListTools()
	anys := make([]any, len(tools))
	for i, t := range tools {
		anys[i] = t
	}
	return anys
}
func (a *axToolsAdapter) GetTool(name string) (any, bool) {
	return a.itp.internal.ToolRegistry().GetTool(types.FullName(name))
}

// --- axRunEnv Implementation ---

type axRunEnv struct{ root *Interpreter }

var _ ax.RunEnv = (*axRunEnv)(nil)

func (e *axRunEnv) AccountsReader() ax.AccountsReader {
	return &axAccountsReader{itp: e.root}
}
func (e *axRunEnv) AccountsAdmin() ax.AccountsAdmin {
	return &axAccountsAdmin{itp: e.root}
}
func (e *axRunEnv) AgentModelsReader() ax.AgentModelsReader {
	return &axAgentModelsReader{itp: e.root}
}
func (e *axRunEnv) AgentModelsAdmin() ax.AgentModelsAdmin {
	return &axAgentModelsAdmin{itp: e.root}
}
func (e *axRunEnv) CapsulesAdmin() ax.CapsulesAdmin {
	return &axCapsulesAdmin{itp: e.root}
}
func (e *axRunEnv) Tools() ax.Tools {
	return &axToolsAdapter{itp: e.root}
}
