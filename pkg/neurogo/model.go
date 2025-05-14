// NeuroScript Version: 0.3.0
// File version: 0.2.3
// Added maxEmitBufferLines constant and refined App.GetTUImodel access.
// filename: pkg/neurogo/model.go
// nlines: 300 // Approximate
// risk_rating: HIGH
package neurogo

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"

	// "github.com/charmbracelet/bubbles/viewport" // Viewports now managed by individual screens
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxEmitBufferLines = 1000 // Buffer size for script emits

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

	inputPaneFocusedStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("205"))
	inputPaneBlurredStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("240"))
	paneTitleStyle           = lipgloss.NewStyle().Bold(true).Padding(0, 1).Background(lipgloss.Color("237")).Foreground(lipgloss.Color("252"))
	screenPaneContainerStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(lipgloss.Color("240"))
)

type message struct {
	sender string
	text   string
}

type FocusTarget int

const (
	FocusLeftInput FocusTarget = iota
	FocusRightInput
	FocusLeftPane
	FocusRightPane
	totalFocusTargets
)

type keyMap struct {
	Quit, Help, Tab, ShiftTab, ScrollUp, ScrollDown, ScrollLeft, ScrollRight, PageUp, PageDown, CycleLeftScreen, CycleRightScreen key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.ShiftTab, k.CycleLeftScreen, k.CycleRightScreen, k.Help, k.Quit}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.ShiftTab, k.CycleLeftScreen, k.CycleRightScreen, k.Help, k.Quit},
		{k.ScrollUp, k.ScrollDown, k.ScrollLeft, k.ScrollRight, k.PageUp, k.PageDown},
	}
}

var defaultKeyMap = keyMap{
	Quit:             key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	Help:             key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Tab:              key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next focus")),
	ShiftTab:         key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("s+tab", "prev focus")),
	CycleLeftScreen:  key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("ctrl+b", "cycle left screen")),
	CycleRightScreen: key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "cycle right screen")),
	ScrollUp:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "scroll up")),
	ScrollDown:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "scroll down")),
	ScrollLeft:       key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "scroll L")),
	ScrollRight:      key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "scroll R")),
	PageUp:           key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "pgup")),
	PageDown:         key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "pgdn")),
}

type model struct {
	app        *App
	keyMap     keyMap
	teaProgram *tea.Program
	spinner    spinner.Model
	help       help.Model

	leftScreens           []Screen
	rightScreens          []Screen
	currentLeftScreenIdx  int
	currentRightScreenIdx int

	leftInputArea  textarea.Model
	rightInputArea textarea.Model

	systemMessages             []message
	emittedLines               []string
	initialScriptToRun         string
	initialScriptRunning       bool
	lastError                  error
	isWaitingForAI             bool
	isSyncing                  bool
	quitting                   bool
	ready                      bool
	helpVisible                bool
	currentActivity            string
	patchStatus                string
	focusTarget                FocusTarget
	width, height              int
	lastDisplayedWMDefinitions []*core.AIWorkerDefinition
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func newModel(app *App, initialScriptPath string) model {
	inputDefaultStyle := lipgloss.NewStyle()
	inputFocusedPromptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	inputBlurredPromptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	inputPlaceholderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

	leftTA := textarea.New()
	leftTA.Placeholder = "System Commands (e.g., //chat)"
	leftTA.Prompt = "SYS $ "
	leftTA.CharLimit = 0
	leftTA.FocusedStyle.Base = inputPaneFocusedStyle
	leftTA.BlurredStyle.Base = inputPaneBlurredStyle
	leftTA.FocusedStyle.Prompt = inputFocusedPromptStyle
	leftTA.BlurredStyle.Prompt = inputBlurredPromptStyle
	leftTA.FocusedStyle.Text = inputDefaultStyle
	leftTA.BlurredStyle.Text = inputDefaultStyle
	leftTA.FocusedStyle.Placeholder = inputPlaceholderStyle
	leftTA.BlurredStyle.Placeholder = inputPlaceholderStyle
	leftTA.KeyMap.InsertNewline.SetEnabled(false)

	rightTA := textarea.New()
	rightTA.Placeholder = "Contextual Input"
	rightTA.Prompt = "INPUT > "
	rightTA.CharLimit = 0
	rightTA.FocusedStyle.Base = inputPaneFocusedStyle
	rightTA.BlurredStyle.Base = inputPaneBlurredStyle
	rightTA.FocusedStyle.Prompt = inputFocusedPromptStyle
	rightTA.BlurredStyle.Prompt = inputBlurredPromptStyle
	rightTA.FocusedStyle.Text = inputDefaultStyle
	rightTA.BlurredStyle.Text = inputDefaultStyle
	rightTA.FocusedStyle.Placeholder = inputPlaceholderStyle
	rightTA.BlurredStyle.Placeholder = inputPlaceholderStyle
	rightTA.KeyMap.InsertNewline.SetEnabled(true)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	h := help.New()
	h.ShowAll = false

	tempWidth, tempHeight := 80, 24 // Initial placeholder sizes, SetSize will correct

	// Pass app to screen constructors, they might need it immediately or store for later via Init/Update
	scriptOutputScreen := NewScriptOutputScreen(app, tempWidth/2, tempHeight-8)
	wmStatusScreen := NewWMStatusScreen(app, tempWidth/2, tempHeight-8)
	aiReplyScreen := NewAIReplyScreen(app, tempWidth/2, tempHeight-8)

	leftScreens := []Screen{scriptOutputScreen, wmStatusScreen}
	rightScreens := []Screen{aiReplyScreen}

	m := model{
		app:                        app, // app field in model struct
		keyMap:                     defaultKeyMap,
		spinner:                    sp,
		help:                       h,
		leftScreens:                leftScreens,
		rightScreens:               rightScreens,
		currentLeftScreenIdx:       0,
		currentRightScreenIdx:      0,
		leftInputArea:              leftTA,
		rightInputArea:             rightTA,
		systemMessages:             []message{{sender: "System", text: "Welcome! Focus: Left Input. Ctrl+B/N. ? for help."}},
		emittedLines:               make([]string, 0, maxEmitBufferLines),
		initialScriptToRun:         initialScriptPath,
		initialScriptRunning:       (initialScriptPath != "" && app != nil),
		currentActivity:            If(initialScriptPath != "" && app != nil, fmt.Sprintf("Executing: %s...", filepath.Base(initialScriptPath)), "").(string),
		focusTarget:                FocusLeftInput,
		helpVisible:                false,
		ready:                      false,
		isSyncing:                  false,
		patchStatus:                "",
		lastDisplayedWMDefinitions: make([]*core.AIWorkerDefinition, 0),
	}
	// The App instance needs a way to reference this model if screens are to call back via app.GetTUImodel()
	// This is typically done after newModel returns, e.g., in tui.Start()

	m.leftInputArea.Focus() // Default focus on app start

	if app != nil && app.GetLogger() != nil {
		app.GetLogger().Debug("TUI model (Screen architecture) initialized.")
	}
	return m
}

func (m *model) SetTeaProgram(p *tea.Program) { m.teaProgram = p }

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd

	if len(m.leftScreens) > 0 && m.currentLeftScreenIdx < len(m.leftScreens) {
		if screenCmd := m.leftScreens[m.currentLeftScreenIdx].Init(m.app); screenCmd != nil {
			cmds = append(cmds, screenCmd)
		}
	}
	if len(m.rightScreens) > 0 && m.currentRightScreenIdx < len(m.rightScreens) {
		if screenCmd := m.rightScreens[m.currentRightScreenIdx].Init(m.app); screenCmd != nil {
			cmds = append(cmds, screenCmd)
		}
	}

	cmds = append(cmds, m.updateFocusStates()) // Set initial focus correctly

	if m.initialScriptRunning {
		scriptCmd := func() tea.Msg {
			if m.app.GetLogger() != nil {
				m.app.GetLogger().Debug("Executing initial TUI script...", "path", m.initialScriptToRun)
			}
			ctxToUse := context.Background()
			if m.app != nil && m.app.Context() != nil { // Check m.app for nil
				ctxToUse = m.app.Context()
			}
			// Check m.app before calling ExecuteScriptFile
			if m.app != nil {
				err := m.app.ExecuteScriptFile(ctxToUse, m.initialScriptToRun)
				return initialScriptDoneMsg{Path: m.initialScriptToRun, Err: err}
			}
			return initialScriptDoneMsg{Path: m.initialScriptToRun, Err: fmt.Errorf("app context not available for initial script")}
		}
		cmds = append(cmds, scriptCmd)
		cmds = append(cmds, m.spinner.Tick)
	}
	return tea.Batch(cmds...)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *model) getActiveLeftScreen() Screen {
	if len(m.leftScreens) > 0 && m.currentLeftScreenIdx >= 0 && m.currentLeftScreenIdx < len(m.leftScreens) {
		return m.leftScreens[m.currentLeftScreenIdx]
	}
	return nil
}

func (m *model) getActiveRightScreen() Screen {
	if len(m.rightScreens) > 0 && m.currentRightScreenIdx >= 0 && m.currentRightScreenIdx < len(m.rightScreens) {
		return m.rightScreens[m.currentRightScreenIdx]
	}
	return nil
}
