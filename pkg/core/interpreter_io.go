// NeuroScript Version: 0.4.2
// File version: 1
// Purpose: Manages the standard I/O streams for the interpreter.
// filename: pkg/core/interpreter_io.go
// nlines: 48
// risk_rating: LOW
package core

import (
	"io"
	"os"
)

func (i *Interpreter) SetStdout(writer io.Writer) {
	if writer == nil {
		i.logger.Warn("Attempted to set nil stdout writer on interpreter, using os.Stdout as fallback.")
		i.stdout = os.Stdout
		return
	}
	i.stdout = writer
}

func (i *Interpreter) Stdout() io.Writer {
	if i.stdout == nil {
		return os.Stdout
	}
	return i.stdout
}

func (i *Interpreter) SetStderr(writer io.Writer) {
	if writer == nil {
		i.logger.Warn("Attempted to set nil stderr writer on interpreter, using os.Stderr as fallback.")
		i.stderr = os.Stderr
		return
	}
	i.stderr = writer
}

func (i *Interpreter) Stderr() io.Writer {
	if i.stderr == nil {
		return os.Stderr
	}
	return i.stderr
}

func (i *Interpreter) SetStdin(reader io.Reader) {
	if reader == nil {
		i.logger.Warn("Attempted to set nil stdin reader on interpreter, using os.Stdin as fallback.")
		i.stdin = os.Stdin
		return
	}
	i.stdin = reader
}

func (i *Interpreter) Stdin() io.Reader {
	if i.stdin == nil {
		return os.Stdin
	}
	return i.stdin
}
