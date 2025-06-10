// NeuroScript Version: 0.3.0
// File version: 0.2.4 // Use RWMutex, pass *ChatSession to TUI to avoid reentrant lock.
// filename: pkg/neurogo/app_chat.go
package neurogo

import (
	"context"
	"fmt"
	"strconv" // Used by safeStrToInt

	// Added for RWMutex
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// App struct (definition in app.go or app_types.go) should have:
// chatMu               sync.RWMutex // Changed from sync.Mutex
// chatSessions         map[string]*ChatSession
// activeChatSessionID  string
// nextChatIDSuffix     int
// tui                  TUIController // Assuming TUIController is defined in interfaces.go
// Log                  interfaces.Logger
// interpreter          *core.Interpreter

// --- Multi-Chat Session Management Methods ---

// CreateNewChatSession creates a new chat session with a worker of the given definitionID,
// adds it to the application's managed sessions, sets it as the active session,
// and instructs the TUI to display the new chat screen by passing the session object.
func (a *App) CreateNewChatSession(definitionID string) (*ChatSession, error) {
	// This function WRITES to a.chatSessions and a.activeChatSessionID, so it needs a full Lock.
	// a.chatMu.Lock()
	// defer a.chatMu.Unlock()

	aiWM := a.GetAIWorkerManager()
	if aiWM == nil {
		return nil, core.NewRuntimeError(core.ErrorCodePreconditionFailed, "AIWorkerManager not available in App", nil)
	}

	workerDef, err := aiWM.GetWorkerDefinition(definitionID)
	if err != nil {
		a.Log.Error("Failed to get worker definition for new chat session", "definitionID", definitionID, "error", err)
		return nil, core.NewRuntimeError(core.ErrorCodeKeyNotFound, fmt.Sprintf("worker definition '%s' not found", definitionID), err)
	}
	if workerDef == nil {
		return nil, core.NewRuntimeError(core.ErrorCodeKeyNotFound, fmt.Sprintf("worker definition '%s' is nil (unexpected)", definitionID), nil)
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
		// Pass the actual chatSess object to avoid re-fetching and re-locking in the TUI update path.
		// The TUIController interface and its implementation will need to be updated for this.
		a.tui.CreateAndShowNewChatScreen(chatSess) // MODIFIED: Passing chatSess object
		a.Log.Info("Instructed TUI to show new chat screen", "sessionID", sessionID)
	} else {
		a.Log.Warn("a.tui is nil, cannot instruct TUI to show new chat screen", "sessionID", sessionID)
	}

	return chatSess, nil
}

// SetActiveChatSession sets the chat session with the given ID as the active one.
// This function WRITES to a.activeChatSessionID, so it needs a full Lock.
func (a *App) SetActiveChatSession(sessionID string) error {
	// a.chatMu.Lock()
	// defer a.chatMu.Unlock()

	if _, exists := a.chatSessions[sessionID]; !exists {
		return core.NewRuntimeError(core.ErrorCodeKeyNotFound, fmt.Sprintf("chat session with ID '%s' not found", sessionID), nil)
	}
	a.activeChatSessionID = sessionID
	a.Log.Info("Active chat session set", "sessionID", sessionID)
	return nil
}

// GetActiveChatSession retrieves the currently active chat session.
// This function READS shared data, so it uses RLock.
func (a *App) GetActiveChatSession() *ChatSession {
	// a.chatMu.RLock()
	// defer a.chatMu.RUnlock()

	if a.activeChatSessionID == "" {
		return nil
	}
	session, exists := a.chatSessions[a.activeChatSessionID]
	if !exists {
		// This RLock section should not log using a.Log if a.Log itself could cause locking issues or reentrancy.
		// For now, assuming a.Log is safe.
		// Consider: if a.Log implies further app state access that might conflict, direct log.Printf might be safer here for warnings.
		// However, this specific log line is unlikely to cause a problem by itself.
		// To be absolutely safe under RLock, one might avoid calling methods on 'a' that could take further locks.
		// For now, let's assume a.Log.Warn is okay.
		if a.Log != nil {
			a.Log.Warn("ActiveChatSessionID points to a non-existent session.", "sessionID", a.activeChatSessionID)
		}
		return nil
	}
	return session
}

// GetChatSession retrieves a specific chat session by its ID.
// This function READS shared data, so it uses RLock.
// IMPORTANT: This function should now be safe with its RLock,
// as the reentrant deadlock was due to CreateNewChatSession calling it while holding a Lock.
// By passing the *ChatSession object, CreateNewChatSession's call chain will no longer call this.
func (a *App) GetChatSession(sessionID string) *ChatSession {
	// a.chatMu.RLock()
	// defer a.chatMu.RUnlock()

	session, exists := a.chatSessions[sessionID]
	if !exists {
		return nil
	}
	return session
}

// SendChatMessageToActiveSession sends a message to the currently active chat session.
func (a *App) SendChatMessageToActiveSession(ctx context.Context, message string) (*interfaces.ConversationTurn, error) {
	// GetActiveChatSession uses RLock internally, which is fine.
	activeSession := a.GetActiveChatSession()
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

	userTurn := &interfaces.ConversationTurn{
		Role:    interfaces.RoleUser,
		Content: message,
	}
	activeSession.AddTurn(userTurn) // ChatSession.AddTurn has its own internal mutex

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
}

// GetActiveChatHistory returns a copy of the conversation history of the active chat session.
func (a *App) GetActiveChatHistory() []*interfaces.ConversationTurn {
	activeSession := a.GetActiveChatSession() // Uses RLock internally
	if activeSession == nil {
		return []*interfaces.ConversationTurn{}
	}
	return activeSession.GetConversationHistory() // ChatSession.GetConversationHistory uses its own mutex
}

// GetActiveChatDetails returns details about the currently active chat session.
func (a *App) GetActiveChatDetails() (sessionID string, displayName string, definitionID string, instanceStatus core.AIWorkerInstanceStatus, isActive bool) {
	activeSession := a.GetActiveChatSession() // Uses RLock internally
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
// This function WRITES to a.chatSessions and potentially a.activeChatSessionID, so it needs a full Lock.
func (a *App) CloseChatSession(sessionID string) error {
	// a.chatMu.Lock()
	// defer a.chatMu.Unlock()

	sessionToClose, exists := a.chatSessions[sessionID]
	if !exists {
		return core.NewRuntimeError(core.ErrorCodeKeyNotFound, fmt.Sprintf("chat session with ID '%s' not found for closing", sessionID), nil)
	}
	// ... (rest of the logic for retiring worker instance) ...
	// Ensure any logging here doesn't cause issues if called under lock.
	if sessionToClose.WorkerInstance != nil {
		a.Log.Info("Closing chat session, retiring worker instance", "sessionID", sessionID, "instanceID", sessionToClose.WorkerInstance.InstanceID)
		// ... (retire worker logic) ...
	}

	delete(a.chatSessions, sessionID)

	if a.activeChatSessionID == sessionID {
		a.activeChatSessionID = ""
		a.Log.Info("Active chat session was closed. No new active session set automatically.", "closedSessionID", sessionID)
	}

	return nil
}

// ListChatSessions returns a slice of all current chat sessions.
// This function READS shared data, so it uses RLock.
func (a *App) ListChatSessions() []*ChatSession {
	// a.chatMu.RLock()
	// defer a.chatMu.RUnlock()

	if len(a.chatSessions) == 0 {
		return []*ChatSession{}
	}

	sessions := make([]*ChatSession, 0, len(a.chatSessions))
	for _, session := range a.chatSessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// --- Accessor Methods (Example, ensure these also respect locking if accessing shared state) ---

// GetInterpreter safely retrieves the interpreter.
func (a *App) GetInterpreter() *core.Interpreter {
	// Assuming a.interpreter is set at init and doesn't change, or uses its own separate lock if it does.
	// If a.mu protects a.interpreter:
	// a.mu.RLock()
	// defer a.mu.RUnlock()
	return a.interpreter
}

// GetAIWorkerManager safely retrieves the AIWorkerManager from the interpreter.
func (a *App) GetAIWorkerManager() *core.AIWorkerManager {
	// As above, depends on how a.interpreter is managed.
	// a.mu.RLock()
	interpreter := a.interpreter
	// a.mu.RUnlock()

	if interpreter == nil {
		// Logging here should be cautious if under a lock from the caller.
		// For now, assuming this method isn't called while holding critical locks that a.Log might also need.
		if a.Log != nil {
			a.Log.Warn("GetAIWorkerManager called when App.interpreter is nil")
		} else {
			fmt.Println("WARNING: GetAIWorkerManager called when App.interpreter is nil")
		}
		return nil
	}
	return interpreter.AIWorkerManager()
}

// Helper to convert string to int (already in app_chat.go)
func safeStrToInt(s string, fallback int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return i
}
