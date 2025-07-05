// In cmd/ng/setup.go

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/neurogo"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// initializeLogger sets up the application's logger based on configuration.
func initializeLogger(levelStr string, filePath string) (interfaces.Logger, error) {
	parsedLevel, parseErr := logging.LogLevelFromString(levelStr)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "[NEUROGO_LoggerInit_DIAG] Error parsing log level string '%s': %v. Using default log level INFO for diagnostics.\n", levelStr, parseErr)
		parsedLevel = interfaces.LogLevelInfo
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

	diagLogger, _ := logging.NewSimpleSlogAdapter(writer, parsedLevel)
	diagLogger.Debug("[NEUROGO_LoggerInit_DIAG] Diagnostic logger created. Attempting to initialize main logger.", "configured_level_str", levelStr, "parsed_level_for_diag", parsedLevel, "output_file_target", filePath)

	appLogger, appLoggerErr := logging.NewSimpleSlogAdapter(writer, parsedLevel)
	if appLoggerErr != nil {
		errMsg := fmt.Sprintf("Failed to create main application SlogAdapter: %v", appLoggerErr)
		diagLogger.Error(errMsg)

		if parseErr != nil {
			return logging.NewNoOpLogger(), fmt.Errorf("log level parsing error ('%s'): %w (and main logger creation also failed: %v)", levelStr, parseErr, appLoggerErr)
		}
		return logging.NewNoOpLogger(), appLoggerErr
	}

	appLogger.Info("Logger initialized", "level", parsedLevel, "output_target", ifElse(filePath != "", filePath, "stderr"))

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

// InitializeCoreComponents now only sets up the interpreter.
func InitializeCoreComponents(app *neurogo.App, logger interfaces.Logger, llmClient interfaces.LLMClient) (tool.Runtime, error) {
	if llmClient == nil {
		err := fmt.Errorf("InitializeCoreComponents received a nil LLM client")
		logger.Error(err.Error())
		return nil, err
	}
	logger.Debug("InitializeCoreComponents received LLMClient.")

	// Interpreter setup
	if app.Config.SandboxDir == "" {
		err := fmt.Errorf("sandbox directory is not configured in app.Config")
		logger.Error(err.Error())
		return nil, err
	}

	initialGlobals := make(map[string]interface{})
	initialIncludes := make([]string, 0)

	// Corrected: Using functional options to create the interpreter and handling single return value.
	interp := interpreter.NewInterpreter(
		interpreter.WithLogger(logger),
		interpreter.WithLLMClient(llmClient),
		interpreter.WithSandboxDir(app.Config.SandboxDir),
		interpreter.WithInitialGlobals(initialGlobals),
		interpreter.WithInitialIncludes(initialIncludes),
	)
	if interp == nil {
		err := fmt.Errorf("failed to create interpreter")
		logger.Error(err.Error())
		return nil, err
	}
	app.SetInterpreter(interp)
	logger.Debug("Interpreter created and sandbox set.", "sandbox_path", app.Config.SandboxDir)

	return interp, nil
}
