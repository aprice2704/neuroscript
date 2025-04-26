// filename: pkg/core/security.go
package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp" // Make sure regexp is imported
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/google/generative-ai-go/genai"
)

// // Sentinel errors for security violations
// var (
// 	ErrToolDenied         = errors.New("tool denied by policy")
// 	ErrToolNotAllowed     = errors.New("tool not allowlisted")
// 	ErrToolBlocked        = errors.New("tool blocked by specific policy")
// 	ErrPathViolation      = errors.New("path violation")
// 	ErrNullByteInArgument = errors.New("argument contains null byte")
// 	ErrInternalSecurity   = errors.New("internal security configuration error")
// )

// SecurityLayer enforces security policies for LLM-initiated tool calls.
type SecurityLayer struct {
	allowlist    map[string]bool
	denylist     map[string]bool
	sandboxRoot  string // Unexported field storing the validated path
	toolRegistry *ToolRegistry
	logger       interfaces.Logger
}

// NewSecurityLayer creates a new security layer instance.
func NewSecurityLayer(allowlistTools []string, denylistSet map[string]bool, sandboxRoot string, registry *ToolRegistry, logger interfaces.Logger) *SecurityLayer {
	allowlistMap := make(map[string]bool)
	for _, tool := range allowlistTools {
		// Ensure the key includes the TOOL. prefix for consistency
		if !strings.HasPrefix(tool, "TOOL.") {
			// This might happen if the list file contains base names.
			// Log a warning or normalize? Let's assume list file is correct for now.
			// logger.Warn("SEC] Tool name '%s' in allowlist might be missing 'TOOL.' prefix.", tool)
		}
		allowlistMap[tool] = true
	}

	// Ensure the provided sandboxRoot is clean and absolute
	cleanSandboxRoot := "/" // Default to root if error occurs? Maybe better to error out?
	absSandboxRoot, err := filepath.Abs(sandboxRoot)
	if err != nil {
		logger.Error("SEC] Failed to get absolute path for sandbox root %q: %v. Sandboxing may not function correctly.", sandboxRoot, err)
		// Use the original cleaned path as fallback if Abs fails
		cleanSandboxRoot = filepath.Clean(sandboxRoot)
	} else {
		cleanSandboxRoot = filepath.Clean(absSandboxRoot)
	}

	// Ensure logger is not nil
	if logger == nil {
		panic("Security must have valid logger")
	}

	logger.Info("[SEC] Initialized Security Layer.")
	logger.Info("[SEC] Allowlisted tools (initial): %v", allowlistTools) // This might log qualified names depending on input list
	deniedToolNames := make([]string, 0, len(denylistSet))
	for tool := range denylistSet {
		deniedToolNames = append(deniedToolNames, tool)
	}
	logger.Info("[SEC] Denied tools: %v", deniedToolNames)
	logger.Info("[SEC] Sandbox Root Set To: %s", cleanSandboxRoot)

	if registry == nil {
		logger.Info("[WARN SEC] SecurityLayer initialized with nil ToolRegistry. Argument validation/execution will fail.")
	}

	return &SecurityLayer{
		allowlist:    allowlistMap,
		denylist:     denylistSet,
		sandboxRoot:  cleanSandboxRoot, // Store the cleaned, potentially absolute path
		toolRegistry: registry,
		logger:       logger,
	}
}

// +++ ADDED: SandboxRoot Getter +++
// SandboxRoot returns the configured root directory for sandboxing file operations.
func (sl *SecurityLayer) SandboxRoot() string {
	return sl.sandboxRoot
}

// --- END ADDED ---

// GetToolDeclarations generates the list of genai.Tool objects for allowlisted tools.
func (sl *SecurityLayer) GetToolDeclarations() ([]*genai.Tool, error) {
	if sl.toolRegistry == nil {
		sl.logger.Error(" SEC] Cannot get tool declarations: ToolRegistry is nil.")
		return nil, errors.New("security layer tool registry is not initialized")
	}

	declarations := []*genai.Tool{}
	allTools := sl.toolRegistry.GetAllTools() // Gets map[string]ToolImplementation (base name -> impl)

	sl.logger.Warn("Generating declarations for %d registered tools...", len(allTools))

	for baseName, impl := range allTools {
		qualifiedName := "TOOL." + baseName // Construct qualified name
		// Check allow/deny lists using the qualified name
		if sl.allowlist[qualifiedName] && !sl.denylist[qualifiedName] {
			sl.logger.Warn("Generating declaration for allowlisted tool: %s", qualifiedName)
			schema := &genai.Schema{
				Type:        genai.TypeObject,
				Properties:  map[string]*genai.Schema{},
				Required:    []string{},
				Description: impl.Spec.Description,
			}

			validSchema := true // Flag to track if schema generation succeeds
			for _, argSpec := range impl.Spec.Args {
				genaiType, typeErr := argSpec.Type.ToGenaiType()
				if typeErr != nil {
					sl.logger.Error("SEC] Failed to convert type for arg '%s' in tool '%s': %v. Skipping tool declaration.", argSpec.Name, qualifiedName, typeErr)
					validSchema = false
					break // Stop processing args for this tool
				}
				schema.Properties[argSpec.Name] = &genai.Schema{
					Type:        genaiType,
					Description: argSpec.Description,
				}
				if argSpec.Required {
					schema.Required = append(schema.Required, argSpec.Name)
				}
			}

			if validSchema {
				declarations = append(declarations, &genai.Tool{
					FunctionDeclarations: []*genai.FunctionDeclaration{{
						Name:        qualifiedName, // Use qualified name for LLM
						Description: impl.Spec.Description,
						Parameters:  schema,
					}},
				})
				sl.logger.Warn("Added declaration for: %s", qualifiedName)
			}
		} else {
			// sl.logger.Warn("Skipping declaration for tool '%s' (Base: %s) (not allowlisted or is denied).", qualifiedName, baseName)
		}
	}
	sl.logger.Warn("Generated %d total tool declarations.", len(declarations))
	return declarations, nil
}

// ExecuteToolCall validates and executes a requested tool call.
// Expects qualifiedToolName like "TOOL.ReadFile".
func (sl *SecurityLayer) ExecuteToolCall(interpreter *Interpreter, fc genai.FunctionCall) (genai.Part, error) {
	qualifiedToolName := fc.Name
	rawArgs := fc.Args

	if sl.toolRegistry == nil {
		err := errors.New("internal security error: tool registry is not available")
		sl.logger.Error("SEC ExecuteToolCall] %v", err)
		return CreateErrorFunctionResultPart(qualifiedToolName, err), err
	}

	// Validation Phase (expects qualified name)
	validatedArgsMap, validationErr := sl.ValidateToolCall(qualifiedToolName, rawArgs)
	if validationErr != nil {
		sl.logger.Debug("ExecuteToolCall] Validation failed for tool '%s': %v", qualifiedToolName, validationErr)
		return CreateErrorFunctionResultPart(qualifiedToolName, validationErr), validationErr
	}

	// Execution Phase
	baseToolName := strings.TrimPrefix(qualifiedToolName, "TOOL.")
	toolImpl, found := sl.toolRegistry.GetTool(baseToolName) // Use base name for registry lookup
	if !found {
		err := fmt.Errorf("tool implementation '%s' not found in registry despite passing validation", baseToolName)
		sl.logger.Error("SEC ExecuteToolCall] %v", err)
		return CreateErrorFunctionResultPart(qualifiedToolName, err), err
	}

	// Convert validated map back to []interface{}
	orderedArgs := make([]interface{}, len(toolImpl.Spec.Args))
	for i, argSpec := range toolImpl.Spec.Args {
		val, exists := validatedArgsMap[argSpec.Name]
		if !exists {
			if !argSpec.Required {
				orderedArgs[i] = nil
			} else {
				err := fmt.Errorf("internal error: required arg '%s' missing post-validation for tool '%s'", argSpec.Name, qualifiedToolName)
				sl.logger.Error("SEC ExecuteToolCall] %v", err)
				return CreateErrorFunctionResultPart(qualifiedToolName, err), err
			}
		} else {
			orderedArgs[i] = val
		}
	}

	sl.logger.Debug("ExecuteToolCall] Executing tool '%s' (Base: %s)...", qualifiedToolName, baseToolName)
	resultValue, execErr := toolImpl.Func(interpreter, orderedArgs)
	if execErr != nil {
		sl.logger.Error("SEC ExecuteToolCall] Execution failed for tool '%s': %v", qualifiedToolName, execErr)
		return CreateErrorFunctionResultPart(qualifiedToolName, execErr), execErr
	}

	// Format Success Response
	sl.logger.Debug("ExecuteToolCall] Execution successful for tool '%s'.", qualifiedToolName)
	responseMap := make(map[string]interface{})
	switch v := resultValue.(type) {
	case map[string]interface{}:
		responseMap = v
	case []interface{}:
		responseMap["result_list"] = v
	case string, int, int64, float32, float64, bool:
		responseMap["result"] = v
	case nil:
		responseMap["status"] = "success"
	default:
		responseMap["result"] = fmt.Sprintf("%v", v)
		sl.logger.Warn("SEC ExecuteToolCall] Tool '%s' returned unexpected type %T", qualifiedToolName, v)
	}

	sl.logger.Debug("ExecuteToolCall] Formatted response for '%s': %v", qualifiedToolName, responseMap)
	return genai.FunctionResponse{
		Name:     qualifiedToolName, // Use qualified name in response back to LLM
		Response: responseMap,
	}, nil
}

// ValidateToolCall checks denylist, allowlist, high-risk status, and delegates argument validation.
// Expects the qualified tool name (e.g., "TOOL.ReadFile").
func (sl *SecurityLayer) ValidateToolCall(qualifiedToolName string, rawArgs map[string]interface{}) (map[string]interface{}, error) {
	sl.logger.Warn("Validating request for tool: %s with raw args: %v", qualifiedToolName, rawArgs)

	// 1. Denylist Check
	if sl.denylist[qualifiedToolName] {
		sl.logger.Warn("DENIED: Tool %q is explicitly denied by denylist.", qualifiedToolName)
		return nil, fmt.Errorf("tool %q denied: %w", qualifiedToolName, ErrToolDenied)
	}
	sl.logger.Warn("Tool '%s' is not denied.", qualifiedToolName)

	// 2. Allowlist Check
	if !sl.allowlist[qualifiedToolName] {
		sl.logger.Warn("DENIED: Tool %q is not in the allowlist for LLM execution.", qualifiedToolName)
		return nil, fmt.Errorf("tool %q not allowed: %w", qualifiedToolName, ErrToolNotAllowed)
	}
	sl.logger.Warn("Tool '%s' is allowlisted.", qualifiedToolName)

	// 3. High-Risk Tool Check (using qualified name)
	if qualifiedToolName == "TOOL.ExecuteCommand" {
		sl.logger.Warn("DENIED: Tool %q is blocked by policy.", qualifiedToolName)
		return nil, fmt.Errorf("tool %q blocked: %w", qualifiedToolName, ErrToolBlocked)
	}

	// 4. Get Tool Specification using Base Name
	if sl.toolRegistry == nil {
		sl.logger.Error("SEC] ToolRegistry not available during validation for %q.", qualifiedToolName)
		return nil, fmt.Errorf("tool registry unavailable for %q: %w", qualifiedToolName, ErrInternalSecurity)
	}
	baseToolName := strings.TrimPrefix(qualifiedToolName, "TOOL.")
	toolImpl, found := sl.toolRegistry.GetTool(baseToolName)
	if !found {
		sl.logger.Error("SEC] Allowlisted tool %q (base: %s) not found in registry.", qualifiedToolName, baseToolName)
		return nil, fmt.Errorf("tool %q implementation not found: %w", qualifiedToolName, ErrInternalSecurity)
	}
	toolSpec := toolImpl.Spec
	sl.logger.Warn("Validating args for '%s' against spec (Base: %s, Spec Args: %d)", qualifiedToolName, baseToolName, len(toolSpec.Args))

	// 5. Delegate Argument Validation (Ensure security_validation.go exists and is correct)
	validatedArgs, validationErr := sl.validateArgumentsAgainstSpec(toolSpec, rawArgs)
	if validationErr != nil {
		sl.logger.Warn("DENIED (Argument Validation): Tool %q, Error: %v", qualifiedToolName, validationErr)
		return nil, validationErr
	}

	sl.logger.Warn("All arguments for '%s' validated successfully. Validated Args: %v", qualifiedToolName, validatedArgs)
	return validatedArgs, nil
}

// SanitizeFilename (Implementation unchanged from user provided version)
func SanitizeFilename(name string) string {
	// ... (implementation unchanged) ...
	name = strings.ReplaceAll(name, " ", "_")
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	name = replacer.Replace(name)
	removeChars := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	name = removeChars.ReplaceAllString(name, "")
	name = strings.Trim(name, "._-")
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")
	for strings.Contains(name, "..") {
		name = strings.ReplaceAll(name, "..", "_")
	} // Basic protection
	name = strings.Trim(name, "._-") // Trim again
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")
	const maxLength = 100 // Limit length
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
	} // Append underscore for reserved names
	return name
}

// SecureFilePath (Implementation unchanged from user provided version)
func SecureFilePath(filePath, allowedDir string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty: %w", ErrPathViolation)
	}
	if strings.Contains(filePath, "\x00") {
		return "", fmt.Errorf("file path contains null byte: %w", ErrNullByteInArgument)
	}
	if filepath.IsAbs(filePath) {
		return "", fmt.Errorf("input file path %q must be relative, not absolute: %w", filePath, ErrPathViolation)
	}

	absAllowedDir, err := filepath.Abs(allowedDir)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path for allowed directory %q: %w", allowedDir, err)
	}
	absAllowedDir = filepath.Clean(absAllowedDir)

	absCleanedPath := filepath.Join(absAllowedDir, filePath)
	absCleanedPath = filepath.Clean(absCleanedPath)

	prefixToCheck := absAllowedDir
	if prefixToCheck != string(filepath.Separator) && !strings.HasSuffix(prefixToCheck, string(filepath.Separator)) {
		prefixToCheck += string(filepath.Separator)
	}

	if absCleanedPath != absAllowedDir && !strings.HasPrefix(absCleanedPath, prefixToCheck) {
		details := fmt.Sprintf("relative path %q resolves to %q which is outside the allowed directory %q", filePath, absCleanedPath, absAllowedDir)
		return "", fmt.Errorf("%s: %w", details, ErrPathViolation)
	}
	return absCleanedPath, nil
}
