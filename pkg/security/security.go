// NeuroScript Version: 0.3.1
// File version: 0.0.7 // Correct ToolRegistry interface usage.
// nlines: 273
// risk_rating: HIGH
// filename: pkg/core/security.go
package core

import (
	"fmt"
	"path/filepath"
	"regexp" // Make sure regexp is imported
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/google/generative-ai-go/genai"
)

// SecurityLayer enforces security policies for LLM-initiated tool calls.
type SecurityLayer struct {
	allowlist    map[string]bool // Stores qualified tool names (TOOL.xxx)
	denylist     map[string]bool // Stores qualified tool names (TOOL.xxx)
	sandboxRoot  string          // Unexported field storing the validated path
	toolRegistry tool.ToolRegistry    // <<< CHANGED: Use the interface type directly
	logger       interfaces.Logger
}

// NewSecurityLayer creates a new security layer instance.
// <<< CHANGED: registry parameter is now ToolRegistry (interface type)
func NewSecurityLayer(allowlistTools []string, denylistSet map[string]bool, sandboxRoot string, registry tool.ToolRegistry, logger interfaces.Logger) *SecurityLayer {
	if logger == nil {
		// This should ideally return an error or use a default logger,
		// but panicking ensures it's caught during development.
		panic("SecurityLayer must have a valid logger")
	}
	allowlistMap := make(map[string]bool)
	for _, tool := range allowlistTools {
		qualifiedName := tool
		if !strings.HasPrefix(tool, "TOOL.") {
			qualifiedName = "TOOL." + tool
			logger.Warn("[SEC] Tool name in allowlist normalized", "original_name", tool, "qualified_name", qualifiedName, "detail", "Ensure config uses qualified names.")
		}
		allowlistMap[qualifiedName] = true
	}
	normalizedDenylist := make(map[string]bool)
	// deniedToolNamesOriginal := make([]string, 0, len(denylistSet)) // Keep if needed for detailed logging elsewhere
	for tool, denied := range denylistSet {
		// deniedToolNamesOriginal = append(deniedToolNamesOriginal, tool)
		qualifiedName := tool
		if !strings.HasPrefix(tool, "TOOL.") {
			qualifiedName = "TOOL." + tool
			logger.Warn("[SEC] Tool name in denylist normalized", "original_name", tool, "qualified_name", qualifiedName, "detail", "Ensure config uses qualified names.")
		}
		normalizedDenylist[qualifiedName] = denied
	}
	cleanSandboxRoot := "/" // Default to a restrictive root if issues occur
	if sandboxRoot == "" {
		logger.Warn("[SEC] Sandbox root not provided, using '/' as a fallback. This is highly restrictive.", "provided_root", sandboxRoot)
	} else {
		absSandboxRoot, err := filepath.Abs(sandboxRoot)
		if err != nil {
			logger.Error("[SEC] Failed to get absolute path for sandbox root. Using '/' as fallback.", "sandbox_root", sandboxRoot, "error", err)
		} else {
			cleanSandboxRoot = filepath.Clean(absSandboxRoot)
		}
	}

	sl := &SecurityLayer{
		allowlist:    allowlistMap,
		denylist:     normalizedDenylist,
		sandboxRoot:  cleanSandboxRoot,
		toolRegistry: registry, // <<< CHANGED: Direct assignment of the interface
		logger:       logger,
	}

	logger.Debug("[SEC] Initialized Security Layer.")
	allowlistedNames := make([]string, 0, len(sl.allowlist))
	for tool := range sl.allowlist {
		allowlistedNames = append(allowlistedNames, tool)
	}
	logger.Debug("[SEC] Allowlisted tools (normalized)", "tools", strings.Join(allowlistedNames, ", "))

	deniedNamesNormalized := make([]string, 0, len(sl.denylist))
	for tool := range sl.denylist {
		deniedNamesNormalized = append(deniedNamesNormalized, tool)
	}
	logger.Debug("[SEC] Denied tools (normalized)", "tools", strings.Join(deniedNamesNormalized, ", "))
	logger.Debug("[SEC] Sandbox Root Set To", "path", sl.sandboxRoot)

	if sl.toolRegistry == nil { // Check if the provided interface is nil
		logger.Warn("[SEC] SecurityLayer initialized with nil ToolRegistry. Tool validation/execution will likely fail.")
	}
	return sl
}

// SandboxRoot returns the configured root directory for sandboxing file operations.
func (sl *SecurityLayer) SandboxRoot() string { return sl.sandboxRoot }

// GetToolDeclarations generates the list of genai.Tool objects for allowlisted tools.
func (sl *SecurityLayer) GetToolDeclarations() ([]*genai.Tool, error) {
	if sl.toolRegistry == nil {
		sl.logger.Error("[SEC] Cannot get tool declarations: ToolRegistry is nil.")
		return nil, fmt.Errorf("%w: security layer tool registry is not initialized", lang.ErrConfiguration)
	}
	declarations := []*genai.Tool{}
	// Calls on sl.toolRegistry (which is now ToolRegistry interface type) should work correctly.
	allToolSpecs := sl.toolRegistry.ListTools()
	sl.logger.Debug("[SEC] Generating declarations from registered tool specs", "count", len(allToolSpecs))

	for _, spec := range allToolSpecs {
		baseName := spec.Name               // e.g., "ReadFile"
		qualifiedName := "TOOL." + baseName // e.g., "TOOL.ReadFile"

		// Check allowlist and denylist using the qualified name
		isAllowed := sl.allowlist[qualifiedName]
		isDenied := sl.denylist[qualifiedName]

		if isAllowed && !isDenied {
			sl.logger.Debug("[SEC] Generating declaration for allowlisted/not-denied tool", "qualified_name", qualifiedName)
			schema := &genai.Schema{
				Type:        genai.TypeObject,
				Properties:  map[string]*genai.Schema{},
				Required:    []string{},
				Description: spec.Description,
			}
			validSchema := true
			for _, argSpec := range spec.Args {
				genaiType, typeErr := argSpec.Type.ToGenaiType()
				if typeErr != nil {
					sl.logger.Error("[SEC] Failed to convert arg type for tool declaration", "arg_name", argSpec.Name, "arg_type", argSpec.Type, "tool_name", qualifiedName, "error", typeErr)
					validSchema = false
					break // Stop processing args for this tool if one is bad
				}
				schema.Properties[argSpec.Name] = &genai.Schema{Type: genaiType, Description: argSpec.Description}
				if argSpec.Required {
					schema.Required = append(schema.Required, argSpec.Name)
				}
			}

			if validSchema {
				declarations = append(declarations, &genai.Tool{
					FunctionDeclarations: []*genai.FunctionDeclaration{
						{
							Name:        qualifiedName,
							Description: spec.Description,
							Parameters:  schema,
						},
					},
				})
				sl.logger.Debug("[SEC] Added declaration", "qualified_name", qualifiedName)
			} else {
				sl.logger.Warn("[SEC] Skipping declaration due to invalid schema", "qualified_name", qualifiedName)
			}
		} else {
			sl.logger.Debug("[SEC] Tool not included in declarations (not allowed or explicitly denied)", "qualified_name", qualifiedName, "is_allowed", isAllowed, "is_denied", isDenied)
		}
	}
	sl.logger.Debug("[SEC] Generated tool declarations.", "count", len(declarations))
	return declarations, nil
}

// Executeinterfaces.ToolCall validates and executes a requested tool call.
func (sl *SecurityLayer) ExecuteToolCall(interpreter *neurogo.Interpreter, fc genai.FunctionCall) (genai.Part, error) {
	qualifiedToolName := fc.Name
	rawArgs := fc.Args

	sl.logger.Debug("[SEC Executeinterfaces.ToolCall] Received request", "tool_name", qualifiedToolName)

	if sl.toolRegistry == nil {
		err := fmt.Errorf("%w: tool registry is not available in security layer", lang.ErrInternalSecurity)
		sl.logger.Error("[SEC Executeinterfaces.ToolCall] Critical internal error", "error", err, "tool_name", qualifiedToolName)
		return CreateErrorFunctionResultPart(qualifiedToolName, err), err
	}

	validatedArgsMap, validationErr := sl.ValidateToolCall(qualifiedToolName, rawArgs)
	if validationErr != nil {
		sl.logger.Warn("[SEC Executeinterfaces.ToolCall] Validation failed for tool call", "tool_name", qualifiedToolName, "error", validationErr)
		return CreateErrorFunctionResultPart(qualifiedToolName, validationErr), validationErr
	}

	sl.logger.Debug("[SEC Executeinterfaces.ToolCall] Tool call validated, proceeding to execution", "tool_name", qualifiedToolName)

	baseToolName := strings.TrimPrefix(qualifiedToolName, "TOOL.")
	// Calls on sl.toolRegistry (which is now ToolRegistry interface type) should work correctly.
	toolImpl, found := sl.toolRegistry.GetTool(baseToolName)
	if !found {
		// This case should ideally be caught by Validateinterfaces.ToolCall if the tool isn't in the registry at all.
		// However, if it was removed between validation and execution, or if Validateinterfaces.ToolCall's check isn't exhaustive.
		err := fmt.Errorf("%w: tool implementation '%s' not found post-validation (tool name: %s)", lang.ErrInternalSecurity, baseToolName, qualifiedToolName)
		sl.logger.Error("[SEC Executeinterfaces.ToolCall] Critical internal error", "error", err)
		return CreateErrorFunctionResultPart(qualifiedToolName, err), err
	}

	// Reconstruct ordered arguments based on ToolSpec
	orderedArgs := make([]interface{}, len(toolImpl.Spec.Args))
	conversionOk := true
	for i, argSpec := range toolImpl.Spec.Args {
		val, exists := validatedArgsMap[argSpec.Name]
		if !exists {
			if argSpec.Required {
				// This should also be caught by Validateinterfaces.ToolCall, but as a safeguard:
				err := fmt.Errorf("%w: required arg '%s' missing post-validation for tool '%s'", lang.ErrInternalSecurity, argSpec.Name, qualifiedToolName)
				sl.logger.Error("[SEC Executeinterfaces.ToolCall] Critical internal error: missing required arg", "error", err)
				conversionOk = false // Mark as failed
				break
			}
			orderedArgs[i] = nil // Optional arg not provided
		} else {
			orderedArgs[i] = val
		}
	}

	if !conversionOk {
		// If argument reconstruction failed (e.g. required arg missing after validation, which is unlikely but possible if validation logic has holes)
		err := fmt.Errorf("%w: failed to reconstruct ordered args post-validation for tool '%s'", lang.ErrInternalSecurity, qualifiedToolName)
		sl.logger.Error("[SEC Executeinterfaces.ToolCall] Critical internal error: arg reconstruction failed", "error", err)
		return CreateErrorFunctionResultPart(qualifiedToolName, err), err
	}

	sl.logger.Debug("[SEC Executeinterfaces.ToolCall] Executing tool function", "qualified_name", qualifiedToolName, "base_name", baseToolName)
	resultValue, execErr := toolImpl.Func(interpreter, orderedArgs)
	if execErr != nil {
		sl.logger.Error("[SEC Executeinterfaces.ToolCall] Tool execution failed", "tool_name", qualifiedToolName, "error", execErr)
		// It's important that execErr is a well-formed error, ideally a RuntimeError
		return CreateErrorFunctionResultPart(qualifiedToolName, execErr), execErr
	}

	sl.logger.Debug("[SEC Executeinterfaces.ToolCall] Tool execution successful", "tool_name", qualifiedToolName)
	return CreateSuccessFunctionResultPart(qualifiedToolName, resultValue, sl.logger), nil
}

// Validateinterfaces.ToolCall checks denylist, allowlist, high-risk status, and delegates argument validation.
func (sl *SecurityLayer) ValidateToolCall(qualifiedToolName string, rawArgs map[string]interface{}) (map[string]interface{}, error) {
	sl.logger.Debug("[SEC Validateinterfaces.ToolCall] Validating request", "tool_name", qualifiedToolName)

	if sl.denylist[qualifiedToolName] {
		sl.logger.Warn("[SEC Validateinterfaces.ToolCall] DENIED (Denylist)", "tool_name", qualifiedToolName)
		return nil, fmt.Errorf("tool %q denied by denylist: %w", qualifiedToolName, lang.ErrToolDenied)
	}
	if !sl.allowlist[qualifiedToolName] {
		sl.logger.Warn("[SEC Validateinterfaces.ToolCall] DENIED (Not Allowlisted)", "tool_name", qualifiedToolName)
		return nil, fmt.Errorf("tool %q not allowed by allowlist: %w", qualifiedToolName, lang.ErrToolNotAllowed)
	}

	// Example of a hardcoded policy, can be expanded
	if qualifiedToolName == "TOOL.ExecuteCommand" { // Assuming ExecuteCommand is the qualified name
		sl.logger.Warn("[SEC Validateinterfaces.ToolCall] DENIED (Blocked by Policy)", "tool_name", qualifiedToolName, "reason", "ExecuteCommand is disabled")
		return nil, fmt.Errorf("tool %q blocked by security policy: %w", qualifiedToolName, lang.ErrToolBlocked)
	}

	if sl.toolRegistry == nil {
		sl.logger.Error("[SEC Validateinterfaces.ToolCall] ToolRegistry not available during validation.", "tool_name", qualifiedToolName)
		return nil, fmt.Errorf("%w: tool registry unavailable for validation of '%s'", lang.ErrInternalSecurity, qualifiedToolName)
	}

	baseToolName := strings.TrimPrefix(qualifiedToolName, "TOOL.")
	// Calls on sl.toolRegistry (which is now ToolRegistry interface type) should work correctly.
	toolImpl, found := sl.toolRegistry.GetTool(baseToolName)
	if !found {
		sl.logger.Error("[SEC Validateinterfaces.ToolCall] Allowlisted tool implementation not found in registry.", "qualified_name", qualifiedToolName, "base_name", baseToolName)
		return nil, fmt.Errorf("%w: allowlisted tool '%s' (base: '%s') implementation not found", lang.ErrInternalSecurity, qualifiedToolName, baseToolName)
	}

	toolSpec := toolImpl.Spec
	sl.logger.Debug("[SEC Validateinterfaces.ToolCall] Found tool spec for validation", "tool_name", qualifiedToolName, "base_name", baseToolName, "args_count_in_spec", len(toolSpec.Args))

	// Delegate to the argument validation logic (assumed to be in security_validation.go or similar)
	validatedArgs, validationErr := sl.validateArgumentsAgainstSpec(toolSpec, rawArgs)
	if validationErr != nil {
		sl.logger.Warn("[SEC Validateinterfaces.ToolCall] DENIED (Argument Validation Failed)", "tool_name", qualifiedToolName, "error", validationErr)
		return nil, validationErr // Return the specific validation error
	}

	sl.logger.Debug("[SEC Validateinterfaces.ToolCall] Arguments validated successfully.", "tool_name", qualifiedToolName)
	return validatedArgs, nil
}

// SanitizeFilename (Implementation unchanged from previous correction)
func SanitizeFilename(name string) string {
	if name == "" {
		return "default_sanitized_name"
	}
	if strings.Contains(name, "\x00") { // Check for null bytes
		return "invalid_null_byte_name" // Or handle error appropriately
	}
	name = strings.ReplaceAll(name, " ", "_") // Replace spaces with underscores
	replacer := strings.NewReplacer(          // Define characters to be replaced
		"/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	name = replacer.Replace(name)
	removeChars := regexp.MustCompile(`[^a-zA-Z0-9._-]`) // Allow alphanumeric, dot, underscore, hyphen
	name = removeChars.ReplaceAllString(name, "")
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_") // Replace multiple underscores with one
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-") // Replace multiple hyphens with one
	name = strings.Trim(name, "._-")                               // Trim leading/trailing unsafe chars
	for strings.Contains(name, "..") {                             // Avoid ".." for path traversal
		name = strings.ReplaceAll(name, "..", "_") // Replace with underscore
	}
	name = strings.Trim(name, "._-")                               // Trim again after ".." replacement
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_") // Consolidate again
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-") // Consolidate again
	const maxLength = 100                                          // Max filename length
	if len(name) > maxLength {
		name = name[:maxLength]
		name = strings.TrimRight(name, "._-") // Ensure it doesn't end with unsafe char after truncation
	}
	if name == "" { // If everything was stripped
		name = "default_sanitized_name"
	}
	reserved := []string{ // Common reserved names (Windows)
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	upperName := strings.ToUpper(name)
	baseName := upperName
	if dotIndex := strings.LastIndex(upperName, "."); dotIndex != -1 { // Check basename without extension
		baseName = upperName[:dotIndex]
	}
	for _, r := range reserved {
		if upperName == r || baseName == r {
			name = name + "_" // Append underscore if it's a reserved name
			break
		}
	}
	return name
}
