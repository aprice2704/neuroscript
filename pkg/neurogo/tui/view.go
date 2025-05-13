// NeuroScript Version: 0.3.0
// File version: 0.0.9
// Dynamic title for Local Output pane based on display mode.
// filename: pkg/neurogo/tui/view.go
// nlines: 85
// risk_rating: MEDIUM
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI based on the current model state.
func (m model) View() string {
	if m.quitting {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Quitting neurogo TUI...")
	}
	if !m.ready {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Initializing TUI, please wait...")
	}

	var localOutputContainerStyle, aiOutputContainerStyle lipgloss.Style

	if m.focusIndex == focusLocalOutput {
		localOutputContainerStyle = localOutputFocusedStyle
	} else {
		localOutputContainerStyle = localOutputBlurredStyle
	}
	if m.focusIndex == focusAIOutput {
		aiOutputContainerStyle = aiOutputFocusedStyle
	} else {
		aiOutputContainerStyle = aiOutputBlurredStyle
	}

	localInputView := m.localInput.View()
	aiInputView := m.aiInput.View()

	// --- Render Local Output Pane (Area A) ---
	var loTitle string
	switch m.localOutputDisplayMode {
	case localOutputModeScript:
		loTitle = "Script Output"
	case localOutputModeWMStatus:
		loTitle = "Worker Manager Status"
	default:
		loTitle = "Local Output" // Fallback
	}
	loTitleString := paneTitleStyle.Render(loTitle)
	localOutputContentView := m.localOutput.View()

	localOutputPane := localOutputContainerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			loTitleString,
			localOutputContentView,
		),
	)

	// --- Render AI Output Pane (Area B) ---
	aiTitleString := paneTitleStyle.Render("AI Output")
	aiViewContent := m.aiOutput.View()
	aiOutputPane := aiOutputContainerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			aiTitleString,
			aiViewContent,
		),
	)

	// --- Assemble Layout ---
	leftColumn := lipgloss.JoinVertical(lipgloss.Left,
		localOutputPane,
		localInputView,
	)

	rightColumn := lipgloss.JoinVertical(lipgloss.Left,
		aiOutputPane,
		aiInputView,
	)

	mainApplicationArea := lipgloss.JoinHorizontal(lipgloss.Top,
		leftColumn,
		rightColumn,
	)

	statusBarView := m.renderStatusBar(m.width)
	helpViewContent := ""
	if m.helpVisible {
		helpViewContent = m.help.View(m.keyMap)
	}

	finalRender := lipgloss.JoinVertical(lipgloss.Left,
		mainApplicationArea,
		statusBarView,
		helpViewContent,
	)

	return finalRender
}
