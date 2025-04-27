// internal/adapters/slogadapter/slogadapter.go
package adapters // Note: Your previous code showed 'package adapters', ensure this is correct or adjust if needed.

import (
	"fmt"
	"log/slog"
	"os"

	// Ensure this import path matches where your interface is defined
	"github.com/aprice2704/neuroscript/pkg/interfaces" // Adjusted based on previous turn
)

// SlogAdapter adapts the standard log/slog Logger to the interfaces.Logger interface.
type SlogAdapter struct {
	logger *slog.Logger
}

// Compile-time check to ensure SlogAdapter implements the updated interfaces.Logger
var _ interfaces.Logger = (*SlogAdapter)(nil)

// NewSlogAdapter creates a new adapter instance.
// It returns an error if the provided logger is nil.
func NewSlogAdapter(logger *slog.Logger) (*SlogAdapter, error) {
	if logger == nil {
		// Decision from previous turn: Return error, do not panic here.
		// Panic should happen in the component constructor receiving the logger.
		return nil, fmt.Errorf("slog.Logger cannot be nil")
	}
	return &SlogAdapter{logger: logger}, nil
}

// --- Standard slog-style logging ---

// Debug logs a message at the Debug level.
func (a *SlogAdapter) Debug(msg string, args ...any) {
	a.logger.Debug(msg, args...)
}

// Info logs a message at the Info level.
func (a *SlogAdapter) Info(msg string, args ...any) {
	a.logger.Info(msg, args...)
}

// Warn logs a message at the Warn level.
func (a *SlogAdapter) Warn(msg string, args ...any) {
	a.logger.Warn(msg, args...)
}

// Error logs a message at the Error level.
func (a *SlogAdapter) Error(msg string, args ...any) {
	a.logger.Error(msg, args...)
}

// --- Printf-style logging ---

// Debugf logs a formatted message at the Debug level.
func (a *SlogAdapter) Debugf(format string, args ...any) {
	// Format the message using Sprintf
	msg := fmt.Sprintf(format, args...)
	// Pass the formatted message to the underlying logger's Debug method.
	// Do not pass the original args... here, as slog expects key-value pairs.
	a.logger.Debug(msg)
}

// Infof logs a formatted message at the Info level.
func (a *SlogAdapter) Infof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.logger.Info(msg)
}

// Warnf logs a formatted message at the Warn level.
func (a *SlogAdapter) Warnf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.logger.Warn(msg)
}

// Errorf logs a formatted message at the Error level.
func (a *SlogAdapter) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.logger.Error(msg)
}

// --- Optional: With method ---
/*
func (a *SlogAdapter) With(args ...any) interfaces.Logger {
	newLogger := a.logger.With(args...)
	// Ignoring error because NewSlogAdapter only errors if logger is nil,
	// and a.logger is guaranteed non-nil if 'a' was correctly created.
	// If NewSlogAdapter's error conditions change, this might need adjustment.
	newAdapter, _ := NewSlogAdapter(newLogger)
	return newAdapter
}
*/

// SimpleTestLogger provides a basic slog.Logger for testing purposes.
// Logs to Stderr at Warn level with source info.
func SimpleTestLogger() *SlogAdapter {
	defhandlerOpts := &slog.HandlerOptions{
		Level:     slog.LevelWarn,
		AddSource: true, // include source file and line number
	}
	handler := slog.NewTextHandler(os.Stderr, defhandlerOpts) // Log to Stderr
	s := slog.New(handler)
	sa, err := NewSlogAdapter(s)
	if err != nil {
		panic("Could not init slogadapter for SimpleTestLogger")
	}
	return sa
}
