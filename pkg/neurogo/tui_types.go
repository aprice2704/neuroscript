// NeuroScript Version: 0.4.0
// File version: 0.1.0 // Initial creation for TUI refactor
// Description: Defines types used by the TUI, primarily tviewAppPointers.
// filename: pkg/neurogo/tui_types.go
package neurogo

import (
	"io"

	"github.com/rivo/tview"
)

// tviewAppPointers holds references to common tview components and TUI state.
// It acts as a context for TUI operations.
type tviewAppPointers struct {
	tviewApp	*tview.Application	// The main tview application.
	grid		*tview.Grid		// The main layout grid.

	// Panes (implemented as tview.Pages to hold multiple screens)
	localOutputView	*tview.Pages	// Pane A: Typically for script output, AIWM status, help.
	aiOutputView	*tview.Pages	// Pane B: Typically for chat sessions, debug logs, help.

	// Input Areas
	localInputArea	*tview.TextArea	// Pane C: Input for Pane A or system commands.
	aiInputArea	*tview.TextArea	// Pane D: Input for Pane B (e.g., chat messages).

	statusBar	*tview.TextView	// Bottom status bar.

	// Focus Management
	focusablePrimitives	[]tview.Primitive	// Ordered list of primitives that can receive focus.
	currentFocusIndex	int			// Index into focusablePrimitives.
	numFocusablePrimitives	int			// Cache len(focusablePrimitives).
	// Indices to quickly find specific panes in focusablePrimitives (if needed, currently not directly used this way)
	paneAIndex	int
	paneBIndex	int
	paneCIndex	int
	paneDIndex	int

	app			*App	// Reference to the main neurogo application logic.
	initialActivityText	string	// Text for the status bar on startup.

	// Screen Management
	// Slices holding the PrimitiveScreener implementations for each pane.
	leftScreens	[]PrimitiveScreener
	rightScreens	[]PrimitiveScreener
	// Map to quickly access ChatConversationScreen instances by their sessionID.
	chatScreenMap	map[string]*ChatConversationScreen

	debugScreen	*DynamicOutputScreen	// Dedicated screen for debug messages.

	// State for currently displayed screen in each pane.
	leftShowing	int	// Index into leftScreens.
	rightShowing	int	// Index into rightScreens.
	helpScreenIndex	int	// Index of the help screen in leftScreens (for '?' shortcut).

	originalStdout	io.Writer	// To restore interpreter's stdout after TUI exits.
}