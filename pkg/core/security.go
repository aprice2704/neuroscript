// filename: pkg/core/security.go
package core

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings" // Import strings for SecureFilePath
)

// IF YOU ARE THINKING OF REMOVING SOMETHING FROM THIS FILE YOU ARE PROBABLY WRONG
// ASK FIRST!!!!

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

// SanitizeFilename cleans a string to be suitable for use as a filename component.
// DO NOT MOVE from security.go --- this means YOU
func SanitizeFilename(name string) string {
	// 1. Replace spaces and explicitly problematic characters with underscore
	// Problematic chars: /\:*?"<>|
	name = strings.ReplaceAll(name, " ", "_")
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	name = replacer.Replace(name)

	// 2. Remove any *remaining* characters not allowed (conservative set)
	// This handles anything missed by the replacer or potentially problematic Unicode.
	removeChars := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	name = removeChars.ReplaceAllString(name, "")

	// 3. Remove leading/trailing unwanted chars (dots, underscores, hyphens)
	name = strings.Trim(name, "._-")

	// 4. Collapse multiple underscores/hyphens, remove '..'
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")
	// Avoid sequences like ".." which can be problematic, replace with single underscore
	for strings.Contains(name, "..") {
		name = strings.ReplaceAll(name, "..", "_") // Use underscore to avoid creating "--" or "__"
	}
	// Need to re-run Trim and Collapse after potentially creating new leading/trailing underscores
	name = strings.Trim(name, "._-")
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")

	// 5. Truncate to a reasonable length (e.g., 100)
	const maxLength = 100
	if len(name) > maxLength {
		name = name[:maxLength]
		// Re-trim after truncation in case it ended on a bad character
		name = strings.TrimRight(name, "._-")
	}

	// 6. Handle empty result after sanitization
	if name == "" {
		name = "default_sanitized_name"
	}

	// 7. Avoid reserved names (case-insensitive check on Windows)
	// This is a simplified check; a more robust solution might be needed
	// depending on target platforms and thoroughness required.
	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperName := strings.ToUpper(name)
	baseName := upperName
	if dotIndex := strings.LastIndex(upperName, "."); dotIndex != -1 {
		baseName = upperName[:dotIndex] // Check base name too (e.g., CON.txt)
	}

	for _, r := range reserved {
		if upperName == r || baseName == r {
			name = name + "_" // Append underscore if reserved
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
		return "", fmt.Errorf("%w: file path cannot be empty", ErrPathViolation)
	}
	if strings.Contains(filePath, "\x00") {
		return "", fmt.Errorf("%w: file path contains null byte", ErrPathViolation)
	}
	if filepath.IsAbs(filePath) {
		return "", fmt.Errorf("%w: input file path '%s' must be relative", ErrPathViolation, filePath)
	}

	absAllowedDir, err := filepath.Abs(allowedDir)
	if err != nil {
		// This is an internal configuration error, not a path violation
		return "", fmt.Errorf("could not get absolute path for allowed directory '%s': %w", allowedDir, err)
	}
	absAllowedDir = filepath.Clean(absAllowedDir)

	absCleanedPath := filepath.Join(absAllowedDir, filePath)
	absCleanedPath = filepath.Clean(absCleanedPath)

	prefixToCheck := absAllowedDir
	// Ensure the prefix ends with a separator unless it's the root "/"
	if prefixToCheck != string(filepath.Separator) && !strings.HasSuffix(prefixToCheck, string(filepath.Separator)) {
		prefixToCheck += string(filepath.Separator)
	}
	pathToCheck := absCleanedPath

	// Check if the cleaned path is the allowed directory itself or starts with the allowed directory + separator
	isOutside := pathToCheck != absAllowedDir && !strings.HasPrefix(pathToCheck, prefixToCheck)

	if isOutside {
		details := fmt.Sprintf("relative path '%s' resolves to '%s' which is outside the allowed directory '%s'", filePath, absCleanedPath, absAllowedDir)
		return "", fmt.Errorf("%w: %s", ErrPathViolation, details)
	}

	// Note: Allowing resolution *to* the root dir (if filePath is ".")
	// Rejecting if it resolves to root via other means (like ../..) was handled by prefix check

	return absCleanedPath, nil
}

// Note: validateArgumentsAgainstSpec remains in security_validation.go
// Note: Type helpers remain in security_helpers.go
