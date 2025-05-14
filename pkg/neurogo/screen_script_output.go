// NeuroScript Version: 0.3.0
// File version: 0.0.3
// Corrected access to main model's emittedLines via app.GetTUImodel().
// filename: pkg/neurogo/screen_script_output.go
// nlines: 95
// risk_rating: LOW
package neurogo

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type ScriptOutputScreen struct {
	app      *App
	viewport viewport.Model
	width    int
	height   int
}

func NewScriptOutputScreen(app *App, width, height int) *ScriptOutputScreen {
	vp := viewport.New(width, height)
	return &ScriptOutputScreen{
		app:      app,
		viewport: vp,
		width:    width,
		height:   height,
	}
}

func (s *ScriptOutputScreen) Init(app *App) tea.Cmd {
	s.app = app
	s.updateViewportContent()
	s.viewport.GotoBottom()
	return nil
}

func (s *ScriptOutputScreen) updateViewportContent() {
	if s.app != nil && s.app.GetTUImodel() != nil {
		// Access emittedLines via app.GetTUImodel()
		// Create a copy for safety if the underlying slice might change during join
		linesCopy := make([]string, len(s.app.GetTUImodel().emittedLines))
		copy(linesCopy, s.app.GetTUImodel().emittedLines)
		s.viewport.SetContent(strings.Join(linesCopy, "\n"))
	} else {
		s.viewport.SetContent("Error: ScriptOutputScreen cannot access TUI model data.")
		if s.app != nil && s.app.GetLogger() != nil {
			s.app.GetLogger().Warn("ScriptOutputScreen: updateViewportContent - GetTUImodel is nil or app is nil")
		}
	}
}

func (s *ScriptOutputScreen) Update(msg tea.Msg, app *App) (Screen, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	s.app = app

	switch msg := msg.(type) {
	case scriptEmitMsg:
		// The main model already appends to m.emittedLines.
		// This screen just needs to re-render that data.
		s.updateViewportContent() // Re-fetch from app.GetTUImodel().emittedLines
		s.viewport.GotoBottom()
		return s, nil

	case refreshViewMsg:
		if msg.ScreenName == "" || msg.ScreenName == s.Name() {
			s.updateViewportContent()
			s.viewport.GotoBottom()
		}
		return s, nil
	}

	s.viewport, cmd = s.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return s, tea.Batch(cmds...)
}

func (s *ScriptOutputScreen) View(width, height int) string {
	return s.viewport.View()
}

func (s *ScriptOutputScreen) Name() string {
	return "Script Output"
}

func (s *ScriptOutputScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.viewport.Width = width
	s.viewport.Height = height
	s.updateViewportContent() // Refresh content on resize
}

func (s *ScriptOutputScreen) GetInputBubble() *textarea.Model {
	return nil
}

func (s *ScriptOutputScreen) HandleSubmit(app *App) tea.Cmd {
	return nil
}

func (s *ScriptOutputScreen) Focus(app *App) tea.Cmd {
	s.app = app
	s.updateViewportContent()
	s.viewport.GotoBottom()
	return nil
}

func (s *ScriptOutputScreen) Blur(app *App) tea.Cmd {
	return nil
}
