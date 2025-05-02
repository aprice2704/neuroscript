// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 20:06:00 PDT // Updated timestamp
// filename: pkg/core/security_helpers.go
package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	// "os" // Not strictly needed for path manipulation/checks
)

// GetSandboxPath joins a relative path with the sandbox root, returning the absolute path.
// It does NOT perform any validation.
// Deprecated: Use ResolveAndSecurePath instead for safer path handling.
func GetSandboxPath(sandboxRoot, relativePath string) string {
	// This helper should ideally also use filepath.Abs on sandboxRoot first for robustness
	absRoot, _ := filepath.Abs(sandboxRoot) // Ignore error for deprecated func?
	if absRoot == "" {
		absRoot = "."
	}
	return filepath.Join(absRoot, relativePath)
}

// IsPathInSandbox checks if the given path is within the allowed sandbox directory.
// Returns true if the path is valid and within bounds, false otherwise.
func IsPathInSandbox(sandboxRoot, inputPath string) (bool, error) {
	// Use ResolveAndSecurePath internally for consistent logic
	_, err := ResolveAndSecurePath(inputPath, sandboxRoot)
	if err != nil {
		// If the error indicates it's outside the root (ErrPathViolation), return false, nil error.
		// If it's another error (like null byte, internal security), return false and the error.
		if errors.Is(err, ErrPathViolation) {
			return false, nil // It's not in the sandbox, but not necessarily an "error" state for this check
		}
		return false, err // Propagate other validation errors
	}
	// If ResolveAndSecurePath succeeded, the path is inside.
	return true, nil
}

// ResolveAndSecurePath resolves an input path (absolute or relative TO THE ALLOWED ROOT)
// to an absolute path and validates it is contained within the allowed directory root.
// Returns the validated *absolute* path or an error (wrapping ErrPathViolation or others).
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
		// This is likely an internal config error if allowedRoot is invalid
		return "", fmt.Errorf("%w: could not get absolute path for allowed root %q: %v", ErrInternalSecurity, allowedRoot, err)
	}
	absAllowedRoot = filepath.Clean(absAllowedRoot)

	// --- CORRECTED Step 2: Resolve inputPath relative to absAllowedRoot ---
	resolvedPath := ""
	if filepath.IsAbs(inputPath) {
		// If input is already absolute, just clean it.
		// Security check later will ensure it's within the allowed root.
		resolvedPath = filepath.Clean(inputPath)
	} else {
		// If input is relative, join it with the *absolute allowed root*
		resolvedPath = filepath.Join(absAllowedRoot, inputPath)
		// Clean the resulting path (handles redundant separators, .. elements)
		resolvedPath = filepath.Clean(resolvedPath)
	}
	// --- End CORRECTION ---

	// 3. Check for containment: resolvedPath must be exactly absAllowedRoot or a descendant.
	prefixToCheck := absAllowedRoot
	// Add trailing separator if missing AND if it's not the root directory itself "/"
	// This handles cases like allowedRoot="/tmp/sandbox" and path="/tmp/sandboxExt" correctly
	if prefixToCheck != string(filepath.Separator) && !strings.HasSuffix(prefixToCheck, string(filepath.Separator)) {
		prefixToCheck += string(filepath.Separator)
	}

	// Check if the cleaned, resolved path is exactly the allowed root OR starts with the prefix
	if resolvedPath != absAllowedRoot && !strings.HasPrefix(resolvedPath, prefixToCheck) {
		details := fmt.Sprintf("path %q (resolves to %q) is outside the allowed root %q", inputPath, resolvedPath, absAllowedRoot)
		return "", fmt.Errorf("%s: %w", details, ErrPathViolation)
	}

	// 4. Return the validated *absolute* path.
	return resolvedPath, nil
}
