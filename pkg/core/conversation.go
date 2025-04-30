// filename: pkg/core/conversation.go
package core

import (
	"fmt"
	"log"
	"strings" // Added for text formatting

	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/generative-ai-go/genai"
)

// ConversationManager holds the state of an interaction with the LLM agent.
type ConversationManager struct {
	History []*genai.Content // Stores the conversation turns
	logger  logging.Logger
}

// NewConversationManager creates a new manager.
func NewConversationManager(logger logging.Logger) *ConversationManager {
	// Ensure logger is not nil
	effectiveLogger := logger
	if effectiveLogger == nil {
		panic("ConversationManager requires a valid logger")
	}
	return &ConversationManager{
		History: make([]*genai.Content, 0),
		logger:  effectiveLogger,
	}
}

// AddUserMessage appends a user message to the history.
func (cm *ConversationManager) AddUserMessage(text string) {
	cm.History = append(cm.History, &genai.Content{
		Role:  "user",
		Parts: []genai.Part{genai.Text(text)},
	})
	cm.logger.Info("[CONVO] Added User: %q", text)
}

// +++ ADDED: AddModelMessage +++
// AddModelMessage adds a pure text response from the model to the history.
func (cm *ConversationManager) AddModelMessage(text string) {
	cm.History = append(cm.History, &genai.Content{
		Role:  "model",
		Parts: []genai.Part{genai.Text(text)},
	})
	// Provide a snippet for logging, handle potential empty strings
	logSnippet := text
	maxLogLen := 80
	if len(logSnippet) > maxLogLen {
		logSnippet = logSnippet[:maxLogLen] + "..."
	}
	cm.logger.Info("[CONVO] Added Model Text: %q", logSnippet)
}

// AddModelResponse adds a model's response (potentially including function calls) to the history.
// NOTE: This might be replaced by more granular additions below if preferred.
func (cm *ConversationManager) AddModelResponse(candidate *genai.Candidate) error {
	if candidate == nil || candidate.Content == nil {
		cm.logger.Warn("CONVO] Attempted to add nil candidate or candidate content.")
		return nil // Return nil, as it's not a critical error for the manager itself
	}

	// Ensure the role is 'model' if unset, reject if it's something else
	if candidate.Content.Role == "" {
		candidate.Content.Role = "model"
	}
	if candidate.Content.Role != "model" {
		err := fmt.Errorf("attempted to add non-model content (Role: %s) as model response", candidate.Content.Role)
		cm.logger.Error("CONVO] %v", err)
		return err // Return error as this indicates misuse
	}

	cm.History = append(cm.History, candidate.Content)

	// Enhanced logging based on parts
	if len(candidate.Content.Parts) == 0 {
		cm.logger.Info("[CONVO] Added Model response with no parts.")
	} else {
		logMsgs := []string{}
		for _, part := range candidate.Content.Parts {
			switch v := part.(type) {
			case genai.Text:
				logSnippet := string(v)
				maxLogLen := 40
				if len(logSnippet) > maxLogLen {
					logSnippet = logSnippet[:maxLogLen] + "..."
				}
				logMsgs = append(logMsgs, fmt.Sprintf("Text(%q)", logSnippet))
			case genai.FunctionCall:
				logMsgs = append(logMsgs, fmt.Sprintf("FunctionCall(%s)", v.Name))
			default:
				logMsgs = append(logMsgs, fmt.Sprintf("UnknownPart(%T)", v))
			}
		}
		cm.logger.Info("[CONVO] Added Model response with parts: [%s]", strings.Join(logMsgs, ", "))
	}
	return nil
}

// +++ ADDED: AddFunctionCallMessage +++
// AddFunctionCallMessage adds a function call request from the model to the history.
// This should be called *before* the function is executed.
func (cm *ConversationManager) AddFunctionCallMessage(fc genai.FunctionCall) {
	cm.History = append(cm.History, &genai.Content{
		Role:  "model", // The FunctionCall is part of the *model's* turn
		Parts: []genai.Part{fc},
	})
	cm.logger.Info("[CONVO] Added Model FunctionCall request: %s", fc.Name)
}

// AddFunctionResponse adds the result of a function execution back into the history.
// Deprecated: Use AddFunctionResultMessage instead for clarity.
func (cm *ConversationManager) AddFunctionResponse(toolName string, responseData map[string]interface{}) error {
	cm.logger.Warn("CONVO] AddFunctionResponse is deprecated, use AddFunctionResultMessage.")
	// Create the part and call the new method
	part := genai.FunctionResponse{
		Name:     toolName,
		Response: responseData,
	}
	return cm.AddFunctionResultMessage(part)
}

// +++ ADDED: AddFunctionResultMessage +++
// AddFunctionResultMessage adds a function response part to the history.
// This should be called *after* the function is executed.
func (cm *ConversationManager) AddFunctionResultMessage(part genai.Part) error {
	// Basic validation: Ensure the part is actually a FunctionResponse
	fr, ok := part.(genai.FunctionResponse)
	if !ok {
		err := fmt.Errorf("attempted to add part of type %T as function result, expected genai.FunctionResponse", part)
		cm.logger.Error("CONVO] %v", err)
		return err
	}

	cm.History = append(cm.History, &genai.Content{
		Role:  "function", // Role is 'function' for results
		Parts: []genai.Part{fr},
	})
	cm.logger.Info("[CONVO] Added FunctionResponse result for: %s", fr.Name)
	return nil
}

// GetHistory returns the current conversation history.
func (cm *ConversationManager) GetHistory() []*genai.Content {
	// Return a copy to prevent external modification? For now, return original.
	return cm.History
}

// ClearHistory resets the conversation.
func (cm *ConversationManager) ClearHistory() {
	cm.History = make([]*genai.Content, 0)
	cm.logger.Info("[CONVO] History cleared.")
}

// --- Helper for creating error responses ---

// +++ ADDED: CreateErrorFunctionResultPart +++
// CreateErrorFunctionResultPart formats an error into a FunctionResponse part.
func CreateErrorFunctionResultPart(toolName string, execErr error) genai.Part {
	errStr := "Unknown execution error"
	if execErr != nil {
		errStr = execErr.Error()
	}
	return genai.FunctionResponse{
		Name: toolName,
		Response: map[string]interface{}{
			// Standardize error response structure
			"error": errStr,
		},
	}
}

// --- Helper for parsing safety settings (Example, unchanged) ---
func parseSafetySettings(settings []string) ([]*genai.SafetySetting, error) {
	parsed := make([]*genai.SafetySetting, 0, len(settings))
	// --- Implementation commented out ---
	if len(settings) > 0 {
		log.Println("[WARN CONVO] parseSafetySettings function needs implementation.")
	}
	return parsed, nil
}
