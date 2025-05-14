// NeuroScript Version: 0.3.0
// File version: 0.0.2
// Refined Screen interface for TUI views, including input handling.
// filename: pkg/neurogo/screen.go
// nlines: 40
// risk_rating: MEDIUM
package neurogo

import (
	"github.com/charmbracelet/bubbles/textarea" // For screen-specific input
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents a distinct view or mode within a TUI pane.
// Each screen is responsible for its own state, updates, and rendering.
type Screen interface {
	// Init is called when the screen becomes active or is first created.
	// It can return commands to be executed, such as an initial data load
	// or focusing an internal component. `app` provides access to global resources.
	Init(app *App) tea.Cmd

	// Update handles TUI messages (key presses, window events, custom messages).
	// `app` provides access to global resources like AIWorkerManager, Logger, etc.
	// It returns the updated Screen (can be itself or a new screen if the update
	// logic dictates a screen change) and any commands to be executed by bubbletea.
	Update(msg tea.Msg, app *App) (Screen, tea.Cmd)

	// View renders the current state of the screen as a string, given the
	// available width and height for its pane.
	View(width, height int) string

	// Name returns a unique, human-readable identifier for the screen.
	// Useful for debugging or potentially for UI elements.
	Name() string

	// SetSize informs the screen of its allocated dimensions. This is typically
	// called by the main model when the window size changes or when the screen
	// is first activated in a pane of a certain size.
	SetSize(width, height int)

	// GetInputBubble returns the textarea model if this screen has a primary input area.
	// Returns nil if the screen doesn't manage a dedicated input bubble (e.g., view-only screens).
	// The main TUI model will manage the focus and rendering of this bubble if provided.
	GetInputBubble() *textarea.Model

	// HandleSubmit is called by the main TUI model when the user submits input
	// from this screen's associated input bubble (i.e., when GetInputBubble() is not nil
	// and the user presses Enter on that input area).
	// The screen is responsible for retrieving the input value (e.g., via screen.GetInputBubble().Value())
	// and processing it. It can return a command, which might include messages to the main
	// model (e.g., to switch screens, interact with other services via App).
	HandleSubmit(app *App) tea.Cmd

	// Focus is called when this screen (or the pane it resides in) becomes the active focus.
	// The screen should use this to focus its internal components, particularly its input bubble if it has one.
	Focus(app *App) tea.Cmd

	// Blur is called when this screen (or the pane it resides in) loses focus.
	// The screen should use this to blur its internal components.
	Blur(app *App) tea.Cmd
}
