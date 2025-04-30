// filename: pkg/neurogo/app_interface.go
package neurogo

import (
	"context"
	"time"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/generative-ai-go/genai" // Needed for ApiFileInfo placeholder
)

// AppAccess defines the interface for components (like the TUI)
// to access necessary application state and configuration.
type AppAccess interface {
	GetModelName() string
	GetSyncDir() string
	GetSandboxDir() string
	GetSyncFilter() string
	GetSyncIgnoreGitignore() bool
	GetLogger() logging.Logger
	GetLLMClient() core.LLMClient
	GetInterpreter() *core.Interpreter // Added for potential TUI access
	// Add other necessary methods, e.g., GetAgentContext() *AgentContext
}

// PatchHandler defines the interface for applying patches.
type PatchHandler interface {
	ApplyPatch(ctx context.Context, patchJSON string) error
	// Add VerifyPatch method if needed later
}

// InterpreterGetter defines an interface for getting the core interpreter.
// Used by components that need access but shouldn't know about the full App.
type InterpreterGetter interface {
	GetInterpreter() *core.Interpreter
}

// +++ Placeholder Type Definition +++
// This type is used by the commented-out API file listing logic.
// It's defined here temporarily because core.ApiFileInfo is undefined.
// The actual fields might differ once core.HelperListApiFiles is implemented.
type ApiFileInfo struct {
	Name        string          // Example: "files/abcdef123"
	DisplayName string          // Example: "my_document.txt"
	URI         string          // Example: "https://generativelanguage.googleapis.com/..."
	State       genai.FileState // Example: genai.FileStateActive
	SizeBytes   int64
	CreateTime  time.Time
	UpdateTime  time.Time
	// Add other fields as needed based on the actual API response
}

// +++ End Placeholder +++

// Consider adding other interfaces as needed, e.g., AgentContextAccessor
