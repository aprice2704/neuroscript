// filename: pkg/neurogo/tui/model.go
package tui

import (
	"github.com/aprice2704/neuroscript/pkg/neurogo"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// --- Styles ---
var (
	inactiveStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	userStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	aiStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
	aiToolCallStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	sysToolCallStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("141"))
	toolResultStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	patchStatusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	systemStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	statusBarSyle    = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("250")).Padding(0, 1)
	// viewportStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), false, false, false, false) // Keep styles simple initially
	// inputAreaStyle   = lipgloss.NewStyle()
)

// --- message struct ---
type message struct {
	sender string
	text   string
}

// --- keyMap struct and implementation for help.KeyMap ---
// keyMap defines additional keybindings.
// CORRECTED: Implements help.KeyMap interface
type keyMap struct {
	Quit key.Binding
	Help key.Binding
	// Add more bindings later (Submit, Sync, ToggleSys, etc.)
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k keyMap) ShortHelp() []key.Binding {
	// Return only the bindings you want shown in the condensed help view
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	// Return bindings grouped by line for the full help view
	return [][]key.Binding{
		{k.Help, k.Quit}, // First row
		// Add more rows as needed: {k.Submit, k.Sync},
	}
}

// defaultKeyMap provides default keybindings.
var defaultKeyMap = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}

// --- model struct ---
type model struct {
	app               *neurogo.App
	viewport          viewport.Model
	textarea          textarea.Model
	spinner           spinner.Model
	help              help.Model
	keyMap            keyMap // Now implements help.KeyMap
	messages          []message
	sender            string
	lastError         error
	isWaitingForAI    bool
	activeToolMessage string
	patchStatus       string
	quitting          bool
	ready             bool
	helpVisible       bool
	aiModelName       string
	localFileCount    int
	apiFileCount      int
	syncUploads       int
	syncDeletes       int
	width             int
	height            int
}

// --- newModel constructor ---
func newModel(app *neurogo.App) model {
	ta := textarea.New()
	ta.Placeholder = "Enter prompt, /sync, /m, or /quit..."
	ta.Focus()
	ta.Prompt = "> "
	ta.CharLimit = 0
	ta.SetHeight(3)
	// Enter should not submit by default; handled in Update
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(10, 5) // Placeholder size
	// vp.UseHighPerformanceRenderer = true // Consider later

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	h := help.New()
	h.ShowAll = true // Start with full help shown during dev

	modelName := "unknown"
	if app != nil && app.Config != nil {
		modelName = app.Config.ModelName
	}

	initialMessages := []message{
		{sender: "System", text: "Welcome to neurogo TUI mode. Type '?' for help, Ctrl+C to quit."},
	}

	return model{
		app:            app,
		textarea:       ta,
		viewport:       vp,
		spinner:        sp,
		help:           h,
		keyMap:         defaultKeyMap, // Use our keyMap instance
		messages:       initialMessages,
		sender:         "You",
		aiModelName:    modelName,
		isWaitingForAI: false,
		helpVisible:    true, // Start with help visible
		ready:          false,
	}
}
