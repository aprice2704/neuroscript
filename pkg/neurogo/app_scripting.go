// NeuroScript Version: 0.3.1
// File version: 0.4.2
// Purpose: Fixes incorrect return value counts in error paths within processNeuroScriptContent.
// filename: pkg/neurogo/app_scripting.go

package neurogo

import (
	"context"
	"fmt"
	"time"
)

// processNeuroScriptContent is the core logic for parsing a script string and loading its definitions.
func (a *App) processNeuroScriptContent(content, sourceName string, interp *Interpreter) (map[string]string, error) {
	if interp == nil {
		return nil, fmt.Errorf("cannot process script from '%s': interpreter is nil", sourceName)
	}
	a.Log.Debug("Processing NeuroScript content.", "source", sourceName)

	parser := arserAPI(a.Log)
	parseResultTree, parseErr := parser.Parse(content)
	if parseErr != nil {
		a.Log.Error("Parsing failed.", "source", sourceName, "error", parseErr)
		return nil, fmt.Errorf("parsing from %s failed: %w", sourceName, parseErr)
	}
	if parseResultTree == nil {
		a.Log.Error("Parsing returned nil result without errors.", "source", sourceName)
		return nil, fmt.Errorf("internal parsing error: nil result for %s", sourceName)
	}
	a.Log.Debug("Parsing successful.", "source", sourceName)

	astBuilder := STBuilder(a.Log)
	programAST, fileMetadata, buildErr := astBuilder.Build(parseResultTree)
	if buildErr != nil {
		a.Log.Error("AST building failed.", "source", sourceName, "error", buildErr)
		return fileMetadata, fmt.Errorf("AST building for %s failed: %w", sourceName, buildErr)
	}
	if programAST == nil {
		a.Log.Error("AST building returned nil program without errors.", "source", sourceName)
		return fileMetadata, fmt.Errorf("internal AST building error: nil program for %s", sourceName)
	}
	if fileMetadata == nil {
		fileMetadata = make(map[string]string)
	}
	a.Log.Debug("AST building successful.", "source", sourceName, "procedures", len(programAST.Procedures))

	if programAST.Procedures != nil {
		for name, proc := range programAST.Procedures {
			if proc == nil {
				a.Log.Warn("Skipping nil procedure found in AST map.", "name", name, "source", sourceName)
				continue
			}
			if err := interp.AddProcedure(*proc); err != nil {
				a.Log.Error("Failed to add procedure to interpreter.", "procedure", name, "source", sourceName, "error", err)
				return fileMetadata, fmt.Errorf("failed to add procedure '%s' from '%s': %w", name, sourceName, err)
			}
			a.Log.Debug("Added procedure to interpreter.", "procedure", name, "source", sourceName)
		}
	}
	a.Log.Debug("Finished processing content.", "source", sourceName, "procedures_added", len(programAST.Procedures))
	return fileMetadata, nil
}

// LoadScriptString parses a script from a string and loads its function definitions
// into the interpreter without executing any top-level code. It is agnostic of the script's origin.
func (app *App) LoadScriptString(ctx context.Context, scriptContent string) (map[string]string, error) {
	app.Log.Debug("--- Loading Script String ---")

	interpreter := app.GetInterpreter()
	if interpreter == nil {
		return nil, fmt.Errorf("cannot load script: interpreter is nil")
	}

	// Use a generic source name as this function only deals with content.
	fileMeta, err := app.processNeuroScriptContent(scriptContent, "<string>", interpreter)
	if err != nil {
		return nil, err
	}

	app.Log.Debug("Script loaded successfully. No execution.")
	return fileMeta, nil
}

// RunProcedure explicitly executes a procedure that has been loaded into the interpreter.
func (app *App) RunProcedure(ctx context.Context, procedureToRun string, scriptArgs []string) (any, error) {
	startTime := time.Now()
	interpreter := app.GetInterpreter()
	if interpreter == nil {
		return nil, fmt.Errorf("cannot run procedure: interpreter is nil")
	}

	if procedureToRun == "" {
		return nil, fmt.Errorf("cannot run procedure: procedure name is empty")
	}

	wrappedArgs := make([]e, len(scriptArgs))
	for i, argStr := range scriptArgs {
		wrapped, err := (argStr)
		if err != nil {
			return nil, fmt.Errorf("failed to wrap script argument '%s': %w", argStr, err)
		}
		wrappedArgs[i] = wrapped
	}

	app.Log.Debug("Executing procedure.", "name", procedureToRun)
	results, runErr := interpreter.RunProcedure(procedureToRun, wrappedArgs...)
	if runErr != nil {
		return nil, fmt.Errorf("error executing procedure '%s': %w", procedureToRun, runErr)
	}

	app.Log.Info("Procedure executed successfully.", "procedure", procedureToRun, "duration", time.Since(startTime))
	return results, nil
}
