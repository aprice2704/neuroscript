// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Add CreateSuccessFunctionResultPart
// filename: pkg/core/security_helpers.go
package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/logging" // Needed for logger in CreateSuccess
	"github.com/google/generative-ai-go/genai"      // Needed for genai types
)

// CreateErrorFunctionResultPart formats a tool execution error into a genai.Part suitable
// for returning to the LLM. It wraps the error message in a standard map structure.
// Assumes this function already exists or is desired here.
func CreateErrorFunctionResultPart(qualifiedToolName string, execErr error) genai.Part {
	errMsg := "unknown execution error"
	if execErr != nil {
		errMsg = execErr.Error() // Use the error's message
	}
	// Log the error before creating the response? Security Layer already logs.
	return genai.FunctionResponse{
		Name: qualifiedToolName,
		Response: map[string]interface{}{
			"error": fmt.Sprintf("Tool execution failed: %s", errMsg),
		},
	}
}

// CreateSuccessFunctionResultPart formats a successful tool execution result into a genai.Part.
// It attempts to intelligently format common result types (maps, slices, primitives)
// into a map structure for the LLM response.
func CreateSuccessFunctionResultPart(qualifiedToolName string, resultValue interface{}, logger logging.Logger) genai.Part {
	responseMap := make(map[string]interface{})

	switch v := resultValue.(type) {
	case map[string]interface{}:
		// If the tool already returned a map, use it directly.
		// Avoid potential key collisions by merging instead? For now, direct use.
		responseMap = v
		// Add a default status if not present?
		if _, ok := responseMap["status"]; !ok {
			responseMap["status"] = "success"
		}
	case []interface{}:
		// If it's a slice, wrap it in a "result_list" key.
		responseMap["result_list"] = v
		responseMap["status"] = "success"
	case []string: // Handle specific common slice types
		responseMap["result_list"] = v
		responseMap["status"] = "success"
	case string, int, int64, float32, float64, bool:
		// For primitive types, wrap them in a "result" key.
		responseMap["result"] = v
		responseMap["status"] = "success"
	case nil:
		// If the tool returned nil explicitly, indicate success without a specific result.
		responseMap["status"] = "success (no explicit result returned)"
	default:
		// For other types, attempt to format them as a string in the "result" key.
		formattedResult := fmt.Sprintf("%v", v)
		responseMap["result"] = formattedResult
		responseMap["status"] = "success (formatted)"
		if logger != nil { // Check logger exists
			logger.Warn("Tool returned unexpected type, formatting as string",
				"tool", qualifiedToolName,
				"type", fmt.Sprintf("%T", v),
				"formatted_result", formattedResult)
		}
	}

	// Log the final response map being sent back? Maybe too verbose.
	// logger.Debug("Formatted success response", "tool", qualifiedToolName, "responseMap", responseMap)

	return genai.FunctionResponse{
		Name:     qualifiedToolName, // Use qualified name back to LLM
		Response: responseMap,
	}
}

// --- Other existing helpers ---

// Deprecated: Use ResolveAndSecurePath instead for safer path handling.
func GetSandboxPath(sandboxRoot, relativePath string) string {
	// ... (implementation unchanged) ...
	absRoot, _ := filepath.Abs(sandboxRoot)
	if absRoot == "" {
		absRoot = "."
	}
	return filepath.Join(absRoot, relativePath)
}

// IsPathInSandbox checks if the given path is within the allowed sandbox directory.
// Returns true if the path is valid and within bounds, false otherwise.
func IsPathInSandbox(sandboxRoot, inputPath string) (bool, error) {
	// ... (implementation unchanged) ...
	_, err := ResolveAndSecurePath(inputPath, sandboxRoot)
	if err != nil {
		if errors.Is(err, ErrPathViolation) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ResolveAndSecurePath resolves an input path (absolute or relative TO THE ALLOWED ROOT)
// to an absolute path and validates it is contained within the allowed directory root.
// Returns the validated *absolute* path or an error (wrapping ErrPathViolation or others).
func ResolveAndSecurePath(inputPath, allowedRoot string) (string, error) {
	// ... (implementation unchanged) ...
	if inputPath == "" {
		return "", fmt.Errorf("input path cannot be empty: %w", ErrPathViolation)
	}
	if strings.Contains(inputPath, "\x00") {
		return "", fmt.Errorf("input path contains null byte: %w", ErrNullByteInArgument)
	}

	absAllowedRoot, err := filepath.Abs(allowedRoot)
	if err != nil {
		return "", fmt.Errorf("%w: could not get absolute path for allowed root %q: %v", ErrInternalSecurity, allowedRoot, err)
	}
	absAllowedRoot = filepath.Clean(absAllowedRoot)

	resolvedPath := ""
	if filepath.IsAbs(inputPath) {
		resolvedPath = filepath.Clean(inputPath)
	} else {
		resolvedPath = filepath.Join(absAllowedRoot, inputPath)
		resolvedPath = filepath.Clean(resolvedPath)
	}

	prefixToCheck := absAllowedRoot
	if prefixToCheck != string(filepath.Separator) && !strings.HasSuffix(prefixToCheck, string(filepath.Separator)) {
		prefixToCheck += string(filepath.Separator)
	}

	if resolvedPath != absAllowedRoot && !strings.HasPrefix(resolvedPath, prefixToCheck) {
		details := fmt.Sprintf("path %q (resolves to %q) is outside the allowed root %q", inputPath, resolvedPath, absAllowedRoot)
		return "", fmt.Errorf("%s: %w", details, ErrPathViolation)
	}

	return resolvedPath, nil
}
