// NeuroScript Version: 0.3.0
// File version: 0.0.3 // Adjusted to compiler feedback on core func signatures
// filename: pkg/neurogo/screen_chat.go
// nlines: 205 // Approximate
// risk_rating: HIGH
package neurogo

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/adapters" // For NewNoOpLogger
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging" // For logging.Logger type
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/generative-ai-go/genai"
)

type ChatScreen struct {
	app                 *App
	viewport            viewport.Model
	inputArea           textarea.Model
	conversationManager *core.ConversationManager
	width               int
	height              int
	workerDefinitionID  string
	workerInstanceID    string
	screenName          string
}

func NewChatScreen(app *App, width, height int, definitionID, instanceID, screenName string) *ChatScreen {
	vp := viewport.New(width, height-inputAreaDefaultVisibleLines-1)
	vp.Style = lipgloss.NewStyle().Padding(0, 1)

	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 0
	ta.SetWidth(width)
	ta.SetHeight(inputAreaDefaultVisibleLines)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.KeyMap.InsertNewline.SetEnabled(true)

	var cmLogger logging.Logger // Declare as interface type
	if app != nil && app.GetLogger() != nil {
		cmLogger = app.GetLogger()
	} else {
		cmLogger = adapters.NewNoOpLogger() // Fallback if app or its logger is nil
	}

	// Adjusting call based on compiler error: "want (logging.Logger)"
	// This assumes NewConversationManager now only takes a logger.
	// DefinitionID and InstanceID would need to be set differently if still needed by ConversationManager.
	cm := core.NewConversationManager(cmLogger)
	// If definitionID and instanceID are still relevant for ConversationManager,
	// they might need to be set via methods like:
	// cm.SetDefinitionID(definitionID)
	// cm.SetInstanceID(instanceID)
	// For now, these are not passed to constructor.

	return &ChatScreen{
		app:                 app,
		viewport:            vp,
		inputArea:           ta,
		conversationManager: cm,
		width:               width,
		height:              height,
		workerDefinitionID:  definitionID,
		workerInstanceID:    instanceID,
		screenName:          screenName,
	}
}

func (s *ChatScreen) Init(app *App) tea.Cmd {
	s.app = app
	// s.conversationManager.AddModelMessage(fmt.Sprintf("Chatting with %s. Type /end to close.", s.screenName))
	// Let's use AddModelResponse for consistency if it's the primary way to add model turns.
	// Or ensure ConversationManager can handle simple string messages for system.
	// For now, assuming AddModelMessage is available and correct.
	if s.conversationManager != nil {
		s.conversationManager.AddModelMessage(fmt.Sprintf("Chatting with %s. Type /end to close.", s.screenName))
	}
	s.renderConversation()
	return textarea.Blink
}

func (s *ChatScreen) Update(msg tea.Msg, app *App) (Screen, tea.Cmd) {
	var vpCmd, taCmd tea.Cmd // Removed batchCmd
	var cmds []tea.Cmd
	s.app = app

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s.inputArea.Focused() {
			s.inputArea, taCmd = s.inputArea.Update(msg)
			cmds = append(cmds, taCmd)
		} else {
			s.viewport, vpCmd = s.viewport.Update(msg)
			cmds = append(cmds, vpCmd)
		}
		return s, tea.Batch(cmds...)

	case aiResponseMsg:
		if msg.TargetScreenName != s.Name() {
			return s, nil
		}
		if msg.Err != nil {
			errMsgText := fmt.Sprintf("AI Error: %v", msg.Err)
			s.conversationManager.AddModelMessage(errMsgText)
			if app != nil && app.GetLogger() != nil {
				app.GetLogger().Error("ChatScreen received AI error", "screen", s.Name(), "error", msg.Err)
			}
		} else if msg.ResponseCandidate != nil {
			aiText := msg.ResponseCandidate.Content // Use Content field
			if aiText == "" && len(msg.ResponseCandidate.ToolCalls) > 0 {
				aiText = fmt.Sprintf("[Requesting tool calls: %d]", len(msg.ResponseCandidate.ToolCalls))
			}
			s.conversationManager.AddModelMessage(aiText)

			// FinishReason is not on core.ConversationTurn.
			// If needed, this information must be sourced differently or added to ConversationTurn.
			// finishReason := msg.ResponseCandidate.FinishReason // This field does not exist
			// if finishReason != "" && finishReason != "STOP" {
			// 	s.conversationManager.AddModelMessage(fmt.Sprintf("[Finish Reason: %s]", finishReason))
			// }
		} else {
			s.conversationManager.AddModelMessage("[AI provided no response content]")
		}
		s.renderConversation()
		s.inputArea.Focus()
		return s, textarea.Blink

	case systemMessageToChatScreen:
		if msg.TargetScreenName != s.Name() {
			return s, nil
		}
		s.conversationManager.AddModelMessage(fmt.Sprintf("[System]: %s", msg.Content))
		s.renderConversation()
	}

	s.viewport, vpCmd = s.viewport.Update(msg)
	s.inputArea, taCmd = s.inputArea.Update(msg)
	cmds = append(cmds, vpCmd, taCmd)

	return s, tea.Batch(cmds...)
}

func (s *ChatScreen) View(width, height int) string {
	return lipgloss.JoinVertical(lipgloss.Left,
		s.viewport.View(),
		s.inputArea.View(),
	)
}

func (s *ChatScreen) Name() string {
	return s.screenName
}

func (s *ChatScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
	inputAreaHeight := s.inputArea.LineCount()
	if inputAreaHeight < inputAreaDefaultVisibleLines {
		inputAreaHeight = inputAreaDefaultVisibleLines
	}
	if inputAreaHeight > height/2 {
		inputAreaHeight = height / 2
	}
	s.inputArea.SetHeight(inputAreaHeight)
	s.inputArea.SetWidth(width)

	s.viewport.Width = width
	s.viewport.Height = height - inputAreaHeight - 1
	s.renderConversation()
}

func (s *ChatScreen) GetInputBubble() *textarea.Model {
	return &s.inputArea
}

func (s *ChatScreen) HandleSubmit(app *App) tea.Cmd {
	s.app = app
	userInput := s.inputArea.Value()
	s.inputArea.Reset()
	s.inputArea.Focus()

	if strings.TrimSpace(userInput) == "" {
		return nil
	}

	if strings.ToLower(strings.TrimSpace(userInput)) == "/end" {
		if app != nil && app.GetLogger() != nil {
			app.GetLogger().Info("ChatScreen sending close message", "screen", s.Name())
		}
		if s.workerInstanceID != "" && app != nil && app.GetAIWorkerManager() != nil {
			go func() {
				// Adjusting call to RetireWorkerInstance based on compiler error
				err := app.GetAIWorkerManager().RetireWorkerInstance(
					s.workerInstanceID,
					"Chat ended by user.",
					core.InstanceStatusRetiredCompleted, // finalStatus
					core.TokenUsageMetrics{},            // finalTokenUsage (empty struct)
					nil,                                 // performanceRecords (nil slice)
				)
				if err != nil && app.GetLogger() != nil {
					app.GetLogger().Error("Failed to retire worker instance on chat end", "instanceID", s.workerInstanceID, "error", err)
				}
			}()
		}
		return func() tea.Msg { return closeScreenMsg{ScreenName: s.Name()} }
	}

	s.conversationManager.AddUserMessage(userInput)
	s.renderConversation()

	genaiHistory := s.conversationManager.GetHistory()
	// Adjusting call based on compiler error: core.ConvertGenaiContentsToConversationTurns returns 1 value
	coreHistory := core.ConvertGenaiContentsToConversationTurns(genaiHistory)
	// Since it returns one value, we assume no error is returned or error handling is implicit.
	// This is risky; ideally, the function should return an error.
	// If coreHistory could be nil on error, add a check:
	if coreHistory == nil {
		if app != nil && app.GetLogger() != nil {
			app.GetLogger().Error("Failed to convert genai history to core turns (conversion returned nil)", "screen", s.Name())
		}
		s.conversationManager.AddModelMessage("[Error preparing message for AI: conversion failed]")
		s.renderConversation()
		return nil
	}

	return func() tea.Msg {
		return sendAIChatMsg{
			OriginatingScreenName: s.Name(),
			InstanceID:            s.workerInstanceID,
			DefinitionID:          s.workerDefinitionID,
			History:               coreHistory,
		}
	}
}

func (s *ChatScreen) Focus(app *App) tea.Cmd {
	s.app = app
	s.renderConversation()
	s.inputArea.Focus()
	return textarea.Blink
}

func (s *ChatScreen) Blur(app *App) tea.Cmd {
	s.inputArea.Blur()
	return nil
}

func (s *ChatScreen) renderConversation() {
	var sb strings.Builder
	history := s.conversationManager.GetHistory()

	for _, content := range history {
		if content == nil {
			continue
		}
		var speakerStyle lipgloss.Style
		speaker := "Unknown"
		text := ""

		switch content.Role {
		case "user":
			speakerStyle = userStyle
			speaker = "You"
		case "model":
			speakerStyle = aiStyle
			speaker = "AI"
		default:
			speakerStyle = systemStyle
			speaker = strings.Title(content.Role)
		}

		for _, part := range content.Parts {
			if txt, ok := part.(genai.Text); ok {
				text += string(txt)
			} else if fc, ok := part.(genai.FunctionCall); ok {
				text += fmt.Sprintf("\n  [Tool Call: %s(%v)]", fc.Name, fc.Args)
				speakerStyle = aiToolCallStyle
			} else if fr, ok := part.(genai.FunctionResponse); ok {
				text += fmt.Sprintf("\n  [Tool Result for %s: %v]", fr.Name, fr.Response)
				speakerStyle = toolResultStyle
			}
		}

		sb.WriteString(speakerStyle.Render(speaker+":") + "\n")
		sb.WriteString("  " + text + "\n\n")
	}

	s.viewport.SetContent(sb.String())
	s.viewport.GotoBottom()
}

type systemMessageToChatScreen struct {
	TargetScreenName string
	Content          string
}
