package core

import (
	"bufio"
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

// parseDocstring parses the raw content of a comment block into a Docstring struct.
func parseDocstring(content string) Docstring {
	doc := Docstring{
		Inputs: make(map[string]string), // Initialize map
	}
	var currentSection *string // Pointer to the string field currently being appended to
	inInputSection := false    // Flag to track if we are parsing INPUTS lines

	// Map keywords (uppercase) to pointers to the struct fields
	// *** FIX: Change value type from **string to *string ***
	sectionMap := map[string]*string{
		"PURPOSE:":   &doc.Purpose,
		"OUTPUT:":    &doc.Output,
		"ALGORITHM:": &doc.Algorithm,
		"CAVEATS:":   &doc.Caveats,
		"EXAMPLES:":  &doc.Examples,
		// INPUTS: is handled separately below
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentLines []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		foundSectionHeader := false

		// Check for standard section headers (uppercase)
		upperTrimmedLine := strings.ToUpper(trimmedLine)
		for prefix, targetPtr := range sectionMap {
			if strings.HasPrefix(upperTrimmedLine, prefix) {
				// Finalize previous section (if any)
				if currentSection != nil {
					*currentSection = strings.TrimSpace(strings.Join(currentLines, "\n"))
				}
				// Start new section
				currentSection = targetPtr                                                          // *** FIX: Assign the pointer directly ***
				currentLines = []string{strings.TrimSpace(strings.TrimPrefix(trimmedLine, prefix))} // Add first line content (case-preserved)
				inInputSection = false                                                              // No longer in inputs section
				foundSectionHeader = true
				break
			}
		}
		if foundSectionHeader {
			continue
		}

		// Check specifically for INPUTS: header
		if strings.HasPrefix(upperTrimmedLine, "INPUTS:") {
			// Finalize previous section
			if currentSection != nil {
				*currentSection = strings.TrimSpace(strings.Join(currentLines, "\n"))
			}
			currentSection = nil      // Clear current section pointer
			inInputSection = true     // Set input section flag
			currentLines = []string{} // Reset lines buffer

			inputContent := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "INPUTS:"))
			if strings.ToLower(inputContent) != "none" && inputContent != "" {
				parseInputLine(inputContent, &doc) // Parse the first line if not "None"
			}
			continue // Move to next line
		}

		// If it's not a section header, append to the current section or parse as input line
		if inInputSection {
			parseInputLine(trimmedLine, &doc)
		} else if currentSection != nil {
			// Append line with its original leading whitespace relative to the section
			// if len(currentLines) > 0 || trimmedLine != "" { // Avoid adding initial empty lines directly under header
			currentLines = append(currentLines, line) // Append raw line to preserve indentation within section
			// }
		}
		// Ignore lines before the first section header if not in inputs
	}

	// Finalize the very last section being processed
	if currentSection != nil {
		*currentSection = strings.TrimSpace(strings.Join(currentLines, "\n"))
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[Error] Scanner error in parseDocstring: %v\n", err)
		// Optionally return partial doc or an error indicator
	}

	return doc
}

// Helper for parsing INPUTS lines
func parseInputLine(line string, doc *Docstring) {
	trimmedLine := strings.TrimSpace(line)
	// Expecting "- name: description"
	if strings.HasPrefix(trimmedLine, "-") {
		parts := strings.SplitN(strings.TrimSpace(trimmedLine[1:]), ":", 2)
		if len(parts) == 2 {
			inputName := strings.TrimSpace(parts[0])
			inputDesc := strings.TrimSpace(parts[1])
			if inputName != "" {
				if doc.Inputs == nil { // Ensure map is initialized
					doc.Inputs = make(map[string]string)
				}
				doc.Inputs[inputName] = inputDesc
			}
		}
	}
}
