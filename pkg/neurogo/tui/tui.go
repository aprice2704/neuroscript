// NeuroScript Version: 0.3.0
// File version: 0.0.2 // Accept initialScriptPath, set up TUIEmitWriter, restore stdout.
// filename: pkg/neurogo/tui/tui.go
// nlines: 55 // Approximate
// risk_rating: MEDIUM
package tui

import (
	"fmt"
	"os" // Imported for os.Stdout

	tea "github.com/charmbracelet/bubbletea"
	// AppAccess provides GetInterpreter(), GetLogger(), and potentially other app interactions.
	// The 'app' instance passed to Start() will implement AppAccess.
)

// Start initializes and runs the Bubble Tea TUI.
// It now accepts an initialScriptPath, which if non-empty, will be executed
// by the TUI after initialization.
func Start(app AppAccess, initialScriptPath string) error {
	if app == nil {
		return fmt.Errorf("cannot start TUI with a nil application reference (AppAccess)")
	}

	logger := app.GetLogger()
	if logger == nil {
		// This should ideally be caught earlier or AppAccess should guarantee a logger.
		fmt.Fprintln(os.Stderr, "Critical Warning: TUI Start received a nil logger from app access. TUI logging will be impaired.")
		// Consider using a NoOpLogger if available globally or returning an error.
	} else {
		logger.Debug("TUI Start: Initializing...")
	}

	// Pass the initial script path to the model constructor.
	// The model will store it and its Init() method can return a command to run it.
	model := newModel(app, initialScriptPath)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Set the tea.Program instance on the model if the model needs to send Cmds directly.
	// This is also useful if TUIEmitWriter were to be created/managed by the model itself.
	model.SetTeaProgram(p) // Assuming model has SetTeaProgram

	interpreter := app.GetInterpreter()
	if interpreter != nil {
		// Ensure interpreter's stdout is correctly managed for TUI EMIT statements.
		originalStdoutWriter := interpreter.Stdout() // Assuming Interpreter has a Stdout() getter or public field.
		// If not, we can't store the specific original, just restore to os.Stdout.

		emitWriter := NewTUIEmitWriter(p)
		interpreter.SetStdout(emitWriter)
		if logger != nil {
			logger.Debug("Switched interpreter stdout to TUIEmitWriter.")
		}

		defer func() {
			// Restore the original stdout writer when the TUI exits.
			// If we couldn't get the specific original, restore to os.Stdout.
			if originalStdoutWriter != nil {
				interpreter.SetStdout(originalStdoutWriter)
				if logger != nil {
					logger.Debug("Restored interpreter stdout to its original writer after TUI exit.")
				}
			} else {
				interpreter.SetStdout(os.Stdout) // Default fallback
				if logger != nil {
					logger.Debug("Restored interpreter stdout to os.Stdout after TUI exit (original not available).")
				}
			}
		}()
	} else {
		if logger != nil {
			logger.Error("TUI Start: Interpreter is nil. EMIT statements from scripts may not appear in TUI.")
		}
	}

	if logger != nil {
		logger.Debug("Starting Bubble Tea program...")
	}
	// p.Run() is blocking. The deferred function will execute after it returns.
	_, err := p.Run()

	if err != nil {
		if logger != nil {
			logger.Error("Bubble Tea program run failed", "error", err)
		}
		// The deferred restore of stdout will still happen.
		return fmt.Errorf("error running TUI: %w", err)
	}

	if logger != nil {
		logger.Debug("Bubble Tea program finished.")
	}
	return nil
}
