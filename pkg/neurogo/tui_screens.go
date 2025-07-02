// NeuroScript Version: 0.4.0
// File version: 0.4.4 // Corrected InputHandler signatures for PrimitiveScreeners
// filename: pkg/neurogo/tui_screens.go
package neurogo

import (
	"fmt"
	"io"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Screener interface (remains the same)
type Screener interface {
	Name() string
	Title() string
	Contents() string
}

// StaticScreen (remains the same)
type StaticScreen struct {
	name		string
	title		string
	contents	string
}

func NewStaticScreen(name, title, contents string) *StaticScreen {
	return &StaticScreen{name: name, title: title, contents: contents}
}
func (ss *StaticScreen) Name() string		{ return ss.name }
func (ss *StaticScreen) Title() string		{ return ss.title }
func (ss *StaticScreen) Contents() string	{ return ss.contents }

var _ Screener = (*StaticScreen)(nil)

// PrimitiveScreener interface (remains the same)
type PrimitiveScreener interface {
	Name() string
	Title() string
	Primitive() tview.Primitive
	OnFocus(appFocusController func(p tview.Primitive))
	OnBlur()
	InputHandler() func(event *tcell.EventKey, setFocusFunc func(p tview.Primitive)) *tcell.EventKey
	IsFocusable() bool
}

// StaticPrimitiveScreen
type StaticPrimitiveScreen struct {
	name		string
	title		string
	contents	string	// Kept for Contents() method if Screener interface is also implemented
	textView	*tview.TextView
}

func NewStaticPrimitiveScreen(name, title, contents string) *StaticPrimitiveScreen {
	tv := tview.NewTextView().
		SetText(contents).
		SetWordWrap(true).
		SetScrollable(true).
		SetRegions(true).
		SetDynamicColors(true)
	return &StaticPrimitiveScreen{
		name:		name,
		title:		title,
		contents:	contents,
		textView:	tv,
	}
}
func (sps *StaticPrimitiveScreen) Name() string					{ return sps.name }
func (sps *StaticPrimitiveScreen) Title() string				{ return sps.title }
func (sps *StaticPrimitiveScreen) Primitive() tview.Primitive			{ return sps.textView }
func (sps *StaticPrimitiveScreen) OnFocus(setFocus func(p tview.Primitive))	{ setFocus(sps.textView) }
func (sps *StaticPrimitiveScreen) OnBlur()					{}

// InputHandler for StaticPrimitiveScreen
func (sps *StaticPrimitiveScreen) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
		// Get the TextView's own input handler function
		handler := sps.textView.InputHandler()
		if handler != nil {
			// Call the TextView's handler. It doesn't return the event.
			handler(event, setFocus)
		}
		// The PrimitiveScreener interface expects this function to return the event
		// if it's not fully "consumed". Since TextView's default handler handles scrolling etc.,
		// and we don't add other keybindings here that would consume events,
		// we return the original event to allow further processing by global handlers.
		return event
	}
}
func (sps *StaticPrimitiveScreen) IsFocusable() bool	{ return true }
func (sps *StaticPrimitiveScreen) Contents() string	{ return sps.contents }

var _ PrimitiveScreener = (*StaticPrimitiveScreen)(nil)
var _ Screener = (*StaticPrimitiveScreen)(nil)	// If it also implements the simpler Screener

// DynamicOutputScreen
type DynamicOutputScreen struct {
	mu		sync.Mutex
	name		string
	title		string
	textView	*tview.TextView
	app		*tview.Application
}

func NewDynamicOutputScreen(name, title string, app *tview.Application) *DynamicOutputScreen {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetRegions(true).
		SetWordWrap(true)
	return &DynamicOutputScreen{
		name:		name,
		title:		title,
		textView:	tv,
		app:		app,
	}
}

func (s *DynamicOutputScreen) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.textView == nil {
		return 0, fmt.Errorf("DynamicOutputScreen.textView is nil")
	}
	if s.app == nil {
		// Fallback: try to write directly if no app, but this might not redraw correctly
		// For robust behavior, 'app' should always be provided.
		// This is a temporary workaround if app is somehow nil.
		// The proper fix is to ensure 'app' is always passed during construction.
		currentText := s.textView.GetText(false)
		s.textView.SetText(currentText + string(p))
		s.textView.ScrollToEnd()
		return len(p), nil
		// return 0, fmt.Errorf("DynamicOutputScreen.app is nil, cannot reliably Write to TextView")
	}
	return s.textView.Write(p)	// This uses tview's app.QueueUpdateDraw internally
}

func (s *DynamicOutputScreen) FlushBufferToTextView() {
	// This method might be less critical if Write directly updates the TextView.
	// However, if there's a desire to manually trigger a scroll or ensure draw:
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.textView != nil {
		s.textView.ScrollToEnd()	// Ensure it's scrolled to end after any potential batch update
		if s.app != nil {
			s.app.QueueUpdateDraw(func() {})
		}
	}
}

func (s *DynamicOutputScreen) SetName(n string) *DynamicOutputScreen {
	s.name = n
	return s
}
func (s *DynamicOutputScreen) SetTitle(t string) *DynamicOutputScreen {
	s.title = t
	return s
}
func (s *DynamicOutputScreen) Name() string	{ return s.name }
func (s *DynamicOutputScreen) Title() string	{ return s.title }

func (s *DynamicOutputScreen) Contents() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.textView == nil {
		return ""
	}
	return s.textView.GetText(false)
}

func (s *DynamicOutputScreen) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.textView != nil {
		s.textView.Clear()
	}
}

func (s *DynamicOutputScreen) Primitive() tview.Primitive		{ return s.textView }
func (s *DynamicOutputScreen) OnFocus(setFocus func(p tview.Primitive))	{ setFocus(s.textView) }
func (s *DynamicOutputScreen) OnBlur()					{}

// InputHandler for DynamicOutputScreen
func (s *DynamicOutputScreen) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
		// Get the TextView's own input handler function
		handler := s.textView.InputHandler()
		if handler != nil {
			// Call the TextView's handler. It doesn't return the event.
			handler(event, setFocus)
		}
		// Return the original event to allow further processing by global handlers.
		return event
	}
}
func (s *DynamicOutputScreen) IsFocusable() bool	{ return true }

var _ Screener = (*DynamicOutputScreen)(nil)
var _ io.Writer = (*DynamicOutputScreen)(nil)
var _ PrimitiveScreener = (*DynamicOutputScreen)(nil)

var helpText = fmt.Sprintf(`[green]Navigation:[white]

[yellow]Tab[white]         Cycles focus: [blue]Left Input (C)[white] -> [blue]Right Input (D)[white] -> [blue]Right Pane (B)[white] -> [blue]Left Pane (A)[white] -> (loop)
[yellow]Shift+Tab[white]    Cycles focus: [blue]Left Input (C)[white] -> [blue]Left Pane (A)[white] -> [blue]Right Pane (B)[white] -> [blue]Right Input (D)[white] -> (loop)

[green]Pane Content Cycling:[white]

[yellow]Ctrl+B[white]        Cycles Left Pane (A) screens (e.g., ScriptOut, AIWM, Help)
[yellow]Ctrl+N[white]        Cycles Right Pane (B) screens (e.g., Chat, DebugLog, HelpRight)
[yellow]Ctrl+F[white]        (If implemented) Next screen in Right Pane (B)
[yellow]Ctrl+P[white]        (If implemented) Previous screen in Right Pane (B)

[green]Commands (in Input Areas C or D):[white]

[yellow]//system_cmd[white]   System-level command
[yellow]/screen_cmd[white]   Screen-specific command (if supported by active screen)
[yellow](other text)[white]  Input for the active screen or system

[green]Other Controls:[white]

[yellow]?[white]             Toggle Help Display in Left Pane (when input not focused)
[yellow]Ctrl+C[white]        Copy content of focused pane to clipboard
[yellow]Ctrl+Q[white]        Quit application
`)