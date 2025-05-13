// NeuroScript Version: 0.3.0
// File version: 0.1.10
// Call FormatWMStatusView from screen_wm_status.go for WM status display.
// filename: pkg/neurogo/update.go
// nlines: 350 // Approximate
// risk_rating: HIGH
package neurogo

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const maxEmitBufferLines = 200

// Update handles messages received by the TUI model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds                                                       []tea.Cmd
		localInputCmd, aiInputCmd, localOutputVPCmd, aiOutputVPCmd tea.Cmd
		spinnerCmd                                                 tea.Cmd
		keyHandled                                                 bool
	)

	m.localInput, localInputCmd = m.localInput.Update(msg)
	cmds = append(cmds, localInputCmd)
	m.aiInput, aiInputCmd = m.aiInput.Update(msg)
	cmds = append(cmds, aiInputCmd)

	if mMouseMsg, ok := msg.(tea.MouseMsg); ok && (mMouseMsg.Type == tea.MouseWheelDown || mMouseMsg.Type == tea.MouseWheelUp) {
		if m.focusIndex == focusLocalOutput {
			m.localOutput, localOutputVPCmd = m.localOutput.Update(msg)
		} else if m.focusIndex == focusAIOutput {
			m.aiOutput, aiOutputVPCmd = m.aiOutput.Update(msg)
		}
	} else {
		m.localOutput, localOutputVPCmd = m.localOutput.Update(msg)
		m.aiOutput, aiOutputVPCmd = m.aiOutput.Update(msg)
	}

	if localOutputVPCmd != nil {
		cmds = append(cmds, localOutputVPCmd)
	}
	if aiOutputVPCmd != nil {
		cmds = append(cmds, aiOutputVPCmd)
	}

	if m.isWaitingForAI || m.isSyncing || m.initialScriptRunning || m.patchStatus != "" {
		m.spinner, spinnerCmd = m.spinner.Update(msg)
		cmds = append(cmds, spinnerCmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keyMap.Quit) {
			m.quitting = true
			return m, tea.Quit
		}

		if m.initialScriptRunning || m.isSyncing {
			if key.Matches(msg, m.keyMap.Help) {
				m.helpVisible = !m.helpVisible
				m.help.ShowAll = m.helpVisible
				keyHandled = true
			}
			if keyHandled || msg.Type == tea.KeyCtrlC {
				return m, tea.Batch(cmds...)
			}
			return m, tea.Batch(cmds...)
		}

		if !keyHandled {
			if key.Matches(msg, m.keyMap.CycleLocalOutput) {
				m.localOutputDisplayMode = (m.localOutputDisplayMode + 1) % totalLocalOutputModes
				var newContent string
				displayModeName := "Unknown View"
				switch m.localOutputDisplayMode {
				case localOutputModeScript:
					newContent = strings.Join(m.emittedLines, "\n")
					displayModeName = "Script Output"
				case localOutputModeWMStatus:
					newContent = FormatWMStatusView(m.app) // Call new function
					displayModeName = "Worker Manager Status"
				}
				m.localOutput.SetContent(newContent)
				m.localOutput.GotoTop()
				m.addMessage("System", fmt.Sprintf("Local output view changed to: %s", displayModeName))
				keyHandled = true
			} else if key.Matches(msg, m.keyMap.Tab) {
				currentOrder := []int{focusLocalInput, focusLocalOutput, focusAIOutput, focusAIInput}
				currentIndex := -1
				for i, fi := range currentOrder {
					if fi == m.focusIndex {
						currentIndex = i
						break
					}
				}
				if currentIndex != -1 {
					m.focusIndex = currentOrder[(currentIndex+1)%len(currentOrder)]
				} else {
					m.focusIndex = focusLocalInput
				}
				m.updateFocus()
				keyHandled = true
			} else if key.Matches(msg, m.keyMap.ShiftTab) {
				currentOrder := []int{focusLocalInput, focusAIInput, focusAIOutput, focusLocalOutput}
				currentIndex := -1
				for i, fi := range currentOrder {
					if fi == m.focusIndex {
						currentIndex = i
						break
					}
				}
				if currentIndex != -1 {
					m.focusIndex = currentOrder[(currentIndex+1)%len(currentOrder)]
				} else {
					m.focusIndex = focusLocalInput
				}
				m.updateFocus()
				keyHandled = true
			} else if key.Matches(msg, m.keyMap.Help) {
				m.helpVisible = !m.helpVisible
				m.help.ShowAll = m.helpVisible
				keyHandled = true
			}

			if msg.Type == tea.KeyEnter && (m.focusIndex == focusLocalInput || m.focusIndex == focusAIInput) {
				if m.focusIndex == focusLocalInput {
					cmdValue := strings.TrimSpace(m.localInput.Value())
					m.addMessage("Command", cmdValue)
					m.localInput.Reset()
					m.lastError = nil
					switch {
					case cmdValue == "quit" || cmdValue == "/quit" || cmdValue == "exit":
						m.quitting = true
						return m, tea.Quit
					case cmdValue == "?" || cmdValue == "/help":
						m.helpVisible = !m.helpVisible
						m.help.ShowAll = m.helpVisible
						m.addMessage("System", fmt.Sprintf("Help toggled %v.", m.helpVisible))
					case cmdValue == "/sync":
						if !m.isSyncing {
							m.isSyncing = true
							m.currentActivity = "Syncing..."
							m.addMessage("System", m.currentActivity)
							cmds = append(cmds, m.spinner.Tick, m.runSyncCmd())
						} else {
							m.addMessage("System", "Sync already in progress.")
						}
					case strings.HasPrefix(cmdValue, "/run "):
						scriptPath := strings.TrimSpace(strings.TrimPrefix(cmdValue, "/run "))
						if scriptPath != "" {
							m.addMessage("System", fmt.Sprintf("Executing script: %s", scriptPath))
							m.initialScriptRunning = true
							m.currentActivity = fmt.Sprintf("Running: %s", filepath.Base(scriptPath))
							cmds = append(cmds, m.spinner.Tick, m.executeSpecificScriptCmd(scriptPath))
						} else {
							m.addMessage("System", "Usage: /run <path_to_script>")
						}
					default:
						m.addMessage("System", fmt.Sprintf("Unknown local command: '%s'", cmdValue))
						m.lastError = fmt.Errorf("unknown local command: %s", cmdValue)
					}
					keyHandled = true
				} else if m.focusIndex == focusAIInput {
					promptValue := strings.TrimSpace(m.aiInput.Value())
					if promptValue != "" {
						m.addMessage("You", promptValue)
						m.aiInput.Reset()
						m.aiOutput.GotoBottom()
						m.isWaitingForAI = true
						m.currentActivity = "AI is thinking..."
						m.lastError = nil
						m.addMessage("System", "Placeholder: AI query sent to processing logic...")
						cmds = append(cmds, m.spinner.Tick)
					}
					keyHandled = true
				}
			}

			if !keyHandled && (m.focusIndex == focusLocalOutput || m.focusIndex == focusAIOutput) {
				activeViewport := &m.localOutput
				if m.focusIndex == focusAIOutput {
					activeViewport = &m.aiOutput
				}
				scrollAmount := 1
				switch {
				case key.Matches(msg, m.keyMap.ScrollUp):
					activeViewport.LineUp(1)
					keyHandled = true
				case key.Matches(msg, m.keyMap.ScrollDown):
					activeViewport.LineDown(1)
					keyHandled = true
				case key.Matches(msg, m.keyMap.ScrollLeft):
					activeViewport.ScrollLeft(scrollAmount)
					keyHandled = true
				case key.Matches(msg, m.keyMap.ScrollRight):
					activeViewport.ScrollRight(scrollAmount)
					keyHandled = true
				case key.Matches(msg, m.keyMap.PageUp):
					activeViewport.ViewUp()
					keyHandled = true
				case key.Matches(msg, m.keyMap.PageDown):
					activeViewport.ViewDown()
					keyHandled = true
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.ready = true
		}
		m.setSizes(msg.Width, msg.Height)
		m.aiOutput.GotoBottom()
		m.localOutput.GotoBottom()

	case scriptEmitMsg:
		content := strings.TrimRight(msg.Content, "\n")
		m.addMessage("emit", content)

	case initialScriptDoneMsg:
		m.initialScriptRunning = false
		m.currentActivity = ""
		scriptBaseName := filepath.Base(msg.Path)
		if msg.Err != nil {
			m.lastError = msg.Err
			m.addMessage("System", errorStyle.Render(fmt.Sprintf("Initial script '%s' FAILED: %v", scriptBaseName, msg.Err)))
			if m.app.GetLogger() != nil {
				m.app.GetLogger().Error("Initial script execution failed", "path", msg.Path, "error", msg.Err)
			}
		} else {
			m.addMessage("System", fmt.Sprintf("Initial script '%s' completed successfully.", scriptBaseName))
			if m.app.GetLogger() != nil {
				m.app.GetLogger().Debug("Initial script execution succeeded", "path", msg.Path)
			}
		}
		m.localOutput.GotoBottom()

	case syncCompleteMsg:
		m.isSyncing = false
		m.currentActivity = ""
		m.lastError = msg.err
		summary := "Sync completed."
		if msg.stats != nil {
			summary = fmt.Sprintf("Sync: Up:%d Del:%d",
				If(msg.stats["files_uploaded"] != nil, msg.stats["files_uploaded"], 0).(int64),
				If(msg.stats["files_deleted_api"] != nil, msg.stats["files_deleted_api"], 0).(int64))
		}
		if msg.err != nil {
			summary = fmt.Sprintf("%s. Error: %v", summary, msg.err)
			m.addMessage("System", errorStyle.Render(summary))
		} else {
			m.addMessage("System", summary)
		}
		m.aiOutput.GotoBottom()

	case errMsg:
		m.lastError = msg.err
		m.isWaitingForAI = false
		m.isSyncing = false
		m.initialScriptRunning = false
		m.currentActivity = ""
		m.patchStatus = ""
		m.addMessage("System", errorStyle.Render(fmt.Sprintf("ERROR: %v", msg.err)))
		m.aiOutput.GotoBottom()
	}

	return m, tea.Batch(cmds...)
}

// updateFocus sets focus to the currently selected pane.
func (m *model) updateFocus() {
	m.localInput.Blur()
	m.aiInput.Blur()
	activePaneMsg := ""
	switch m.focusIndex {
	case focusLocalInput:
		m.localInput.Focus()
		activePaneMsg = "Focus: Local Input ($)"
	case focusAIInput:
		m.aiInput.Focus()
		activePaneMsg = "Focus: AI Input (>)"
	case focusLocalOutput:
		activePaneMsg = "Focus: Local Output (Scroll with keys/mouse)"
	case focusAIOutput:
		activePaneMsg = "Focus: AI Output (Scroll with keys/mouse)"
	}
	if activePaneMsg != "" {
		m.addMessage("System", activePaneMsg)
	}
}

// executeSpecificScriptCmd creates a command to run a script file.
func (m *model) executeSpecificScriptCmd(scriptPath string) tea.Cmd {
	return func() tea.Msg {
		if m.app == nil {
			return initialScriptDoneMsg{Path: scriptPath, Err: fmt.Errorf("application access not available to run script")}
		}
		if m.app.GetLogger() != nil {
			m.app.GetLogger().Debug("Executing script via TUI command...", "path", scriptPath)
		}
		ctxToUse := context.Background()
		if m.app.Context() != nil {
			ctxToUse = m.app.Context()
		}
		err := m.app.ExecuteScriptFile(ctxToUse, scriptPath)
		return initialScriptDoneMsg{Path: scriptPath, Err: err}
	}
}
