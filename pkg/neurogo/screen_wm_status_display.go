// NeuroScript Version: 0.3.0
// File version: 0.0.3 // Used GetTUImodel() to access model from app.
// Functional WMStatusScreen using updated FormatWMStatusView.
// filename: pkg/neurogo/screen_wm_status_display.go
// nlines: 90
// risk_rating: MEDIUM
package neurogo

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type WMStatusScreen struct {
	app      *App
	viewport viewport.Model
	width    int
	height   int
}

func NewWMStatusScreen(app *App, width, height int) *WMStatusScreen {
	vp := viewport.New(width, height)
	return &WMStatusScreen{
		app:      app,
		viewport: vp,
		width:    width,
		height:   height,
	}
}

func (s *WMStatusScreen) Init(app *App) tea.Cmd {
	s.app = app
	s.updateContent()
	return nil
}

func (s *WMStatusScreen) updateContent() {
	if s.app == nil {
		s.viewport.SetContent(wmErrorStyle.Render("WMStatusScreen: App reference not available."))
		return
	}
	definitions, formattedView := FormatWMStatusView(s.app)

	tuiModel := s.app.GetTUImodel() // Use getter
	if tuiModel != nil {            // Check if nil
		tuiModel.lastDisplayedWMDefinitions = definitions
		if s.app.GetLogger() != nil {
			s.app.GetLogger().Debug("WMStatusScreen updated main model's lastDisplayedWMDefinitions", "count", len(definitions))
		}
	}

	s.viewport.SetContent(formattedView)
	s.viewport.GotoTop()
}

func (s *WMStatusScreen) Update(msg tea.Msg, app *App) (Screen, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	s.app = app

	switch msg := msg.(type) {
	case tea.KeyMsg:
		s.viewport, cmd = s.viewport.Update(msg)
		cmds = append(cmds, cmd)
		return s, tea.Batch(cmds...)

	case refreshViewMsg:
		s.updateContent()
		return s, nil
	case tea.WindowSizeMsg:
		break
	}

	s.viewport, cmd = s.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return s, tea.Batch(cmds...)
}

func (s *WMStatusScreen) View(width, height int) string {
	return s.viewport.View()
}

func (s *WMStatusScreen) Name() string {
	return "Worker Manager Status"
}

func (s *WMStatusScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.viewport.Width = width
	s.viewport.Height = height
	s.updateContent()
}

func (s *WMStatusScreen) GetInputBubble() *textarea.Model {
	return nil
}

func (s *WMStatusScreen) HandleSubmit(app *App) tea.Cmd {
	return nil
}

func (s *WMStatusScreen) Focus(app *App) tea.Cmd {
	s.app = app
	s.updateContent()
	return nil
}

func (s *WMStatusScreen) Blur(app *App) tea.Cmd {
	return nil
}
