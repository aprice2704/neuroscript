// NeuroScript Version: 0.3.0
// File version: 10
// Purpose: A simple CLI tool to run NeuroScript files with slog-based logging.
// filename: cmd/ng/main.go
// nlines: 181

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/api"

	// This blank import is crucial. It registers all standard NeuroScript tool
	// bundles so they are available to the interpreter.
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all"
)

// slogAdapter makes the standard slog.Logger compatible with the
// NeuroScript interfaces.Logger interface.
type slogAdapter struct {
	*slog.Logger
}

// SetLevel is a pass-through to the underlying slog handler's level.
// Note: This is a simple implementation; a more robust one might
// require a custom slog.Handler to allow dynamic level changes.
func (l *slogAdapter) SetLevel(level api.LogLevel) {
	// This is intentionally a no-op for this simple adapter,
	// as changing the level of the default slog handler after creation is complex.
	// The level is set at initialization.
}

func (l *slogAdapter) Debug(msg string, args ...interface{}) {
	l.Logger.Debug(msg, args...)
}

func (l *slogAdapter) Debugf(format string, args ...interface{}) {
	l.Logger.Debug(fmt.Sprintf(format, args...))
}

func (l *slogAdapter) Info(msg string, args ...interface{}) {
	l.Logger.Info(msg, args...)
}

func (l *slogAdapter) Infof(format string, args ...interface{}) {
	l.Logger.Info(fmt.Sprintf(format, args...))
}

func (l *slogAdapter) Warn(msg string, args ...interface{}) {
	l.Logger.Warn(msg, args...)
}

func (l *slogAdapter) Warnf(format string, args ...interface{}) {
	l.Logger.Warn(fmt.Sprintf(format, args...))
}

func (l *slogAdapter) Error(msg string, args ...interface{}) {
	l.Logger.Error(msg, args...)
}

func (l *slogAdapter) Errorf(format string, args ...interface{}) {
	l.Logger.Error(fmt.Sprintf(format, args...))
}

func main() {
	// 1. Define and parse command-line arguments.
	logLevelFlag := flag.String("loglevel", "error", "Set the log level: debug, info, warn, error")
	flag.Parse()
	scriptFiles := flag.Args()

	if len(scriptFiles) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: ng [-loglevel <level>] <file1.ns> [file2.ns] ...")
		os.Exit(1)
	}

	// 2. Set up the slog logger based on the command-line flag.
	var level slog.Level
	switch strings.ToLower(*logLevelFlag) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		fmt.Fprintf(os.Stderr, "Invalid log level %q. Defaulting to 'error'.\n", *logLevelFlag)
		level = slog.LevelError
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slogger := slog.New(handler)
	logger := &slogAdapter{Logger: slogger}

	// 3. Configure the HostContext for the interpreter.
	hostCtx, err := api.NewHostContextBuilder().
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithStdin(os.Stdin).
		WithLogger(logger).
		WithEmitFunc(func(v api.Value) {
			unwrapped, err := api.Unwrap(v)
			if err != nil {
				logger.Errorf("emit: failed to unwrap value: %v", err)
				return
			}
			if unwrapped != nil {
				if str, ok := unwrapped.(string); ok {
					fmt.Fprintln(os.Stdout, str)
				} else {
					fmt.Fprintf(os.Stdout, "%v\n", unwrapped)
				}
			}
		}).
		Build()
	if err != nil {
		slogger.Error("Failed to build host context", "error", err)
		os.Exit(1)
	}

	// 4. Create a wildcard capability to grant all permissions.
	allCaps := api.NewCapability("*", "*")

	// 5. Create a new NeuroScript interpreter instance.
	interp := api.NewConfigInterpreter(
		[]string{"*"}, // Allow all tools
		[]api.Capability{allCaps},
		api.WithHostContext(hostCtx),
	)
	interp.SetTurnContext(context.Background())

	// 6. Read, parse, and load each script file in append mode.
	for _, filename := range scriptFiles {
		src, err := os.ReadFile(filename)
		if err != nil {
			logger.Errorf("Failed to read file %q: %v", filename, err)
			os.Exit(1)
		}

		tree, err := api.Parse(src, 0)
		if err != nil {
			logger.Errorf("Failed to parse file %q: %v", filename, err)
			os.Exit(1)
		}

		if err := interp.AppendScript(tree); err != nil {
			logger.Errorf("Failed to load definitions from %q: %v", filename, err)
			os.Exit(1)
		}
		logger.Infof("Successfully loaded script: %s", filename)
	}

	// 7. Execute the 'command' blocks from the loaded scripts.
	logger.Info("Executing command blocks...")
	result, err := interp.ExecuteCommands()
	if err != nil {
		logger.Errorf("Script execution failed: %v", err)
		os.Exit(1)
	}

	// 8. Print the final result if it's not nil.
	if result != nil {
		unwrapped, err := api.Unwrap(result)
		if err != nil {
			logger.Errorf("Failed to unwrap result value: %v", err)
			os.Exit(1)
		}
		if unwrapped != nil {
			if str, ok := unwrapped.(string); ok {
				fmt.Fprintln(os.Stdout, str)
			} else {
				fmt.Fprintf(os.Stdout, "%v\n", unwrapped)
			}
		}
	}

	logger.Info("Execution finished successfully.")
}
