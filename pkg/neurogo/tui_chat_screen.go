// NeuroScript Version: 0.4.0
// File version: 0.1.3 // Corrected compiler errors: _sID unused, tr.IDProvider
// Description: Implements the ChatConversationScreen for the TUI.
// filename: pkg/neurogo/tui_chat_screen.go
package neurogo

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ChatConversationScreen displays the conversation with an AI worker.
type ChatConversationScreen struct {
	name      string
	title     string
	textView  *tview.TextView
	app       *App   // To get active chat details for title and AIWM
	sessionID string // ID of the chat session this screen represents
}

// NewChatConversationScreen creates a new chat screen.
func NewChatConversationScreen(app *App, sessionID string, initialTitle string) *ChatConversationScreen {
	if app == nil {
		panic("ChatConversationScreen requires a non-nil app instance")
	}
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	tv.SetBorder(false)

	shortSessionIDForName := sessionID
	if len(sessionID) > 8 {
		shortSessionIDForName = sessionID[:8]
	}

	return &ChatConversationScreen{
		app:       app,
		name:      fmt.Sprintf("Chat-%s", shortSessionIDForName),
		title:     initialTitle,
		textView:  tv,
		sessionID: sessionID,
	}
}

// Name returns the screen's short identifier.
func (cs *ChatConversationScreen) Name() string { return cs.name }

// Title returns the screen's current title.
func (cs *ChatConversationScreen) Title() string {
	cs.updateTitle()
	return cs.title
}

func (cs *ChatConversationScreen) updateTitle() {
	var sess *ChatSession
	if cs.sessionID != "" {
		sess = cs.app.GetChatSession(cs.sessionID)
	} else {
		// This case should ideally not be hit if screens are always created with a sessionID
		// However, as a fallback, it could check the globally active session.
		cs.app.Log.Warn("ChatConversationScreen.updateTitle called with empty sessionID, attempting to use global active session.", "screenName", cs.name)
		sess = cs.app.GetActiveChatSession()
	}

	newTitle := "Chat (No Session)" // Default if session is nil
	if sess != nil && sess.WorkerInstance != nil {
		var statusColor string
		status := sess.WorkerInstance.Status
		switch status {
		case core.InstanceStatusIdle:
			statusColor = "[green]"
		case core.InstanceStatusBusy:
			statusColor = "[yellow]"
		case core.InstanceStatusError:
			statusColor = "[red]"
		default:
			statusColor = "[white]"
		}

		titleName := sess.DisplayName
		if titleName == "" && sess.DefinitionID != "" {
			if aiWM := cs.app.GetAIWorkerManager(); aiWM != nil {
				if def, err := aiWM.GetWorkerDefinition(sess.DefinitionID); err == nil && def != nil {
					titleName = def.Name
				}
			}
		}
		if titleName == "" { // Ultimate fallback
			titleName = sess.DefinitionID
		}

		shortInstID := sess.WorkerInstance.InstanceID
		if len(shortInstID) > 8 {
			shortInstID = shortInstID[:8]
		}
		newTitle = fmt.Sprintf("%s (%s) %s%s[-]", EscapeTviewTags(titleName), shortInstID, statusColor, status)
	} else if cs.sessionID == "" { // If this screen has no sessionID and there was no active global session
		// Use underscore for unused sessionID from GetActiveChatDetails
		_, dispName, _defID, _instStatus, isActive := cs.app.GetActiveChatDetails()
		if isActive {
			var statusColor string
			switch _instStatus {
			case core.InstanceStatusIdle:
				statusColor = "[green]"
			case core.InstanceStatusBusy:
				statusColor = "[yellow]"
			case core.InstanceStatusError:
				statusColor = "[red]"
			default:
				statusColor = "[white]"
			}
			titleName := dispName
			if titleName == "" {
				titleName = _defID
			}
			newTitle = fmt.Sprintf("Chat: %s %s%s[-]", EscapeTviewTags(titleName), statusColor, _instStatus)
		} else {
			newTitle = "Chat (Inactive)"
		}
	}
	cs.title = newTitle
}

// Primitive returns the tview.Primitive for this screen.
func (cs *ChatConversationScreen) Primitive() tview.Primitive {
	cs.updateTitle()
	return cs.textView
}

// OnFocus is called when the screen gains focus.
func (cs *ChatConversationScreen) OnFocus(setFocus func(p tview.Primitive)) {
	setFocus(cs.textView)
	cs.textView.ScrollToEnd()
}

// OnBlur is called when the screen loses focus.
func (cs *ChatConversationScreen) OnBlur() {
}

// InputHandler returns the input handler for this screen.
func (cs *ChatConversationScreen) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
		textViewHandlerFunc := cs.textView.InputHandler()
		if textViewHandlerFunc != nil {
			textViewHandlerFunc(event, setFocus)
		}
		return event
	}
}

// IsFocusable indicates if the screen's primitive can be focused.
func (cs *ChatConversationScreen) IsFocusable() bool {
	return true
}

// UpdateConversation formats and sets the chat history on the TextView.
func (cs *ChatConversationScreen) UpdateConversation() {
	if cs.textView == nil || cs.app == nil {
		return
	}

	var history []*interfaces.ConversationTurn
	if cs.sessionID != "" {
		session := cs.app.GetChatSession(cs.sessionID)
		if session != nil {
			history = session.GetConversationHistory()
		} else {
			cs.textView.SetText(fmt.Sprintf("[red]Error: Chat session %s not found.[-]", EscapeTviewTags(cs.sessionID)))
			return
		}
	} else {
		// This screen has no specific session ID; this indicates a potential logic error
		// as chat screens should be tied to a session.
		cs.app.Log.Warn("ChatConversationScreen.UpdateConversation called for a screen with no sessionID.", "screenName", cs.name)
		cs.textView.SetText(fmt.Sprintf("[red]Error: Screen %s has no associated chat session ID.[-]", EscapeTviewTags(cs.name)))
		history = []*interfaces.ConversationTurn{} // Show empty
	}

	cs.textView.Clear()
	var builder strings.Builder
	for _, turn := range history {
		roleLabel := turn.Role
		roleColor := "white"

		switch turn.Role {
		case interfaces.RoleUser:
			roleLabel = "You"
			roleColor = "blue"
		case interfaces.RoleAssistant, "model":
			roleLabel = "AI"
			roleColor = "green"
		case interfaces.RoleTool, "function":
			roleLabel = "Tool"
			roleColor = "yellow"
		default:
			roleColor = "grey"
		}

		content := turn.Content
		if content == "" && len(turn.ToolCalls) > 0 {
			var tcContent strings.Builder
			tcContent.WriteString("requests Tool Call(s):")
			for _, tc := range turn.ToolCalls {
				tcContent.WriteString(fmt.Sprintf("\n  - %s(%s)", tc.Name, EscapeTviewTags(fmt.Sprintf("%v", tc.Arguments))))
			}
			content = tcContent.String()
		} else if content == "" && len(turn.ToolResults) > 0 {
			var trContent strings.Builder
			trContent.WriteString("provides Tool Result(s):")
			for _, tr := range turn.ToolResults {
				// Use tr.ID directly as per interfaces.ToolResult definition
				trContent.WriteString(fmt.Sprintf("\n  - ID %s: %s (Error: %s)",
					EscapeTviewTags(tr.ID), // Corrected to use tr.ID
					EscapeTviewTags(fmt.Sprintf("%v", tr.Result)),
					EscapeTviewTags(tr.Error)))
			}
			content = trContent.String()
		}
		builder.WriteString(fmt.Sprintf("[::b][%s]%s:[::-] %s\n\n", roleColor, strings.Title(string(roleLabel)), EscapeTviewTags(content)))
	}
	cs.textView.SetText(builder.String())
	cs.textView.ScrollToEnd()
	cs.updateTitle()
}

var _ PrimitiveScreener = (*ChatConversationScreen)(nil)
