// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Corrected to use req.ProviderParams and req.ModelName, fixing compiler errors.
// filename: pkg/provider/httpprovider/httpprovider.go
// nlines: 108
// risk_rating: HIGH

package httpprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/types"
)

// Provider implements the provider.AIProvider interface for any generic HTTP-based LLM API.
// It is configured at runtime using fields from the AgentModel.
type Provider struct {
	// Client is an HTTP client to use for requests.
	// If nil, http.DefaultClient is used.
	Client *http.Client
}

// New creates a new instance of the generic HTTP provider.
func New(opts ...Option) *Provider {
	p := &Provider{
		Client: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Option defines a function for configuring the Provider.
type Option func(*Provider)

// WithHTTPClient sets a custom http.Client for the provider.
func WithHTTPClient(client *http.Client) Option {
	return func(p *Provider) {
		p.Client = client
	}
}

// Chat sends a request to a generic HTTP API based on AgentModel configuration.
func (p *Provider) Chat(ctx context.Context, req types.AIRequest) (*types.AIResponse, error) {
	// 1. Extract HTTP configuration from the ProviderParams
	//    This is the FIX for "req.Model undefined".
	config, err := extractConfig(req.ProviderParams, req.AgentModelName)
	if err != nil {
		return nil, fmt.Errorf("httpprovider: %w", err)
	}

	// 2. Build the interpolation context
	//    This is the FIX for "req.Model.Model undefined".
	interpContext := map[string]string{
		"MODEL":  req.ModelName,
		"APIKEY": req.APIKey,
		"PROMPT": req.Prompt, // This is the full AEIOU envelope
	}

	// 3. Build Request Body
	// We pass the raw template and let the helper handle interpolation.
	bodyBytes, err := buildRequestBody(config.BodyTemplate, interpContext)
	if err != nil {
		return nil, fmt.Errorf("httpprovider: failed to build request body: %w", err)
	}

	// 4. Build Request Headers
	headers := buildRequestHeaders(config.Headers, interpContext)

	// 5. Create and Execute HTTP Request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, config.URL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("httpprovider: failed to create http request: %w", err)
	}
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("httpprovider: http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("httpprovider: failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse a standard error message from the provider
		errMsg := parseErrorResponse(respBody, config.ErrorPath)
		if errMsg != "" {
			return nil, fmt.Errorf("httpprovider: api error: %s (status %d)", errMsg, resp.StatusCode)
		}
		// Fallback
		return nil, fmt.Errorf("httpprovider: api returned non-ok status '%s' with body: %s", resp.Status, string(respBody))
	}

	// 6. Parse the successful response
	var data interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("httpprovider: failed to unmarshal response body as JSON: %w", err)
	}

	// 7. Extract the text content using the response path
	textContent, err := extractResponseText(data, config.ResponsePath)
	if err != nil {
		// If extraction fails, dump the whole response as a fallback
		textContent = fmt.Sprintf("httpprovider: %v. Full response: %s", err, string(respBody))
		// We don't return an error here, as we *did* get a response.
	}

	// 8. Return the standard AIResponse
	// TODO: Add token counting extraction if paths are provided in config
	return &types.AIResponse{
		TextContent: strings.TrimSpace(textContent),
	}, nil
}
