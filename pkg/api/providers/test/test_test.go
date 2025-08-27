// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Adds a comprehensive, table-driven test for the Chat function to ensure it strictly rejects non-envelope prompts.
// filename: pkg/api/providers/test/test_test.go
// nlines: 89
// risk_rating: LOW

package test

import (
	"context"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// TestTestProvider_Chat verifies the core logic of the mock provider.
// It ensures that the provider strictly requires valid AEIOU envelopes
// and correctly handles different orchestration content.
func TestTestProvider_Chat(t *testing.T) {
	p := New()
	ctx := context.Background()

	// Helper to create a valid prompt envelope
	createPrompt := func(content string) string {
		env := &aeiou.Envelope{Orchestration: content}
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
			name:        "Strictly fails on non-envelope string",
			prompt:      "this is just a plain string",
			expectErr:   true,
			mustContain: "test provider requires a valid AEIOU envelope",
		},
		{
			name:      "Strictly fails on empty string",
			prompt:    "",
			expectErr: true,
		},
		{
			name:      "Strictly fails on incomplete envelope",
			prompt:    "<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:START>>>",
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
	parsed, err := aeiou.RobustParse(wrapped)
	if err != nil {
		t.Fatalf("aeiou.RobustParse failed to parse the wrapped response: %v\n---Envelope---\n%s", err, wrapped)
	}

	// 3. Verify the content was correctly placed in the ACTIONS section.
	if !strings.Contains(parsed.Actions, originalContent) {
		t.Errorf("ACTIONS section does not contain the original content.\nGot:\n%s", parsed.Actions)
	}
	if !strings.Contains(parsed.Actions, "command") {
		t.Error("ACTIONS section is missing the 'command' keyword.")
	}
}
