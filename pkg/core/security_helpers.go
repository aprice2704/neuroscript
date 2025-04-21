// filename: pkg/core/security_helpers.go
package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

// GetSandboxPath joins a relative path with the sandbox root, returning the absolute path.
// It does NOT perform any validation.
func GetSandboxPath(sandboxRoot, relativePath string) string {
	return filepath.Join(sandboxRoot, relativePath)
}

// IsPathInSandbox checks if the given path is within the allowed sandbox directory.
// It's a variant of SecureFilePath that returns only a boolean and a simpler error.
func IsPathInSandbox(sandboxRoot, filePath string) (bool, error) {
	if filePath == "" {
		return false, fmt.Errorf("file path cannot be empty")
	}
	if strings.Contains(filePath, "\x00") {
		return false, fmt.Errorf("file path contains null byte")
	}
	if filepath.IsAbs(filePath) {
		return false, fmt.Errorf("input file path '%s' must be relative", filePath)
	}

	absAllowedDir, err := filepath.Abs(sandboxRoot)
	if err != nil {
		// This is an internal configuration error, not a path violation
		return false, fmt.Errorf("could not get absolute path for allowed directory '%s': %w", sandboxRoot, err)
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

	isInSandbox := pathToCheck == absAllowedDir || strings.HasPrefix(pathToCheck, prefixToCheck)

	return isInSandbox, nil
}

// +++ NEW FUNCTION +++
// ResolveAndSecurePath resolves an input path (potentially relative to CWD)
// to an absolute path and validates it against an allowed directory root.
// Returns the validated absolute path or an error wrapping ErrPathViolation.
func ResolveAndSecurePath(inputPath, allowedRoot string) (string, error) {
	if inputPath == "" {
		return "", fmt.Errorf("input path cannot be empty: %w", ErrPathViolation)
	}
	if strings.Contains(inputPath, "\x00") {
		return "", fmt.Errorf("input path contains null byte: %w", ErrNullByteInArgument)
	}

	// 1. Resolve the allowed directory root to an absolute, clean path.
	absAllowedRoot, err := filepath.Abs(allowedRoot)
	if err != nil {
		// This is likely an internal config error
		return "", fmt.Errorf("could not get absolute path for allowed root %q: %w", allowedRoot, ErrInternalSecurity)
	}
	absAllowedRoot = filepath.Clean(absAllowedRoot)

	// 2. Resolve the inputPath to an absolute, clean path (relative to CWD).
	absInputPath, err := filepath.Abs(inputPath)
	if err != nil {
		// Error resolving the user-provided path itself
		return "", fmt.Errorf("could not resolve absolute path for %q: %w", inputPath, err)
	}
	absInputPath = filepath.Clean(absInputPath)

	// 3. Check for containment: absInputPath must be exactly absAllowedRoot or a descendant.
	prefixToCheck := absAllowedRoot
	// Add trailing separator if missing AND if it's not the root directory itself "/"
	if prefixToCheck != string(filepath.Separator) && !strings.HasSuffix(prefixToCheck, string(filepath.Separator)) {
		prefixToCheck += string(filepath.Separator)
	}

	// Check if the cleaned input path is exactly the allowed root OR starts with the prefix
	if absInputPath != absAllowedRoot && !strings.HasPrefix(absInputPath, prefixToCheck) {
		details := fmt.Sprintf("path %q (resolves to %q) is outside the allowed root %q", inputPath, absInputPath, absAllowedRoot)
		return "", fmt.Errorf("%s: %w", details, ErrPathViolation)
	}

	// 4. Return the validated *absolute* input path.
	return absInputPath, nil
}
