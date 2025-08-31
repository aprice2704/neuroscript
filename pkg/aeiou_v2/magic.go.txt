// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Provides a centralized helper function to create V2 magic strings, abstracting the constant and format.
// filename: neuroscript/pkg/aeiou/magic.go
// nlines: 36
// risk_rating: LOW

package aeiou

import (
	"encoding/json"
	"fmt"
)

const magicConstant = "NSENVELOPE_MAGIC_9E3B6F2D"
const protocolVersion = "V2"

// Wrap formats a string according to the NeuroScript V2 envelope protocol.
// The type should be a standard section type like "START", "END", "ACTIONS", etc.
// If a payload is provided, it must be a struct that can be marshaled to JSON.
func Wrap(sectionType SectionType, payload interface{}) (string, error) {
	// The new format is <<<MAGIC:VERSION:TYPE>>> or <<<MAGIC:VERSION:TYPE:JSON_PAYLOAD>>>
	if payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return "", fmt.Errorf("failed to marshal payload for section %s: %w", sectionType, err)
		}
		return fmt.Sprintf("<<<%s:%s:%s:%s>>>", magicConstant, protocolVersion, sectionType, string(payloadBytes)), nil
	}
	return fmt.Sprintf("<<<%s:%s:%s>>>", magicConstant, protocolVersion, sectionType), nil
}
