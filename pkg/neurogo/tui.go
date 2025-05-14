// NeuroScript Version: 0.3.0
// File version: 0.0.6
// Set tuiModelInstance on App in Start function.
// filename: pkg/neurogo/tui.go
// nlines: 65
// risk_rating: MEDIUM
package neurogo

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Start initializes and runs the Bubble Tea TUI.
func Start(app *App, initialScriptPath string) error {
	if app == nil {
		return fmt.Errorf("cannot start TUI with a nil application reference")
	}

	logger := app.GetLogger()
	if logger == nil {
		fmt.Fprintln(os.Stderr, "Critical Warning: TUI Start received a nil logger. TUI logging will be impaired.")
	} else {
		logger.Debug("TUI Start: Initializing...")
	}

	// Create the TUI model (value type)
	tuiModelValue := newModel(app, initialScriptPath)
	// Set the TUI model instance on the App so screens can access it if necessary
	app.SetTUImodel(&tuiModelValue) // Pass address of the model

	// Pass the address of the model to tea.NewProgram
	p := tea.NewProgram(&tuiModelValue, tea.WithAltScreen(), tea.WithMouseCellMotion())
	tuiModelValue.SetTeaProgram(p) // SetTeaProgram is on the value receiver for model

	interpreter := app.GetInterpreter()
	var originalStdoutWriter io.Writer
	if interpreter != nil {
		originalStdoutWriter = interpreter.Stdout()
		emitWriter := NewTUIEmitWriter(p)
		interpreter.SetStdout(emitWriter)
		if logger != nil {
			logger.Debug("Switched interpreter stdout to TUIEmitWriter.")
		}
	} else if logger != nil {
		logger.Error("TUI Start: Interpreter not available. EMIT statements might not appear in TUI.")
	}

	if logger != nil {
		logger.Debug("Starting Bubble Tea program...")
	}

	// tea.Program.Run() returns tea.Model, which is an interface.
	// We need to type assert it back to *model to access our specific fields.
	finalTeaModelInterface, runErr := p.Run()

	// Restore original stdout
	if interpreter != nil {
		if originalStdoutWriter != nil {
			interpreter.SetStdout(originalStdoutWriter)
			if logger != nil {
				logger.Debug("Restored interpreter stdout to original writer.")
			}
		} else {
			interpreter.SetStdout(os.Stdout) // Fallback
			if logger != nil {
				logger.Debug("Restored interpreter stdout to os.Stdout (original was nil).")
			}
		}
	}

	if finalTeaModel, ok := finalTeaModelInterface.(*model); ok {
		if finalTeaModel.lastError != nil && runErr == nil { // If model had an error but p.Run() didn't
			runErr = fmt.Errorf("TUI model exited with internal error: %w", finalTeaModel.lastError)
		}
	} else if runErr == nil { // p.Run() was ok, but model type assertion failed
		// This case should ideally not happen if NewProgram received *model
		logger.Error("TUI finished, but final model type assertion failed.", "expected_type", "*model", "actual_type", fmt.Sprintf("%T", finalTeaModelInterface))
	}

	if runErr != nil {
		if logger != nil {
			logger.Error("Bubble Tea program run failed", "error", runErr)
		}
		return fmt.Errorf("error running TUI: %w", runErr)
	}

	if logger != nil {
		logger.Debug("Bubble Tea program finished.")
	}
	return nil
}
