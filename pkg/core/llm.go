// NeuroScript Version: 0.3.1
// File version: 0.1.6 // Changed INFO logs to DEBUG
// filename: pkg/core/llm.go
package core

import (
	"context"
	"fmt"
	"strings" // For processing model response

	// logging.Logger is from "github.com/aprice2704/neuroscript/pkg/logging"
	// coreNoOpLogger is from "pkg/core/utils.go" (same package 'core')
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid" // For generating IDs for ToolCalls
	"google.golang.org/api/option"
)

// LLMClient interface definition is in llm_types.go

// --- Concrete LLM Client Implementation ---

// concreteLLMClient represents the actual implementation that talks to an LLM API.
type concreteLLMClient struct {
	apiKey      string
	apiHost     string
	logger      logging.Logger // Use the logging.Logger interface
	enabled     bool
	modelID     string
	genaiClient *genai.Client
}

// Ensure concreteLLMClient implements the LLMClient interface.
var _ LLMClient = (*concreteLLMClient)(nil)

// NewLLMClient creates a new instance of the LLM client.
// It acts as a factory, returning an LLMClient interface type.
// If enabled is false, or essential parameters are missing, it returns an internal NoOp client.
func NewLLMClient(apiKey, apiHost, modelID string, logger logging.Logger, enabled bool) LLMClient {
	if logger == nil {
		// Use the coreNoOpLogger defined in utils.go (accessible as it's in the same package 'core')
		logger = &coreNoOpLogger{} // This is from pkg/core/utils.go
		logger.Debug("NewLLMClient: nil logger provided, using internal coreNoOpLogger.")
	}

	// Use the internal newCoreInternalNoOpLLMClient defined at the end of this file.
	noopClientFactory := func() LLMClient {
		return newCoreInternalNoOpLLMClient(logger)
	}

	if !enabled {
		logger.Debug("LLM client explicitly disabled. Using internal NoOpLLMClient.") // Changed from Info
		return noopClientFactory()
	}

	logger.Debug("Attempting to create concrete LLM client.", "host", apiHost, "model_id", modelID, "enabled", enabled) // Changed from Info
	if apiKey == "" {
		logger.Error("API Key is missing for enabled LLM client.")
		logger.Warn("API Key missing, falling back to internal NoOpLLMClient.")
		return noopClientFactory()
	}
	if modelID == "" {
		logger.Error("ModelID is missing for enabled LLM client.")
		logger.Warn("ModelID missing, falling back to internal NoOpLLMClient.")
		return noopClientFactory()
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		logger.Error("Failed to initialize GenAI client", "error", err)
		logger.Warn("GenAI client init failed, falling back to internal NoOpLLMClient.")
		return noopClientFactory()
	}
	logger.Debug("Concrete LLM client created successfully.") // Changed from Info
	return &concreteLLMClient{
		apiKey:      apiKey,
		apiHost:     apiHost,
		logger:      logger, // Store the passed-in or coreNoOpLogger
		enabled:     true,
		modelID:     modelID,
		genaiClient: client,
	}
}

// convertCoreTurnsToGenaiContents converts NeuroScript ConversationTurns to genai.Content format.
// (No changes from file version 0.1.4 for this function)
func convertCoreTurnsToGenaiContents(turns []*ConversationTurn, logger logging.Logger) []*genai.Content {
	var genaiContents []*genai.Content
	turnRoleMap := map[Role]string{
		RoleUser:      "user",
		RoleAssistant: "model",
		RoleSystem:    "model", // Or handle as SystemInstruction
		RoleTool:      "function",
	}

	for _, turn := range turns {
		if turn == nil {
			continue
		}
		genaiRole, ok := turnRoleMap[turn.Role]
		if !ok {
			logger.Warnf("Unknown role in ConversationTurn: %s. Defaulting to 'user'.", turn.Role)
			genaiRole = "user"
		}
		if turn.Role == RoleSystem {
			logger.Debugf("System role in turn converted to '%s' for genai.Content history. Consider using SystemInstruction for GenAI models.", genaiRole)
		}

		var parts []genai.Part
		if turn.Content != "" {
			parts = append(parts, genai.Text(turn.Content))
		}

		if turn.Role == RoleAssistant && len(turn.ToolCalls) > 0 {
			for _, tc := range turn.ToolCalls {
				if tc != nil {
					parts = append(parts, genai.FunctionCall{Name: tc.Name, Args: tc.Arguments})
				}
			}
		}

		if turn.Role == RoleTool && len(turn.ToolResults) > 0 {
			for _, tr := range turn.ToolResults {
				if tr != nil {
					var funcName string
					// Critical: This logic needs to be robust. The name of the function that was called MUST be used.
					// If ToolResult doesn't reliably provide this, it's a systemic issue.
					// For now, using a placeholder that highlights this problem.
					// A potential solution is to ensure that the system which executes the tool and creates the
					// ToolResult also includes the original ToolCall's Name (or the function name directly)
					// in the ToolResult struct or as context when creating the ToolResult turn.
					// For example, if ToolResult had a `FunctionName string` field: funcName = tr.FunctionName
					funcName = "placeholder_MUST_BE_ACTUAL_FUNCTION_NAME"
					if tr.ID != "" { // If ID might be used to find name elsewhere
						logger.Warnf("CRITICAL: Function name for ToolResult ID '%s' is using a placeholder '%s'. This must be the actual function name that was called.", tr.ID, funcName)
					} else {
						logger.Warnf("CRITICAL: Function name for ToolResult is using a placeholder '%s' due to missing context. This must be the actual function name.", funcName)
					}

					responseMap := map[string]interface{}{"output": tr.Result}
					if tr.Error != "" {
						responseMap["error_message"] = tr.Error
					}
					parts = append(parts, genai.FunctionResponse{Name: funcName, Response: responseMap})
				}
			}
		}
		if len(parts) > 0 {
			genaiContents = append(genaiContents, &genai.Content{Role: genaiRole, Parts: parts})
		}
	}
	return genaiContents
}

// Ask sends a request to the actual LLM API (Google GenAI).
// (No changes from file version 0.1.4 for this function's core logic)
func (c *concreteLLMClient) Ask(ctx context.Context, turns []*ConversationTurn) (*ConversationTurn, error) {
	c.logger.Debug("ConcreteLLMClient Ask called", "turn_count", len(turns), "model_id", c.modelID)
	if !c.enabled || c.genaiClient == nil {
		c.logger.Warn("Ask called on disabled or uninitialized concrete LLM client.")
		return &ConversationTurn{
			Role:       RoleAssistant,
			Content:    "LLM client not enabled or initialized.",
			TokenUsage: TokenUsageMetrics{},
		}, nil
	}

	model := c.genaiClient.GenerativeModel(c.modelID)
	// TODO: Apply GenerationConfig & SafetySettings from AIWorkerDefinition or other context.

	genaiHistory := convertCoreTurnsToGenaiContents(turns, c.logger)

	if len(genaiHistory) == 0 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "Ask requires at least one turn (the user prompt)", ErrInvalidArgument)
	}

	chat := model.StartChat()
	if len(genaiHistory) > 1 {
		chat.History = genaiHistory[:len(genaiHistory)-1]
	}

	lastTurnInput := genaiHistory[len(genaiHistory)-1]
	if lastTurnInput.Role != "user" && lastTurnInput.Role != "function" {
		c.logger.Warnf("Last turn input to GenAI model was role '%s', expected 'user' or 'function'. This might lead to unexpected behavior with model %s.", lastTurnInput.Role, c.modelID)
	}

	resp, err := chat.SendMessage(ctx, lastTurnInput.Parts...)

	if err != nil {
		c.logger.Error("GenAI SendMessage failed", "error", err, "model_id", c.modelID)
		return nil, NewRuntimeError(ErrorCodeLLMError, fmt.Sprintf("LLM communication error with model %s", c.modelID), err)
	}
	if resp == nil {
		c.logger.Error("GenAI SendMessage returned a nil response object.", "model_id", c.modelID)
		return nil, NewRuntimeError(ErrorCodeLLMError, "LLM returned nil response object", ErrLLMError)
	}

	responseTurn := &ConversationTurn{
		Role:       RoleAssistant,
		Content:    "",
		TokenUsage: TokenUsageMetrics{},
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		candidate := resp.Candidates[0]
		var responseTextBuilder strings.Builder
		for _, part := range candidate.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				responseTextBuilder.WriteString(string(txt))
			}
		}
		responseTurn.Content = responseTextBuilder.String()
	} else {
		errMsg := "LLM returned no valid candidates or content."
		// Using .String() for genai.FinishReason as it's an enum.
		if resp.PromptFeedback != nil && resp.PromptFeedback.BlockReason != genai.BlockReasonUnspecified {
			errMsg = fmt.Sprintf("Prompt blocked by API (Reason: %s, Model: %s)", resp.PromptFeedback.BlockReason.String(), c.modelID)
		} else if len(resp.Candidates) > 0 && resp.Candidates[0].FinishReason != genai.FinishReasonUnspecified {
			errMsg = fmt.Sprintf("LLM returned no content. FinishReason: %s (Model: %s)", resp.Candidates[0].FinishReason.String(), c.modelID)
		}
		c.logger.Warn(errMsg, "model_id", c.modelID)
		if resp.PromptFeedback != nil && resp.PromptFeedback.BlockReason != genai.BlockReasonUnspecified {
			return responseTurn, NewRuntimeError(ErrorCodeLLMError, errMsg, ErrLLMError)
		}
	}

	if resp.UsageMetadata != nil {
		responseTurn.TokenUsage.InputTokens = int64(resp.UsageMetadata.PromptTokenCount)
		responseTurn.TokenUsage.OutputTokens = int64(resp.UsageMetadata.CandidatesTokenCount)
		responseTurn.TokenUsage.TotalTokens = int64(resp.UsageMetadata.TotalTokenCount)
		c.logger.Debugf("Token usage from GenAI (Ask): Input=%d, Output=%d, Total=%d (Model: %s)",
			responseTurn.TokenUsage.InputTokens, responseTurn.TokenUsage.OutputTokens, responseTurn.TokenUsage.TotalTokens, c.modelID)
	} else {
		c.logger.Warnf("GenAI response missing UsageMetadata. Token counts will be zero/approximated. (Model: %s)", c.modelID)
	}

	return responseTurn, nil
}

// AskWithTools sends a request with tools to the actual LLM API.
// (No changes from file version 0.1.4 for this function's core logic, other than FinishReason fix)
func (c *concreteLLMClient) AskWithTools(ctx context.Context, turns []*ConversationTurn, tools []ToolDefinition) (*ConversationTurn, []*ToolCall, error) {
	c.logger.Debug("ConcreteLLMClient AskWithTools called", "turn_count", len(turns), "tool_count", len(tools), "model_id", c.modelID)
	if !c.enabled || c.genaiClient == nil {
		c.logger.Warn("AskWithTools called on disabled or uninitialized concrete LLM client.")
		return &ConversationTurn{
			Role:       RoleAssistant,
			Content:    "LLM client not enabled or initialized (tools).",
			TokenUsage: TokenUsageMetrics{},
		}, nil, nil
	}

	model := c.genaiClient.GenerativeModel(c.modelID)
	var genaiAPITools []*genai.Tool
	if len(tools) > 0 {
		genaiAPITools = make([]*genai.Tool, 0, len(tools))
		for _, tDef := range tools {
			if tDef.Name == "" {
				c.logger.Warnf("Skipping tool with empty name in AskWithTools.")
				continue
			}
			var paramsSchema *genai.Schema
			if tDef.InputSchema != nil {
				if gs, okGs := tDef.InputSchema.(*genai.Schema); okGs {
					paramsSchema = gs
				} else {
					c.logger.Debugf("Tool %s: InputSchema to genai.Schema conversion is currently a placeholder or requires InputSchema to be *genai.Schema.", tDef.Name)
					paramsSchema = &genai.Schema{Type: genai.TypeObject} // Default to empty object if not already a *genai.Schema
				}
			}
			genaiAPITools = append(genaiAPITools, &genai.Tool{
				FunctionDeclarations: []*genai.FunctionDeclaration{{
					Name:        tDef.Name,
					Description: tDef.Description,
					Parameters:  paramsSchema,
				}},
			})
		}
		if len(genaiAPITools) > 0 {
			model.Tools = genaiAPITools
		} else {
			c.logger.Debug("AskWithTools: No valid tools converted for GenAI model.")
		}
	}

	genaiHistory := convertCoreTurnsToGenaiContents(turns, c.logger)
	chat := model.StartChat()
	if len(genaiHistory) > 1 {
		chat.History = genaiHistory[:len(genaiHistory)-1]
	}

	var lastTurnInputParts []genai.Part
	if len(genaiHistory) > 0 {
		lastTurnInput := genaiHistory[len(genaiHistory)-1]
		lastTurnInputParts = lastTurnInput.Parts
		if lastTurnInput.Role != "user" && lastTurnInput.Role != "function" {
			c.logger.Warnf("Last turn input to GenAI model (tools) was role '%s', expected 'user' or 'function'. (Model: %s)", lastTurnInput.Role, c.modelID)
		}
	} else {
		return nil, nil, NewRuntimeError(ErrorCodeArgMismatch, "AskWithTools requires at least one turn", ErrInvalidArgument)
	}

	resp, err := chat.SendMessage(ctx, lastTurnInputParts...)
	if err != nil {
		c.logger.Error("GenAI SendMessage (with tools) failed", "error", err, "model_id", c.modelID)
		return nil, nil, NewRuntimeError(ErrorCodeLLMError, fmt.Sprintf("LLM communication error with tools (model %s)", c.modelID), err)
	}
	if resp == nil {
		c.logger.Error("GenAI SendMessage (with tools) returned a nil response object.", "model_id", c.modelID)
		return nil, nil, NewRuntimeError(ErrorCodeLLMError, "LLM returned nil response object (with tools)", ErrLLMError)
	}

	responseTurn := &ConversationTurn{
		Role:       RoleAssistant,
		Content:    "",
		ToolCalls:  []*ToolCall{}, // Ensure it's non-nil
		TokenUsage: TokenUsageMetrics{},
	}
	var coreToolCallsOutput []*ToolCall // Explicitly define so it can be nil if no calls

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		candidate := resp.Candidates[0]
		var responseTextBuilder strings.Builder
		for _, part := range candidate.Content.Parts {
			switch p := part.(type) {
			case genai.Text:
				responseTextBuilder.WriteString(string(p))
			case genai.FunctionCall:
				coreTC := &ToolCall{
					ID:        uuid.NewString(),
					Name:      p.Name,
					Arguments: p.Args,
				}
				coreToolCallsOutput = append(coreToolCallsOutput, coreTC)
			default:
				c.logger.Warnf("Unhandled part type in AskWithTools response: %T", p)
			}
		}
		responseTurn.Content = responseTextBuilder.String()
		if len(coreToolCallsOutput) > 0 {
			responseTurn.ToolCalls = coreToolCallsOutput
		}
	} else {
		errMsg := "LLM returned no valid candidates or content (with tools)."
		if resp.PromptFeedback != nil && resp.PromptFeedback.BlockReason != genai.BlockReasonUnspecified {
			errMsg = fmt.Sprintf("Prompt blocked by API (Reason: %s, Model: %s, WithTools)", resp.PromptFeedback.BlockReason.String(), c.modelID)
		} else if len(resp.Candidates) > 0 && resp.Candidates[0].FinishReason != genai.FinishReasonUnspecified { // Corrected comparison
			errMsg = fmt.Sprintf("LLM returned no content (with tools). FinishReason: %s (Model: %s)", resp.Candidates[0].FinishReason.String(), c.modelID)
		}
		c.logger.Warn(errMsg, "model_id", c.modelID)
		if resp.PromptFeedback != nil && resp.PromptFeedback.BlockReason != genai.BlockReasonUnspecified {
			return responseTurn, coreToolCallsOutput, NewRuntimeError(ErrorCodeLLMError, errMsg, ErrLLMError)
		}
	}

	if resp.UsageMetadata != nil {
		responseTurn.TokenUsage.InputTokens = int64(resp.UsageMetadata.PromptTokenCount)
		responseTurn.TokenUsage.OutputTokens = int64(resp.UsageMetadata.CandidatesTokenCount)
		responseTurn.TokenUsage.TotalTokens = int64(resp.UsageMetadata.TotalTokenCount)
		c.logger.Debugf("Token usage from GenAI (AskWithTools): Input=%d, Output=%d, Total=%d (Model: %s)",
			responseTurn.TokenUsage.InputTokens, responseTurn.TokenUsage.OutputTokens, responseTurn.TokenUsage.TotalTokens, c.modelID)
	} else {
		c.logger.Warnf("GenAI response (AskWithTools) missing UsageMetadata. Token counts will be zero/approximated. (Model: %s)", c.modelID)
	}

	return responseTurn, coreToolCallsOutput, nil
}

// Embed generates embeddings using the actual LLM API.
// (No changes from file version 0.1.4 for this function's core logic)
func (c *concreteLLMClient) Embed(ctx context.Context, text string) ([]float32, error) {
	embeddingModelID := "embedding-001"
	c.logger.Debug("ConcreteLLMClient Embed called", "text_length", len(text), "embedding_model_id", embeddingModelID)

	if !c.enabled || c.genaiClient == nil {
		c.logger.Warn("Embed called on disabled or uninitialized concrete LLM client.")
		return []float32{}, nil
	}

	em := c.genaiClient.EmbeddingModel(embeddingModelID)
	if em == nil {
		errMsg := fmt.Sprintf("failed to get embedding model from genAI client for model: %s", embeddingModelID)
		c.logger.Error(errMsg)
		return nil, NewRuntimeError(ErrorCodeLLMError, errMsg, ErrLLMError)
	}

	res, err := em.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		c.logger.Error("GenAI EmbedContent failed", "error", err, "embedding_model_id", embeddingModelID)
		return nil, NewRuntimeError(ErrorCodeLLMError, fmt.Sprintf("LLM embedding generation failed with model %s", embeddingModelID), err)
	}

	if res == nil || res.Embedding == nil || len(res.Embedding.Values) == 0 {
		c.logger.Warn("GenAI EmbedContent returned nil or empty embedding.", "embedding_model_id", embeddingModelID)
		return nil, NewRuntimeError(ErrorCodeLLMError, "LLM returned empty embedding", ErrLLMError)
	}

	return res.Embedding.Values, nil
}

// Client returns the underlying *genai.Client.
func (c *concreteLLMClient) Client() *genai.Client {
	return c.genaiClient
}

// --- Internal No-Op LLM Client Implementation (to avoid import cycle with adapters) ---

type coreInternalNoOpLLMClient struct {
	logger logging.Logger // Use the logging.Logger interface
}

// Ensure coreInternalNoOpLLMClient implements the LLMClient interface.
var _ LLMClient = (*coreInternalNoOpLLMClient)(nil)

// newCoreInternalNoOpLLMClient is a private constructor for the internal no-op client.
func newCoreInternalNoOpLLMClient(logger logging.Logger) LLMClient {
	// Logger is already guaranteed to be non-nil by NewLLMClient.
	logger.Debug("Creating internal coreInternalNoOpLLMClient.") // Adjusted log message for clarity, Changed from Info
	return &coreInternalNoOpLLMClient{logger: logger}
}

func (c *coreInternalNoOpLLMClient) Ask(ctx context.Context, turns []*ConversationTurn) (*ConversationTurn, error) {
	c.logger.Debug("coreInternalNoOpLLMClient Ask called")
	return &ConversationTurn{
		Role:       RoleAssistant,
		Content:    "No-op response from internal client.",
		TokenUsage: TokenUsageMetrics{},
	}, nil
}

func (c *coreInternalNoOpLLMClient) AskWithTools(ctx context.Context, turns []*ConversationTurn, tools []ToolDefinition) (*ConversationTurn, []*ToolCall, error) {
	c.logger.Debug("coreInternalNoOpLLMClient AskWithTools called")
	return &ConversationTurn{
		Role:       RoleAssistant,
		Content:    "No-op response from internal client (tools).",
		TokenUsage: TokenUsageMetrics{},
	}, nil, nil
}

func (c *coreInternalNoOpLLMClient) Embed(ctx context.Context, text string) ([]float32, error) {
	c.logger.Debug("coreInternalNoOpLLMClient Embed called")
	return []float32{}, nil
}

func (c *coreInternalNoOpLLMClient) Client() *genai.Client {
	return nil
}
