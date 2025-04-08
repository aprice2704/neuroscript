// Package blocks extracts fenced code blocks (```lang ... ```) from text content.
// This file specifically handles extracting metadata comments from the *start*
// of the extracted block's raw content using the shared metadata package.
package blocks

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/neurodata/metadata"
)

// LookForMetadata uses the shared metadata extractor to find metadata
// at the beginning of a block's raw content.
func LookForMetadata(rawContent string) (map[string]string, error) {
	// (Rest of function is unchanged from previous correct version)
	fmt.Printf("[DEBUG BLOCKS Metadata] Calling shared metadata.Extract\n")
	metaMap, err := metadata.Extract(rawContent)
	if err != nil {
		fmt.Printf("[ERROR BLOCKS Metadata] Error calling metadata.Extract: %v\n", err)
		return make(map[string]string), fmt.Errorf("failed to extract metadata from block content: %w", err)
	}
	fmt.Printf("[DEBUG BLOCKS Metadata] Finished metadata.Extract. Found: %v\n", metaMap)
	return metaMap, nil
}
