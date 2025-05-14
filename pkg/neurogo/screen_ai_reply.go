// NeuroScript Version: 0.3.0
// File version: 0.0.4 // Removed unused 'refreshed' variable. Ensured refreshViewMsg logic is standard.
// filename: pkg/neurogo/screen_ai_reply.go
// nlines: 100
// risk_rating: LOW
package neurogo

import (
	"strings"
	// "time" // Not directly used in this version of the file

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AIReplyScreen struct {
	app      *App
	viewport viewport.Model
	width    int
	height   int
}

func NewAIReplyScreen(app *App, width, height int) *AIReplyScreen {
	vp := viewport.New(width, height)
	// Initial content will be set in Init
	return &AIReplyScreen{
		app:      app,
		viewport: vp,
		width:    width,
		height:   height,
	}
}

func (s *AIReplyScreen) Init(app *App) tea.Cmd {
	s.app = app
	s.updateViewportContent()
	s.viewport.GotoBottom()
	return nil
}

func (s *AIReplyScreen) updateViewportContent() {
	if s.app == nil || s.app.GetTUImodel() == nil {
		s.viewport.SetContent(errorStyle.Render("AIReplyScreen: App/Model not available.")) // Use errorStyle
		if s.app != nil && s.app.GetLogger() != nil {
			s.app.GetLogger().Warn("AIReplyScreen: updateViewportContent - GetTUImodel is nil or app is nil")
		}
		return
	}

	var content strings.Builder
	// Access systemMessages via app.GetTUImodel()
	// Make a copy to avoid issues if the underlying slice is modified elsewhere during iteration.
	// This is good practice, though less critical if systemMessages is append-only during a single Update cycle.
	tuiModel := s.app.GetTUImodel()
	messagesToRender := make([]message, len(tuiModel.systemMessages))
	copy(messagesToRender, tuiModel.systemMessages)

	for _, msg := range messagesToRender {
		var style lipgloss.Style
		senderPrefix := msg.sender

		switch strings.ToLower(msg.sender) {
		case "you", "user":
			style = userStyle
		case "ai", "model":
			style = aiStyle
		case "system":
			style = systemStyle
		case "emit": // This screen might not show emits, but if it does via systemMessages
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("172"))
			senderPrefix = "SCRIPT"
		case "error":
			style = errorStyle
		case "tool result":
			style = toolResultStyle
		default:
			style = systemStyle
		}

		if msg.sender != "" {
			content.WriteString(style.Render(senderPrefix + ": "))
		}
		// If msg.text is already styled (e.g. an error message), this will re-style it.
		// Consider if msg.text should be rendered as-is if already styled.
		content.WriteString(msg.text)
		content.WriteString("\n")
	}
	s.viewport.SetContent(content.String())
}

func (s *AIReplyScreen) Update(msg tea.Msg, app *App) (Screen, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	s.app = app // Ensure app context is current

	// This screen primarily displays messages from m.systemMessages.
	// It should update its content when those messages change.
	// The main model's Update loop adds to systemMessages.
	// This screen will re-render its content on Focus, SetSize, or specific refresh messages.

	switch msgTyped := msg.(type) {
	// Messages that imply systemMessages might have changed in the main model
	case initialScriptDoneMsg, syncCompleteMsg, errMsg, aiResponseMsg, closeScreenMsg, sendAIChatMsg, updateStatusBarMsg:
		s.updateViewportContent()
		s.viewport.GotoBottom()

	case refreshViewMsg: // msgTyped is now of type refreshViewMsg
		if msgTyped.ScreenName == "" || msgTyped.ScreenName == s.Name() {
			s.updateViewportContent()
			s.viewport.GotoBottom()
		}
		// tea.KeyMsg for scrolling is handled by the viewport update below
	}

	// Default viewport update for scrolling, mouse events, etc.
	s.viewport, cmd = s.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return s, tea.Batch(cmds...)
}

func (s *AIReplyScreen) View(width, height int) string {
	// Content is updated in Update, Focus, or SetSize. View just renders.
	return s.viewport.View()
}

func (s *AIReplyScreen) Name() string {
	return "System/AI Log"
}

func (s *AIReplyScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.viewport.Width = width
	s.viewport.Height = height
	s.updateViewportContent() // Refresh content on resize as wrapping might change
}

func (s *AIReplyScreen) GetInputBubble() *textarea.Model { return nil }
func (s *AIReplyScreen) HandleSubmit(app *App) tea.Cmd   { return nil }

func (s *AIReplyScreen) Focus(app *App) tea.Cmd {
	s.app = app
	s.updateViewportContent() // Refresh content when focused
	s.viewport.GotoBottom()   // Scroll to bottom when focused
	return nil
}

func (s *AIReplyScreen) Blur(app *App) tea.Cmd {
	return nil
}
