// NeuroScript Version: 0.3.0
// File version: 0.0.1
// Moved script processing and execution logic from app.go
// filename: pkg/neurogo/app_scripting.go
// nlines: 162 // Based on original line counts of moved methods
// risk_rating: LOW // Verbatim move of existing code
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
// Moved here from app_script.go
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
	parseResultTree, parseErr := parser.Parse(content) // Matched to the provided app.go
	if parseErr != nil {
		a.Log.Error("Parsing failed.", "path", filePath, "error", parseErr)
		// The provided app.go returns a single error, so we adapt if parseErr is a slice
		// For now, assuming parseErr is a single error object as per its usage here.
		// If parser.Parse returns []*core.ErrorNode as per index, this part needs adjustment
		// based on how app.go intends to handle it.
		// Reverting to exactly match the provided app.go structure for parseErr:
		return nil, nil, fmt.Errorf("parsing file %s failed: %w", filePath, parseErr)
	}
	if parseResultTree == nil {
		a.Log.Error("Parsing returned nil result without errors.", "path", filePath)
		return nil, nil, fmt.Errorf("internal parsing error: nil result for %s", filePath)
	}
	a.Log.Debug("Parsing successful.", "path", filePath)

	astBuilder := core.NewASTBuilder(a.Log) // Matched to the provided app.go
	// Matched to the provided app.go:
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
			// Matched to the provided app.go: if err := interp.AddProcedure(*proc); err != nil {
			// This assumes AddProcedure takes a core.Procedure value.
			// The index shows AddProcedure(name string, proc *Procedure) error.
			// For strict adherence to *verbatim move*, I will keep what's in the uploaded app.go.
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
// Moved here from app_script.go
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
		a.Log.Debug("Processing library path.", "path", absPath) // Changed from Info

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

// ExecuteScriptFile loads and runs the main procedure of a given script file.
func (app *App) ExecuteScriptFile(ctx context.Context, scriptPath string) error {
	startTime := time.Now()
	app.Log.Debug("--- Executing Script File ---", "path", scriptPath) // Changed from Info

	interpreter := app.GetInterpreter() // Use safe getter
	if interpreter == nil {
		app.Log.Error("Interpreter is not initialized.")
		return fmt.Errorf("cannot execute script: interpreter is nil")
	}

	// Load libraries first (idempotent, but ensures they are loaded if not already)
	if err := app.loadLibraries(interpreter); err != nil {
		app.Log.Error("Failed to load libraries before executing script", "error", err)
		// Decide whether to proceed or return error
		return fmt.Errorf("error loading libraries for script %s: %w", scriptPath, err)
	}

	// Process the main script file
	_, fileMeta, err := app.processNeuroScriptFile(scriptPath, interpreter)
	if err != nil {
		return fmt.Errorf("failed to process script %s: %w", scriptPath, err)
	}
	if fileMeta == nil {
		fileMeta = make(map[string]string)
		app.Log.Warn("No file metadata available from script file processing.", "path", scriptPath)
	}

	// Determine target procedure: flag -> metadata -> default 'main'
	procedureToRun := app.Config.TargetArg
	if procedureToRun == "" {
		if metaTarget, ok := fileMeta["target"]; ok && metaTarget != "" {
			procedureToRun = metaTarget
			app.Log.Debug("Using target procedure from script metadata.", "procedure", procedureToRun) // Changed from Info
		} else {
			procedureToRun = "main"
			app.Log.Debug("No target specified via flag or metadata, defaulting to 'main'.") // Changed from Info
		}
	} else {
		app.Log.Debug("Using target procedure from -target flag.", "procedure", procedureToRun) // Changed from Info
	}

	// Prepare arguments (simple map for now)
	procArgsMap := make(map[string]interface{})
	for i, arg := range app.Config.ProcArgs {
		procArgsMap[fmt.Sprintf("arg%d", i+1)] = arg
	}
	if app.Config.TargetArg != "" && procedureToRun == app.Config.TargetArg {
		// Avoid duplicating target arg if it was explicitly set
	} else if app.Config.TargetArg != "" {
		// If target was specified but we are running 'main', maybe pass it as 'target'?
		procArgsMap["target"] = app.Config.TargetArg
	}

	app.Log.Debug("Executing procedure.", "name", procedureToRun, "args_count", len(procArgsMap)) // Changed from Info
	execStartTime := time.Now()

	// --- RunProcedure Call ---
	// Sticking with the temporary fix from provided app.go: Run without args from map for now.
	// The provided app.go shows:
	// results, runErr = interpreter.RunProcedure(procedureToRun) // No args passed
	// The index for Interpreter.RunProcedure shows: (ctx context.Context, procName string, args ...interface{}) (interface{}, error)
	// For strict verbatim move of the *logic block* from the user's app.go, I will replicate its way of calling RunProcedure,
	// even if it means not passing ctx or the map args.
	var runErr error
	var results interface{}
	if len(procArgsMap) > 0 {
		app.Log.Warn("Procedure arguments provided via flags/map, but current RunProcedure call doesn't support passing them easily. Executing procedure without these arguments.", "procedure", procedureToRun)
		results, runErr = interpreter.RunProcedure(procedureToRun) // No args passed, as in provided app.go
	} else {
		results, runErr = interpreter.RunProcedure(procedureToRun) // No args passed, as in provided app.go
	}
	// --- End RunProcedure Call ---

	execEndTime := time.Now()
	duration := execEndTime.Sub(execStartTime)
	app.Log.Debug("Procedure execution finished.", "name", procedureToRun, "duration", duration) // Changed from Info

	if runErr != nil {
		app.Log.Error("Script execution failed.", "procedure", procedureToRun, "error", runErr)
		return fmt.Errorf("error executing procedure '%s': %w", procedureToRun, runErr)
	}

	app.Log.Debug("Script executed successfully.", "procedure", procedureToRun) // Changed from Info
	if results != nil {
		app.Log.Debug("Script Result Value", "result", fmt.Sprintf("%+v", results))
	} else {
		app.Log.Debug("Script Result Value: nil")
	}

	totalDuration := time.Since(startTime)
	app.Log.Debug("--- Script File Execution Finished ---", "path", scriptPath, "total_duration", totalDuration) // Changed from Info
	return nil
}
