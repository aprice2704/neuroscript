// NeuroScript Version: 0.3.1
// File version: 2
// Purpose: Methods for tool registration and execution in the interpreter.
// filename: pkg/core/interpreter_tools.go
// nlines: 110
// risk_rating: MEDIUM

package core

import (
	"errors"
	"fmt"
	"time"
)

type ToolHandler interface {
	CallTool(toolName string, methodName string, args map[string]any) (any, error)
}

func (i *Interpreter) SetExternalToolHandler(handler ToolHandler) {
	i.externalHandler = handler
}

func (i *Interpreter) executeInternalTool(impl ToolImplementation, args map[string]interface{}) (interface{}, error) {
	validatedArgs := make([]interface{}, len(impl.Spec.Args))
	for idx, argSpec := range impl.Spec.Args {
		value, provided := args[argSpec.Name]
		if !provided {
			if argSpec.Required {
				return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("tool '%s': missing required argument '%s'", impl.Spec.Name, argSpec.Name), ErrArgumentMismatch)
			}
			validatedArgs[idx] = nil
		} else {
			validatedArgs[idx] = value
		}
	}

	result, err := impl.Func(i, validatedArgs)
	if err != nil {
		if _, ok := err.(*RuntimeError); !ok {
			return nil, NewRuntimeError(ErrorCodeToolExecutionFailed, fmt.Sprintf("tool '%s' execution failed: %v", impl.Spec.Name, err), err)
		}
		return nil, err
	}
	return result, nil
}

func (i *Interpreter) ExecuteTool(toolName string, args map[string]interface{}) (interface{}, error) {
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
			return nil, NewRuntimeError(ErrorCodeRateLimited, fmt.Sprintf("tool '%s' rate limit exceeded", toolName), ErrRateLimited)
		}
		validTimestamps = append(validTimestamps, now)
		i.ToolCallTimestamps[toolName] = validTimestamps
	}

	impl, found := i.GetTool(toolName)
	if found {
		return i.executeInternalTool(impl, args)
	}

	if i.externalHandler != nil {
		methodName, ok := args["method"].(string)
		if !ok {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("external tool call to '%s' requires a 'method' argument", toolName), ErrArgumentMismatch)
		}
		result, err := i.externalHandler.CallTool(toolName, methodName, args)
		if err != nil {
			return nil, NewRuntimeError(ErrorCodeToolExecutionFailed, fmt.Sprintf("external tool '%s' failed: %v", toolName, err), err)
		}
		return result, nil
	}

	return nil, NewRuntimeError(ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", toolName), ErrToolNotFound)
}

func (i *Interpreter) ToolRegistry() ToolRegistry {
	return i
}

func (i *Interpreter) RegisterTool(impl ToolImplementation) error {
	if i.toolRegistry == nil {
		return errors.New("internal error: tool registry not initialized")
	}
	return i.toolRegistry.RegisterTool(impl)
}

func (i *Interpreter) GetTool(name string) (ToolImplementation, bool) {
	if i.toolRegistry == nil {
		return ToolImplementation{}, false
	}
	return i.toolRegistry.GetTool(name)
}

func (i *Interpreter) ListTools() []ToolSpec {
	if i.toolRegistry == nil {
		return []ToolSpec{}
	}
	return i.toolRegistry.ListTools()
}

func (i *Interpreter) SetInternalToolRegistry(registry *ToolRegistryImpl) {
	if registry != nil && registry.interpreter != i {
		registry.interpreter = i
	}
	i.toolRegistry = registry
}

func (i *Interpreter) InternalToolRegistry() *ToolRegistryImpl {
	if i.toolRegistry == nil {
		panic("FATAL: Interpreter's internal toolRegistry field is nil")
	}
	return i.toolRegistry
}
