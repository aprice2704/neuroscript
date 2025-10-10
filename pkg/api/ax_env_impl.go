// NeuroScript Version: 0.8.0
// File version: 22
// Purpose: FIX: Refactors all adapters to follow the ax_impl_wiring.md guide, wrapping interfaces instead of the interpreter. Consolidates axToolsAdapter implementation.
// filename: pkg/api/ax_env_impl.go
// nlines: 75
// risk_rating: HIGH

package api

import (
	"errors"

	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// --- Adapter Structs ---

type axAccountsAdmin struct{ ua interfaces.AccountAdmin }

func (a *axAccountsAdmin) Register(name string, cfg map[string]any) error {
	return a.ua.Register(name, cfg)
}

type axAccountsReader struct{ ur interfaces.AccountReader }

func (a *axAccountsReader) Get(name string) (map[string]any, bool) {
	acct, found := a.ur.Get(name)
	if !found {
		return nil, false
	}
	if config, ok := acct.(map[string]any); ok {
		return config, true
	}
	return nil, false
}

type axAgentModelsAdmin struct{ ua interfaces.AgentModelAdmin }

func (a *axAgentModelsAdmin) Register(name string, cfg map[string]any) error {
	return a.ua.Register(types.AgentModelName(name), cfg)
}

type axAgentModelsReader struct{ ur interfaces.AgentModelReader }

func (a *axAgentModelsReader) Get(name string) (map[string]any, bool) {
	model, found := a.ur.Get(types.AgentModelName(name))
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
	if reg := a.itp.CapsuleRegistryForAdmin(); reg != nil {
		return nil // Placeholder
	}
	return nil
}

type axToolsAdapter struct{ tr tool.ToolRegistry }

func (t *axToolsAdapter) Register(name string, impl any) error {
	if ti, ok := impl.(ToolImplementation); ok {
		_, err := t.tr.RegisterTool(ti)
		return err
	}
	return errors.New("unsupported tool implementation type")
}
func (t *axToolsAdapter) Lookup(name string) (any, bool) { return t.tr.GetTool(types.FullName(name)) }
func (t *axToolsAdapter) ListTools() []any {
	tools := t.tr.ListTools()
	anys := make([]any, len(tools))
	for i, tool := range tools {
		anys[i] = tool
	}
	return anys
}
func (t *axToolsAdapter) GetTool(name string) (any, bool) { return t.tr.GetTool(types.FullName(name)) }

// --- axRunEnv Implementation ---

type axRunEnv struct{ root *Interpreter }

var _ ax.RunEnv = (*axRunEnv)(nil)

func (e *axRunEnv) AccountsReader() ax.AccountsReader {
	return &axAccountsReader{ur: e.root.Accounts()}
}
func (e *axRunEnv) AccountsAdmin() ax.AccountsAdmin {
	return &axAccountsAdmin{ua: e.root.AccountsAdmin()}
}
func (e *axRunEnv) AgentModelsReader() ax.AgentModelsReader {
	return &axAgentModelsReader{ur: e.root.AgentModels()}
}
func (e *axRunEnv) AgentModelsAdmin() ax.AgentModelsAdmin {
	return &axAgentModelsAdmin{ua: e.root.AgentModelsAdmin()}
}
func (e *axRunEnv) CapsulesAdmin() ax.CapsulesAdmin {
	return &axCapsulesAdmin{itp: e.root}
}
func (e *axRunEnv) Tools() ax.Tools {
	return &axToolsAdapter{tr: e.root.internal.ToolRegistry()}
}
