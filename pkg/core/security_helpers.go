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
