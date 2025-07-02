// NeuroScript Version: 0.4.0
// File version: 0.1.0
// Purpose: Implements a TUI screen to display the AIWorkerManager.String() output.
// filename: pkg/neurogo/tui_screen_aiwm_string.go
// nlines: 87
// risk_rating: LOW
package neurogo

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AIWMStringScreen displays the full string output of the AIWorkerManager.
type AIWMStringScreen struct {
	app		*App	// To access AIWorkerManager
	name		string
	title		string
	textView	*tview.TextView
}

// NewAIWMStringScreen creates a new screen for displaying AIWorkerManager.String().
func NewAIWMStringScreen(app *App) *AIWMStringScreen {
	if app == nil {
		// This should ideally not happen if app is always passed.
		// Consider logging this panic if a logger is available, or handle differently.
		panic("AIWMStringScreen: app cannot be nil")
	}
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(false)	// AIWM.String() might have its own formatting; wrapping might distort.
	tv.SetBorder(false)

	s := &AIWMStringScreen{
		app:		app,
		name:		"AIWMString",
		title:		"AIWM Full Status",
		textView:	tv,
	}
	// Log creation if a TUI logger is available, similar to AIWMStatusScreen
	// Example: if app.tui != nil { app.tui.LogToDebugScreen("[AIWM_STRING_NEW] NewAIWMStringScreen created for %s.", s.name) }
	return s
}

// Name returns the screen's short identifier.
func (s *AIWMStringScreen) Name() string	{ return s.name }

// Title returns the screen's current title.
func (s *AIWMStringScreen) Title() string	{ return s.title }

// Primitive returns the tview.Primitive for this screen (the TextView).
// It updates the content when the primitive is first requested.
func (s *AIWMStringScreen) Primitive() tview.Primitive {
	s.updateContent()
	return s.textView
}

// OnFocus is called when the screen gains focus. It updates the content.
func (s *AIWMStringScreen) OnFocus(setFocus func(p tview.Primitive)) {
	s.updateContent()
	setFocus(s.textView)
	s.textView.ScrollToBeginning()	// Show the start of the status string
}

// OnBlur is called when the screen loses focus. (No action needed here)
func (s *AIWMStringScreen) OnBlur()	{}

// InputHandler returns the input handler for this screen.
// It defaults to the TextView's input handler for scrolling.
func (s *AIWMStringScreen) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
		if s.textView != nil {
			handler := s.textView.InputHandler()
			if handler != nil {
				// Let the textView's default handler manage scrolling, etc.
				handler(event, setFocus)
				return event	// Return event as it might not be fully consumed
			}
		}
		return event
	}
}

// IsFocusable indicates if the screen's primitive can be focused.
func (s *AIWMStringScreen) IsFocusable() bool {
	return true
}

// updateContent fetches the AIWorkerManager.String() output and updates the TextView.
func (s *AIWMStringScreen) updateContent() {
	if s.app == nil || s.textView == nil {
		if s.textView != nil {
			s.textView.SetText("[red]Error: App or TextView not properly initialized for AIWMStringScreen[-]")
		}
		// Log error if logger is available
		// Example: if s.app != nil && s.app.Log != nil { s.app.Log.Error("App or TextView nil in AIWMStringScreen.updateContent") }
		return
	}

	aiwm := s.app.GetAIWorkerManager()
	if aiwm == nil {
		s.textView.SetText("[yellow]AIWorkerManager not yet initialized or is not available.[-]")
		return
	}

	statusString := aiwm.ColourString()	// Get the string output from AIWorkerManager
	s.textView.SetText(statusString)	// Escape any tview tags in the string
	s.textView.ScrollToBeginning()
}

// Ensure AIWMStringScreen implements PrimitiveScreener.
var _ PrimitiveScreener = (*AIWMStringScreen)(nil)