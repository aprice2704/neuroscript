// filename: pkg/adapters/noop_logger.go
package adapters

import "github.com/aprice2704/neuroscript/pkg/interfaces"

// NoOpLogger is a logger implementation that performs no actions.
// It satisfies the updated interfaces.Logger interface which includes
// Printf-style formatting methods.
type NoOpLogger struct{}

// NewNoOpLogger creates a new instance of NoOpLogger.
func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

// Ensure NoOpLogger implements the updated interfaces.Logger at compile time.
var _ interfaces.Logger = (*NoOpLogger)(nil)

func (l *NoOpLogger) Debug(msg string, args ...any)      {}
func (l *NoOpLogger) Info(msg string, args ...any)       {}
func (l *NoOpLogger) Warn(msg string, args ...any)       {}
func (l *NoOpLogger) Error(msg string, args ...any)      {}
func (l *NoOpLogger) Debugf(format string, args ...any)  {}
func (l *NoOpLogger) Infof(format string, args ...any)   {}
func (l *NoOpLogger) Warnf(format string, args ...any)   {}
func (l *NoOpLogger) Errorf(format string, args ...any)  {}
func (a *NoOpLogger) SetLevel(level interfaces.LogLevel) {}
