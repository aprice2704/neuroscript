// NeuroScript Version: 0.3.0
// File version: 0.0.2 // Layout main and emit viewports side-by-side.
// filename: pkg/neurogo/tui/view.go
// nlines: 60 // Approximate
// risk_rating: LOW
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI based on the current model state.
func (m model) View() string {
	if m.quitting {
		// Centered "Quitting..." message
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Quitting neurogo TUI...")
	}
	if !m.ready {
		// Centered "Initializing..." message
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Initializing TUI, please wait...")
	}

	// --- Render Individual Components ---
	// These .View() calls will use the styles (including distinct borders)
	// set on them in model.go during newModel.
	mainConversationView := m.viewport.View()
	emitLogView := m.emitLogViewport.View()

	statusBarView := m.renderStatusBar(m.width) // Assumes m.renderStatusBar exists

	commandInputView := m.commandInput.View()
	promptInputView := m.promptInput.View()

	// --- Assemble Layout ---

	// 1. Side-by-side viewports for main conversation and emit log
	// Their widths (m.viewport.Width, m.emitLogViewport.Width) and shared height
	// are determined by the setSizes function (in update_helpers.go).
	topViewportsArea := lipgloss.JoinHorizontal(lipgloss.Top,
		mainConversationView,
		emitLogView,
	)

	// 2. Input areas (command and prompt) joined horizontally.
	// The setSizes function determines their individual widths.
	// The 1-character offset issue:
	// - If commandInput's style has right padding/margin, or promptInput's style has left padding/margin,
	//   that will create space.
	// - If both have full borders, their adjacent borders will create a visual separation.
	//   The current styles in model.go (e.g., focusedCommandStyle, focusedPromptStyle)
	//   use RoundedBorder or NormalBorder which draw all sides.
	// The `inputSeparatorWidth` in `setSizes` was an attempt to manage space *between* them.
	// If the borders themselves provide enough visual separation, `inputSeparatorWidth` could be 0.
	// Or, if an explicit space is desired, it could be:
	//   `lipgloss.JoinHorizontal(lipgloss.Bottom, commandInputView, " ", promptInputView)`
	//   and `setSizes` would need to account for that 1 character.
	// For now, assuming direct join and borders provide separation.
	inputControlsArea := lipgloss.JoinHorizontal(lipgloss.Bottom,
		commandInputView,
		promptInputView,
	)

	// 3. Help view (conditionally displayed at the bottom)
	helpViewContent := ""
	if m.helpVisible {
		helpViewContent = m.help.View(m.keyMap) // keyMap from model
	}

	// 4. Main content area: top viewports above the input area
	mainApplicationArea := lipgloss.JoinVertical(lipgloss.Left,
		topViewportsArea,
		inputControlsArea,
	)

	// 5. Final assembly: main application area, then status bar, then help view
	// This stacks them vertically.
	finalRender := lipgloss.JoinVertical(lipgloss.Left,
		mainApplicationArea,
		statusBarView,
		helpViewContent, // This will be empty if help is not visible
	)

	return finalRender
}

// Note: m.renderStatusBar() is assumed to be defined in update_helpers.go or view_helpers.go
// and should correctly use m.width for its own layout.
// The m.renderMessages() is used by setSizes/update loop to populate m.viewport.SetContent().
