// filename: pkg/neurogo/tui/update_helpers.go
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
// It assumes m.viewport is ready and sized.
func (m *model) addMessage(sender, text string) {
	// Prevent excessive message buildup (optional)
	if len(m.messages) > 1000 {
		m.messages = m.messages[len(m.messages)-500:] // Keep last 500
	}
	m.messages = append(m.messages, message{sender: sender, text: text})
	if m.ready {
		// Ensure content is set *before* scrolling
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
	}
}

// renderMessages formats the message history for the viewport.
func (m *model) renderMessages() string {
	var content strings.Builder
	maxMessages := 200 // Limit history displayed
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
			style = systemStyle.Copy().Italic(true).Foreground(lipgloss.Color("13")) // Magenta Debug
		case "Command":
			style = systemStyle.Copy().Foreground(lipgloss.Color("14")) // Cyan command log
		// Add other sender types as needed
		default:
			style = systemStyle
		}
		// Ensure message text doesn't contain excessive newlines if not intended
		processedText := strings.ReplaceAll(msg.text, "\r\n", "\n") // Normalize newlines
		// Basic formatting: Sender: Text
		content.WriteString(style.Render(fmt.Sprintf("%s: %s\n", msg.sender, processedText)))
	}
	// Ensure viewport gets at least one line if messages are empty to prevent layout collapse
	finalContent := content.String()
	if strings.TrimSpace(finalContent) == "" {
		return " " // Return a single space to ensure viewport doesn't collapse
	}
	return finalContent
}

// renderStatusBar formats the status bar content.
func (m *model) renderStatusBar(width int) string {
	if !m.ready {
		return ""
	}

	// Basic info
	modelInfo := fmt.Sprintf("AI: %s", m.aiModelName)
	// Simplified file info for now, update based on actual state later if needed
	fileInfo := fmt.Sprintf("Files(L:%d/R:%d)", m.localFileCount, m.apiFileCount)
	syncInfo := fmt.Sprintf("Sync(Up:%d/Del:%d)", m.syncUploads, m.syncDeletes) // Use updated stats fields
	left := strings.Join([]string{modelInfo, fileInfo, syncInfo}, " | ")

	// Activity / Error indicator
	activity := ""
	if m.isWaitingForAI {
		activity = fmt.Sprintf("%s Waiting for AI...", m.spinner.View())
	} else if m.isSyncing { // Check sync status
		activity = fmt.Sprintf("%s %s", m.spinner.View(), m.currentActivity) // Use currentActivity
	} else if m.patchStatus != "" { // Keep patch status if needed
		activity = fmt.Sprintf("%s %s", m.spinner.View(), m.patchStatus)
	} else if m.currentActivity != "" { // Display other activities if not syncing/waiting for AI
		activity = fmt.Sprintf("%s %s", m.spinner.View(), m.currentActivity)
	}

	errorMsg := ""
	if m.lastError != nil {
		errorMsg = errorStyle.Render(fmt.Sprintf("Error: %v", m.lastError))
	}

	right := ""
	if errorMsg != "" {
		right = errorMsg // Error takes precedence
	} else {
		right = activity
	}

	// Calculate available space for the separator, ensuring it's not negative
	separatorWidth := width - lipgloss.Width(left) - lipgloss.Width(right) - statusBarSyle.GetHorizontalPadding()*2 // Account for padding
	if separatorWidth < 0 {
		separatorWidth = 0
	}
	separator := strings.Repeat(" ", separatorWidth)

	finalStatus := lipgloss.JoinHorizontal(lipgloss.Top, left, separator, right)

	// Use PlaceHorizontal to ensure the final string fits the width, padding if necessary
	return statusBarSyle.Render(lipgloss.PlaceHorizontal(width, lipgloss.Left, finalStatus, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceForeground(statusBarSyle.GetBackground())))
}

// Ensure errMsg type is defined (e.g., in msgs.go)
// Ensure syncCompleteMsg type is defined (e.g., in msgs.go)

// --- Command Function for Async Sync (Now a method on model) ---

// runSyncCmd performs the sync operation in the background and returns a message.
// It accesses necessary app configuration and clients via the AppAccess interface
// stored in the model (m.app).
func (m *model) runSyncCmd() tea.Cmd {
	return func() tea.Msg {
		// Use a local logger for this specific operation, obtained via the interface
		logger := m.app.GetLogger() // Get logger instance
		if logger == nil {
			// Fallback if logger retrieval fails, though App should prevent this
			return errMsg{fmt.Errorf("TUI:runSyncCmd - Logger not available")}
		}

		// Access config and clients via interface methods on m.app
		if m.app == nil {
			logger.Error("Sync command failed: app reference (via interface) not available in TUI model")
			return errMsg{fmt.Errorf("app reference (via interface) not available in TUI model")}
		}

		syncDir := m.app.GetSyncDir()
		if syncDir == "" {
			logger.Error("Sync command failed: Sync directory not configured.")
			return errMsg{fmt.Errorf("sync directory not configured")}
		}

		// --- Get Interpreter (Replaces LLM Client Check) ---
		interp := m.app.GetInterpreter() // Use interface getter
		if interp == nil {
			logger.Error("Sync command failed: Interpreter not available.")
			return errMsg{fmt.Errorf("interpreter not available for sync operation")}
		}
		// --- End Interpreter Check ---

		// Validate Sync Directory securely relative to current working directory or sandbox
		// The interpreter's FileAPI already knows the sandbox root.
		// We still need the configured syncDir relative path.
		fileAPI := interp.FileAPI()
		if fileAPI == nil {
			logger.Error("Sync command failed: Interpreter's FileAPI is nil.")
			return errMsg{fmt.Errorf("interpreter FileAPI is nil, cannot resolve sync path")}
		}

		// Resolve the sync path securely using FileAPI
		absSyncDir, secErr := fileAPI.ResolvePath(syncDir)
		if secErr != nil {
			// Error message from ResolvePath already contains the sandbox root [cite: 2]
			// *** FIX: Removed ", "base", fileAPI.SandboxRoot()" ***
			logger.Error("Sync command failed: Invalid sync directory path.", "input_path", syncDir, "error", secErr)
			return errMsg{fmt.Errorf("invalid sync directory path '%s': %w", syncDir, secErr)}
		}

		// Stat check
		dirInfo, statErr := os.Stat(absSyncDir)
		if statErr != nil {
			errMsgFmt := "failed to stat sync directory %s: %w"
			if os.IsNotExist(statErr) {
				errMsgFmt = "sync directory does not exist: %s: %w"
			}
			logger.Error(errMsgFmt, absSyncDir, statErr)

			// Return a user-friendly error message
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

		// Use interface getters for config needed by the helper
		syncFilter := m.app.GetSyncFilter()
		ignoreGitignore := m.app.GetSyncIgnoreGitignore()

		// Call the core sync helper, passing the interpreter
		// The helper will use the interpreter to get the logger and LLM client if needed
		stats, syncErr := core.SyncDirectoryUpHelper(
			ctx,
			absSyncDir, // Resolved absolute path
			syncFilter,
			ignoreGitignore,
			interp, // Pass the Interpreter
			// Removed logger argument
		)

		// Return the result message
		return syncCompleteMsg{stats: stats, err: syncErr}
	}
}

// setSizes helper - recalculates component sizes.
// Assumes renderStatusBar, renderMessages, m.help.View exist elsewhere (e.g., view.go or update_helpers.go)
func (m *model) setSizes(width, height int) {
	const commandInputWidth = 25
	const inputSeparatorWidth = 1

	m.help.Width = width

	// Ensure inputs have minimum height (e.g., 1)
	cmdHeight := max(1, m.commandInput.Height())
	promptHeight := max(1, m.promptInput.Height())
	// Take the max height for the input row layout
	inputsRowHeight := max(cmdHeight, promptHeight)

	statusBarHeight := lipgloss.Height(m.renderStatusBar(width)) // Assumes renderStatusBar exists
	helpHeight := 0
	if m.helpVisible { // Use m.helpVisible state flag
		helpHeight = lipgloss.Height(m.help.View(m.keyMap))
	}

	const verticalMargin = 1 // Margin between viewport and inputs/status/help
	viewportHeight := height - inputsRowHeight - statusBarHeight - helpHeight - verticalMargin
	if viewportHeight < 1 { // Ensure viewport has at least height 1
		viewportHeight = 1
	}

	m.commandInput.SetWidth(commandInputWidth)
	// Calculate prompt width dynamically
	promptInputWidth := width - commandInputWidth - inputSeparatorWidth
	if promptInputWidth < 10 { // Ensure minimum prompt width
		promptInputWidth = 10
	}
	m.promptInput.SetWidth(promptInputWidth)

	m.viewport.Width = width
	m.viewport.Height = viewportHeight

	// Update content only if ready, otherwise viewport might not be fully initialized
	if m.ready {
		m.viewport.SetContent(m.renderMessages()) // Assumes renderMessages exists
	}
}

// Local max helper
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// NOTE: Ensure errMsg, syncCompleteMsg are defined in msgs.go
// NOTE: Ensure addMessage, renderMessages, renderStatusBar are defined elsewhere (e.g., view.go or update_helpers.go)
