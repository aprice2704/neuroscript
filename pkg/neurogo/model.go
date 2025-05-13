// NeuroScript Version: 0.3.0
// File version: 0.1.12
// Use tui.AppAccess interface.
// filename: pkg/neurogo/tui/model.go
// nlines: 225
// risk_rating: HIGH
package neurogo

import (
	"context"
	"fmt"
	"path/filepath"

	// "strings" // Already in use by FormatWMStatusView if it were here

	// "github.com/aprice2704/neuroscript/pkg/core" // core types used via interfaces
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Styles --- (styles remain the same)
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

	localOutputBlurredStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("51"))
	aiOutputBlurredStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("69"))
	localOutputFocusedStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("205"))
	aiOutputFocusedStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("205"))

	localInputFocusedStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("205"))
	localInputBlurredStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("240"))
	aiInputFocusedStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("205"))
	aiInputBlurredStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("240"))

	paneTitleStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1)
)

type message struct {
	sender string
	text   string
}

const (
	focusLocalInput = iota
	focusAIInput
	focusLocalOutput
	focusAIOutput
	totalFocusPanes
)

const (
	localOutputModeScript = iota
	localOutputModeWMStatus
	totalLocalOutputModes
)

type keyMap struct {
	Quit, Help, Tab, ShiftTab, ScrollUp, ScrollDown, ScrollLeft, ScrollRight, PageUp, PageDown, CycleLocalOutput key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.ShiftTab, k.CycleLocalOutput, k.Help, k.Quit}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.ShiftTab, k.CycleLocalOutput, k.Help, k.Quit},
		{k.ScrollUp, k.ScrollDown, k.ScrollLeft, k.ScrollRight, k.PageUp, k.PageDown},
	}
}

var defaultKeyMap = keyMap{
	Quit:             key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	Help:             key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Tab:              key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next pane")),
	ShiftTab:         key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("s+tab", "prev pane")),
	CycleLocalOutput: key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("ctrl+b", "cycle local view")),
	ScrollUp:         key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "scroll up")),
	ScrollDown:       key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "scroll down")),
	ScrollLeft:       key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "L")),
	ScrollRight:      key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "R")),
	PageUp:           key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "pgup")),
	PageDown:         key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "pgdn")),
}

type model struct {
	app                                                     *App
	localOutput, aiOutput                                   viewport.Model
	localInput, aiInput                                     textarea.Model
	spinner                                                 spinner.Model
	help                                                    help.Model
	keyMap                                                  keyMap
	teaProgram                                              *tea.Program
	messages                                                []message
	emittedLines                                            []string
	initialScriptToRun                                      string
	initialScriptRunning                                    bool
	sender                                                  string
	lastError                                               error
	isWaitingForAI, isSyncing, quitting, ready, helpVisible bool
	currentActivity, patchStatus                            string
	focusIndex                                              int
	aiModelName                                             string
	localFileCount, apiFileCount, syncUploads, syncDeletes  int
	width, height                                           int

	localOutputDisplayMode int
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func newModel(app *App, initialScriptPath string) model { // app is now tui.AppAccess
	plainInternalStyle := lipgloss.NewStyle()
	focusedPromptTextSyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredPromptTextSyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	placeholderTextSyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

	li := textarea.New()
	li.Placeholder = "Local Input (/cmd or script path)"
	li.Focus()
	li.Prompt = "$ "
	li.CharLimit = 0
	li.FocusedStyle.Base = localInputFocusedStyle
	li.BlurredStyle.Base = localInputBlurredStyle
	li.FocusedStyle.Prompt = focusedPromptTextSyle
	li.BlurredStyle.Prompt = blurredPromptTextSyle
	li.FocusedStyle.Text = plainInternalStyle
	li.BlurredStyle.Text = plainInternalStyle
	li.FocusedStyle.Placeholder = placeholderTextSyle
	li.BlurredStyle.Placeholder = placeholderTextSyle
	li.FocusedStyle.CursorLine = plainInternalStyle
	li.BlurredStyle.CursorLine = plainInternalStyle
	li.FocusedStyle.CursorLineNumber = plainInternalStyle
	li.BlurredStyle.CursorLineNumber = plainInternalStyle
	li.FocusedStyle.EndOfBuffer = plainInternalStyle
	li.BlurredStyle.EndOfBuffer = plainInternalStyle
	li.KeyMap.InsertNewline.SetEnabled(false)

	ai := textarea.New()
	ai.Placeholder = "AI Input (Enter prompt)"
	ai.Prompt = "> "
	ai.CharLimit = 0
	ai.FocusedStyle.Base = aiInputFocusedStyle
	ai.BlurredStyle.Base = aiInputBlurredStyle
	ai.FocusedStyle.Prompt = focusedPromptTextSyle
	ai.BlurredStyle.Prompt = blurredPromptTextSyle
	ai.FocusedStyle.Text = plainInternalStyle
	ai.BlurredStyle.Text = plainInternalStyle
	ai.FocusedStyle.Placeholder = placeholderTextSyle
	ai.BlurredStyle.Placeholder = placeholderTextSyle
	ai.FocusedStyle.CursorLine = plainInternalStyle
	ai.BlurredStyle.CursorLine = plainInternalStyle
	ai.FocusedStyle.CursorLineNumber = plainInternalStyle
	ai.BlurredStyle.CursorLineNumber = plainInternalStyle
	ai.FocusedStyle.EndOfBuffer = plainInternalStyle
	ai.BlurredStyle.EndOfBuffer = plainInternalStyle
	ai.KeyMap.InsertNewline.SetEnabled(true)

	loVP := viewport.New(10, 10)
	loVP.SetContent("")
	loVP.Style = plainInternalStyle

	aiVP := viewport.New(10, 10)
	aiVP.SetContent("Welcome to NeuroScript TUI!")
	aiVP.Style = plainInternalStyle

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	h := help.New()
	h.ShowAll = false
	modelName := "unknown"
	if app != nil { // app is tui.AppAccess
		modelName = app.GetModelName()
	}
	initialMessages := []message{{sender: "System", text: "Focus: Local Input. Tab/Shift+Tab. ? for help."}}
	if app != nil && app.GetLogger() != nil {
		app.GetLogger().Debug("TUI model initialized.")
	}
	scriptIsPending := initialScriptPath != "" && app != nil
	return model{
		app:                    app, // app is tui.AppAccess
		initialScriptToRun:     initialScriptPath,
		initialScriptRunning:   scriptIsPending,
		currentActivity:        If(scriptIsPending, fmt.Sprintf("Executing: %s...", filepath.Base(initialScriptPath)), "").(string),
		localInput:             li,
		aiInput:                ai,
		localOutput:            loVP,
		aiOutput:               aiVP,
		spinner:                sp,
		help:                   h,
		keyMap:                 defaultKeyMap,
		messages:               initialMessages,
		emittedLines:           []string{},
		sender:                 "You",
		aiModelName:            modelName,
		focusIndex:             focusLocalInput,
		helpVisible:            false,
		ready:                  false,
		isSyncing:              false,
		patchStatus:            "",
		localOutputDisplayMode: localOutputModeScript,
	}
}

func (m *model) SetTeaProgram(p *tea.Program) { m.teaProgram = p }

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, textarea.Blink)
	if m.initialScriptToRun != "" && m.app != nil { // app is tui.AppAccess
		scriptCmd := func() tea.Msg {
			if m.app.GetLogger() != nil {
				m.app.GetLogger().Debug("Executing initial TUI script...", "path", m.initialScriptToRun)
			}
			ctxToUse := context.Background()
			if m.app.Context() != nil {
				ctxToUse = m.app.Context()
			}
			err := m.app.ExecuteScriptFile(ctxToUse, m.initialScriptToRun) // Call on tui.AppAccess
			return initialScriptDoneMsg{Path: m.initialScriptToRun, Err: err}
		}
		cmds = append(cmds, scriptCmd)
		if m.initialScriptRunning {
			cmds = append(cmds, m.spinner.Tick)
		}
	}
	return tea.Batch(cmds...)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
