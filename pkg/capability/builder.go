// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Provides helper functions for building and parsing Capability structs.
// filename: pkg/policy/capability/builder.go
// nlines: 60
// risk_rating: MEDIUM

package capability

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidCapabilityFormat is returned when a capability string is malformed.
	ErrInvalidCapabilityFormat = errors.New("invalid capability format")
)

// New creates a new Capability struct with a single verb.
// It's a convenient helper for the most common use case.
func New(resource, verb string, scopes ...string) Capability {
	return Capability{
		Resource: resource,
		Verbs:    []string{verb},
		Scopes:   scopes,
	}
}

// NewWithVerbs creates a new Capability struct with multiple verbs.
func NewWithVerbs(resource string, verbs []string, scopes []string) Capability {
	return Capability{
		Resource: resource,
		Verbs:    verbs,
		Scopes:   scopes,
	}
}

// Parse creates a Capability struct from a string representation.
// The expected format is "resource:verb1,verb2:scope1,scope2,...".
// The scope part is optional.
func Parse(s string) (Capability, error) {
	parts := strings.SplitN(s, ":", 3)
	if len(parts) < 2 {
		return Capability{}, fmt.Errorf("%w: must have at least 'resource:verb'", ErrInvalidCapabilityFormat)
	}

	resource := strings.TrimSpace(parts[0])
	verbsStr := strings.TrimSpace(parts[1])

	if resource == "" {
		return Capability{}, fmt.Errorf("%w: resource cannot be empty", ErrInvalidCapabilityFormat)
	}
	if verbsStr == "" {
		return Capability{}, fmt.Errorf("%w: verbs cannot be empty", ErrInvalidCapabilityFormat)
	}

	verbs := strings.Split(verbsStr, ",")

	var scopes []string
	if len(parts) == 3 {
		scopes = strings.Split(parts[2], ",")
	}

	return Capability{
		Resource: resource,
		Verbs:    verbs,
		Scopes:   scopes,
	}, nil
}

// MustParse is a helper that wraps Parse and panics on error.
// It is intended for use in tests or variable initializations.
func MustParse(s string) Capability {
	c, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return c
}
