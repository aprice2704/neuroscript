// NeuroScript Version: 0.3.0
// File version: 0.1.5
// TUI-specific AppAccess getters moved to app_access.go
// filename: pkg/neurogo/app.go
// nlines: 78 // Approximate
// risk_rating: LOW
package neurogo

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/neurodata/models"
)

// App orchestrates the NeuroScript application components.
type App struct {
	Config *Config
	Log    logging.Logger

	// LLM Interaction
	llmClient core.LLMClient

	// Interpreter and State
	interpreter     *core.Interpreter
	aiWorkerManager *core.AIWorkerManager
	mu              sync.RWMutex

	// Patch Handling
	patchHandler PatchHandler

	// Loaded Schema (optional, depends on application needs)
	loadedSchema *models.Schema
}

// NewApp creates a new NeuroGo application instance.
func NewApp(logger logging.Logger) *App {
	if logger == nil {
		logger = adapters.NewNoOpLogger()
		fmt.Fprintf(os.Stderr, "Critical Warning: NewApp called with nil logger. Using NoOpLogger.\n")
	} else {
		logger.Debug("Creating new App instance.")
	}
	return &App{
		Log:    logger,
		Config: NewConfig(),
	}
}

// SetInterpreter sets the interpreter instance for the app.
func (app *App) SetInterpreter(interpreter *core.Interpreter) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.interpreter = interpreter
}

// SetAIWorkerManager sets the AI Worker Manager instance for the app.
func (app *App) SetAIWorkerManager(manager *core.AIWorkerManager) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.aiWorkerManager = manager
	if app.interpreter != nil {
		app.interpreter.SetAIWorkerManager(manager)
	} else {
		app.Log.Warn("SetAIWorkerManager called before interpreter was set.")
	}
}

// GetInterpreter returns the application's core interpreter instance safely.
func (app *App) GetInterpreter() *core.Interpreter {
	app.mu.RLock()
	defer app.mu.RUnlock()
	if app.interpreter == nil {
		app.Log.Error("GetInterpreter called before interpreter was initialized!")
	}
	return app.interpreter
}

// GetAIWorkerManager returns the AI Worker Manager instance safely.
func (app *App) GetAIWorkerManager() *core.AIWorkerManager {
	app.mu.RLock()
	defer app.mu.RUnlock()
	if app.aiWorkerManager == nil {
		app.Log.Warn("GetAIWorkerManager called before AI Worker Manager was initialized.")
	}
	return app.aiWorkerManager
}

// Run method placeholder. The actual execution logic will depend on the mode
// and might be orchestrated from main.go or specific mode handlers.
func (app *App) Run(ctx context.Context) error {
	app.Log.Error("App.Run called directly, this logic should now be in main.go or specific handlers.")
	return fmt.Errorf("App.Run is deprecated; execution flow managed by main.go")
}

// --- Methods generally useful for AppAccess interface ---

func (a *App) GetLogger() logging.Logger {
	if a.Log == nil {
		fmt.Fprintf(os.Stderr, "Warning: GetLogger called when App.Log is nil. Returning NoOpLogger.\n")
		return adapters.NewNoOpLogger()
	}
	return a.Log
}

func (a *App) GetLLMClient() core.LLMClient {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.llmClient == nil {
		logger := a.GetLogger()
		logger.Warn("GetLLMClient called when App.llmClient is nil.")
		// As per original app.go, returning nil here if llmClient is not set.
		return nil
	}
	return a.llmClient
}

// Context returns a base context for the application.
// This method is part of satisfying the tui.AppAccess interface and can be general.
func (a *App) Context() context.Context {
	return context.Background()
}
