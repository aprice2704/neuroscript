// NeuroScript Version: 0.3.0
// File version: 0.0.3 (Removed duplicate CreateLLMClient)
// filename: pkg/neurogo/app_init.go
package neurogo

import (
	"context" // For App context initialization
	"fmt"

	// os is no longer needed here directly if CreateLLMClient is removed
	// "os"

	// For robust path joining
	// "path/filepath" // No longer needed here directly

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// NewApp creates and initializes a new App instance.
func NewApp(config *Config, logger interfaces.Logger, llmclient interfaces.LLMClient) (*App, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		logger = logging.NewNoLogger()
		logger.Warn("No logger provided to NewApp, using NoOpLogger.")
	}

	logger.Infof("NeuroScript grammar %s", core.GrammarVersion)

	appCtx, cancelFunc := context.WithCancel(context.Background())

	a := &App{
		Config:     config,
		Log:        logger,
		appCtx:     appCtx,
		cancelFunc: cancelFunc,
		// chatSessions map is initialized in InitializeCoreComponents
	}

	// If an LLM client is passed in (e.g., from tests or specific setup), use it.
	// Otherwise, InitializeCoreComponents will create one.
	if llmclient != nil {
		a.llmClient = llmclient
		a.Log.Info("LLM Client provided to NewApp and assigned.")
	} else {
		// llmClient will be created in InitializeCoreComponents
		a.Log.Debug("No LLM client provided to NewApp; will be created during core component initialization.")
	}

	a.Log.Debug("Basic App struct initialized. Core components (including LLMClient if not provided) will be initialized by Run->InitializeCoreComponents.")
	return a, nil
}

// CreateLLMClient method REMOVED from here, as it's now in app.go
// func (app *App) CreateLLMClient() (interfaces.LLMClient, error) { ... }
