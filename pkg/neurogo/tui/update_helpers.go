// NeuroScript Version: 0.3.0
// File version: 0.0.1 // Adjust setSizes for side-by-side main and emit viewports.
// filename: pkg/neurogo/tui/update_helpers.go
// nlines: 200 // Approximate
// risk_rating: MEDIUM
package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core" // Keep for direct logging if needed
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// addMessage is a helper to append a message to the model's list and scroll.
func (m *model) addMessage(sender, text string) {
	if len(m.messages) > 1000 {
		m.messages = m.messages[len(m.messages)-500:]
	}
	m.messages = append(m.messages, message{sender: sender, text: text})
	if m.ready {
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
	}
}

// renderMessages formats the message history for the viewport.
func (m *model) renderMessages() string {
	var content strings.Builder
	maxMessages := 200
	start := 0
	if len(m.messages) > maxMessages {
		start = len(m.messages) - maxMessages
	}
	visibleMessages := m.messages[start:]

	for _, msg := range visibleMessages {
		var style lipgloss.Style
		switch msg.sender {
		case "You":
			style = userStyle
		case "AI":
			style = aiStyle
		case "System":
			style = systemStyle
		case "Debug":
			style = systemStyle.Copy().Italic(true).Foreground(lipgloss.Color("13"))
		case "Command":
			style = systemStyle.Copy().Foreground(lipgloss.Color("14"))
		default:
			style = systemStyle
		}
		processedText := strings.ReplaceAll(msg.text, "\r\n", "\n")
		content.WriteString(style.Render(fmt.Sprintf("%s: %s\n", msg.sender, processedText)))
	}
	finalContent := content.String()
	if strings.TrimSpace(finalContent) == "" {
		return " "
	}
	return finalContent
}

// renderStatusBar formats the status bar content.
func (m *model) renderStatusBar(width int) string {
	if !m.ready {
		return ""
	}
	modelInfo := fmt.Sprintf("AI: %s", m.aiModelName)
	fileInfo := fmt.Sprintf("Files(L:%d/R:%d)", m.localFileCount, m.apiFileCount)
	syncInfo := fmt.Sprintf("Sync(Up:%d/Del:%d)", m.syncUploads, m.syncDeletes)
	left := strings.Join([]string{modelInfo, fileInfo, syncInfo}, " | ")

	activity := ""
	if m.isWaitingForAI {
		activity = fmt.Sprintf("%s Waiting for AI...", m.spinner.View())
	} else if m.isSyncing {
		activity = fmt.Sprintf("%s %s", m.spinner.View(), m.currentActivity)
	} else if m.patchStatus != "" {
		activity = fmt.Sprintf("%s %s", m.spinner.View(), m.patchStatus)
	} else if m.currentActivity != "" {
		activity = fmt.Sprintf("%s %s", m.spinner.View(), m.currentActivity)
	}

	errorMsg := ""
	if m.lastError != nil {
		errorMsg = errorStyle.Render(fmt.Sprintf("Error: %v", m.lastError))
	}

	right := ""
	if errorMsg != "" {
		right = errorMsg
	} else {
		right = activity
	}

	separatorWidth := width - lipgloss.Width(left) - lipgloss.Width(right) - statusBarSyle.GetHorizontalPadding()*2
	if separatorWidth < 0 {
		separatorWidth = 0
	}
	separator := strings.Repeat(" ", separatorWidth)
	finalStatus := lipgloss.JoinHorizontal(lipgloss.Top, left, separator, right)
	return statusBarSyle.Render(lipgloss.PlaceHorizontal(width, lipgloss.Left, finalStatus, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceForeground(statusBarSyle.GetBackground())))
}

// runSyncCmd performs the sync operation in the background and returns a message.
func (m *model) runSyncCmd() tea.Cmd {
	return func() tea.Msg {
		logger := m.app.GetLogger()
		if logger == nil {
			return errMsg{fmt.Errorf("TUI:runSyncCmd - Logger not available")}
		}
		if m.app == nil {
			logger.Error("Sync command failed: app reference (via interface) not available in TUI model")
			return errMsg{fmt.Errorf("app reference (via interface) not available in TUI model")}
		}
		syncDir := m.app.GetSyncDir()
		if syncDir == "" {
			logger.Error("Sync command failed: Sync directory not configured.")
			return errMsg{fmt.Errorf("sync directory not configured")}
		}
		interp := m.app.GetInterpreter()
		if interp == nil {
			logger.Error("Sync command failed: Interpreter not available.")
			return errMsg{fmt.Errorf("interpreter not available for sync operation")}
		}
		fileAPI := interp.FileAPI()
		if fileAPI == nil {
			logger.Error("Sync command failed: Interpreter's FileAPI is nil.")
			return errMsg{fmt.Errorf("interpreter FileAPI is nil, cannot resolve sync path")}
		}
		absSyncDir, secErr := fileAPI.ResolvePath(syncDir)
		if secErr != nil {
			logger.Error("Sync command failed: Invalid sync directory path.", "input_path", syncDir, "error", secErr)
			return errMsg{fmt.Errorf("invalid sync directory path '%s': %w", syncDir, secErr)}
		}
		dirInfo, statErr := os.Stat(absSyncDir)
		if statErr != nil {
			errMsgFmt := "failed to stat sync directory %s: %w"
			if os.IsNotExist(statErr) {
				errMsgFmt = "sync directory does not exist: %s: %w"
			}
			logger.Error(errMsgFmt, absSyncDir, statErr)
			if os.IsNotExist(statErr) {
				return errMsg{fmt.Errorf("sync directory does not exist: %s", absSyncDir)}
			}
			return errMsg{fmt.Errorf("cannot access sync directory %s", absSyncDir)}
		}
		if !dirInfo.IsDir() {
			logger.Error("Sync command failed: Sync path is not a directory.", "path", absSyncDir)
			return errMsg{fmt.Errorf("sync path is not a directory: %s", absSyncDir)}
		}
		ctx := context.Background()
		syncFilter := m.app.GetSyncFilter()
		ignoreGitignore := m.app.GetSyncIgnoreGitignore()
		stats, syncErr := core.SyncDirectoryUpHelper(ctx, absSyncDir, syncFilter, ignoreGitignore, interp)
		return syncCompleteMsg{stats: stats, err: syncErr}
	}
}

// setSizes helper - recalculates component sizes.
func (m *model) setSizes(width, height int) {
	// Constants for layout proportions and minimums
	const commandInputFixedWidth = 25 // Fixed width for command input
	const inputSeparatorWidth = 1     // Separator between command and prompt inputs
	const mainViewportRatio = 0.6     // Main viewport gets 60% of the shared width
	const minInputHeight = 1
	const minViewportHeight = 3 // Minimum height for viewports
	const minPromptWidth = 20   // Minimum width for the prompt input area

	m.help.Width = width

	// --- Calculate Heights ---
	// Heights for fixed elements
	cmdHeight := max(minInputHeight, m.commandInput.Height())   // Use minInputHeight
	promptHeight := max(minInputHeight, m.promptInput.Height()) // Use minInputHeight
	inputsRowHeight := max(cmdHeight, promptHeight)             // The row of inputs takes the max of their heights

	statusBarHeight := lipgloss.Height(m.renderStatusBar(width))
	helpHeight := 0
	if m.helpVisible {
		helpHeight = lipgloss.Height(m.help.View(m.keyMap))
	}

	// Available height for the main content area (main viewport + emit log viewport)
	// Subtract one more for a potential margin or to prevent being too cramped
	const verticalMarginBetweenViewportsAndInputs = 1
	availableHeightForViewports := height - inputsRowHeight - statusBarHeight - helpHeight - verticalMarginBetweenViewportsAndInputs
	if availableHeightForViewports < minViewportHeight {
		availableHeightForViewports = minViewportHeight
	}

	// --- Set Heights ---
	m.viewport.Height = availableHeightForViewports
	m.emitLogViewport.Height = availableHeightForViewports // Both viewports share the same height

	// --- Calculate Widths ---
	// Inputs
	m.commandInput.SetWidth(commandInputFixedWidth)
	promptInputWidth := width - commandInputFixedWidth - inputSeparatorWidth
	if promptInputWidth < minPromptWidth {
		promptInputWidth = minPromptWidth
	}
	m.promptInput.SetWidth(promptInputWidth)

	// Viewports (Main and Emit Log, side-by-side)
	// They share the total width.
	mainVPWidth := int(float64(width) * mainViewportRatio)
	emitLogVPWidth := width - mainVPWidth

	// Ensure minimum widths for viewports if they are too small
	if mainVPWidth < 10 { // Arbitrary minimum
		mainVPWidth = 10
		emitLogVPWidth = max(10, width-mainVPWidth) // Adjust emit log if main was clamped
	}
	if emitLogVPWidth < 10 { // Arbitrary minimum
		emitLogVPWidth = 10
		mainVPWidth = max(10, width-emitLogVPWidth) // Adjust main if emit log was clamped
	}
	// Final check to prevent overlap if both clamped due to very small total width
	if mainVPWidth+emitLogVPWidth > width {
		// Prioritize main viewport in extreme cases
		mainVPWidth = max(10, width-10)
		emitLogVPWidth = width - mainVPWidth
	}

	m.viewport.Width = mainVPWidth
	m.emitLogViewport.Width = emitLogVPWidth

	// Update content now that sizes are set
	if m.ready {
		m.viewport.SetContent(m.renderMessages())
		m.emitLogViewport.SetContent(strings.Join(m.emittedLines, "\n")) // Update emit log content
		m.viewport.GotoBottom()
		m.emitLogViewport.GotoBottom()
	}
}

// Local max helper
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
