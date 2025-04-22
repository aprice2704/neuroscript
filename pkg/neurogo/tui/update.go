// filename: pkg/neurogo/tui/update.go
package tui

import (
	"fmt"
	"strings"

	// Import core for SecureFilePath and SyncDirectoryUpHelper

	// DO NOT import "github.com/aprice2704/neuroscript/pkg/neurogo"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"

	// Keep lipgloss for styles potentially used here
	tea "github.com/charmbracelet/bubbletea"
)

// --- TUI Update Function ---

// Init runs initialization commands for the TUI model.
func (m model) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages received by the TUI model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds           []tea.Cmd
		cmdInputCmd    tea.Cmd
		promptInputCmd tea.Cmd
		viewportCmd    tea.Cmd
		spinnerCmd     tea.Cmd
		keyHandled     bool
	)

	// Handle component updates first
	m.viewport, viewportCmd = m.viewport.Update(msg)
	cmds = append(cmds, viewportCmd)

	// Update spinner only if needed
	if m.isWaitingForAI || m.isSyncing || m.currentActivity != "" {
		m.spinner, spinnerCmd = m.spinner.Update(msg)
		cmds = append(cmds, spinnerCmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global keys first
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			m.quitting = true
			m.addMessage("System", "Quitting...")
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.helpVisible = m.help.ShowAll
			if m.ready {
				m.setSizes(m.width, m.height) // Recalculate layout
			}
			m.viewport.GotoBottom()
			keyHandled = true

		case key.Matches(msg, m.keyMap.Tab):
			if m.isSyncing { // Ignore tab during sync
				return m, nil
			}
			// m.addMessage("Debug", fmt.Sprintf("Tab pressed. Old focus: %d", m.focusedInput)) // Optional debug
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

		// Handle Enter specifically for the focused input
		case msg.Type == tea.KeyEnter:
			if m.isSyncing { // Ignore Enter during sync
				return m, nil
			}

			if m.focusedInput == focusCommand {
				cmdValue := strings.TrimSpace(m.commandInput.Value())
				m.addMessage("Command", cmdValue) // Echo command
				m.commandInput.Reset()
				m.lastError = nil

				switch cmdValue {
				case "quit", "/quit", "exit":
					m.quitting = true
					m.addMessage("System", "Quitting...")
					return m, tea.Quit // Return immediately on quit
				case "?", "/help":
					m.help.ShowAll = !m.help.ShowAll
					m.helpVisible = m.help.ShowAll
					if m.ready {
						m.setSizes(m.width, m.height) // Recalculate layout
					}
					m.addMessage("System", fmt.Sprintf("Help toggled %v", m.helpVisible))
				case "/sync":
					if m.isSyncing {
						m.addMessage("System", "Sync already in progress.")
					} else {
						m.isSyncing = true
						m.currentActivity = "Syncing files..."
						m.lastError = nil // Clear previous errors
						m.addMessage("System", m.currentActivity)
						// Call the method on the model, not the standalone function
						cmds = append(cmds, m.spinner.Tick, m.runSyncCmd())
					}
				default:
					m.addMessage("System", fmt.Sprintf("Unknown command: '%s'", cmdValue))
					m.lastError = fmt.Errorf("unknown command: %s", cmdValue)
				}
				keyHandled = true // Enter in command input is handled

			} else if m.focusedInput == focusPrompt {
				// Assuming Enter submits the prompt
				promptValue := strings.TrimSpace(m.promptInput.Value())
				if promptValue != "" {
					m.addMessage("You", promptValue)
					m.promptInput.Reset()
					m.viewport.GotoBottom()
					m.isWaitingForAI = true
					m.currentActivity = "Waiting for AI..."
					m.lastError = nil
					cmds = append(cmds, m.spinner.Tick)
					m.addMessage("System", "Sending prompt to AI...")
					// TODO: Implement command to send prompt via interface
					// Example: cmds = append(cmds, m.sendPromptCmd(promptValue))
				}
				keyHandled = true // Enter in prompt input is handled
			}
		} // End inner switch for specific keys

		// If the key wasn't handled globally or by Enter, pass it to the focused input
		if !keyHandled && !m.isSyncing { // Don't process other keys during sync
			if m.focusedInput == focusCommand {
				m.commandInput, cmdInputCmd = m.commandInput.Update(msg)
				cmds = append(cmds, cmdInputCmd)
			} else {
				m.promptInput, promptInputCmd = m.promptInput.Update(msg)
				cmds = append(cmds, promptInputCmd)
			}
		}

	// --- Other Message Types ---
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// setSizes needs to be called whether ready or not to handle resize
		m.setSizes(msg.Width, msg.Height)
		if !m.ready {
			m.ready = true
			// No need to SetContent here, setSizes should handle it
		}
		m.viewport.GotoBottom()

	case syncCompleteMsg:
		m.isSyncing = false
		m.currentActivity = ""
		m.lastError = msg.err // Store potential sync error

		summary := "Sync complete."
		if msg.stats != nil {
			// Safely extract stats using type assertion with checking
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
			if val, ok := msg.stats["delete_errors"].(int64); ok {
				deleteErrors = val
			}
			if val, ok := msg.stats["list_api_errors"].(int64); ok {
				listErrors = val
			}
			if val, ok := msg.stats["hash_errors"].(int64); ok {
				hashErrors = val
			}
			if val, ok := msg.stats["walk_errors"].(int64); ok {
				walkErrors = val
			}
			if val, ok := msg.stats["files_ignored"].(int64); ok {
				ignored = val
			}
			if val, ok := msg.stats["files_scanned"].(int64); ok {
				scanned = val
			}
			if val, ok := msg.stats["files_up_to_date"].(int64); ok {
				upToDate = val
			}
			if val, ok := msg.stats["files_updated_api"].(int64); ok {
				updatedAPI = val
			}

			totalErrors := uploadErrors + deleteErrors + listErrors + hashErrors + walkErrors

			// Update status bar info (convert int64 to int)
			m.syncUploads = int(uploads + updatedAPI)
			m.syncDeletes = int(deletes)
			// TODO: Update local/api file counts? Requires another command/interface method

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
		m.isWaitingForAI = false // Stop spinner if error occurred
		m.isSyncing = false      // Stop sync state if error occurred
		m.currentActivity = ""
		m.addMessage("System", errorStyle.Render(fmt.Sprintf("Error: %v", msg.err)))
		m.viewport.GotoBottom()

	} // End main switch msg.(type)

	// --- Final Focus Update ---
	// Ensure visual focus matches state after processing message
	if m.focusedInput == focusCommand {
		if !m.commandInput.Focused() {
			m.commandInput.Focus() // Re-focus if necessary
		}
		m.promptInput.Blur()
	} else {
		if !m.promptInput.Focused() {
			m.promptInput.Focus() // Re-focus if necessary
		}
		m.commandInput.Blur()
	}

	// Update viewport content if messages changed
	m.viewport.SetContent(m.renderMessages()) // Assumes renderMessages exists
	m.viewport.GotoBottom()                   // Keep viewport at bottom

	return m, tea.Batch(cmds...)
}
