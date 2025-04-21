// filename: pkg/neurogo/tui/view.go
package tui

import "github.com/charmbracelet/lipgloss"

// View renders the TUI based on the current model state.
func (m model) View() string {
	if m.quitting {
		return "Quitting neurogo TUI...\n"
	}
	if !m.ready {
		return "Initializing TUI, waiting for window size..."
	}

	// --- Render Components ---
	viewportView := m.viewport.View()
	textareaView := m.textarea.View()
	// CORRECTED: Call with width, matches updated definition in update.go
	statusBarView := m.renderStatusBar(m.width)

	helpView := ""
	if m.helpVisible {
		// CORRECTED: Pass m.keyMap (which now implements help.KeyMap)
		helpView = m.help.View(m.keyMap)
	}

	// --- Assemble Layout ---
	mainContentView := lipgloss.JoinVertical(lipgloss.Left,
		viewportView,
		textareaView,
	)

	finalView := lipgloss.JoinVertical(lipgloss.Bottom,
		mainContentView,
		statusBarView,
		helpView,
	)

	return finalView
}
