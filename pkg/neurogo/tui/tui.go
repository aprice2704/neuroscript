// filename: pkg/neurogo/tui/tui.go
package tui

import (
	"fmt"
	"log" // Import log for nil check fallback

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
	debugLogger := app.GetLogger()
	if debugLogger == nil {
		// Fallback if logger isn't available via interface (shouldn't happen ideally)
		log.Println("WARN: TUI Start called with app returning nil DebugLogger.")
		debugLogger = log.New(log.Writer(), "DEBUG-TUI-FALLBACK: ", log.LstdFlags|log.Lshortfile)
	}

	// Pass the app interface to the model constructor.
	model := newModel(app) // Pass the interface

	p := tea.NewProgram(model, tea.WithAltScreen())

	debugLogger.Println("Starting Bubble Tea program...")
	_, err := p.Run() // This blocks until the TUI quits

	if err != nil {
		// Log the error using the app's logger if possible
		debugLogger.Debug("Bubble Tea program run failed: %v", err)
		return fmt.Errorf("error running TUI: %w", err)
	}
	debugLogger.Println("Bubble Tea program finished.")
	return nil // Return nil on clean exit
}
