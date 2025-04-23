// filename: pkg/neurogo/agent_tools.go
// UPDATED: Add TOOL.AgentPin implementation and registration.
package neurogo

import (
	// Needed if AgentPinFile is restored
	"errors"
	"fmt"
	"path/filepath" // For path cleaning
	"strings"       // For error joining fallback

	"github.com/aprice2704/neuroscript/pkg/core" // Import core
)

// --- Agent Configuration Tools ---

// (toolAgentSetSandbox, toolAgentSetModel, toolAgentSetAllowlist unchanged)
func toolAgentSetSandbox(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (handle, path), got %d", core.ErrValidationArgCount, len(args))
	}
	handle, okH := args[0].(string)
	path, okP := args[1].(string)
	if !okH || !okP {
		return nil, fmt.Errorf("%w: expected string arguments for handle and path", core.ErrValidationTypeMismatch)
	}
	if handle == "" || path == "" {
		return nil, fmt.Errorf("%w: handle and path cannot be empty", core.ErrValidationRequiredArgNil)
	}
	obj, err := interpreter.GetHandleValue(handle, HandlePrefixAgentContext)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AgentSetSandbox: failed to get AgentContext: %w", err)
	}
	agentCtx, ok := obj.(*AgentContext)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentSetSandbox: %w: handle '%s' did not contain expected *neurogo.AgentContext", core.ErrInternalTool, handle)
	}
	cleanedPath := filepath.Clean(path)
	agentCtx.SetSandboxDir(cleanedPath)
	interpreter.Logger().Printf("[TOOL AgentSetSandbox] Set sandbox directory to '%s' for handle '%s'", cleanedPath, handle)
	return nil, nil
}
func toolAgentSetModel(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (handle, name), got %d", core.ErrValidationArgCount, len(args))
	}
	handle, okH := args[0].(string)
	name, okN := args[1].(string)
	if !okH || !okN {
		return nil, fmt.Errorf("%w: expected string arguments for handle and name", core.ErrValidationTypeMismatch)
	}
	if handle == "" || name == "" {
		return nil, fmt.Errorf("%w: handle and name cannot be empty", core.ErrValidationRequiredArgNil)
	}
	obj, err := interpreter.GetHandleValue(handle, HandlePrefixAgentContext)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AgentSetModel: failed to get AgentContext: %w", err)
	}
	agentCtx, ok := obj.(*AgentContext)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentSetModel: %w: handle '%s' did not contain expected *neurogo.AgentContext", core.ErrInternalTool, handle)
	}
	agentCtx.SetModelName(name)
	interpreter.Logger().Printf("[TOOL AgentSetModel] Set model name to '%s' for handle '%s'", name, handle)
	return nil, nil
}
func toolAgentSetAllowlist(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (handle, path), got %d", core.ErrValidationArgCount, len(args))
	}
	handle, okH := args[0].(string)
	path, okP := args[1].(string)
	if !okH || !okP {
		return nil, fmt.Errorf("%w: expected string arguments for handle and path", core.ErrValidationTypeMismatch)
	}
	if handle == "" {
		return nil, fmt.Errorf("%w: handle cannot be empty", core.ErrValidationRequiredArgNil)
	}
	cleanedPath := ""
	if path != "" {
		cleanedPath = filepath.Clean(path)
	} else {
		interpreter.Logger().Printf("[TOOL AgentSetAllowlist] Warning: Setting empty allowlist path for handle '%s', effectively disabling allowlist.", handle)
	}
	obj, err := interpreter.GetHandleValue(handle, HandlePrefixAgentContext)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AgentSetAllowlist: failed to get AgentContext: %w", err)
	}
	agentCtx, ok := obj.(*AgentContext)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentSetAllowlist: %w: handle '%s' did not contain expected *neurogo.AgentContext", core.ErrInternalTool, handle)
	}
	agentCtx.SetAllowlistPath(cleanedPath)
	interpreter.Logger().Printf("[TOOL AgentSetAllowlist] Set allowlist path to '%s' for handle '%s'", cleanedPath, handle)
	return nil, nil
}

// --- REMOVED: toolAgentPinFile (replaced by UpsertAs + AgentPin) ---

// --- NEW: TOOL.AgentPin ---
// toolAgentPin takes a map (from UpsertAs) and adds the file info to the pinned list.
func toolAgentPin(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	// Args: handle (string), fileInfoMap (map)
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (handle, fileInfoMap), got %d", core.ErrValidationArgCount, len(args))
	}
	handle, okH := args[0].(string)
	fileInfoMap, okM := args[1].(map[string]interface{}) // Expect map[string]interface{}
	if !okH {
		return nil, fmt.Errorf("%w: expected string handle argument", core.ErrValidationTypeMismatch)
	}
	if !okM {
		return nil, fmt.Errorf("%w: expected map fileInfoMap argument, got %T", core.ErrValidationTypeMismatch, args[1])
	}
	if handle == "" {
		return nil, fmt.Errorf("%w: handle cannot be empty", core.ErrValidationRequiredArgNil)
	}

	logger := interpreter.Logger()

	// 1. Get AgentContext
	obj, err := interpreter.GetHandleValue(handle, HandlePrefixAgentContext)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AgentPin: failed to get AgentContext: %w", err)
	}
	agentCtx, ok := obj.(*AgentContext)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentPin: %w: handle '%s' did not contain expected *neurogo.AgentContext", core.ErrInternalTool, handle)
	}

	// 2. Extract info from map
	// Expecting {"displayName": string, "uri": string} from UpsertAs
	displayNameVal, okDN := fileInfoMap["displayName"]
	uriVal, okURI := fileInfoMap["uri"]
	if !okDN || !okURI {
		return nil, fmt.Errorf("TOOL.AgentPin: %w: fileInfoMap missing 'displayName' or 'uri' keys", core.ErrValidationArgValue)
	}

	displayName, okDNStr := displayNameVal.(string)
	uri, okURIStr := uriVal.(string)
	if !okDNStr || !okURIStr {
		return nil, fmt.Errorf("TOOL.AgentPin: %w: fileInfoMap 'displayName' (%T) or 'uri' (%T) not strings", core.ErrValidationArgValue, displayNameVal, uriVal)
	}
	if displayName == "" || uri == "" {
		return nil, fmt.Errorf("TOOL.AgentPin: %w: fileInfoMap 'displayName' or 'uri' cannot be empty", core.ErrValidationArgValue)
	}
	// Use displayName as the key for pinning, as relativePath isn't available from UpsertAs(contents, name)
	keyPath := displayName

	// 3. Pin file in AgentContext
	pinErr := agentCtx.PinFile(keyPath, uri)
	if pinErr != nil {
		logger.Printf("[TOOL AgentPin] Error pinning file in AgentContext: %v", pinErr)
		return nil, fmt.Errorf("TOOL.AgentPin: failed to pin file '%s' (URI: %s) in context: %w", keyPath, uri, pinErr)
	}

	logger.Printf("[TOOL AgentPin] Successfully pinned '%s' (URI: %s)", keyPath, uri)
	return nil, nil // Success
}

// TODO: Implement TOOL.AgentSyncDirectory(handle string, path string, filter string, ignoreGitignore bool) error

// --- Registration ---

// RegisterAgentTools registers all tools specific to agent configuration with the provided registry.
// UPDATED: Add AgentPin registration
func RegisterAgentTools(registry *core.ToolRegistry) error {
	tools := []core.ToolImplementation{
		// Existing tools...
		{Spec: core.ToolSpec{Name: "AgentSetSandbox", Description: "Sets the agent's sandbox directory.", Args: []core.ArgSpec{{Name: "agentCtxHandle", Type: core.ArgTypeString, Required: true}, {Name: "path", Type: core.ArgTypeString, Required: true}}, ReturnType: core.ArgTypeAny}, Func: toolAgentSetSandbox},
		{Spec: core.ToolSpec{Name: "AgentSetModel", Description: "Sets the AI model name for the agent.", Args: []core.ArgSpec{{Name: "agentCtxHandle", Type: core.ArgTypeString, Required: true}, {Name: "name", Type: core.ArgTypeString, Required: true}}, ReturnType: core.ArgTypeAny}, Func: toolAgentSetModel},
		{Spec: core.ToolSpec{Name: "AgentSetAllowlist", Description: "Sets the path to the tool allowlist file for the agent.", Args: []core.ArgSpec{{Name: "agentCtxHandle", Type: core.ArgTypeString, Required: true}, {Name: "path", Type: core.ArgTypeString, Required: true}}, ReturnType: core.ArgTypeAny}, Func: toolAgentSetAllowlist},
		// --- REMOVED AgentPinFile Registration ---
		// --- NEW Tool Registration ---
		{
			Spec: core.ToolSpec{
				Name: "AgentPin", Description: "Adds a file (identified by map from UpsertAs) to the agent's persistent context.",
				Args: []core.ArgSpec{
					{Name: "agentCtxHandle", Type: core.ArgTypeString, Required: true, Description: "Handle to the agent context object."},
					{Name: "fileInfoMap", Type: core.ArgTypeMap, Required: true, Description: "Map containing 'displayName' and 'uri' from TOOL.UpsertAs."},
				}, ReturnType: core.ArgTypeAny, // No return value
			}, Func: toolAgentPin,
		},
		// Add TOOL.AgentSyncDirectory spec here
	}

	var errs []error
	for _, tool := range tools {
		if err := registry.RegisterTool(tool); err != nil {
			errs = append(errs, fmt.Errorf("failed to register agent tool %s: %w", tool.Spec.Name, err))
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
