// NeuroScript Version: 0.4.0
// File version: 0.1.0
// Description: Implements the ChatConversationScreen for the TUI.
// filename: pkg/neurogo/tui_chat_screen.go
// nlines: 130 // Approximate
// risk_rating: MEDIUM
package neurogo

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ChatConversationScreen displays the conversation with an AI worker.
type ChatConversationScreen struct {
	name     string
	title    string
	textView *tview.TextView
	app      *App // To get active chat details for title and AIWM
}

// NewChatConversationScreen creates a new chat screen.
func NewChatConversationScreen(app *App) *ChatConversationScreen {
	if app == nil {
		panic("ChatConversationScreen requires a non-nil app instance")
	}
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	tv.SetBorder(false)

	return &ChatConversationScreen{
		app:      app,
		name:     "Chat",            // Static name for identification
		title:    "Chat (Inactive)", // Initial title, updated by Primitive()
		textView: tv,
	}
}

// Name returns the screen's short identifier.
func (cs *ChatConversationScreen) Name() string { return cs.name }

// Title returns the screen's current title.
func (cs *ChatConversationScreen) Title() string { return cs.title }

// Primitive returns the tview.Primitive for this screen.
// It updates the title based on the current chat state.
func (cs *ChatConversationScreen) Primitive() tview.Primitive {
	defID, instID, status, isActive := cs.app.GetActiveChatInstanceDetails()
	newTitle := "Chat (Inactive)"

	if isActive {
		var statusColor string
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
		shortInstID := instID
		if len(instID) > 8 {
			shortInstID = instID[:8]
		}

		defName := defID // Fallback to DefinitionID
		if aiWM := cs.app.GetAIWorkerManager(); aiWM != nil && cs.app.activeChatInstance != nil {
			// Bug #1 Fix: Call GetWorkerDefinition instead of GetDefinition
			if def, err := aiWM.GetWorkerDefinition(cs.app.activeChatInstance.DefinitionID); err == nil && def != nil {
				defName = def.Name
			} else if err != nil {
				// Log error if needed: cs.app.Log.Warnf("Failed to get def name for chat title: %v", err)
			}
		}
		newTitle = fmt.Sprintf("Chat: %s (%s) %s%s[-]", defName, shortInstID, statusColor, status)
	}

	cs.title = newTitle
	// cs.textView.SetTitle(cs.title) // TextView itself doesn't usually have a title if it's the page content.
	// The title is typically for the Page in a Pages view or for status bars.
	return cs.textView
}

// OnFocus is called when the screen gains focus.
func (cs *ChatConversationScreen) OnFocus(setFocus func(p tview.Primitive)) {
	setFocus(cs.textView)
	cs.textView.ScrollToEnd() // Ensure latest message is visible
}

// OnBlur is called when the screen loses focus.
func (cs *ChatConversationScreen) OnBlur() {
	// No specific action needed on blur for now.
}

// InputHandler returns the input handler for this screen.
// Bug #2 Fix: Ensure the returned function matches the PrimitiveScreener interface.
func (cs *ChatConversationScreen) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
		// Get the TextView's own input handler function
		// (which has signature: func(event *tcell.EventKey, setFocus func(p tview.Primitive)))
		textViewHandlerFunc := cs.textView.InputHandler()
		textViewHandlerFunc(event, setFocus) // Call it; it doesn't return the event.

		// The PrimitiveScreener interface expects this function to return the event
		// if it's not fully "consumed" by this screen's specific logic.
		// Since TextView's default handler handles scrolling etc., and we don't add
		// other keybindings here that would consume events, we return the original event
		// to allow further processing by global handlers if necessary.
		return event
	}
}

// IsFocusable indicates if the screen's primitive can be focused.
func (cs *ChatConversationScreen) IsFocusable() bool {
	return true // The TextView should be focusable for scrolling.
}

// UpdateConversation formats and sets the chat history on the TextView.
func (cs *ChatConversationScreen) UpdateConversation(history []*core.ConversationTurn) {
	if cs.textView == nil {
		return
	}
	cs.textView.Clear() // Clear previous content
	var builder strings.Builder
	for _, turn := range history {
		roleLabel := turn.Role
		roleColor := "white" // Default color

		switch turn.Role {
		case core.RoleUser:
			roleLabel = "You"
			roleColor = "blue"
		case core.RoleAssistant, "model": // "model" is common from genai
			roleLabel = "AI"
			roleColor = "green"
		case core.RoleTool, "function": // "function" is common from genai
			roleLabel = "Tool"
			roleColor = "yellow"
		default:
			roleColor = "grey" // Or handle as unknown
		}

		content := turn.Content
		if content == "" && len(turn.ToolCalls) > 0 {
			var tcContent strings.Builder
			tcContent.WriteString("requests Tool Call(s):")
			for _, tc := range turn.ToolCalls {
				// Ensure arguments are displayable, potentially Marshal to JSON for complex args
				tcContent.WriteString(fmt.Sprintf("\n  - %s(%s)", tc.Name, EscapeTviewTags(fmt.Sprintf("%v", tc.Arguments))))
			}
			content = tcContent.String()
		} else if content == "" && len(turn.ToolResults) > 0 {
			var trContent strings.Builder
			trContent.WriteString("provides Tool Result(s):")
			for _, tr := range turn.ToolResults {
				trContent.WriteString(fmt.Sprintf("\n  - ID %s: %s (Error: %s)",
					EscapeTviewTags(tr.ID),
					EscapeTviewTags(fmt.Sprintf("%v", tr.Result)),
					EscapeTviewTags(tr.Error)))
			}
			content = trContent.String()
		}
		// Ensure all parts of the content are escaped
		builder.WriteString(fmt.Sprintf("[::b][%s]%s:[::-] %s\n\n", roleColor, strings.Title(string(roleLabel)), EscapeTviewTags(content)))
	}
	cs.textView.SetText(builder.String())
	cs.textView.ScrollToEnd()
	// cs.Primitive() // Calling Primitive here could update the title, but might be redundant if draw cycle handles it.
	// If the title depends on content (e.g. last message summary), then yes. For now, title depends on instance state.
}

// Ensure ChatConversationScreen satisfies the PrimitiveScreener interface.
var _ PrimitiveScreener = (*ChatConversationScreen)(nil)
