// NeuroScript Version: 0.3.0
// File version: 0.2.6
// Removed wm package dependency and corrected all error code constants.
// filename: pkg/neurogo/app_chat.go
package neurogo

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// --- Multi-Chat Session Management Methods ---

func (a *App) CreateNewChatSession(definitionID string) (*ChatSession, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Since AIWorkerManager is removed, we cannot create new sessions this way for now.
	// This will need to be re-implemented. Returning an error.
	return nil, lang.NewRuntimeError(lang.ErrorCodeNotImplemented, "AIWorkerManager is not available, cannot create chat session", nil)

	/*
		// --- Original logic commented out ---
		aiWM := a.GetAIWorkerManager()
		if aiWM == nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodePreconditionFailed, "AIWorkerManager not available in App", nil)
		}

		workerDef, err := aiWM.GetWorkerDefinition(definitionID)
		if err != nil {
			a.Log.Error("Failed to get worker definition for new chat session", "definitionID", definitionID, "error", err)
			return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("worker definition '%s' not found", definitionID), err)
		}
		if workerDef == nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("worker definition '%s' is nil (unexpected)", definitionID), nil)
		}

		newInstance, err := aiWM.SpawnWorkerInstance(definitionID, nil, nil)
		if err != nil {
			a.Log.Error("Failed to spawn worker instance for new chat session", "definitionID", definitionID, "error", err)
			return nil, err
		}

		a.nextChatIDSuffix++
		defIDPrefix := ""
		if len(workerDef.DefinitionID) > 0 {
			defIDPrefix = workerDef.DefinitionID[:min(8, len(workerDef.DefinitionID))]
		}
		sessionIDPart := workerDef.Name
		if sessionIDPart == "" {
			sessionIDPart = "Chat"
		}
		sessionID := fmt.Sprintf("%s-%s-%d", sessionIDPart, defIDPrefix, a.nextChatIDSuffix)

		displayName := fmt.Sprintf("%s #%d", workerDef.Name, a.nextChatIDSuffix)
		if len(displayName) > 30 {
			namePrefix := workerDef.Name
			if len(workerDef.Name) > 20 {
				namePrefix = workerDef.Name[:20]
			}
			displayName = namePrefix + fmt.Sprintf("... #%d", a.nextChatIDSuffix)
		}

		chatSess := NewChatSession(sessionID, displayName, definitionID, newInstance)
		if a.chatSessions == nil {
			a.chatSessions = make(map[string]*ChatSession)
		}
		a.chatSessions[sessionID] = chatSess
		a.activeChatSessionID = sessionID

		a.Log.Info("New chat session created and activated (data)",
			"sessionID", sessionID,
			"displayName", displayName,
			"definitionID", definitionID,
			"instanceID", newInstance.InstanceID)

		if a.tui != nil {
			a.tui.CreateAndShowNewChatScreen(chatSess)
			a.Log.Info("Instructed TUI to show new chat screen", "sessionID", sessionID)
		} else {
			a.Log.Warn("a.tui is nil, cannot instruct TUI to show new chat screen", "sessionID", sessionID)
		}

		return chatSess, nil
	*/
}

func (a *App) SetActiveChatSession(sessionID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.chatSessions[sessionID]; !exists {
		return lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("chat session with ID '%s' not found", sessionID), nil)
	}
	a.activeChatSessionID = sessionID
	a.Log.Info("Active chat session set", "sessionID", sessionID)
	return nil
}

func (a *App) GetActiveChatSession() *ChatSession {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.activeChatSessionID == "" {
		return nil
	}
	session, exists := a.chatSessions[a.activeChatSessionID]
	if !exists {
		if a.Log != nil {
			a.Log.Warn("ActiveChatSessionID points to a non-existent session.", "sessionID", a.activeChatSessionID)
		}
		return nil
	}
	return session
}

func (a *App) GetChatSession(sessionID string) *ChatSession {
	a.mu.RLock()
	defer a.mu.RUnlock()

	session, exists := a.chatSessions[sessionID]
	if !exists {
		return nil
	}
	return session
}

func (a *App) SendChatMessageToActiveSession(ctx context.Context, message string) (*interfaces.ConversationTurn, error) {
	activeSession := a.GetActiveChatSession()
	if activeSession == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodePreconditionFailed, "no active chat session to send message to", nil)
	}
	// Since WorkerInstance is removed, commenting out this check.
	// if activeSession.WorkerInstance == nil {
	// 	return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("active session '%s' has a nil WorkerInstance", activeSession.SessionID), nil)
	// }

	a.Log.Info("Sending message to active chat session", "sessionID", activeSession.SessionID, "definitionID", activeSession.DefinitionID)

	userTurn := &interfaces.ConversationTurn{
		Role:    interfaces.RoleUser,
		Content: message,
	}
	activeSession.AddTurn(userTurn)

	// Since WorkerInstance is removed, we cannot process messages this way for now.
	// This will need to be re-implemented. Returning an error.
	return nil, lang.NewRuntimeError(lang.ErrorCodeNotImplemented, "WorkerInstance is not available to process chat message", nil)

	/*
		// --- Original logic commented out ---
		aiResponseTurn, err := activeSession.WorkerInstance.ProcessChatMessage(ctx, message)
		if err != nil {
			a.Log.Error("Error processing chat message by worker instance",
				"sessionID", activeSession.SessionID,
				"instanceID", activeSession.WorkerInstance.InstanceID,
				"error", err)
			errorTurn := &interfaces.ConversationTurn{
				Role:    interfaces.RoleSystem,
				Content: fmt.Sprintf("Error processing message: %v", err),
			}
			activeSession.AddTurn(errorTurn)
			return nil, err
		}

		if aiResponseTurn != nil {
			activeSession.AddTurn(aiResponseTurn)
			a.Log.Info("Received response from chat worker", "sessionID", activeSession.SessionID)
		} else {
			a.Log.Warn("ProcessChatMessage returned nil turn and nil error.", "sessionID", activeSession.SessionID)
		}

		return aiResponseTurn, nil
	*/
}

func (a *App) GetActiveChatHistory() []*interfaces.ConversationTurn {
	activeSession := a.GetActiveChatSession()
	if activeSession == nil {
		return []*interfaces.ConversationTurn{}
	}
	return activeSession.GetConversationHistory()
}

func (a *App) GetActiveChatDetails() (sessionID string, displayName string, definitionID string, instanceStatus string, isActive bool) {
	activeSession := a.GetActiveChatSession()
	if activeSession != nil { // Cannot check for WorkerInstance anymore
		return activeSession.SessionID,
			activeSession.DisplayName,
			activeSession.DefinitionID,
			"unknown", // Placeholder for status
			true
	}
	return "", "", "", "", false
}

func (a *App) CloseChatSession(sessionID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, exists := a.chatSessions[sessionID]
	if !exists {
		return lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("chat session with ID '%s' not found for closing", sessionID), nil)
	}

	delete(a.chatSessions, sessionID)

	if a.activeChatSessionID == sessionID {
		a.activeChatSessionID = ""
		a.Log.Info("Active chat session was closed. No new active session set automatically.", "closedSessionID", sessionID)
	}

	return nil
}

func (a *App) ListChatSessions() []*ChatSession {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(a.chatSessions) == 0 {
		return []*ChatSession{}
	}

	sessions := make([]*ChatSession, 0, len(a.chatSessions))
	for _, session := range a.chatSessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// GetAIWorkerManager is commented out as it depends on the removed wm package.
/*
func (a *App) GetAIWorkerManager() *wm.AIWorkerManager {
	a.mu.RLock()
	interpreter := a.interpreter
	a.mu.RUnlock()

	if interpreter == nil {
		if a.Log != nil {
			a.Log.Warn("GetAIWorkerManager called when App.interpreter is nil")
		} else {
			fmt.Println("WARNING: GetAIWorkerManager called when App.interpreter is nil")
		}
		return nil
	}
	// This method no longer exists on the interpreter
	// return interpreter.AIWorkerManager()
	return nil
}
*/

func safeStrToInt(s string, fallback int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return i
}
