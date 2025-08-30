// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Defines core interfaces and types for metadata parsing.
// filename: pkg/metadata/metadata.go
// nlines: 33
// risk_rating: LOW

package metadata

import (
	"fmt"
	"io"
)

// Store holds key-value metadata extracted from a file.
type Store map[string]string

// Parser defines the interface for a file parser that can extract a metadata Store.
type Parser interface {
	// Parse reads from r, extracts metadata, and returns the metadata store,
	// the content with metadata stripped, and any error encountered.
	Parse(r io.Reader) (Store, []byte, error)
}

// CheckRequired verifies that the metadata store contains all the specified required keys.
// It returns an error listing all missing keys.
func (s Store) CheckRequired(keys ...string) error {
	var missing []string
	for _, k := range keys {
		if _, ok := s[k]; !ok {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required metadata keys: %v", missing)
	}
	return nil
}
