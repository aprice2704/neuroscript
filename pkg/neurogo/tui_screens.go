// NeuroScript Version: 0.4.0
// File version: 0.4.0
// filename: pkg/neurogo/tui_screens.go
// nlines: 80 // Approximate
// risk_rating: LOW
// Short description: Defines Screener interface and implementations for TUI.
// Changes:
// - DynamicOutputScreen now simply buffers output. Refresh callback removed for this phase.
// - Focus on initial display of buffered content. Live updates deferred.
// - Updated Screener interface documentation.

package neurogo

import (
	"fmt"
	"strings"
	"sync"
)

// Screener defines the interface for all displayable screens within TUI panes.
// Its methods provide the necessary information for the TUI to display and manage the screen.
type Screener interface {
	Name() string     // Returns a short, unique, human-readable name or identifier for the screen. Used in status bars, logs.
	Title() string    // Returns the title to be displayed, typically at the top of the pane where the screen is rendered.
	Contents() string // Returns the main text content to be displayed. For dynamic screens, this reflects their current state.
}

// --- StaticScreen ---
// StaticScreen is a Screener implementation for screens with fixed, predefined content.
type StaticScreen struct {
	name     string
	title    string
	contents string
}

// NewStaticScreen creates a new StaticScreen.
func NewStaticScreen(name, title, contents string) *StaticScreen {
	return &StaticScreen{
		name:     name,
		title:    title,
		contents: contents,
	}
}

// Name returns the name of the StaticScreen.
func (ss *StaticScreen) Name() string {
	return ss.name
}

// Title returns the title of the StaticScreen.
func (ss *StaticScreen) Title() string {
	return ss.title
}

// Contents returns the pre-defined content of the StaticScreen.
func (ss *StaticScreen) Contents() string {
	return ss.contents
}

// --- DynamicOutputScreen ---
// DynamicOutputScreen is a Screener implementation that buffers content written to it
// (as an io.Writer). Its Contents() method returns the current buffer.
// For this version, it does not actively push updates to the TUI; the TUI must
// call Contents() to get the latest state (e.g., during setScreen).
type DynamicOutputScreen struct {
	mu      sync.Mutex
	name    string
	title   string
	builder strings.Builder
}

// NewDynamicOutputScreen creates a new DynamicOutputScreen.
func NewDynamicOutputScreen(name, title string) *DynamicOutputScreen {
	return &DynamicOutputScreen{
		name:  name,
		title: title,
		// builder is initialized as empty strings.Builder
	}
}

// Name returns the name of the DynamicOutputScreen.
func (s *DynamicOutputScreen) Name() string {
	// No lock needed as name is immutable after creation
	return s.name
}

// Title returns the title of the DynamicOutputScreen.
func (s *DynamicOutputScreen) Title() string {
	// No lock needed as title is immutable after creation
	return s.title
}

// Contents returns the current buffered content as a string.
func (s *DynamicOutputScreen) Contents() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.builder.String()
}

// Write implements io.Writer. It appends data to the internal buffer.
// For this simplified version, it does NOT trigger any TUI refresh callback.
func (s *DynamicOutputScreen) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, err = s.builder.Write(p)
	return n, err
}

// Clear resets the internal buffer.
// Note: This clear will only be visible in the TUI if Contents() is called again
// (e.g., by setScreen or a future refresh mechanism).
func (s *DynamicOutputScreen) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.builder.Reset()
}

// helpText is used by the default Help screen.
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
