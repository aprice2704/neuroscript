// NeuroScript Version: 0.3.1
// File version: 0.0.7 // Correct ToolRegistry interface usage.
// nlines: 273
// risk_rating: HIGH
// filename: pkg/security/security.go
package security

import (
	"fmt"
	"path/filepath"
	"regexp" // Make sure regexp is imported
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/google/generative-ai-go/genai"
)

type ADlist = map[types.FullName]bool

// SecurityLayer enforces security policies for LLM-initiated tool calls.
type SecurityLayer struct {
	allowlist    map[types.FullName]bool
	denylist     map[types.FullName]bool
	sandboxRoot  string            // Unexported field storing the validated path
	toolRegistry tool.ToolRegistry // FIXED: Changed from policy.ToolRegistry
	logger       interfaces.Logger
}

// NewSecurityLayer creates a new security layer instance.
func NewSecurityLayer(allowlistTools ADlist, denylistSet ADlist, sandboxRoot string, registry tool.ToolRegistry, logger interfaces.Logger) *SecurityLayer {

	if logger == nil {
		// This should ideally return an error or use a default logger,
		// but panicking ensures it's caught during development.
		panic("SecurityLayer must have a valid logger")
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
		allowlist:    allowlistTools,
		denylist:     denylistSet,
		sandboxRoot:  cleanSandboxRoot,
		toolRegistry: registry, // <<< CHANGED: Direct assignment of the interface
		logger:       logger,
	}

	logger.Debug("[SEC] Initialized Security Layer.")
	allowlistedNames := make([]string, 0, len(sl.allowlist))
	for tool := range sl.allowlist {
		allowlistedNames = append(allowlistedNames, string(tool))
	}
	logger.Debug("[SEC] Allowlisted tools", "tools", strings.Join(allowlistedNames, ", "))

	deniedNamesNormalized := make([]string, 0, len(sl.denylist))
	for tool := range sl.denylist {
		deniedNamesNormalized = append(deniedNamesNormalized, string(tool))
	}
	logger.Debug("[SEC] Denied tools", "tools", strings.Join(deniedNamesNormalized, ", "))
	logger.Debug("[SEC] Sandbox Root Set To", "path", sl.sandboxRoot)

	if sl.toolRegistry == nil { // Check if the provided interface is nil
		logger.Warn("[SEC] SecurityLayer initialized with nil ToolRegistry. Tool validation/execution will likely fail.")
	}
	return sl
}

// SandboxRoot returns the configured root directory for sandboxing file operations.
func (sl *SecurityLayer) SandboxRoot() string { return sl.sandboxRoot }

// GetToolDeclarations generates the list of genai.Tool objects for allowlisted tools.
// func (sl *SecurityLayer) GetToolDeclarations() ([]*genai.Tool, error) {
// 	if sl.toolRegistry == nil {
// 		sl.logger.Error("[SEC] Cannot get tool declarations: ToolRegistry is nil.")
// 		return nil, fmt.Errorf("%w: security layer tool registry is not initialized", lang.ErrConfiguration)
// 	}
// 	declarations := []*genai.Tool{}
// 	// Calls on sl.toolRegistry (which is now ToolRegistry interface type) should work correctly.
// 	allToolSpecs := sl.toolRegistry.ListTools()
// 	sl.logger.Debug("[SEC] Generating declarations from registered tool specs", "count", len(allToolSpecs))

// 	for _, spec := range allToolSpecs {
// 		// Check allowlist and denylist using the qualified name
// 		isAllowed := sl.allowlist[spec.FullName]
// 		isDenied := sl.denylist[spec.FullName]

// 		if isAllowed && !isDenied {
// 			sl.logger.Debug("[SEC] Generating declaration for allowlisted/not-denied tool", "qualified_name", spec.FullName)
// 			schema := &genai.Schema{
// 				Type:        genai.TypeObject,
// 				Properties:  map[string]*genai.Schema{},
// 				Required:    []string{},
// 				Description: spec.Description,
// 			}
// 			validSchema := true
// 			for _, argSpec := range spec.Args {
// 				// FIXED: Replaced direct method call with a call to the new utility function
// 				genaiType, typeErr := ai.ToGenaiType(argSpec.Type)
// 				if typeErr != nil {
// 					sl.logger.Error("[SEC] Failed to convert arg type for tool declaration", "arg_name", argSpec.Name, "arg_type", argSpec.Type, "tool_name", spec.FullName, "error", typeErr)
// 					validSchema = false
// 					break // Stop processing args for this tool if one is bad
// 				}
// 				schema.Properties[argSpec.Name] = &genai.Schema{Type: genaiType, Description: argSpec.Description}
// 				if argSpec.Required {
// 					schema.Required = append(schema.Required, argSpec.Name)
// 				}
// 			}

// 			if validSchema {
// 				declarations = append(declarations, &genai.Tool{
// 					FunctionDeclarations: []*genai.FunctionDeclaration{
// 						{
// 							Name:        string(spec.FullName),
// 							Description: spec.Description,
// 							Parameters:  schema,
// 						},
// 					},
// 				})
// 				sl.logger.Debug("[SEC] Added declaration", "qualified_name", spec.FullName)
// 			} else {
// 				sl.logger.Warn("[SEC] Skipping declaration due to invalid schema", "qualified_name", spec.FullName)
// 			}
// 		} else {
// 			sl.logger.Debug("[SEC] Tool not included in declarations (not allowed or explicitly denied)", "full_name", spec.FullName, "is_allowed", isAllowed, "is_denied", isDenied)
// 		}
// 	}
// 	sl.logger.Debug("[SEC] Generated tool declarations.", "count", len(declarations))
// 	return declarations, nil
// }

// ExecuteToolCall validates and executes a requested tool call.
func (sl *SecurityLayer) ExecuteToolCall(interpreter tool.Runtime, fc genai.FunctionCall) (genai.Part, error) {
	fulltoolname := types.FullName(fc.Name)
	rawArgs := fc.Args

	sl.logger.Debug("[SEC ExecuteToolCall] Received request", "tool_name", fulltoolname)

	if sl.toolRegistry == nil {
		err := fmt.Errorf("%w: tool registry is not available in security layer", lang.ErrInternalSecurity)
		sl.logger.Error("[SEC ExecuteToolCall] Critical internal error", "error", err, "tool_name", fulltoolname)
		return CreateErrorFunctionResultPart(fulltoolname, err), err
	}

	validatedArgsMap, validationErr := sl.ValidateToolCall(fulltoolname, rawArgs)
	if validationErr != nil {
		sl.logger.Warn("[SEC ExecuteToolCall] Validation failed for tool call", "tool_name", fulltoolname, "error", validationErr)
		return CreateErrorFunctionResultPart(fulltoolname, validationErr), validationErr
	}

	sl.logger.Debug("[SEC ExecuteToolCall] Tool call validated, proceeding to execution", "tool_name", fulltoolname)

	toolImpl, found := sl.toolRegistry.GetTool(fulltoolname)
	if !found {
		// This case should ideally be caught by ValidateToolCall if the tool isn't in the registry at all.
		// However, if it was removed between validation and execution, or if ValidateToolCall's check isn't exhaustive.
		err := fmt.Errorf("%w: tool implementation '%s' not found post-validation (tool name: %s)", lang.ErrInternalSecurity, fulltoolname, fulltoolname)
		sl.logger.Error("[SEC ExecuteToolCall] Critical internal error", "error", err)
		return CreateErrorFunctionResultPart(fulltoolname, err), err
	}

	// Reconstruct ordered arguments based on ToolSpec
	orderedArgs := make([]interface{}, len(toolImpl.Spec.Args))
	conversionOk := true
	for i, argSpec := range toolImpl.Spec.Args {
		val, exists := validatedArgsMap[argSpec.Name]
		if !exists {
			if argSpec.Required {
				// This should also be caught by ValidateToolCall, but as a safeguard:
				err := fmt.Errorf("%w: required arg '%s' missing post-validation for tool '%s'", lang.ErrInternalSecurity, argSpec.Name, fulltoolname)
				sl.logger.Error("[SEC ExecuteToolCall] Critical internal error: missing required arg", "error", err)
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
		err := fmt.Errorf("%w: failed to reconstruct ordered args post-validation for tool '%s'", lang.ErrInternalSecurity, fulltoolname)
		sl.logger.Error("[SEC ExecuteToolCall] Critical internal error: arg reconstruction failed", "error", err)
		return CreateErrorFunctionResultPart(fulltoolname, err), err
	}

	sl.logger.Debug("[SEC ExecuteToolCall] Executing tool function", "qualified_name", fulltoolname, "full_name", fulltoolname)
	resultValue, execErr := toolImpl.Func(interpreter, orderedArgs)
	if execErr != nil {
		sl.logger.Error("[SEC ExecuteToolCall] Tool execution failed", "tool_name", fulltoolname, "error", execErr)
		// It's important that execErr is a well-formed error, ideally a RuntimeError
		return CreateErrorFunctionResultPart(fulltoolname, execErr), execErr
	}

	sl.logger.Debug("[SEC ExecuteToolCall] Tool execution successful", "tool_name", fulltoolname)
	return CreateSuccessFunctionResultPart(fulltoolname, resultValue, sl.logger), nil
}

// ValidateToolCall checks denylist, allowlist, high-risk status, and delegates argument validation.
func (sl *SecurityLayer) ValidateToolCall(fulltoolname types.FullName, rawArgs map[string]interface{}) (map[string]interface{}, error) {
	sl.logger.Debug("[SEC ValidateToolCall] Validating request", "tool_name", fulltoolname)

	if sl.denylist[fulltoolname] {
		sl.logger.Warn("[SEC ValidateToolCall] DENIED (Denylist)", "tool_name", fulltoolname)
		return nil, fmt.Errorf("tool %q denied by denylist: %w", fulltoolname, lang.ErrToolDenied)
	}
	if !sl.allowlist[fulltoolname] {
		sl.logger.Warn("[SEC ValidateToolCall] DENIED (Not Allowlisted)", "tool_name", fulltoolname)
		return nil, fmt.Errorf("tool %q not allowed by allowlist: %w", fulltoolname, lang.ErrToolNotAllowed)
	}

	// Example of a hardcoded policy, can be expanded
	if fulltoolname == "TOOL.ExecuteCommand" { // Assuming ExecuteCommand is the qualified name
		sl.logger.Warn("[SEC ValidateToolCall] DENIED (Blocked by Policy)", "tool_name", fulltoolname, "reason", "ExecuteCommand is disabled")
		return nil, fmt.Errorf("tool %q blocked by security policy: %w", fulltoolname, lang.ErrToolBlocked)
	}

	if sl.toolRegistry == nil {
		sl.logger.Error("[SEC ValidateToolCall] ToolRegistry not available during validation.", "tool_name", fulltoolname)
		return nil, fmt.Errorf("%w: tool registry unavailable for validation of '%s'", lang.ErrInternalSecurity, fulltoolname)
	}

	toolImpl, found := sl.toolRegistry.GetTool(fulltoolname)
	if !found {
		sl.logger.Error("[SEC ValidateToolCall] Allowlisted tool implementation not found in registry.", "qualified_name", fulltoolname, "full_name", fulltoolname)
		return nil, fmt.Errorf("%w: allowlisted tool '%s' (base: '%s') implementation not found", lang.ErrInternalSecurity, fulltoolname, fulltoolname)
	}

	toolSpec := toolImpl.Spec
	sl.logger.Debug("[SEC ValidateToolCall] Found tool spec for validation", "tool_name", fulltoolname, "base_name", fulltoolname, "args_count_in_spec", len(toolSpec.Args))

	// Delegate to the argument validation logic (assumed to be in security_validation.go or similar)
	validatedArgs, validationErr := sl.validateArgumentsAgainstSpec(toolSpec, rawArgs)
	if validationErr != nil {
		sl.logger.Warn("[SEC ValidateToolCall] DENIED (Argument Validation Failed)", "tool_name", fulltoolname, "error", validationErr)
		return nil, validationErr // Return the specific validation error
	}

	sl.logger.Debug("[SEC ValidateToolCall] Arguments validated successfully.", "tool_name", fulltoolname)
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
