// filename: pkg/neurogo/tui_chat_methods.go
// NeuroScript Version: 0.4.0
// File version: 0.1.13 // Corrected invalid type assertion on chatScreen
// Description: Contains methods for TUI chat view management.

package neurogo

import (
	"fmt"	// Only for EscapeTviewTags if used in status messages, etc.
)

// CreateAndShowNewChatScreen implements TUIController.
// It accepts a *ChatSession and calls switchToChatViewAndUpdate directly.
func (tvP *tviewAppPointers) CreateAndShowNewChatScreen(session *ChatSession) {
	if tvP.tviewApp == nil {
		tvP.LogToDebugScreen("[CREATE_SHOW_CHAT] tviewApp is nil. Cannot update TUI for session %s.", session.SessionID)
		if tvP.app != nil && tvP.app.Log != nil {
			tvP.app.Log.Error("tviewApp is nil in CreateAndShowNewChatScreen", "sessionID", session.SessionID)
		}
		return
	}
	if session == nil {
		tvP.LogToDebugScreen("[CREATE_SHOW_CHAT] Received nil session. Aborting.")
		if tvP.app != nil && tvP.app.Log != nil {
			tvP.app.Log.Error("CreateAndShowNewChatScreen received nil session")
		}
		return
	}

	tvP.LogToDebugScreen("[CREATE_SHOW_CHAT] Attempting DIRECT TUI update for sessionID: %s, displayName: %s", session.SessionID, session.DisplayName)

	// Call directly as we are on the main TUI goroutine.
	tvP.switchToChatViewAndUpdate(session)

	tvP.LogToDebugScreen("[CREATE_SHOW_CHAT] DIRECT TUI update for session %s completed.", session.SessionID)
}

// switchToChatViewAndUpdate ensures the specified chat session's screen is visible and focused.
// It now accepts a *ChatSession object directly.
// This method MUST be called on the TUI's main goroutine.
func (tvP *tviewAppPointers) switchToChatViewAndUpdate(session *ChatSession) {
	if session == nil || session.SessionID == "" {
		tvP.LogToDebugScreen("[SWITCH_CHAT] Called with nil session or empty sessionID. Aborting.")
		return
	}
	sessionID := session.SessionID	// Extract for convenience
	tvP.LogToDebugScreen("[SWITCH_CHAT] Attempting to switch/update to sessionID: %s", sessionID)

	chatScreen, exists := tvP.chatScreenMap[sessionID]
	if !exists {
		tvP.LogToDebugScreen("[SWITCH_CHAT] Screen for session %s not in map. Creating new.", sessionID)
		initialTitle := session.DisplayName
		if initialTitle == "" {	// Fallback
			initialTitle = fmt.Sprintf("Chat: %s", session.DefinitionID)
		}

		chatScreen = NewChatConversationScreen(tvP.app, sessionID, initialTitle)
		tvP.chatScreenMap[sessionID] = chatScreen
		tvP.addScreen(chatScreen, false)	// Add to Pane B (rightScreens)
		tvP.LogToDebugScreen("[SWITCH_CHAT] Created and added new chat screen for session %s. Title: '%s'. Total right screens: %d. AI Output View pages: %d",
			sessionID, initialTitle, len(tvP.rightScreens), tvP.aiOutputView.GetPageCount())
	} else {
		tvP.LogToDebugScreen("[SWITCH_CHAT] Screen for session %s found in map.", sessionID)
		// chatScreen is already *ChatConversationScreen.
		// Call its updateTitle method to refresh its title based on the current session state.
		if chatScreen != nil {	// Should not be nil if exists == true, but good check.
			chatScreen.updateTitle()	// This will use the screen's own logic to refresh the title.
			tvP.LogToDebugScreen("[SWITCH_CHAT] Updated title for existing screen %s.", sessionID)
		}
	}

	// Proceed only if chatScreen is valid (it should be if !exists or if it existed and was not nil)
	if chatScreen == nil {
		tvP.LogToDebugScreen("[SWITCH_CHAT] chatScreen is nil for session %s after get/create. Aborting further TUI updates for this session.", sessionID)
		return
	}

	chatScreenIndex := tvP.getScreenIndex(chatScreen, false)
	if chatScreenIndex != -1 {
		tvP.LogToDebugScreen("[SWITCH_CHAT] Screen index for session %s in rightScreens is %d.", sessionID, chatScreenIndex)
		tvP.setScreen(chatScreenIndex, false)	// Make the chat screen visible

		tvP.LogToDebugScreen("[SWITCH_CHAT] Pane B (AI Output) should now show page for session %s. Attempting to focus AI Input Area (Pane D).", sessionID)

		aiInputAreaIndex := -1
		for i, prim := range tvP.focusablePrimitives {
			if prim == tvP.aiInputArea {
				aiInputAreaIndex = i
				break
			}
		}

		if aiInputAreaIndex != -1 {
			if tvP.currentFocusIndex != aiInputAreaIndex {
				delta := aiInputAreaIndex - tvP.currentFocusIndex
				tvP.LogToDebugScreen("[SWITCH_CHAT] Current focus index %d, target %d (AIInputArea). Calling dFocus with delta: %d", tvP.currentFocusIndex, aiInputAreaIndex, delta)
				tvP.dFocus(delta)
			} else {
				tvP.LogToDebugScreen("[SWITCH_CHAT] AI Input Area already has logical focus (index %d). Calling dFocus(0) to re-apply styles/focus.", tvP.currentFocusIndex)
				tvP.dFocus(0)
			}
		} else {
			tvP.LogToDebugScreen("[SWITCH_CHAT] AI Input Area (Pane D) not found in focusablePrimitives. Cannot programmatically set focus via dFocus.")
		}
	} else {
		// This case should ideally not be hit if chatScreen was successfully created/retrieved and added.
		tvP.LogToDebugScreen("[SWITCH_CHAT] CRITICAL ERROR: ChatScreen for session %s (Name: %s) NOT FOUND in rightScreens list.", sessionID, chatScreen.Name())
	}
}