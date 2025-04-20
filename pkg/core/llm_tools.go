// filename: pkg/core/llm_tools.go
package core

import (
	"context"
	"errors"
	"fmt"

	// "log" // Logger comes from interpreter
	// No longer need NewLLMClient from core.llm

	"github.com/google/generative-ai-go/genai"
)

// --- Existing toolAskLLM ---
func toolAskLLM(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.AskLLM requires exactly one argument (prompt string), got %d", len(args))
	}
	prompt, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.AskLLM argument must be a string, got %T", args[0])
	}
	if prompt == "" {
		return nil, errors.New("TOOL.AskLLM prompt cannot be empty")
	}

	// --- MODIFIED: Use Interpreter's LLMClient ---
	llmClient := interpreter.llmClient                 // Get client from interpreter
	if llmClient == nil || llmClient.Client() == nil { // Check underlying client too
		interpreter.Logger().Println("[ERROR TOOL.AskLLM] LLM client not available via interpreter.")
		return nil, errors.New("TOOL.AskLLM: LLM client not available or not initialized")
	}
	// --- END MODIFIED ---

	ctx := context.Background()
	// Call method on the existing client instance
	response, err := llmClient.CallLLM(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("TOOL.AskLLM failed: %w", err)
	}
	return response, nil
}

// --- NEW Tool: AskLLMWithFiles ---
func toolAskLLMWithFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.AskLLMWithFiles requires exactly two arguments (prompt_text string, file_uris list), got %d", len(args))
	}

	promptText, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.AskLLMWithFiles: first argument (prompt_text) must be a string, got %T", args[0])
	}

	fileURIsArg, ok := args[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("TOOL.AskLLMWithFiles: second argument (file_uris) must be a list, got %T", args[1])
	}

	fileURIs := []string{}
	for i, item := range fileURIsArg {
		uri, ok := item.(string)
		if !ok || uri == "" {
			interpreter.Logger().Printf("[WARN TOOL.AskLLMWithFiles] Skipping invalid/empty URI at index %d in file_uris list.", i)
			continue
		}
		fileURIs = append(fileURIs, uri)
	}

	if len(fileURIs) == 0 {
		interpreter.Logger().Println("[WARN TOOL.AskLLMWithFiles] file_uris list contained no valid URIs.")
		return nil, errors.New("TOOL.AskLLMWithFiles: requires at least one valid file URI in the list")
	}

	parts := []genai.Part{}
	interpreter.Logger().Printf("[TOOL.AskLLMWithFiles] Preparing parts. Files: %d, Prompt: %q", len(fileURIs), promptText)
	for _, uri := range fileURIs {
		parts = append(parts, genai.FileData{URI: uri})
		interpreter.Logger().Printf("[TOOL.AskLLMWithFiles] Added FileData: %s", uri)
	}
	if promptText != "" {
		parts = append(parts, genai.Text(promptText))
		interpreter.Logger().Printf("[TOOL.AskLLMWithFiles] Added Text part.")
	} else {
		interpreter.Logger().Printf("[TOOL.AskLLMWithFiles] No text prompt provided, sending files only.")
	}

	// --- MODIFIED: Use Interpreter's LLMClient ---
	llmClient := interpreter.llmClient                 // Get client from interpreter
	if llmClient == nil || llmClient.Client() == nil { // Check underlying client too
		interpreter.Logger().Println("[ERROR TOOL.AskLLMWithFiles] LLM client not available via interpreter.")
		return nil, errors.New("TOOL.AskLLMWithFiles: LLM client not available or not initialized")
	}
	// --- END MODIFIED ---

	ctx := context.Background()
	// Call method on the existing client instance
	resp, err := llmClient.CallLLMWithParts(ctx, parts, nil)

	if err != nil {
		return nil, fmt.Errorf("TOOL.AskLLMWithFiles LLM call failed: %w", err)
	}

	// Existing response processing logic
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		part := resp.Candidates[0].Content.Parts[0]
		if text, ok := part.(genai.Text); ok {
			interpreter.Logger().Printf("[TOOL.AskLLMWithFiles] Received text response.")
			return string(text), nil
		}
	}
	interpreter.Logger().Printf("[WARN TOOL.AskLLMWithFiles] Received non-text or empty response.")
	return "", errors.New("TOOL.AskLLMWithFiles received non-text or empty response")
}

// --- Registration Function ---
// (Registration logic remains unchanged)
func registerLLMTools(registry *ToolRegistry) error {
	var err error
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "AskLLM",
			Description: "Sends a single text prompt to the LLM and returns the text response. This call is stateless.",
			Args: []ArgSpec{
				{Name: "prompt", Type: ArgTypeString, Required: true, Description: "The text prompt to send to the LLM."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolAskLLM,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool AskLLM: %w", err)
	}

	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "AskLLMWithFiles",
			Description: "Sends a request to the LLM including both text prompt and references to uploaded files (via their API URIs). Returns the text response.",
			Args: []ArgSpec{
				{Name: "prompt_text", Type: ArgTypeString, Required: true, Description: "The text prompt to accompany the files."},
				{Name: "file_uris", Type: ArgTypeList, Required: true, Description: "A list of strings, where each string is a File API URI (e.g., 'files/...') for an uploaded file."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolAskLLMWithFiles,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool AskLLMWithFiles: %w", err)
	}

	return nil
}
