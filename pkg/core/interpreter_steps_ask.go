// filename: pkg/core/interpreter_steps_ask.go
package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	// Assuming other necessary imports like interfaces Logger etc.
)

// executeAskAI handles the "call askAI ..." step.
func (i *Interpreter) executeAskAI(step Step, stepNum int, evaluatedArgs []interface{}) (interface{}, error) {
	i.Logger().Info("[DEBUG-INTERP]   Executing AskAI (Step %d)", stepNum+1)

	// --- Argument Validation ---
	if len(evaluatedArgs) == 0 || len(evaluatedArgs) > 2 {
		// Basic check: Allow prompt (string) and optional context (map or string?)
		// Refine this based on more detailed spec if available.
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("askAI expects 1 or 2 arguments (prompt string, [optional context]), got %d", len(evaluatedArgs)),
			ErrArgumentMismatch)
	}

	prompt, promptOk := evaluatedArgs[0].(string)
	if !promptOk {
		return nil, NewRuntimeError(ErrorCodeType,
			fmt.Sprintf("askAI first argument (prompt) must be a string, got %T", evaluatedArgs[0]),
			ErrInvalidFunctionArgument)
	}

	// Optional context argument handling (placeholder)
	// var requestContext interface{} = nil
	// if len(evaluatedArgs) == 2 {
	//  requestContext = evaluatedArgs[1]
	//  // TODO: Process context if needed (e.g., format into prompt preamble)
	// }
	// --- End Argument Validation ---

	if i.llmClient == nil {
		i.Logger().Error("[ERROR INTERP] askAI called but LLMClient is not configured.")
		return nil, NewRuntimeError(ErrorCodeLLMError, "LLM client not configured in interpreter", ErrInternal) // Or a more specific sentinel?
	}

	// --- LLM Call ---
	i.Logger().Info("[INFO INTERP] Calling LLM (Model: %s) with prompt: %q", i.modelName, prompt) // Log prompt potentially truncated
	// Simple text generation for now. Adapt if history/context needs more complex handling.
	// Use a background context for now. Consider making it configurable/passed down.
	ctx := context.Background()
	model := i.llmClient.Client().GenerativeModel(i.modelName)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))

	if err != nil {
		i.Logger().Error("[ERROR INTERP] LLM content generation failed: %v", err)
		// Wrap the LLM error
		return nil, NewRuntimeError(ErrorCodeLLMError, "LLM API call failed", fmt.Errorf("generating content: %w", err))
	}

	// --- Process Response ---
	// Assuming a simple text response is expected. Need to handle potential errors/empty responses.
	// This part needs refinement based on how genai library structures responses and potential errors within responses.
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		i.Logger().Warn("[WARN INTERP] LLM response was empty or malformed.")
		// Return empty string or a specific error? Return empty string for now.
		return "", nil // Considered successful call, but empty result.
	}

	// Extract text from the first candidate's first part. Adapt if multiple parts/candidates are expected.
	// Ensure the part is actually text.
	textContent := ""
	if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		textContent = string(textPart)
	} else {
		i.Logger().Warn("[WARN INTERP] LLM response part was not text: %T", resp.Candidates[0].Content.Parts[0])
		// Return empty string or error? Let's return empty string.
		return "", nil
	}

	i.Logger().Info("[INFO INTERP] LLM call successful. Response length: %d", len(textContent))
	// Return the text content
	return textContent, nil
}

// executeAskHuman handles the "call askHuman ..." step.
func (i *Interpreter) executeAskHuman(step Step, stepNum int, evaluatedArgs []interface{}) (interface{}, error) {
	i.Logger().Info("[DEBUG-INTERP]   Executing AskHuman (Step %d)", stepNum+1)

	// --- Argument Validation ---
	if len(evaluatedArgs) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("askHuman expects 1 argument (prompt string), got %d", len(evaluatedArgs)),
			ErrArgumentMismatch)
	}
	prompt, ok := evaluatedArgs[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType,
			fmt.Sprintf("askHuman first argument (prompt) must be a string, got %T", evaluatedArgs[0]),
			ErrInvalidFunctionArgument)
	}
	// --- End Argument Validation ---

	// --- Placeholder Implementation ---
	// Log the prompt and return a placeholder response.
	// In a real implementation, this would involve interacting with a UI or external system.
	i.Logger().Info("[INFO INTERP] AskHuman: Prompt for user: %q", prompt)
	placeholderResponse := fmt.Sprintf("[Placeholder response for: %s]", prompt)

	// Return placeholder response
	return placeholderResponse, nil
}

// executeAskComputer handles the "call askComputer ..." step.
// Delegates to tool execution logic.
func (i *Interpreter) executeAskComputer(step Step, stepNum int, evaluatedArgs []interface{}) (interface{}, error) {
	i.Logger().Info("[DEBUG-INTERP]   Executing AskComputer (Step %d)", stepNum+1)

	// --- Argument Validation ---
	if len(evaluatedArgs) == 0 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			"askComputer requires at least one argument (tool name string)",
			ErrArgumentMismatch)
	}
	toolName, ok := evaluatedArgs[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType,
			fmt.Sprintf("askComputer first argument (tool name) must be a string, got %T", evaluatedArgs[0]),
			ErrInvalidFunctionArgument)
	}
	// --- End Argument Validation ---

	// --- Delegate to Tool Call ---
	// Synthesize arguments for the tool (excluding the tool name itself)
	toolArgs := evaluatedArgs[1:]

	// Re-evaluate: executeCall expects *AST nodes* in step.Args, then evaluates them.
	// We already have evaluated args. We need to call the *tool execution* logic directly,
	// bypassing the argument evaluation part of executeCall.

	// Let's extract the relevant part from executeCall's tool logic:
	toolImpl, found := i.ToolRegistry().GetTool(toolName)
	if !found {
		errMsg := fmt.Sprintf("tool '%s' (called via askComputer) not found in registry", toolName)
		return nil, NewRuntimeError(ErrorCodeToolNotFound, errMsg, fmt.Errorf("%s: %w", errMsg, ErrToolNotFound))
	}

	// Validate and Convert the ALREADY EVALUATED arguments (excluding the tool name itself)
	validatedAndConvertedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, toolArgs)
	if validationErr != nil {
		code := ErrorCodeArgMismatch
		if errors.Is(validationErr, ErrValidationTypeMismatch) {
			code = ErrorCodeType
		} else if errors.Is(validationErr, ErrValidationArgCount) {
			code = ErrorCodeArgMismatch
		}
		return nil, NewRuntimeError(code, fmt.Sprintf("argument validation failed for tool '%s' (called via askComputer)", toolName), fmt.Errorf("validating args for %s: %w", toolName, validationErr))
	}

	// Execute the tool directly
	i.Logger().Debug("[DEBUG-INTERP]     Executing Tool '%s' (via askComputer)...", toolName)
	toolResult, toolErr := toolImpl.Func(i, validatedAndConvertedArgs)

	if toolErr != nil {
		if re, ok := toolErr.(*RuntimeError); ok {
			return nil, re // Already a RuntimeError
		}
		code := ErrorCodeToolSpecific
		return nil, NewRuntimeError(code, fmt.Sprintf("tool '%s' (via askComputer) execution failed", toolName), fmt.Errorf("executing tool %s: %w", toolName, toolErr))
	}

	i.Logger().Debug("[DEBUG-INTERP]     Tool '%s' (via askComputer) execution successful.", toolName)
	return toolResult, nil
	// --- End Delegate to Tool Call ---
}
