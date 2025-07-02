// NeuroScript Version: 0.3.1
// File version: 1.4.0
// Purpose: Corrected the return signature of NewTestInterpreter to return an error, fixing a type mismatch in tests.
// filename: pkg/logging/helpers.go

package logging

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// NOTE: This file assumes a 'logging_flags.go' and 'utils.go' exist in the package
// to provide TestVerbose and coreNoOpLogger.

// --- Internal Test Logger ---
type TestLogger struct {
	t	*testing.T
	out	io.Writer
}

var _ interfaces.Logger = (*TestLogger)(nil)

func NewTestLogger(t *testing.T) interfaces.Logger {
	if TestVerbose != nil && !*TestVerbose {
		return &utils.coreNoOpLogger{}
	}
	return &TestLogger{t: t, out: os.Stderr}
}

func (l *TestLogger) logStructured(level string, msg string, args ...any) {
	var sb strings.Builder
	sb.WriteString(level)
	sb.WriteString(" ")
	sb.WriteString(msg)
	for i := 0; i < len(args); i += 2 {
		sb.WriteString(fmt.Sprintf(" %v=%v", args[i], args[i+1]))
	}
	l.t.Log(sb.String())
}

func (l *TestLogger) Debug(msg string, args ...any)	{ l.logStructured("[DEBUG]", msg, args...) }
func (l *TestLogger) Info(msg string, args ...any)	{ l.logStructured("[INFO]", msg, args...) }
func (l *TestLogger) Warn(msg string, args ...any)	{ l.logStructured("[WARN]", msg, args...) }
func (l *TestLogger) Error(msg string, args ...any)	{ l.logStructured("[ERROR]", msg, args...) }

func (l *TestLogger) SetLevel(level interfaces.LogLevel)	{}
func (l *TestLogger) Debugf(format string, args ...any)		{ l.t.Logf("[DEBUG] "+format, args...) }
func (l *TestLogger) Infof(format string, args ...any)		{ l.t.Logf("[INFO] "+format, args...) }
func (l *TestLogger) Warnf(format string, args ...any)		{ l.t.Logf("[WARN] "+format, args...) }
func (l *TestLogger) Errorf(format string, args ...any)		{ l.t.Logf("[ERROR] "+format, args...) }
func (l *TestLogger) With(args ...any) interfaces.Logger	{ return l }

// --- End Test Logger ---

func IsRunningInTestMode() bool {
	return flag.Lookup("test.v") != nil
}