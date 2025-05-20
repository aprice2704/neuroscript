package neurogo

import (
	"fmt"
	// "github.com/rivo/tview" // Not directly needed if tviewAppPointers is defined elsewhere
	// "github.com/gdamore/tcell/v2" // Not directly needed
)

// switchToChatViewAndUpdate ensures the specified chat session's screen is visible and focused.
// If the screen doesn't exist for the sessionID, it creates it.
func (tvP *tviewAppPointers) switchToChatViewAndUpdate(sessionID string) {
	if sessionID == "" {
		tvP.LogToDebugScreen("[SWITCH_CHAT] Called with empty sessionID. Aborting.")
		return
	}
	tvP.LogToDebugScreen("[SWITCH_CHAT] Attempting to switch/update to sessionID: %s", sessionID)

	// All UI manipulations must happen on the main goroutine.
	tvP.tviewApp.QueueUpdateDraw(func() {
		chatScreen, exists := tvP.chatScreenMap[sessionID]
		if !exists {
			tvP.LogToDebugScreen("[SWITCH_CHAT] Screen for session %s not in map. Creating new.", sessionID)
			session := tvP.app.GetChatSession(sessionID)
			if session == nil {
				tvP.LogToDebugScreen("[SWITCH_CHAT] CRITICAL: Failed to get session details for new chat screen: %s. Aborting switch.", sessionID)
				if tvP.statusBar != nil {
					tvP.statusBar.SetText(fmt.Sprintf("[red]Error: Chat session %s not found![-]", sessionID))
				}
				return // Return from the anonymous function
			}
			initialTitle := session.DisplayName
			if initialTitle == "" {
				initialTitle = fmt.Sprintf("Chat: %s", session.DefinitionID)
			}

			chatScreen = NewChatConversationScreen(tvP.app, sessionID, initialTitle)
			tvP.chatScreenMap[sessionID] = chatScreen
			tvP.addScreen(chatScreen, false) // Add to Pane B (rightScreens)
			tvP.LogToDebugScreen("[SWITCH_CHAT] Created and added new chat screen for session %s. Title: '%s'. Total right screens: %d", sessionID, initialTitle, len(tvP.rightScreens))
		} else {
			tvP.LogToDebugScreen("[SWITCH_CHAT] Screen for session %s found in map.", sessionID)
		}

		chatScreenIndex := tvP.getScreenIndex(chatScreen, false)
		if chatScreenIndex != -1 {
			tvP.LogToDebugScreen("[SWITCH_CHAT] Screen index for session %s is %d.", sessionID, chatScreenIndex)

			_, currentPrimitiveInPaneB := tvP.aiOutputView.GetFrontPage()
			var currentVisibleScreenerInPaneB PrimitiveScreener
			if currentPrimitiveInPaneB != nil {
				currentVisibleScreenerInPaneB, _ = tvP.getScreenerFromPrimitive(currentPrimitiveInPaneB, false)
			}

			if currentVisibleScreenerInPaneB != chatScreen {
				// Corrected the potentially problematic log line by pre-calculating the screener name string
				currentScreenerNameStr := "nil"
				if currentVisibleScreenerInPaneB != nil {
					currentScreenerNameStr = currentVisibleScreenerInPaneB.Name()
				}
				tvP.LogToDebugScreen("[SWITCH_CHAT] Current visible screener in Pane B (%s) is not target (%s). Switching.",
					currentScreenerNameStr,
					chatScreen.Name())
				tvP.setScreen(chatScreenIndex, false)
			} else {
				tvP.LogToDebugScreen("[SWITCH_CHAT] Target chat screen for session %s is already visible.", sessionID)
			}

			tvP.LogToDebugScreen("[SWITCH_CHAT] Updating conversation and title for session %s.", sessionID)
			chatScreen.UpdateConversation()
			chatScreen.Primitive()

			tvP.LogToDebugScreen("[SWITCH_CHAT] Setting focus to AI Input Area (Pane D).")
			targetFocusIndex := -1
			for i, prim := range tvP.focusablePrimitives {
				if prim == tvP.aiInputArea {
					targetFocusIndex = i
					break
				}
			}

			if targetFocusIndex != -1 {
				currentActualFocus := tvP.tviewApp.GetFocus()
				currentIndexInCycle := -1
				for i, p := range tvP.focusablePrimitives {
					if p == currentActualFocus {
						currentIndexInCycle = i
						break
					}
				}
				// Use currentFocusIndex as tracked by dFocus, ensuring it's valid
				if tvP.currentFocusIndex < 0 || tvP.currentFocusIndex >= len(tvP.focusablePrimitives) {
					if len(tvP.focusablePrimitives) > 0 {
						tvP.currentFocusIndex = 0
					}
				}
				// If a more precise current index was found in our cycle list, use it
				if currentIndexInCycle != -1 {
					tvP.currentFocusIndex = currentIndexInCycle
				}

				if tvP.currentFocusIndex != targetFocusIndex {
					delta := targetFocusIndex - tvP.currentFocusIndex
					tvP.LogToDebugScreen("[SWITCH_CHAT] Current focus index %d, target %d (AIInputArea). dFocus delta: %d", tvP.currentFocusIndex, targetFocusIndex, delta)
					tvP.dFocus(delta)
				} else {
					tvP.LogToDebugScreen("[SWITCH_CHAT] AI Input Area already determined to be focused (index %d). Re-applying styles with dFocus(0).", tvP.currentFocusIndex)
					tvP.dFocus(0)
				}
			} else {
				tvP.LogToDebugScreen("[SWITCH_CHAT] AI Input Area (Pane D) not found in focusablePrimitives. Cannot set focus.")
			}
		} else {
			tvP.LogToDebugScreen("[SWITCH_CHAT] CRITICAL ERROR: ChatScreen for session %s (new or existing) not found in rightScreens list via getScreenIndex. This should not happen.", sessionID)
		}
		tvP.updateStatusText()
	})
}
