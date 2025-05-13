// NeuroScript Version: 0.3.0
// File version: 0.1.5
// Initialize AIWorkerManager with default definitions and dedicated sandbox subdir.
// Corrected NewInterpreter and NewAIWorkerManager calls based on user's setup.go.
// Added diagnostic print for logger initialization.
// filename: cmd/ng/setup.go
// nlines: 95 // Approximate
// risk_rating: MEDIUM
package main

import (
	"fmt"
	"io"
	"log/slog" // Import slog for slog.Level
	"os"
	"path/filepath"
	"strings" // Import strings for LogLevelFromString

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/neurogo"
)

// initializeLogger sets up the application's logger based on configuration.
func initializeLogger(levelStr string, filePath string) (logging.Logger, error) {
	level, err := adapters.LogLevelFromString(levelStr)
	if err != nil {
		// Fallback for diagnostic print if adapters.LogLevelFromString fails
		fmt.Fprintf(os.Stderr, "[NEUROGO_LoggerInit_DIAG] Error parsing log level string '%s': %v. Defaulting to INFO for diagnostic print.\n", levelStr, err)
		// Attempt to provide a default slog.Level for the diagnostic print, or make it conditional
		//	var tempSlogLvl slog.Level = slog.LevelInfo // Default for safety
		//	fmt.Fprintf(os.Stderr, "[NEUROGO_LoggerInit_DIAG] LogLevelString: '%s', Parsed logging.LogLevel: %v (error path), Effective Slog Level for setup: %s\n", levelStr, level, tempSlogLvl.String())
		return nil, fmt.Errorf("invalid log level: %q: %w", levelStr, err)
	}

	// Diagnostic print to os.Stderr before logger is created
	var slogEquivalentLevel slog.Level
	switch level {
	case logging.LogLevelDebug:
		slogEquivalentLevel = slog.LevelDebug
	case logging.LogLevelInfo:
		slogEquivalentLevel = slog.LevelInfo
	case logging.LogLevelWarn:
		slogEquivalentLevel = slog.LevelWarn
	case logging.LogLevelError:
		slogEquivalentLevel = slog.LevelError
	default:
		slogEquivalentLevel = slog.LevelInfo // Should not happen if LogLevelFromString is robust
	}
	fmt.Fprintf(os.Stderr, "[NEUROGO_LoggerInit_DIAG] LogLevelString: '%s', Parsed logging.LogLevel: %v, Effective Slog Level for setup: %s\n", strings.ToLower(levelStr), level, slogEquivalentLevel.String())

	var output io.Writer = os.Stderr
	var fileCloser io.Closer

	if filePath != "" {
		dir := filepath.Dir(filePath)
		if dir != "" && dir != "." { // Ensure directory creation only if a directory path is present
			if mkDirErr := os.MkdirAll(dir, 0755); mkDirErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not create log directory %s: %v. Attempting to use file directly.\n", dir, mkDirErr)
			}
		}

		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %s: %w", filePath, err)
		}
		output = file
		fileCloser = file // Store for closing
	}

	// Use NewSimpleSlogAdapter from pkg/adapters
	logger, _ := adapters.NewSimpleSlogAdapter(output, level)

	// If logging to a file, return a logger that also handles closing the file.
	if fileCloser != nil {
		return struct {
			logging.Logger
			io.Closer
		}{logger, fileCloser}, nil
	}

	return logger, nil
}

// initializeCoreComponents sets up the interpreter and AI worker manager.
func initializeCoreComponents(app *neurogo.App, logger logging.Logger, llmClient core.LLMClient) (*core.Interpreter, *core.AIWorkerManager, error) {
	if app == nil || app.Config == nil {
		return nil, nil, fmt.Errorf("application or application config is nil, cannot initialize core components")
	}
	if logger == nil {
		return nil, nil, fmt.Errorf("logger is nil, cannot initialize core components")
	}
	// LLM Client is now passed in, so we don't create it here.
	// if llmClient == nil, some operations might fail, but core components can still init.

	interpreter, err := core.NewInterpreter(logger, llmClient, app.Config.SandboxDir,
		map[string]interface{}{}, []string{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create interpreter: %w", err)
	}
	app.SetInterpreter(interpreter) // Store interpreter in app
	logger.Debug("Interpreter created and sandbox set.", "sandbox_path", app.Config.SandboxDir)

	// AI Worker Manager setup
	aiWmSandboxDir := filepath.Join(app.Config.SandboxDir, ".neuroscript_aiwm")
	if err := os.MkdirAll(aiWmSandboxDir, 0750); err != nil {
		logger.Error("Failed to create AIWorkerManager sandbox subdirectory", "path", aiWmSandboxDir, "error", err)
		return interpreter, nil, fmt.Errorf("failed to create AIWorkerManager sandbox directory '%s': %w", aiWmSandboxDir, err)
	}
	logger.Debug("AIWorkerManager sandbox directory ensured", "path", aiWmSandboxDir)

	aiWm, err := core.NewAIWorkerManager(logger, aiWmSandboxDir, llmClient, core.AIWorkerDefinitions_Default, "")
	if err != nil {
		logger.Error("Failed to create AI Worker Manager", "error", err)
		return interpreter, nil, fmt.Errorf("failed to create AI Worker Manager: %w", err)
	}
	if aiWm != nil {
		app.SetAIWorkerManager(aiWm)
		// This Info log is appropriate as it's a summary of component initialization.
		logger.Infof("AI Worker Manager available.")
		defs := aiWm.ListWorkerDefinitions(nil)
		logger.Debug("AIWorkerManager initial definitions loaded check", "count", len(defs))
	} else {
		logger.Warn("AI Worker Manager could not be initialized.")
	}

	return interpreter, aiWm, nil
}
