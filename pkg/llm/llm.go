// NeuroScript Version: 0.5.2
// File version: 15
// Purpose: Fixed minor compiler errors related to genai API usage and struct fields.
// filename: pkg/llm/llm.go
// nlines: 222
// risk_rating: MEDIUM

package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// concreteLLMClient is the internal implementation that talks to the Google Gemini API.
// It implements the interfaces.LLMClient interface.
type concreteLLMClient struct {
	genaiClient *genai.Client
	model       *genai.GenerativeModel
	logger      interfaces.Logger
}

// Statically assert that our concrete implementation satisfies the new interface.
var _ interfaces.LLMClient = (*concreteLLMClient)(nil)

// NewLLMClient creates and returns a client for interacting with the Gemini API.
func NewLLMClient(apiKey string, modelName string, logger interfaces.Logger) (interfaces.LLMClient, error) {
	if apiKey == "" {
		logger.Warn("NewLLMClient: API key is not set. Using internal NoOpLLMClient.")
		return newCoreInternalNoOpLLMClient(logger), nil
	}
	if modelName == "" {
		modelName = "gemini-1.5-flash-latest"
		logger.Debug("NewLLMClient: model name not provided, defaulting to", "model", modelName)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel(modelName)

	return &concreteLLMClient{
		genaiClient: client,
		model:       model,
		logger:      logger,
	}, nil
}

// Client returns the underlying *genai.Client for helpers that need it.
func (c *concreteLLMClient) Client() *genai.Client {
	return c.genaiClient
}

// Ask sends the conversation history to the Gemini API.
func (c *concreteLLMClient) Ask(ctx context.Context, turns []*interfaces.ConversationTurn) (*interfaces.ConversationTurn, error) {
	c.logger.Debug("Sending request to LLM...", "turn_count", len(turns))
	session := c.model.StartChat()
	session.History = convertTurnsToGenaiContents(turns)

	resp, err := session.SendMessage(ctx, genai.Text("continue"))
	if err != nil {
		return nil, fmt.Errorf("LLM SendMessage failed: %w", err)
	}

	return genaiContentToTurn(resp.Candidates[0].Content)
}

// AskWithTools sends the conversation history and available tools to the LLM.
func (c *concreteLLMClient) AskWithTools(ctx context.Context, turns []*interfaces.ConversationTurn, tools []interfaces.ToolDefinition) (*interfaces.ConversationTurn, []*interfaces.ToolCall, error) {
	c.logger.Debug("Sending request to LLM with tools...", "turn_count", len(turns), "tool_count", len(tools))

	genaiTools, err := convertToolsToGenai(tools)
	if err != nil {
		return nil, nil, fmt.Errorf("could not convert tools to genai format: %w", err)
	}
	c.model.Tools = genaiTools

	session := c.model.StartChat()
	session.History = convertTurnsToGenaiContents(turns)

	resp, err := session.SendMessage(ctx, genai.Text("continue"))
	if err != nil {
		return nil, nil, fmt.Errorf("LLM SendMessage with tools failed: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, nil, fmt.Errorf("LLM returned no candidates")
	}

	// Check for tool calls
	if len(resp.Candidates[0].Content.Parts) > 0 {
		if fc, ok := resp.Candidates[0].Content.Parts[0].(genai.FunctionCall); ok {
			toolCalls := []*interfaces.ToolCall{
				{
					Name:      types.FullName(fc.Name),
					Arguments: fc.Args,
				},
			}
			return nil, toolCalls, nil
		}
	}

	// If no tool calls, it's a regular response
	responseTurn, err := genaiContentToTurn(resp.Candidates[0].Content)
	if err != nil {
		return nil, nil, err
	}
	return responseTurn, nil, nil
}

// Embed generates vector embeddings for the given text.
func (c *concreteLLMClient) Embed(ctx context.Context, text string) ([]float32, error) {
	em := c.genaiClient.EmbeddingModel("text-embedding-004")
	res, err := em.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}
	if res == nil || res.Embedding == nil {
		return nil, fmt.Errorf("received nil embedding from API")
	}
	return res.Embedding.Values, nil
}

// --- NoOp Client Implementation ---

type coreInternalNoOpLLMClient struct {
	logger interfaces.Logger
}

func newCoreInternalNoOpLLMClient(logger interfaces.Logger) *coreInternalNoOpLLMClient {
	return &coreInternalNoOpLLMClient{logger: logger}
}

func (c *coreInternalNoOpLLMClient) Ask(ctx context.Context, history []*interfaces.ConversationTurn) (*interfaces.ConversationTurn, error) {
	return &interfaces.ConversationTurn{Role: interfaces.RoleAssistant, Content: ""}, nil
}

func (c *coreInternalNoOpLLMClient) AskWithTools(ctx context.Context, turns []*interfaces.ConversationTurn, tools []interfaces.ToolDefinition) (*interfaces.ConversationTurn, []*interfaces.ToolCall, error) {
	return &interfaces.ConversationTurn{Role: interfaces.RoleAssistant, Content: ""}, nil, nil
}

func (c *coreInternalNoOpLLMClient) Embed(ctx context.Context, text string) ([]float32, error) {
	return nil, nil
}

func (c *coreInternalNoOpLLMClient) Client() *genai.Client {
	return nil
}

// --- Helper Functions ---

func convertTurnsToGenaiContents(turns []*interfaces.ConversationTurn) []*genai.Content {
	history := make([]*genai.Content, 0, len(turns))
	for _, turn := range turns {
		var role string
		switch turn.Role {
		case interfaces.RoleUser:
			role = "user"
		case interfaces.RoleAssistant:
			role = "model"
		case interfaces.RoleTool:
			continue
		default:
			continue
		}

		parts := []genai.Part{genai.Text(turn.Content)}

		if turn.ToolCalls != nil {
			for _, tc := range turn.ToolCalls {
				parts = append(parts, genai.FunctionCall{Name: string(tc.Name), Args: tc.Arguments})
			}
		}

		if turn.Role == interfaces.RoleTool && turn.ToolResults != nil {
			for _, tr := range turn.ToolResults {
				parts = append(parts, genai.FunctionResponse{Name: tr.ID, Response: map[string]any{"result": tr.Result, "error": tr.Error}})
			}
		}

		history = append(history, &genai.Content{Parts: parts, Role: role})
	}
	return history
}

func genaiContentToTurn(content *genai.Content) (*interfaces.ConversationTurn, error) {
	if content == nil || len(content.Parts) == 0 {
		return nil, fmt.Errorf("LLM returned no content")
	}

	turn := &interfaces.ConversationTurn{Role: interfaces.RoleAssistant}

	for _, part := range content.Parts {
		switch p := part.(type) {
		case genai.Text:
			turn.Content = string(p)
		case genai.FunctionCall:
		default:
			return nil, fmt.Errorf("unhandled part type in LLM response: %T", p)
		}
	}
	return turn, nil
}

func convertToolsToGenai(tools []interfaces.ToolDefinition) ([]*genai.Tool, error) {
	if len(tools) == 0 {
		return nil, nil
	}

	genaiFuncs := make([]*genai.FunctionDeclaration, len(tools))
	for i, tool := range tools {
		schemaBytes, err := json.Marshal(tool.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema for tool %q: %w", tool.Name, err)
		}

		genaiSchema := &genai.Schema{}
		if err := json.Unmarshal(schemaBytes, genaiSchema); err != nil {
			return nil, fmt.Errorf("failed to unmarshal schema for tool %q: %w", tool.Name, err)
		}

		genaiFuncs[i] = &genai.FunctionDeclaration{
			Name:        string(tool.Name),
			Description: tool.Description,
			Parameters:  genaiSchema,
		}
	}

	return []*genai.Tool{{FunctionDeclarations: genaiFuncs}}, nil
}
