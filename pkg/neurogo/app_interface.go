// NeuroScript Version: 0.3.0
// File version: 0.0.4
// Ensured GetAIWorkerManager and GetLogger (as logging.Logger) are in AppAccess.
// This is the primary AppAccess interface for the application.
// filename: pkg/neurogo/app_interface.go
// nlines: 28
// risk_rating: LOW
package neurogo

import (
	"context"
	"time"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/generative-ai-go/genai"
)

// AppAccess defines the interface for components (like the TUI model)
// to access necessary application state and configuration from the neurogo.App.
type AppAccess interface {
	GetModelName() string
	GetSyncDir() string
	GetSandboxDir() string
	GetSyncFilter() string
	GetSyncIgnoreGitignore() bool
	GetLogger() logging.Logger // Must return logging.Logger
	GetLLMClient() core.LLMClient
	GetInterpreter() *core.Interpreter
	GetAIWorkerManager() *core.AIWorkerManager // Must have this method
	Context() context.Context
	ExecuteScriptFile(ctx context.Context, scriptPath string) error
}

// PatchHandler defines the interface for applying patches.
type PatchHandler interface {
	ApplyPatch(ctx context.Context, patchJSON string) error
}

// InterpreterGetter defines an interface for getting the core interpreter.
type InterpreterGetter interface {
	GetInterpreter() *core.Interpreter
}

// ApiFileInfo placeholder
type ApiFileInfo struct {
	Name        string
	DisplayName string
	URI         string
	State       genai.FileState
	SizeBytes   int64
	CreateTime  time.Time
	UpdateTime  time.Time
}
