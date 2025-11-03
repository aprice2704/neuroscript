// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Populates the 'Description' field in the parsed capsule struct.
// filename: pkg/api/capsule.go
// nlines: 67
// risk_rating: MEDIUM

package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/metadata"
)

// NewCapsuleStore creates a new capsule store, optionally initialized
// with a set of registries. The store searches registries in the order
// they are provided.
func NewCapsuleStore(initial ...*CapsuleRegistry) *CapsuleStore {
	// We can cast the slice directly because api.CapsuleRegistry is an
	// alias for capsule.Registry.
	return capsule.NewStore(initial...)
}

// DefaultCapsuleRegistry returns the singleton registry that contains
// all built-in capsules that are embedded in the NeuroScript binary.
// This registry is populated by an init() function in the capsule package.
func DefaultCapsuleRegistry() *CapsuleRegistry {
	return capsule.DefaultRegistry()
}

// ParseCapsule parses a raw byte slice of capsule content, validates its
// metadata, and returns a populated api.Capsule struct.
//
// This function enforces all metadata requirements from
// 'capsule_metadata_requirements.md':
//   - Required fields: '::id', '::version', '::description'.
//   - '::id' format: Must start with 'capsule/' and contain only [a-z0-9_-].
//   - '::version' format: Must be a whole integer.
//
// It automatically calculates the SHA256, Size, and final ID.
func ParseCapsule(content []byte) (*Capsule, error) {
	// 1. Parse the metadata and content body using the internal metadata package
	//    We use bytes.NewReader as it implements the io.ReadSeeker interface
	//    required by the parser.
	reader := bytes.NewReader(content)
	meta, contentBody, _, err := metadata.ParseWithAutoDetect(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse capsule metadata: %w", err)
	}

	// 2. Validate required metadata fields
	extractor := metadata.NewExtractor(meta)
	if err := extractor.CheckRequired("id", "version", "description"); err != nil {
		return nil, fmt.Errorf("missing required metadata: %w", err)
	}

	// 3. Extract and validate core fields based on registry rules
	name := extractor.MustGet("id")
	version := extractor.MustGet("version")
	description := extractor.MustGet("description") // Get the description

	// Use the internal, re-exported validation function
	if err := capsule.ValidateName(name); err != nil {
		return nil, fmt.Errorf("invalid capsule '::id' %q: %w", name, err)
	}
	// Use the internal validation logic for versions
	if _, err := strconv.Atoi(version); err != nil {
		return nil, fmt.Errorf("invalid capsule '::version' %q: must be an integer", version)
	}

	// 4. Populate the capsule struct
	priority, _ := extractor.GetIntOr("priority", 100)
	cleanContent := string(bytes.TrimSpace(contentBody))

	// Calculate SHA256 and Size, mirroring internal logic
	sum := sha256.Sum256([]byte(cleanContent))
	sha := hex.EncodeToString(sum[:])

	cap := &Capsule{
		Name:        name,
		Version:     version,
		Description: description, // <-- FIX: Populate the field
		ID:          fmt.Sprintf("%s@%s", name, version),
		MIME:        extractor.GetOr("mime", "text/plain; charset=utf-8"),
		Content:     cleanContent,
		Priority:    priority,
		SHA256:      sha,
		Size:        len(cleanContent),
	}

	return cap, nil
}
