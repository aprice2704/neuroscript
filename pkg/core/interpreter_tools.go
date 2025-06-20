// NeuroScript Version: 0.3.1
// File version: 3
// Purpose: Implements the Generic Adapter pattern. executeInternalTool now handles unwrapping Value args and wrapping primitive results, per the contract.
// filename: pkg/core/interpreter_tools.go
// nlines: 115
// risk_rating: HIGH

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

// executeInternalTool is the Generic Adapter Bridge for all internal tools.
// It accepts wrapped Values, unwraps them into primitives for the tool's Go
// function, then wraps the primitive result back into a Value.
func (i *Interpreter) executeInternalTool(impl ToolImplementation, args map[string]Value) (Value, error) {
	// UNWRAP arguments from Value -> interface{}
	validatedArgs := make([]interface{}, len(impl.Spec.Args))
	for idx, argSpec := range impl.Spec.Args {
		value, provided := args[argSpec.Name]
		if !provided {
			if argSpec.Required {
				return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("tool '%s': missing required argument '%s'", impl.Spec.Name, argSpec.Name), ErrArgumentMismatch)
			}
			validatedArgs[idx] = nil // Use nil for optional, unprovided args
		} else {
			// This is the UNWRAP step
			unwrappedValue := Unwrap(value)
			validatedArgs[idx] = unwrappedValue
		}
	}

	// CALL the primitive-based Go function
	result, err := impl.Func(i, validatedArgs)
	if err != nil {
		if _, ok := err.(*RuntimeError); !ok {
			return nil, NewRuntimeError(ErrorCodeToolExecutionFailed, fmt.Sprintf("tool '%s' execution failed: %v", impl.Spec.Name, err), err)
		}
		return nil, err
	}

	// WRAP the result from interface{} -> Value
	return Wrap(result)
}

// ExecuteTool is called by the interpreter's evaluation logic.
// It accepts a map of argument names to Values.
func (i *Interpreter) ExecuteTool(toolName string, args map[string]Value) (Value, error) {
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
		// Note: External tool handling would also need a similar unwrap/wrap bridge
		// if it were to be fully integrated with the Value system. This part remains
		// primitive-based for now.
		unwrappedArgs := make(map[string]any, len(args))
		for k, v := range args {
			unwrappedArgs[k] = Unwrap(v) // Ignoring error for simplicity here
		}
		methodName, ok := unwrappedArgs["method"].(string)
		if !ok {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("external tool call to '%s' requires a 'method' argument", toolName), ErrArgumentMismatch)
		}
		result, err := i.externalHandler.CallTool(toolName, methodName, unwrappedArgs)
		if err != nil {
			return nil, NewRuntimeError(ErrorCodeToolExecutionFailed, fmt.Sprintf("external tool '%s' failed: %v", toolName, err), err)
		}
		return Wrap(result)
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
