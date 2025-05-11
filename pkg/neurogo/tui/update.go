// NeuroScript Version: 0.3.0
// File version: 0.0.2 // Handle initialScriptDoneMsg, manage initialScriptRunning state and spinner.
// filename: pkg/neurogo/tui/update.go
// nlines: 230 // Approximate
// risk_rating: MEDIUM
package tui

import (
	"fmt"
	"path/filepath" // For filepath.Base
	"strings"

	// Import core for SecureFilePath and SyncDirectoryUpHelper
	// DO NOT import "github.com/aprice2704/neuroscript/pkg/neurogo"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const maxEmitBufferLines = 200 // Max number of EMIT lines to keep in the TUI buffer

// Update handles messages received by the TUI model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds           []tea.Cmd
		cmdInputCmd    tea.Cmd
		promptInputCmd tea.Cmd
		viewportCmd    tea.Cmd
		emitLogVPCmd   tea.Cmd
		spinnerCmd     tea.Cmd
		keyHandled     bool
	)

	// Handle component updates first, including our new emitLogViewport
	m.viewport, viewportCmd = m.viewport.Update(msg)
	m.emitLogViewport, emitLogVPCmd = m.emitLogViewport.Update(msg)
	cmds = append(cmds, viewportCmd, emitLogVPCmd)

	// Update spinner only if needed (covers AI, Sync, and initial script)
	// initialScriptRunning is set in newModel if a script path is provided
	if m.isWaitingForAI || m.isSyncing || m.initialScriptRunning {
		m.spinner, spinnerCmd = m.spinner.Update(msg)
		cmds = append(cmds, spinnerCmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Allow Ctrl+C to quit even during blocking operations like initial script or sync
		if key.Matches(msg, m.keyMap.Quit) {
			m.quitting = true
			// Add message only if not already in a critical blocking state that might clear/overwrite it
			if !m.initialScriptRunning && !m.isSyncing {
				m.addMessage("System", "Quitting...")
			}
			return m, tea.Quit
		}

		// If initial script is running, or syncing, generally ignore other key presses
		if m.initialScriptRunning || m.isSyncing {
			keyHandled = true // Effectively ignore other keys
			// We could allow specific keys like help ('?') if desired, but keeping it simple for now.
			// return m, tea.Batch(cmds...) // Return early, only processing spinner and viewport updates
		}

		if !keyHandled { // Process other keys if not in a blocking state or already handled
			switch {
			// case key.Matches(msg, m.keyMap.Quit): // Moved up
			// 	m.quitting = true
			// 	m.addMessage("System", "Quitting...")
			// 	return m, tea.Quit

			case key.Matches(msg, m.keyMap.Help):
				m.help.ShowAll = !m.help.ShowAll
				m.helpVisible = m.help.ShowAll
				if m.ready {
					m.setSizes(m.width, m.height)
				}
				m.viewport.GotoBottom()
				keyHandled = true

			case key.Matches(msg, m.keyMap.Tab):
				m.focusedInput = (m.focusedInput + 1) % 2
				if m.focusedInput == focusCommand {
					m.promptInput.Blur()
					m.commandInput.Focus()
					m.addMessage("System", "Focus: Command Input ($)")
				} else {
					m.commandInput.Blur()
					m.promptInput.Focus()
					m.addMessage("System", "Focus: Prompt Input (>)")
				}
				keyHandled = true

			case msg.Type == tea.KeyEnter:
				if m.focusedInput == focusCommand {
					cmdValue := strings.TrimSpace(m.commandInput.Value())
					m.addMessage("Command", cmdValue)
					m.commandInput.Reset()
					m.lastError = nil

					switch { // Use switch without expression for cleaner prefix checking
					case cmdValue == "quit" || cmdValue == "/quit" || cmdValue == "exit":
						m.quitting = true
						m.addMessage("System", "Quitting...")
						return m, tea.Quit
					case cmdValue == "?" || cmdValue == "/help":
						m.help.ShowAll = !m.help.ShowAll
						m.helpVisible = m.help.ShowAll
						if m.ready {
							m.setSizes(m.width, m.height)
						}
						m.addMessage("System", fmt.Sprintf("Help toggled %v", m.helpVisible))
					case cmdValue == "/sync":
						if m.isSyncing {
							m.addMessage("System", "Sync already in progress.")
						} else {
							m.isSyncing = true
							m.currentActivity = "Syncing files..."
							m.lastError = nil
							m.addMessage("System", m.currentActivity)
							cmds = append(cmds, m.spinner.Tick, m.runSyncCmd())
						}
					case strings.HasPrefix(cmdValue, "/run "):
						scriptPath := strings.TrimSpace(strings.TrimPrefix(cmdValue, "/run "))
						if scriptPath != "" {
							m.addMessage("System", fmt.Sprintf("Executing script: %s", scriptPath))
							m.initialScriptRunning = true // Use this state for any TUI-triggered script
							m.currentActivity = fmt.Sprintf("Running: %s...", filepath.Base(scriptPath))
							// We need a way to differentiate the "done" message for this script
							// vs the initial startup script. For now, initialScriptDoneMsg might be okay,
							// or we create a generic scriptDoneMsg.
							// For simplicity, let's reuse initialScriptDoneMsg and the logic.
							cmds = append(cmds, m.spinner.Tick, m.executeSpecificScriptCmd(scriptPath))
						} else {
							m.addMessage("System", "Usage: /run <path_to_script.ns>")
						}
					default:
						m.addMessage("System", fmt.Sprintf("Unknown command: '%s'", cmdValue))
						m.lastError = fmt.Errorf("unknown command: %s", cmdValue)
					}
					keyHandled = true

				} else if m.focusedInput == focusPrompt {
					promptValue := strings.TrimSpace(m.promptInput.Value())
					if promptValue != "" {
						m.addMessage("You", promptValue)
						m.promptInput.Reset()
						m.viewport.GotoBottom()
						m.isWaitingForAI = true
						m.currentActivity = "Querying AI..." // More generic
						m.lastError = nil
						cmds = append(cmds, m.spinner.Tick)
						// cmds = append(cmds, m.sendPromptToAICmd(promptValue)) // This would be your actual AI call
						m.addMessage("System", "Sending to AI...") // Placeholder until AI call is made
					}
					keyHandled = true
				}
			}

			if !keyHandled {
				if m.focusedInput == focusCommand {
					m.commandInput, cmdInputCmd = m.commandInput.Update(msg)
					cmds = append(cmds, cmdInputCmd)
				} else {
					m.promptInput, promptInputCmd = m.promptInput.Update(msg)
					cmds = append(cmds, promptInputCmd)
				}
			}
		} // End of !keyHandled block

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.setSizes(msg.Width, msg.Height)
		if !m.ready {
			m.ready = true
		}
		m.viewport.GotoBottom()
		m.emitLogViewport.GotoBottom()

	case scriptEmitMsg:
		content := strings.TrimRight(msg.Content, "\n")
		m.emittedLines = append(m.emittedLines, content)
		if len(m.emittedLines) > maxEmitBufferLines {
			m.emittedLines = m.emittedLines[len(m.emittedLines)-maxEmitBufferLines:]
		}
		m.emitLogViewport.SetContent(strings.Join(m.emittedLines, "\n"))
		m.emitLogViewport.GotoBottom()

	case initialScriptDoneMsg:
		m.initialScriptRunning = false
		m.currentActivity = "" // Clear general activity

		if msg.Err != nil {
			m.lastError = msg.Err
			errorMsg := fmt.Sprintf("Script '%s' failed: %v", filepath.Base(msg.Path), msg.Err)
			m.addMessage("System", errorStyle.Render(errorMsg))
			if m.app.GetLogger() != nil {
				m.app.GetLogger().Error("Script execution failed in TUI", "path", msg.Path, "error", msg.Err)
			}
		} else {
			successMsg := fmt.Sprintf("Script '%s' completed.", filepath.Base(msg.Path))
			m.addMessage("System", successMsg)
			if m.app.GetLogger() != nil {
				m.app.GetLogger().Debug("Script execution successful in TUI", "path", msg.Path)
			}
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		m.emitLogViewport.GotoBottom()

	case syncCompleteMsg:
		m.isSyncing = false
		m.currentActivity = ""
		m.lastError = msg.err
		summary := "Sync complete."
		if msg.stats != nil {
			var uploads, deletes, uploadErrors, deleteErrors, listErrors, hashErrors, walkErrors, ignored, scanned, upToDate, updatedAPI int64
			if val, ok := msg.stats["files_uploaded"].(int64); ok {
				uploads = val
			}
			if val, ok := msg.stats["files_deleted_api"].(int64); ok {
				deletes = val
			}
			if val, ok := msg.stats["upload_errors"].(int64); ok {
				uploadErrors = val
			}
			// ... (rest of stat handling as in your existing file) ...
			totalErrors := uploadErrors + deleteErrors + listErrors + hashErrors + walkErrors
			m.syncUploads = int(uploads + updatedAPI)
			m.syncDeletes = int(deletes)
			summary = fmt.Sprintf("Sync done. Scanned:%d Ignored:%d UpToDate:%d Uploaded:%d Updated:%d Deleted:%d Errors:%d",
				scanned, ignored, upToDate, uploads, updatedAPI, deletes, totalErrors)

		}
		if msg.err != nil {
			summary = fmt.Sprintf("%s Error: %v", summary, msg.err)
			m.addMessage("System", errorStyle.Render(summary))
		} else {
			m.addMessage("System", summary)
		}
		m.viewport.GotoBottom()

	case errMsg:
		m.lastError = msg.err
		m.isWaitingForAI = false
		m.isSyncing = false
		m.initialScriptRunning = false // Clear this flag on any error too
		m.currentActivity = ""
		m.addMessage("System", errorStyle.Render(fmt.Sprintf("Error: %v", msg.err)))
		m.viewport.GotoBottom()
	}

	if m.focusedInput == focusCommand {
		if !m.commandInput.Focused() {
			m.commandInput.Focus()
		}
		m.promptInput.Blur()
	} else {
		if !m.promptInput.Focused() {
			m.promptInput.Focus()
		}
		m.commandInput.Blur()
	}

	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()
	m.emitLogViewport.SetContent(strings.Join(m.emittedLines, "\n")) // Ensure emit log always up to date
	m.emitLogViewport.GotoBottom()

	return m, tea.Batch(cmds...)
}

// executeSpecificScriptCmd is a helper to create a command for running a script by path.
// This can be used by a /run command, for example.
func (m *model) executeSpecificScriptCmd(scriptPath string) tea.Cmd {
	return func() tea.Msg {
		if m.app == nil {
			return initialScriptDoneMsg{Path: scriptPath, Err: fmt.Errorf("app service not available to run script")}
		}
		if m.app.GetLogger() != nil {
			m.app.GetLogger().Debug("Executing script from TUI command...", "path", scriptPath)
		}
		err := m.app.ExecuteScriptFile(m.app.Context(), scriptPath) // Assuming app has Context() or use context.Background()
		return initialScriptDoneMsg{Path: scriptPath, Err: err}
	}
}
