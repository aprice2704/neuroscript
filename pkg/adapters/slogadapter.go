// NeuroScript Version: 0.3.0
// File version: 0.1.4
// Removed diagnostic prints from Debug/Debugf.
// filename: pkg/adapters/slogadapter.go
// nlines: 90 // Approximate
// risk_rating: MEDIUM
package adapters

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/logging"
)

// SlogAdapter wraps slog.Logger to implement the logging.Logger interface.
type SlogAdapter struct {
	logger *slog.Logger
	opts   *slog.HandlerOptions
	level  logging.LogLevel
}

// NewSimpleSlogAdapter creates a new SlogAdapter.
func NewSimpleSlogAdapter(output io.Writer, level logging.LogLevel) (logging.Logger, error) {
	if output == nil {
		output = os.Stderr
	}

	var slogLevel slog.Level
	switch level {
	case logging.LogLevelDebug:
		slogLevel = slog.LevelDebug
	case logging.LogLevelInfo:
		slogLevel = slog.LevelInfo
	case logging.LogLevelWarn:
		slogLevel = slog.LevelWarn
	case logging.LogLevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	levelVar := new(slog.LevelVar)
	levelVar.Set(slogLevel)

	opts := &slog.HandlerOptions{
		Level: levelVar,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Time(slog.TimeKey, a.Value.Time().UTC().Round(0))
			}
			return a
		},
	}

	return &SlogAdapter{
		logger: slog.New(slog.NewTextHandler(output, opts)),
		opts:   opts,
		level:  level,
	}, nil
}

// Debug logs a message at DebugLevel.
func (a *SlogAdapter) Debug(msg string, args ...interface{}) {
	a.logger.Debug(msg, args...)
}

// Info logs a message at InfoLevel.
func (a *SlogAdapter) Info(msg string, args ...interface{}) {
	a.logger.Info(msg, args...)
}

// Warn logs a message at WarnLevel.
func (a *SlogAdapter) Warn(msg string, args ...interface{}) {
	a.logger.Warn(msg, args...)
}

// Error logs a message at ErrorLevel.
func (a *SlogAdapter) Error(msg string, args ...interface{}) {
	a.logger.Error(msg, args...)
}

// Debugf logs a formatted message at DebugLevel.
func (a *SlogAdapter) Debugf(format string, v ...interface{}) {
	a.logger.Debug(fmt.Sprintf(format, v...))
}

// Infof logs a formatted message at InfoLevel.
func (a *SlogAdapter) Infof(format string, v ...interface{}) {
	a.logger.Info(fmt.Sprintf(format, v...))
}

// Warnf logs a formatted message at WarnLevel.
func (a *SlogAdapter) Warnf(format string, v ...interface{}) {
	a.logger.Warn(fmt.Sprintf(format, v...))
}

// Errorf logs a formatted message at ErrorLevel.
func (a *SlogAdapter) Errorf(format string, v ...interface{}) {
	a.logger.Error(fmt.Sprintf(format, v...))
}

// SetLevel changes the logger's level.
func (a *SlogAdapter) SetLevel(level logging.LogLevel) {
	a.level = level
	var slogLevel slog.Level
	switch level {
	case logging.LogLevelDebug:
		slogLevel = slog.LevelDebug
	case logging.LogLevelInfo:
		slogLevel = slog.LevelInfo
	case logging.LogLevelWarn:
		slogLevel = slog.LevelWarn
	case logging.LogLevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}
	if lv, ok := a.opts.Level.(*slog.LevelVar); ok {
		lv.Set(slogLevel)
	} else {
		fmt.Fprintf(os.Stderr, "[SLOG_ADAPTER_ERROR] SetLevel: opts.Level is not a *slog.LevelVar, cannot dynamically set level.\n")
	}
}

// LogLevelFromString converts a string to a logging.LogLevel.
func LogLevelFromString(levelStr string) (logging.LogLevel, error) {
	switch strings.ToLower(levelStr) {
	case "debug":
		return logging.LogLevelDebug, nil
	case "info":
		return logging.LogLevelInfo, nil
	case "warn", "warning":
		return logging.LogLevelWarn, nil
	case "error", "err":
		return logging.LogLevelError, nil
	default:
		return logging.LogLevelInfo, fmt.Errorf("unknown log level: %s", levelStr)
	}
}
