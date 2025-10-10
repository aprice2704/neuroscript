// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Reverted Logger() to GetLogger() to minimize external API changes.
// filename: pkg/interpreter/interpreter_io.go
// nlines: 48
// risk_rating: LOW
package interpreter

import (
	"fmt"
	"io"
	"os"
)

// Println satisfies the tool.Runtime interface, providing a way for tools to print output.
func (i *Interpreter) Println(a ...any) {
	fmt.Fprintln(i.Stdout(), a...)
}

func (i *Interpreter) SetStdout(writer io.Writer) {
	if writer == nil {
		i.GetLogger().Warn("Attempted to set nil stdout writer on interpreter, using os.Stdout as fallback.")
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
		i.GetLogger().Warn("Attempted to set nil stderr writer on interpreter, using os.Stderr as fallback.")
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
		i.GetLogger().Warn("Attempted to set nil stdin reader on interpreter, using os.Stdin as fallback.")
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
