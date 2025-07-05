// NeuroScript Version: 0.4.0
// File version: 0.1.1
// Commented out implementation due to removal of the wm package.
// filename: pkg/neurogo/tui_screen_aiwm_string.go
package neurogo

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AIWMStringScreen displays the full string output of the AIWorkerManager.
type AIWMStringScreen struct {
	app      *App
	name     string
	title    string
	textView *tview.TextView
}

// NewAIWMStringScreen creates a new screen for displaying AIWorkerManager.String().
func NewAIWMStringScreen(app *App) *AIWMStringScreen {
	if app == nil {
		panic("AIWMStringScreen: app cannot be nil")
	}
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(false)
	tv.SetBorder(false)

	s := &AIWMStringScreen{
		app:      app,
		name:     "AIWMString",
		title:    "AIWM Full Status",
		textView: tv,
	}
	return s
}

// Name returns the screen's short identifier.
func (s *AIWMStringScreen) Name() string { return s.name }

// Title returns the screen's current title.
func (s *AIWMStringScreen) Title() string { return s.title }

// Primitive returns the tview.Primitive for this screen (the TextView).
func (s *AIWMStringScreen) Primitive() tview.Primitive {
	s.updateContent()
	return s.textView
}

// OnFocus is called when the screen gains focus. It updates the content.
func (s *AIWMStringScreen) OnFocus(setFocus func(p tview.Primitive)) {
	s.updateContent()
	setFocus(s.textView)
	s.textView.ScrollToBeginning()
}

// OnBlur is called when the screen loses focus. (No action needed here)
func (s *AIWMStringScreen) OnBlur() {}

// InputHandler returns the input handler for this screen.
func (s *AIWMStringScreen) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
		if s.textView != nil {
			handler := s.textView.InputHandler()
			if handler != nil {
				handler(event, setFocus)
				return event
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
		return
	}

	// NOTE: Implementation is commented out as it depends on the AIWorkerManager,
	// which has been temporarily removed.
	s.textView.SetText("[yellow]AIWorkerManager functionality is currently disabled.[-]")

	/*
		aiwm := s.app.GetAIWorkerManager()
		if aiwm == nil {
			s.textView.SetText("[yellow]AIWorkerManager not yet initialized or is not available.[-]")
			return
		}

		statusString := aiwm.ColourString()
		s.textView.SetText(statusString)
		s.textView.ScrollToBeginning()
	*/
}

// Ensure AIWMStringScreen implements PrimitiveScreener.
var _ PrimitiveScreener = (*AIWMStringScreen)(nil)
