// filename: pkg/core/tools_helpers.go
package core

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp" // Added regexp
	"strings"
)

// SecureFilePath cleans and ensures the **relative** path is within the allowed directory (cwd).
// Rejects absolute paths.
func SecureFilePath(filePath, allowedDir string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}
	if strings.Contains(filePath, "\x00") {
		return "", fmt.Errorf("file path contains null byte")
	}
	if filepath.IsAbs(filePath) {
		return "", fmt.Errorf("input file path '%s' must be relative", filePath)
	}

	absAllowedDir, err := filepath.Abs(allowedDir)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path for allowed directory '%s': %w", allowedDir, err)
	}
	absAllowedDir = filepath.Clean(absAllowedDir)

	absCleanedPath := filepath.Join(absAllowedDir, filePath)
	absCleanedPath = filepath.Clean(absCleanedPath)

	prefixToCheck := absAllowedDir
	if !strings.HasSuffix(prefixToCheck, string(filepath.Separator)) {
		prefixToCheck += string(filepath.Separator)
	}
	pathToCheck := absCleanedPath

	if pathToCheck != absAllowedDir && !strings.HasPrefix(pathToCheck, prefixToCheck) {
		return "", fmt.Errorf("relative path '%s' resolves to '%s' which is outside the allowed directory '%s'", filePath, absCleanedPath, absAllowedDir)
	}
	if absCleanedPath == absAllowedDir && filepath.Clean(filePath) != "." {
		return "", fmt.Errorf("path '%s' resolves to the allowed directory root '%s', which is not permitted unless '.' was specified", filePath, absCleanedPath)
	}

	return absCleanedPath, nil
}

// runGitCommand executes a git command with the given arguments.
func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
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

// --- ADDED Exported SanitizeFilename ---

// SanitizeFilename cleans a string to be suitable for use as a filename component.
// This is the canonical implementation moved from utils.go.
func SanitizeFilename(name string) string {
	// 1. Replace common separators with underscore
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")

	// 2. Remove characters not allowed in filenames (conservative set)
	// Compile regex inside the function or keep it package-level if used frequently
	removeChars := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	name = removeChars.ReplaceAllString(name, "")

	// 3. Remove leading/trailing unwanted chars
	name = strings.Trim(name, "._-")

	// 4. Collapse multiple separators
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")
	name = strings.ReplaceAll(name, "..", "_") // Avoid '..' sequences
	name = regexp.MustCompile(`\.{2,}`).ReplaceAllString(name, ".")

	// 5. Truncate to a reasonable length
	const maxLength = 100
	if len(name) > maxLength {
		// Try to cut nicely
		lastSep := strings.LastIndexAny(name[:maxLength], "_-.")
		if lastSep > maxLength/2 {
			name = name[:lastSep]
		} else {
			name = name[:maxLength]
		}
		name = strings.TrimRight(name, "._-") // Ensure it doesn't end badly after cut
	}

	// 6. Handle empty result
	if name == "" {
		name = "default_sanitized_name"
	}

	// 7. Avoid reserved names (case-insensitive check)
	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperName := strings.ToUpper(name)
	// Also check base name without extension if applicable (e.g., file.CON) - simple check for now
	baseName := upperName
	if dotIndex := strings.LastIndex(upperName, "."); dotIndex != -1 {
		baseName = upperName[:dotIndex]
	}
	for _, r := range reserved {
		if upperName == r || baseName == r {
			name = name + "_" // Append underscore if reserved
			break
		}
	}

	return name
}

// --- END ADDED SanitizeFilename ---
