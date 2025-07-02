// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Added FileStorageClient interface and FileMetadata struct to support testable file API tools.
// filename: pkg/interfaces/llm.go
// nlines: 60
// risk_rating: MEDIUM

package interfaces

import (
	"context"
	"time"

	"github.com/google/generative-ai-go/genai"
)

// Role defines the speaker role in a conversation turn.
type Role string

const (
	RoleSystem	Role	= "system"
	RoleUser	Role	= "user"
	RoleAssistant	Role	= "assistant"
	RoleTool	Role	= "tool"
)

// FileMetadata holds structured information about a file stored with the LLM provider.
type FileMetadata struct {
	FileID		string
	FileName	string
	DisplayName	string
	SizeBytes	int64
	CreateTime	time.Time
	State		string
	MimeType	string
}

// FileStorageClient defines the interface for interacting with an LLM's file storage.
type FileStorageClient interface {
	ListFiles(ctx context.Context) ([]FileMetadata, error)
	UploadFile(ctx context.Context, path string, displayName string) (FileMetadata, error)
	DeleteFile(ctx context.Context, fileID string) error
	GetFile(ctx context.Context, fileID string) (FileMetadata, error)
}

// LLMClient defines the interface for interacting with a Large Language Model.
type LLMClient interface {

	// Ask sends a conversation history to the LLM and expects a response turn.
	Ask(ctx context.Context, turns []*ConversationTurn) (*ConversationTurn, error)

	// AskWithTools sends a conversation history and available tools, expecting
	// either a response turn or a request to call specific tools.
	AskWithTools(ctx context.Context, turns []*ConversationTurn, tools []ToolDefinition) (*ConversationTurn, []*ToolCall, error)

	// Embed generates vector embeddings for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)

	// Client returns the underlying *genai.Client, if available, otherwise nil.
	// This can be used for provider-specific operations not covered by the interface.
	Client() *genai.Client
}

// ConversationTurn represents a single turn in a conversation with the LLM.
type ConversationTurn struct {
	Role		Role		`json:"role"`
	Content		string		`json:"content"`
	ToolCalls	[]*ToolCall	`json:"tool_calls,omitempty"`
	ToolResults	[]*ToolResult	`json:"tool_results,omitempty"`
}

// ToolCall represents a request from the LLM to call a specific tool.
type ToolCall struct {
	ID		string		`json:"id"`
	Name		string		`json:"name"`
	Arguments	map[string]any	`json:"arguments"`
}

// ToolResult represents the result of executing a tool call.
type ToolResult struct {
	ID	string	`json:"id"`
	Result	any	`json:"result"`
	Error	string	`json:"error,omitempty"`
}

// ToolDefinition describes a tool that can be made available to the LLM.
type ToolDefinition struct {
	Name		string	`json:"name"`
	Description	string	`json:"description,omitempty"`
	InputSchema	any	`json:"input_schema"`
}