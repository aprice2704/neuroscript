// filename: pkg/core/llm_types.go
package core

import (
	"context"
	"fmt"

	// Need genai import for the new Client() method return type
	"github.com/google/generative-ai-go/genai"
	// TokenUsageMetrics is defined in ai_worker_types.go, ensure it's accessible
	// If not directly, this implies ai_worker_types.go is in the same package or imported.
	// Assuming it's accessible as `TokenUsageMetrics` directly or as `core.TokenUsageMetrics`
	// For this file, direct accessibility implies it's in the same package 'core',
	// which aligns with `ai_worker_types.go` also being in `package core`.
)

// Role defines the speaker role in a conversation turn.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool" // For tool execution results, distinct from model's function call request
)

// ConversationTurn represents a single turn in a conversation.
type ConversationTurn struct {
	Role        Role              `json:"role"`                   // Speaker role (system, user, assistant, tool)
	Content     string            `json:"content"`                // Text content of the turn
	ToolCalls   []*ToolCall       `json:"tool_calls,omitempty"`   // List of tool calls requested by the assistant
	ToolResults []*ToolResult     `json:"tool_results,omitempty"` // List of results from tool calls (added by the system/tool)
	TokenUsage  TokenUsageMetrics `json:"token_usage,omitempty"`  // <<< ADDED FIELD: Token usage for generating this turn (primarily for assistant turns)
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

// String returns a string representation of the ConversationTurn.
func (t *ConversationTurn) String() string {
	base := fmt.Sprintf("[%s]: %s", t.Role, t.Content)
	if len(t.ToolCalls) > 0 {
		calls := ""
		for _, tc := range t.ToolCalls {
			calls += fmt.Sprintf("\n  ToolCall(ID: %s, Name: %s, Args: %v)", tc.ID, tc.Name, tc.Arguments)
		}
		base += calls
	}
	if len(t.ToolResults) > 0 {
		results := ""
		for _, tr := range t.ToolResults {
			resStr := fmt.Sprintf("%v", tr.Result)
			if tr.Error != "" {
				resStr = fmt.Sprintf("Error: %s", tr.Error)
			}
			results += fmt.Sprintf("\n  ToolResult(ID: %s, Result: %s)", tr.ID, resStr)
		}
		base += results
	}
	if t.TokenUsage.TotalTokens > 0 || t.TokenUsage.InputTokens > 0 || t.TokenUsage.OutputTokens > 0 { // Only show if non-zero
		base += fmt.Sprintf("\n  Tokens(In: %d, Out: %d, Total: %d)", t.TokenUsage.InputTokens, t.TokenUsage.OutputTokens, t.TokenUsage.TotalTokens)
	}
	return base
}
