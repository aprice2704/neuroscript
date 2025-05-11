// NeuroScript Version: 0.3.0
// File version: 0.0.2 // Removed redeclaration of scriptEmitMsg.
// filename: pkg/neurogo/tui/tui_writer.go
// nlines: 38 // Approximate
// risk_rating: LOW
package tui

import (
	"fmt" // For error formatting

	tea "github.com/charmbracelet/bubbletea"
)

// TUIEmitWriter is an io.Writer that captures output (typically from EMIT statements)
// and sends it as a scriptEmitMsg to the TUI's tea.Program instance.
// scriptEmitMsg is defined in msgs.go
type TUIEmitWriter struct {
	teaProgram *tea.Program
}

// NewTUIEmitWriter creates a new TUIEmitWriter.
// It requires a pointer to the active tea.Program to send messages to the TUI.
func NewTUIEmitWriter(program *tea.Program) *TUIEmitWriter {
	if program == nil {
		// This is a programming error.
		// For now, assume program is valid based on calling context.
	}
	return &TUIEmitWriter{teaProgram: program}
}

// Write implements the io.Writer interface.
// It converts the byte slice 'p' to a string and sends it as the Content
// of a scriptEmitMsg to the TUI.
func (tew *TUIEmitWriter) Write(p []byte) (n int, err error) {
	if tew.teaProgram == nil {
		return 0, fmt.Errorf("TUIEmitWriter: tea.Program is nil, cannot send emit message")
	}

	content := string(p)
	tew.teaProgram.Send(scriptEmitMsg{Content: content}) // scriptEmitMsg from msgs.go
	return len(p), nil
}
