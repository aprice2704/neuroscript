package neurogo

import "fmt"

type Screener interface {
	Name() string
	Title() string
	Contents() string
}

type StaticScreen struct {
	title    string
	contents string
	name     string
}

// Methods of Screener i/f

func (h *StaticScreen) Title() string {
	return h.title
}

func (h *StaticScreen) Name() string {
	return h.name
}

func (h *StaticScreen) Contents() string {
	return h.contents
}

var helpText = fmt.Sprintf(
	`[green]Navigation:[white]

[yellow]Tab[white] cycles focus: [blue]Left Input (C)[white] -> [blue]Right Input (D)[white] -> [blue]Right Pane (B)[white] -> [blue]Left Pane (A)[white] -> (loop)
[yellow]Shift+Tab[white] cycles focus: [blue]Left Input (C)[white] -> [blue]Left Pane (A)[white] -> [blue]Right Pane (B)[white] -> [blue]Right Input (D)[white] -> (loop)

[green]Pane Content Cycling:[white]

[yellow]Ctrl+B[white] cycles Left Pane (A) screens
[yellow]Ctrl+N[white] cycles Right Pane (B) screens

[green]Commands:[white]

[yellow]//system_command [args][white] - System-level command
[yellow]/screen_command [args][white] - Screen-specific command
[yellow]regular text input[white] - Input for the active Screen

[green]Other:[white]

[yellow]?[white] - Toggle Help Display
[yellow]Ctrl+C[white] - Quit
`)
