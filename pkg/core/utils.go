// filename: pkg/core/utils.go
package core

import (
	"bufio"
	"fmt"
	"strings"
	// No longer need regexp here
)

// --- Utility Helpers ---

// trimCodeFences removes code fences (```) from the beginning and end of a code string.
// Handles optional language identifiers and whitespace.
func trimCodeFences(code string) string {
	trimmed := strings.TrimSpace(code)
	lines := strings.Split(trimmed, "\n")

	if len(lines) == 0 {
		return trimmed
	}

	startFenceFound := false
	endFenceFound := false

	firstLineTrimmed := strings.TrimSpace(lines[0])
	if strings.HasPrefix(firstLineTrimmed, "```") {
		restOfLine := strings.TrimSpace(firstLineTrimmed[3:])
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
			lines = lines[1:]
		}
	}

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

// --- REMOVED sanitizeFilename - Moved to tools_helpers.go ---

// parseDocstring extracts structured information from a COMMENT: block.
func parseDocstring(content string) Docstring {
	doc := Docstring{Inputs: make(map[string]string)}
	var currentSection *string
	inInputSection := false

	sectionMap := map[string]*string{
		"PURPOSE:":   &doc.Purpose,
		"OUTPUT:":    &doc.Output,
		"ALGORITHM:": &doc.Algorithm,
		"CAVEATS:":   &doc.Caveats,
		"EXAMPLES:":  &doc.Examples,
		// LANG_VERSION is handled separately if needed, or assumed part of CAVEATS/etc.
		// Add LANG_VERSION here if direct parsing is desired:
		"LANG_VERSION:": &doc.LangVersion,
	}

	var inputLines *[]string = &doc.InputLines

	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentLines []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		upperTrimmedLine := strings.ToUpper(trimmedLine)

		foundSectionHeader := false

		// Check standard sections first
		for prefix, targetPtr := range sectionMap {
			if strings.HasPrefix(upperTrimmedLine, prefix) {
				if currentSection != nil {
					*currentSection = strings.TrimSpace(strings.Join(currentLines, "\n"))
				}
				currentSection = targetPtr
				// Use TrimPrefix which handles the prefix case-insensitively if needed, but ToUpper ensures match
				currentLines = []string{strings.TrimSpace(trimmedLine[len(prefix):])} // Get content after prefix
				inInputSection = false
				foundSectionHeader = true
				break
			}
		}
		if foundSectionHeader {
			continue
		}

		// Check INPUTS: specifically
		if strings.HasPrefix(upperTrimmedLine, "INPUTS:") {
			if currentSection != nil {
				*currentSection = strings.TrimSpace(strings.Join(currentLines, "\n"))
			}
			currentSection = nil // Clear target ptr
			inInputSection = true
			currentLines = []string{} // Reset lines for input parsing logic
			inputContent := strings.TrimSpace(trimmedLine[len("INPUTS:"):])
			if strings.ToLower(inputContent) != "none" && inputContent != "" {
				parseInputLine(inputContent, &doc) // Pass address of doc
				if inputLines != nil {
					*inputLines = append(*inputLines, inputContent)
				}
			}
			continue
		}

		// Append line to the current section or inputs
		if inInputSection {
			parseInputLine(trimmedLine, &doc) // Pass address of doc
			if inputLines != nil {
				*inputLines = append(*inputLines, line)
			}
		} else if currentSection != nil {
			currentLines = append(currentLines, line)
		}
	}

	// Finalize last section
	if currentSection != nil {
		*currentSection = strings.TrimSpace(strings.Join(currentLines, "\n"))
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[Error] Scanner error in parseDocstring: %v\n", err)
	}

	return doc
}

// parseInputLine parses a single line within the INPUTS section.
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
