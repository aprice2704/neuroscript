// NeuroScript Version: 0.7.0
// File version: 10
// Purpose: Fixed a URL construction bug by removing the extra slash before the model name, which was causing a 404 error due to a duplicated 'models/' path segment.
// filename: pkg/provider/google/google.go
// nlines: 174
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

const apiBaseURL = "https://generativelanguage.googleapis.com/v1/"

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

type geminiListModelsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// ListModels fetches a list of available model names from the Google AI API.
func (p *Provider) ListModels(apiKey string) ([]string, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Google provider requires an API key to list models")
	}

	url := fmt.Sprintf("%smodels?key=%s", apiBaseURL, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request to list models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google api returned non-ok status for list models: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read list models response body: %w", err)
	}

	var listResp geminiListModelsResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal list models response: %w", err)
	}

	var modelNames []string
	for _, model := range listResp.Models {
		modelNames = append(modelNames, model.Name)
	}

	return modelNames, nil
}

// Chat sends a request to a Google AI model. It requires the prompt to be
// a valid AEIOU envelope and will fail if it cannot be parsed.
func (p *Provider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	if req.APIKey == "" {
		return nil, fmt.Errorf("Google provider requires an API key")
	}

	promptToParse := req.Prompt
	if markerPos := strings.LastIndex(req.Prompt, aeiou.Wrap(aeiou.SectionStart)); markerPos != -1 {
		promptToParse = req.Prompt[markerPos:]
	}

	_, _, err := aeiou.Parse(strings.NewReader(promptToParse))
	if err != nil {
		return nil, fmt.Errorf("google provider requires a valid AEIOU envelope prompt, but parsing failed: %w", err)
	}

	// FIX: The model name from the API already includes "models/".
	// The base URL should not also include it.
	url := fmt.Sprintf("%s%s:generateContent?key=%s", apiBaseURL, req.ModelName, req.APIKey)

	requestPayload := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: req.Prompt}}},
		},
	}
	bodyBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var geminiResp geminiResponse
		if json.Unmarshal(respBody, &geminiResp) == nil && geminiResp.Error != nil {
			return nil, fmt.Errorf("google api error: %s (status %d)", geminiResp.Error.Message, resp.StatusCode)
		}
		return nil, fmt.Errorf("google api returned non-ok status '%s' with body: %s", resp.Status, string(respBody))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal successful response body: %w", err)
	}

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
