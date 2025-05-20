// NeuroScript Version: 0.4.0
// File version: 0.1.0
// Description: Utility functions and types for the TUI package.
// filename: pkg/neurogo/tui_utils.go
// nlines: 40 // Approximate
// risk_rating: LOW
package neurogo

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// posmod calculates a non-negative S_size for a given number and modulus.
// Useful for cycling through slices.
func posmod(a, b int) (c int) {
	if b == 0 { // Avoid division by zero if screens list is empty or other modulus is zero
		return 0
	}
	c = a % b
	if c < 0 {
		c += b
	}
	return c
}

// EscapeTviewTags escapes '[' characters for safe display in tview's dynamic color tags.
// Note: This is a basic escape. More complex situations might need a robust parser
// or ensuring content doesn't unintentionally form valid tags.
func EscapeTviewTags(s string) string {
	s = strings.ReplaceAll(s, "[", "[[")
	// To correctly escape ']', one would use '[]]', but this may conflict
	// if ']' is intended as part of a tag like [-].
	// Generally, escaping '[' is the primary concern for accidental tag formation.
	return s
}

// tviewWriter is an io.Writer that writes to a tview.TextView and queues an update.
// Potentially useful for redirecting general output streams to a TUI pane.
type tviewWriter struct {
	app      *tview.Application
	textView *tview.TextView
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
	// Note: tview.TextView.Write is not inherently thread-safe if called
	// from goroutines other than the main tview event loop.
	// QueueUpdateDraw should be used if writes can happen concurrently.
	// However, for direct calls from within tview handlers, this might be okay.
	// For safety, especially if this writer is used broadly:
	// if tw.app != nil {
	//    tw.app.QueueUpdate(func() {
	//        n, err = tw.textView.Write(p) // Perform actual write on main thread
	//    })
	//    tw.app.Draw() // Ensure redraw is scheduled
	//    return len(p), nil // Assume success, error handling within QueueUpdate is complex
	// } else {
	// Direct write if no app to queue with (e.g., during setup before app.Run)
	n, err = tw.textView.Write(p)
	// }
	// The original had QueueUpdateDraw(func(){}) which is just a redraw trigger.
	// If Write itself needs to be on the main thread, the above QueueUpdate is better.
	// For now, keeping it simple and assuming writes are managed correctly by caller context.
	if tw.app != nil {
		tw.app.QueueUpdateDraw(func() {}) // Trigger redraw
	}
	return
}
