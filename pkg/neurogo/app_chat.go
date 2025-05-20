package neurogo

import (
	"context"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// --- Chat Session Management Methods ---

// StartChatWithWorker attempts to start or resume a chat with a worker of the given definitionID.
// If a chat with the same definitionID is already active and the instance is usable, it's returned.
// Otherwise, a new instance is spawned.
func (a *App) StartChatWithWorker(definitionID string) (*core.AIWorkerInstance, error) {
	a.chatMu.Lock()
	defer a.chatMu.Unlock()

	aiWM := a.GetAIWorkerManager()
	if aiWM == nil {
		return nil, core.NewRuntimeError(core.ErrorCodePreconditionFailed, "AIWorkerManager not available in App", nil)
	}

	// Check if we already have an active chat with this definitionID
	if a.activeChatDefinitionID == definitionID && a.activeChatInstance != nil {
		// Verify instance status (e.g., not retired)
		// GetWorkerInstance will confirm if it's still in the activeInstances map of AIWM
		currentInstance, err := aiWM.GetWorkerInstance(a.activeChatInstanceID)
		if err == nil && currentInstance != nil &&
			currentInstance.Status != core.InstanceStatusRetiredCompleted &&
			currentInstance.Status != core.InstanceStatusRetiredError &&
			currentInstance.Status != core.InstanceStatusRetiredExhausted {
			a.Log.Info("Resuming active chat session", "definitionID", definitionID, "instanceID", a.activeChatInstanceID)
			return a.activeChatInstance, nil
		}
		// If instance is retired or not found in AIWM, clear it and proceed to spawn a new one.
		a.Log.Info("Previous chat instance for definition is unusable or not found, spawning new.", "definitionID", definitionID, "oldInstanceID", a.activeChatInstanceID)
		a.clearActiveChatUnsafe() // Clear previous if unusable
	} else if a.activeChatInstance != nil {
		// If there's an active chat, but for a different definition, clear it.
		a.Log.Info("Switching active chat definition.", "oldDefinitionID", a.activeChatDefinitionID, "newDefinitionID", definitionID)
		a.clearActiveChatUnsafe()
	}

	a.Log.Info("Starting new chat session", "definitionID", definitionID)
	// Spawn a new instance. For chat, we typically don't provide instanceConfigOverrides or initialFileContexts at this stage,
	// unless specific chat profiles are implemented.
	newInstance, err := aiWM.SpawnWorkerInstance(definitionID, nil, nil)
	if err != nil {
		a.Log.Error("Failed to spawn worker instance for chat", "definitionID", definitionID, "error", err)
		return nil, err // err from SpawnWorkerInstance should be a RuntimeError
	}

	a.activeChatInstance = newInstance
	a.activeChatDefinitionID = newInstance.DefinitionID
	a.activeChatInstanceID = newInstance.InstanceID
	a.Log.Info("Chat session started successfully", "definitionID", a.activeChatDefinitionID, "instanceID", a.activeChatInstanceID)

	return newInstance, nil
}

// SendChatMessageToActiveWorker sends a message to the currently active chat instance.
func (a *App) SendChatMessageToActiveWorker(ctx context.Context, message string) (*core.ConversationTurn, error) {
	a.chatMu.Lock()
	instance := a.activeChatInstance
	instanceID := a.activeChatInstanceID
	defID := a.activeChatDefinitionID
	a.chatMu.Unlock() // Unlock before potentially long-running call

	if instance == nil {
		return nil, core.NewRuntimeError(core.ErrorCodePreconditionFailed, "no active chat instance to send message to", nil)
	}

	a.Log.Info("Sending message to active chat worker", "definitionID", defID, "instanceID", instanceID)
	responseTurn, err := instance.ProcessChatMessage(ctx, message)
	if err != nil {
		a.Log.Error("Error processing chat message by worker instance", "definitionID", defID, "instanceID", instanceID, "error", err)
		// Potentially update app's view of instance status if error indicates retirement
		// For now, ProcessChatMessage handles instance's internal status.
		return nil, err
	}

	a.Log.Info("Received response from chat worker", "definitionID", defID, "instanceID", instanceID)
	return responseTurn, nil
}

// GetActiveChatHistory returns a copy of the conversation history of the active chat instance.
func (a *App) GetActiveChatHistory() []*core.ConversationTurn {
	a.chatMu.Lock()
	defer a.chatMu.Unlock()

	if a.activeChatInstance == nil || a.activeChatInstance.ConversationHistory == nil {
		return []*core.ConversationTurn{} // Return empty slice, not nil
	}

	// Return a copy to prevent modification issues if TUI iterates while history changes
	historyCopy := make([]*core.ConversationTurn, len(a.activeChatInstance.ConversationHistory))
	for i, turn := range a.activeChatInstance.ConversationHistory {
		turnCopy := *turn // Shallow copy of the turn struct
		historyCopy[i] = &turnCopy
	}
	return historyCopy
}

// GetActiveChatInstanceDetails returns details about the currently active chat.
func (a *App) GetActiveChatInstanceDetails() (definitionID string, instanceID string, status core.AIWorkerInstanceStatus, isActive bool) {
	a.chatMu.Lock()
	defer a.chatMu.Unlock()

	if a.activeChatInstance != nil {
		return a.activeChatDefinitionID, a.activeChatInstanceID, a.activeChatInstance.Status, true
	}
	return "", "", "", false
}

// ClearActiveChat clears the current active chat session information from the App.
// It does not necessarily retire the instance in the AIWorkerManager.
func (a *App) ClearActiveChat() {
	a.chatMu.Lock()
	defer a.chatMu.Unlock()
	a.clearActiveChatUnsafe()
}

// clearActiveChatUnsafe is an internal helper, assumes caller holds chatMu.
func (a *App) clearActiveChatUnsafe() {
	if a.activeChatInstance != nil {
		a.Log.Info("Clearing active chat session", "definitionID", a.activeChatDefinitionID, "instanceID", a.activeChatInstanceID)
		// Note: Retiring the instance here could be an option if App "owns" chat instances.
		// For now, just detaching. AIWM's policies or explicit commands would retire it.
		// Example: if needed: a.GetAIWorkerManager().RetireWorkerInstance(a.activeChatInstanceID, "Chat session ended by user", core.InstanceStatusRetiredCompleted, a.activeChatInstance.SessionTokenUsage, nil)
	}
	a.activeChatInstance = nil
	a.activeChatDefinitionID = ""
	a.activeChatInstanceID = ""
}

// AIWorkerManager returns the application's AIWorkerManager instance.
// It provides read-only access from other parts of the neurogo package.
func (a *App) AIWorkerManager() *core.AIWorkerManager {
	// This method implies interpreter is already set.
	// Adding a nil check for robustness, though design should ensure interpreter exists.
	a.mu.RLock()
	interp := a.interpreter
	a.mu.RUnlock()
	if interp == nil {
		a.Log.Warn("AIWorkerManager() called when App.interpreter is nil")
		return nil
	}
	return interp.AIWorkerManager()
}

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
		a.Log.Warn("GetAIWorkerManager called when interpreter is nil.")
		return nil
	}
	return interpreter.AIWorkerManager()
}
