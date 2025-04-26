// internal/adapters/slogadapter/slogadapter.go
package slogadapter

import (
	"fmt"
	"log/slog"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// SlogAdapter adapts the standard log/slog Logger to the interfaces.Logger interface.
type SlogAdapter struct {
	logger *slog.Logger
}

// Compile-time check to ensure SlogAdapter implements interfaces.Logger
var _ interfaces.Logger = (*SlogAdapter)(nil)

// NewSlogAdapter creates a new adapter instance.
// It returns an error if the provided logger is nil.
func NewSlogAdapter(logger *slog.Logger) (*SlogAdapter, error) {
	if logger == nil {
		// Or consider returning a default no-op logger instead of an error -- NO
		return nil, fmt.Errorf("slog.Logger cannot be nil")
	}
	return &SlogAdapter{logger: logger}, nil
}

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

// --- Optional: If you add With to the interface ---
/*
func (a *SlogAdapter) With(args ...any) interfaces.Logger {
	// Create a new adapter with the child logger that includes the attributes
	newLogger := a.logger.With(args...)
	// We can ignore the error here because NewSlogAdapter only errors if logger is nil,
	// and a.logger is guaranteed not to be nil if 'a' was created correctly.
	newAdapter, _ := NewSlogAdapter(newLogger)
	return newAdapter
}
*/
