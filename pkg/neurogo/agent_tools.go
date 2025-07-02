// NeuroScript Version: 0.3.0
// File version: 0.1.4
// Correct ToolRegistry type in RegisterAgentTools
// filename: pkg/neurogo/agent_tools.go
// nlines: 190
// risk_rating: LOW
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
func toolAgentSetSandbox(interpreter * Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (handle, path), got %d",  alidationArgCount, len(args))
	}
	handle, okH := args[0].(string)
	path, okP := args[1].(string)
	if !okH || !okP {
		return nil, fmt.Errorf("%w: expected string arguments for handle and path",  alidationTypeMismatch)
	}
	if handle == "" || path == "" {
		return nil, fmt.Errorf("%w: handle and path cannot be empty",  alidationRequiredArgNil)
	}
	obj, err := interpreter.GetHandleValue(handle, HandlePrefixAgentContext)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AgentSetSandbox: failed to get AgentContext: %w", err)
	}
	agentCtx, ok := obj.(*AgentContext)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentSetSandbox: %w: handle '%s' did not contain expected *neurogo.AgentContext",  nternalTool, handle)
	}
	cleanedPath := filepath.Clean(path)
	agentCtx.SetSandboxDir(cleanedPath)
	interpreter.Logger().Debug("[TOOL AgentSetSandbox] Set sandbox directory to '%s' for handle '%s'", cleanedPath, handle)
	return nil, nil
}
func toolAgentSetModel(interpreter * rpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (handle, name), got %d",  alidationArgCount, len(args))
	}
	handle, okH := args[0].(string)
	name, okN := args[1].(string)
	if !okH || !okN {
		return nil, fmt.Errorf("%w: expected string arguments for handle and name",  alidationTypeMismatch)
	}
	if handle == "" || name == "" {
		return nil, fmt.Errorf("%w: handle and name cannot be empty",  alidationRequiredArgNil)
	}
	obj, err := interpreter.GetHandleValue(handle, HandlePrefixAgentContext)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AgentSetModel: failed to get AgentContext: %w", err)
	}
	agentCtx, ok := obj.(*AgentContext)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentSetModel: %w: handle '%s' did not contain expected *neurogo.AgentContext",  nternalTool, handle)
	}
	agentCtx.SetModelName(name)
	interpreter.Logger().Debug("[TOOL AgentSetModel] Set model name to '%s' for handle '%s'", name, handle)
	return nil, nil
}
func toolAgentSetAllowlist(interpreter * rpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (handle, path), got %d",  alidationArgCount, len(args))
	}
	handle, okH := args[0].(string)
	path, okP := args[1].(string)
	if !okH || !okP {
		return nil, fmt.Errorf("%w: expected string arguments for handle and path",  alidationTypeMismatch)
	}
	if handle == "" {
		return nil, fmt.Errorf("%w: handle cannot be empty",  alidationRequiredArgNil)
	}
	cleanedPath := ""
	if path != "" {
		cleanedPath = filepath.Clean(path)
	} else {
		interpreter.Logger().Debug("[TOOL AgentSetAllowlist] Warning: Setting empty allowlist path for handle '%s', effectively disabling allowlist.", handle)
	}
	obj, err := interpreter.GetHandleValue(handle, HandlePrefixAgentContext)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AgentSetAllowlist: failed to get AgentContext: %w", err)
	}
	agentCtx, ok := obj.(*AgentContext)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentSetAllowlist: %w: handle '%s' did not contain expected *neurogo.AgentContext",  nternalTool, handle)
	}
	agentCtx.SetAllowlistPath(cleanedPath)
	interpreter.Logger().Debug("[TOOL AgentSetAllowlist] Set allowlist path to '%s' for handle '%s'", cleanedPath, handle)
	return nil, nil
}

// --- REMOVED: toolAgentPinFile (replaced by UpsertAs + AgentPin) ---

// --- NEW: TOOL.AgentPin ---
// toolAgentPin takes a map (from UpsertAs) and adds the file info to the pinned list.
func toolAgentPin(interpreter * rpreter, args []interface{}) (interface{}, error) {
	// Args: handle (string), fileInfoMap (map)
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: expected 2 arguments (handle, fileInfoMap), got %d",  alidationArgCount, len(args))
	}
	handle, okH := args[0].(string)
	fileInfoMap, okM := args[1].(map[string]interface{}) // Expect map[string]interface{}
	if !okH {
		return nil, fmt.Errorf("%w: expected string handle argument",  alidationTypeMismatch)
	}
	if !okM {
		return nil, fmt.Errorf("%w: expected map fileInfoMap argument, got %T",  alidationTypeMismatch, args[1])
	}
	if handle == "" {
		return nil, fmt.Errorf("%w: handle cannot be empty",  alidationRequiredArgNil)
	}

	logger := interpreter.Logger()

	// 1. Get AgentContext
	obj, err := interpreter.GetHandleValue(handle, HandlePrefixAgentContext)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AgentPin: failed to get AgentContext: %w", err)
	}
	agentCtx, ok := obj.(*AgentContext)
	if !ok {
		return nil, fmt.Errorf("TOOL.AgentPin: %w: handle '%s' did not contain expected *neurogo.AgentContext",  nternalTool, handle)
	}

	// 2. Extract info from map
	// Expecting {"displayName": string, "uri": string} from UpsertAs
	displayNameVal, okDN := fileInfoMap["displayName"]
	uriVal, okURI := fileInfoMap["uri"]
	if !okDN || !okURI {
		return nil, fmt.Errorf("TOOL.AgentPin: %w: fileInfoMap missing 'displayName' or 'uri' keys",  alidationArgValue)
	}

	displayName, okDNStr := displayNameVal.(string)
	uri, okURIStr := uriVal.(string)
	if !okDNStr || !okURIStr {
		return nil, fmt.Errorf("TOOL.AgentPin: %w: fileInfoMap 'displayName' (%T) or 'uri' (%T) not strings",  alidationArgValue, displayNameVal, uriVal)
	}
	if displayName == "" || uri == "" {
		return nil, fmt.Errorf("TOOL.AgentPin: %w: fileInfoMap 'displayName' or 'uri' cannot be empty",  alidationArgValue)
	}
	// Use displayName as the key for pinning, as relativePath isn't available from UpsertAs(contents, name)
	keyPath := displayName

	// 3. Pin file in AgentContext
	pinErr := agentCtx.PinFile(keyPath, uri)
	if pinErr != nil {
		logger.Error("[TOOL AgentPin] Error pinning file in AgentContext: %v", pinErr)
		return nil, fmt.Errorf("TOOL.AgentPin: failed to pin file '%s' (URI: %s) in context: %w", keyPath, uri, pinErr)
	}

	logger.Debug("[TOOL AgentPin] Successfully pinned '%s' (URI: %s)", keyPath, uri)
	return nil, nil // Success
}

// TODO: Implement TOOL.AgentSyncDirectory(handle string, path string, filter string, ignoreGitignore bool) error

// --- Registration ---

// RegisterAgentTools registers all tools specific to agent configuration with the provided registry.
// UPDATED: Add AgentPin registration
// CORRECTED: Changed registry type from * Registry to  Too stry
func RegisterAgentTools(registry  Registry) error {
	tools := [] Implementation{
		// Existing tools...
		{Spec:  Spec{Name: "AgentSetSandbox", Description: "Sets the agent's sandbox directory.", Args: [] Arg {Name: "agentCtxHandle", Type:  ArgType g, Required: true}, {Name: "path", Type:  ArgTypeStri equired: true}}, ReturnType:  ArgTypeAny}, Fu oolAgentSetSandbox},
		{Spec:  Spec{Name: "AgentSetModel", Description: "Sets the AI model name for the agent.", Args: [] Arg {Name: "agentCtxHandle", Type:  ArgType g, Required: true}, {Name: "name", Type:  ArgTypeStri equired: true}}, ReturnType:  ArgTypeAny}, Fu oolAgentSetModel},
		{Spec:  Spec{Name: "AgentSetAllowlist", Description: "Sets the path to the tool allowlist file for the agent.", Args: [] Arg {Name: "agentCtxHandle", Type:  ArgType g, Required: true}, {Name: "path", Type:  ArgTypeStri equired: true}}, ReturnType:  ArgTypeAny}, Fu oolAgentSetAllowlist},
		// --- REMOVED AgentPinFile Registration ---
		// --- NEW Tool Registration ---
		{
			Spec:  Spec{
				Name: "AgentPin", Description: "Adds a file (identified by map from UpsertAs) to the agent's persistent context.",
				Args: [] pec{
					{Name: "agentCtxHandle", Type:  ypeString, Required: true, Description: "Handle to the agent context object."},
					{Name: "fileInfoMap", Type:  ypeMap, Required: true, Description: "Map containing 'displayName' and 'uri' from TOOL.UpsertAs."},
				}, ReturnType:  ypeAny, // No return value
			}, Func: toolAgentPin,
		},
		// Add TOOL.AgentSyncDirectory spec here
	}

	var errs []error
	for _, tool := range tools {
		// This call is now correct because 'registry' is  Registry (interface)
		// and 'RegisterTool' is a method on that interface.
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
