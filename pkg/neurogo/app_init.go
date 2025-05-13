// NeuroScript Version: 0.3.0
// File version: 0.0.1
// Moved LLM client initialization logic from app.go
// filename: pkg/neurogo/app_init.go
// nlines: 40 // Approximate based on moved method
// risk_rating: LOW // Moving existing code, no functional changes
package neurogo

import (
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	// logging.Logger is used via app.Log
)

// CreateLLMClient creates the LLM client based on config.
// Renamed from createLLMClient for clarity as it's called externally now.
func (app *App) CreateLLMClient() (core.LLMClient, error) {
	if app.Config == nil {
		return nil, fmt.Errorf("cannot create LLM client: app config is nil")
	}

	// LLM is now implicitly enabled if API key is provided.
	// Create NoOp client only if API key is truly missing.
	apiKey := app.Config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("NEUROSCRIPT_API_KEY") // Standardized env var name
		if apiKey == "" {
			app.Log.Debug("API key is missing in config and environment variable (NEUROSCRIPT_API_KEY). Creating NoOpLLMClient.") // Changed from Info
			return adapters.NewNoOpLLMClient(), nil
		}
		app.Log.Debug("Using LLM API key from environment variable NEUROSCRIPT_API_KEY.") // Changed from Info
	} else {
		app.Log.Debug("Using LLM API key from configuration.")
	}

	app.Log.Debug("Creating real LLMClient.") // Changed from Info
	apiHost := app.Config.APIHost
	modelName := app.Config.ModelName

	// Ensure logger is not nil before passing
	logger := app.Log
	if logger == nil {
		logger = adapters.NewNoOpLogger() // Safety fallback
	}

	// Use NewLLMClient from core package
	llmClient := core.NewLLMClient(apiKey, apiHost, modelName, logger, !app.Config.Insecure) // Pass insecure flag inversely
	if llmClient == nil {
		app.Log.Error("core.NewLLMClient returned nil unexpectedly.")
		return nil, fmt.Errorf("failed to create LLM client instance (core.NewLLMClient returned nil)")
	}

	app.Log.Debug("Real LLMClient created.", "host", apiHost, "model", modelName) // Changed from Info
	return llmClient, nil
}
