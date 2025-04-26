package neurogo

import (
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// --- Interface Implementation Methods (tui.AppAccess) ---

func (a *App) GetModelName() string {
	if a.Config != nil && a.Config.ModelName != "" {
		return a.Config.ModelName
	}
	return "unknown"
}

func (a *App) GetSyncDir() string {
	if a.Config != nil {
		return a.Config.SyncDir
	}
	return "." // Default from NewConfig
}

func (a *App) GetSandboxDir() string {
	if a.Config != nil {
		return a.Config.SandboxDir
	}
	return "." // Default from NewConfig
}

func (a *App) GetSyncFilter() string {
	if a.Config != nil {
		return a.Config.SyncFilter
	}
	return ""
}

func (a *App) GetSyncIgnoreGitignore() bool {
	if a.Config != nil {
		return a.Config.SyncIgnoreGitignore
	}
	return false
}

func (a *App) GetLogger() interfaces.Logger {
	// Ensure non-nil
	if a.Logger == nil {
		panic("Must have a valid logger")
	}
	return a.Logger
}

func (a *App) GetLLMClient() *core.LLMClient {
	return a.llmClient
}

// --- End Interface Implementation Methods ---
