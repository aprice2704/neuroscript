// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Remove duplicate CreateErrorFunctionResultPart.
// filename: pkg/core/conversation.go
package core

import (
	"fmt"
	"log"
	"strings" // Added for text formatting

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/google/generative-ai-go/genai"
)

// ConversationManager holds the state of an interaction with the LLM agent.
type ConversationManager struct {
	History []*genai.Content // Stores the conversation turns
	logger  interfaces.Logger
}

// NewConversationManager creates a new manager.
func NewConversationManager(logger interfaces.Logger) *ConversationManager {
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
	cm.logger.Debug("[CONVO] Added User: %q", text)
}

// AddModelMessage adds a pure text response from the model to the history.
func (cm *ConversationManager) AddModelMessage(text string) {
	cm.History = append(cm.History, &genai.Content{
		Role:  "model",
		Parts: []genai.Part{genai.Text(text)},
	})
	logSnippet := text
	maxLogLen := 80
	if len(logSnippet) > maxLogLen {
		logSnippet = logSnippet[:maxLogLen] + "..."
	}
	cm.logger.Debug("[CONVO] Added Model Text: %q", logSnippet)
}

// AddModelResponse adds a model's response (potentially including function calls) to the history.
func (cm *ConversationManager) AddModelResponse(candidate *genai.Candidate) error {
	if candidate == nil || candidate.Content == nil {
		cm.logger.Warn("CONVO] Attempted to add nil candidate or candidate content.")
		return nil
	}
	if candidate.Content.Role == "" {
		candidate.Content.Role = "model"
	}
	if candidate.Content.Role != "model" {
		err := fmt.Errorf("attempted to add non-model content (Role: %s) as model response", candidate.Content.Role)
		cm.logger.Error("CONVO] %v", err)
		return err
	}
	cm.History = append(cm.History, candidate.Content)
	if len(candidate.Content.Parts) == 0 {
		cm.logger.Debug("[CONVO] Added Model response with no parts.")
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
		cm.logger.Debug("[CONVO] Added Model response with parts: [%s]", strings.Join(logMsgs, ", "))
	}
	return nil
}

// AddFunctionCallMessage adds a function call request from the model to the history.
func (cm *ConversationManager) AddFunctionCallMessage(fc genai.FunctionCall) {
	cm.History = append(cm.History, &genai.Content{
		Role:  "model", // The FunctionCall is part of the *model's* turn
		Parts: []genai.Part{fc},
	})
	cm.logger.Debug("[CONVO] Added Model FunctionCall request: %s", fc.Name)
}

// AddFunctionResultMessage adds a function response part to the history.
func (cm *ConversationManager) AddFunctionResultMessage(part genai.Part) error {
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
	cm.logger.Debug("[CONVO] Added FunctionResponse result for: %s", fr.Name)
	return nil
}

// GetHistory returns the current conversation history.
func (cm *ConversationManager) GetHistory() []*genai.Content { return cm.History }

// ClearHistory resets the conversation.
func (cm *ConversationManager) ClearHistory() {
	cm.History = make([]*genai.Content, 0)
	cm.logger.Debug("[CONVO] History cleared.")
}

// --- Helper for creating error responses (REMOVED - now lives in security_helpers.go) ---
/*
func CreateErrorFunctionResultPart(toolName string, execErr error) genai.Part {
	// ... implementation removed ...
}
*/

// --- Helper for parsing safety settings (Example, unchanged) ---
func parseSafetySettings(settings []string) ([]*genai.SafetySetting, error) {
	parsed := make([]*genai.SafetySetting, 0, len(settings))
	if len(settings) > 0 {
		log.Println("[WARN CONVO] parseSafetySettings function needs implementation.")
	}
	return parsed, nil
}
