// pkg/core/tools_helpers.go
package core

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// secureFilePath cleans and ensures the **relative** path is within the allowed directory (cwd).
// Rejects absolute paths.
func secureFilePath(filePath, allowedDir string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}
	if strings.Contains(filePath, "\x00") {
		return "", fmt.Errorf("file path contains null byte")
	}
	// Reject absolute paths directly
	if filepath.IsAbs(filePath) {
		return "", fmt.Errorf("input file path '%s' must be relative", filePath)
	}

	absAllowedDir, err := filepath.Abs(allowedDir)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path for allowed directory '%s': %w", allowedDir, err)
	}
	absAllowedDir = filepath.Clean(absAllowedDir)

	// Join the allowed dir with the relative path
	absCleanedPath := filepath.Join(absAllowedDir, filePath)
	absCleanedPath = filepath.Clean(absCleanedPath) // Clean the final result

	// Check if the final path is still within the allowed directory.
	// Use filepath.Separator to handle OS differences robustly.
	// Add the separator to absAllowedDir to ensure we match full directory prefixes.
	prefixToCheck := absAllowedDir
	if !strings.HasSuffix(prefixToCheck, string(filepath.Separator)) {
		prefixToCheck += string(filepath.Separator)
	}
	pathToCheck := absCleanedPath
	// Allow matching the directory itself if the input was exactly "."
	if pathToCheck != absAllowedDir && !strings.HasPrefix(pathToCheck, prefixToCheck) {
		return "", fmt.Errorf("relative path '%s' resolves to '%s' which is outside the allowed directory '%s'", filePath, absCleanedPath, absAllowedDir)
	}

	// Check if it resolves *exactly* to the allowed directory root, unless the input was explicitly "."
	// This prevents operations directly on the root unless intended via "."
	if absCleanedPath == absAllowedDir && filepath.Clean(filePath) != "." {
		return "", fmt.Errorf("path '%s' resolves to the allowed directory root '%s', which is not permitted for this operation unless '.' was specified", filePath, absCleanedPath)
	}

	return absCleanedPath, nil // Return the safe, absolute, cleaned path
}

// runGitCommand executes a git command with the given arguments.
func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		// Quote arguments containing spaces for clearer error messages
		quotedArgs := make([]string, len(args))
		for i, arg := range args {
			if strings.Contains(arg, " ") {
				quotedArgs[i] = fmt.Sprintf("%q", arg)
			} else {
				quotedArgs[i] = arg
			}
		}
		return fmt.Errorf("git command 'git %s' failed: %v\nStderr: %s", strings.Join(quotedArgs, " "), err, stderr.String())
	}
	return nil
}
