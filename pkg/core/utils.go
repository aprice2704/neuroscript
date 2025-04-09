// pkg/core/utils.go
package core

import (
	"bufio"
	// "bytes" // No longer needed here
	"fmt"
	// "os/exec" // No longer needed here
	// "path/filepath" // No longer needed here
	"regexp"
	"strings"
)

// --- Utility Helpers ---

// trimCodeFences removes code fences (```) from the beginning and end of a code string.
// Handles optional language identifiers and whitespace.
func trimCodeFences(code string) string {
	trimmed := strings.TrimSpace(code)
	// Test comment line before variable assignment
	lines := strings.Split(trimmed, "\n")

	if len(lines) == 0 {
		return trimmed // Return trimmed original if empty after trimming
	}

	startFenceFound := false
	endFenceFound := false

	// Check first line for start fence (``` optional_lang)
	firstLineTrimmed := strings.TrimSpace(lines[0])
	if strings.HasPrefix(firstLineTrimmed, "```") {
		// Check if it's just the fence or fence + lang id
		restOfLine := strings.TrimSpace(firstLineTrimmed[3:])
		// Allow empty or valid identifier chars for lang ID
		isValidLangID := true
		if len(restOfLine) > 0 {
			for _, r := range restOfLine {
				if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
					isValidLangID = false
					break
				}
			}
		}
		if isValidLangID {
			startFenceFound = true
			lines = lines[1:] // Remove the first line
		}
	}

	// Check last line for end fence (```)
	if len(lines) > 0 {
		lastLineTrimmed := strings.TrimSpace(lines[len(lines)-1])
		if lastLineTrimmed == "```" {
			endFenceFound = true
			lines = lines[:len(lines)-1] // Remove the last line
		}
	}

	// Return joined content only if at least one fence was removed
	if startFenceFound || endFenceFound {
		return strings.TrimSpace(strings.Join(lines, "\n"))
	}

	// Otherwise return the originally trimmed string
	return trimmed
}

// sanitizeFilename cleans a string to be suitable for use as a filename.
func sanitizeFilename(name string) string {
	// 1. Replace common separators with underscore
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")

	// 2. Remove characters not allowed in filenames (conservative set)
	removeChars := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	name = removeChars.ReplaceAllString(name, "")

	// 3. Remove leading/trailing unwanted chars (dots, underscores, hyphens)
	name = strings.TrimLeft(name, "._-")
	name = strings.TrimRight(name, "._-")

	// 4. Collapse multiple underscores/hyphens/dots
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")
	// Be careful with dots to avoid issues like '..'
	name = strings.ReplaceAll(name, "..", "_") // Replace '..' with underscore first
	name = regexp.MustCompile(`\.{2,}`).ReplaceAllString(name, ".")

	// 5. Truncate to a reasonable maximum length
	const maxLength = 100
	if len(name) > maxLength {
		// Try to cut at a separator if possible
		lastSep := strings.LastIndexAny(name[:maxLength], "_-.")
		if lastSep > maxLength/2 { // Avoid cutting too early
			name = name[:lastSep]
		} else {
			name = name[:maxLength]
		}
		// Ensure it doesn't end with a separator after truncation
		name = strings.TrimRight(name, "._-")
	}

	// 6. Handle empty result
	if name == "" {
		name = "default_skill_name" // Provide a default
	}

	// 7. Avoid reserved filenames (Windows mainly, but good practice)
	// Check base name without extension if relevant later
	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperName := strings.ToUpper(name)
	for _, r := range reserved {
		if upperName == r {
			name = name + "_" // Append underscore if reserved
			break
		}
	}

	return name
}

// runGitCommand moved to tools_helpers.go
// SecureFilePath moved to tools_helpers.go

// parseDocstring extracts structured information from a COMMENT: block.
func parseDocstring(content string) Docstring {
	doc := Docstring{Inputs: make(map[string]string)} // 'doc' is a value here
	var currentSection *string                        // Pointer to the current section's string field in Docstring
	inInputSection := false

	// Map section headers (uppercase) to pointers in the Docstring struct
	sectionMap := map[string]*string{
		"PURPOSE:":   &doc.Purpose,
		"OUTPUT:":    &doc.Output,
		"ALGORITHM:": &doc.Algorithm,
		"CAVEATS:":   &doc.Caveats,
		"EXAMPLES:":  &doc.Examples,
	}

	// Special handling slice for INPUTS:
	var inputLines *[]string = &doc.InputLines // Store raw lines for input

	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentLines []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		upperTrimmedLine := strings.ToUpper(trimmedLine)

		foundSectionHeader := false

		// Check standard sections
		for prefix, targetPtr := range sectionMap { // targetPtr is *string
			if strings.HasPrefix(upperTrimmedLine, prefix) {
				// Finalize previous section
				if currentSection != nil {
					*currentSection = strings.TrimSpace(strings.Join(currentLines, "\n"))
				} else if inInputSection && inputLines != nil {
					// Finalize INPUTS section if we were in it
				}

				// Start new section
				currentSection = targetPtr                                                          // Assign the pointer directly
				currentLines = []string{strings.TrimSpace(strings.TrimPrefix(trimmedLine, prefix))} // Start with content after header
				inInputSection = false
				foundSectionHeader = true
				break
			}
		}

		if foundSectionHeader {
			continue
		}

		// Check INPUTS: section specifically
		if strings.HasPrefix(upperTrimmedLine, "INPUTS:") {
			// Finalize previous section
			if currentSection != nil {
				*currentSection = strings.TrimSpace(strings.Join(currentLines, "\n"))
			}

			// Start INPUTS section
			currentSection = nil // Not writing to a simple string field
			inInputSection = true
			currentLines = []string{} // Reset lines for input parsing
			inputContent := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "INPUTS:"))
			if strings.ToLower(inputContent) != "none" && inputContent != "" {
				// *** CORRECT CALL: Passing address (&doc) which is type *Docstring ***
				parseInputLine(inputContent, &doc)
				if inputLines != nil {
					*inputLines = append(*inputLines, inputContent) // Add raw line
				}
			}
			continue
		}

		// Append line to the current section or inputs
		if inInputSection {
			// *** CORRECT CALL: Passing address (&doc) which is type *Docstring ***
			parseInputLine(trimmedLine, &doc) // Parse structure
			if inputLines != nil {
				*inputLines = append(*inputLines, line) // Add raw line
			}
		} else if currentSection != nil {
			currentLines = append(currentLines, line)
		}
		// Lines before the first section header are ignored
	}

	// Finalize the last section after the loop
	if currentSection != nil {
		*currentSection = strings.TrimSpace(strings.Join(currentLines, "\n"))
	} else if inInputSection && inputLines != nil {
		// Finalize INPUTS section if it was the last one
	}

	// Handle scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("[Error] Scanner error in parseDocstring: %v\n", err)
	}

	return doc // Return the Docstring value
}

// parseInputLine parses a single line within the INPUTS section.
// *** CORRECT SIGNATURE: Expects a pointer *Docstring ***
func parseInputLine(line string, doc *Docstring) {
	trimmedLine := strings.TrimSpace(line)
	// Expect lines like "- name: description"
	if strings.HasPrefix(trimmedLine, "-") {
		parts := strings.SplitN(strings.TrimSpace(trimmedLine[1:]), ":", 2)
		if len(parts) == 2 {
			inputName := strings.TrimSpace(parts[0])
			inputDesc := strings.TrimSpace(parts[1])
			if inputName != "" {
				if doc.Inputs == nil { // Access field via pointer
					doc.Inputs = make(map[string]string)
				}
				doc.Inputs[inputName] = inputDesc // Access field via pointer
			}
		}
	}
}
