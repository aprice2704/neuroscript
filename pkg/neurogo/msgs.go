// NeuroScript Version: 0.3.0
// File version: 0.0.5 // Changed LLMResponseCandidate to ConversationTurn
// Corrected aiResponseMsg to use core.LLMResponseCandidate.
// filename: pkg/neurogo/msgs.go
// nlines: 55
// risk_rating: LOW
package neurogo

import (
	"time"

	"github.com/aprice2704/neuroscript/pkg/core" // Ensure core is imported
)

// errMsg is used to signal an error to the TUI.
type errMsg struct{ err error }

// Error returns the error message.
func (e errMsg) Error() string { return e.err.Error() }

// scriptEmitMsg is used to send lines emitted by a NeuroScript script to the TUI.
type scriptEmitMsg struct{ Content string }

// initialScriptDoneMsg signals that the initial startup script has finished.
type initialScriptDoneMsg struct {
	Path string
	Err  error
}

// syncCompleteMsg signals that a sync operation has finished.
type syncCompleteMsg struct {
	err   error
	stats map[string]interface{} // Or a more structured type for stats
}

// patchAppliedMsg signals that a patch has been applied.
type patchAppliedMsg struct {
	Summary string
	Err     error
}

// --- Screen and Chat Specific Messages ---

// closeScreenMsg signals the main model to close/remove a screen.
type closeScreenMsg struct {
	ScreenName string // The unique name of the screen to close
}

// aiResponseMsg is sent by the main model (or an AI call handler) to a specific ChatScreen
// with the AI's response.
type aiResponseMsg struct {
	TargetScreenName  string
	ResponseCandidate *core.ConversationTurn // Changed from LLMResponseCandidate
	Err               error
}

// sendAIChatMsg is sent by a ChatScreen to the main model to request an AI call.
type sendAIChatMsg struct {
	OriginatingScreenName string
	InstanceID            string
	DefinitionID          string
	History               []*core.ConversationTurn // Uses core type
}

// updateStatusBarMsg is used to update the status bar text.
type updateStatusBarMsg string

// refreshViewMsg is used to signal a refresh of a specific screen or all screens.
type refreshViewMsg struct {
	ScreenName string // Optional: if empty, refresh current or all.
	Timestamp  time.Time
}
