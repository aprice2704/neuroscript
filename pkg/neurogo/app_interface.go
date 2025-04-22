package neurogo

import (
	"io"
	"log"

	"github.com/aprice2704/neuroscript/pkg/core"
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

func (a *App) GetDebugLogger() *log.Logger {
	// Ensure non-nil
	if a.DebugLog == nil {
		a.DebugLog = log.New(io.Discard, "DEBUG-FALLBACK: ", log.LstdFlags|log.Lshortfile)
	}
	return a.DebugLog
}

func (a *App) GetInfoLogger() *log.Logger {
	// Ensure non-nil
	if a.InfoLog == nil {
		a.InfoLog = log.New(io.Discard, "INFO-FALLBACK: ", log.LstdFlags)
	}
	return a.InfoLog
}

func (a *App) GetErrorLogger() *log.Logger {
	// Ensure non-nil
	if a.ErrorLog == nil {
		a.ErrorLog = log.New(io.Discard, "ERROR-FALLBACK: ", log.LstdFlags|log.Lshortfile)
	}
	return a.ErrorLog
}

func (a *App) GetLLMClient() *core.LLMClient {
	return a.llmClient
}

// --- End Interface Implementation Methods ---
