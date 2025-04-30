// filename: pkg/logging/logger.go
package logging

// Logger defines a standard interface for logging operations,
// compatible with log/slog levels and including Printf-style formatting methods.
// This interface is defined in its own package to avoid import cycles.
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

	// Note: Consider adding Fatal and With if needed universally,
	// but they were not in the last interface definition provided.
	// Fatal(msg string, args ...any)
	// With(args ...any) Logger
}
