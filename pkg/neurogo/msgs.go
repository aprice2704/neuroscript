// NeuroScript Version: 0.3.0
// File version: 0.0.1 // Added scriptEmitMsg and initialScriptDoneMsg
// filename: pkg/neurogo/tui/msgs.go
package neurogo

import (
	"github.com/google/generative-ai-go/genai"
	// Import core package if core types are needed in messages
	// "github.com/aprice2704/neuroscript/pkg/core"
)

// errMsg is used to bubble errors up through the Bubble Tea runtime.
type errMsg struct{ err error }

// Error satisfies the error interface.
func (e errMsg) Error() string { return e.err.Error() }

// --- Script EMIT message ---

// scriptEmitMsg is the message sent to the TUI when a script EMIT statement
// is processed by the TUIEmitWriter.
type scriptEmitMsg struct {
	Content string
}

// --- Initial Script Execution Messages ---

// initialScriptDoneMsg is sent when the initial startup script (if provided)
// has finished executing. It includes the path of the script and any error
// that occurred during its execution.
type initialScriptDoneMsg struct {
	Path string
	Err  error
}

// --- Placeholder Message Types (Existing) ---
// These messages will be sent by commands when background operations complete.

// llmResponseMsg carries the response from an LLM call.
type llmResponseMsg struct {
	response *genai.GenerateContentResponse // Or relevant part of the response
}

// toolResultMsg carries the result (or error) from a tool execution.
type toolResultMsg struct {
	toolName string
	result   interface{} // Can be map[string]interface{}, string, etc.
	err      error
}

// syncCompleteMsg carries the results of a file sync operation.
// Updated: Stats map uses interface{} to match SyncDirectoryUpHelper return type.
type syncCompleteMsg struct {
	stats map[string]interface{} // Use interface{} and perform type assertions later
	err   error
}

// statusUpdateMsg can be used for generic status updates (e.g., file counts).
type statusUpdateMsg struct {
	message string
	// Add specific fields if needed, e.g., file counts
	// LocalFiles int
	// ApiFiles   int
}

// Add other message types as needed, e.g., for patch application results.
// type patchResultMsg struct { err error }
