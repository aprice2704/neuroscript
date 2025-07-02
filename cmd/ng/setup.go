// In cmd/ng/setup.go

package main

import (
	"fmt"
	"io" // Import slog for slog.Level explicit casting if necessary
	"os"
	"path/filepath"

	// Keep for InitializeCoreComponents
	// "strings" // Not directly used in this corrected function

	"github.com/aprice2704/neuroscript/pkg/adapters" // Keep for InitializeCoreComponents
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/neurogo"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// initializeLogger sets up the application's logger based on configuration.
func initializeLogger(levelStr string, filePath string) (interfaces.Logger, error) {
	parsedLevel, parseErr := adapters.LogLevelFromString(levelStr)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "[NEUROGO_LoggerInit_DIAG] Error parsing log level string '%s': %v. Using default log level INFO for diagnostics.\n", levelStr, parseErr)
		parsedLevel = interfaces.LogLevelInfo // Use a default for the diagnostic logger to function
	}

	var writer io.Writer = os.Stderr
	if filePath != "" {
		// #nosec G304 -- File path is user-supplied.
		f, errOpen := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if errOpen != nil {
			fmt.Fprintf(os.Stderr, "[NEUROGO_LoggerInit_DIAG] Error opening log file '%s': %v. Using stderr.\n", filePath, errOpen)
		} else {
			writer = f
		}
	}

	// adapters.NewSimpleSlogAdapter returns interfaces.Logger (and no error)
	diagLogger, _ := adapters.NewSimpleSlogAdapter(writer, parsedLevel)
	// For logging the level, pass 'parsedLevel' directly. slog handles its Level types.
	// If a string is explicitly needed elsewhere: slog.Level(parsedLevel).String()
	diagLogger.Debug("[NEUROGO_LoggerInit_DIAG] Diagnostic logger created. Attempting to initialize main logger.", "configured_level_str", levelStr, "parsed_level_for_diag", parsedLevel, "output_file_target", filePath)

	// adapters.NewSlogAdapter returns (*SlogAdapter, error)
	// *SlogAdapter implements interfaces.Logger
	appLogger, appLoggerErr := adapters.NewSimpleSlogAdapter(writer, parsedLevel)
	if appLoggerErr != nil {
		errMsg := fmt.Sprintf("Failed to create main application SlogAdapter: %v", appLoggerErr)
		diagLogger.Error(errMsg) // Use diagLogger to report this failure

		// Return the most relevant error.
		if parseErr != nil {
			// Return a NoOpLogger if main logger creation failed, to satisfy the interface.
			return logging.NewNoOpLogger(), fmt.Errorf("log level parsing error ('%s'): %w (and main logger creation also failed: %v)", levelStr, parseErr, appLoggerErr)
		}
		return logging.NewNoOpLogger(), appLoggerErr
	}

	// Log successful initialization. Pass 'parsedLevel' directly.
	appLogger.Info("Logger initialized", "level", parsedLevel, "output_target", ifElse(filePath != "", filePath, "stderr"))

	// If there was an initial parsing error for the log level, return that error
	// so the application knows the configuration wasn't fully respected.
	if parseErr != nil {
		return appLogger, fmt.Errorf("log level configuration error ('%s' was invalid, used default): %w", levelStr, parseErr)
	}

	return appLogger, nil
}

// ifElse helper function
func ifElse(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func InitializeCoreComponents(app *neurogo.App, logger interfaces.Logger, llmClient interfaces.LLMClient) (tool.RunTime, *runtime.AIWorkerManager, error) {
	// LLM Client is now passed in as an argument and should already be set on the App instance by NewApp.
	// No need to create or set it here.
	if llmClient == nil {
		// This should ideally be caught earlier in main.go after app.CreateLLMClient()
		err := fmt.Errorf("InitializeCoreComponents received a nil LLM client")
		logger.Error(err.Error())
		return nil, nil, err
	}
	logger.Debug("InitializeCoreComponents received LLMClient.")

	// Interpreter setup
	if app.Config.SandboxDir == "" {
		err := fmt.Errorf("sandbox directory is not configured in app.Config")
		logger.Error(err.Error())
		return nil, nil, err // Return nil for AIWM as well if this fails
	}

	initialGlobals := make(map[string]interface{})
	initialIncludes := make([]string, 0)

	interpreter, err := nterpreter(logger, llmClient, app.Config.SandboxDir, initialGlobals, initialIncludes)
	if err != nil {
		logger.Error("Failed to create  rpreter", "error", err)
		return nil, nil, fmt.Errorf("failed to create interpreter: %w", err)
	}
	app.SetInterpreter(interpreter)
	logger.Debug("Interpreter created and sandbox set.", "sandbox_path", app.Config.SandboxDir)

	// AI Worker Manager setup
	aiWmSandboxDir := filepath.Join(app.Config.SandboxDir, ".neuroscript_aiwm")
	if err := os.MkdirAll(aiWmSandboxDir, 0750); err != nil {
		logger.Error("Failed to create AIWorkerManager sandbox subdirectory", "path", aiWmSandboxDir, "error", err)
		return interpreter, nil, fmt.Errorf("failed to create AIWorkerManager sandbox directory '%s': %w", aiWmSandboxDir, err)
	}
	logger.Debug("AIWorkerManager sandbox directory ensured", "path", aiWmSandboxDir)

	defaultDefsContent := ""
	if rkerDefinitions_Default != "" {
		defaultDefsContent = rkerDefinitions_Default
	}

	aiWm, errManager := IWorkerManager(logger, aiWmSandboxDir, llmClient, defaultDefsContent, "")
	if errManager != nil {
		logger.Error("Failed to create AI Worker Manager", "error", errManager)
		// Interpreter was created, but AIWM failed.
		return interpreter, nil, fmt.Errorf("failed to create AI Worker Manager: %w", errManager)
	}
	app.SetAIWorkerManager(aiWm)
	logger.Infof("AI Worker Manager initialized and available.")
	// defs := aiWm.ListWorkerDefinitions(nil) // Assuming ListWorkerDefinitions doesn't require an arg or takes nil
	// logger.Debug("AI Worker definitions loaded", "count", len(defs))

	return interpreter, aiWm, nil
}
