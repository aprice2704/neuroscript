// NeuroScript Version: 0.3.0
// File version: 0.2.2 // Corrected ErrorCodeNotFound to ErrorCodeKeyNotFound
// filename: pkg/neurogo/app_chat.go
package neurogo

import (
	"context"
	"fmt"
	"strconv"

	// Still needed for LastActivityAt in ChatSession, and potentially for logging.
	"github.com/aprice2704/neuroscript/pkg/core"
	// "github.com/google/uuid" // Alternative for session IDs
)

// --- Multi-Chat Session Management Methods ---

// CreateNewChatSession creates a new chat session with a worker of the given definitionID,
// adds it to the application's managed sessions, and sets it as the active session.
func (a *App) CreateNewChatSession(definitionID string) (*ChatSession, error) {
	a.chatMu.Lock()
	defer a.chatMu.Unlock()

	aiWM := a.GetAIWorkerManager() // Uses App's GetAIWorkerManager
	if aiWM == nil {
		return nil, core.NewRuntimeError(core.ErrorCodePreconditionFailed, "AIWorkerManager not available in App", nil)
	}

	// Get definition details for naming
	workerDef, err := aiWM.GetWorkerDefinition(definitionID)
	if err != nil {
		a.Log.Error("Failed to get worker definition for new chat session", "definitionID", definitionID, "error", err)
		// Use ErrorCodeKeyNotFound as the definition ID was not found
		return nil, core.NewRuntimeError(core.ErrorCodeKeyNotFound, fmt.Sprintf("worker definition '%s' not found", definitionID), err)
	}
	if workerDef == nil { // Should be caught by GetWorkerDefinition error, but defensive check.
		// Use ErrorCodeKeyNotFound as the definition ID resolved to a nil definition
		return nil, core.NewRuntimeError(core.ErrorCodeKeyNotFound, fmt.Sprintf("worker definition '%s' is nil (unexpected)", definitionID), nil)
	}

	// Spawn a new instance.
	newInstance, err := aiWM.SpawnWorkerInstance(definitionID, nil, nil)
	if err != nil {
		a.Log.Error("Failed to spawn worker instance for new chat session", "definitionID", definitionID, "error", err)
		return nil, err // err from SpawnWorkerInstance should be a RuntimeError
	}

	// Generate SessionID and DisplayName
	a.nextChatIDSuffix++
	// Ensure DefinitionID is not empty before slicing
	defIDPrefix := ""
	if len(workerDef.DefinitionID) > 0 {
		defIDPrefix = workerDef.DefinitionID[:min(8, len(workerDef.DefinitionID))]
	}
	sessionID := fmt.Sprintf("%s-%s-%d", workerDef.Name, defIDPrefix, a.nextChatIDSuffix)
	// Alternative using UUID: sessionID := uuid.NewString()

	displayName := fmt.Sprintf("%s #%d", workerDef.Name, a.nextChatIDSuffix)
	if len(displayName) > 30 { // Keep display name reasonably short for TUI
		// Ensure workerDef.Name is not empty before slicing
		namePrefix := workerDef.Name
		if len(workerDef.Name) > 20 {
			namePrefix = workerDef.Name[:20]
		}
		displayName = namePrefix + fmt.Sprintf("... #%d", a.nextChatIDSuffix)
	}

	chatSess := NewChatSession(sessionID, displayName, definitionID, newInstance)
	if a.chatSessions == nil { // Should have been initialized in App.initializeCoreComponents
		a.chatSessions = make(map[string]*ChatSession)
	}
	a.chatSessions[sessionID] = chatSess
	a.activeChatSessionID = sessionID

	a.Log.Info("New chat session created and activated",
		"sessionID", sessionID,
		"displayName", displayName,
		"definitionID", definitionID,
		"instanceID", newInstance.InstanceID)

	return chatSess, nil
}

// SetActiveChatSession sets the chat session with the given ID as the active one.
func (a *App) SetActiveChatSession(sessionID string) error {
	a.chatMu.Lock()
	defer a.chatMu.Unlock()

	if _, exists := a.chatSessions[sessionID]; !exists {
		// Use ErrorCodeKeyNotFound as the session ID was not found in the map
		return core.NewRuntimeError(core.ErrorCodeKeyNotFound, fmt.Sprintf("chat session with ID '%s' not found", sessionID), nil)
	}
	a.activeChatSessionID = sessionID
	a.Log.Info("Active chat session set", "sessionID", sessionID)
	return nil
}

// GetActiveChatSession retrieves the currently active chat session.
// Returns nil if no session is active or the active ID is invalid.
func (a *App) GetActiveChatSession() *ChatSession {
	a.chatMu.Lock() // Lock for reading activeChatSessionID and chatSessions
	defer a.chatMu.Unlock()

	if a.activeChatSessionID == "" {
		return nil
	}
	session, exists := a.chatSessions[a.activeChatSessionID]
	if !exists {
		a.Log.Warn("ActiveChatSessionID points to a non-existent session.", "sessionID", a.activeChatSessionID)
		return nil
	}
	return session
}

// GetChatSession retrieves a specific chat session by its ID.
// Returns nil if the session ID is not found.
func (a *App) GetChatSession(sessionID string) *ChatSession {
	a.chatMu.Lock()
	defer a.chatMu.Unlock()

	session, exists := a.chatSessions[sessionID]
	if !exists {
		return nil
	}
	return session
}

// SendChatMessageToActiveSession sends a message to the currently active chat session.
// It updates the session's conversation history with both the user's message and the AI's response.
func (a *App) SendChatMessageToActiveSession(ctx context.Context, message string) (*core.ConversationTurn, error) {
	activeSession := a.GetActiveChatSession() // This method handles its own locking for read
	if activeSession == nil {
		return nil, core.NewRuntimeError(core.ErrorCodePreconditionFailed, "no active chat session to send message to", nil)
	}
	if activeSession.WorkerInstance == nil {
		return nil, core.NewRuntimeError(core.ErrorCodeInternal, fmt.Sprintf("active session '%s' has a nil WorkerInstance", activeSession.SessionID), nil)
	}

	a.Log.Info("Sending message to active chat session",
		"sessionID", activeSession.SessionID,
		"definitionID", activeSession.DefinitionID,
		"instanceID", activeSession.WorkerInstance.InstanceID)

	// 1. Create and add user's turn to the session's history
	userTurn := &core.ConversationTurn{
		Role:    core.RoleUser,
		Content: message,
		// Timestamp removed
	}
	activeSession.AddTurn(userTurn) // AddTurn in ChatSession handles its own mutex

	aiResponseTurn, err := activeSession.WorkerInstance.ProcessChatMessage(ctx, message)
	if err != nil {
		a.Log.Error("Error processing chat message by worker instance",
			"sessionID", activeSession.SessionID,
			"instanceID", activeSession.WorkerInstance.InstanceID,
			"error", err)
		errorTurn := &core.ConversationTurn{
			Role:    core.RoleSystem,
			Content: fmt.Sprintf("Error processing message: %v", err),
			// Timestamp removed
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
}

// GetActiveChatHistory returns a copy of the conversation history of the active chat session.
func (a *App) GetActiveChatHistory() []*core.ConversationTurn {
	activeSession := a.GetActiveChatSession() // Handles its own locking
	if activeSession == nil {
		return []*core.ConversationTurn{}
	}
	return activeSession.GetConversationHistory() // ChatSession.GetConversationHistory handles its own locking
}

// GetActiveChatDetails returns details about the currently active chat session.
func (a *App) GetActiveChatDetails() (sessionID string, displayName string, definitionID string, instanceStatus core.AIWorkerInstanceStatus, isActive bool) {
	activeSession := a.GetActiveChatSession() // Handles its own locking
	if activeSession != nil && activeSession.WorkerInstance != nil {
		return activeSession.SessionID,
			activeSession.DisplayName,
			activeSession.DefinitionID,
			activeSession.WorkerInstance.Status,
			true
	}
	return "", "", "", "", false
}

// CloseChatSession removes a chat session from management.
// It also attempts to retire the associated AIWorkerInstance.
func (a *App) CloseChatSession(sessionID string) error {
	a.chatMu.Lock()
	defer a.chatMu.Unlock()

	sessionToClose, exists := a.chatSessions[sessionID]
	if !exists {
		// Use ErrorCodeKeyNotFound as the session ID was not found in the map
		return core.NewRuntimeError(core.ErrorCodeKeyNotFound, fmt.Sprintf("chat session with ID '%s' not found for closing", sessionID), nil)
	}
	if sessionToClose.WorkerInstance == nil {
		a.Log.Warn("Chat session to close has a nil WorkerInstance. Cannot retire.", "sessionID", sessionID)
	} else {
		a.Log.Info("Closing chat session", "sessionID", sessionID, "instanceID", sessionToClose.WorkerInstance.InstanceID)
		aiWM := a.GetAIWorkerManager()
		if aiWM != nil {
			err := aiWM.RetireWorkerInstance(
				sessionToClose.WorkerInstance.InstanceID,
				"Chat session closed by user",
				core.InstanceStatusRetiredCompleted,
				sessionToClose.WorkerInstance.SessionTokenUsage,
				nil,
			)
			if err != nil {
				a.Log.Warn("Failed to retire worker instance during chat session closure",
					"sessionID", sessionID,
					"instanceID", sessionToClose.WorkerInstance.InstanceID,
					"error", err)
			} else {
				a.Log.Info("Successfully retired worker instance for closed chat session",
					"instanceID", sessionToClose.WorkerInstance.InstanceID)
			}
		} else {
			a.Log.Warn("AIWorkerManager was nil, cannot retire instance for session.", "sessionID", sessionID)
		}
	}

	delete(a.chatSessions, sessionID)

	if a.activeChatSessionID == sessionID {
		a.activeChatSessionID = ""
		a.Log.Info("Active chat session was closed. No new active session set automatically.", "closedSessionID", sessionID)
	}

	return nil
}

// ListChatSessions returns a slice of all current chat sessions.
// The order is not guaranteed.
func (a *App) ListChatSessions() []*ChatSession {
	a.chatMu.Lock()
	defer a.chatMu.Unlock()

	if len(a.chatSessions) == 0 {
		return []*ChatSession{}
	}

	sessions := make([]*ChatSession, 0, len(a.chatSessions))
	for _, session := range a.chatSessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// --- Accessor Methods ---

// GetInterpreter safely retrieves the interpreter.
func (a *App) GetInterpreter() *core.Interpreter {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.interpreter
}

// GetAIWorkerManager safely retrieves the AIWorkerManager from the interpreter.
func (a *App) GetAIWorkerManager() *core.AIWorkerManager {
	a.mu.RLock()
	interpreter := a.interpreter
	a.mu.RUnlock()

	if interpreter == nil {
		errMsg := "GetAIWorkerManager called when App.interpreter is nil"
		if a.Log != nil {
			a.Log.Warn(errMsg)
		} else {
			fmt.Println("WARNING: " + errMsg)
		}
		return nil
	}
	return interpreter.AIWorkerManager()
}

// Helper to convert string to int
func safeStrToInt(s string, fallback int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return i
}
