// filename: pkg/interfaces/logger.go
package interfaces

// Logger defines a standard interface for logging operations,
// compatible with log/slog levels and including Printf-style formatting methods.
type Logger interface {
	// --- Standard slog-style logging (structured key-value pairs) ---
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)

	// --- Printf-style logging (formatted string message) ---
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

// --- ADDED EXPORTED NoOpLogger ---
// NoOpLogger is an implementation of Logger that does nothing.
type NoOpLogger struct{}

func (l *NoOpLogger) Debug(msg string, args ...any) {}
func (l *NoOpLogger) Info(msg string, args ...any)  {}
func (l *NoOpLogger) Warn(msg string, args ...any)  {}
func (l *NoOpLogger) Error(msg string, args ...any) {}

func (l *NoOpLogger) Debugf(format string, args ...any) {}
func (l *NoOpLogger) Infof(format string, args ...any)  {}
func (l *NoOpLogger) Warnf(format string, args ...any)  {}
func (l *NoOpLogger) Errorf(format string, args ...any) {}

// --- Ensure it implements the interface ---
var _ Logger = (*NoOpLogger)(nil)
