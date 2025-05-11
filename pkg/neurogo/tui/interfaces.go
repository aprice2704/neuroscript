// NeuroScript Version: 0.3.0
// File version: 0.0.1 // Add ExecuteScriptFile and Context methods.
// filename: pkg/neurogo/tui/interfaces.go
// nlines: 25 // Approximate
// risk_rating: LOW
package tui

import (
	"context" // Added for Context method and ExecuteScriptFile

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// AppAccess defines the methods the TUI components need
// to interact with the main application state and configuration,
// without directly importing the neurogo package.
type AppAccess interface {
	// Config Accessors
	GetModelName() string
	GetSyncDir() string
	GetSandboxDir() string
	GetSyncFilter() string
	GetSyncIgnoreGitignore() bool
	// Add other config getters as needed

	// Logger Accessor
	GetLogger() logging.Logger

	// Client Accessors
	GetLLMClient() core.LLMClient

	// Interpreter Accessor
	GetInterpreter() *core.Interpreter

	// Script Execution
	ExecuteScriptFile(ctx context.Context, scriptPath string) error

	// Context Accessor
	Context() context.Context // Provides a general context if needed by TUI operations
}
