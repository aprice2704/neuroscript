// internal/interfaces/logger.go
package interfaces

// Required for the format string methods if implemented directly

// Logger defines a standard interface for logging operations,
// compatible with log/slog levels and including Printf-style formatting methods.
type Logger interface {
	// --- Standard slog-style logging (structured key-value pairs) ---

	// Debug logs a message at Debug level with optional key-value pairs.
	Debug(msg string, args ...any) // args should be key-value pairs
	// Info logs a message at Info level with optional key-value pairs.
	Info(msg string, args ...any)
	// Warn logs a message at Warn level with optional key-value pairs.
	Warn(msg string, args ...any)
	// Error logs a message at Error level with optional key-value pairs.
	Error(msg string, args ...any)

	// --- Printf-style logging (formatted string message) ---

	// Debugf logs a formatted message at the Debug level.
	Debugf(format string, args ...any)
	// Infof logs a formatted message at the Info level.
	Infof(format string, args ...any)
	// Warnf logs a formatted message at the Warn level.
	Warnf(format string, args ...any)
	// Errorf logs a formatted message at the Error level.
	Errorf(format string, args ...any)

	// --- Optional Additions ---
	// With returns a new Logger that includes the given attributes in all subsequent logs.
	// With(args ...any) Logger // Uncomment if you want to add persistent attributes
}
