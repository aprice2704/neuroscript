// NeuroScript Version: 0.3.0
// File version: 0.1.2 // Corrected SLogAdapter name, corrected NewAIWorkerManager args order and types.
// Contains setup functions for the ng application.
// filename: cmd/ng/setup.go
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core" // Now directly used
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/neurogo" // Now directly used
)

// initializeLogger sets up the application's logger based on configuration.
func initializeLogger(levelStr string, filePath string) (logging.Logger, error) {

	level, err := adapters.LogLevelFromString(levelStr)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %q", levelStr)
	}

	var output io.Writer = os.Stderr
	var closer func() error // To store the file closer if needed

	if filePath != "" {
		dir := filepath.Dir(filePath)
		if mkDirErr := os.MkdirAll(dir, 0755); mkDirErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not create log directory %s: %v. Attempting to use file directly.\n", dir, mkDirErr)
		}

		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %q: %w", filePath, err)
		}
		output = f
		closer = f.Close // Store the closer method
	}

	// CORRECTED: Use NewSimpleSlogAdapter
	loggerAdapter, err := adapters.NewSimpleSlogAdapter(output, level)
	if err != nil {
		// Close the file if opened, even if adapter creation failed
		if closer != nil {
			_ = closer() // Ignore error on close during error path
		}
		return nil, fmt.Errorf("failed to create logger adapter: %w", err)
	}

	// Assuming NewSimpleSlogAdapter handles closing *os.File if output is one,
	// or that main() handles closing the logger if it implements io.Closer.
	return loggerAdapter, nil
}

// initializeCoreComponents creates and configures the core Interpreter, LLMClient, and AIWorkerManager.
func initializeCoreComponents(app *neurogo.App, logger logging.Logger, absSandboxDir string) (core.LLMClient, *core.Interpreter, *core.AIWorkerManager, error) {
	if app == nil || app.Config == nil {
		return nil, nil, nil, fmt.Errorf("initializeCoreComponents: app or app.Config is nil")
	}
	if logger == nil {
		return nil, nil, nil, fmt.Errorf("initializeCoreComponents: logger is nil")
	}

	llmClient, err := app.CreateLLMClient() // Uses app.Config internally
	if err != nil {
		logger.Error("Failed to create LLM client", "error", err)
		return nil, nil, nil, fmt.Errorf("failed to create LLM client: %w", err)
	}
	logger.Info("LLM Client created successfully")

	// Call NewInterpreter with the correct signature
	interpreter, err := core.NewInterpreter(logger, llmClient, app.Config.SandboxDir, nil, app.Config.LibPaths)
	if err != nil {
		logger.Error("Failed to create core Interpreter", "error", err)
		return nil, nil, nil, fmt.Errorf("failed to create core Interpreter: %w", err)
	}
	app.SetInterpreter(interpreter) // Store the interpreter in the App struct
	logger.Info("Core Interpreter created successfully")

	// Define paths for AI Worker Manager persistence files within the sandbox
	// These paths are NOT passed to NewAIWorkerManager anymore, but might be used
	// by tools later to load/save content.
	// definitionsFilePath := filepath.Join(absSandboxDir, neurogo.DefaultDefinitionsFilename)
	// performanceDataFilePath := filepath.Join(absSandboxDir, neurogo.DefaultPerformanceDataFilename)

	// CORRECTED: Argument order and types for NewAIWorkerManager:
	// logger, sandboxDir, llmClient, initialDefinitionsContent, initialPerformanceContent
	// Pass empty strings for initial content; loading from files is now handled by tools if needed.
	aiWm, err := core.NewAIWorkerManager(logger, app.Config.SandboxDir, llmClient, "", "")
	if err != nil {
		logger.Error("Failed to create AI Worker Manager", "error", err)
		return llmClient, interpreter, nil, fmt.Errorf("failed to create AI Worker Manager: %w", err)
	}
	if aiWm != nil {
		app.SetAIWorkerManager(aiWm) // Store the AIWM in the App struct
		logger.Info("AI Worker Manager created successfully")
	} else {
		logger.Warn("AI Worker Manager is nil after creation attempt without explicit error.")
	}

	return llmClient, interpreter, aiWm, nil
}
