// NeuroScript Version: 0.3.0
// File version: 0.1.7
// Corrected tool function signatures to accept tool.Runtime and perform type assertion.
// filename: pkg/neurogo/agent_tools.go
// nlines: 210
// risk_rating: LOW
package neurogo

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- Agent Configuration Tools ---

func toolAgentSetSandbox(rt tool.Runtime, args []interface{}) (interface{}, error) {
	i, ok := rt.(*interpreter.Interpreter)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentSetSandbox: runtime is not a valid interpreter")
	}

	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (handle, path), got %d", lang.ErrValidationArgCount, len(args))
	}
	handle, okH := args[0].(string)
	path, okP := args[1].(string)
	if !okH || !okP {
		return nil, fmt.Errorf("%w: expected string arguments for handle and path", lang.ErrValidationTypeMismatch)
	}
	if handle == "" || path == "" {
		return nil, fmt.Errorf("%w: handle and path cannot be empty", lang.ErrValidationRequiredArgNil)
	}
	obj, err := i.GetHandleValue(handle, HandlePrefixAgentContext)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AgentSetSandbox: failed to get AgentContext: %w", err)
	}
	agentCtx, ok := obj.(*AgentContext)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentSetSandbox: %w: handle '%s' did not contain expected *neurogo.AgentContext", lang.ErrInternalTool, handle)
	}
	cleanedPath := filepath.Clean(path)
	agentCtx.SetSandboxDir(cleanedPath)
	i.Logger().Debug("[TOOL AgentSetSandbox] Set sandbox directory to '%s' for handle '%s'", cleanedPath, handle)
	return nil, nil
}

func toolAgentSetModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	i, ok := rt.(*interpreter.Interpreter)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentSetModel: runtime is not a valid interpreter")
	}

	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (handle, name), got %d", lang.ErrValidationArgCount, len(args))
	}
	handle, okH := args[0].(string)
	name, okN := args[1].(string)
	if !okH || !okN {
		return nil, fmt.Errorf("%w: expected string arguments for handle and name", lang.ErrValidationTypeMismatch)
	}
	if handle == "" || name == "" {
		return nil, fmt.Errorf("%w: handle and name cannot be empty", lang.ErrValidationRequiredArgNil)
	}
	obj, err := i.GetHandleValue(handle, HandlePrefixAgentContext)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AgentSetModel: failed to get AgentContext: %w", err)
	}
	agentCtx, ok := obj.(*AgentContext)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentSetModel: %w: handle '%s' did not contain expected *neurogo.AgentContext", lang.ErrInternalTool, handle)
	}
	agentCtx.SetModelName(name)
	i.Logger().Debug("[TOOL AgentSetModel] Set model name to '%s' for handle '%s'", name, handle)
	return nil, nil
}

func toolAgentSetAllowlist(rt tool.Runtime, args []interface{}) (interface{}, error) {
	i, ok := rt.(*interpreter.Interpreter)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentSetAllowlist: runtime is not a valid interpreter")
	}
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (handle, path), got %d", lang.ErrValidationArgCount, len(args))
	}
	handle, okH := args[0].(string)
	path, okP := args[1].(string)
	if !okH || !okP {
		return nil, fmt.Errorf("%w: expected string arguments for handle and path", lang.ErrValidationTypeMismatch)
	}
	if handle == "" {
		return nil, fmt.Errorf("%w: handle cannot be empty", lang.ErrValidationRequiredArgNil)
	}
	cleanedPath := ""
	if path != "" {
		cleanedPath = filepath.Clean(path)
	} else {
		i.Logger().Debug("[TOOL AgentSetAllowlist] Warning: Setting empty allowlist path for handle '%s', effectively disabling allowlist.", handle)
	}
	obj, err := i.GetHandleValue(handle, HandlePrefixAgentContext)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AgentSetAllowlist: failed to get AgentContext: %w", err)
	}
	agentCtx, ok := obj.(*AgentContext)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentSetAllowlist: %w: handle '%s' did not contain expected *neurogo.AgentContext", lang.ErrInternalTool, handle)
	}
	agentCtx.SetAllowlistPath(cleanedPath)
	i.Logger().Debug("[TOOL AgentSetAllowlist] Set allowlist path to '%s' for handle '%s'", cleanedPath, handle)
	return nil, nil
}

func toolAgentPin(rt tool.Runtime, args []interface{}) (interface{}, error) {
	i, ok := rt.(*interpreter.Interpreter)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentPin: runtime is not a valid interpreter")
	}

	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (handle, fileInfoMap), got %d", lang.ErrValidationArgCount, len(args))
	}
	handle, okH := args[0].(string)
	fileInfoMap, okM := args[1].(map[string]interface{})
	if !okH {
		return nil, fmt.Errorf("%w: expected string handle argument", lang.ErrValidationTypeMismatch)
	}
	if !okM {
		return nil, fmt.Errorf("%w: expected map fileInfoMap argument, got %T", lang.ErrValidationTypeMismatch, args[1])
	}
	if handle == "" {
		return nil, fmt.Errorf("%w: handle cannot be empty", lang.ErrValidationRequiredArgNil)
	}

	logger := i.Logger()

	obj, err := i.GetHandleValue(handle, HandlePrefixAgentContext)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AgentPin: failed to get AgentContext: %w", err)
	}
	agentCtx, ok := obj.(*AgentContext)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentPin: %w: handle '%s' did not contain expected *neurogo.AgentContext", lang.ErrInternalTool, handle)
	}

	displayNameVal, okDN := fileInfoMap["displayName"]
	uriVal, okURI := fileInfoMap["uri"]
	if !okDN || !okURI {
		return nil, fmt.Errorf("TOOL.AgentPin: %w: fileInfoMap missing 'displayName' or 'uri' keys", lang.ErrValidationArgValue)
	}

	displayName, okDNStr := displayNameVal.(string)
	uri, okURIStr := uriVal.(string)
	if !okDNStr || !okURIStr {
		return nil, fmt.Errorf("TOOL.AgentPin: %w: fileInfoMap 'displayName' (%T) or 'uri' (%T) not strings", lang.ErrValidationArgValue, displayNameVal, uriVal)
	}
	if displayName == "" || uri == "" {
		return nil, fmt.Errorf("TOOL.AgentPin: %w: fileInfoMap 'displayName' or 'uri' cannot be empty", lang.ErrValidationArgValue)
	}
	keyPath := displayName

	pinErr := agentCtx.PinFile(keyPath, uri)
	if pinErr != nil {
		logger.Error("[TOOL AgentPin] Error pinning file in AgentContext: %v", pinErr)
		return nil, fmt.Errorf("TOOL.AgentPin: failed to pin file '%s' (URI: %s) in context: %w", keyPath, uri, pinErr)
	}

	logger.Debug("[TOOL AgentPin] Successfully pinned '%s' (URI: %s)", keyPath, uri)
	return nil, nil
}

// RegisterAgentTools registers all tools specific to agent configuration with the provided registry.
func RegisterAgentTools(registry tool.ToolRegistry) error {
	tools := []tool.ToolImplementation{
		{Spec: tool.ToolSpec{Name: "AgentSetSandbox", Description: "Sets the agent's sandbox directory.", Args: []tool.ArgSpec{{Name: "agentCtxHandle", Type: tool.ArgTypeString, Required: true}, {Name: "path", Type: tool.ArgTypeString, Required: true}}, ReturnType: tool.ArgTypeAny}, Func: toolAgentSetSandbox},
		{Spec: tool.ToolSpec{Name: "AgentSetModel", Description: "Sets the AI model name for the agent.", Args: []tool.ArgSpec{{Name: "agentCtxHandle", Type: tool.ArgTypeString, Required: true}, {Name: "name", Type: tool.ArgTypeString, Required: true}}, ReturnType: tool.ArgTypeAny}, Func: toolAgentSetModel},
		{Spec: tool.ToolSpec{Name: "AgentSetAllowlist", Description: "Sets the path to the tool allowlist file for the agent.", Args: []tool.ArgSpec{{Name: "agentCtxHandle", Type: tool.ArgTypeString, Required: true}, {Name: "path", Type: tool.ArgTypeString, Required: true}}, ReturnType: tool.ArgTypeAny}, Func: toolAgentSetAllowlist},
		{
			Spec: tool.ToolSpec{
				Name: "AgentPin", Description: "Adds a file (identified by map from UpsertAs) to the agent's persistent context.",
				Args: []tool.ArgSpec{
					{Name: "agentCtxHandle", Type: tool.ArgTypeString, Required: true, Description: "Handle to the agent context object."},
					{Name: "fileInfoMap", Type: tool.ArgTypeMap, Required: true, Description: "Map containing 'displayName' and 'uri' from TOOL.UpsertAs."},
				}, ReturnType: tool.ArgTypeAny,
			}, Func: toolAgentPin,
		},
	}

	var errs []error
	for _, t := range tools {
		if _, err := registry.RegisterTool(t); err != nil {
			errs = append(errs, fmt.Errorf("failed to register agent tool %s: %w", t.Spec.Name, err))
		}
	}

	if len(errs) > 0 {
		errorMessages := make([]string, len(errs))
		for i, e := range errs {
			errorMessages[i] = e.Error()
		}
		return errors.New(strings.Join(errorMessages, "; "))
	}
	return nil
}
