// NeuroScript Version: 0.3.1
// File version: 0.0.9 // Remove debug Printf statements.
// nlines: 105 // Approximate
// risk_rating: HIGH // Security-critical path validation
// filename: pkg/core/security_helpers.go
package core

import (
	"errors"
	"fmt"
	"os" // Import os for PathSeparator
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/logging" // Needed for logger in CreateSuccess
	"github.com/google/generative-ai-go/genai"      // Needed for genai types
)

// --- CreateErrorFunctionResultPart unchanged ---
func CreateErrorFunctionResultPart(qualifiedToolName string, execErr error) genai.Part {
	errMsg := "unknown execution error"
	if execErr != nil {
		errMsg = execErr.Error() // Use the error's message
	}
	return genai.FunctionResponse{
		Name: qualifiedToolName,
		Response: map[string]interface{}{
			"error": fmt.Sprintf("Tool execution failed: %s", errMsg),
		},
	}
}

// --- CreateSuccessFunctionResultPart unchanged ---
func CreateSuccessFunctionResultPart(qualifiedToolName string, resultValue interface{}, logger logging.Logger) genai.Part {
	responseMap := make(map[string]interface{})
	switch v := resultValue.(type) {
	case map[string]interface{}:
		responseMap = v
		if _, ok := responseMap["status"]; !ok {
			responseMap["status"] = "success"
		}
	case []map[string]interface{}:
		responseMap["result_list"] = v
		responseMap["status"] = "success"
	case []interface{}:
		responseMap["result_list"] = v
		responseMap["status"] = "success"
	case []string:
		responseMap["result_list"] = v
		responseMap["status"] = "success"
	case string, int, int64, float32, float64, bool:
		responseMap["result"] = v
		responseMap["status"] = "success"
	case nil:
		responseMap["status"] = "success (no explicit result returned)"
	default:
		formattedResult := fmt.Sprintf("%v", v)
		responseMap["result"] = formattedResult
		responseMap["status"] = "success (formatted)"
		if logger != nil {
			logger.Warn("Tool returned unexpected type, formatting as string", "tool", qualifiedToolName, "type", fmt.Sprintf("%T", v), "formatted_result", formattedResult)
		}
	}
	return genai.FunctionResponse{Name: qualifiedToolName, Response: responseMap}
}

// --- Deprecated: GetSandboxPath unchanged ---
func GetSandboxPath(sandboxRoot, relativePath string) string {
	absRoot, _ := filepath.Abs(sandboxRoot)
	if absRoot == "" {
		absRoot = "."
	}
	return filepath.Join(absRoot, relativePath)
}

// --- IsPathInSandbox unchanged ---
func IsPathInSandbox(sandboxRoot, inputPath string) (bool, error) {
	_, err := ResolveAndSecurePath(inputPath, sandboxRoot)
	if err != nil {
		if re, ok := err.(*RuntimeError); ok && errors.Is(re.Wrapped, ErrPathViolation) {
			return false, nil // Specific path violation is not an error for the check, just means "false"
		}
		return false, err // Other errors during resolution are returned
	}
	return true, nil // No error means path is inside
}

// ResolveAndSecurePath resolves an input path (expected to be relative to allowedRoot)
// to an absolute path and validates it is contained within the allowed directory root.
// Returns the validated *absolute* path or a *RuntimeError.
func ResolveAndSecurePath(inputPath, allowedRoot string) (string, error) {
	// --- Input Validation ---
	if inputPath == "" {
		return "", NewRuntimeError(ErrorCodeArgMismatch, "input path cannot be empty", ErrInvalidArgument)
	}
	if strings.Contains(inputPath, "\x00") {
		return "", NewRuntimeError(ErrorCodeSecurity, "input path contains null byte", ErrNullByteInArgument)
	}
	if filepath.IsAbs(inputPath) {
		errMsg := fmt.Sprintf("input file path %q must be relative, not absolute", inputPath)
		return "", NewRuntimeError(ErrorCodeSecurity, errMsg, ErrPathViolation)
	}

	// --- Resolve Paths ---
	absAllowedRoot, err := filepath.Abs(allowedRoot)
	if err != nil {
		return "", NewRuntimeError(ErrorCodeConfiguration, fmt.Sprintf("could not get absolute path for allowed root %q: %v", allowedRoot, err), errors.Join(ErrConfiguration, err))
	}
	absAllowedRoot = filepath.Clean(absAllowedRoot)

	resolvedPath := filepath.Join(absAllowedRoot, inputPath)
	resolvedPath = filepath.Clean(resolvedPath) // Critical: Simplifies ../ etc.

	// --- Robust Check: Use filepath.Rel ---
	rel, err := filepath.Rel(absAllowedRoot, resolvedPath)
	if err != nil {
		// This error might occur if paths are on different volumes on Windows, etc.
		details := fmt.Sprintf("internal error checking path relationship between %q and %q", absAllowedRoot, resolvedPath)
		return "", NewRuntimeError(ErrorCodeInternal, details, errors.Join(ErrInternalSecurity, err))
	}

	// --- IsOutside Check using path components ---
	parts := strings.Split(rel, string(os.PathSeparator))
	isOutside := false
	// If the first path component after splitting is "..", it's outside.
	if len(parts) > 0 && parts[0] == ".." {
		isOutside = true
	}
	// Handle the case where rel is exactly ".." which Split might return as [".."]
	if rel == ".." {
		isOutside = true
	}
	// Ensure the root itself (rel == ".") is not considered outside.
	if rel == "." {
		isOutside = false
	}
	// --- End Check ---

	if isOutside {
		details := fmt.Sprintf("relative path %q resolves to %q which is outside the allowed directory %q", inputPath, resolvedPath, absAllowedRoot)
		return "", NewRuntimeError(ErrorCodeSecurity, details, ErrPathViolation)
	}

	return resolvedPath, nil
}

// SecureFilePath wraps ResolveAndSecurePath
func SecureFilePath(relativePath, allowedRoot string) (string, error) {
	return ResolveAndSecurePath(relativePath, allowedRoot)
}
