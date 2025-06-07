// NeuroScript Version: 0.3.0
// File version: 0.0.4 (Refactored to separate content processing from file I/O)
// Purpose: Provide helpers for loading and processing NeuroScript files and strings.
// filename: pkg/neurogo/app_scripting.go

package neurogo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// processNeuroScriptContent is the core logic for parsing a script string and loading it.
// The sourceName is used for logging and error reporting (e.g., a file path or "<string>").
func (a *App) processNeuroScriptContent(content, sourceName string, interp *core.Interpreter) ([]*core.Procedure, map[string]string, error) {
	if interp == nil {
		return nil, nil, fmt.Errorf("cannot process script from '%s': interpreter is nil", sourceName)
	}
	a.Log.Debug("Processing NeuroScript content.", "source", sourceName)

	parser := core.NewParserAPI(a.Log)
	parseResultTree, parseErr := parser.Parse(content)
	if parseErr != nil {
		a.Log.Error("Parsing failed.", "source", sourceName, "error", parseErr)
		return nil, nil, fmt.Errorf("parsing from %s failed: %w", sourceName, parseErr)
	}
	if parseResultTree == nil {
		a.Log.Error("Parsing returned nil result without errors.", "source", sourceName)
		return nil, nil, fmt.Errorf("internal parsing error: nil result for %s", sourceName)
	}
	a.Log.Debug("Parsing successful.", "source", sourceName)

	astBuilder := core.NewASTBuilder(a.Log)
	programAST, fileMetadata, buildErr := astBuilder.Build(parseResultTree)
	if buildErr != nil {
		a.Log.Error("AST building failed.", "source", sourceName, "error", buildErr)
		return nil, fileMetadata, fmt.Errorf("AST building for %s failed: %w", sourceName, buildErr)
	}
	if programAST == nil {
		a.Log.Error("AST building returned nil program without errors.", "source", sourceName)
		return nil, fileMetadata, fmt.Errorf("internal AST building error: nil program for %s", sourceName)
	}
	if fileMetadata == nil {
		fileMetadata = make(map[string]string)
	}
	a.Log.Debug("AST building successful.", "source", sourceName, "procedures", len(programAST.Procedures))

	definedProcs := []*core.Procedure{}
	if programAST.Procedures != nil {
		for name, proc := range programAST.Procedures {
			if proc == nil {
				a.Log.Warn("Skipping nil procedure found in AST map.", "name", name, "source", sourceName)
				continue
			}
			if err := interp.AddProcedure(*proc); err != nil {
				a.Log.Error("Failed to add procedure to interpreter.", "procedure", name, "source", sourceName, "error", err)
				return definedProcs, fileMetadata, fmt.Errorf("failed to add procedure '%s' from '%s': %w", name, sourceName, err)
			}
			a.Log.Debug("Added procedure to interpreter.", "procedure", name, "source", sourceName)
			definedProcs = append(definedProcs, proc)
		}
	}
	a.Log.Debug("Finished processing content.", "source", sourceName, "procedures_added", len(definedProcs))
	return definedProcs, fileMetadata, nil
}

// processNeuroScriptFile is now a thin wrapper that reads a file and calls processNeuroScriptContent.
func (a *App) processNeuroScriptFile(filePath string, interp *core.Interpreter) ([]*core.Procedure, map[string]string, error) {
	a.Log.Debug("Reading NeuroScript file.", "path", filePath)
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		a.Log.Error("Failed to read script file.", "path", filePath, "error", err)
		return nil, nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return a.processNeuroScriptContent(string(contentBytes), filePath, interp)
}

// loadLibraries processes all files specified in Config.LibPaths. (This function remains unchanged)
func (a *App) loadLibraries(interpreter *core.Interpreter) error {
	if interpreter == nil {
		return fmt.Errorf("cannot load libraries: interpreter is nil")
	}
	a.Log.Debug("Loading libraries from paths.", "paths", a.Config.LibPaths)
	for _, libPath := range a.Config.LibPaths {
		absPath, err := filepath.Abs(libPath)
		if err != nil {
			a.Log.Warn("Could not get absolute path for library path, skipping.", "path", libPath, "error", err)
			continue
		}
		a.Log.Debug("Processing library path.", "path", absPath)

		info, err := os.Stat(absPath)
		if err != nil {
			a.Log.Warn("Could not stat library path, skipping.", "path", absPath, "error", err)
			continue
		}

		if info.IsDir() {
			err := filepath.WalkDir(absPath, func(path string, d os.DirEntry, walkErr error) error {
				if walkErr != nil {
					a.Log.Warn("Error accessing path during library walk, skipping.", "path", path, "error", walkErr)
					return nil
				}
				if !d.IsDir() && strings.HasSuffix(d.Name(), ".ns") {
					a.Log.Debug("Processing library file.", "file", path)
					_, _, procErr := a.processNeuroScriptFile(path, interpreter)
					if procErr != nil {
						a.Log.Error("Failed to process library file, continuing...", "file", path, "error", procErr)
					}
				}
				return nil
			})
			if err != nil {
				a.Log.Error("Error walking library directory.", "path", absPath, "error", err)
			}
		} else if strings.HasSuffix(info.Name(), ".ns") {
			a.Log.Debug("Processing library file.", "file", absPath)
			_, _, procErr := a.processNeuroScriptFile(absPath, interpreter)
			if procErr != nil {
				a.Log.Error("Failed to process library file.", "file", absPath, "error", procErr)
			}
		} else {
			a.Log.Warn("Library path is not a directory or .ns file, skipping.", "path", absPath)
		}
	}
	a.Log.Debug("Finished loading libraries.")
	return nil
}

// ExecuteScriptFile loads and runs the target procedure of a given script file.
func (app *App) ExecuteScriptFile(ctx context.Context, scriptPath string) error {
	startTime := time.Now()
	app.Log.Debug("--- Executing Script File ---", "path", scriptPath)

	interpreter := app.GetInterpreter()
	if interpreter == nil {
		return fmt.Errorf("cannot execute script: interpreter is nil")
	}

	if err := app.loadLibraries(interpreter); err != nil {
		return fmt.Errorf("error loading libraries for script %s: %w", scriptPath, err)
	}

	// This function now correctly calls the refactored core logic.
	_, fileMeta, err := app.processNeuroScriptFile(scriptPath, interpreter)
	if err != nil {
		return fmt.Errorf("failed to process script %s: %w", scriptPath, err)
	}
	if fileMeta == nil {
		fileMeta = make(map[string]string)
	}

	// Determine procedure to run
	procedureToRun := app.Config.TargetArg
	if procedureToRun == "" {
		if metaTarget, ok := fileMeta["target"]; ok && metaTarget != "" {
			procedureToRun = metaTarget
		} else {
			procedureToRun = "main"
		}
	}

	// Prepare arguments
	scriptCLIArgs := app.Config.ProcArgs
	interpreterArgs := make([]interface{}, len(scriptCLIArgs))
	for i, argStr := range scriptCLIArgs {
		interpreterArgs[i] = argStr
	}

	// Execute the procedure
	app.Log.Debug("Executing procedure.", "name", procedureToRun)
	_, runErr := interpreter.RunProcedure(procedureToRun, interpreterArgs...)
	if runErr != nil {
		return fmt.Errorf("error executing procedure '%s' in script '%s': %w", procedureToRun, scriptPath, runErr)
	}

	totalDuration := time.Since(startTime)
	app.Log.Debug("--- Script File Execution Finished ---", "path", scriptPath, "total_duration", totalDuration)
	return nil
}

// In pkg/neurogo/app.go or app_script.go

// ExecuteScriptString parses and runs the target procedure of a given script string.
func (app *App) ExecuteScriptString(ctx context.Context, scriptName, scriptContent string, initialVars map[string]interface{}) (any, error) {
	startTime := time.Now()
	app.Log.Debug("--- Executing Script String ---", "name", scriptName)

	interpreter := app.GetInterpreter()
	if interpreter == nil {
		return nil, fmt.Errorf("cannot execute script: interpreter is nil")
	}

	if err := app.loadLibraries(interpreter); err != nil {
		return nil, fmt.Errorf("error loading libraries for script '%s': %w", scriptName, err)
	}

	// CORRECTED LINE: Handle all 3 return values from the content processor.
	_, fileMeta, err := app.processNeuroScriptContent(scriptContent, scriptName, interpreter)
	if err != nil {
		return nil, err // The error from processNeuroScriptContent is already descriptive
	}
	if fileMeta == nil {
		fileMeta = make(map[string]string)
	}

	if initialVars != nil {
		for key, value := range initialVars {
			interpreter.SetVariable(key, value)
		}
	}

	procedureToRun := app.Config.TargetArg
	if procedureToRun == "" {
		if metaTarget, ok := fileMeta["target"]; ok && metaTarget != "" {
			procedureToRun = metaTarget
		} else {
			procedureToRun = "main"
		}
	}

	interpreterArgs := make([]interface{}, len(app.Config.ProcArgs))
	for i, argStr := range app.Config.ProcArgs {
		interpreterArgs[i] = argStr
	}

	results, runErr := interpreter.RunProcedure(procedureToRun, interpreterArgs...)
	if runErr != nil {
		return nil, fmt.Errorf("error executing procedure '%s' from script '%s': %w", procedureToRun, scriptName, runErr)
	}

	app.Log.Info("Script executed successfully.", "script_name", scriptName, "procedure", procedureToRun, "duration", time.Since(startTime))
	return results, nil
}
