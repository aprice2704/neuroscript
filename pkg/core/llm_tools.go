// filename: pkg/core/llm_tools.go
package core

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai" // Keep for genai.Part potentially
)

// callLLM is a helper function used by tools to interact with the LLM.
func callLLM(ctx context.Context, llmClient LLMClient, prompt string) (string, error) {
	if llmClient == nil {
		return "", fmt.Errorf("LLM client is nil")
	}
	turns := []*ConversationTurn{{Role: RoleUser, Content: prompt}}
	responseTurn, err := llmClient.Ask(ctx, turns)
	if err != nil {
		return "", fmt.Errorf("LLM Ask failed: %w", err)
	}
	if responseTurn == nil {
		return "", fmt.Errorf("LLM returned nil response without error")
	}
	return responseTurn.Content, nil
}

// callLLMWithParts sends a multimodal request. Needs redesign or type assertion.
func callLLMWithParts(ctx context.Context, llmClient LLMClient, parts []genai.Part) (string, error) {
	if llmClient == nil {
		return "", fmt.Errorf("LLM client is nil")
	}
	// --- PROBLEM: Standard LLMClient interface doesn't support []genai.Part directly. ---
	// --- See previous responses for potential solutions (type assertion, encoding) ---

	// --- Simulating encoding into a single turn (likely insufficient) ---
	textContent := ""
	for _, p := range parts {
		if str, ok := p.(genai.Text); ok {
			textContent += string(str) + "\n"
		} else {
			// Placeholder for non-text parts
			textContent += fmt.Sprintf("[Non-text part: %T]\n", p)
		}
	}
	turns := []*ConversationTurn{{Role: RoleUser, Content: textContent}}
	responseTurn, err := llmClient.Ask(ctx, turns)
	if err != nil {
		return "", fmt.Errorf("LLM Ask failed simulating parts: %w", err)
	}
	if responseTurn == nil {
		return "", fmt.Errorf("LLM nil response simulating parts")
	}
	return responseTurn.Content, nil

	// --- Fallback Error (if no solution implemented) ---
	// return "", fmt.Errorf("callLLMWithParts requires specific LLMClient support")
}

// --- Tool Implementations ---

// TOOL.LLM.Ask
func toolLLMAsk(ctx context.Context, interp *Interpreter, args map[string]interface{}) (interface{}, error) {
	// Use the helper function defined in tools_helpers.go
	prompt, err := getStringArg(args, "prompt")
	if err != nil {
		return nil, err
	}
	if interp.llmClient == nil {
		return nil, fmt.Errorf("LLM client not configured in interpreter")
	}
	response, err := callLLM(ctx, interp.llmClient, prompt)
	if err != nil {
		return nil, err // Error already wrapped by callLLM
	}
	return response, nil
}

// TOOL.LLM.AskWithParts
func toolLLMAskWithParts(ctx context.Context, interp *Interpreter, args map[string]interface{}) (interface{}, error) {
	partsArg, ok := args["parts"]
	if !ok {
		return nil, fmt.Errorf("missing required argument 'parts'")
	}

	var parts []genai.Part
	if partsSlice, ok := partsArg.([]interface{}); ok {
		for idx, p := range partsSlice {
			// Attempt to convert interface{} back to genai.Part
			// This is simplified and likely needs more robust handling
			if text, ok := p.(string); ok {
				parts = append(parts, genai.Text(text))
			} else {
				// Handle other potential part types (e.g., maps representing blobs) here
				return nil, fmt.Errorf("cannot convert 'parts' element at index %d (type %T) to genai.Part; complex parts conversion not implemented", idx, p)
			}
		}
	} else {
		return nil, fmt.Errorf("invalid argument type for 'parts': expected a list, got %T", partsArg)
	}

	if interp.llmClient == nil {
		return nil, fmt.Errorf("LLM client not configured in interpreter")
	}
	// Use the helper, acknowledging its limitations
	response, err := callLLMWithParts(ctx, interp.llmClient, parts)
	if err != nil {
		return nil, err // Error already wrapped by callLLMWithParts
	}
	return response, nil
}

// RegisterLLMTools registers the LLM interaction tools.
// Assumes ToolRegistry and ToolImplementation types are defined correctly.
func RegisterLLMTools(registry *ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("cannot register tools: ToolRegistry is nil")
	}

	// Tool: LLM.Ask
	err := registry.RegisterTool(ToolImplementation{
		Definition: ToolDefinition{
			Name:        "LLM.Ask",
			Description: "Sends a text prompt to the configured LLM and returns the text response.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prompt": map[string]interface{}{"type": "string", "description": "The text prompt to send to the LLM."},
				},
				"required": []string{"prompt"},
			},
			// OutputSchema removed
		},
		Execute: toolLLMAsk,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool LLM.Ask: %w", err)
	}

	// Tool: LLM.AskWithParts
	err = registry.RegisterTool(ToolImplementation{
		Definition: ToolDefinition{
			Name:        "LLM.AskWithParts",
			Description: "Sends a multimodal prompt (text and other data parts) to the LLM. NOTE: Requires specific LLM client support and careful 'parts' argument construction.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"parts": map[string]interface{}{
						"type":        "array",
						"description": "A list of prompt parts (e.g., text strings).",
						"items":       map[string]interface{}{"type": "string"}, // Simplified schema
					},
				},
				"required": []string{"parts"},
			},
			// OutputSchema removed
		},
		Execute: toolLLMAskWithParts,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool LLM.AskWithParts: %w", err)
	}

	return nil // Indicate success
}

// --- Assumed Type Definitions (ensure they exist) ---
// type ToolRegistry struct { ... }
// func (tr *ToolRegistry) RegisterTool(impl ToolImplementation) error
// type ToolImplementation struct { Definition ToolDefinition; Execute ToolExecutorFunc }
// type ToolDefinition struct { Name string; Description string; InputSchema any }
// type ToolExecutorFunc func(ctx context.Context, interp *Interpreter, args map[string]interface{}) (interface{}, error)
