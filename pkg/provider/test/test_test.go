// NeuroScript Version: 0.7.0
// File version: 6
// Purpose: Updated tests to confirm the provider correctly ignores prepended text and finds the final valid envelope in the prompt.
// filename: pkg/provider/test/test_test.go
// nlines: 100
// risk_rating: LOW

package test

import (
	"context"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// TestTestProvider_EnvelopeHandling verifies the core logic of the mock provider.
// It ensures that the provider correctly handles prompts with and without prepended
// text, while still rejecting prompts that lack a valid AEIOU envelope.
func TestTestProvider_EnvelopeHandling(t *testing.T) {
	p := New()
	ctx := context.Background()

	// Helper to create a valid prompt envelope
	createPrompt := func(content string) string {
		env := &aeiou.Envelope{UserData: content, Actions: "command endcommand"}
		payload, _ := env.Compose()
		return payload
	}

	testCases := []struct {
		name        string
		prompt      string
		expectErr   bool
		mustContain string // For successful responses or error messages
	}{
		{
			name:        "Valid prompt for 'ping'",
			prompt:      createPrompt("ping"),
			expectErr:   false,
			mustContain: "test_provider_ok:pong",
		},
		{
			name:        "Valid prompt for 'llm'",
			prompt:      createPrompt("What is a large language model?"),
			expectErr:   false,
			mustContain: "A large language model is a neural network",
		},
		{
			name:        "Succeeds with prepended text before a valid envelope",
			prompt:      "Some leading text...\n" + createPrompt("ping"),
			expectErr:   false,
			mustContain: "test_provider_ok:pong",
		},
		{
			name:        "Fails on non-envelope string",
			prompt:      "this is just a plain string",
			expectErr:   true,
			mustContain: "test provider requires a valid AEIOU envelope",
		},
		{
			name:      "Fails on empty string",
			prompt:    "",
			expectErr: true,
		},
		{
			name:      "Fails on incomplete envelope",
			prompt:    "<<<NSENV:V3:START>>>",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := provider.AIRequest{Prompt: tc.prompt}
			resp, err := p.Chat(ctx, req)

			if tc.expectErr {
				if err == nil {
					t.Fatal("Expected an error, but got nil")
				}
				if tc.mustContain != "" && !strings.Contains(err.Error(), tc.mustContain) {
					t.Errorf("Expected error to contain %q, but got: %v", tc.mustContain, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}

			if !strings.Contains(resp.TextContent, tc.mustContain) {
				t.Errorf("Response does not contain expected text %q.\nGot:\n%s", tc.mustContain, resp.TextContent)
			}
		})
	}
}

// TestTestProvider_WrapResponseInAEIOU verifies the internal helper for wrapping
// responses produces a valid and parsable AEIOU envelope.
func TestTestProvider_WrapResponseInAEIOU(t *testing.T) {
	originalContent := "hello, this is the test content"

	// 1. Generate the envelope using the helper.
	wrapped, err := WrapResponseInAEIOU(originalContent)
	if err != nil {
		t.Fatalf("WrapResponseInAEIOU failed: %v", err)
	}

	// 2. Parse the generated envelope to ensure it's valid.
	parsed, _, err := aeiou.Parse(strings.NewReader(wrapped))
	if err != nil {
		t.Fatalf("aeiou.Parse failed to parse the wrapped response: %v\n---Envelope---\n%s", err, wrapped)
	}

	// 3. Verify the content was correctly placed in the ACTIONS section.
	if !strings.Contains(parsed.Actions, originalContent) {
		t.Errorf("ACTIONS section does not contain the original content.\nGot:\n%s", parsed.Actions)
	}
	if !strings.Contains(parsed.Actions, "command") {
		t.Error("ACTIONS section is missing the 'command' keyword.")
	}
}
