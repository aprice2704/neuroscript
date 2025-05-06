// NeuroScript Version: 0.3.1
// File version: 0.0.6 // Fix variable names inside SecureFilePath function.
// filename: pkg/core/security.go
package core

import (
	"fmt"
	"path/filepath"
	"regexp" // Make sure regexp is imported
	"strings"

	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/generative-ai-go/genai"
)

// SecurityLayer enforces security policies for LLM-initiated tool calls.
type SecurityLayer struct {
	allowlist    map[string]bool // Stores qualified tool names (TOOL.xxx)
	denylist     map[string]bool // Stores qualified tool names (TOOL.xxx)
	sandboxRoot  string          // Unexported field storing the validated path
	toolRegistry *ToolRegistry
	logger       logging.Logger
}

// NewSecurityLayer creates a new security layer instance.
func NewSecurityLayer(allowlistTools []string, denylistSet map[string]bool, sandboxRoot string, registry *ToolRegistry, logger logging.Logger) *SecurityLayer {
	// ... (implementation unchanged from previous correction) ...
	if logger == nil {
		panic("Security must have valid logger")
	}
	allowlistMap := make(map[string]bool)
	for _, tool := range allowlistTools {
		qualifiedName := tool
		if !strings.HasPrefix(tool, "TOOL.") {
			qualifiedName = "TOOL." + tool
			logger.Warn("SEC] Tool name '%s' in allowlist normalized to '%s'. Ensure config uses qualified names.", tool, qualifiedName)
		}
		allowlistMap[qualifiedName] = true
	}
	normalizedDenylist := make(map[string]bool)
	deniedToolNamesOriginal := make([]string, 0, len(denylistSet))
	for tool, denied := range denylistSet {
		deniedToolNamesOriginal = append(deniedToolNamesOriginal, tool)
		qualifiedName := tool
		if !strings.HasPrefix(tool, "TOOL.") {
			qualifiedName = "TOOL." + tool
			logger.Warn("SEC] Tool name '%s' in denylist normalized to '%s'. Ensure config uses qualified names.", tool, qualifiedName)
		}
		normalizedDenylist[qualifiedName] = denied
	}
	cleanSandboxRoot := "/"
	absSandboxRoot, err := filepath.Abs(sandboxRoot)
	if err != nil {
		logger.Error("SEC] Failed to get absolute path for sandbox root %q: %v. Using '/' as fallback.", sandboxRoot, err)
	} else {
		cleanSandboxRoot = filepath.Clean(absSandboxRoot)
	}
	logger.Info("[SEC] Initialized Security Layer.")
	allowlistedNames := make([]string, 0, len(allowlistMap))
	for tool := range allowlistMap {
		allowlistedNames = append(allowlistedNames, tool)
	}
	logger.Info("[SEC] Allowlisted tools (normalized): %v", allowlistedNames)
	deniedNamesNormalized := make([]string, 0, len(normalizedDenylist))
	for tool := range normalizedDenylist {
		deniedNamesNormalized = append(deniedNamesNormalized, tool)
	}
	logger.Info("[SEC] Denied tools (normalized): %v", deniedNamesNormalized)
	logger.Info("[SEC] Sandbox Root Set To: %s", cleanSandboxRoot)
	if registry == nil {
		logger.Warn("[SEC] SecurityLayer initialized with nil ToolRegistry. Validation/execution will fail.")
	}
	return &SecurityLayer{allowlist: allowlistMap, denylist: normalizedDenylist, sandboxRoot: cleanSandboxRoot, toolRegistry: registry, logger: logger}
}

// SandboxRoot returns the configured root directory for sandboxing file operations.
func (sl *SecurityLayer) SandboxRoot() string { return sl.sandboxRoot }

// GetToolDeclarations generates the list of genai.Tool objects for allowlisted tools.
func (sl *SecurityLayer) GetToolDeclarations() ([]*genai.Tool, error) {
	// ... (implementation unchanged from previous correction) ...
	if sl.toolRegistry == nil {
		sl.logger.Error("SEC] Cannot get tool declarations: ToolRegistry is nil.")
		return nil, fmt.Errorf("%w: security layer tool registry is not initialized", ErrConfiguration)
	}
	declarations := []*genai.Tool{}
	allToolSpecs := sl.toolRegistry.ListTools()
	sl.logger.Debug("Generating declarations for %d registered tool specs...", len(allToolSpecs))
	for _, spec := range allToolSpecs {
		baseName := spec.Name
		qualifiedName := "TOOL." + baseName
		isAllowed := sl.allowlist[qualifiedName]
		isDenied := sl.denylist[qualifiedName]
		if isAllowed && !isDenied {
			sl.logger.Debug("Generating declaration for allowlisted/not-denied tool", "qualifiedName", qualifiedName)
			schema := &genai.Schema{Type: genai.TypeObject, Properties: map[string]*genai.Schema{}, Required: []string{}, Description: spec.Description}
			validSchema := true
			for _, argSpec := range spec.Args {
				genaiType, typeErr := argSpec.Type.ToGenaiType()
				if typeErr != nil {
					sl.logger.Error("SEC] Failed to convert type for arg in tool declaration", "arg", argSpec.Name, "tool", qualifiedName, "error", typeErr)
					validSchema = false
					break
				}
				schema.Properties[argSpec.Name] = &genai.Schema{Type: genaiType, Description: argSpec.Description}
				if argSpec.Required {
					schema.Required = append(schema.Required, argSpec.Name)
				}
			}
			if validSchema {
				declarations = append(declarations, &genai.Tool{FunctionDeclarations: []*genai.FunctionDeclaration{{Name: qualifiedName, Description: spec.Description, Parameters: schema}}})
				sl.logger.Debug("Added declaration for", "qualifiedName", qualifiedName)
			} else {
				sl.logger.Warn("Skipping declaration due to invalid schema", "qualifiedName", qualifiedName)
			}
		}
	}
	sl.logger.Info("Generated %d total tool declarations.", len(declarations))
	return declarations, nil
}

// ExecuteToolCall validates and executes a requested tool call.
func (sl *SecurityLayer) ExecuteToolCall(interpreter *Interpreter, fc genai.FunctionCall) (genai.Part, error) {
	// ... (implementation unchanged from previous correction - returns nil error on success) ...
	qualifiedToolName := fc.Name
	rawArgs := fc.Args
	if sl.toolRegistry == nil {
		err := fmt.Errorf("%w: tool registry is not available", ErrInternalSecurity)
		sl.logger.Error("SEC ExecuteToolCall] %v", err)
		return CreateErrorFunctionResultPart(qualifiedToolName, err), err
	}
	validatedArgsMap, validationErr := sl.ValidateToolCall(qualifiedToolName, rawArgs)
	if validationErr != nil {
		sl.logger.Warn("SEC ExecuteToolCall] Validation failed", "tool", qualifiedToolName, "error", validationErr)
		return CreateErrorFunctionResultPart(qualifiedToolName, validationErr), validationErr
	}
	baseToolName := strings.TrimPrefix(qualifiedToolName, "TOOL.")
	toolImpl, found := sl.toolRegistry.GetTool(baseToolName)
	if !found {
		err := fmt.Errorf("%w: tool implementation '%s' not found post-validation", ErrInternalSecurity, baseToolName)
		sl.logger.Error("SEC ExecuteToolCall] %v", err)
		return CreateErrorFunctionResultPart(qualifiedToolName, err), err
	}
	orderedArgs := make([]interface{}, len(toolImpl.Spec.Args))
	conversionOk := true
	for i, argSpec := range toolImpl.Spec.Args {
		val, exists := validatedArgsMap[argSpec.Name]
		if !exists {
			if !argSpec.Required {
				orderedArgs[i] = nil
			} else {
				err := fmt.Errorf("%w: required arg '%s' missing post-validation for tool '%s'", ErrInternalSecurity, argSpec.Name, qualifiedToolName)
				sl.logger.Error("SEC ExecuteToolCall] %v", err)
				conversionOk = false
				break
			}
		} else {
			orderedArgs[i] = val
		}
	}
	if !conversionOk {
		err := fmt.Errorf("%w: failed to reconstruct ordered args post-validation for tool '%s'", ErrInternalSecurity, qualifiedToolName)
		return CreateErrorFunctionResultPart(qualifiedToolName, err), err
	}
	sl.logger.Debug("SEC ExecuteToolCall] Executing tool", "qualifiedName", qualifiedToolName, "baseName", baseToolName)
	resultValue, execErr := toolImpl.Func(interpreter, orderedArgs)
	if execErr != nil {
		sl.logger.Error("SEC ExecuteToolCall] Execution failed", "tool", qualifiedToolName, "error", execErr)
		return CreateErrorFunctionResultPart(qualifiedToolName, execErr), execErr
	}
	sl.logger.Debug("SEC ExecuteToolCall] Execution successful", "tool", qualifiedToolName)
	return CreateSuccessFunctionResultPart(qualifiedToolName, resultValue, sl.logger), nil // Return nil error on success
}

// ValidateToolCall checks denylist, allowlist, high-risk status, and delegates argument validation.
func (sl *SecurityLayer) ValidateToolCall(qualifiedToolName string, rawArgs map[string]interface{}) (map[string]interface{}, error) {
	// ... (implementation unchanged from previous correction) ...
	sl.logger.Debug("SEC Validating request", "tool", qualifiedToolName /*, "raw_args", rawArgs */)
	if sl.denylist[qualifiedToolName] {
		sl.logger.Warn("SEC DENIED (Denylist)", "tool", qualifiedToolName)
		return nil, fmt.Errorf("tool %q denied: %w", qualifiedToolName, ErrToolDenied)
	}
	if !sl.allowlist[qualifiedToolName] {
		sl.logger.Warn("SEC DENIED (Not Allowlisted)", "tool", qualifiedToolName)
		return nil, fmt.Errorf("tool %q not allowed: %w", qualifiedToolName, ErrToolNotAllowed)
	}
	if qualifiedToolName == "TOOL.ExecuteCommand" {
		sl.logger.Warn("SEC DENIED (Blocked Policy)", "tool", qualifiedToolName)
		return nil, fmt.Errorf("tool %q blocked: %w", qualifiedToolName, ErrToolBlocked)
	}
	if sl.toolRegistry == nil {
		sl.logger.Error("SEC] ToolRegistry not available during validation.", "tool", qualifiedToolName)
		return nil, fmt.Errorf("%w: tool registry unavailable for %q", ErrInternalSecurity, qualifiedToolName)
	}
	baseToolName := strings.TrimPrefix(qualifiedToolName, "TOOL.")
	toolImpl, found := sl.toolRegistry.GetTool(baseToolName)
	if !found {
		sl.logger.Error("SEC] Allowlisted tool implementation not found in registry.", "qualifiedName", qualifiedToolName, "baseName", baseToolName)
		return nil, fmt.Errorf("%w: allowlisted tool '%s' implementation not found", ErrInternalSecurity, qualifiedToolName)
	}
	toolSpec := toolImpl.Spec
	sl.logger.Debug("SEC] Found tool spec for validation", "tool", qualifiedToolName, "baseName", baseToolName, "specArgsCount", len(toolSpec.Args))
	validatedArgs, validationErr := sl.validateArgumentsAgainstSpec(toolSpec, rawArgs)
	if validationErr != nil {
		sl.logger.Warn("SEC DENIED (Argument Validation)", "tool", qualifiedToolName, "error", validationErr)
		return nil, validationErr
	}
	sl.logger.Debug("SEC] Arguments validated successfully.", "tool", qualifiedToolName /*, "validatedArgs", validatedArgs */)
	return validatedArgs, nil
}

// SanitizeFilename (Implementation unchanged from previous correction)
func SanitizeFilename(name string) string {
	// ... (implementation unchanged) ...
	if name == "" {
		return "default_sanitized_name"
	}
	if strings.Contains(name, "\x00") {
		return "invalid_null_byte_name"
	}
	name = strings.ReplaceAll(name, " ", "_")
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	name = replacer.Replace(name)
	removeChars := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	name = removeChars.ReplaceAllString(name, "")
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")
	name = strings.Trim(name, "._-")
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

// SecureFilePath validates a path against the sandbox root.
// *** FIXED variable names inside this function ***
func SecureFilePath(filePath, allowedDir string) (string, error) {
	// Basic validation
	if filePath == "" { // Use filePath (parameter name)
		return "", fmt.Errorf("file path cannot be empty: %w", ErrPathViolation)
	}
	if strings.Contains(filePath, "\x00") { // Use filePath
		return "", fmt.Errorf("file path contains null byte: %w", ErrNullByteInArgument)
	}
	if filepath.IsAbs(filePath) { // Use filePath
		return "", fmt.Errorf("input file path %q must be relative, not absolute: %w", filePath, ErrPathViolation)
	}

	// Resolve allowed directory
	absAllowedDir, err := filepath.Abs(allowedDir) // Use allowedDir (parameter name)
	if err != nil {
		// Use allowedDir in error message
		return "", fmt.Errorf("%w: could not get absolute path for allowed directory %q: %v", ErrInternalSecurity, allowedDir, err)
	}
	absAllowedDir = filepath.Clean(absAllowedDir)

	// Join allowed directory with input path
	absCleanedPath := filepath.Join(absAllowedDir, filePath) // Use allowedDir and filePath
	absCleanedPath = filepath.Clean(absCleanedPath)

	// Check containment
	prefixToCheck := absAllowedDir
	if prefixToCheck != string(filepath.Separator) && !strings.HasSuffix(prefixToCheck, string(filepath.Separator)) {
		prefixToCheck += string(filepath.Separator)
	}

	if absCleanedPath != absAllowedDir && !strings.HasPrefix(absCleanedPath, prefixToCheck) {
		// Use filePath and absAllowedDir in error message
		details := fmt.Sprintf("relative path %q resolves to %q which is outside the allowed directory %q", filePath, absCleanedPath, absAllowedDir)
		return "", fmt.Errorf("%s: %w", details, ErrPathViolation)
	}

	return absCleanedPath, nil
}
