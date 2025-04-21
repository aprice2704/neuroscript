// filename: pkg/neurogo/tui/tui.go
package tui

import (
	"fmt"
	// "log" // Using app's logger is preferred

	"github.com/aprice2704/neuroscript/pkg/neurogo" // To access App struct
	tea "github.com/charmbracelet/bubbletea"
)

// Start initializes and runs the Bubble Tea TUI.
// It takes the initialized neurogo App as input to access config and clients.
func Start(app *neurogo.App) error {
	if app == nil {
		return fmt.Errorf("cannot start TUI with a nil application reference")
	}
	if app.DebugLog == nil {
		// Fallback if logger isn't set, though app initialization should ensure this
		// Consider using standard log if app.DebugLog must be nil sometimes
		return fmt.Errorf("application debug logger is not initialized")
	}

	// Pass the app instance (or specific parts like config, llmClient)
	// to the model constructor. newModel must be defined in model.go
	model := newModel(app)

	// Start Bubble Tea. Use WithAltScreen() for full-screen.
	// Use WithMouseCellMotion() if mouse interaction might be needed later.
	// Use WithContext to allow graceful shutdown on context cancellation if needed later.
	p := tea.NewProgram(model, tea.WithAltScreen()) //, tea.WithMouseCellMotion()) // Pass app context? tea.WithContext(app.Ctx)

	// Log TUI start/stop for debugging using the app's logger
	app.DebugLog.Println("Starting Bubble Tea program...")
	_, err := p.Run() // This blocks until the TUI quits
	// Log the error *before* returning it
	if err != nil {
		// Use ErrorLog for actual errors encountered during runtime
		if app.ErrorLog != nil {
			app.ErrorLog.Printf("Bubble Tea program run failed: %v", err)
		}
		return fmt.Errorf("error running TUI: %w", err)
	}
	app.DebugLog.Println("Bubble Tea program finished.")
	return nil // Return nil on clean exit
}

// Helper function to exit cleanly (can be called via a tea.Cmd if needed later)
// func quitCmd() tea.Cmd {
// 	return tea.Quit
// }
