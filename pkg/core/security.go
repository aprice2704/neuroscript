// filename: pkg/core/security.go
package core

import (
	"fmt"
	"log"
	"path/filepath"
)

// SecurityLayer enforces security policies for LLM-initiated tool calls.
type SecurityLayer struct {
	allowlist    map[string]bool
	denylist     map[string]bool // Added denylist map
	sandboxRoot  string
	toolRegistry *ToolRegistry
	logger       *log.Logger
}

// NewSecurityLayer creates a new security layer instance.
// Now accepts allowlist (slice) and denylist (map).
func NewSecurityLayer(allowlistTools []string, denylistSet map[string]bool, sandboxRoot string, registry *ToolRegistry, logger *log.Logger) *SecurityLayer {
	allowlistMap := make(map[string]bool)
	for _, tool := range allowlistTools {
		allowlistMap[tool] = true
	}
	cleanSandboxRoot := filepath.Clean(sandboxRoot)

	// Log effective lists
	logger.Printf("[SEC] Initialized Security Layer.")
	logger.Printf("[SEC] Allowlisted tools (initial): %v", allowlistTools)
	deniedToolNames := make([]string, 0, len(denylistSet))
	for tool := range denylistSet {
		deniedToolNames = append(deniedToolNames, tool)
	}
	logger.Printf("[SEC] Denied tools: %v", deniedToolNames)
	logger.Printf("[SEC] Sandbox Root: %s", cleanSandboxRoot)

	if registry == nil {
		logger.Printf("[WARN SEC] SecurityLayer initialized with nil ToolRegistry. Argument validation will be significantly limited.")
	}

	return &SecurityLayer{
		allowlist:    allowlistMap,
		denylist:     denylistSet, // Store the denylist map
		sandboxRoot:  cleanSandboxRoot,
		toolRegistry: registry,
		logger:       logger,
	}
}

// ValidateToolCall checks denylist, allowlist, high-risk status, and delegates argument validation.
func (sl *SecurityLayer) ValidateToolCall(toolName string, rawArgs map[string]interface{}) (map[string]interface{}, error) {
	sl.logger.Printf("[SEC] Validating request for tool: %s with raw args: %v", toolName, rawArgs)

	// 1. Check Denylist FIRST
	if sl.denylist[toolName] {
		errMsg := fmt.Sprintf("tool '%s' is explicitly denied by denylist", toolName)
		sl.logger.Printf("[SEC] DENIED: %s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	sl.logger.Printf("[SEC] Tool '%s' is not denied.", toolName)

	// 2. Check Allowlist (only if not denied)
	if !sl.allowlist[toolName] {
		errMsg := fmt.Sprintf("tool '%s' is not in the allowlist for LLM execution", toolName)
		sl.logger.Printf("[SEC] DENIED: %s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	sl.logger.Printf("[SEC] Tool '%s' is allowlisted.", toolName)

	// 3. High-Risk Tool Check (Example - this check could also use the denylist)
	if toolName == "TOOL.ExecuteCommand" {
		// This check is somewhat redundant if ExecuteCommand is typically denied,
		// but provides defense in depth if denylist loading fails or is misconfigured.
		errMsg := "tool 'TOOL.ExecuteCommand' is blocked by default for LLM execution"
		sl.logger.Printf("[SEC] DENIED: %s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	// TODO: Add more checks/configurability for risky tools like WriteFile

	// 4. Get Tool Specification
	if sl.toolRegistry == nil {
		sl.logger.Printf("[WARN SEC] ToolRegistry not available. Skipping argument validation for tool '%s'. Returning raw args.", toolName)
		// !! SECURITY RISK: Returning raw args without validation !!
		return rawArgs, nil
	}
	toolImpl, found := sl.toolRegistry.GetTool(toolName)
	if !found {
		// Should be caught by allowlist check generally, but good to double-check
		return nil, fmt.Errorf("internal security error: allowlisted tool '%s' not found in registry", toolName)
	}
	toolSpec := toolImpl.Spec
	sl.logger.Printf("[SEC] Validating args for '%s' against spec (%d args)", toolName, len(toolSpec.Args))

	// 5. Delegate Argument Validation (calls function in security_validation.go)
	validatedArgs, validationErr := sl.validateArgumentsAgainstSpec(toolSpec, rawArgs)
	if validationErr != nil {
		sl.logger.Printf("[SEC] DENIED (Argument Validation): %v", validationErr)
		return nil, validationErr
	}

	// If all checks passed:
	sl.logger.Printf("[SEC] All arguments for '%s' validated successfully. Validated Args: %v", toolName, validatedArgs)
	return validatedArgs, nil
}

// Note: validateArgumentsAgainstSpec remains in security_validation.go
// Note: Type helpers remain in security_helpers.go
