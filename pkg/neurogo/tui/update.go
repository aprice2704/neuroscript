// filename: pkg/neurogo/tui/update.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Init runs initialization commands for the TUI model.
// CORRECTED: This is the single definition of Init.
func (m model) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages received by the TUI model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// Process updates for components first
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			m.quitting = true
			m.messages = append(m.messages, message{sender: "System", text: "Quitting..."})
			if m.ready { // Update viewport only if ready
				m.viewport.SetContent(m.renderMessages())
			}
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Help):
			// Use the help component's toggle method
			m.help.ShowAll = !m.help.ShowAll
			m.helpVisible = m.help.ShowAll // Keep state consistent
			// Recalculate sizes because help height changes
			if m.ready {
				m.setSizes(m.width, m.height)
				m.viewport.GotoBottom()
			}
			return m, nil

			// TODO: Handle Enter/Submit Key
			// ...

		default:
			if m.ready && m.textarea.Focused() {
				m.viewport.GotoBottom()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			// Calculate initial sizes *before* setting content
			m.setSizes(msg.Width, m.height)
			m.viewport.SetContent(m.renderMessages()) // Render initial messages
			m.ready = true
		} else {
			m.setSizes(msg.Width, m.height) // Recalculate sizes
		}
		m.viewport.GotoBottom()

	case spinner.TickMsg:
		var spinCmd tea.Cmd
		if m.isWaitingForAI || m.activeToolMessage != "" || m.patchStatus != "" {
			m.spinner, spinCmd = m.spinner.Update(msg)
			cmds = append(cmds, spinCmd)
		}

	case errMsg: // Handles errors wrapped in our custom message
		m.lastError = msg.err
		m.isWaitingForAI = false
		m.activeToolMessage = ""
		m.patchStatus = ""
		m.addMessage("System", errorStyle.Render(fmt.Sprintf("Error: %v", msg.err)))
		m.viewport.GotoBottom() // Ensure error is visible

		// --- Placeholder Handlers ---
		// case llmResponseMsg: ...
		// case toolResultMsg: ...
		// case syncCompleteMsg: ...
		// case statusUpdateMsg: ...

	} // End main switch

	return m, tea.Batch(cmds...)
}

// setSizes helper - recalculates component sizes.
func (m *model) setSizes(width, height int) {
	m.help.Width = width // Help uses full width

	textAreaHeight := m.textarea.Height()
	// CORRECTED: Pass width to renderStatusBar for accurate height calculation
	statusBarHeight := lipgloss.Height(m.renderStatusBar(width))
	helpHeight := 0
	// Use the component's state directly
	if m.help.ShowAll {
		helpHeight = lipgloss.Height(m.help.View(m.keyMap))
	}

	const verticalMargin = 1 // Optional gap
	viewportHeight := height - textAreaHeight - statusBarHeight - helpHeight - verticalMargin
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	// Apply sizes
	m.textarea.SetWidth(width)
	m.viewport.Width = width
	m.viewport.Height = viewportHeight

	// Content might need re-rendering if width changed, affecting wrapping
	if m.ready {
		m.viewport.SetContent(m.renderMessages())
	}
}

// addMessage helper - appends message and updates viewport.
func (m *model) addMessage(sender, text string) {
	m.messages = append(m.messages, message{sender: sender, text: text})
	if m.ready {
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
	}
}

// renderMessages helper - formats history for viewport.
func (m *model) renderMessages() string {
	var sb strings.Builder
	for i, msg := range m.messages {
		var style lipgloss.Style
		prefix := ""
		switch msg.sender {
		case "You":
			style, prefix = userStyle, "You: "
		case "AI":
			style, prefix = aiStyle, "AI: "
		case "AIToolCall":
			style, prefix = aiToolCallStyle, "[AI Tool Call]: "
		case "SysToolCall":
			style, prefix = sysToolCallStyle, "[Tool Call]: "
		case "ToolResult":
			style, prefix = toolResultStyle, "[Tool Result]: "
		case "Patch":
			style, prefix = patchStatusStyle, "[Patch]: "
		case "System":
			style, prefix = systemStyle, "[System]: "
		default:
			style, prefix = inactiveStyle, "["+msg.sender+"]: "
		}
		renderedMsg := style.Render(prefix + msg.text)
		sb.WriteString(renderedMsg)
		if i < len(m.messages)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// renderStatusBar helper - creates status bar string.
// CORRECTED: Added width parameter to definition
func (m *model) renderStatusBar(width int) string {
	if !m.ready {
		return ""
	} // Don't render if not ready

	modelInfo := fmt.Sprintf("Model: %s", m.aiModelName)
	fileInfo := fmt.Sprintf("Files(L:%d/R:%d)", m.localFileCount, m.apiFileCount)
	syncInfo := ""
	if m.syncUploads > 0 || m.syncDeletes > 0 {
		syncInfo = fmt.Sprintf("Sync(↑%d/↓%d)", m.syncUploads, m.syncDeletes)
	}
	leftParts := []string{modelInfo, fileInfo}
	if syncInfo != "" {
		leftParts = append(leftParts, syncInfo)
	}
	left := strings.Join(leftParts, " | ")

	activity := ""
	if m.isWaitingForAI {
		activity = m.spinner.View() + " Waiting for AI..."
	} else if m.activeToolMessage != "" {
		activity = m.spinner.View() + " " + m.activeToolMessage
	} else if m.patchStatus != "" {
		activity = m.spinner.View() + " " + m.patchStatus
	}

	errorInfo := ""
	if m.lastError != nil {
		errStr := "ERR: " + m.lastError.Error()
		maxErrWidth := width / 3
		if len(errStr) > maxErrWidth {
			errStr = errStr[:maxErrWidth-3] + "..."
		}
		errorInfo = errorStyle.Render(errStr)
	}

	right := activity
	if activity != "" && errorInfo != "" {
		right += " | " + errorInfo
	} else {
		right += errorInfo
	}

	// Calculate gap width *after* potentially combining activity and error
	padding := statusBarSyle.GetHorizontalPadding()
	gapWidth := width - lipgloss.Width(left) - lipgloss.Width(right) - padding
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := strings.Repeat(" ", gapWidth)

	finalStatus := lipgloss.JoinHorizontal(lipgloss.Top, left, gap, right)
	// Apply style *after* joining
	return statusBarSyle.Width(width).Render(finalStatus)
}
