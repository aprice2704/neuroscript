// NeuroScript Version: 0.3.0
// File version: 0.0.5
// Use tui.AppAccess interface for Start function.
// Correctly handle p.Run() return values and type assertion.
// filename: pkg/neurogo/tui/tui.go
// nlines: 60
// risk_rating: MEDIUM
package tui

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Start initializes and runs the Bubble Tea TUI.
func Start(app AppAccess, initialScriptPath string) error { // app is tui.AppAccess
	if app == nil {
		return fmt.Errorf("cannot start TUI with a nil application reference (tui.AppAccess)")
	}

	logger := app.GetLogger()
	if logger == nil {
		fmt.Fprintln(os.Stderr, "Critical Warning: TUI Start received a nil logger from app access. TUI logging will be impaired.")
	} else {
		logger.Debug("TUI Start: Initializing...")
	}

	tuiModel := newModel(app, initialScriptPath) // Renamed to tuiModel to avoid conflict with package name

	p := tea.NewProgram(tuiModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
	tuiModel.SetTeaProgram(p) // Use tuiModel here

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
		logger.Error("TUI Start: Interpreter not available via AppAccess. EMIT statements from scripts might not appear in TUI.")
	}

	if logger != nil {
		logger.Debug("Starting Bubble Tea program...")
	}

	finalTeaModel, runErr := p.Run() // Correctly capture both return values

	// Access final model state if needed, e.g., for error checking or logging
	if m, ok := finalTeaModel.(model); ok { // Corrected type assertion to use the struct name `model`
		if m.lastError != nil && runErr == nil {
			runErr = fmt.Errorf("TUI model exited with internal error: %w", m.lastError)
		}
	}

	if interpreter != nil {
		if originalStdoutWriter != nil {
			interpreter.SetStdout(originalStdoutWriter)
			if logger != nil {
				logger.Debug("Restored interpreter stdout to its original writer after TUI exit.")
			}
		} else {
			interpreter.SetStdout(os.Stdout)
			if logger != nil {
				logger.Debug("Restored interpreter stdout to os.Stdout after TUI exit (original not available).")
			}
		}
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
