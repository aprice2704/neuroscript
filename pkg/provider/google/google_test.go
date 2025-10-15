// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Replaced the mock parsing test with a live integration test that calls the real Google API. On failure, it now attempts to list available models for easier debugging.
// filename: pkg/provider/google/google_test.go
// nlines: 65
// risk_rating: HIGH

package google

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// TestGoogleProvider_LiveChat performs a live integration test against the
// Google Gemini API. It requires the GEMINI_API_KEY environment variable to be set.
func TestGoogleProvider_LiveChat(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: GEMINI_API_KEY is not set.")
	}

	p := New()
	ctx := context.Background()
	modelName := "models/gemini-2.5-flash" // A known stable model

	// Create a minimal, valid envelope for the request.
	envelope, err := (&aeiou.Envelope{
		UserData: "Live test prompt",
		Actions:  "command endcommand",
	}).Compose()
	if err != nil {
		t.Fatalf("Failed to compose valid envelope: %v", err)
	}

	req := provider.AIRequest{
		Prompt:    envelope,
		APIKey:    apiKey,
		ModelName: modelName,
	}

	// Attempt the chat call.
	_, err = p.Chat(ctx, req)

	// If there's an error, check if it's a "model not found" error.
	// If so, list the available models to help with debugging.
	if err != nil {
		if strings.Contains(err.Error(), "is not found for API version") || strings.Contains(err.Error(), "status 404") {
			t.Logf("Chat request failed with a model-not-found error: %v", err)
			t.Log("Attempting to list available models...")

			models, listErr := p.ListModels(apiKey)
			if listErr != nil {
				t.Fatalf("...failed to list models as well: %v", listErr)
			}

			t.Logf("...successfully retrieved models:\n%s", strings.Join(models, "\n"))
			t.Fail() // Ensure the original test still fails.
		} else {
			// For any other error, just fail the test.
			t.Fatalf("Chat request failed with an unexpected error: %v", err)
		}
	} else {
		t.Log("Live chat call successful.")
	}
}
