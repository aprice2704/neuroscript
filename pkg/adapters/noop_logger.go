// filename: pkg/adapters/noop_logger.go
package adapters

import "github.com/aprice2704/neuroscript/pkg/logging"

// NoOpLogger is a logger implementation that performs no actions.
// It satisfies the updated logging.Logger interface which includes
// Printf-style formatting methods.
type NoOpLogger struct{}

// NewNoOpLogger creates a new instance of NoOpLogger.
func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

// Ensure NoOpLogger implements the updated logging.Logger at compile time.
var _ logging.Logger = (*NoOpLogger)(nil)

// --- Standard slog-style logging ---

// Debug logs a debug message. Does nothing in NoOpLogger.
func (l *NoOpLogger) Debug(msg string, args ...any) {}

// Info logs an informational message. Does nothing in NoOpLogger.
func (l *NoOpLogger) Info(msg string, args ...any) {}

// Warn logs a warning message. Does nothing in NoOpLogger.
func (l *NoOpLogger) Warn(msg string, args ...any) {}

// Error logs an error message. Does nothing in NoOpLogger.
func (l *NoOpLogger) Error(msg string, args ...any) {}

// --- Printf-style logging ---

// Debugf logs a formatted debug message. Does nothing in NoOpLogger.
func (l *NoOpLogger) Debugf(format string, args ...any) {}

// Infof logs a formatted informational message. Does nothing in NoOpLogger.
func (l *NoOpLogger) Infof(format string, args ...any) {}

// Warnf logs a formatted warning message. Does nothing in NoOpLogger.
func (l *NoOpLogger) Warnf(format string, args ...any) {}

// Errorf logs a formatted error message. Does nothing in NoOpLogger.
func (l *NoOpLogger) Errorf(format string, args ...any) {}

func (a *NoOpLogger) SetLevel(level logging.LogLevel) {
}

// Note: The logging.Logger provided does not include Fatal or With methods.
// If they were intended, the interface definition would need updating.
// Fatal logs a fatal message and exits. Does nothing in NoOpLogger (does not exit).
// func (l *NoOpLogger) Fatal(msg string, args ...any) {}

// With returns a new logger with the specified attributes. Returns the same NoOpLogger.
// func (l *NoOpLogger) With(args ...any) logging.Logger {
// 	 return l // Return the same instance as it has no state
// }
