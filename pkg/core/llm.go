// filename: pkg/core/llm.go
package core

import (
	"bytes"
	"context" // Import context
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Configuration constants
const (
	defaultGeminiAPIEndpointBase = "https://generativelanguage.googleapis.com/v1beta/models/"
	defaultGeminiModel           = "gemini-1.5-flash-latest" // Default model
)

// --- LLM Client ---

// LLMClient holds configuration and state for interacting with the LLM API.
type LLMClient struct {
	apiKey     string
	modelName  string
	endpoint   string
	httpClient *http.Client
	logger     *log.Logger
}

// NewLLMClient creates a new client.
// Uses environment variable GEMINI_API_KEY if apiKey is empty.
func NewLLMClient(apiKey string, logger *log.Logger) *LLMClient {
	effectiveAPIKey := apiKey
	if effectiveAPIKey == "" {
		effectiveAPIKey = os.Getenv("GEMINI_API_KEY")
		if effectiveAPIKey == "" {
			logger.Printf("[WARN LLM] API Key not provided and GEMINI_API_KEY environment variable not set. LLM calls will fail.")
		} else {
			logger.Printf("[INFO LLM] Using API Key from GEMINI_API_KEY environment variable.")
		}
	} else {
		logger.Printf("[INFO LLM] Using provided API Key.")
	}

	return &LLMClient{
		apiKey:     effectiveAPIKey,
		modelName:  defaultGeminiModel,
		endpoint:   defaultGeminiAPIEndpointBase,
		httpClient: &http.Client{Timeout: 180 * time.Second}, // Increased timeout further
		logger:     logger,
	}
}

// SetModel allows changing the model used by the client.
func (c *LLMClient) SetModel(modelName string) {
	if modelName != "" {
		c.modelName = modelName
		c.logger.Printf("[INFO LLM] Set model to: %s", modelName)
	}
}

// --- API Interaction Methods ---

// CallLLMAgent performs a single turn of interaction with the LLM, potentially involving function calls.
// Takes conversation history and tool declarations. Returns the full API response.
func (c *LLMClient) CallLLMAgent(ctx context.Context, conversationHistory []GeminiContent, availableTools []GeminiTool) (*GeminiResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("LLM API key is not set")
	}

	// Construct the full API URL
	url := fmt.Sprintf("%s%s:generateContent?key=%s", c.endpoint, c.modelName, c.apiKey)

	// Prepare the request body
	requestBody := GeminiRequest{
		Contents: conversationHistory,
		Tools:    availableTools, // Include tool declarations
		// TODO: Add configurable GenerationConfig or SafetySettings here if needed
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("LLM client: failed to marshal request body: %w", err)
	}

	c.logger.Printf("[DEBUG LLM] Sending agent request to %s (model: %s)", url, c.modelName)
	// Avoid logging full body by default due to size/sensitivity
	// c.logger.Printf("[DEBUG LLM] Request Body: %s", string(requestBodyBytes))

	// Make the HTTP POST request with context
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("LLM client: failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json") // Explicitly accept JSON

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		// Handle context cancellation or timeout
		if ctx.Err() == context.Canceled {
			return nil, fmt.Errorf("LLM client: request canceled: %w", ctx.Err())
		} else if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("LLM client: request deadline exceeded: %w", ctx.Err())
		}
		return nil, fmt.Errorf("LLM client: failed to send request to Gemini API: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("LLM client: failed to read response body: %w", err)
	}

	c.logger.Printf("[DEBUG LLM] Received response status: %d", resp.StatusCode)
	// Avoid logging full body by default
	// c.logger.Printf("[DEBUG LLM] Response Body: %s", string(responseBodyBytes))

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM client: API request failed with status %d: %s", resp.StatusCode, string(responseBodyBytes))
	}

	// Parse the response JSON
	var responseBody GeminiResponse
	err = json.Unmarshal(responseBodyBytes, &responseBody)
	if err != nil {
		return nil, fmt.Errorf("LLM client: failed to unmarshal response body: %w\nRaw Body: %s", err, string(responseBodyBytes))
	}

	// --- Basic Validation/Logging of Response ---
	// (Moved logging logic inside, similar to previous version)
	if len(responseBody.Candidates) == 0 {
		if responseBody.PromptFeedback != nil && responseBody.PromptFeedback.BlockReason != "" {
			c.logger.Printf("[WARN LLM] Request blocked by safety settings: %s (Message: %s)",
				responseBody.PromptFeedback.BlockReason, responseBody.PromptFeedback.BlockReasonMessage)
		} else {
			c.logger.Printf("[WARN LLM] No candidates received in LLM response.")
		}
	} else {
		firstCandidate := responseBody.Candidates[0]
		c.logger.Printf("[DEBUG LLM] Candidate Finish Reason: %s", firstCandidate.FinishReason)
		if len(firstCandidate.Content.Parts) > 0 {
			firstPart := firstCandidate.Content.Parts[0]
			if firstPart.FunctionCall != nil {
				c.logger.Printf("[DEBUG LLM] Received FunctionCall request: %s", firstPart.FunctionCall.Name)
			} else if firstPart.Text != "" {
				snippet := firstPart.Text[:min(len(firstPart.Text), 80)]
				if len(firstPart.Text) > 80 {
					snippet += "..."
				}
				c.logger.Printf("[DEBUG LLM] Received Text response (snippet): %s", snippet)
			}
		} else {
			c.logger.Printf("[WARN LLM] First candidate has no parts.")
		}
	}

	return &responseBody, nil
}

// CallLLMAPI is the legacy function used by CALL LLM in standard NeuroScript.
// It performs a simple single-turn text generation.
func CallLLMAPI(prompt string) (string, error) {
	// This function can likely remain as is, using default model/endpoint.
	// Alternatively, it could be refactored to use an LLMClient instance
	// configured globally or passed via Interpreter, but let's keep it simple for now.

	apiKey := os.Getenv("GEMINI_API_KEY") // Keep using env var directly here
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	url := fmt.Sprintf("%s%s:generateContent?key=%s", defaultGeminiAPIEndpointBase, defaultGeminiModel, apiKey)

	requestBody := GeminiRequest{
		Contents: []GeminiContent{{Parts: []GeminiPart{{Text: prompt}}, Role: "user"}},
		// No tools for simple call
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Use a default client/timeout for this legacy function
	httpClient := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Gemini API: %w", err)
	}
	defer resp.Body.Close()

	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini API request failed with status %d: %s", resp.StatusCode, string(responseBodyBytes))
	}

	var responseBody GeminiResponse
	err = json.Unmarshal(responseBodyBytes, &responseBody)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w\nRaw Body: %s", err, string(responseBodyBytes))
	}

	// --- Extract Text (legacy logic) ---
	if len(responseBody.Candidates) == 0 {
		if responseBody.PromptFeedback != nil && responseBody.PromptFeedback.BlockReason != "" {
			return "", fmt.Errorf("request blocked by safety settings: %s", responseBody.PromptFeedback.BlockReason)
		}
		return "", fmt.Errorf("no candidates received from Gemini API")
	}
	if len(responseBody.Candidates[0].Content.Parts) > 0 {
		if responseBody.Candidates[0].Content.Parts[0].Text != "" {
			return responseBody.Candidates[0].Content.Parts[0].Text, nil
		}
		// Add specific check for unexpected function call in simple API
		if responseBody.Candidates[0].Content.Parts[0].FunctionCall != nil {
			return "", fmt.Errorf("simple CallLLMAPI received unexpected FunctionCall response")
		}
	}

	return "", fmt.Errorf("no text content found in the first candidate")
}
