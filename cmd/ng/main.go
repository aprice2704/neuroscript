// NeuroScript Version: 0.3.0
// File version: 7
// Purpose: A simple CLI tool to run NeuroScript files, replacing the old 'ng' tool.
// filename: cmd/ng/main.go
// nlines: 161

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aprice2704/neuroscript/pkg/api"

	// This blank import is crucial. It registers all standard NeuroScript tool
	// bundles so they are available to the interpreter.
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all"
)

// simpleLogger is a basic adapter to make the standard Go logger
// satisfy the interfaces.Logger requirement of the HostContext.
type simpleLogger struct {
	*log.Logger
}

// SetLevel is a no-op method to satisfy the interfaces.Logger interface.
// The standard logger does not have log levels, but the method must match the required signature.
func (l *simpleLogger) SetLevel(level api.LogLevel) {}

// Debug logs a debug message.
func (l *simpleLogger) Debug(msg string, args ...interface{}) {
	l.Printf("DEBUG: "+msg, args...)
}

// Debugf logs a formatted debug message.
func (l *simpleLogger) Debugf(format string, args ...interface{}) {
	l.Printf("DEBUG: "+format, args...)
}

// Info logs an info message.
func (l *simpleLogger) Info(msg string, args ...interface{}) {
	l.Printf("INFO: "+msg, args...)
}

// Infof logs a formatted info message.
func (l *simpleLogger) Infof(format string, args ...interface{}) {
	l.Printf("INFO: "+format, args...)
}

// Warn logs a warning message.
func (l *simpleLogger) Warn(msg string, args ...interface{}) {
	l.Printf("WARN: "+msg, args...)
}

// Warnf logs a formatted warning message.
func (l *simpleLogger) Warnf(format string, args ...interface{}) {
	l.Printf("WARN: "+format, args...)
}

// Error logs an error message.
func (l *simpleLogger) Error(msg string, args ...interface{}) {
	l.Printf("ERROR: "+msg, args...)
}

// Errorf logs a formatted error message.
func (l *simpleLogger) Errorf(format string, args ...interface{}) {
	l.Printf("ERROR: "+format, args...)
}

func main() {
	// 1. Parse command-line arguments.
	flag.Parse()
	scriptFiles := flag.Args()

	if len(scriptFiles) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: ng <file1.ns> [file2.ns] ...")
		os.Exit(1)
	}

	// 2. Set up a basic logger that satisfies the required interface.
	baseLogger := log.New(os.Stderr, "ng: ", log.LstdFlags)
	logger := &simpleLogger{Logger: baseLogger}

	// 3. Configure the HostContext for the interpreter, including the EmitFunc.
	hostCtx, err := api.NewHostContextBuilder().
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithStdin(os.Stdin).
		WithLogger(logger).
		WithEmitFunc(func(v api.Value) {
			// This function handles the 'emit' statement from the script.
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
		logger.Fatalf("Failed to build host context: %v", err)
	}

	// 4. Create a wildcard capability to grant all permissions.
	// The signature requires group and name; "*" for both creates a universal grant.
	allCaps := api.NewCapability("*", "*")

	// 5. Create a new NeuroScript interpreter instance using NewConfigInterpreter.
	interp := api.NewConfigInterpreter(
		[]string{"*"}, // Use the wildcard "*" to allow all tools.
		[]api.Capability{allCaps},
		api.WithHostContext(hostCtx),
	)

	// 6. Read, parse, and load each script file in append mode.
	for _, filename := range scriptFiles {
		src, err := os.ReadFile(filename)
		if err != nil {
			logger.Fatalf("Failed to read file %q: %v", filename, err)
		}

		tree, err := api.Parse(src, 0)
		if err != nil {
			logger.Fatalf("Failed to parse file %q: %v", filename, err)
		}

		if err := interp.AppendScript(tree); err != nil {
			logger.Fatalf("Failed to load definitions from %q: %v", filename, err)
		}
		logger.Info("Successfully loaded script: %s", filename)
	}

	// 7. Execute the 'command' blocks from the loaded scripts.
	logger.Info("Executing command blocks...")
	result, err := interp.ExecuteCommands()
	if err != nil {
		logger.Fatalf("Script execution failed: %v", err)
	}

	// 8. Print the final result if it's not the exported Nil value.
	if result != nil {
		unwrapped, err := api.Unwrap(result)
		if err != nil {
			logger.Fatalf("Failed to unwrap result value: %v", err)
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
