// NeuroScript Version: 0.5.2
// File version: 6
// Purpose: Corrected the CallTool signature to perfectly match the tool.Runtime interface, resolving the final compiler error.
// filename: pkg/interpreter/interpreter_tools.go
// nlines: 120
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// ToolHandler defines the interface for an external system that can execute tools.
type ToolHandler interface {
	CallTool(toolName string, methodName string, args map[string]any) (any, error)
}

// SetExternalToolHandler registers a handler for external tool calls.
func (i *Interpreter) SetExternalToolHandler(handler ToolHandler) {
	i.externalHandler = handler
}

// CallTool satisfies the tool.Runtime interface.
// FIX: The signature now correctly matches the interface: (string, []any) (any, error)
func (i *Interpreter) CallTool(toolName string, args []any) (any, error) {
	impl, found := i.GetTool(toolName)
	if !found {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", toolName), lang.ErrToolNotFound)
	}

	// Convert the []any slice to the map[string]lang.Value expected by executeInternalTool.
	// This logic assumes positional arguments based on the tool's spec.
	argsMap := make(map[string]lang.Value)
	for idx, argSpec := range impl.Spec.Args {
		if idx < len(args) {
			wrappedArg, err := lang.Wrap(args[idx])
			if err != nil {
				return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("failed to wrap argument for tool '%s'", toolName), err)
			}
			argsMap[argSpec.Name] = wrappedArg
		}
	}

	resultVal, err := i.executeInternalTool(impl, argsMap)
	if err != nil {
		return nil, err
	}

	// Unwrap the result to 'any' to match the interface signature.
	return lang.Unwrap(resultVal), nil
}

// executeInternalTool is the Generic Adapter Bridge for all internal tools.
func (i *Interpreter) executeInternalTool(impl tool.ToolImplementation, args map[string]lang.Value) (lang.Value, error) {
	// UNWRAP arguments from Value -> interface{}
	validatedArgs := make([]interface{}, len(impl.Spec.Args))
	for idx, argSpec := range impl.Spec.Args {
		value, provided := args[argSpec.Name]
		if !provided {
			if argSpec.Required {
				return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("tool '%s': missing required argument '%s'", impl.Spec.Name, argSpec.Name), lang.ErrArgumentMismatch)
			}
			validatedArgs[idx] = nil // Use nil for optional, unprovided args
		} else {
			unwrappedValue := lang.Unwrap(value)
			validatedArgs[idx] = unwrappedValue
		}
	}

	// CALL the primitive-based Go function
	result, err := impl.Func(i, validatedArgs)
	if err != nil {
		if _, ok := err.(*lang.RuntimeError); !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("tool '%s' execution failed: %v", impl.Spec.Name, err), err)
		}
		return nil, err
	}

	// WRAP the result from interface{} -> Value
	return lang.Wrap(result)
}

// ExecuteTool is the primary entry point for the interpreter's 'call' statement.
func (i *Interpreter) ExecuteTool(toolName string, args map[string]lang.Value) (lang.Value, error) {
	now := time.Now()
	if i.rateLimitCount > 0 && i.rateLimitDuration > 0 {
		timestamps := i.ToolCallTimestamps[toolName]
		validTimestamps := []time.Time{}
		cutoff := now.Add(-i.rateLimitDuration)
		for _, ts := range timestamps {
			if ts.After(cutoff) {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		if len(validTimestamps) >= i.rateLimitCount {
			return nil, lang.NewRuntimeError(lang.ErrorCodeRateLimited, fmt.Sprintf("tool '%s' rate limit exceeded", toolName), lang.ErrRateLimited)
		}
		validTimestamps = append(validTimestamps, now)
		i.ToolCallTimestamps[toolName] = validTimestamps
	}

	impl, found := i.GetTool(toolName)
	if found {
		return i.executeInternalTool(impl, args)
	}

	if i.externalHandler != nil {
		unwrappedArgs := make(map[string]any, len(args))
		for k, v := range args {
			unwrappedArgs[k] = lang.Unwrap(v)
		}
		methodName, ok := unwrappedArgs["method"].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("external tool call to '%s' requires a 'method' argument", toolName), lang.ErrArgumentMismatch)
		}
		result, err := i.externalHandler.(ToolHandler).CallTool(toolName, methodName, unwrappedArgs)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("external tool '%s' failed: %v", toolName, err), err)
		}
		return lang.Wrap(result)
	}

	return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", toolName), lang.ErrToolNotFound)
}

func (i *Interpreter) ToolRegistry() tool.ToolRegistry {
	return i
}

func (i *Interpreter) RegisterTool(impl tool.ToolImplementation) error {
	if i.tools == nil {
		return errors.New("internal error: tool registry not initialized")
	}
	return i.tools.RegisterTool(impl)
}

func (i *Interpreter) GetTool(name string) (tool.ToolImplementation, bool) {
	if i.tools == nil {
		return tool.ToolImplementation{}, false
	}
	return i.tools.GetTool(name)
}

func (i *Interpreter) ListTools() []tool.ToolSpec {
	if i.tools == nil {
		return []tool.ToolSpec{}
	}
	return i.tools.ListTools()
}

func (i *Interpreter) SetInternalToolRegistry(registry tool.ToolRegistry) {
	i.tools = registry
}

func (i *Interpreter) InternalToolRegistry() tool.ToolRegistry {
	if i.tools == nil {
		panic("FATAL: Interpreter's internal toolRegistry field is nil")
	}
	return i.tools
}
