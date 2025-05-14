// NeuroScript Version: 0.3.0
// File version: 0.1.0
// Refactor View method for Screen architecture.
// filename: pkg/neurogo/view.go
// nlines: 120 // Approximate
// risk_rating: MEDIUM
package neurogo

import (
	"fmt"
	"strings"

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

	// --- Calculate Dimensions ---
	statusBarHeight := 1 // From model or constants
	helpLines := 0
	if m.helpVisible {
		helpLines = strings.Count(m.help.View(m.keyMap), "\n") + 1
	}

	// Get current input area heights (they might dynamically resize based on content)
	// For fixed layout, we might define a max height or use LineHeight() * fixed_lines
	leftInputRenderHeight := lipgloss.Height(m.leftInputArea.View())
	rightInputRenderHeight := lipgloss.Height(m.rightInputArea.View())
	if leftInputRenderHeight < 1 {
		leftInputRenderHeight = 1
	} // Ensure minimum height
	if rightInputRenderHeight < 1 {
		rightInputRenderHeight = 1
	}

	// Calculate available height for screen content panes
	// inputPaneVerticalPadding := inputPaneBlurredStyle.GetVerticalFrameSize() // Assuming blurred for consistent height calc
	inputPaneVerticalPadding := m.leftInputArea.BlurredStyle.Base.GetVerticalFrameSize()

	availableScreenHeight := m.height - statusBarHeight - helpLines - leftInputRenderHeight - inputPaneVerticalPadding // Assuming both inputs take same effective height slot
	if availableScreenHeight < 1 {
		availableScreenHeight = 1
	}

	// Pane widths
	// Ensure these calculations are robust, especially if m.width is small
	leftPaneContainerWidth := m.width / 2
	rightPaneContainerWidth := m.width - leftPaneContainerWidth

	// Inner content width for screens (after accounting for container padding/border)
	leftScreenContentWidth := leftPaneContainerWidth - screenPaneContainerStyle.GetHorizontalFrameSize()
	rightScreenContentWidth := rightPaneContainerWidth - screenPaneContainerStyle.GetHorizontalFrameSize()

	// --- Render Active Screens ---
	var leftScreenView, rightScreenView string
	var leftScreenTitle, rightScreenTitle string

	activeLeftScreen := m.getActiveLeftScreen()
	if activeLeftScreen != nil {
		leftScreenView = activeLeftScreen.View(leftScreenContentWidth, availableScreenHeight)
		leftScreenTitle = activeLeftScreen.Name()
	} else {
		leftScreenView = lipgloss.Place(leftScreenContentWidth, availableScreenHeight, lipgloss.Center, lipgloss.Center, "(No active left screen)")
		leftScreenTitle = "Left Pane"
	}

	activeRightScreen := m.getActiveRightScreen()
	if activeRightScreen != nil {
		rightScreenView = activeRightScreen.View(rightScreenContentWidth, availableScreenHeight)
		rightScreenTitle = activeRightScreen.Name()
	} else {
		rightScreenView = lipgloss.Place(rightScreenContentWidth, availableScreenHeight, lipgloss.Center, lipgloss.Center, "(No active right screen)")
		rightScreenTitle = "Right Pane"
	}

	// --- Prepare Pane Titles ---
	leftPaneTitleStr := paneTitleStyle.Render(fmt.Sprintf("Left: %s", leftScreenTitle))
	rightPaneTitleStr := paneTitleStyle.Render(fmt.Sprintf("Right: %s", rightScreenTitle))

	// --- Render Screen Panes with Titles ---
	leftScreenPane := screenPaneContainerStyle.Width(leftPaneContainerWidth).Height(availableScreenHeight + lipgloss.Height(leftPaneTitleStr)).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			leftPaneTitleStr,
			leftScreenView,
		),
	)
	rightScreenPane := screenPaneContainerStyle.Width(rightPaneContainerWidth).Height(availableScreenHeight + lipgloss.Height(rightPaneTitleStr)).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			rightPaneTitleStr,
			rightScreenView,
		),
	)

	// --- Render Global Input Areas ---
	// Styles for input areas are now part of the textarea.Model (m.leftInputArea.FocusedStyle.Base, etc.)
	// The View() method of textarea already applies these styles.
	leftInputView := m.leftInputArea.View()
	rightInputView := m.rightInputArea.View()

	// --- Assemble Columns ---
	leftColumn := lipgloss.JoinVertical(lipgloss.Left,
		leftScreenPane,
		leftInputView,
	)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left,
		rightScreenPane,
		rightInputView,
	)

	// --- Assemble Main Application Area ---
	mainApplicationArea := lipgloss.JoinHorizontal(lipgloss.Top,
		leftColumn,
		rightColumn,
	)

	// --- Render Status Bar and Help ---
	statusBarView := m.renderStatusBar(m.width) // Assuming renderStatusBar is now a method on *model
	helpViewContent := ""
	if m.helpVisible {
		helpViewContent = m.help.View(m.keyMap)
	}

	// --- Final Vertical Join ---
	finalRender := lipgloss.JoinVertical(lipgloss.Top, // Use Top alignment
		mainApplicationArea,
		statusBarView,
		helpViewContent,
	)

	return finalRender
}

// renderStatusBar (moved from update_helpers.go or model.go, now a method if it uses model fields)
// For simplicity, assuming it's defined elsewhere or we inline parts of it.
// This is a simplified version. The actual one uses m.spinner, m.currentActivity etc.
func (m *model) renderStatusBar(width int) string {
	activity := m.currentActivity
	if m.isWaitingForAI || m.isSyncing || m.initialScriptRunning {
		activity = m.spinner.View() + " " + activity
	}
	if m.lastError != nil {
		activity = errorStyle.Render(fmt.Sprintf("ERR: %s", m.lastError.Error()))
	} else if m.patchStatus != "" {
		activity = patchStatusStyle.Render(m.patchStatus)
	}

	// Simplified status bar rendering
	statusText := fmt.Sprintf(" %s | Focus: %s ", activity, m.focusTarget.String()) // Assuming FocusTarget has a String() method

	// Ensure status bar doesn't overflow
	if lipgloss.Width(statusText) > width {
		statusText = statusText[:width-3] + "..."
	}
	return statusBarSyle.Width(width).Render(statusText)
}

// String method for FocusTarget enum to be used in status bar
func (ft FocusTarget) String() string {
	switch ft {
	case FocusLeftInput:
		return "Left Input"
	case FocusRightInput:
		return "Right Input"
	case FocusLeftPane:
		return "Left Pane"
	case FocusRightPane:
		return "Right Pane"
	default:
		return "Unknown Focus"
	}
}
