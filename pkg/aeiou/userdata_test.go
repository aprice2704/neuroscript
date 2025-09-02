// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines tests for the AEIOU v3 USERDATA JSON schema validator.
// filename: aeiou/userdata_test.go
// nlines: 83
// risk_rating: LOW

package aeiou

import (
	"errors"
	"testing"
)

func TestParseAndValidateUserData(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectErrIs error
	}{
		{
			name:  "Valid minimal payload",
			input: `{"subject":"test","fields":{}}`,
		},
		{
			name:  "Valid full payload",
			input: `{"subject":"test","brief":"a test","fields":{"key":"val"}}`,
		},
		{
			name:        "Invalid JSON",
			input:       `{"subject":test}`,
			expectErrIs: ErrUserDataSchema,
		},
		{
			name:        "Missing subject",
			input:       `{"brief":"a test","fields":{}}`,
			expectErrIs: ErrUserDataSchema,
		},
		{
			name:        "Missing fields",
			input:       `{"subject":"test"}`,
			expectErrIs: ErrUserDataSchema,
		},
		{
			name:        "Subject is not a string",
			input:       `{"subject":123,"fields":{}}`,
			expectErrIs: ErrUserDataSchema,
		},
		{
			name:        "Brief is not a string",
			input:       `{"subject":"test","brief":123,"fields":{}}`,
			expectErrIs: ErrUserDataSchema,
		},
		{
			name:        "Fields is not an object",
			input:       `{"subject":"test","fields":"not-an-object"}`,
			expectErrIs: ErrUserDataSchema,
		},
		{
			name:        "Fields is an array",
			input:       `{"subject":"test","fields":[]}`,
			expectErrIs: ErrUserDataSchema,
		},
		{
			name:        "Empty input",
			input:       ``,
			expectErrIs: ErrUserDataSchema,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := ParseAndValidateUserData(tc.input)

			if tc.expectErrIs != nil {
				if !errors.Is(err, tc.expectErrIs) {
					t.Fatalf("Expected error target %v, got %v", tc.expectErrIs, err)
				}
				if payload != nil {
					t.Errorf("Expected nil payload on error, but got %+v", payload)
				}
				return
			}

			if err != nil {
				t.Fatalf("Validation failed unexpectedly: %v", err)
			}
			if payload == nil {
				t.Fatal("Expected a valid payload, but got nil")
			}
		})
	}
}
