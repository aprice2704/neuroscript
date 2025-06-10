// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Reconstructed central LLM interface and defined concrete ConversationTurn struct.
// filename: pkg/interfaces/llm.go
// nlines: 28
// risk_rating: LOW

package interfaces

import (
	"context"

	"github.com/google/generative-ai-go/genai" // correct import is "github.com/google/generative-ai-go/genai"
)

// Role defines the speaker role in a conversation turn.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// LLMClient defines the interface for interacting with a Large Language Model.
type LLMClient interface {
	// Ask sends a conversation history to the LLM and expects a response turn.
	// The returned ConversationTurn should have TokenUsage populated if available.
	Ask(ctx context.Context, turns []*ConversationTurn) (*ConversationTurn, error)

	// AskWithTools sends a conversation history and available tools, expecting
	// either a response turn or a request to call specific tools.
	// The returned ConversationTurn should have TokenUsage populated if available.
	AskWithTools(ctx context.Context, turns []*ConversationTurn, tools []ToolDefinition) (*ConversationTurn, []*ToolCall, error)

	// Embed generates vector embeddings for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)

	// Client returns the underlying *genai.Client, if available, otherwise nil.
	// This allows helpers needing the specific client type to access it safely.
	Client() *genai.Client
}

// ConversationTurn represents a single turn in a conversation.
type ConversationTurn struct {
	Role        Role          `json:"role"`                   // Speaker role (system, user, assistant, tool)
	Content     string        `json:"content"`                // Text content of the turn
	ToolCalls   []*ToolCall   `json:"tool_calls,omitempty"`   // List of tool calls requested by the assistant
	ToolResults []*ToolResult `json:"tool_results,omitempty"` // List of results from tool calls (added by the system/tool)
}

// ToolCall represents a request from the LLM to call a specific tool.
type ToolCall struct {
	ID        string         `json:"id"`        // Unique identifier for the tool call instance
	Name      string         `json:"name"`      // Name of the tool to call
	Arguments map[string]any `json:"arguments"` // Arguments for the tool call, structured as a map
}

// ToolResult represents the result of executing a tool call.
type ToolResult struct {
	ID     string `json:"id"`              // ID matching the corresponding ToolCall
	Result any    `json:"result"`          // Result data from the tool execution (can be string, number, bool, list, map)
	Error  string `json:"error,omitempty"` // Error message if the tool execution failed
}

// ToolDefinition describes a tool that can be made available to the LLM.
type ToolDefinition struct {
	Name        string `json:"name"`                  // Name of the tool
	Description string `json:"description,omitempty"` // Description of what the tool does
	InputSchema any    `json:"input_schema"`          // JSON Schema object describing the input parameters
}
