// filename: pkg/neurogo/tui/msgs.go
package tui

import (
	"github.com/google/generative-ai-go/genai"
	// Import core package if core types are needed in messages
	// "github.com/aprice2704/neuroscript/pkg/core"
)

// errMsg is used to bubble errors up through the Bubble Tea runtime.
type errMsg struct{ err error }

// Error satisfies the error interface.
func (e errMsg) Error() string { return e.err.Error() }

// --- Placeholder Message Types ---
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
type syncCompleteMsg struct {
	stats map[string]int // Use the stats map defined in core.SyncDirectoryUpHelper
	err   error
}

// statusUpdateMsg can be used for generic status updates (e.g., file counts).
type statusUpdateMsg struct {
	message string
	// Add specific fields if needed, e.g., file counts
	// LocalFiles int
	// ApiFiles   int
}

// Add other message types as needed, e.g., for patch application results.
// type patchResultMsg struct { err error }
