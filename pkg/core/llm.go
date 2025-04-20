// filename: pkg/core/llm.go
package core

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// LLMClient wraps the genai.Client and stores the configured model name.
type LLMClient struct {
	client    *genai.Client
	logger    *log.Logger
	modelName string
}

// NewLLMClient creates a new LLM client instance.
func NewLLMClient(apiKey string, modelName string, logger *log.Logger) *LLMClient {
	ctx := context.Background()
	effectiveAPIKey := apiKey
	if effectiveAPIKey == "" {
		effectiveAPIKey = os.Getenv("GEMINI_API_KEY")
	}

	if effectiveAPIKey == "" {
		logger.Println("[ERROR LLM] No API key provided via flag or GEMINI_API_KEY env var. LLM calls will fail.")
		return &LLMClient{client: nil, logger: logger}
	}

	// +++ ADDED: Log endpoint context +++
	// The client doesn't directly expose the URL, but we know the service
	logger.Println("[INFO LLM] Creating GenAI client for Google AI API...")

	client, err := genai.NewClient(ctx, option.WithAPIKey(effectiveAPIKey))
	if err != nil {
		logger.Printf("[ERROR LLM] Failed to create GenAI client: %v", err)
		return &LLMClient{client: nil, logger: logger}
	}
	logger.Println("[INFO LLM] GenAI client created successfully.")

	effectiveModelName := modelName
	if effectiveModelName == "" {
		effectiveModelName = "gemini-1.5-pro-latest" // Default model
		logger.Printf("[INFO LLM] No model name provided, using default: %s", effectiveModelName)
	} else {
		// Log the provided model name as INFO level as well
		logger.Printf("[INFO LLM] Configured to use model: %s", effectiveModelName)
	}

	return &LLMClient{
		client:    client,
		logger:    logger,
		modelName: effectiveModelName,
	}
}

// CallLLMAgent sends a request to the LLM agent model using StartChat.
func (c *LLMClient) CallLLMAgent(ctx context.Context, requestContext LLMRequestContext, tools []*genai.Tool) (*genai.GenerateContentResponse, error) {
	if c.client == nil {
		return nil, errors.New("LLM client not initialized (missing API key?)")
	}

	// Use DebugLog for detailed call info
	c.logger.Printf("[DEBUG LLM CallLLMAgent] Using model: %s", c.modelName)
	model := c.client.GenerativeModel(c.modelName)
	model.Tools = tools
	cs := model.StartChat()
	cs.History = requestContext.History

	parts := []genai.Part{}
	lastUserText := ""

	if len(requestContext.History) > 0 {
		lastMsg := requestContext.History[len(requestContext.History)-1]
		if lastMsg.Role == "user" && len(lastMsg.Parts) > 0 {
			if textPart, ok := lastMsg.Parts[0].(genai.Text); ok {
				lastUserText = string(textPart)
			}
		}
	}

	if len(requestContext.FileURIs) > 0 {
		c.logger.Printf("[DEBUG LLM CallLLMAgent] Adding %d file URI(s) to request.", len(requestContext.FileURIs))
		for _, uri := range requestContext.FileURIs {
			if uri == "" {
				c.logger.Println("[WARN LLM CallLLMAgent] Skipping empty file URI provided in context.")
				continue
			}
			c.logger.Printf("[DEBUG LLM CallLLMAgent] Adding FileData part for URI: %s", uri)
			parts = append(parts, genai.FileData{URI: uri})
		}
	}

	if lastUserText != "" {
		c.logger.Printf("[DEBUG LLM CallLLMAgent] Adding last user text part: %q", lastUserText)
		parts = append(parts, genai.Text(lastUserText))
	}

	if len(parts) == 0 {
		c.logger.Println("[WARN LLM CallLLMAgent] No parts constructed to send for this turn.")
		return nil, errors.New("no content parts to send in CallLLMAgent")
	}

	c.logger.Printf("[DEBUG LLM CallLLMAgent] Sending message via StartChat. Part count: %d", len(parts))
	resp, err := cs.SendMessage(ctx, parts...)
	if err != nil {
		c.logger.Printf("[ERROR LLM CallLLMAgent] SendMessage failed: %v", err)
		return nil, err
	}
	c.logger.Printf("[DEBUG LLM CallLLMAgent] Received response from SendMessage.")
	return resp, nil
}

// CallLLM is a simpler function for non-agent, stateless calls using GenerateContent.
func (c *LLMClient) CallLLM(ctx context.Context, prompt string) (string, error) {
	if c.client == nil {
		return "", errors.New("LLM client not initialized (missing API key?)")
	}
	c.logger.Printf("[DEBUG LLM CallLLM] Using model: %s", c.modelName) // Debug log
	model := c.client.GenerativeModel(c.modelName)
	c.logger.Printf("[DEBUG LLM CallLLM] Sending simple prompt: %q", prompt)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		c.logger.Printf("[ERROR LLM CallLLM] GenerateContent failed: %v", err)
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		part := resp.Candidates[0].Content.Parts[0]
		if text, ok := part.(genai.Text); ok {
			c.logger.Printf("[DEBUG LLM CallLLM] Received simple text response.")
			return string(text), nil
		}
	}
	c.logger.Printf("[WARN LLM CallLLM] Received non-text or empty response.")
	return "", errors.New("received non-text or empty response")
}

// CallLLMWithParts calls the LLM with specific parts (stateless).
func (c *LLMClient) CallLLMWithParts(ctx context.Context, partsToCall []genai.Part, tools []*genai.Tool) (*genai.GenerateContentResponse, error) {
	if c.client == nil {
		return nil, errors.New("LLM client not initialized")
	}
	c.logger.Printf("[DEBUG LLM CallLLMWithParts] Using model: %s", c.modelName) // Debug log
	model := c.client.GenerativeModel(c.modelName)
	model.Tools = tools
	c.logger.Printf("[DEBUG LLM CallLLMWithParts] Sending %d parts.", len(partsToCall))
	resp, err := model.GenerateContent(ctx, partsToCall...)
	if err != nil {
		c.logger.Printf("[ERROR LLM CallLLMWithParts] GenerateContent failed: %v", err)
		return nil, err
	}
	c.logger.Printf("[DEBUG LLM CallLLMWithParts] Received response.")
	return resp, nil
}
