// filename: pkg/core/conversation.go
package core

import (
	"errors" // Added for error creation
	"fmt"
	"log"

	// For safety settings parsing
	"github.com/google/generative-ai-go/genai"
)

// ConversationManager holds the state of an interaction with the LLM agent.
type ConversationManager struct {
	History []*genai.Content // Stores the conversation turns
	logger  *log.Logger
}

// NewConversationManager creates a new manager.
func NewConversationManager(logger *log.Logger) *ConversationManager {
	return &ConversationManager{
		History: make([]*genai.Content, 0),
		logger:  logger,
	}
}

// AddUserMessage appends a user message to the history.
func (cm *ConversationManager) AddUserMessage(text string) {
	cm.History = append(cm.History, &genai.Content{
		Role:  "user",
		Parts: []genai.Part{genai.Text(text)},
	})
	cm.logger.Printf("[CONVO] Added User: %q", text)
}

// AddModelResponse adds a model's response (text or function call) to the history.
func (cm *ConversationManager) AddModelResponse(candidate *genai.Candidate) error {
	if candidate == nil || candidate.Content == nil {
		cm.logger.Printf("[WARN CONVO] Attempted to add nil candidate or candidate content.")
		return nil
	}

	if candidate.Content.Role == "" {
		candidate.Content.Role = "model"
	}
	if candidate.Content.Role != "model" {
		err := fmt.Errorf("attempted to add non-model content (Role: %s) as model response", candidate.Content.Role)
		cm.logger.Printf("[ERROR CONVO] %v", err)
		return err
	}

	cm.History = append(cm.History, candidate.Content)

	if len(candidate.Content.Parts) > 0 {
		part := candidate.Content.Parts[0]
		if textPart, ok := part.(genai.Text); ok {
			cm.logger.Printf("[CONVO] Added Model Text (snippet): %s...", string(textPart)[:min(len(string(textPart)), 80)])
		} else if fcPart, ok := part.(genai.FunctionCall); ok {
			cm.logger.Printf("[CONVO] Added Model FunctionCall: %s", fcPart.Name)
		} else {
			cm.logger.Printf("[CONVO] Added Model response with unknown part type: %T", part)
		}
	} else {
		cm.logger.Printf("[CONVO] Added Model response with no parts.")
	}
	return nil
}

// AddFunctionResponse adds the result of a function execution back into the history.
func (cm *ConversationManager) AddFunctionResponse(toolName string, responseData map[string]interface{}) error {
	if len(cm.History) == 0 {
		err := errors.New("cannot add function response: conversation history is empty")
		cm.logger.Printf("[ERROR CONVO] %v", err)
		return err
	}
	lastContent := cm.History[len(cm.History)-1]
	if lastContent.Role != "model" || len(lastContent.Parts) == 0 {
		err := errors.New("cannot add function response: last message was not from model or had no parts")
		cm.logger.Printf("[ERROR CONVO] %v", err)
		return err
	}
	lastPart := lastContent.Parts[0]
	if _, ok := lastPart.(genai.FunctionCall); !ok {
		err := fmt.Errorf("cannot add function response: last message part was %T, not FunctionCall", lastPart)
		cm.logger.Printf("[ERROR CONVO] %v", err)
		return err
	}

	cm.History = append(cm.History, &genai.Content{
		Role: "function",
		Parts: []genai.Part{
			genai.FunctionResponse{
				Name:     toolName,
				Response: responseData,
			},
		},
	})
	cm.logger.Printf("[CONVO] Added FunctionResponse for: %s", toolName)
	return nil
}

// GetHistory returns the current conversation history.
func (cm *ConversationManager) GetHistory() []*genai.Content {
	return cm.History
}

// ClearHistory resets the conversation.
func (cm *ConversationManager) ClearHistory() {
	cm.History = make([]*genai.Content, 0)
	cm.logger.Printf("[CONVO] History cleared.")
}

// --- Helper for parsing safety settings (Example) ---
func parseSafetySettings(settings []string) ([]*genai.SafetySetting, error) {
	parsed := make([]*genai.SafetySetting, 0, len(settings))

	// --- *** MODIFIED: Commented out problematic mapping logic *** ---
	/*
		validThresholds := map[string]genai.HarmBlockThreshold{
			"BLOCK_NONE":                       genai.HarmBlockThresholdNone,
			"BLOCK_LOW_AND_ABOVE":              genai.HarmBlockThresholdLowAndAbove,
			"BLOCK_MEDIUM_AND_ABOVE":           genai.HarmBlockThresholdMediumAndAbove,
			"BLOCK_ONLY_HIGH":                  genai.HarmBlockThresholdOnlyHigh,
			"HARM_BLOCK_THRESHOLD_UNSPECIFIED": genai.HarmBlockThresholdUnspecified,
		}
		validCategories := map[string]genai.HarmCategory{
			"HARM_CATEGORY_HARASSMENT":        genai.HarmCategoryHarassment,
			"HARM_CATEGORY_HATE_SPEECH":       genai.HarmCategoryHateSpeech,
			"HARM_CATEGORY_SEXUALLY_EXPLICIT": genai.HarmCategorySexuallyExplicit,
			"HARM_CATEGORY_DANGEROUS_CONTENT": genai.HarmCategoryDangerousContent,
			"HARM_CATEGORY_UNSPECIFIED":       genai.HarmCategoryUnspecified,
		}

		for _, s := range settings {
			parts := strings.SplitN(s, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid safety setting format: %q, expected CATEGORY=THRESHOLD", s)
			}
			categoryStr := strings.TrimSpace(parts[0])
			thresholdStr := strings.TrimSpace(parts[1])

			category, catOk := validCategories[categoryStr]
			if !catOk {
				return nil, fmt.Errorf("invalid safety category string: %q in setting %q", categoryStr, s)
			}
			threshold, thrOk := validThresholds[thresholdStr]
			if !thrOk {
				return nil, fmt.Errorf("invalid safety threshold string: %q in setting %q", thresholdStr, s)
			}

			parsed = append(parsed, &genai.SafetySetting{
				Category:  category,
				Threshold: threshold,
			})
		}
	*/
	// --- End Modified Section ---

	// Return empty slice and nil error until properly implemented
	if len(settings) > 0 {
		log.Println("[WARN CONVO] parseSafetySettings function needs implementation to map strings to genai constants.")
	}
	return parsed, nil
}
