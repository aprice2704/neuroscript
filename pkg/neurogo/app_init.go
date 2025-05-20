// NeuroScript Version: 0.3.0
// File version: 0.0.2 (Corrected core API calls, multi-chat init)
// filename: pkg/neurogo/app_init.go
package neurogo

import (
	"context" // For App context initialization
	"fmt"
	"os"

	// For robust path joining
	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// NewApp creates and initializes a new App instance.
// The signature of NewApp here is func NewApp(config *Config, logger logging.Logger) (*App, error)
// as per your latest app.go. The CreateLLMClient is a method on *App.
// Test files calling NewApp with a third llmClient argument will need to be updated
// or NewApp signature would need to change (which is not the current direction).
func NewApp(config *Config, logger logging.Logger) (*App, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		logger = adapters.NewNoOpLogger()
		logger.Warn("No logger provided to NewApp, using NoOpLogger.")
	}

	appCtx, cancelFunc := context.WithCancel(context.Background())

	// Initialize App struct with new multi-chat fields and other fields from your app.go
	a := &App{
		Config:     config,
		Log:        logger,
		appCtx:     appCtx,
		cancelFunc: cancelFunc,
	}

	// Initialize LLM Client using the instance method CreateLLMClient
	// CreateLLMClient is defined in this file (app_init.go, from your upload).
	llmClient, err := a.CreateLLMClient()
	if err != nil {
		// CreateLLMClient already logs and returns a NoOp client on significant error.
		a.Log.Warnf("Error during LLM client creation attempt: %v. A NoOp client might be in use.", err)
	}
	a.llmClient = llmClient // Assign the created (or NoOp) client
	if a.llmClient != nil { // Should always be non-nil due to CreateLLMClient's fallback
		a.Log.Info("LLM Client initialized.")
	}

	a.Log.Debug("Basic App struct initialized. Core components will be initialized by Run->initializeCoreComponents.")
	return a, nil
}

// CreateLLMClient function as provided in your app_init.go (version 0.0.1)
// This function creates an LLM client based on application configuration.
func (app *App) CreateLLMClient() (core.LLMClient, error) {
	if app.Config == nil {
		return nil, fmt.Errorf("cannot create LLM client: app config is nil")
	}

	apiKey := app.Config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("NEUROSCRIPT_API_KEY") // Standardized env var name
		if apiKey == "" {
			if app.Log != nil { // Check logger
				app.Log.Debug("API key is missing in config and environment variable (NEUROSCRIPT_API_KEY). Creating NoOpLLMClient.")
			}
			// Assuming adapters.NewNoOpLLMClient() is available from your pkg/adapters/noop_llmclient.go
			return adapters.NewNoOpLLMClient(), nil
		}
		if app.Log != nil {
			app.Log.Debug("Using LLM API key from environment variable NEUROSCRIPT_API_KEY.")
		}
	} else {
		if app.Log != nil {
			app.Log.Debug("Using LLM API key from configuration.")
		}
	}

	if app.Log != nil {
		app.Log.Debug("Creating real LLMClient.")
	}
	apiHost := app.Config.APIHost
	modelName := app.Config.ModelName // This is ModelID for NewLLMClient

	loggerToUse := app.Log
	if loggerToUse == nil { // Should be set by NewApp
		loggerToUse = adapters.NewNoOpLogger() // Safety fallback
	}

	// Based on pkg/core/llm.go:
	// NewLLMClient(apiKey, apiHost, modelID string, logger Logger, verifyTLS bool) LLMClient
	llmClient := core.NewLLMClient(apiKey, apiHost, modelName, loggerToUse, !app.Config.Insecure) // verifyTLS is !Insecure

	if llmClient == nil {
		// This case implies core.NewLLMClient can return nil, indicating a critical failure.
		if app.Log != nil {
			app.Log.Error("core.NewLLMClient returned nil unexpectedly. This indicates a critical LLM client creation failure.")
		}
		// It's better to return an error if the real client creation fails this way
		// and let the caller decide on a NoOp if appropriate.
		// However, your original CreateLLMClient implies it *always* returns a client or NoOp for API key missing.
		// If core.NewLLMClient returns nil for other reasons, returning an error is good.
		// For consistency with API key missing path, returning NoOp here too if llmClient is nil.
		return adapters.NewNoOpLLMClient(), fmt.Errorf("core.NewLLMClient returned nil for real client creation")
	}
	if app.Log != nil {
		app.Log.Debug("Real LLMClient created.", "host", apiHost, "model", modelName)
	}
	return llmClient, nil
}
