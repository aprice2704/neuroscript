// filename: pkg/neurogo/tui/view.go
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI based on the current model state.
func (m model) View() string {
	if m.quitting {
		return "Quitting neurogo TUI...\n"
	}
	if !m.ready {
		return "Initializing TUI, waiting for window size..."
	}

	// --- Render Components ---
	// Render viewport (style set in model constructor should apply)
	viewportView := m.viewport.View()

	statusBarView := m.renderStatusBar(m.width) // Status bar content

	// Render command and prompt inputs (styles applied internally now)
	commandView := m.commandInput.View()
	promptView := m.promptInput.View()

	// --- Reverted to Horizontal Join ---
	inputsView := lipgloss.JoinHorizontal(lipgloss.Top,
		commandView,
		promptView,
	)
	// ---

	helpView := ""
	if m.helpVisible {
		helpView = m.help.View(m.keyMap)
	}

	// --- Assemble Layout ---
	// Main content view (viewport above inputs)
	mainContentView := lipgloss.JoinVertical(lipgloss.Left,
		viewportView,
		inputsView,
	)

	// Combine everything
	finalView := lipgloss.JoinVertical(lipgloss.Bottom,
		mainContentView,
		statusBarView,
		helpView,
	)

	return finalView
}
