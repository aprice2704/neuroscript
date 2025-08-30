// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Adds tests to ensure the Google provider strictly fails when given a non-envelope prompt.
// filename: pkg/api/providers/google/google_test.go
// nlines: 55
// risk_rating: LOW

package google

import (
	"context"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/provider"
)

// TestGoogleProvider_StrictEnvelopeParsing verifies that the Google provider's
// Chat function fails correctly when the prompt is not a valid AEIOU envelope.
// These tests do not require a live API key as they should fail before any
// network request is made.
func TestGoogleProvider_StrictEnvelopeParsing(t *testing.T) {
	p := New()
	ctx := context.Background()

	testCases := []struct {
		name        string
		prompt      string
		expectErr   bool
		mustContain string
	}{
		{
			name:        "Fails on plain string",
			prompt:      "this is not an envelope",
			expectErr:   true,
			mustContain: "google provider requires a valid AEIOU envelope",
		},
		{
			name:        "Fails on empty string",
			prompt:      "",
			expectErr:   true,
			mustContain: "google provider requires a valid AEIOU envelope",
		},
		{
			name:        "Fails on incomplete envelope",
			prompt:      "<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:START>>>",
			expectErr:   true,
			mustContain: "google provider requires a valid AEIOU envelope",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We provide a dummy API key because the validation happens before the key is used.
			req := provider.AIRequest{Prompt: tc.prompt, APIKey: "dummy-key"}
			_, err := p.Chat(ctx, req)

			if !tc.expectErr {
				t.Fatal("Expected an error, but got nil")
			}
			if err == nil {
				t.Fatal("Expected an error but got nil")
			}
			if !strings.Contains(err.Error(), tc.mustContain) {
				t.Errorf("Expected error to contain %q, but got: %v", tc.mustContain, err)
			}
		})
	}
}
