// filename: pkg/core/conversation.go
package core

import (
	"fmt"
	"log"
	"strings" // For safety settings parsing
)

// ConversationManager holds the state of an interaction with the LLM agent.
type ConversationManager struct {
	History []GeminiContent // Stores the conversation turns
	// Add session-specific security config here later
	// allowlist []string
	// sandboxRoot string
	logger *log.Logger
}

// NewConversationManager creates a new manager.
func NewConversationManager(logger *log.Logger) *ConversationManager {
	return &ConversationManager{
		History: make([]GeminiContent, 0),
		logger:  logger,
	}
}

// AddUserMessage appends a user message to the history.
func (cm *ConversationManager) AddUserMessage(text string) {
	cm.History = append(cm.History, GeminiContent{
		Role:  "user",
		Parts: []GeminiPart{{Text: text}},
	})
	cm.logger.Printf("[CONVO] Added User: %q", text)
}

// AddModelResponse adds a model's response (text or function call) to the history.
func (cm *ConversationManager) AddModelResponse(candidate GeminiCandidate) error {
	// Need to ensure the candidate.Content is correctly structured for history
	if candidate.Content.Role == "" {
		candidate.Content.Role = "model" // Ensure role is set
	}
	if candidate.Content.Role != "model" {
		return fmt.Errorf("attempted to add non-model content (%s) as model response", candidate.Content.Role)
	}

	cm.History = append(cm.History, candidate.Content)

	// Log what was added
	if len(candidate.Content.Parts) > 0 {
		part := candidate.Content.Parts[0]
		if part.Text != "" {
			cm.logger.Printf("[CONVO] Added Model Text (snippet): %s...", part.Text[:min(len(part.Text), 80)])
		} else if part.FunctionCall != nil {
			cm.logger.Printf("[CONVO] Added Model FunctionCall: %s", part.FunctionCall.Name)
		} else {
			cm.logger.Printf("[CONVO] Added Model response with unknown part structure.")
		}
	} else {
		cm.logger.Printf("[CONVO] Added Model response with no parts.")
	}
	return nil
}

// AddFunctionResponse adds the result of a function execution back into the history.
// It MUST follow a model message containing the corresponding FunctionCall.
func (cm *ConversationManager) AddFunctionResponse(toolName string, responseData map[string]interface{}) error {
	// Basic validation: Ensure the last message was a model requesting a function call
	if len(cm.History) == 0 {
		return fmt.Errorf("cannot add function response: conversation history is empty")
	}
	lastContent := cm.History[len(cm.History)-1]
	if lastContent.Role != "model" || len(lastContent.Parts) == 0 || lastContent.Parts[0].FunctionCall == nil {
		return fmt.Errorf("cannot add function response: last message was not a model function call")
	}
	// Optional: Could check if lastContent.Parts[0].FunctionCall.Name matches toolName

	cm.History = append(cm.History, GeminiContent{
		Role: "function", // Gemini API v1beta uses "function" role for results
		Parts: []GeminiPart{{
			FunctionResponse: &GeminiFunctionResponse{
				Name:     toolName,
				Response: responseData,
			},
		}},
	})
	cm.logger.Printf("[CONVO] Added FunctionResponse for: %s", toolName)
	return nil
}

// GetHistory returns the current conversation history.
func (cm *ConversationManager) GetHistory() []GeminiContent {
	// Implement context window management here if needed (e.g., truncation)
	// For now, return the full history.
	return cm.History
}

// ClearHistory resets the conversation.
func (cm *ConversationManager) ClearHistory() {
	cm.History = make([]GeminiContent, 0)
	cm.logger.Printf("[CONVO] History cleared.")
}

// --- Helper for parsing safety settings (Example) ---
// This might belong elsewhere, maybe config loading, but shows parsing logic.
func parseSafetySettings(settings []string) ([]GeminiSafetySetting, error) {
	parsed := make([]GeminiSafetySetting, 0, len(settings))
	validThresholds := map[string]bool{
		"BLOCK_NONE":                       true,
		"BLOCK_LOW_AND_ABOVE":              true,
		"BLOCK_MEDIUM_AND_ABOVE":           true,
		"BLOCK_ONLY_HIGH":                  true,
		"HARM_BLOCK_THRESHOLD_UNSPECIFIED": true, // Default?
	}
	// Categories are like HARM_CATEGORY_SEXUALLY_EXPLICIT, HARM_CATEGORY_HATE_SPEECH, etc.
	for _, s := range settings {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid safety setting format: %q, expected CATEGORY=THRESHOLD", s)
		}
		category := strings.TrimSpace(parts[0])
		threshold := strings.TrimSpace(parts[1])
		if !validThresholds[threshold] {
			return nil, fmt.Errorf("invalid safety threshold: %q in setting %q", threshold, s)
		}
		// Basic category check (can be more specific)
		if !strings.HasPrefix(category, "HARM_CATEGORY_") {
			return nil, fmt.Errorf("invalid safety category format: %q in setting %q", category, s)
		}
		parsed = append(parsed, GeminiSafetySetting{
			Category:  category,
			Threshold: threshold,
		})
	}
	return parsed, nil
}
