// NeuroScript Version: 0.4.0
// File version: 0.4.2 // Added DynamicPrimitiveOutputScreen
// filename: pkg/neurogo/tui_screens.go
// nlines: 230 // Approximate
// risk_rating: LOW
// Short description: Defines Screener interfaces and implementations for TUI.
// Changes:
// - Added DynamicPrimitiveOutputScreen implementing PrimitiveScreener and io.Writer.

package neurogo

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// --- Existing Screener Interface (from your v0.4.0) ---
type Screener interface {
	Name() string
	Title() string
	Contents() string
}

// --- Existing StaticScreen (from your v0.4.0) ---
type StaticScreen struct {
	name     string
	title    string
	contents string
}

func NewStaticScreen(name, title, contents string) *StaticScreen {
	return &StaticScreen{name: name, title: title, contents: contents}
}
func (ss *StaticScreen) Name() string     { return ss.name }
func (ss *StaticScreen) Title() string    { return ss.title }
func (ss *StaticScreen) Contents() string { return ss.contents }

var _ Screener = (*StaticScreen)(nil)
var _ Screener = (*DynamicOutputScreen)(nil)
var _ io.Writer = (*DynamicOutputScreen)(nil)

// --- NEW PrimitiveScreener Interface ---
type PrimitiveScreener interface {
	Name() string
	Title() string
	Primitive() tview.Primitive
	OnFocus(appFocusController func(p tview.Primitive))
	OnBlur()
	InputHandler() func(event *tcell.EventKey, setFocusFunc func(p tview.Primitive)) *tcell.EventKey
	IsFocusable() bool
}

// --- StaticPrimitiveScreen (Implements PrimitiveScreener) ---
type StaticPrimitiveScreen struct {
	name     string
	title    string
	contents string
	textView *tview.TextView
}

func NewStaticPrimitiveScreen(name, title, contents string) *StaticPrimitiveScreen {
	tv := tview.NewTextView().
		SetText(contents).
		SetWordWrap(true).
		SetScrollable(true).
		SetDynamicColors(true)
	return &StaticPrimitiveScreen{
		name:     name,
		title:    title,
		contents: contents,
		textView: tv,
	}
}
func (sps *StaticPrimitiveScreen) Name() string               { return sps.name }
func (sps *StaticPrimitiveScreen) Title() string              { return sps.title }
func (sps *StaticPrimitiveScreen) Primitive() tview.Primitive { return sps.textView }
func (sps *StaticPrimitiveScreen) OnFocus(setFocus func(p tview.Primitive)) { /* setFocus(sps.textView) can be called here if needed */
}
func (sps *StaticPrimitiveScreen) OnBlur() {}
func (sps *StaticPrimitiveScreen) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
	return nil
}
func (sps *StaticPrimitiveScreen) IsFocusable() bool { return true }
func (sps *StaticPrimitiveScreen) Contents() string  { return sps.contents }

var _ PrimitiveScreener = (*StaticPrimitiveScreen)(nil)

// --- Existing DynamicOutputScreen (from your v0.4.0) ---
type DynamicOutputScreen struct {
	mu       sync.Mutex
	name     string
	title    string
	builder  strings.Builder
	textView *tview.TextView
}

func NewDynamicOutputScreen(name, title string) *DynamicOutputScreen {
	return &DynamicOutputScreen{name: name, title: title, textView: tview.NewTextView()}
}

func (s *DynamicOutputScreen) FlushBufferToTextView() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.textView != nil {
		s.textView.SetText(s.builder.String())
		s.textView.ScrollToBeginning() // Or ScrollToEnd
	}
}

func (s *DynamicOutputScreen) SetName(n string) *DynamicOutputScreen {
	s.name = n
	return s
}
func (s *DynamicOutputScreen) SetTitle(t string) *DynamicOutputScreen {
	s.name = t
	return s
}
func (s *DynamicOutputScreen) Name() string  { return s.name }
func (s *DynamicOutputScreen) Title() string { return s.title }
func (s *DynamicOutputScreen) Contents() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.builder.String()
}
func (s *DynamicOutputScreen) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, err = s.builder.Write(p)
	return n, err
}
func (s *DynamicOutputScreen) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.builder.Reset()
}

func (s *DynamicOutputScreen) Primitive() tview.Primitive               { return s.textView }
func (s *DynamicOutputScreen) OnFocus(setFocus func(p tview.Primitive)) {}
func (s *DynamicOutputScreen) OnBlur()                                  {}
func (s *DynamicOutputScreen) InputHandler() func(
	event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
	return nil
}
func (s *DynamicOutputScreen) IsFocusable() bool { return true }

// helpText (from your v0.4.0 file)
var helpText = fmt.Sprintf(
	`[green]Navigation:[white]

[yellow]Tab[white] cycles focus: [blue]Left Input (C)[white] -> [blue]Right Input (D)[white] -> [blue]Right Pane (B)[white] -> [blue]Left Pane (A)[white] -> (loop)
[yellow]Shift+Tab[white] cycles focus: [blue]Left Input (C)[white] -> [blue]Left Pane (A)[white] -> [blue]Right Pane (B)[white] -> [blue]Right Input (D)[white] -> (loop)

[green]Pane Content Cycling:[white]

[yellow]Ctrl+B[white] cycles Left Pane (A) screens
[yellow]Ctrl+N[white] cycles Right Pane (B) screens

[green]Commands:[white]

[yellow]//system_command [args][white] - System-level command
[yellow]/screen_command [args][white] - Screen-specific command (not yet fully implemented for specific screens)
[yellow]regular text input[white] - Input for the active Screen or system

[green]Other:[white]

[yellow]?[white] - Toggle Help Display in Left Pane
[yellow]Ctrl+C[white] - Quit
`)
