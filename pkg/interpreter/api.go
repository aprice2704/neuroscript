// NeuroScript Version: 0.8.0
// File version: 12
// Purpose: Patched Run method to correctly call un-exported runProcedure.
// filename: pkg/interpreter/api.go
// nlines: 120
// risk_rating: MEDIUM

package interpreter

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Run is the public entrypoint for executing a named procedure.
func (i *Interpreter) Run(procName string, args ...lang.Value) (lang.Value, error) {
	// FIX: The internal method is 'runProcedure' and 'args' must be expanded.
	return i.runProcedure(procName, args...)
}

// ExecuteCommands is the public entrypoint for executing all top-level command blocks.
func (i *Interpreter) ExecuteCommands() (lang.Value, error) {
	return i.executeCommands()
}

// HasEmitFunc returns true if a custom emit handler is configured in the HostContext.
func (i *Interpreter) HasEmitFunc() bool {
	return i.hostContext != nil && i.hostContext.EmitFunc != nil
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
	i.rootInterpreter().state.providersMu.Lock()
	defer i.rootInterpreter().state.providersMu.Unlock()
	if i.rootInterpreter().state.providers == nil {
		i.rootInterpreter().state.providers = make(map[string]provider.AIProvider)
	}
	i.rootInterpreter().state.providers[name] = p
}

// GetProvider retrieves a registered AIProvider by name.
func (i *Interpreter) GetProvider(name string) (provider.AIProvider, bool) {
	i.rootInterpreter().state.providersMu.RLock()
	defer i.rootInterpreter().state.providersMu.RUnlock()
	p, found := i.rootInterpreter().state.providers[name]
	return p, found
}

// NTools returns the number of registered tools.
func (i *Interpreter) NTools() (ntools int) {
	return i.tools.NTools()
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

// Logger returns the interpreter's configured logger from the HostContext.
func (i *Interpreter) Logger() interfaces.Logger {
	if i.hostContext == nil || i.hostContext.Logger == nil {
		panic("FATAL: Interpreter has no logger configured in its HostContext.")
	}
	return i.hostContext.Logger
}

// GetLogger satisfies the tool.Runtime interface by wrapping the Logger method.
func (i *Interpreter) GetLogger() interfaces.Logger {
	return i.Logger()
}

// LLM satisfies the tool.Runtime interface.
func (i *Interpreter) LLM() interfaces.LLMClient {
	// The LLM client is a root-level resource.
	return i.rootInterpreter().aiWorker
}

// GetAndClearWhisperBuffer retrieves the content of the default 'self' buffer and clears it.
func (i *Interpreter) GetAndClearWhisperBuffer() string {
	return i.bufferManager.GetAndClear(DefaultSelfHandle)
}

// GetTurnContext returns the context for the current AEIOU turn.
func (i *Interpreter) GetTurnContext() context.Context {
	if i.turnCtx == nil {
		return context.Background()
	}
	return i.turnCtx
}

// setTurnContext sets the context for the current turn.
func (i *Interpreter) setTurnContext(ctx context.Context) {
	i.turnCtx = ctx
}

// GetToolSpec satisfies the eval.Runtime interface.
func (i *Interpreter) GetToolSpec(toolName types.FullName) (tool.ToolSpec, bool) {
	t, ok := i.tools.GetTool(toolName)
	if !ok {
		return tool.ToolSpec{}, false
	}
	return t.Spec, true
}
