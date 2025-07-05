// filename: pkg/neurogo/chat_session.go
// File version: 1.1
// Corrected undefined types for wm.AIWorkerInstance.
package neurogo

import (
	"sync"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// ChatSession represents an active, independent conversation with a specific AIWorkerInstance.
type ChatSession struct {
	SessionID    string // Unique identifier for this session (e.g., defID-timestamp or UUID)
	DisplayName  string // User-facing name, e.g., "C(AgentName)-1"
	DefinitionID string // ID of the AIWorkerDefinition used for this session

	//	WorkerInstance *wm.AIWorkerInstance           // The AI instance handling this chat
	Conversation []*interfaces.ConversationTurn // The history of this specific chat session

	CreatedAt      time.Time // Timestamp of when the session was created
	LastActivityAt time.Time // Timestamp of the last message or significant activity

	mu sync.RWMutex // Protects concurrent access to the Conversation slice
}

// NewChatSession creates and initializes a new chat session.
// The AIWorkerInstance should already be obtained/created by the AIWorkerManager.
func NewChatSession(sessionID, displayName, definitionID string) *ChatSession {
	// if instance == nil {
	// 	// This is a programming error; a session cannot exist without an instance.
	// 	panic(fmt.Sprintf("cannot create ChatSession '%s' (DisplayName: '%s') for DefinitionID '%s': AIWorkerInstance is nil",
	// 		sessionID, displayName, definitionID))
	// }
	return &ChatSession{
		SessionID:    sessionID,
		DisplayName:  displayName,
		DefinitionID: definitionID,
		//	WorkerInstance: instance,
		Conversation:   make([]*interfaces.ConversationTurn, 0), // Initialize with an empty history
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}
}

// AddTurn appends a conversation turn to this session's history in a thread-safe manner.
func (cs *ChatSession) AddTurn(turn *interfaces.ConversationTurn) {
	if cs == nil {
		// Log or handle error: attempt to add turn to a nil ChatSession
		return
	}
	if turn == nil {
		// Log or handle error: attempt to add a nil turn
		return
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Conversation = append(cs.Conversation, turn)
	cs.LastActivityAt = time.Now()
}

// GetConversationHistory returns a new slice containing pointers to the conversation turns
// for safe, concurrent-read access. The turns themselves are not deep-copied.
func (cs *ChatSession) GetConversationHistory() []*interfaces.ConversationTurn {
	if cs == nil {
		return []*interfaces.ConversationTurn{} // Return an empty slice if cs is nil
	}
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// Create a copy of the slice to prevent external modifications to the slice itself (e.g., reordering).
	// The interfaces.ConversationTurn pointers are copied.
	historyCopy := make([]*interfaces.ConversationTurn, len(cs.Conversation))
	copy(historyCopy, cs.Conversation)
	return historyCopy
}

// ClearConversation clears all messages from this session's history in a thread-safe manner.
func (cs *ChatSession) ClearConversation() {
	if cs == nil {
		return
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Conversation = make([]*interfaces.ConversationTurn, 0) // Replace with a new empty slice
	cs.LastActivityAt = time.Now()
}
