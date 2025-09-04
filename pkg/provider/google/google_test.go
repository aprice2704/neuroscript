// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Corrected the test by creating a fully-formed, valid envelope for the test case and refining the error assertion to correctly distinguish between expected network errors and unexpected parsing failures.
// filename: pkg/provider/google/google_test.go
// nlines: 80
// risk_rating: LOW

package google

import (
	"context"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// TestGoogleProvider_EnvelopeParsing verifies that the Google provider's Chat
// function correctly handles various prompt formats, including those with
// prepended text before a valid AEIOU envelope.
func TestGoogleProvider_EnvelopeParsing(t *testing.T) {
	p := New()
	ctx := context.Background()
	// FIX: A valid, parsable envelope must have an ACTIONS section.
	validEnvelope, _ := (&aeiou.Envelope{UserData: "test", Actions: "command endcommand"}).Compose()

	testCases := []struct {
		name        string
		prompt      string
		expectErr   bool
		mustContain string
	}{
		{
			name:        "Fails on plain string with no envelope",
			prompt:      "this is not an envelope",
			expectErr:   true,
			mustContain: "requires a valid AEIOU envelope",
		},
		{
			name:        "Fails on empty string",
			prompt:      "",
			expectErr:   true,
			mustContain: "requires a valid AEIOU envelope",
		},
		{
			name:        "Fails on incomplete envelope",
			prompt:      "<<<NSENV:V3:START>>>",
			expectErr:   true,
			mustContain: "requires a valid AEIOU envelope",
		},
		{
			name:        "Succeeds with prepended text before a valid envelope",
			prompt:      "This is some bootstrap text.\n\n" + validEnvelope,
			expectErr:   false, // This should now pass parsing without error.
			mustContain: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We provide a dummy API key because the validation happens before the key is used.
			req := provider.AIRequest{Prompt: tc.prompt, APIKey: "dummy-key"}
			_, err := p.Chat(ctx, req)

			if tc.expectErr {
				if err == nil {
					t.Fatal("Expected an error but got nil")
				}
				if !strings.Contains(err.Error(), tc.mustContain) {
					t.Errorf("Expected error to contain %q, but got: %v", tc.mustContain, err)
				}
			} else {
				// FIX: The assertion must check that the error, if one occurs, is NOT the parsing error.
				// We expect a failure later from the network call, but the parsing part should succeed.
				if err != nil && strings.Contains(err.Error(), "requires a valid AEIOU envelope") {
					t.Errorf("Test failed: parsing failed unexpectedly with error: %v", err)
				}
			}
		})
	}
}
