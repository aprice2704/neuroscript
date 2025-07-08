// NeuroScript Version: 0.3.0
// File version: 0.0.4
// Corrected reference to GrammarVersion to use the 'lang' package.
// filename: pkg/neurogo/app_init.go
package neurogo

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // load tools
)

// NewApp creates and initializes a new App instance.
func NewApp(config *Config, logger interfaces.Logger, llmclient interfaces.LLMClient) (*App, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		logger = logging.NewNoOpLogger()
		logger.Warn("No logger provided to NewApp, using NoOpLogger.")
	}

	logger.Infof("NeuroScript grammar %s", lang.GrammarVersion) // Corrected reference

	appCtx, cancelFunc := context.WithCancel(context.Background())

	a := &App{
		Config:     config,
		Log:        logger,
		appCtx:     appCtx,
		cancelFunc: cancelFunc,
	}

	if llmclient != nil {
		a.llmClient = llmclient
		a.Log.Info("LLM Client provided to NewApp and assigned.")
	} else {
		a.Log.Debug("No LLM client provided to NewApp; will be created during core component initialization.")
	}

	a.Log.Debug("Basic App struct initialized. Core components (including LLMClient if not provided) will be initialized by Run->InitializeCoreComponents.")
	return a, nil
}
