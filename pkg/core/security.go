// filename: pkg/core/security.go
package core

import (
	// Import errors package
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

// IF YOU ARE THINKING OF REMOVING SOMETHING FROM THIS FILE YOU ARE PROBABLY WRONG
// ASK FIRST!!!!

// SecurityLayer enforces security policies for LLM-initiated tool calls.
type SecurityLayer struct {
	allowlist    map[string]bool
	denylist     map[string]bool
	sandboxRoot  string
	toolRegistry *ToolRegistry
	logger       *log.Logger
}

// NewSecurityLayer creates a new security layer instance.
func NewSecurityLayer(allowlistTools []string, denylistSet map[string]bool, sandboxRoot string, registry *ToolRegistry, logger *log.Logger) *SecurityLayer {
	allowlistMap := make(map[string]bool)
	for _, tool := range allowlistTools {
		allowlistMap[tool] = true
	}
	cleanSandboxRoot := filepath.Clean(sandboxRoot)

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
		denylist:     denylistSet,
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
		sl.logger.Printf("[SEC] DENIED: Tool %q is explicitly denied by denylist.", toolName)
		// Use Sentinel Error + Wrapping
		return nil, fmt.Errorf("tool %q denied: %w", toolName, ErrToolDenied)
	}
	sl.logger.Printf("[SEC] Tool '%s' is not denied.", toolName)

	// 2. Check Allowlist (only if not denied)
	if !sl.allowlist[toolName] {
		sl.logger.Printf("[SEC] DENIED: Tool %q is not in the allowlist for LLM execution.", toolName)
		// Use Sentinel Error + Wrapping
		return nil, fmt.Errorf("tool %q not allowed: %w", toolName, ErrToolNotAllowed)
	}
	sl.logger.Printf("[SEC] Tool '%s' is allowlisted.", toolName)

	// 3. High-Risk Tool Check
	if toolName == "TOOL.ExecuteCommand" {
		sl.logger.Printf("[SEC] DENIED: Tool 'TOOL.ExecuteCommand' is blocked by policy.")
		// Use Sentinel Error + Wrapping
		return nil, fmt.Errorf("tool %q blocked: %w", toolName, ErrToolBlocked)
	}
	// TODO: Add more checks/configurability for risky tools like WriteFile

	// 4. Get Tool Specification
	if sl.toolRegistry == nil {
		sl.logger.Printf("[WARN SEC] ToolRegistry not available. Skipping argument validation for tool '%s'. Returning raw args.", toolName)
		// !! SECURITY RISK: Returning raw args without validation !!
		// Returning error here instead of raw args to prevent insecure operation
		return nil, fmt.Errorf("tool registry unavailable for tool %q: %w", toolName, ErrInternalSecurity)
	}
	toolImpl, found := sl.toolRegistry.GetTool(toolName)
	if !found {
		sl.logger.Printf("[SEC] DENIED: Allowlisted tool %q not found in registry.", toolName)
		// Should be caught by allowlist check generally, but good to double-check
		// Use Sentinel Error + Wrapping
		return nil, fmt.Errorf("allowlisted tool %q not found in registry: %w", toolName, ErrInternalSecurity)
	}
	toolSpec := toolImpl.Spec
	sl.logger.Printf("[SEC] Validating args for '%s' against spec (%d args)", toolName, len(toolSpec.Args))

	// 5. Delegate Argument Validation (calls function in security_validation.go)
	validatedArgs, validationErr := sl.validateArgumentsAgainstSpec(toolSpec, rawArgs) // Variable is 'validationErr'
	if validationErr != nil {
		// Logging should happen within validateArgumentsAgainstSpec or its callees
		// *** FIXED TYPO: Use validationErr ***
		sl.logger.Printf("[SEC] DENIED (Argument Validation): Tool %q, Error: %v", toolName, validationErr)
		// Return error directly, assumes it's already properly formed/wrapped
		// *** FIXED TYPO: Use validationErr ***
		return nil, validationErr
	}

	// If all checks passed:
	sl.logger.Printf("[SEC] All arguments for '%s' validated successfully. Validated Args: %v", toolName, validatedArgs)
	return validatedArgs, nil
}

// SanitizeFilename cleans a string to be suitable for use as a filename component.
// DO NOT MOVE from security.go --- this means YOU
func SanitizeFilename(name string) string {
	// ... (Sanitization logic remains the same) ...
	name = strings.ReplaceAll(name, " ", "_")
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	name = replacer.Replace(name)
	removeChars := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	name = removeChars.ReplaceAllString(name, "")
	name = strings.Trim(name, "._-")
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")
	for strings.Contains(name, "..") {
		name = strings.ReplaceAll(name, "..", "_")
	}
	name = strings.Trim(name, "._-")
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")
	const maxLength = 100
	if len(name) > maxLength {
		name = name[:maxLength]
		name = strings.TrimRight(name, "._-")
	}
	if name == "" {
		name = "default_sanitized_name"
	}
	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperName := strings.ToUpper(name)
	baseName := upperName
	if dotIndex := strings.LastIndex(upperName, "."); dotIndex != -1 {
		baseName = upperName[:dotIndex]
	}
	for _, r := range reserved {
		if upperName == r || baseName == r {
			name = name + "_"
			break
		}
	}
	return name
}

// SecureFilePath cleans and ensures the **relative** path is within the allowed directory (cwd).
// Rejects absolute paths. Returns the cleaned absolute path on success.
// DO NOT MOVE from security.go --- this means YOU
func SecureFilePath(filePath, allowedDir string) (string, error) {
	if filePath == "" {
		// Wrap sentinel error
		return "", fmt.Errorf("file path cannot be empty: %w", ErrPathViolation)
	}
	if strings.Contains(filePath, "\x00") {
		// Use correct sentinel error + Wrap
		return "", fmt.Errorf("file path contains null byte: %w", ErrNullByteInArgument)
	}
	if filepath.IsAbs(filePath) {
		// Wrap sentinel error
		return "", fmt.Errorf("input file path %q must be relative: %w", filePath, ErrPathViolation)
	}

	absAllowedDir, err := filepath.Abs(allowedDir)
	if err != nil {
		// Wrap internal config error
		return "", fmt.Errorf("could not get absolute path for allowed directory %q: %w", allowedDir, err)
	}
	absAllowedDir = filepath.Clean(absAllowedDir)

	absCleanedPath := filepath.Join(absAllowedDir, filePath)
	absCleanedPath = filepath.Clean(absCleanedPath)

	prefixToCheck := absAllowedDir
	if prefixToCheck != string(filepath.Separator) && !strings.HasSuffix(prefixToCheck, string(filepath.Separator)) {
		prefixToCheck += string(filepath.Separator)
	}
	pathToCheck := absCleanedPath

	isOutside := pathToCheck != absAllowedDir && !strings.HasPrefix(pathToCheck, prefixToCheck)

	if isOutside {
		details := fmt.Sprintf("relative path %q resolves to %q which is outside the allowed directory %q", filePath, absCleanedPath, absAllowedDir)
		// Wrap sentinel error
		return "", fmt.Errorf("%s: %w", details, ErrPathViolation)
	}

	return absCleanedPath, nil
}
