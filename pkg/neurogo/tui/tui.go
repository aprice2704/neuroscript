// filename: pkg/neurogo/tui/tui.go
package tui

import (
	"fmt"
	// Import log for nil check fallback
	// No neurogo import needed here
	tea "github.com/charmbracelet/bubbletea"
)

// Start initializes and runs the Bubble Tea TUI.
// It takes an implementation of the AppAccess interface to interact
// with the main application logic without causing import cycles.
func Start(app AppAccess) error {
	if app == nil {
		return fmt.Errorf("cannot start TUI with a nil application reference")
	}

	// Use the interface method to get the logger
	logger := app.GetLogger()

	// Pass the app interface to the model constructor.
	model := newModel(app) // Pass the interface

	p := tea.NewProgram(model, tea.WithAltScreen())

	logger.Debug("Starting Bubble Tea program...")
	_, err := p.Run() // This blocks until the TUI quits

	if err != nil {
		// Log the error using the app's logger if possible
		logger.Debug("Bubble Tea program run failed: %v", err)
		return fmt.Errorf("error running TUI: %w", err)
	}
	logger.Debug("Bubble Tea program finished.")
	return nil // Return nil on clean exit
}
