// filename: pkg/neurogo/tui/model.go
package tui

import (
	// Keep existing bubbletea imports
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	// DO NOT import "github.com/aprice2704/neuroscript/pkg/neurogo"
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

	// --- Debug Styles (Re-added borders) ---
	focusedCommandStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")) // Focused cmd (Pink border)
	blurredCommandStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))  // Blurred cmd (Gray border)
	focusedPromptStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")) // Focused prompt (Pink border)
	blurredPromptStyle  = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))  // Blurred prompt (Gray border)
)

// --- message struct ---
type message struct {
	sender string
	text   string
}

// --- keyMap struct and implementation for help.KeyMap ---
type keyMap struct {
	Quit key.Binding
	Help key.Binding
	Tab  key.Binding // Added Tab
	// Enter key handled directly in Update based on focus
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.Help, k.Quit},
		// Add more rows as needed
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
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch input"),
	),
}

// Constants for focused input
const (
	focusCommand = 0
	focusPrompt  = 1
)

// --- model struct ---
type model struct {
	// Use the interface type, not the concrete struct pointer
	app AppAccess

	// UI Components
	viewport     viewport.Model
	commandInput textarea.Model // Input for commands like /sync, quit, ?
	promptInput  textarea.Model // Input for AI prompts
	spinner      spinner.Model
	help         help.Model
	keyMap       keyMap

	// State
	messages        []message
	sender          string // Should always be "You" when sending prompt
	lastError       error
	isWaitingForAI  bool
	currentActivity string // Renamed from activeToolMessage for broader use (e.g., "Syncing...")
	isSyncing       bool   // Added state for sync operation
	patchStatus     string
	quitting        bool
	ready           bool
	helpVisible     bool
	focusedInput    int // 0 for command, 1 for prompt

	// Status Bar Info (initialized via interface)
	aiModelName    string
	localFileCount int
	apiFileCount   int
	syncUploads    int // These will be updated by syncCompleteMsg
	syncDeletes    int // These will be updated by syncCompleteMsg

	// Terminal Size
	width  int
	height int
}

// --- newModel constructor ---
// Change signature to accept the interface
func newModel(app AppAccess) model {
	// Command Input (smaller, focused first)
	cmdInput := textarea.New()
	cmdInput.Placeholder = "/cmd"
	cmdInput.Focus()
	cmdInput.Prompt = "$ "
	cmdInput.CharLimit = 200
	cmdInput.SetHeight(1)
	cmdInput.SetWidth(20)
	cmdInput.FocusedStyle.Base = focusedCommandStyle // Apply debug border style
	cmdInput.BlurredStyle.Base = blurredCommandStyle // Apply debug border style
	cmdInput.KeyMap.InsertNewline.SetEnabled(false)

	// Prompt Input (larger)
	promptInput := textarea.New()
	promptInput.Placeholder = "Enter prompt for AI..."
	promptInput.Prompt = "> "
	promptInput.CharLimit = 0
	promptInput.SetHeight(3)
	promptInput.SetWidth(50)
	promptInput.FocusedStyle.Base = focusedPromptStyle // Apply debug border style
	promptInput.BlurredStyle.Base = blurredPromptStyle // Apply debug border style
	promptInput.KeyMap.InsertNewline.SetEnabled(true)

	vp := viewport.New(10, 5)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	h := help.New()
	h.ShowAll = false

	// Get model name via the interface method
	modelName := "unknown"
	if app != nil {
		modelName = app.GetModelName()
	}

	initialMessages := []message{
		{sender: "System", text: "Welcome! Focus is on Command input. Type commands like 'quit', '?', or Tab to switch."},
	}

	// Any logging here should use app.GetDebugLogger()
	if app != nil && app.GetDebugLogger() != nil {
		app.GetDebugLogger().Println("TUI model initialized.")
	}

	return model{
		app:             app, // Store the interface
		commandInput:    cmdInput,
		promptInput:     promptInput,
		viewport:        vp,
		spinner:         sp,
		help:            h,
		keyMap:          defaultKeyMap,
		messages:        initialMessages,
		sender:          "You",
		aiModelName:     modelName, // Store the retrieved name
		focusedInput:    focusCommand,
		helpVisible:     false,
		ready:           false,
		isSyncing:       false,
		currentActivity: "",
	}
}
