// NeuroScript Version: 0.3.0
// File version: 0.1.1 // Corrected AI call logic
// filename: pkg/neurogo/update_initiate_ai_chat.go

package neurogo

import (
	"fmt"
	"strings"

	// "time" // Only if ConversationTurn needs a timestamp filled here

	"github.com/aprice2704/neuroscript/pkg/core" // Added import for core types
	tea "github.com/charmbracelet/bubbletea"
)

// initiateAIChatCall handles the logic for making an AI call from a ChatScreen.
// This version uses AIWorkerManager's ExecuteStatelessTask.
func (m *model) initiateAIChatCall(msg sendAIChatMsg) tea.Cmd {
	return func() tea.Msg {
		if m.app == nil || m.app.GetLogger() == nil {
			return errMsg{fmt.Errorf("app or logger not available in initiateAIChatCall")}
		}
		logger := m.app.GetLogger()
		logger.Info("Initiating AI Chat Call (Stateless)", "originScreen", msg.OriginatingScreenName, "instanceID", msg.InstanceID, "definitionID", msg.DefinitionID)

		aiWM := m.app.GetAIWorkerManager()
		if aiWM == nil {
			logger.Error("AIWorkerManager not available to initiate chat call.")
			return aiResponseMsg{TargetScreenName: msg.OriginatingScreenName, Err: fmt.Errorf("AIWorkerManager not available")}
		}

		// Construct the prompt from msg.History.
		// For a stateless call, typically the last user message or a summary of history is used.
		var promptBuilder strings.Builder
		if len(msg.History) > 0 {
			// This example takes the content of the last turn in history as the prompt.
			// You might need a more sophisticated way to build the prompt based on your conversation flow.
			lastTurn := msg.History[len(msg.History)-1]
			if lastTurn != nil && lastTurn.Content != "" { // Check for nil turn and empty content
				promptBuilder.WriteString(lastTurn.Content)
			} else {
				logger.Warn("initiateAIChatCall: Last history turn is nil or has empty content.")
				// Decide if this is an error or if an empty prompt is permissible.
				// For now, let's assume an empty prompt from empty history content is not an error,
				// but an empty history list is.
			}
		}

		if promptBuilder.Len() == 0 && len(msg.History) == 0 { // Check if prompt is still empty due to empty history
			logger.Warn("initiateAIChatCall: History is empty, cannot generate prompt.")
			return aiResponseMsg{TargetScreenName: msg.OriginatingScreenName, Err: fmt.Errorf("cannot make AI call with empty history")}
		}
		prompt := promptBuilder.String()

		// Get the LLMClient from the App.
		llmClient := m.app.GetLLMClient() // Assumes App has GetLLMClient() -> core.LLMClient
		if llmClient == nil {
			logger.Error("LLMClient not available via App to initiate chat call.")
			return aiResponseMsg{TargetScreenName: msg.OriginatingScreenName, Err: fmt.Errorf("LLMClient not available")}
		}

		configOverrides := make(map[string]interface{}) // Provide actual overrides if any

		// Call ExecuteStatelessTask with the correct signature.
		respTurnContent, _, err := aiWM.ExecuteStatelessTask(msg.DefinitionID, llmClient, prompt, configOverrides)
		if err != nil {
			logger.Error("AI stateless chat call failed", "error", err, "screen", msg.OriginatingScreenName)
			return aiResponseMsg{TargetScreenName: msg.OriginatingScreenName, Err: err}
		}

		// Wrap the string response in a core.ConversationTurn for aiResponseMsg.
		conversationTurn := &core.ConversationTurn{
			Role:    "model", // Or core.RoleModel if defined
			Content: respTurnContent,
			// Timestamp: time.Now(), // Optional: if your ConversationTurn struct needs it
		}

		logger.Info("AI stateless chat call successful", "screen", msg.OriginatingScreenName)
		return aiResponseMsg{TargetScreenName: msg.OriginatingScreenName, ResponseCandidate: conversationTurn}
	}
}
