// NeuroScript Version: 0.3.0
// File version: 0.1.5
// Minimal nil check for internal s.logger on each log call.
// filename: pkg/logging/slogadapter.go
// nlines: 90 // Approximate
// risk_rating: MEDIUM
package logging

import (
	"fmt"
	"io"
	stdlog "log"		// Standard log package for critical panics
	stdslog "log/slog"	// Aliased to stdslog to avoid conflict with standard log package if used
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// SlogAdapter wraps slog.Logger to implement the interfaces.Logger interface.
type SlogAdapter struct {
	logger	*stdslog.Logger		// Changed to stdslog.Logger
	opts	*stdslog.HandlerOptions	// Changed to stdslog.HandlerOptions
	level	interfaces.LogLevel
}

// checkInternalState ensures the SlogAdapter's internal logger is not nil.
// It panics if the internal logger is nil, as this indicates a setup error.
func (a *SlogAdapter) checkInternalState() {
	if a == nil {
		stdlog.Fatalf("CRITICAL PANIC (SlogAdapter method call): Receiver 'a' (*SlogAdapter) is nil.")
	}
	if a.logger == nil {
		// This means the SlogAdapter instance was created or modified incorrectly,
		// leaving its internal slog.Logger as nil.
		stdlog.Fatalf("CRITICAL PANIC (SlogAdapter method call): Internal slog.Logger (a.logger) is nil in adapter instance: %+v. This is a bug in SlogAdapter setup or instantiation.", a)
	}
}

// NewSimpleSlogAdapter creates a new SlogAdapter.
// The return type is interfaces.Logger as per your original code.
func NewSimpleSlogAdapter(output io.Writer, level interfaces.LogLevel) (interfaces.Logger, error) {
	if output == nil {
		output = os.Stderr
	}

	var slogLevel stdslog.Level	// Changed to stdslog.Level
	switch level {
	case interfaces.LogLevelDebug:
		slogLevel = stdslog.LevelDebug
	case interfaces.LogLevelInfo:
		slogLevel = stdslog.LevelInfo
	case interfaces.LogLevelWarn:
		slogLevel = stdslog.LevelWarn
	case interfaces.LogLevelError:
		slogLevel = stdslog.LevelError
	default:
		slogLevel = stdslog.LevelInfo	// Default to Info
	}

	levelVar := new(stdslog.LevelVar)	// Changed to stdslog.LevelVar
	levelVar.Set(slogLevel)

	opts := &stdslog.HandlerOptions{	// Changed to stdslog.HandlerOptions
		Level:	levelVar,
		ReplaceAttr: func(groups []string, a stdslog.Attr) stdslog.Attr {	// Changed to stdslog.Attr
			if a.Key == stdslog.TimeKey {	// Changed to stdslog.TimeKey
				// Using a.Value.Time().UTC().Round(0) to ensure a Time object and normalize.
				// The Round(0) truncates to the second, removing monotonic clock readings if present,
				// which can simplify time comparison and display if sub-second precision isn't critical here.
				// If you need nanoseconds, you might format it differently or not round.
				return stdslog.Time(stdslog.TimeKey, a.Value.Time().UTC())	// Keep UTC, remove Round(0) if full precision is needed
			}
			return a
		},
	}

	return &SlogAdapter{
		logger:	stdslog.New(stdslog.NewTextHandler(output, opts)),	// Changed to stdslog
		opts:	opts,
		level:	level,
	}, nil
}

// Debug logs a message at DebugLevel.
func (a *SlogAdapter) Debug(msg string, args ...interface{}) {
	a.checkInternalState()	// PANIC if a.logger is nil
	a.logger.Debug(msg, args...)
}

// Info logs a message at InfoLevel.
func (a *SlogAdapter) Info(msg string, args ...interface{}) {
	a.checkInternalState()	// PANIC if a.logger is nil
	a.logger.Info(msg, args...)
}

// Warn logs a message at WarnLevel.
func (a *SlogAdapter) Warn(msg string, args ...interface{}) {
	a.checkInternalState()	// PANIC if a.logger is nil
	a.logger.Warn(msg, args...)
}

// Error logs a message at ErrorLevel.
func (a *SlogAdapter) Error(msg string, args ...interface{}) {
	a.checkInternalState()	// PANIC if a.logger is nil
	a.logger.Error(msg, args...)
}

// Debugf logs a formatted message at DebugLevel.
func (a *SlogAdapter) Debugf(format string, v ...interface{}) {
	a.checkInternalState()	// PANIC if a.logger is nil
	a.logger.Debug(fmt.Sprintf(format, v...))
}

// Infof logs a formatted message at InfoLevel.
func (a *SlogAdapter) Infof(format string, v ...interface{}) {
	a.checkInternalState()	// PANIC if a.logger is nil
	a.logger.Info(fmt.Sprintf(format, v...))
}

// Warnf logs a formatted message at WarnLevel.
func (a *SlogAdapter) Warnf(format string, v ...interface{}) {
	a.checkInternalState()	// PANIC if a.logger is nil
	a.logger.Warn(fmt.Sprintf(format, v...))
}

// Errorf logs a formatted message at ErrorLevel.
func (a *SlogAdapter) Errorf(format string, v ...interface{}) {
	a.checkInternalState()	// PANIC if a.logger is nil
	a.logger.Error(fmt.Sprintf(format, v...))
}

// SetLevel changes the logger's level.
func (a *SlogAdapter) SetLevel(level interfaces.LogLevel) {
	a.checkInternalState()	// Also check here in case SetLevel is called on a faulty adapter
	a.level = level
	var slogLevel stdslog.Level	// Changed to stdslog.Level
	switch level {
	case interfaces.LogLevelDebug:
		slogLevel = stdslog.LevelDebug
	case interfaces.LogLevelInfo:
		slogLevel = stdslog.LevelInfo
	case interfaces.LogLevelWarn:
		slogLevel = stdslog.LevelWarn
	case interfaces.LogLevelError:
		slogLevel = stdslog.LevelError
	default:
		slogLevel = stdslog.LevelInfo
	}

	// Check if a.opts is nil before dereferencing, although it should be set by NewSimpleSlogAdapter
	if a.opts == nil {
		fmt.Fprintf(os.Stderr, "[SLOG_ADAPTER_ERROR] SetLevel: a.opts is nil, cannot dynamically set level.\n")
		return
	}

	if lv, ok := a.opts.Level.(*stdslog.LevelVar); ok {	// Changed to stdslog.LevelVar
		if lv == nil {	// Additional check if the LevelVar itself is nil
			fmt.Fprintf(os.Stderr, "[SLOG_ADAPTER_ERROR] SetLevel: opts.Level is a *slog.LevelVar but the pointer is nil.\n")
			// Optionally re-initialize it, or log error. For now, log and return.
			// Re-initializing here might hide a deeper issue.
			// Example: newLevelVar := new(stdslog.LevelVar); newLevelVar.Set(slogLevel); a.opts.Level = newLevelVar
			return
		}
		lv.Set(slogLevel)
	} else {
		// This message is fine as a fallback if dynamic level setting isn't critical path for error.
		// If this state is unexpected, could also panic.
		fmt.Fprintf(os.Stderr, "[SLOG_ADAPTER_ERROR] SetLevel: opts.Level is not a *slog.LevelVar (type: %T), cannot dynamically set level.\n", a.opts.Level)
	}
}

// LogLevelFromString converts a string to a interfaces.LogLevel.
// This function was already present in your provided code and is kept.
// It uses your `interfaces.LogLevel` constants.
func LogLevelFromString(levelStr string) (interfaces.LogLevel, error) {
	switch strings.ToLower(levelStr) {
	// Assuming your interfaces package defines these constants:
	// interfaces.LogLevelDebug, interfaces.LogLevelInfo, interfaces.LogLevelWarn, interfaces.LogLevelError
	case "debug":
		return interfaces.LogLevelDebug, nil
	case "info":
		return interfaces.LogLevelInfo, nil
	case "warn", "warning":
		return interfaces.LogLevelWarn, nil
	case "error", "err":
		return interfaces.LogLevelError, nil
	default:
		return interfaces.LogLevelInfo, fmt.Errorf("unknown log level: %s", levelStr)
	}
}