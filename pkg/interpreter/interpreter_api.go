// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Added GetTurnContext and setTurnContext methods to manage the AEIOU v3 session context.
// filename: pkg/interpreter/interpreter_api.go
// nlines: 135
// risk_rating: LOW

package interpreter

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// AgentModels provides read-only access to the agent model store.
func (i *Interpreter) AgentModels() interfaces.AgentModelReader {
	return agentmodel.NewAgentModelReader(i.modelStore)
}

// AgentModelsAdmin provides administrative access to the agent model store.
func (i *Interpreter) AgentModelsAdmin() interfaces.AgentModelAdmin {
	return agentmodel.NewAgentModelAdmin(i.modelStore, i.ExecPolicy)
}

// PromptUser satisfies the tool.Runtime interface for user interaction.
func (i *Interpreter) PromptUser(prompt string) (string, error) {
	if _, err := fmt.Fprint(i.Stdout(), prompt+" "); err != nil {
		return "", fmt.Errorf("failed to write prompt to stdout: %w", err)
	}
	reader := bufio.NewReader(i.Stdin())
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read from stdin: %w", err)
	}
	return strings.TrimSpace(response), nil
}

// RegisterProvider allows the host application to register a concrete AIProvider implementation.
func (i *Interpreter) RegisterProvider(name string, p provider.AIProvider) {
	if i.root != nil {
		i.root.RegisterProvider(name, p)
		return
	}
	i.state.providersMu.Lock()
	defer i.state.providersMu.Unlock()
	if i.state.providers == nil {
		i.state.providers = make(map[string]provider.AIProvider)
	}
	i.state.providers[name] = p
}

// GetProvider retrieves a registered AIProvider by name.
func (i *Interpreter) GetProvider(name string) (provider.AIProvider, bool) {
	if i.root != nil {
		return i.root.GetProvider(name)
	}
	i.state.providersMu.RLock()
	defer i.state.providersMu.RUnlock()
	p, found := i.state.providers[name]
	return p, found
}

// NTools returns the number of registered tools.
func (i *Interpreter) NTools() (ntools int) {
	return i.tools.NTools()
}

// LLM returns the configured LLM client.
func (i *Interpreter) LLM() interfaces.LLMClient {
	return i.llmclient
}

// KnownProcedures returns the map of known procedures.
func (i *Interpreter) KnownProcedures() map[string]*ast.Procedure {
	if i.state.knownProcedures == nil {
		return make(map[string]*ast.Procedure)
	}
	return i.state.knownProcedures
}

// ToolRegistry returns the interpreter's tool registry.
func (i *Interpreter) ToolRegistry() tool.ToolRegistry {
	return i.tools
}

// CloneForEventHandler creates a sandboxed clone for event handling.
func (i *Interpreter) CloneForEventHandler() *Interpreter {
	return i.clone()
}

// CloneWithNewVariables creates a clone with a fresh set of variables for procedure calls.
func (i *Interpreter) CloneWithNewVariables() *Interpreter {
	return i.clone()
}

// GetLogger returns the interpreter's configured logger.
func (i *Interpreter) GetLogger() interfaces.Logger {
	return i.logger
}

// SetLastResult sets the value of the 'last' keyword.
func (i *Interpreter) SetLastResult(v lang.Value) {
	i.lastCallResult = v
}

// RegisterEvent registers an event handler.
func (i *Interpreter) RegisterEvent(decl *ast.OnEventDecl) error {
	return i.eventManager.register(decl, i)
}

// SetEmitFunc sets a custom function to handle 'emit' statements.
func (i *Interpreter) SetEmitFunc(f func(lang.Value)) {
	i.customEmitFunc = f
}

// SetWhisperFunc sets a custom function to handle 'whisper' statements.
func (i *Interpreter) SetWhisperFunc(f func(handle, data lang.Value)) {
	i.customWhisperFunc = f
}

// GetAndClearWhisperBuffer retrieves the content of the default 'self' buffer and clears it.
func (i *Interpreter) GetAndClearWhisperBuffer() string {
	return i.bufferManager.GetAndClear(DefaultSelfHandle)
}

// GetTurnContext returns the context for the current AEIOU turn. This is intended
// to satisfy the tool.Runtime interface so tools can access session data.
func (i *Interpreter) GetTurnContext() context.Context {
	if i.turnCtx == nil {
		return context.Background()
	}
	return i.turnCtx
}

// setTurnContext sets the context for the current turn. This is used internally
// by the host loop controller in 'ask'.
func (i *Interpreter) setTurnContext(ctx context.Context) {
	i.turnCtx = ctx
}
