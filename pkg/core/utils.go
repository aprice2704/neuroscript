// pkg/core/utils.go
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

// trimCodeFences removes code fences (```) from the beginning and end of a code string.
// It handles cases with and without content between the fences and also accounts for
// additional whitespace or content on the fence lines.  The function first trims
// leading/trailing whitespace.  Then it checks if the first and last lines are code fences
// (```), accounting for potential extra whitespace or text on those lines. If fences are
// found, they are removed, and the resulting string is returned. Otherwise, the original
// (trimmed) string is returned.
func trimCodeFences(code string) string {
	trimmed := strings.TrimSpace(code)
	lines := strings.Split(trimmed, "\n")
	if len(lines) < 1 {
		return code
	}
	firstLineTrimmed := strings.TrimSpace(lines[0])
	startFenceFound := false
	if strings.HasPrefix(firstLineTrimmed, "```") {
		restOfLine := strings.TrimSpace(firstLineTrimmed[3:])
		if len(restOfLine) == 0 || !strings.ContainsAny(restOfLine, " \t") {
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
	return trimmed
}
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	removeChars := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	name = removeChars.ReplaceAllString(name, "")
	name = strings.TrimLeft(name, "._-")
	name = strings.TrimRight(name, "._-")
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")
	name = regexp.MustCompile(`\.{2,}`).ReplaceAllString(name, ".")
	name = strings.ReplaceAll(name, "..", "_")
	const maxLength = 100
	if len(name) > maxLength {
		lastSep := strings.LastIndexAny(name[:maxLength], "_-.")
		if lastSep > maxLength/2 {
			name = name[:lastSep]
		} else {
			name = name[:maxLength]
		}
		name = strings.TrimRight(name, "._-")
	}
	if name == "" {
		name = "default_skill_name"
	}
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

// secureFilePath cleans and ensures the **relative** path is within the allowed directory (cwd).
// *** REVERTED: Rejects absolute paths. ***
func secureFilePath(filePath, allowedDir string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}
	if strings.Contains(filePath, "\x00") {
		return "", fmt.Errorf("file path contains null byte")
	}
	// *** Reject absolute paths directly ***
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
	if !strings.HasPrefix(absCleanedPath, absAllowedDir) {
		return "", fmt.Errorf("relative path '%s' resolves to '%s' which is outside the allowed directory '%s'", filePath, absCleanedPath, absAllowedDir)
	}

	// Check if it resolves *exactly* to the allowed directory root.
	if absCleanedPath == absAllowedDir && filepath.Clean(filePath) != "." {
		return "", fmt.Errorf("path '%s' resolves to the allowed directory root '%s', which is not permitted for this operation", filePath, absCleanedPath)
	}

	return absCleanedPath, nil // Return the safe, absolute, cleaned path
}

func parseDocstring(content string) Docstring {
	doc := Docstring{Inputs: make(map[string]string)}
	var currentSection *string
	inInputSection := false
	sectionMap := map[string]*string{"PURPOSE:": &doc.Purpose, "OUTPUT:": &doc.Output, "ALGORITHM:": &doc.Algorithm, "CAVEATS:": &doc.Caveats, "EXAMPLES:": &doc.Examples}
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentLines []string
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		foundSectionHeader := false
		upperTrimmedLine := strings.ToUpper(trimmedLine)
		for prefix, targetPtr := range sectionMap {
			if strings.HasPrefix(upperTrimmedLine, prefix) {
				if currentSection != nil {
					*currentSection = strings.TrimSpace(strings.Join(currentLines, "\n"))
				}
				currentSection = targetPtr
				currentLines = []string{strings.TrimSpace(strings.TrimPrefix(trimmedLine, prefix))}
				inInputSection = false
				foundSectionHeader = true
				break
			}
		}
		if foundSectionHeader {
			continue
		}
		if strings.HasPrefix(upperTrimmedLine, "INPUTS:") {
			if currentSection != nil {
				*currentSection = strings.TrimSpace(strings.Join(currentLines, "\n"))
			}
			currentSection = nil
			inInputSection = true
			currentLines = []string{}
			inputContent := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "INPUTS:"))
			if strings.ToLower(inputContent) != "none" && inputContent != "" {
				parseInputLine(inputContent, &doc)
			}
			continue
		}
		if inInputSection {
			parseInputLine(trimmedLine, &doc)
		} else if currentSection != nil {
			currentLines = append(currentLines, line)
		}
	}
	if currentSection != nil {
		*currentSection = strings.TrimSpace(strings.Join(currentLines, "\n"))
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("[Error] Scanner error in parseDocstring: %v\n", err)
	}
	return doc
}
func parseInputLine(line string, doc *Docstring) {
	trimmedLine := strings.TrimSpace(line)
	if strings.HasPrefix(trimmedLine, "-") {
		parts := strings.SplitN(strings.TrimSpace(trimmedLine[1:]), ":", 2)
		if len(parts) == 2 {
			inputName := strings.TrimSpace(parts[0])
			inputDesc := strings.TrimSpace(parts[1])
			if inputName != "" {
				if doc.Inputs == nil {
					doc.Inputs = make(map[string]string)
				}
				doc.Inputs[inputName] = inputDesc
			}
		}
	}
}