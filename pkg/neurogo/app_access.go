// NeuroScript Version: 0.3.0
// File version: 0.0.1
// Moved TUI AppAccess getter methods from app.go
// filename: pkg/neurogo/app_access.go
// nlines: 40 // Approximate based on moved methods
// risk_rating: LOW // Moving existing code, no functional changes
package neurogo

// --- Methods implementing parts of AppAccess interface (e.g., for TUI) ---
// These provide access to configuration or app state needed by external packages like TUI.

func (a *App) GetModelName() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.Config == nil {
		return ""
	}
	return a.Config.ModelName
}

func (a *App) GetSyncDir() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.Config == nil {
		return ""
	}
	return a.Config.SyncDir
}

func (a *App) GetSandboxDir() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.Config == nil {
		return ""
	}
	return a.Config.SandboxDir
}

func (a *App) GetSyncFilter() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.Config == nil {
		return ""
	}
	return a.Config.SyncFilter
}

func (a *App) GetSyncIgnoreGitignore() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.Config == nil {
		return false
	}
	return a.Config.SyncIgnoreGitignore
}