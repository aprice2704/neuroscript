// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Removes mutable Set methods; I/O streams are now configured via HostContext at startup.
// filename: pkg/interpreter/io.go
// nlines: 30
// risk_rating: LOW

package interpreter

import (
	"fmt"
	"io"
)

// Println satisfies the tool.Runtime interface, providing a way for tools to print output.
func (i *Interpreter) Println(a ...any) {
	fmt.Fprintln(i.Stdout(), a...)
}

func (i *Interpreter) Stdout() io.Writer {
	if i.hostContext == nil || i.hostContext.Stdout == nil {
		panic("FATAL: Interpreter has no stdout writer configured in its HostContext.")
	}
	return i.hostContext.Stdout
}

func (i *Interpreter) Stderr() io.Writer {
	if i.hostContext == nil || i.hostContext.Stderr == nil {
		panic("FATAL: Interpreter has no stderr writer configured in its HostContext.")
	}
	return i.hostContext.Stderr
}

func (i *Interpreter) Stdin() io.Reader {
	if i.hostContext == nil || i.hostContext.Stdin == nil {
		panic("FATAL: Interpreter has no stdin reader configured in its HostContext.")
	}
	return i.hostContext.Stdin
}
