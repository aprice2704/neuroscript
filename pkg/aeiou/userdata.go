// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements USERDATA JSON schema validation for AEIOU v3.
// filename: aeiou/userdata.go
// nlines: 48
// risk_rating: LOW

package aeiou

import (
	"encoding/json"
	"fmt"
)

// UserDataPayload defines the required structure of the USERDATA JSON object.
type UserDataPayload struct {
	Subject string `json:"subject"`
	Brief   string `json:"brief,omitempty"`
	// Fields must be a JSON object, so we use json.RawMessage and check.
	Fields json.RawMessage `json:"fields"`
}

// ParseAndValidateUserData parses a JSON string and validates it against the
// minimal USERDATA schema required by AEIOU v3.
func ParseAndValidateUserData(data string) (*UserDataPayload, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal userdata: %v", ErrUserDataSchema, err)
	}

	if _, ok := raw["subject"]; !ok {
		return nil, fmt.Errorf("%w: missing required field 'subject'", ErrUserDataSchema)
	}
	if _, ok := raw["fields"]; !ok {
		return nil, fmt.Errorf("%w: missing required field 'fields'", ErrUserDataSchema)
	}

	var payload UserDataPayload
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal into target struct: %v", ErrUserDataSchema, err)
	}

	// Validate that 'fields' is a JSON object.
	if len(payload.Fields) == 0 || payload.Fields[0] != '{' {
		return nil, fmt.Errorf("%w: field 'fields' must be a JSON object", ErrUserDataSchema)
	}

	return &payload, nil
}
