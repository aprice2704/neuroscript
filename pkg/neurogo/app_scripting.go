// NeuroScript Version: 0.3.0
// File version: 0.0.3
// Purpose: Correctly pass -arg CLI arguments to NeuroScript procedures using interpreter.RunProcedure signature from core_index.json.
// filename: pkg/neurogo/app_scripting.go
// nlines: 166
// risk_rating: MEDIUM
package neurogo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aprice2704/neuroscript/pkg/core"
	// logging.Logger is used via a.Log
	// models.Schema is not used by these functions
)

// processNeuroScriptFile parses a .ns file and adds its procedures to the interpreter.
// Returns the list of procedures defined in THIS file and the file's metadata.
func (a *App) processNeuroScriptFile(filePath string, interp *core.Interpreter) ([]*core.Procedure, map[string]string, error) {
	if interp == nil {
		return nil, nil, fmt.Errorf("cannot process script file '%s': interpreter is nil", filePath)
	}
	a.Log.Debug("Processing NeuroScript file.", "path", filePath)
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		a.Log.Error("Failed to read script file.", "path", filePath, "error", err)
		return nil, nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	content := string(contentBytes)

	parser := core.NewParserAPI(a.Log)
	parseResultTree, parseErr := parser.Parse(content)
	if parseErr != nil {
		a.Log.Error("Parsing failed.", "path", filePath, "error", parseErr)
		return nil, nil, fmt.Errorf("parsing file %s failed: %w", filePath, parseErr)
	}
	if parseResultTree == nil {
		a.Log.Error("Parsing returned nil result without errors.", "path", filePath)
		return nil, nil, fmt.Errorf("internal parsing error: nil result for %s", filePath)
	}
	a.Log.Debug("Parsing successful.", "path", filePath)

	astBuilder := core.NewASTBuilder(a.Log)
	programAST, fileMetadata, buildErr := astBuilder.Build(parseResultTree)
	if buildErr != nil {
		a.Log.Error("AST building failed.", "path", filePath, "error", buildErr)
		return nil, fileMetadata, fmt.Errorf("AST building for %s failed: %w", filePath, buildErr)
	}
	if programAST == nil {
		a.Log.Error("AST building returned nil program without errors.", "path", filePath)
		return nil, fileMetadata, fmt.Errorf("internal AST building error: nil program for %s", filePath)
	}
	if fileMetadata == nil {
		fileMetadata = make(map[string]string)
	}
	a.Log.Debug("AST building successful.", "path", filePath, "procedures", len(programAST.Procedures), "metadata_keys", len(fileMetadata))

	definedProcs := []*core.Procedure{}
	if programAST.Procedures != nil {
		for name, proc := range programAST.Procedures {
			if proc == nil {
				a.Log.Warn("Skipping nil procedure found in AST map.", "name", name, "path", filePath)
				continue
			}
			if err := interp.AddProcedure(*proc); err != nil {
				a.Log.Error("Failed to add procedure to interpreter.", "procedure", name, "path", filePath, "error", err)
				return definedProcs, fileMetadata, fmt.Errorf("failed to add procedure '%s' from '%s': %w", name, filePath, err)
			}
			a.Log.Debug("Added procedure to interpreter.", "procedure", name, "path", filePath)
			definedProcs = append(definedProcs, proc)
		}
	}
	a.Log.Debug("Finished processing file.", "path", filePath, "procedures_added", len(definedProcs), "metadata_keys", len(fileMetadata))
	return definedProcs, fileMetadata, nil
}

// loadLibraries processes all files specified in Config.LibPaths.
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

// ExecuteScriptFile loads and runs the target procedure of a given script file, passing arguments.
func (app *App) ExecuteScriptFile(ctx context.Context, scriptPath string) error { // ctx is kept for potential future use, but not passed to RunProcedure
	startTime := time.Now()
	app.Log.Debug("--- Executing Script File ---", "path", scriptPath, "target_flag", app.Config.TargetArg, "args_flag", app.Config.ProcArgs)

	interpreter := app.GetInterpreter()
	if interpreter == nil {
		app.Log.Error("Interpreter is not initialized.")
		return fmt.Errorf("cannot execute script: interpreter is nil")
	}

	if err := app.loadLibraries(interpreter); err != nil {
		app.Log.Error("Failed to load libraries before executing script", "error", err)
		return fmt.Errorf("error loading libraries for script %s: %w", scriptPath, err)
	}

	_, fileMeta, err := app.processNeuroScriptFile(scriptPath, interpreter)
	if err != nil {
		return fmt.Errorf("failed to process script %s: %w", scriptPath, err)
	}
	if fileMeta == nil {
		fileMeta = make(map[string]string)
	}

	procedureToRun := app.Config.TargetArg
	if procedureToRun == "" {
		if metaTarget, ok := fileMeta["target"]; ok && metaTarget != "" {
			procedureToRun = metaTarget
			app.Log.Debug("Using target procedure from script metadata.", "procedure", procedureToRun)
		} else {
			procedureToRun = "main"
			app.Log.Debug("No target specified via flag or metadata, defaulting to 'main'.")
		}
	} else {
		app.Log.Debug("Using target procedure from -target flag.", "procedure", procedureToRun)
	}

	scriptCLIArgs := app.Config.ProcArgs
	interpreterArgs := make([]interface{}, len(scriptCLIArgs))
	for i, argStr := range scriptCLIArgs {
		interpreterArgs[i] = argStr
	}

	app.Log.Debug("Executing procedure.", "name", procedureToRun, "args_count", len(interpreterArgs), "args", interpreterArgs)
	execStartTime := time.Now()

	// According to core_index.json, Interpreter.RunProcedure signature is:
	// RunProcedure(procName string, args ...interface{}) (interface{}, error)
	// The context `ctx` is not passed here.
	results, runErr := interpreter.RunProcedure(procedureToRun, interpreterArgs...) //

	execEndTime := time.Now()
	duration := execEndTime.Sub(execStartTime)
	app.Log.Debug("Procedure execution finished.", "name", procedureToRun, "duration", duration)

	if runErr != nil {
		app.Log.Error("Script execution failed.", "procedure", procedureToRun, "error", runErr)
		return fmt.Errorf("error executing procedure '%s' in script '%s': %w", procedureToRun, scriptPath, runErr)
	}

	app.Log.Debug("Script executed successfully.", "procedure", procedureToRun)
	if results != nil {
		app.Log.Debug("Script Result Value", "result", fmt.Sprintf("%+v", results))
	} else {
		app.Log.Debug("Script Result Value: nil")
	}

	totalDuration := time.Since(startTime)
	app.Log.Debug("--- Script File Execution Finished ---", "path", scriptPath, "total_duration", totalDuration)
	return nil
}
