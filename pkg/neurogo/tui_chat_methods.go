// NeuroScript Version: 0.4.0
// File version: 0.1.1 // Simplified by removing redundant calls, relying on onPanePageChange
// Description: Contains methods for TUI chat view management.

package neurogo

import (
	"fmt"
	// "github.com/rivo/tview" // Not directly needed if tviewAppPointers is defined elsewhere
	// "github.com/gdamore/tcell/v2" // Not directly needed
)

// switchToChatViewAndUpdate ensures the specified chat session's screen is visible and focused.
// If the screen doesn't exist for the sessionID, it creates it.
// Relies on onPanePageChange to handle content updates for the newly visible screen.
func (tvP *tviewAppPointers) switchToChatViewAndUpdate(sessionID string) {
	if sessionID == "" {
		tvP.LogToDebugScreen("[SWITCH_CHAT] Called with empty sessionID. Aborting.")
		return
	}
	tvP.LogToDebugScreen("[SWITCH_CHAT] Attempting to switch/update to sessionID: %s", sessionID)

	// All UI manipulations must happen on the main goroutine.
	//	tvP.tviewApp.QueueUpdateDraw(func() {
	chatScreen, exists := tvP.chatScreenMap[sessionID]
	if !exists {
		tvP.LogToDebugScreen("[SWITCH_CHAT] Screen for session %s not in map. Creating new.", sessionID)
		session := tvP.app.GetChatSession(sessionID) // Accesses app.chatMu
		if session == nil {
			tvP.LogToDebugScreen("[SWITCH_CHAT] CRITICAL: Failed to get session details for new chat screen: %s. Aborting switch.", sessionID)
			if tvP.statusBar != nil {
				tvP.statusBar.SetText(fmt.Sprintf("[red]Error: Chat session %s not found![-]", EscapeTviewTags(sessionID)))
			}
			return
		}
		initialTitle := session.DisplayName
		if initialTitle == "" {
			initialTitle = fmt.Sprintf("Chat: %s", session.DefinitionID)
		}

		chatScreen = NewChatConversationScreen(tvP.app, sessionID, initialTitle)
		tvP.chatScreenMap[sessionID] = chatScreen
		tvP.addScreen(chatScreen, false) // Add to Pane B (rightScreens) and to aiOutputView (Pages)
		tvP.LogToDebugScreen("[SWITCH_CHAT] Created and added new chat screen for session %s. Title: '%s'. Total right screens: %d", sessionID, initialTitle, len(tvP.rightScreens))
	} else {
		tvP.LogToDebugScreen("[SWITCH_CHAT] Screen for session %s found in map.", sessionID)
	}

	chatScreenIndex := tvP.getScreenIndex(chatScreen, false)
	if chatScreenIndex != -1 {
		tvP.LogToDebugScreen("[SWITCH_CHAT] Screen index for session %s is %d.", sessionID, chatScreenIndex)

		// Make the chat screen visible. This will trigger onPanePageChange,
		// which should handle UpdateConversation and title updates via Primitive().
		// onPanePageChange is also responsible for calling the screen's OnFocus method.
		tvP.setScreen(chatScreenIndex, false) // This calls SwitchToPage -> onPanePageChange

		// Note: The explicit calls to chatScreen.UpdateConversation() and chatScreen.Primitive()
		// that were here previously have been removed.
		// onPanePageChange (called by setScreen) is now expected to handle:
		// 1. Calling screener.OnFocus() (which for ChatConversationScreen just sets tview focus and scrolls).
		// 2. If the screener is a ChatConversationScreen, it calls UpdateConversation().
		// 3. Title updates happen when Primitive() is called by updateStatusText() (which is called by onPanePageChange and dFocus).

		tvP.LogToDebugScreen("[SWITCH_CHAT] Setting focus to AI Input Area (Pane D) after ensuring screen is visible.")
		targetFocusIndex := -1
		for i, prim := range tvP.focusablePrimitives {
			if prim == tvP.aiInputArea {
				targetFocusIndex = i
				break
			}
		}

		if targetFocusIndex != -1 {
			// Use tvP.currentFocusIndex which is maintained by dFocus
			if tvP.currentFocusIndex != targetFocusIndex {
				delta := targetFocusIndex - tvP.currentFocusIndex // dFocus handles posmod
				tvP.LogToDebugScreen("[SWITCH_CHAT] Current focus index %d, target %d (AIInputArea). dFocus delta: %d", tvP.currentFocusIndex, targetFocusIndex, delta)
				tvP.dFocus(delta)
			} else {
				tvP.LogToDebugScreen("[SWITCH_CHAT] AI Input Area already has logical focus (index %d). Re-applying styles with dFocus(0).", tvP.currentFocusIndex)
				tvP.dFocus(0) // Reapply styles and ensure actual tview focus
			}
		} else {
			tvP.LogToDebugScreen("[SWITCH_CHAT] AI Input Area (Pane D) not found in focusablePrimitives. Cannot set focus.")
		}
	} else {
		tvP.LogToDebugScreen("[SWITCH_CHAT] CRITICAL ERROR: ChatScreen for session %s (new or existing) not found in rightScreens list via getScreenIndex. This should not happen.", sessionID)
	}
	// tvP.updateStatusText() // updateStatusText is called by setScreen (via onPanePageChange) and dFocus, so likely not needed here.
	//	})
}
