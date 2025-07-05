// NeuroScript Version: 0.3.0
// File version: 0.0.7
// Removed wm package import and GetAIWorkerManager from AppAccess interface.
// filename: pkg/neurogo/app_interface.go
package neurogo

import (
	"context"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
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
	GetLogger() interfaces.Logger
	GetLLMClient() interfaces.LLMClient
	GetInterpreter() *interpreter.Interpreter
	// GetAIWorkerManager() *wm.AIWorkerManager // This method is removed for now
	Context() context.Context
	ExecuteScriptFile(ctx context.Context, scriptPath string) error
}

// PatchHandler defines the interface for applying patches.
type PatchHandler interface {
	ApplyPatch(ctx context.Context, patchJSON string) error
}

// InterpreterGetter defines an interface for getting the core interpreter.
type InterpreterGetter interface {
	GetInterpreter() interpreter.Interpreter
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

// StandardFileSystem implements FileSystemOperations using the os package.
// type StandardFileSystem struct{}

// func (sfs *StandardFileSystem) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }
// func (sfs *StandardFileSystem) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }
// func (sfs *StandardFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
// 	return os.WriteFile(name, data, perm)
// }
// func (sfs *StandardFileSystem) MkdirAll(path string, perm fs.FileMode) error {
// 	return os.MkdirAll(path, perm)
// }
// func (sfs *StandardFileSystem) Remove(name string) error       { return os.Remove(name) }
// func (sfs *StandardFileSystem) UserHomeDir() (string, error) { return os.UserHomeDir() }
// func (sfs *StandardFileSystem) Getenv(key string) string       { return os.Getenv(key) }
