// NeuroScript Version: 0.3.0
// File version: 0.2.4
// filename: pkg/neurogo/update.go
// nlines: 200 // Approximate
// risk_rating: HIGH
package neurogo

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	statusBarHeight              = 1
	inputAreaDefaultVisibleLines = 3
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keyMap.Quit) {
			m.quitting = true
			if screen := m.getActiveLeftScreen(); screen != nil {
				cmds = append(cmds, screen.Blur(m.app))
			}
			if screen := m.getActiveRightScreen(); screen != nil {
				cmds = append(cmds, screen.Blur(m.app))
			}
			return m, tea.Quit
		}
		if key.Matches(msg, m.keyMap.Help) {
			m.helpVisible = !m.helpVisible
			m.help.ShowAll = m.helpVisible
			m.currentActivity = If(m.helpVisible, "Help Visible", "Help Hidden").(string)
			cmds = append(cmds, m.updateFocusStates())
			return m, tea.Batch(cmds...)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.ready = true
		}

		helpLines := 0
		if m.helpVisible {
			helpLines = strings.Count(m.help.View(m.keyMap), "\n") + 1
		}

		inputContainerVPadding := m.leftInputArea.BlurredStyle.Base.GetVerticalFrameSize()
		inputAreaRenderSlotHeight := inputAreaDefaultVisibleLines
		screenContainerVPadding := screenPaneContainerStyle.GetVerticalFrameSize()

		screenHeight := m.height - statusBarHeight - helpLines - (inputAreaRenderSlotHeight + inputContainerVPadding) - screenContainerVPadding
		if screenHeight < 1 {
			screenHeight = 1
		}

		leftPaneWidth := m.width / 2
		rightPaneWidth := m.width - leftPaneWidth

		screenContentWidthLeft := leftPaneWidth - screenPaneContainerStyle.GetHorizontalFrameSize()
		if screenContentWidthLeft < 0 {
			screenContentWidthLeft = 0
		}
		screenContentWidthRight := rightPaneWidth - screenPaneContainerStyle.GetHorizontalFrameSize()
		if screenContentWidthRight < 0 {
			screenContentWidthRight = 0
		}

		for _, s := range m.leftScreens {
			s.SetSize(screenContentWidthLeft, screenHeight)
		}
		for _, s := range m.rightScreens {
			s.SetSize(screenContentWidthRight, screenHeight)
		}

		m.leftInputArea.SetWidth(leftPaneWidth - m.leftInputArea.BlurredStyle.Base.GetHorizontalFrameSize())
		m.rightInputArea.SetWidth(rightPaneWidth - m.rightInputArea.BlurredStyle.Base.GetHorizontalFrameSize())
		m.help.Width = m.width
		cmds = append(cmds, m.updateFocusStates())
		return m, tea.Batch(cmds...)

	case spinner.TickMsg:
		if m.isWaitingForAI || m.isSyncing || m.initialScriptRunning {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		if screen := m.getActiveLeftScreen(); screen != nil {
			_, screenCmd := screen.Update(msg, m.app)
			cmds = append(cmds, screenCmd)
		}
		if screen := m.getActiveRightScreen(); screen != nil {
			_, screenCmd := screen.Update(msg, m.app)
			cmds = append(cmds, screenCmd)
		}
		return m, tea.Batch(cmds...)

	case initialScriptDoneMsg:
		m.initialScriptRunning = false
		m.currentActivity = ""
		m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("Initial script '%s' done. Error: %v", msg.Path, msg.Err)})
		if msg.Err != nil {
			m.lastError = msg.Err
		}
		if screen := m.getActiveLeftScreen(); screen != nil && screen.Name() == "Script Output" {
			var newScreen Screen
			newScreen, cmd = screen.Update(refreshViewMsg{Timestamp: time.Now()}, m.app)
			m.leftScreens[m.currentLeftScreenIdx] = newScreen
			cmds = append(cmds, cmd)
		}

	case scriptEmitMsg:
		if len(m.emittedLines) >= maxEmitBufferLines {
			m.emittedLines = m.emittedLines[len(m.emittedLines)-maxEmitBufferLines+1:]
		}
		m.emittedLines = append(m.emittedLines, strings.TrimRight(msg.Content, "\n"))
		if screen := m.getActiveLeftScreen(); screen != nil && screen.Name() == "Script Output" {
			var newScreen Screen
			newScreen, cmd = screen.Update(msg, m.app)
			m.leftScreens[m.currentLeftScreenIdx] = newScreen
			cmds = append(cmds, cmd)
		}

	case syncCompleteMsg:
		m.isSyncing = false
		m.currentActivity = ""
		m.lastError = msg.err
		summary := "Sync completed."
		if msg.stats != nil {
			uploaded := If(msg.stats["files_uploaded"] != nil, msg.stats["files_uploaded"], int64(0)).(int64)
			deleted := If(msg.stats["files_deleted_api"] != nil, msg.stats["files_deleted_api"], int64(0)).(int64)
			summary = fmt.Sprintf("Sync: Up:%d Del:%d", uploaded, deleted)
		}
		if msg.err != nil {
			summary = fmt.Sprintf("%s. Error: %v", summary, msg.err)
		}
		m.systemMessages = append(m.systemMessages, message{"System", summary})

	case aiResponseMsg:
		m.isWaitingForAI = false
		m.currentActivity = ""
		if msg.Err != nil {
			m.lastError = msg.Err
			m.currentActivity = fmt.Sprintf("AI Error: %v", msg.Err)
		}
		foundScreen := false
		for i, screen := range m.rightScreens {
			if screen.Name() == msg.TargetScreenName {
				var updatedScreen Screen
				updatedScreen, cmd = screen.Update(msg, m.app)
				m.rightScreens[i] = updatedScreen
				cmds = append(cmds, cmd)
				foundScreen = true
				break
			}
		}
		if !foundScreen && m.app != nil && m.app.GetLogger() != nil {
			m.app.GetLogger().Warn("aiResponseMsg received for unknown or inactive screen", "target", msg.TargetScreenName)
		}

	case sendAIChatMsg:
		m.isWaitingForAI = true
		m.currentActivity = "AI Chat thinking..."
		cmds = append(cmds, m.initiateAIChatCall(msg)) // Definition in update_helpers.go

	case closeScreenMsg:
		cmds = append(cmds, m.handleCloseScreen(msg)) // Definition now in update_helpers.go

	case errMsg:
		m.lastError = msg.err
		m.currentActivity = fmt.Sprintf("ERROR: %v", msg.err)
		m.isWaitingForAI = false
		m.isSyncing = false
		m.initialScriptRunning = false
		if screen := m.getActiveLeftScreen(); screen != nil {
			_, screenCmd := screen.Update(msg, m.app)
			cmds = append(cmds, screenCmd)
		}
		if screen := m.getActiveRightScreen(); screen != nil {
			_, screenCmd := screen.Update(msg, m.app)
			cmds = append(cmds, screenCmd)
		}

	case refreshViewMsg:
		if screen := m.getActiveLeftScreen(); screen != nil && (msg.ScreenName == "" || msg.ScreenName == screen.Name()) {
			var newScreen Screen
			newScreen, cmd = screen.Update(msg, m.app)
			m.leftScreens[m.currentLeftScreenIdx] = newScreen
			cmds = append(cmds, cmd)
		}
		if screen := m.getActiveRightScreen(); screen != nil && (msg.ScreenName == "" || msg.ScreenName == screen.Name()) {
			var newScreen Screen
			newScreen, cmd = screen.Update(msg, m.app)
			m.rightScreens[m.currentRightScreenIdx] = newScreen
			cmds = append(cmds, cmd)
		}
	}

	keyHandled := false
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, m.keyMap.Quit) || key.Matches(keyMsg, m.keyMap.Help) {
			// Already handled
		} else {
			switch {
			case key.Matches(keyMsg, m.keyMap.Tab):
				cmds = append(cmds, m.cycleFocus(false))
				keyHandled = true
			case key.Matches(keyMsg, m.keyMap.ShiftTab):
				cmds = append(cmds, m.cycleFocus(true))
				keyHandled = true
			case key.Matches(keyMsg, m.keyMap.CycleLeftScreen):
				cmds = append(cmds, m.cycleScreen(true))
				keyHandled = true
			case key.Matches(keyMsg, m.keyMap.CycleRightScreen):
				cmds = append(cmds, m.cycleScreen(false))
				keyHandled = true
			case keyMsg.Type == tea.KeyEnter:
				keyHandled = true
				focusedGlobalInput := m.getFocusedGlobalInputArea()
				if focusedGlobalInput != nil {
					inputValue := strings.TrimSpace(focusedGlobalInput.Value())
					if strings.HasPrefix(inputValue, "//") {
						cmds = append(cmds, m.handleSystemCommand(inputValue))
						focusedGlobalInput.Reset()
					} else {
						activeScreen := m.getScreenForFocusedInput()
						if activeScreen != nil && activeScreen.GetInputBubble() != nil {
							activeScreen.GetInputBubble().SetValue(inputValue)
							if submitCmd := activeScreen.HandleSubmit(m.app); submitCmd != nil {
								cmds = append(cmds, submitCmd)
							}
							m.syncInputAreaWithScreen(focusedGlobalInput, activeScreen)
						} else {
							m.systemMessages = append(m.systemMessages, message{"System", "No active input target for Enter."})
							focusedGlobalInput.Reset()
						}
					}
				}

			default:
				focusedGlobalInput := m.getFocusedGlobalInputArea()
				if focusedGlobalInput != nil && focusedGlobalInput.Focused() {
					var currentGlobalInput *textarea.Model
					var activeScreenForInput Screen
					if m.focusTarget == FocusLeftInput {
						currentGlobalInput = &m.leftInputArea
						activeScreenForInput = m.getActiveLeftScreen()
					} else if m.focusTarget == FocusRightInput {
						currentGlobalInput = &m.rightInputArea
						activeScreenForInput = m.getActiveRightScreen()
					}

					if currentGlobalInput != nil {
						*currentGlobalInput, cmd = currentGlobalInput.Update(msg)
						cmds = append(cmds, cmd)
						if activeScreenForInput != nil && activeScreenForInput.GetInputBubble() != nil {
							activeScreenForInput.GetInputBubble().SetValue(currentGlobalInput.Value())
						}
						keyHandled = true
					}
				} else if !keyHandled { // If not handled by global input, pass to focused pane's screen
					var targetScreen Screen
					if m.focusTarget == FocusLeftPane {
						targetScreen = m.getActiveLeftScreen()
					} else if m.focusTarget == FocusRightPane {
						targetScreen = m.getActiveRightScreen()
					}

					if targetScreen != nil {
						var updatedScreen Screen
						updatedScreen, cmd = targetScreen.Update(msg, m.app)
						cmds = append(cmds, cmd)
						if m.focusTarget == FocusLeftPane {
							m.leftScreens[m.currentLeftScreenIdx] = updatedScreen
						} else if m.focusTarget == FocusRightPane { // Ensure it's specifically the right pane
							m.rightScreens[m.currentRightScreenIdx] = updatedScreen
						}
					}
				}
			}
		}
	}

	if m.ready {
		if screen := m.getActiveLeftScreen(); screen != nil {
			m.syncInputAreaWithScreen(&m.leftInputArea, screen)
		}
		if screen := m.getActiveRightScreen(); screen != nil {
			m.syncInputAreaWithScreen(&m.rightInputArea, screen)
		}
	}

	return m, tea.Batch(cmds...)
}

// Suggested location: pkg/neurogo/update_helpers.go
// Ensure this file has the necessary imports (fmt, tea) and package declaration (package neurogo)

// handleCloseScreen manages the removal of a screen from the UI.
func (m *model) handleCloseScreen(msg closeScreenMsg) tea.Cmd {
	var cmds []tea.Cmd
	var screenClosed bool

	// Check left screens
	for i, s := range m.leftScreens {
		if s.Name() == msg.ScreenName {
			if m.app != nil && m.app.GetLogger() != nil {
				m.app.GetLogger().Info("Closing left screen", "name", msg.ScreenName)
			}
			if s.Blur(m.app) != nil { // Assuming Screen interface has Blur
				cmds = append(cmds, s.Blur(m.app))
			}

			// Remove the screen
			m.leftScreens = append(m.leftScreens[:i], m.leftScreens[i+1:]...)
			screenClosed = true

			if len(m.leftScreens) == 0 {
				m.currentLeftScreenIdx = -1 // No active screen
			} else {
				// Adjust current index if necessary
				if i < m.currentLeftScreenIdx {
					m.currentLeftScreenIdx--
				} else if i == m.currentLeftScreenIdx {
					// If the closed screen was the active one, adjust index
					// Try to keep it valid, or move to the previous if it was last
					if m.currentLeftScreenIdx >= len(m.leftScreens) {
						m.currentLeftScreenIdx = len(m.leftScreens) - 1
					}
					// If list became empty and index was 0, it's now -1 (handled above)
					// If list not empty, and index was 0, it stays 0 (new screen at index 0)
				}
			}
			// If focus was on left pane or its input, update focus.
			// updateFocusStates will also sync the input area.
			if m.focusTarget == FocusLeftPane || m.focusTarget == FocusLeftInput {
				cmds = append(cmds, m.updateFocusStates())
			} else {
				// Still sync input area even if focus isn't directly on the left pane's input
				m.syncInputAreaWithScreen(&m.leftInputArea, m.getActiveLeftScreen())
			}
			break
		}
	}

	if screenClosed {
		m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("Screen '%s' closed.", msg.ScreenName)})
		return tea.Batch(cmds...)
	}

	// Check right screens if not found in left
	for i, s := range m.rightScreens {
		if s.Name() == msg.ScreenName {
			if m.app != nil && m.app.GetLogger() != nil {
				m.app.GetLogger().Info("Closing right screen", "name", msg.ScreenName)
			}
			if s.Blur(m.app) != nil { // Assuming Screen interface has Blur
				cmds = append(cmds, s.Blur(m.app))
			}

			m.rightScreens = append(m.rightScreens[:i], m.rightScreens[i+1:]...)
			screenClosed = true

			if len(m.rightScreens) == 0 {
				m.currentRightScreenIdx = -1
			} else {
				if i < m.currentRightScreenIdx {
					m.currentRightScreenIdx--
				} else if i == m.currentRightScreenIdx {
					if m.currentRightScreenIdx >= len(m.rightScreens) {
						m.currentRightScreenIdx = len(m.rightScreens) - 1
					}
				}
			}
			if m.focusTarget == FocusRightPane || m.focusTarget == FocusRightInput {
				cmds = append(cmds, m.updateFocusStates())
			} else {
				m.syncInputAreaWithScreen(&m.rightInputArea, m.getActiveRightScreen())
			}
			break
		}
	}

	if screenClosed {
		m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("Screen '%s' closed.", msg.ScreenName)})
	} else {
		if m.app != nil && m.app.GetLogger() != nil {
			m.app.GetLogger().Warn("Attempted to close screen not found", "name", msg.ScreenName)
		}
		m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("Could not close screen: '%s' (not found).", msg.ScreenName)})
	}

	return tea.Batch(cmds...)
}
