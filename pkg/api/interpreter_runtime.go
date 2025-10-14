// NeuroScript Version: 0.8.0
// File version: 5
// Purpose: Implements the tool.Runtime interface for the public api.Interpreter.
// filename: pkg/api/interpreter_runtime.go
// nlines: 135
// risk_rating: MEDIUM

package api

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Statically assert that *Interpreter satisfies the tool.Runtime interface.
// This will cause a compile-time error if the interface is not fully implemented.
var _ tool.Runtime = (*Interpreter)(nil)

// This file contains the methods that make *api.Interpreter satisfy the
// tool.Runtime interface. This allows the public Interpreter to be passed
// directly to tools that need access to the execution environment.

// Println prints to the interpreter's configured Stdout.
func (i *Interpreter) Println(a ...any) {
	if i.internal == nil || i.internal.HostContext() == nil || i.internal.HostContext().Stdout == nil {
		return
	}
	fmt.Fprintln(i.internal.HostContext().Stdout, a...)
}

// PromptUser is not yet supported in the public API. Tools should not rely on it.
func (i *Interpreter) PromptUser(prompt string) (string, error) {
	return "", errors.New("interactive user prompts are not supported via this runtime")
}

// GetVar retrieves a variable from the interpreter's global scope and unwraps it to a native Go type.
func (i *Interpreter) GetVar(name string) (any, bool) {
	val, ok := i.internal.GetVariable(name)
	if !ok {
		return nil, false
	}
	return lang.Unwrap(val), true
}

// SetVar wraps a native Go value and sets it in the interpreter's global scope.
// If the value cannot be wrapped, an error is logged.
func (i *Interpreter) SetVar(name string, val any) {
	wrappedVal, err := lang.Wrap(val)
	if err != nil {
		if logger := i.GetLogger(); logger != nil {
			logger.Errorf("Failed to set variable '%s': %v", name, err)
		}
		return
	}
	i.internal.SetVariable(name, wrappedVal)
}

// CallTool executes a tool by its full name with native Go arguments.
func (i *Interpreter) CallTool(name types.FullName, args []any) (any, error) {
	wrappedArgs := make([]lang.Value, len(args))
	for idx, arg := range args {
		wrapped, err := lang.Wrap(arg)
		if err != nil {
			return nil, fmt.Errorf("error wrapping argument %d for tool '%s': %w", idx, name, err)
		}
		wrappedArgs[idx] = wrapped
	}

	result, err := i.ToolRegistry().CallFromInterpreter(i, name, wrappedArgs)
	if err != nil {
		return nil, err
	}
	return lang.Unwrap(result), nil
}

// GetLogger returns the interpreter's configured logger from its HostContext.
func (i *Interpreter) GetLogger() interfaces.Logger {
	if i.internal == nil || i.internal.HostContext() == nil {
		return nil
	}
	return i.internal.HostContext().Logger
}

// SandboxDir returns the root directory for sandboxed file operations.
func (i *Interpreter) SandboxDir() string {
	if i.internal == nil {
		return ""
	}
	return i.internal.SandboxDir()
}

// LLM returns the configured LLM client.
func (i *Interpreter) LLM() interfaces.LLMClient {
	if i.internal == nil {
		return nil
	}
	return i.internal.LLM()
}

// RegisterHandle stores a Go object and returns a handle string for it.
func (i *Interpreter) RegisterHandle(obj interface{}, typePrefix string) (string, error) {
	if i.internal == nil {
		return "", errors.New("interpreter not initialized")
	}
	return i.Handles().RegisterHandle(obj, typePrefix)
}

// GetHandleValue retrieves a Go object by its handle.
func (i *Interpreter) GetHandleValue(handle string, expectedTypePrefix string) (interface{}, error) {
	if i.internal == nil {
		return nil, errors.New("interpreter not initialized")
	}
	return i.Handles().GetHandleValue(handle, expectedTypePrefix)
}

// AgentModels returns the read-only view of the AgentModel store.
func (i *Interpreter) AgentModels() interfaces.AgentModelReader {
	if i.internal == nil {
		return nil
	}
	return i.internal.AgentModels()
}

// AgentModelsAdmin returns the administrative view of the AgentModel store.
func (i *Interpreter) AgentModelsAdmin() interfaces.AgentModelAdmin {
	if i.internal == nil {
		return nil
	}
	return i.internal.AgentModelsAdmin()
}

// GetGrantSet returns the capability grants from the current execution policy.
func (i *Interpreter) GetGrantSet() *capability.GrantSet {
	if i.internal == nil || i.internal.ExecPolicy == nil {
		return &capability.GrantSet{}
	}
	return &i.internal.ExecPolicy.Grants
}

// GetExecPolicy returns the current execution policy.
func (i *Interpreter) GetExecPolicy() *policy.ExecPolicy {
	if i.internal == nil {
		return nil
	}
	return i.internal.ExecPolicy
}
