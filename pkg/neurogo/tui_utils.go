// NeuroScript Version: 0.4.0
// File version: 0.1.1 // Added FormatEventKeyForLogging
// Description: Utility functions and types for the TUI package.
// filename: pkg/neurogo/tui_utils.go
package neurogo

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"	// Added for tcell.KeyNames and tcell.EventKey
	"github.com/rivo/tview"
)

// posmod calculates a non-negative S_size for a given number and modulus.
func posmod(a, b int) (c int) {
	if b == 0 {
		return 0
	}
	c = a % b
	if c < 0 {
		c += b
	}
	return c
}

// EscapeTviewTags escapes '[' characters for safe display in tview's dynamic color tags.
func EscapeTviewTags(s string) string {
	s = strings.ReplaceAll(s, "[", "[[")
	return s
}

// FormatEventKeyForLogging provides a human-readable string for a tcell.EventKey.
func FormatEventKeyForLogging(event *tcell.EventKey) string {
	if event == nil {
		return "<nil_event>"
	}
	key := event.Key()
	if key == tcell.KeyRune {
		return fmt.Sprintf("Rune[%c (%d)]", event.Rune(), event.Rune())
	} else if key == tcell.KeyCtrlUnderscore || key == tcell.KeyCtrlSpace {
		// These might not have standard names or might be tricky
		return fmt.Sprintf("Key[%d]", key)
	} else if key < tcell.KeyCtrlA || key > tcell.KeyCtrlZ && (key < tcell.KeyF1 || key > tcell.KeyF64) && key != tcell.KeyBackspace && key != tcell.KeyTab && key != tcell.KeyEnter && key != tcell.KeyEsc && key != tcell.KeyDelete && key != tcell.KeyInsert && key != tcell.KeyHome && key != tcell.KeyEnd && key != tcell.KeyPgUp && key != tcell.KeyPgDn && key != tcell.KeyUp && key != tcell.KeyDown && key != tcell.KeyLeft && key != tcell.KeyRight {
		// Check if there's a name in KeyNames for it
		if name, ok := tcell.KeyNames[key]; ok {
			return name
		}
		return fmt.Sprintf("Key[%d]", key)	// Fallback for other special keys without a common name
	}
	// For standard keys that have names
	if name, ok := tcell.KeyNames[key]; ok {
		return name
	}
	return fmt.Sprintf("Key[%d]", key)	// Fallback if no name found
}

// tviewWriter is an io.Writer that writes to a tview.TextView and queues an update.
type tviewWriter struct {
	app		*tview.Application
	textView	*tview.TextView
}

// NewTviewWriter creates a new tviewWriter.
func NewTviewWriter(app *tview.Application, textView *tview.TextView) *tviewWriter {
	return &tviewWriter{app: app, textView: textView}
}

// Write implements io.Writer.
func (tw *tviewWriter) Write(p []byte) (n int, err error) {
	if tw.textView == nil {
		return 0, fmt.Errorf("tviewWriter.textView is nil")
	}
	n, err = tw.textView.Write(p)
	if tw.app != nil {
		tw.app.QueueUpdateDraw(func() {})	// Trigger redraw
	}
	return
}