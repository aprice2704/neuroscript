// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Corrected the parsing logic to find the last envelope in the prompt, allowing it to ignore any prepended bootstrap text from the connection manager.
// filename: pkg/provider/google/google.go
// nlines: 128
// risk_rating: MEDIUM

package google

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
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

// Chat sends a request to a Google AI model. It requires the prompt to be
// a valid AEIOU envelope and will fail if it cannot be parsed.
func (p *Provider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	if req.APIKey == "" {
		return nil, fmt.Errorf("Google provider requires an API key")
	}

	// The provider contract requires a valid AEIOU envelope. Parse it first.
	// FIX: The prompt may contain bootstrap text. Find the last occurrence of the
	// start marker to parse the real envelope, ignoring any preceding content.
	promptToParse := req.Prompt
	if markerPos := strings.LastIndex(req.Prompt, aeiou.Wrap(aeiou.SectionStart)); markerPos != -1 {
		promptToParse = req.Prompt[markerPos:]
	}

	_, _, err := aeiou.Parse(strings.NewReader(promptToParse))
	if err != nil {
		return nil, fmt.Errorf("google provider requires a valid AEIOU envelope prompt, but parsing failed: %w", err)
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
