// internal/interfaces/logger.go
package interfaces

// Logger defines a standard interface for logging operations,
// compatible with log/slog levels.
type Logger interface {
	Debug(msg string, args ...any) // args should be key-value pairs
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	// Consider adding: With(args ...any) Logger to add persistent attributes
}
