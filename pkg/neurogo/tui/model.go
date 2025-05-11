// NeuroScript Version: 0.3.0
// File version: 0.0.5 // Handle initial script path, set distinct full borders, manage initial script running state.
// filename: pkg/neurogo/tui/model.go
// nlines: 165 // Approximate
// risk_rating: MEDIUM
package tui

import (
	"context"       // For running the initial script
	"fmt"           // For initial script activity message
	"path/filepath" // For getting base name of script path

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Styles ---
var (
	inactiveStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	userStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	aiStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
	aiToolCallStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Italic(true)
	sysToolCallStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Italic(true)
	toolResultStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true)
	patchStatusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	systemStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	statusBarSyle    = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("250")).Padding(0, 1)

	// Styles with distinct, full borders for each component
	viewportStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, true, true, true).BorderForeground(lipgloss.Color("69"))   // Conversation Viewport (Blue border)
	emitLogStyle        = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, true, true, true).BorderForeground(lipgloss.Color("51"))   // Emit Log Viewport (Cyan border)
	focusedCommandStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true, true, true, true).BorderForeground(lipgloss.Color("205")) // Focused Command (Pink border)
	blurredCommandStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, true, true, true).BorderForeground(lipgloss.Color("240"))  // Blurred Command (Gray border)
	focusedPromptStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true, true, true, true).BorderForeground(lipgloss.Color("205")) // Focused Prompt (Pink border)
	blurredPromptStyle  = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, true, true, true).BorderForeground(lipgloss.Color("240"))  // Blurred Prompt (Gray border)
)

// --- message struct ---
type message struct { // For main conversation viewport
	sender string
	text   string
}

// --- keyMap struct and implementation for help.KeyMap ---
type keyMap struct {
	Quit key.Binding
	Help key.Binding
	Tab  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding  { return []key.Binding{k.Tab, k.Help, k.Quit} }
func (k keyMap) FullHelp() [][]key.Binding { return [][]key.Binding{{k.Tab, k.Help, k.Quit}} }

var defaultKeyMap = keyMap{
	Quit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	Help: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
	Tab:  key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch input")),
}

const (
	focusCommand = 0
	focusPrompt  = 1
)

// --- model struct ---
type model struct {
	app AppAccess

	viewport        viewport.Model // Main conversation
	emitLogViewport viewport.Model // For EMIT statements
	commandInput    textarea.Model
	promptInput     textarea.Model
	spinner         spinner.Model
	help            help.Model
	keyMap          keyMap
	teaProgram      *tea.Program // Can be set by tui.Start

	messages             []message
	emittedLines         []string
	initialScriptToRun   string // Path of the script to run on startup, if any
	initialScriptRunning bool   // True if the initial script is currently executing
	sender               string
	lastError            error
	isWaitingForAI       bool
	currentActivity      string
	isSyncing            bool
	patchStatus          string
	quitting             bool
	ready                bool
	helpVisible          bool
	focusedInput         int

	aiModelName    string
	localFileCount int
	apiFileCount   int
	syncUploads    int
	syncDeletes    int

	width  int
	height int
}

// If is a simple helper for conditional assignment.
func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

// newModel constructor
func newModel(app AppAccess, initialScriptPath string) model {
	cmdInput := textarea.New()
	cmdInput.Placeholder = "/cmd or script path"
	cmdInput.Focus() // Command input focused by default
	cmdInput.Prompt = "$ "
	cmdInput.CharLimit = 0
	cmdInput.SetHeight(1)
	cmdInput.FocusedStyle.Base = focusedCommandStyle
	cmdInput.BlurredStyle.Base = blurredCommandStyle
	cmdInput.KeyMap.InsertNewline.SetEnabled(false)

	promptInput := textarea.New()
	promptInput.Placeholder = "Enter prompt for AI..."
	promptInput.Prompt = "> "
	promptInput.CharLimit = 0
	promptInput.SetHeight(3)
	promptInput.FocusedStyle.Base = focusedPromptStyle
	promptInput.BlurredStyle.Base = blurredPromptStyle
	promptInput.KeyMap.InsertNewline.SetEnabled(true)

	vp := viewport.New(10, 5)
	vp.SetContent("Welcome to NeuroScript TUI!")
	vp.Style = viewportStyle // Apply border style

	emitVP := viewport.New(5, 10)
	emitVP.SetContent("--- Script EMIT Log ---")
	emitVP.Style = emitLogStyle // Apply border style

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	h := help.New()
	h.ShowAll = false

	modelName := "unknown"
	if app != nil {
		modelName = app.GetModelName()
	}

	initialMessages := []message{
		{sender: "System", text: "Focus: Command. Type '/help', Tab to switch, or prompt."},
	}

	if app != nil && app.GetLogger() != nil {
		app.GetLogger().Debug("TUI model initialized.")
	}

	scriptIsPending := initialScriptPath != "" && app != nil

	return model{
		app:                  app,
		initialScriptToRun:   initialScriptPath,
		initialScriptRunning: scriptIsPending, // Set true if script is provided
		currentActivity:      If(scriptIsPending, fmt.Sprintf("Executing: %s...", filepath.Base(initialScriptPath)), "").(string),
		commandInput:         cmdInput,
		promptInput:          promptInput,
		viewport:             vp,
		emitLogViewport:      emitVP,
		spinner:              sp,
		help:                 h,
		keyMap:               defaultKeyMap,
		messages:             initialMessages,
		emittedLines:         []string{},
		sender:               "You",
		aiModelName:          modelName,
		focusedInput:         focusCommand,
		helpVisible:          false,
		ready:                false,
		isSyncing:            false,
		patchStatus:          "",
	}
}

// SetTeaProgram allows the main TUI function to set the tea.Program instance
func (m *model) SetTeaProgram(p *tea.Program) {
	m.teaProgram = p
}

// Init is called by Bubble Tea when the model is first started.
func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, textarea.Blink)

	if m.initialScriptToRun != "" && m.app != nil {
		// currentActivity and initialScriptRunning should be set in newModel if a script is pending.
		// This command just runs it.
		scriptCmd := func() tea.Msg {
			if m.app.GetLogger() != nil {
				m.app.GetLogger().Debug("Executing initial TUI script (cmd from Init)...", "path", m.initialScriptToRun)
			}
			// Script execution happens here. EMITs should be routed to TUIEmitWriter.
			err := m.app.ExecuteScriptFile(context.Background(), m.initialScriptToRun)
			return initialScriptDoneMsg{Path: m.initialScriptToRun, Err: err}
		}
		cmds = append(cmds, scriptCmd)
		if m.initialScriptRunning { // If set true in newModel
			cmds = append(cmds, m.spinner.Tick) // Start spinner immediately
		}
	}
	return tea.Batch(cmds...)
}
