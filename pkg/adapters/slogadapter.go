// internal/adapters/slogadapter/slogadapter.go
package adapters

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/aprice2704/neuroscript/pkg/logging"
)

// SlogAdapter adapts the standard log/slog Logger to the logging.Logger interface.
type SlogAdapter struct {
	logger  *slog.Logger
	handler slog.Handler
	opts    *slog.HandlerOptions
}

// Compile-time check to ensure SlogAdapter implements the updated logging.Logger
var _ logging.Logger = (*SlogAdapter)(nil)

// A simple logger to an io output and at a given level
func NewSimpleSlogAdapter(output io.Writer, level logging.LogLevel) (*SlogAdapter, error) {
	lv486 := new(slog.LevelVar)
	lv486.Set(slog.LevelInfo)
	newopts := &slog.HandlerOptions{AddSource: false, Level: lv486}
	newhandler := slog.NewTextHandler(output, newopts)
	newlogger := slog.New(newhandler)
	nadapt := &SlogAdapter{logger: newlogger, handler: newhandler, opts: newopts}
	return nadapt, nil
}

// --- Standard slog-style logging ---

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

// --- Printf-style logging ---

// Debugf logs a formatted message at the Debug level.
func (a *SlogAdapter) Debugf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.logger.Debug(msg)
}

// Infof logs a formatted message at the Info level.
func (a *SlogAdapter) Infof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.logger.Info(msg)
}

// Warnf logs a formatted message at the Warn level.
func (a *SlogAdapter) Warnf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.logger.Warn(msg)
}

// Errorf logs a formatted message at the Error level.
func (a *SlogAdapter) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.logger.Error(msg)
}

func (a *SlogAdapter) SetLevel(level logging.LogLevel) {
	a.opts.Level.(*slog.LevelVar).Set(slog.Level(level))
}
