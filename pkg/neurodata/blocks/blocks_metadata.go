// pkg/neurodata/blocks2/blocks_metadata.go
package blocks

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// LookForMetadata searches the raw content of a fenced block for common metadata patterns
// like '# key: value' or '-- key: value'.
// It returns a map of found metadata key-value pairs.
func LookForMetadata(rawContent string) (map[string]string, error) {
	fmt.Printf("[DEBUG BLOCKS2 Metadata] Starting LookForMetadata\n") // Debug
	metadata := make(map[string]string)
	// Regex for lines like: # id: value  OR -- id: value
	// Allows various keys (id, version, etc.) and captures the value.
	// Keys are restricted to common metadata words for safety/simplicity.
	// Value capture `(.*)` captures the rest of the line after the colon, needs trimming.
	metadataPattern := regexp.MustCompile(`^(?:#|--)\s*(id|version|lang_version|template|template_version|rendering_hint|canonical_format|status|dependsOn|howToUpdate)\s*:\s*(.*)`)

	scanner := bufio.NewScanner(strings.NewReader(rawContent))
	linesChecked := 0
	maxLinesToCheck := 10 // Limit how many lines we check for metadata at the start of the block content

	for scanner.Scan() && linesChecked < maxLinesToCheck {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Stop checking if we hit a line that doesn't look like a comment/metadata
		// unless it's an empty line (allow blank lines between metadata)
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") && !strings.HasPrefix(trimmedLine, "--") {
			fmt.Printf("[DEBUG BLOCKS2 Metadata] Non-metadata line encountered: %q. Stopping metadata scan.\n", trimmedLine) // Debug
			break
		}
		// If it's just an empty line, continue scanning
		if trimmedLine == "" {
			linesChecked++
			continue
		}

		matches := metadataPattern.FindStringSubmatch(line)
		// matches[0] is the full line match
		// matches[1] is the metadata key (e.g., "id", "version")
		// matches[2] is the metadata value (needs trimming)
		if len(matches) == 3 {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2]) // Trim captured value
			// Only add if key not already found (first occurrence wins)
			if _, exists := metadata[key]; !exists {
				metadata[key] = value
				fmt.Printf("[DEBUG BLOCKS2 Metadata] Found metadata: %s = %q\n", key, value) // Debug
			}
		}
		linesChecked++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[ERROR BLOCKS2 Metadata] Scanner error: %v\n", err) // Debug
		return metadata, fmt.Errorf("error scanning block content for metadata: %w", err)
	}

	fmt.Printf("[DEBUG BLOCKS2 Metadata] Finished LookForMetadata. Found: %v\n", metadata) // Debug
	return metadata, nil
}
