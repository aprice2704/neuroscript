// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Provides unit tests for the capability parsing and building logic.
// filename: pkg/policy/capability/builder_test.go
// nlines: 80
// risk_rating: LOW

package capability

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseCapability(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected Capability
		wantErr  error
	}{
		{
			name:  "Full capability with single scope",
			input: "fs:read:/data/file.txt",
			expected: Capability{
				Resource: "fs",
				Verbs:    []string{"read"},
				Scopes:   []string{"/data/file.txt"},
			},
		},
		{
			name:  "Multiple verbs and scopes",
			input: "net:read,write:*.api.com:443,10.0.0.0/8",
			expected: Capability{
				Resource: "net",
				Verbs:    []string{"read", "write"},
				Scopes:   []string{"*.api.com:443", "10.0.0.0/8"},
			},
		},
		{
			name:  "Capability with no scopes",
			input: "model:use",
			expected: Capability{
				Resource: "model",
				Verbs:    []string{"use"},
				Scopes:   nil,
			},
		},
		{
			name:  "Capability with an empty scope part",
			input: "env:read:",
			expected: Capability{
				Resource: "env",
				Verbs:    []string{"read"},
				Scopes:   []string{""},
			},
		},
		{
			name:    "Error - Missing verb part",
			input:   "fs",
			wantErr: ErrInvalidCapabilityFormat,
		},
		{
			name:    "Error - Empty resource",
			input:   ":read",
			wantErr: ErrInvalidCapabilityFormat,
		},
		{
			name:    "Error - Empty verb",
			input:   "fs::/path",
			wantErr: ErrInvalidCapabilityFormat,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("Parse() = %v, want %v", got, tc.expected)
			}
		})
	}
}
