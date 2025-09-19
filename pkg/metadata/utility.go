// NeuroScript Version: 0.3.0
// File version: 3
// Purpose: Provides utilities for extracting and validating metadata, exporting NormalizeKey.
// filename: pkg/metadata/utility.go
// nlines: 98
// risk_rating: LOW
package metadata

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Pre-defined sets of required metadata keys for different schemas.
var (
	// RequiredSourceFileKeys are the essential keys for any source file.
	RequiredSourceFileKeys = []string{"schema", "serialization", "fileversion", "description"}
	// RequiredCapsuleKeys are the essential keys for a capsule markdown file.
	RequiredCapsuleKeys = []string{"schema", "serialization", "id", "version", "description"}
)

// keyNormalizeRegex is used to remove characters that are ignored during key matching.
var keyNormalizeRegex = regexp.MustCompile(`[._-]+`)

// NormalizeKey implements the key matching rule from the spec:
// "the case of the letters, and the characters underscore, dot and dash (_.-) are ignored"
func NormalizeKey(key string) string {
	lower := strings.ToLower(key)
	return keyNormalizeRegex.ReplaceAllString(lower, "")
}

// Extractor provides a safe and convenient way to access values from a metadata Store.
// It automatically normalizes keys for lookups.
type Extractor struct {
	store Store
}

// NewExtractor creates a new extractor for a given metadata store.
func NewExtractor(s Store) *Extractor {
	// We create a new store with normalized keys for efficient lookups.
	normalizedStore := make(Store)
	for k, v := range s {
		normalizedStore[NormalizeKey(k)] = v
	}
	return &Extractor{store: normalizedStore}
}

// Get retrieves a value by key. Returns the value and true if the key exists.
func (e *Extractor) Get(key string) (string, bool) {
	val, ok := e.store[NormalizeKey(key)]
	return val, ok
}

// GetOr retrieves a value by key, returning the provided default value if the key is not found.
func (e *Extractor) GetOr(key string, defaultValue string) string {
	if val, ok := e.Get(key); ok {
		return val
	}
	return defaultValue
}

// MustGet retrieves a value by key. It returns the value, or an empty string if not found.
func (e *Extractor) MustGet(key string) string {
	return e.store[NormalizeKey(key)]
}

// GetInt retrieves a value by key and attempts to parse it as an integer.
func (e *Extractor) GetInt(key string) (int, bool, error) {
	val, ok := e.Get(key)
	if !ok {
		return 0, false, nil
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, true, fmt.Errorf("metadata key %q is not a valid integer: %w", key, err)
	}
	return i, true, nil
}

// GetIntOr retrieves a value by key and parses it as an integer, returning the
// provided default value if the key is not found. If the key is found but the value
// is not a valid integer, it returns an error.
func (e *Extractor) GetIntOr(key string, defaultValue int) (int, error) {
	i, ok, err := e.GetInt(key)
	if err != nil {
		return 0, err // Found the key, but it failed to parse.
	}
	if !ok {
		return defaultValue, nil // Key not found, return default.
	}
	return i, nil // Key found and parsed successfully.
}

// CheckRequired verifies that the underlying store contains all the specified required keys.
// It uses the same normalization logic for checking.
func (e *Extractor) CheckRequired(keys ...string) error {
	var missing []string
	for _, k := range keys {
		if _, ok := e.store[NormalizeKey(k)]; !ok {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required metadata keys: %v", missing)
	}
	return nil
}
