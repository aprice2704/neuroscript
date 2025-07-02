// filename: pkg/interfaces/logger.go
package interfaces

import "log/slog"	// Import slog for level constants mapping

// LogLevel defines the severity level for logging.
// Uses standard slog levels for compatibility.
type LogLevel int

const (
	LogLevelDebug	LogLevel	= LogLevel(slog.LevelDebug)	// Debug level (-4)
	LogLevelInfo	LogLevel	= LogLevel(slog.LevelInfo)	// Info level (0)
	LogLevelWarn	LogLevel	= LogLevel(slog.LevelWarn)	// Warn level (4)
	LogLevelError	LogLevel	= LogLevel(slog.LevelError)	// Error level (8)
)

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

	// --- New method to control verbosity ---
	SetLevel(level LogLevel)

	// Note: Consider adding Fatal and With if needed universally,
	// but they were not in the last interface definition provided.
	// Fatal(msg string, args ...any)
	// With(args ...any) Logger
}