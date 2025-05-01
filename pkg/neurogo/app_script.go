// filename: pkg/neurogo/app_script.go
package neurogo

import (
	"context" // Import errors
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/neurodata/blocks"
	"github.com/aprice2704/neuroscript/pkg/neurodata/checklist"
)

// runScriptMode executes a NeuroScript (.ns) file.
func (a *App) runScriptMode(ctx context.Context) error {
	startTime := time.Now()
	a.Log.Info("--- Running in Script Mode ---")
	a.Log.Info("Script File:", "path", a.Config.ScriptFile)
	// ... (logging args unchanged) ...

	// Direct nil check, no type assertion needed as a.interpreter is *core.Interpreter
	if a.interpreter == nil {
		a.Log.Error("Interpreter is not correctly initialized.", "type", fmt.Sprintf("%T", a.interpreter))
		return fmt.Errorf("cannot run script mode: interpreter is nil")
	}
	interpreter := a.interpreter // Use the field directly

	registry := interpreter.ToolRegistry()
	if registry == nil {
		return fmt.Errorf("tool registry is nil")
	}
	// ... (tool registration unchanged) ...
	a.Log.Debug("Registering core tools for script mode.")
	if err := core.RegisterCoreTools(registry); err != nil {
		a.Log.Error("Failed to register core tools", "error", err)
		// Decide if this should be fatal for script mode
	}
	a.Log.Debug("Registering data tools (Blocks, Checklist).")
	if err := blocks.RegisterBlockTools(registry); err != nil {
		a.Log.Error("Failed to register block tools", "error", err)
	}
	if err := checklist.RegisterChecklistTools(registry); err != nil {
		a.Log.Error("Failed to register checklist tools", "error", err)
	}

	a.Log.Debug("Loading library files.")
	if err := a.loadLibraries(interpreter); err != nil { // Pass the interpreter
		a.Log.Error("Failed to load libraries", "error", err)
		return fmt.Errorf("error loading libraries: %w", err)
	}
	a.Log.Info("Libraries loaded.")

	procedureToRun, err := a.determineProcedureToRun(interpreter) // Pass the interpreter
	if err != nil {
		return err
	}
	if procedureToRun == "" {
		a.Log.Error("No procedure specified to run (use -target or ensure a 'main' procedure exists).")
		// Assuming KnownProcedures exists and returns map[string]*core.Procedure
		// knownProcs := interpreter.KnownProcedures() // Method expects Procedure value, not pointer
		// procNames := make([]string, 0, len(knownProcs))
		// for name := range knownProcs {
		// 	procNames = append(procNames, name)
		// }
		// a.Log.Info("Available procedures:", "procs", procNames) // Log names only
		return fmt.Errorf("no procedure target specified or found")
	}

	a.Log.Info("Target procedure determined.", "name", procedureToRun)

	procArgsMap := make(map[string]interface{})
	// ... (arg map population unchanged) ...
	for i, arg := range a.Config.ProcArgs {
		procArgsMap[fmt.Sprintf("arg%d", i+1)] = arg
	}
	if a.Config.TargetArg != "" {
		procArgsMap["target"] = a.Config.TargetArg
	}

	a.Log.Info("Executing procedure.", "name", procedureToRun, "args_count", len(procArgsMap))
	execStartTime := time.Now()

	// Use the interpreter directly
	procName := procedureToRun
	// arguments := procArgsMap // Map not directly usable with current RunProcedure signature

	// --- FIX for RunProcedure call ---
	// Current RunProcedure expects (string, ...interface{})
	// We don't have the ordered arguments easily here.
	// IDEAL FIX: Change RunProcedure signature in interpreter.go to accept context and map.
	// TEMPORARY FIX: Call assuming no arguments are passed from this map for now.
	// This WILL break if the target procedure requires arguments passed via flags/map.
	var runErr error
	var results interface{}
	if len(procArgsMap) > 0 {
		a.Log.Warn("Procedure arguments provided via flags/map, but current RunProcedure call doesn't support passing them easily. Executing procedure without these arguments.", "procedure", procName)
		// Attempt call without variadic args
		results, runErr = interpreter.RunProcedure(procName)
	} else {
		// Call normally if no map args
		results, runErr = interpreter.RunProcedure(procName)
	}
	// --- End FIX ---

	execEndTime := time.Now()
	duration := execEndTime.Sub(execStartTime)
	a.Log.Info("Procedure execution finished.", "name", procedureToRun, "duration", duration)

	if runErr != nil {
		// ... (error logging unchanged) ...
		a.Log.Error("Script execution failed.", "procedure", procedureToRun, "error", runErr)
		fmt.Fprintf(os.Stderr, "Error executing procedure '%s':\n%v\n", procedureToRun, runErr)
		return fmt.Errorf("script execution failed: %w", runErr)
	}

	// ... (result printing unchanged) ...
	a.Log.Info("Script executed successfully.", "procedure", procedureToRun)
	if results != nil {
		fmt.Println("--- Script Result ---")
		fmt.Printf("%+v\n", results)
		a.Log.Debug("Script Result Value", "result", results)
	} else {
		fmt.Println("--- Script Finished (No explicit return value) ---")
		a.Log.Debug("Script Result Value: nil")
	}

	totalDuration := time.Since(startTime)
	a.Log.Info("--- Script Mode Finished ---", "total_duration", totalDuration)
	return nil
}

// loadLibraries processes all files specified in Config.LibPaths.
func (a *App) loadLibraries(interpreter *core.Interpreter) error { // Accepts correct type
	a.Log.Debug("Loading libraries from paths.", "paths", a.Config.LibPaths)
	for _, libPath := range a.Config.LibPaths {
		absPath, err := filepath.Abs(libPath)
		if err != nil {
			a.Log.Warn("Could not get absolute path for library path, skipping.", "path", libPath, "error", err)
			continue
		}
		a.Log.Info("Processing library path.", "path", absPath)

		info, err := os.Stat(absPath)
		if err != nil {
			a.Log.Warn("Could not stat library path, skipping.", "path", absPath, "error", err)
			continue
		}

		if info.IsDir() {
			err := filepath.WalkDir(absPath, func(path string, d os.DirEntry, walkErr error) error {
				if walkErr != nil {
					a.Log.Warn("Error accessing path during library walk, skipping.", "path", path, "error", walkErr)
					return nil // Skip this item, continue walk
				}
				if !d.IsDir() && strings.HasSuffix(d.Name(), ".ns") {
					a.Log.Debug("Processing library file.", "file", path)
					_, _, procErr := a.processNeuroScriptFile(path, interpreter) // Pass interpreter
					if procErr != nil {
						// Log but continue processing other library files
						a.Log.Error("Failed to process library file, continuing...", "file", path, "error", procErr)
					}
				}
				return nil
			})
			if err != nil {
				a.Log.Error("Error walking library directory.", "path", absPath, "error", err)
				// Consider returning the error if loading libraries should be atomic.
			}
		} else if strings.HasSuffix(info.Name(), ".ns") {
			a.Log.Debug("Processing library file.", "file", absPath)
			_, _, procErr := a.processNeuroScriptFile(absPath, interpreter) // Pass interpreter
			if procErr != nil {
				a.Log.Error("Failed to process library file.", "file", absPath, "error", procErr)
				// Consider returning the error
			}
		} else {
			a.Log.Warn("Library path is not a directory or .ns file, skipping.", "path", absPath)
		}
	}
	a.Log.Debug("Finished loading libraries.")
	return nil
}

// determineProcedureToRun finds the main script file and identifies the procedure to execute.
func (a *App) determineProcedureToRun(interpreter *core.Interpreter) (string, error) { // Accepts correct type
	if a.Config.ScriptFile == "" {
		return "", fmt.Errorf("no script file specified (-script)")
	}

	// Process the main script file after libraries
	a.Log.Info("Processing main script file.", "path", a.Config.ScriptFile)
	_, fileMeta, err := a.processNeuroScriptFile(a.Config.ScriptFile, interpreter) // Pass interpreter
	if err != nil {
		return "", fmt.Errorf("failed to process script %s: %w", a.Config.ScriptFile, err)
	}
	if fileMeta == nil {
		// Initialize if nil to avoid panics later
		fileMeta = make(map[string]string)
		a.Log.Warn("No file metadata available from main script file processing.", "path", a.Config.ScriptFile)
	}

	// Determine target procedure: flag -> metadata -> default 'main'
	targetProc := a.Config.TargetArg
	if targetProc != "" {
		a.Log.Info("Using target procedure from -target flag.", "procedure", targetProc)
		return targetProc, nil
	}

	if defaultTarget, ok := fileMeta["target"]; ok && defaultTarget != "" {
		a.Log.Info("Using target procedure from script metadata.", "procedure", defaultTarget)
		return defaultTarget, nil
	}

	a.Log.Info("No target specified via flag or metadata, defaulting to 'main'.")
	return "main", nil
}

// processNeuroScriptFile parses a .ns file and adds its procedures to the interpreter.
// Returns the list of procedures defined in THIS file and the file's metadata.
func (a *App) processNeuroScriptFile(filePath string, interp *core.Interpreter) ([]*core.Procedure, map[string]string, error) { // Accepts correct type
	a.Log.Debug("Processing NeuroScript file.", "path", filePath)
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		a.Log.Error("Failed to read script file.", "path", filePath, "error", err)
		return nil, nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	content := string(contentBytes)

	parser := core.NewParserAPI(a.Log)
	// Corrected call: Parse only takes source content
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
	// Corrected call: Build returns program, metadata, error
	programAST, fileMetadata, buildErr := astBuilder.Build(parseResultTree)
	if buildErr != nil {
		a.Log.Error("AST building failed.", "path", filePath, "error", buildErr)
		// Return potentially non-nil metadata even on build error
		return nil, fileMetadata, fmt.Errorf("AST building for %s failed: %w", filePath, buildErr)
	}
	if programAST == nil {
		a.Log.Error("AST building returned nil program without errors.", "path", filePath)
		return nil, fileMetadata, fmt.Errorf("internal AST building error: nil program for %s", filePath)
	}
	// Ensure fileMetadata is not nil if programAST is not nil (Build should guarantee this)
	if fileMetadata == nil {
		fileMetadata = make(map[string]string) // Defensive init
	}
	a.Log.Debug("AST building successful.", "path", filePath, "procedures", len(programAST.Procedures), "metadata_keys", len(fileMetadata))

	// Add procedures to interpreter
	definedProcs := []*core.Procedure{} // Slice of pointers
	if programAST.Procedures != nil {   // Check map is not nil
		for name, proc := range programAST.Procedures { // proc is *core.Procedure
			if proc == nil {
				a.Log.Warn("Skipping nil procedure found in AST map.", "name", name, "path", filePath)
				continue
			}
			// Corrected call: Dereference proc because AddProcedure expects Procedure value (based on fetched interpreter.go)
			// IDEAL FIX: Change AddProcedure signature in interpreter.go
			if err := interp.AddProcedure(*proc); err != nil { // DEREFERENCE HERE
				a.Log.Error("Failed to add procedure to interpreter.", "procedure", name, "path", filePath, "error", err)
				// Return error immediately to avoid inconsistent state
				return definedProcs, fileMetadata, fmt.Errorf("failed to add procedure '%s' from '%s': %w", name, filePath, err)
			} else {
				a.Log.Debug("Added procedure to interpreter.", "procedure", name, "path", filePath)
				definedProcs = append(definedProcs, proc) // Append the pointer to the list for this file
			}
		}
	}

	// Metadata now comes from Build return value

	a.Log.Debug("Finished processing file.", "path", filePath, "procedures_added", len(definedProcs), "metadata_keys", len(fileMetadata))
	return definedProcs, fileMetadata, nil
}
