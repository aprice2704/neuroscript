package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time" // Keep for potential future use (e.g., timeouts)
)

const (
	geminiAPIEndpointBase = "https://generativelanguage.googleapis.com/v1beta/models/"
	geminiModel           = "gemini-1.5-flash-latest" // As requested
)

// --- Request Structures --- based on documentation

type GeminiRequest struct {
	Contents         []GeminiContent         `json:"contents"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []GeminiSafetySetting   `json:"safetySettings,omitempty"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"` // Typically "user" for single turn
}

type GeminiPart struct {
	Text string `json:"text"`
	// Add other part types like inlineData if needed later
}

// Optional: Add GenerationConfig and SafetySettings if needed
type GeminiGenerationConfig struct {
	Temperature     float32  `json:"temperature,omitempty"`
	TopP            float32  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}
type GeminiSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"` // e.g., "BLOCK_MEDIUM_AND_ABOVE"
}

// --- Response Structures --- based on documentation

type GeminiResponse struct {
	Candidates     []GeminiCandidate     `json:"candidates"`
	PromptFeedback *GeminiPromptFeedback `json:"promptFeedback,omitempty"`
}

type GeminiCandidate struct {
	Content       GeminiContent        `json:"content"` // Model's response content
	FinishReason  string               `json:"finishReason,omitempty"`
	Index         int                  `json:"index"`
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings,omitempty"`
}

type GeminiSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"` // e.g., "NEGLIGIBLE"
}

type GeminiPromptFeedback struct {
	BlockReason   string               `json:"blockReason,omitempty"`
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings,omitempty"`
}

// CallLLMAPI makes a REST call to the specified Gemini model.
func CallLLMAPI(prompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	// Construct the URL
	url := fmt.Sprintf("%s%s:generateContent?key=%s", geminiAPIEndpointBase, geminiModel, apiKey)

	// Prepare the request body
	requestBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
				// Role: "user", // Optional, defaults might be okay
			},
		},
		// Add GenerationConfig or SafetySettings here if needed
		// GenerationConfig: &GeminiGenerationConfig{ Temperature: 0.7 },
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Make the HTTP POST request
	httpClient := &http.Client{Timeout: 60 * time.Second} // Add a timeout
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Gemini API: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini API request failed with status %d: %s", resp.StatusCode, string(responseBodyBytes))
	}

	// Parse the response JSON
	var responseBody GeminiResponse
	err = json.Unmarshal(responseBodyBytes, &responseBody)
	if err != nil {
		// Include raw body for debugging if unmarshal fails
		return "", fmt.Errorf("failed to unmarshal response body: %w\nRaw Body: %s", err, string(responseBodyBytes))
	}

	// Check for safety blocks or empty candidates
	if len(responseBody.Candidates) == 0 {
		if responseBody.PromptFeedback != nil && responseBody.PromptFeedback.BlockReason != "" {
			return "", fmt.Errorf("request blocked by safety settings: %s", responseBody.PromptFeedback.BlockReason)
		}
		return "", fmt.Errorf("no candidates received from Gemini API")
	}

	// Extract the text from the first candidate
	// Assume the response structure has content.parts as an array and we want the first part's text.
	if len(responseBody.Candidates[0].Content.Parts) > 0 {
		generatedText := responseBody.Candidates[0].Content.Parts[0].Text
		return generatedText, nil
	}

	// Handle cases where no text part is returned
	return "", fmt.Errorf("no text content found in the first candidate")
}
