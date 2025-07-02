// NeuroScript Version: 0.3.1
// File version: 0.0.1 // Correct ToolRegistry interface usage in RegisterLLMTools.
// nlines: 121
// risk_rating: MEDIUM
// filename: pkg/llm/llm_tools.go
package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/google/generative-ai-go/genai"	// Keep for genai.Part potentially
)

// callLLM is a helper function used by tools to interact with the LLM.
func callLLM(ctx context.Context, llmClient interfaces.LLMClient, prompt string) (string, error) {
	if llmClient == nil {
		return "", lang.ErrLLMNotConfigured	// Return specific error
	}
	turns := []*interfaces.ConversationTurn{{Role: interfaces.RoleUser, Content: prompt}}
	responseTurn, err := llmClient.Ask(ctx, turns)
	if err != nil {
		return "", fmt.Errorf("%w: %w", lang.ErrLLMError, err)
	}
	if responseTurn == nil {
		return "", fmt.Errorf("%w: LLM returned nil response without error", lang.ErrLLMError)
	}
	return responseTurn.Content, nil
}

// callLLMWithParts sends a multimodal request.
// Note: This function's current implementation converts all parts to text,
// which might not be suitable for true multimodal inputs.
// The standard LLMClient interface might need extension for direct multimodal support.
func callLLMWithParts(ctx context.Context, llmClient interfaces.LLMClient, parts []genai.Part) (string, error) {
	if llmClient == nil {
		return "", lang.ErrLLMNotConfigured	// Return specific error
	}

	// Convert genai.Part to a text-based representation for the current LLMClient.Ask interface
	var contentBuilder strings.Builder
	for _, p := range parts {
		switch v := p.(type) {
		case genai.Text:
			contentBuilder.WriteString(string(v))
		case genai.Blob:
			// For now, just indicate a blob was present. A real implementation might
			// use a specific encoding or store it and pass a reference if the LLM supports it.
			contentBuilder.WriteString(fmt.Sprintf("[Blob Content (MIME: %s)]", v.MIMEType))
		default:
			contentBuilder.WriteString(fmt.Sprintf("[Unsupported genai.Part type: %T]", p))
		}
		contentBuilder.WriteString("\n")	// Add a newline between parts for readability
	}

	turns := []*interfaces.ConversationTurn{{Role: interfaces.RoleUser, Content: strings.TrimSpace(contentBuilder.String())}}
	responseTurn, err := llmClient.Ask(ctx, turns)
	if err != nil {
		return "", fmt.Errorf("%w: LLM Ask failed while simulating parts: %w", lang.ErrLLMError, err)
	}
	if responseTurn == nil {
		return "", fmt.Errorf("%w: LLM returned nil response while simulating parts", lang.ErrLLMError)
	}
	return responseTurn.Content, nil
}

// --- Tool Implementations ---

// TOOL.LLM.Ask
func toolLLMAsk(interpreter *neurogo.Interpreter, args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "LLM.Ask: expected 1 argument (prompt)", lang.ErrArgumentMismatch)
	}
	prompt, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("LLM.Ask: argument 'prompt' must be a string, got %T", args[0]), lang.ErrInvalidArgument)
	}

	if interpreter.llmClient == nil {
		return nil, lang.ErrLLMNotConfigured
	}

	response, err := callLLM(context.Background(), interpreter.llmClient, prompt)
	if err != nil {
		// callLLM already wraps ErrLLMError, so just return it.
		return nil, err
	}
	return response, nil
}

// TOOL.LLM.AskWithParts
func toolLLMAskWithParts(interpreter *neurogo.Interpreter, args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "LLM.AskWithParts: expected 1 argument (parts)", lang.ErrArgumentMismatch)
	}
	partsArg := args[0]

	var genaiParts []genai.Part
	if partsSlice, ok := partsArg.([]interface{}); ok {
		genaiParts = make([]genai.Part, 0, len(partsSlice))
		for idx, p := range partsSlice {
			// For simplicity, assuming parts are strings for now.
			// A more robust implementation would handle various genai.Part types (Text, Blob, etc.)
			// and might require specific input structures or type assertions.
			if text, ok := p.(string); ok {
				genaiParts = append(genaiParts, genai.Text(text))
			} else {
				// If it's already a genai.Part, use it directly. This is unlikely with current NeuroScript arg passing.
				// else if part, ok := p.(genai.Part); ok {
				// genaiParts = append(genaiParts, part)
				// }
				return nil, lang.NewRuntimeError(lang.ErrorCodeType,
					fmt.Sprintf("LLM.AskWithParts: 'parts' element at index %d (type %T) must be a string for current implementation", idx, p),
					lang.ErrInvalidArgument)
			}
		}
	} else {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("LLM.AskWithParts: invalid argument type for 'parts': expected a list, got %T", partsArg), lang.ErrInvalidArgument)
	}

	if interpreter.llmClient == nil {
		return nil, lang.ErrLLMNotConfigured
	}
	response, err := callLLMWithParts(context.Background(), interpreter.llmClient, genaiParts)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// RegisterLLMTools registers the LLM interaction tools.
// <<< CHANGED: registry parameter is now ToolRegistry (interface type)
func RegisterLLMTools(registry ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("cannot register LLM tools: provided ToolRegistry is nil")
	}
	var err error

	llmAskInputSchema := map[string]interface{}{
		"type":	"object",
		"properties": map[string]interface{}{
			"prompt": map[string]interface{}{"type": "string", "description": "The text prompt to send to the LLM."},
		},
		"required":	[]string{"prompt"},
	}
	llmAskArgs, argsErr := parser.ConvertInputSchemaToArgSpec(llmAskInputSchema)
	if argsErr != nil {
		return fmt.Errorf("failed to convert args for LLM.Ask: %w", argsErr)
	}

	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:		"LLM.Ask",
			Description:	"Sends a text prompt to the configured LLM and returns the text response.",
			Args:		llmAskArgs,
			ReturnType:	parser.ArgTypeString,
		},
		Func:	toolLLMAsk,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool LLM.Ask: %w", err)
	}

	llmAskPartsInputSchema := map[string]interface{}{
		"type":	"object",
		"properties": map[string]interface{}{
			"parts": map[string]interface{}{
				"type":		"array",
				"description":	"A list of prompt parts (e.g., text strings). Complex parts may need specific encoding or LLM client support.",
				"items":	map[string]interface{}{"type": "string"},	// Simplified: assumes list of strings.
			},
		},
		"required":	[]string{"parts"},
	}
	llmAskPartsArgs, argsErr := parser.ConvertInputSchemaToArgSpec(llmAskPartsInputSchema)
	if argsErr != nil {
		return fmt.Errorf("failed to convert args for LLM.AskWithParts: %w", argsErr)
	}

	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:		"LLM.AskWithParts",
			Description:	"Sends a list of parts (currently treated as text strings) as a prompt to the LLM.",
			Args:		llmAskPartsArgs,
			ReturnType:	parser.ArgTypeString,
		},
		Func:	toolLLMAskWithParts,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool LLM.AskWithParts: %w", err)
	}

	return nil
}