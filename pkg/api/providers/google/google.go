// NeuroScript Version: 0.6.0
// File version: 2
// Purpose: Implements a real AIProvider for Google AI models that makes live API calls.
// filename: pkg/api/providers/google/google.go
// nlines: 116
// risk_rating: MEDIUM

package google

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aprice2704/neuroscript/pkg/provider"
)

const apiBaseURL = "https://generativelanguage.googleapis.com/v1beta/models/"

// Provider implements the provider.AIProvider interface for Google's AI services.
type Provider struct{}

// New creates a new instance of the Google AI provider.
func New() *Provider {
	return &Provider{}
}

// --- Gemini API Request/Response Structures ---

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}
type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}
type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata geminiUsage       `json:"usageMetadata"`
	Error         *geminiError      `json:"error"`
}
type geminiCandidate struct {
	Content geminiContent `json:"content"`
}
type geminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
}
type geminiError struct {
	Message string `json:"message"`
}

// Chat sends a request to a Google AI model.
func (p *Provider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	if req.APIKey == "" {
		return nil, fmt.Errorf("Google provider requires an API key")
	}

	// 1. Construct the API endpoint URL.
	url := fmt.Sprintf("%s%s:generateContent?key=%s", apiBaseURL, req.ModelName, req.APIKey)

	// 2. Create the request body.
	requestPayload := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: req.Prompt}}},
		},
	}
	bodyBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 3. Create and send the HTTP request.
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// 4. Read and parse the response body.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	// 5. Check for API errors.
	if resp.StatusCode != http.StatusOK {
		if geminiResp.Error != nil {
			return nil, fmt.Errorf("google api error: %s (status %d)", geminiResp.Error.Message, resp.StatusCode)
		}
		return nil, fmt.Errorf("google api returned non-ok status: %s", resp.Status)
	}

	// 6. Extract the content and construct the final response.
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content found in google api response")
	}
	textContent := geminiResp.Candidates[0].Content.Parts[0].Text

	return &provider.AIResponse{
		TextContent:  textContent,
		InputTokens:  geminiResp.UsageMetadata.PromptTokenCount,
		OutputTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
	}, nil
}
