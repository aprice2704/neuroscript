package core

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// --- Utility Helpers ---

// trimCodeFences removes leading/trailing code fences (``` or ```lang)
// Moved from interpreter_c.go
func trimCodeFences(code string) string {
	trimmed := strings.TrimSpace(code)
	lines := strings.Split(trimmed, "\n")
	if len(lines) < 1 {
		return code
	}
	firstLineTrimmed := strings.TrimSpace(lines[0])
	startFenceFound := false
	// More general check for ``` optionally followed by language hint
	if strings.HasPrefix(firstLineTrimmed, "```") {
		// Check if it's ONLY ``` or ``` plus non-space chars
		restOfLine := strings.TrimSpace(firstLineTrimmed[3:])
		if len(restOfLine) == 0 || !strings.ContainsAny(restOfLine, " \t") { // Allow ``` or ```lang, but not ``` lang with space
			startFenceFound = true
			lines = lines[1:]
		}
	}
	endFenceFound := false
	if len(lines) > 0 {
		lastLineTrimmed := strings.TrimSpace(lines[len(lines)-1])
		if lastLineTrimmed == "```" {
			endFenceFound = true
			lines = lines[:len(lines)-1]
		}
	}
	if startFenceFound || endFenceFound {
		return strings.TrimSpace(strings.Join(lines, "\n"))
	}
	return trimmed // Return original trimmed if no fences found
}

// sanitizeFilename creates a safe filename component.
// Moved from interpreter_c.go
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	// Allow alphanumeric, underscore, hyphen, dot. Remove others.
	removeChars := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	name = removeChars.ReplaceAllString(name, "")
	// Remove leading/trailing dots, underscores, hyphens more carefully
	name = strings.TrimLeft(name, "._-")
	name = strings.TrimRight(name, "._-")
	// Collapse multiple underscores/hyphens/dots
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")
	name = regexp.MustCompile(`\.{2,}`).ReplaceAllString(name, ".") // Avoid .. in middle
	name = strings.ReplaceAll(name, "..", "_")                      // Replace remaining ".." just in case

	const maxLength = 100
	if len(name) > maxLength {
		lastSep := strings.LastIndexAny(name[:maxLength], "_-.")
		if lastSep > maxLength/2 {
			name = name[:lastSep]
		} else {
			name = name[:maxLength]
		}
		name = strings.TrimRight(name, "._-") // Trim again after potential cut
	}
	if name == "" {
		name = "default_skill_name"
	} // Ensure non-empty
	// Avoid OS reserved names (Windows mainly) - simplistic check
	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "LPT1"}
	upperName := strings.ToUpper(name)
	for _, r := range reserved {
		if upperName == r {
			name = name + "_"
			break
		}
	}

	return name
}

// runGitCommand executes a git command.
// Moved from interpreter_c.go
func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		// Ensure args are properly quoted if they contain spaces
		quotedArgs := make([]string, len(args))
		for i, arg := range args {
			if strings.Contains(arg, " ") {
				quotedArgs[i] = fmt.Sprintf("%q", arg) // Use %q for quoting
			} else {
				quotedArgs[i] = arg
			}
		}
		return fmt.Errorf("git command 'git %s' failed: %v\nStderr: %s", strings.Join(quotedArgs, " "), err, stderr.String())
	}
	return nil
}

// secureFilePath cleans and ensures the path is within the allowed directory (cwd).
// Moved from interpreter_c.go
func secureFilePath(filePath, allowedDir string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}
	// Basic check for null bytes
	if strings.Contains(filePath, "\x00") {
		return "", fmt.Errorf("file path contains null byte")
	}

	absAllowedDir, err := filepath.Abs(allowedDir)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path for allowed directory '%s': %w", allowedDir, err)
	}
	absAllowedDir = filepath.Clean(absAllowedDir)

	// Clean the input path itself first to handle relative traversals better
	cleanedInputPath := filepath.Clean(filePath)
	// Prevent absolute paths in the input 'filePath' argument if allowedDir is meant as root
	if filepath.IsAbs(cleanedInputPath) {
		// Allow if it's within allowedDir? Or disallow always? Let's disallow absolute inputs for now.
		// To allow absolute paths that are *within* allowedDir:
		// absCleanedInputPath := filepath.Clean(filePath)
		// if !strings.HasPrefix(absCleanedInputPath, absAllowedDir) {
		//     return "", fmt.Errorf("absolute input path '%s' is outside allowed directory '%s'", absCleanedInputPath, absAllowedDir)
		// }
		// joinedPath = absCleanedInputPath // Use the already absolute path
		// --- Current behavior: Disallow absolute paths in filePath argument ---
		return "", fmt.Errorf("input file path '%s' must be relative", filePath)
	}

	// Join the cleaned relative path to the absolute allowed directory
	joinedPath := filepath.Join(absAllowedDir, cleanedInputPath)

	// Final clean on the joined path
	absCleanedPath := filepath.Clean(joinedPath)

	// Check if the final absolute path starts with the allowed directory path.
	// Add a separator check to prevent cases like /allowed/dir-abc matching /allowed/dir
	if !strings.HasPrefix(absCleanedPath, absAllowedDir+string(filepath.Separator)) && absCleanedPath != absAllowedDir {
		return "", fmt.Errorf("path '%s' resolves to '%s' which is outside the allowed directory '%s'", filePath, absCleanedPath, absAllowedDir)
	}

	// Additional check: Ensure it's not EXACTLY the allowed dir if filePath wasn't empty or "."
	if absCleanedPath == absAllowedDir && filePath != "." && filePath != "" {
		// This prevents targeting the root directory itself when a specific file/subdir was intended.
		// Allow if filePath is "."? Yes, that explicitly means the root.
		return "", fmt.Errorf("path '%s' resolves to the allowed directory root '%s'", filePath, absCleanedPath)
	}

	return absCleanedPath, nil // Return the safe, absolute, cleaned path
}
