// NeuroScript Version: 0.8.0
// File version: 17
// Purpose: FEAT: Adds public setters for stores and other host-configurable components as requested by the API team.
// filename: pkg/interpreter/interpreter_api.go
// nlines: 205
// risk_rating: HIGH

package interpreter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/ax/contract"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// SetRuntime allows the host application to set a custom runtime context.
func (i *Interpreter) SetRuntime(rt tool.Runtime) {
	if rt == nil {
		i.runtime = i
	} else {
		i.runtime = rt
	}
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
	root := i.rootInterpreter()
	root.state.providersMu.Lock()
	defer root.state.providersMu.Unlock()
	if root.state.providers == nil {
		root.state.providers = make(map[string]provider.AIProvider)
	}
	root.state.providers[name] = p
}

// GetProvider retrieves a registered AIProvider by name.
func (i *Interpreter) GetProvider(name string) (provider.AIProvider, bool) {
	root := i.rootInterpreter()
	root.state.providersMu.RLock()
	defer root.state.providersMu.RUnlock()
	p, found := root.state.providers[name]
	return p, found
}

// NTools returns the number of registered tools.
func (i *Interpreter) NTools() (ntools int) {
	if tr, ok := i.catalogs.Tools().(tool.ToolRegistry); ok {
		return tr.NTools()
	}
	return 0
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

// CloneForEventHandler creates a sandboxed clone for event handling.
func (i *Interpreter) CloneForEventHandler() *Interpreter {
	return i.clone()
}

// CloneWithNewVariables creates a clone with a fresh set of variables for procedure calls.
func (i *Interpreter) CloneWithNewVariables() *Interpreter {
	return i.clone()
}

// GetLogger returns the interpreter's configured logger from the parcel.
func (i *Interpreter) GetLogger() interfaces.Logger {
	if i.parcel != nil && i.parcel.Logger() != nil {
		return i.parcel.Logger()
	}
	return logging.NewNoOpLogger() // Fallback
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

// HasEmitFunc returns true if a custom emit function has been set.
func (i *Interpreter) HasEmitFunc() bool {
	return i.customEmitFunc != nil
}

// SetWhisperFunc sets a custom function to handle 'whisper' statements.
func (i *Interpreter) SetWhisperFunc(f func(handle, data lang.Value)) {
	i.customWhisperFunc = f
}

// SetEventHandlerErrorCallback sets the callback for event handler errors.
func (i *Interpreter) SetEventHandlerErrorCallback(f func(eventName, source string, err *lang.RuntimeError)) {
	i.eventHandlerErrorCallback = f
}

// GetAndClearWhisperBuffer retrieves the content of the default 'self' buffer and clears it.
func (i *Interpreter) GetAndClearWhisperBuffer() string {
	return i.bufferManager.GetAndClear(DefaultSelfHandle)
}

// GetTurnContext reconstructs a context.Context from the parcel's HostContext.
func (i *Interpreter) GetTurnContext() context.Context {
	if i.parcel == nil {
		fmt.Fprintf(os.Stderr, "[DEBUG GetTurnContext] Interp ID: %s, Parcel is NIL\n", i.id)
		return context.Background()
	}

	hostCtx := i.parcel.AEIOU()
	ctx := context.Background()
	ctx = context.WithValue(ctx, aeiou.SessionIDKey, hostCtx.SessionID)
	ctx = context.WithValue(ctx, aeiou.TurnIndexKey, hostCtx.TurnIndex)
	ctx = context.WithValue(ctx, aeiou.TurnNonceKey, hostCtx.TurnNonce)

	fmt.Fprintf(os.Stderr, "[DEBUG GetTurnContext] Interp ID: %s, SID: %q, Turn: %d\n", i.id, hostCtx.SessionID, hostCtx.TurnIndex)
	return ctx
}

// SetTurnContext extracts data from a context.Context and stores it in the parcel's HostContext.
func (i *Interpreter) SetTurnContext(ctx context.Context) {
	if i.parcel == nil {
		fmt.Fprintf(os.Stderr, "[DEBUG SetTurnContext] Interp ID: %s, cannot set context on NIL parcel\n", i.id)
		return
	}

	sid, _ := ctx.Value(aeiou.SessionIDKey).(string)
	turn, _ := ctx.Value(aeiou.TurnIndexKey).(int)
	nonce, _ := ctx.Value(aeiou.TurnNonceKey).(string)

	fmt.Fprintf(os.Stderr, "[DEBUG SetTurnContext] Interp ID: %s, Setting context with SID: %q, Turn: %d\n", i.id, sid, turn)

	i.parcel = i.parcel.Fork(func(m *contract.ParcelMut) {
		m.AEIOU = &aeiou.HostContext{
			SessionID: sid,
			TurnIndex: turn,
			TurnNonce: nonce,
		}
	})
}

// SetCapsuleAdminRegistry sets the primary (admin) capsule registry.
// This is intended for setup and testing, and replaces the existing capsule store.
func (i *Interpreter) SetCapsuleAdminRegistry(registry *capsule.Registry) {
	root := i.rootInterpreter()
	if sc, ok := root.catalogs.(*sharedCatalogs); ok {
		sc.capsules = capsule.NewStore(registry)
	}
}

// SetAccountStore sets the account store for the interpreter.
// This is intended for setup and testing.
func (i *Interpreter) SetAccountStore(store *account.Store) {
	root := i.rootInterpreter()
	if sc, ok := root.catalogs.(*sharedCatalogs); ok {
		sc.accounts = store
	}
}

// SetAgentModelStore sets the agent model store for the interpreter.
// This is intended for setup and testing.
func (i *Interpreter) SetAgentModelStore(store *agentmodel.AgentModelStore) {
	root := i.rootInterpreter()
	if sc, ok := root.catalogs.(*sharedCatalogs); ok {
		sc.agentModels = store
	}
}

// SetAITranscript sets the writer for logging AI prompts and responses.
func (i *Interpreter) SetAITranscript(w io.Writer) {
	i.aiTranscript = w
}

// SetEmitter sets the LLM telemetry emitter for the interpreter.
func (i *Interpreter) SetEmitter(e interfaces.Emitter) {
	i.emitter = e
}
