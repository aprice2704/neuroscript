// filename: pkg/core/llm_tools.go
package core

import (
	"context" // Import errors
	"fmt"

	"github.com/google/generative-ai-go/genai" // Keep for genai.Part potentially
)

// callLLM is a helper function used by tools to interact with the LLM.
func callLLM(ctx context.Context, llmClient LLMClient, prompt string) (string, error) {
	if llmClient == nil {
		return "", ErrLLMNotConfigured // Return specific error
	}
	turns := []*ConversationTurn{{Role: RoleUser, Content: prompt}}
	responseTurn, err := llmClient.Ask(ctx, turns)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrLLMError, err)
	}
	if responseTurn == nil {
		return "", fmt.Errorf("%w: LLM returned nil response without error", ErrLLMError)
	}
	return responseTurn.Content, nil
}

// callLLMWithParts sends a multimodal request. Needs redesign or type assertion.
func callLLMWithParts(ctx context.Context, llmClient LLMClient, parts []genai.Part) (string, error) {
	if llmClient == nil {
		return "", ErrLLMNotConfigured // Return specific error
	}
	// --- PROBLEM: Standard LLMClient interface doesn't support []genai.Part directly. ---
	// --- Simulating encoding into a single turn (likely insufficient) ---
	textContent := ""
	for _, p := range parts {
		if str, ok := p.(genai.Text); ok {
			textContent += string(str) + "\n"
		} else {
			textContent += fmt.Sprintf("[Non-text part: %T]\n", p)
		}
	}
	turns := []*ConversationTurn{{Role: RoleUser, Content: textContent}}
	responseTurn, err := llmClient.Ask(ctx, turns)
	if err != nil {
		return "", fmt.Errorf("%w: LLM Ask failed simulating parts: %w", ErrLLMError, err)
	}
	if responseTurn == nil {
		return "", fmt.Errorf("%w: LLM nil response simulating parts", ErrLLMError)
	}
	return responseTurn.Content, nil
}

// --- Tool Implementations ---

// TOOL.LLM.Ask
func toolLLMAsk(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Assumes validation layer ensures args[0] exists and is a string.
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: expected 1 argument (prompt), got %d", ErrArgumentMismatch, len(args))
	}
	prompt, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: argument 'prompt' must be a string, got %T", ErrInvalidArgument, args[0])
	}

	if interpreter.llmClient == nil {
		return nil, ErrLLMNotConfigured
	}

	response, err := callLLM(context.Background(), interpreter.llmClient, prompt)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// TOOL.LLM.AskWithParts
func toolLLMAskWithParts(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Assumes validation layer ensures args[0] exists and is a slice/list.
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: expected 1 argument (parts), got %d", ErrArgumentMismatch, len(args))
	}
	partsArg := args[0] // Should be []interface{} after validation/conversion

	var parts []genai.Part
	if partsSlice, ok := partsArg.([]interface{}); ok {
		parts = make([]genai.Part, 0, len(partsSlice)) // Initialize slice
		for idx, p := range partsSlice {
			if text, ok := p.(string); ok {
				parts = append(parts, genai.Text(text))
			} else {
				return nil, fmt.Errorf("%w: cannot convert 'parts' element at index %d (type %T) to genai.Part; complex parts conversion not implemented", ErrInvalidArgument, idx, p)
			}
		}
	} else {
		return nil, fmt.Errorf("%w: invalid argument type for 'parts': expected a list, got %T", ErrInvalidArgument, partsArg)
	}

	if interpreter.llmClient == nil {
		return nil, ErrLLMNotConfigured
	}
	response, err := callLLMWithParts(context.Background(), interpreter.llmClient, parts)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// RegisterLLMTools registers the LLM interaction tools.
func RegisterLLMTools(registry *ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("cannot register LLM tools: ToolRegistry is nil")
	}
	var err error // Declare error variable

	// Define input schema for LLM.Ask using map literal for helper conversion
	llmAskInputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"prompt": map[string]interface{}{"type": "string", "description": "The text prompt to send to the LLM."},
		},
		"required": []string{"prompt"},
	}
	// Convert schema to ArgSpec slice using the helper (now in ast_builder_helpers.go)
	llmAskArgs, argsErr := ConvertInputSchemaToArgSpec(llmAskInputSchema) // <<< USES HELPER
	if argsErr != nil {
		return fmt.Errorf("failed to convert args for LLM.Ask: %w", argsErr)
	}

	// Tool: LLM.Ask
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "LLM.Ask",
			Description: "Sends a text prompt to the configured LLM and returns the text response.",
			Args:        llmAskArgs,    // Use converted ArgSpec slice
			ReturnType:  ArgTypeString, // Specify return type
		},
		Func: toolLLMAsk,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool LLM.Ask: %w", err)
	}

	// Define input schema for LLM.AskWithParts
	llmAskPartsInputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"parts": map[string]interface{}{
				"type":        "array",
				"description": "A list of prompt parts (e.g., text strings). Complex parts may need specific encoding.",
				"items":       map[string]interface{}{"type": "string"}, // Simplified schema: assumes list of strings for now
			},
		},
		"required": []string{"parts"},
	}
	// Convert schema to ArgSpec slice using the helper
	llmAskPartsArgs, argsErr := ConvertInputSchemaToArgSpec(llmAskPartsInputSchema) // <<< USES HELPER
	if argsErr != nil {
		return fmt.Errorf("failed to convert args for LLM.AskWithParts: %w", argsErr)
	}

	// Tool: LLM.AskWithParts
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "LLM.AskWithParts",
			Description: "Sends a multimodal prompt (e.g., list of text strings) to the LLM. NOTE: Requires specific LLM client support and careful 'parts' argument construction.",
			Args:        llmAskPartsArgs, // Use converted ArgSpec slice
			ReturnType:  ArgTypeString,   // Specify return type
		},
		Func: toolLLMAskWithParts,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool LLM.AskWithParts: %w", err)
	}

	return nil // Indicate success
}

// --- REMOVED Helper Function for Schema Conversion ---
// func ConvertInputSchemaToArgSpec(schema map[string]interface{}) ([]ArgSpec, error) { ... }
