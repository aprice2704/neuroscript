// NeuroScript Version: 0.3.0
// File version: 0.0.3
// Corrected genai.FunctionResponse.Response type handling.
// filename: pkg/core/conversation_helpers.go
// nlines: 125 // Approximate
// risk_rating: MEDIUM
package core

import (
	"fmt"
	"strings"

	// Keep for potential future use of timestamps
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid" // For placeholder ID generation
)

// ConvertGenaiContentsToConversationTurns converts a slice of *genai.Content
// (used by ConversationManager) to a slice of *ConversationTurn (used by LLMClient.Ask).
func ConvertGenaiContentsToConversationTurns(genaiHistory []*genai.Content) []*ConversationTurn {
	if genaiHistory == nil {
		return nil
	}
	coreTurns := make([]*ConversationTurn, 0, len(genaiHistory))
	for _, genaiContent := range genaiHistory {
		if genaiContent == nil {
			continue
		}

		var coreRole Role
		switch strings.ToLower(genaiContent.Role) {
		case "user":
			coreRole = RoleUser
		case "model", "assistant":
			coreRole = RoleAssistant
		case "function":
			coreRole = RoleTool // genai "function" role contains FunctionResponse parts for us
		case "system":
			coreRole = RoleSystem
		default:
			coreRole = Role(genaiContent.Role) // Fallback, might need validation if strict roles are enforced
		}

		turn := &ConversationTurn{
			Role: coreRole,
			// core.ConversationTurn (from your llm_types.go) has:
			// Content     string
			// ToolCalls   []*ToolCall
			// ToolResults []*ToolResult
			// TokenUsage  TokenUsageMetrics
			// It does not have a direct Timestamp field in the provided definition.
		}

		var textContentParts []string
		currentToolCalls := []*ToolCall{}
		currentToolResults := []*ToolResult{}

		for _, genaiPart := range genaiContent.Parts {
			switch p := genaiPart.(type) {
			case genai.Text:
				textContentParts = append(textContentParts, string(p))
			case genai.FunctionCall:
				// core.ToolCall has ID, Name, Arguments. genai.FunctionCall has Name, Args.
				// We need to generate an ID for the core.ToolCall.
				callID := uuid.NewString() // Generate a unique ID for this tool call
				currentToolCalls = append(currentToolCalls, &ToolCall{
					ID:        callID,
					Name:      p.Name,
					Arguments: p.Args,
				})
			case genai.FunctionResponse:
				// This part populates core.ToolResult.
				// The genai.FunctionResponse.Name is the function name.
				// We need to associate this with a ToolCall ID. This mapping is tricky here
				// as genai.FunctionResponse doesn't carry the original call ID.
				// For now, we'll assume the Name can be used to correlate, or that this conversion
				// happens in a context where such correlation is possible.
				// If genaiContent.Role was "function", this part is the primary data.
				currentToolResults = append(currentToolResults, &ToolResult{
					ID:     p.Name, // Placeholder: Ideally this should be the ID of the ToolCall it's responding to.
					Result: p.Response,
					// Error field is not directly in genai.FunctionResponse; errors are usually part of the Response map.
				})
			default:
				textContentParts = append(textContentParts, fmt.Sprintf("[unhandled genai.Part type: %T]", p))
			}
		}
		turn.Content = strings.Join(textContentParts, "\n")
		turn.ToolCalls = currentToolCalls
		// Only assign ToolResults if the role is specifically for tool responses.
		// genai.Content with Role "function" contains genai.FunctionResponse parts.
		if coreRole == RoleTool {
			turn.ToolResults = currentToolResults
		}

		coreTurns = append(coreTurns, turn)
	}
	return coreTurns
}

// ConvertCoreTurnsToGenaiContents converts a slice of *ConversationTurn
// back to a slice of *genai.Content.
func ConvertCoreTurnsToGenaiContents(coreTurns []*ConversationTurn) []*genai.Content {
	if coreTurns == nil {
		return nil
	}
	genaiContents := make([]*genai.Content, 0, len(coreTurns))
	for _, turn := range coreTurns {
		if turn == nil {
			continue
		}

		genaiRole := string(turn.Role)
		if turn.Role == RoleAssistant {
			genaiRole = "model" // Google's genai library uses "model" for assistant role
		} else if turn.Role == RoleTool {
			genaiRole = "function" // Google's genai library uses "function" for tool responses
		}

		content := &genai.Content{
			Role:  genaiRole,
			Parts: make([]genai.Part, 0),
		}

		if turn.Content != "" {
			content.Parts = append(content.Parts, genai.Text(turn.Content))
		}

		for _, tc := range turn.ToolCalls {
			if tc != nil {
				content.Parts = append(content.Parts, genai.FunctionCall{Name: tc.Name, Args: tc.Arguments})
			}
		}

		// This part handles core.ToolResults and converts them to genai.FunctionResponse parts.
		// This is typically when the coreTurn.Role is RoleTool.
		if turn.Role == RoleTool {
			for _, tr := range turn.ToolResults {
				if tr != nil {
					var responseMap map[string]any
					if tr.Error != "" {
						// If there's an error in ToolResult, represent it in the map
						responseMap = map[string]any{"error": tr.Error}
					} else {
						// If tr.Result is already a map[string]any, use it directly.
						if resultMap, ok := tr.Result.(map[string]any); ok {
							responseMap = resultMap
						} else {
							// Otherwise, wrap the result in a map with a default key "value".
							// This ensures the type matches genai.FunctionResponse.Response.
							responseMap = map[string]any{"value": tr.Result}
						}
					}
					// The genai.FunctionResponse.Name should be the name of the function that was called.
					// core.ToolResult.ID should correspond to the ID of the ToolCall.
					// Assuming tr.ID here refers to the function name for this conversion context.
					// This might need a lookup if tr.ID is a call ID and function name is stored elsewhere.
					responsePart := genai.FunctionResponse{
						Name:     tr.ID, // Assuming tr.ID holds the function name here. This is a common ambiguity.
						Response: responseMap,
					}
					content.Parts = append(content.Parts, responsePart)
				}
			}
		}
		genaiContents = append(genaiContents, content)
	}
	return genaiContents
}
