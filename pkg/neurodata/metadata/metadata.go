// Package metadata provides functions for extracting structured metadata
// (formatted as ':: key: value') from the beginning of text content,
// typically used for files or embedded code/data blocks.
package metadata

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// (Rest of the file remains the same as provided in the previous step)
// ...
var metadataPattern = regexp.MustCompile(`^\s*::\s+([a-zA-Z0-9_.-]+)\s*:\s*(.*)`)
var commentOrBlankPattern = regexp.MustCompile(`^\s*($|#|--)`)
var startsWithMetadataPrefix = regexp.MustCompile(`^\s*::`)

func Extract(content string) (map[string]string, error) {
	// ... implementation ...
	// ... rest of the function code ...
	metadata := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		if startsWithMetadataPrefix.MatchString(line) {
			matches := metadataPattern.FindStringSubmatch(line)
			if len(matches) == 3 {
				key := strings.TrimSpace(matches[1])
				value := strings.TrimSpace(matches[2])
				if _, exists := metadata[key]; !exists {
					metadata[key] = value
				}
				continue
			} else {
				continue // Skip malformed '::' lines
			}
		}

		if commentOrBlankPattern.MatchString(line) {
			continue
		}
		break // Stop at first non-metadata/comment/blank line
	}

	if err := scanner.Err(); err != nil {
		return metadata, fmt.Errorf("error scanning content for metadata: %w", err)
	}

	return metadata, nil
}
