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
	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// NewApp creates and initializes a new App instance.
func NewApp(config *Config, logger logging.Logger, llmclient core.LLMClient) (*App, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		logger = adapters.NewNoOpLogger()
		logger.Warn("No logger provided to NewApp, using NoOpLogger.")
	}

	appCtx, cancelFunc := context.WithCancel(context.Background())

	a := &App{
		Config:     config,
		Log:        logger,
		appCtx:     appCtx,
		cancelFunc: cancelFunc,
		// chatSessions map is initialized in initializeCoreComponents
	}

	// If an LLM client is passed in (e.g., from tests or specific setup), use it.
	// Otherwise, initializeCoreComponents will create one.
	if llmclient != nil {
		a.llmClient = llmclient
		a.Log.Info("LLM Client provided to NewApp and assigned.")
	} else {
		// llmClient will be created in initializeCoreComponents
		a.Log.Debug("No LLM client provided to NewApp; will be created during core component initialization.")
	}

	a.Log.Debug("Basic App struct initialized. Core components (including LLMClient if not provided) will be initialized by Run->initializeCoreComponents.")
	return a, nil
}

// CreateLLMClient method REMOVED from here, as it's now in app.go
// func (app *App) CreateLLMClient() (core.LLMClient, error) { ... }
