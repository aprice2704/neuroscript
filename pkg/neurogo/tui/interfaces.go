// filename: pkg/neurogo/tui/interfaces.go
package tui

import (
	"github.com/aprice2704/neuroscript/pkg/core" // Need this for LLMClient type
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// AppAccess defines the methods the TUI components need
// to interact with the main application state and configuration,
// without directly importing the neurogo package.
type AppAccess interface {
	// Config Accessors
	GetModelName() string
	GetSyncDir() string
	GetSandboxDir() string // Needed for secure path validation
	GetSyncFilter() string
	GetSyncIgnoreGitignore() bool
	// Add other config getters as needed

	// Logger Accessors
	GetLogger() logging.Logger

	// Client Accessors
	GetLLMClient() *core.LLMClient // Added
}
