// :: product: NS
// :: majorVersion: 1
// :: fileVersion: 1
// :: description: Minimal identity validation helpers for NeuroScript.
// :: latestChange: Created IsNodeID and IsEntityID helpers.
// :: filename: pkg/interfaces/fdm_node_entity.go
// :: serialization: go

package interfaces

import (
	"strings"
	"unicode"
)

// IsNodeID checks if a string follows the N_ prefix convention.
func IsNodeID(s string) bool {
	return strings.HasPrefix(s, "N_") && len(s) > 2
}

// IsEntityID checks if a string follows the E_ prefix convention.
func IsEntityID(s string) bool {
	return strings.HasPrefix(s, "E_") && len(s) > 2
}

// IsNSHandle checks if a string follows NeuroScript handle conventions.
func IsNSHandle(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Simplified check: alphanumeric and some separators
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '_' && r != '-' && r != ':' {
			return false
		}
	}
	return true
}
