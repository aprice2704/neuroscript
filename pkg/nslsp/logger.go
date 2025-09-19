// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Provides a complete and correct logger adapter that fully implements the interfaces.Logger interface.
// filename: pkg/nslsp/logger.go
// nlines: 75
// risk_rating: LOW

package nslsp

import (
	"fmt"
	"log"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// serverLogger adapts the standard *log.Logger to the interfaces.Logger interface.
type serverLogger struct {
	*log.Logger
	level interfaces.LogLevel
}

// NewServerLogger creates a new logger adapter, defaulting to the Info level.
func NewServerLogger(l *log.Logger) *serverLogger {
	return &serverLogger{
		Logger: l,
		level:  interfaces.LogLevelInfo,
	}
}

// SetLevel sets the minimum log level for the logger to output messages.
func (l *serverLogger) SetLevel(level interfaces.LogLevel) {
	l.level = level
}

// --- slog-style logging ---

func (l *serverLogger) Debug(msg string, args ...any) {
	if l.level <= interfaces.LogLevelDebug {
		l.Println(fmt.Sprintf("[DEBUG] %s %v", msg, args))
	}
}
func (l *serverLogger) Info(msg string, args ...any) {
	if l.level <= interfaces.LogLevelInfo {
		l.Println(fmt.Sprintf("[INFO] %s %v", msg, args))
	}
}
func (l *serverLogger) Warn(msg string, args ...any) {
	if l.level <= interfaces.LogLevelWarn {
		l.Println(fmt.Sprintf("[WARN] %s %v", msg, args))
	}
}
func (l *serverLogger) Error(msg string, args ...any) {
	if l.level <= interfaces.LogLevelError {
		l.Println(fmt.Sprintf("[ERROR] %s %v", msg, args))
	}
}

// --- Printf-style logging ---

func (l *serverLogger) Debugf(format string, args ...any) {
	if l.level <= interfaces.LogLevelDebug {
		l.Printf("[DEBUG] "+format, args...)
	}
}
func (l *serverLogger) Infof(format string, args ...any) {
	if l.level <= interfaces.LogLevelInfo {
		l.Printf("[INFO] "+format, args...)
	}
}
func (l *serverLogger) Warnf(format string, args ...any) {
	if l.level <= interfaces.LogLevelWarn {
		l.Printf("[WARN] "+format, args...)
	}
}
func (l *serverLogger) Errorf(format string, args ...any) {
	if l.level <= interfaces.LogLevelError {
		l.Printf("[ERROR] "+format, args...)
	}
}

// Compile-time check to ensure serverLogger implements interfaces.Logger
var _ interfaces.Logger = (*serverLogger)(nil)
