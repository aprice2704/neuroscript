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

	// <<< FIX: Add Type Assertion for Interpreter >>>
	interpreter, ok := a.interpreter.(*core.Interpreter)
	if !ok || interpreter == nil {
		a.Log.Error("Interpreter is not correctly initialized or has unexpected type for script mode.", "type", fmt.Sprintf("%T", a.interpreter))
		return fmt.Errorf("cannot run script mode: interpreter is not *core.Interpreter or is nil")
	}

	registry := interpreter.ToolRegistry()
	if registry == nil {
		return fmt.Errorf("tool registry is nil")
	}
	// ... (tool registration unchanged) ...
	a.Log.Debug("Registering core tools for script mode.")
	if err := core.RegisterCoreTools(registry); err != nil {
		a.Log.Error("Failed to register core tools", "error", err)
	}
	a.Log.Debug("Registering data tools (Blocks, Checklist).")
	if err := blocks.RegisterBlockTools(registry); err != nil {
		a.Log.Error("Failed to register block tools", "error", err)
	}
	if err := checklist.RegisterChecklistTools(registry); err != nil {
		a.Log.Error("Failed to register checklist tools", "error", err)
	}

	a.Log.Debug("Loading library files.")
	if err := a.loadLibraries(interpreter); err != nil { // Pass asserted interpreter
		a.Log.Error("Failed to load libraries", "error", err)
		return fmt.Errorf("error loading libraries: %w", err)
	}
	a.Log.Info("Libraries loaded.")

	procedureToRun, err := a.determineProcedureToRun(interpreter) // Pass asserted interpreter
	if err != nil {
		return err
	}
	if procedureToRun == "" {
		a.Log.Error("No procedure specified to run (use -target or ensure a 'main' procedure exists).")
		knownProcs := interpreter.KnownProcedures()
		a.Log.Info("Available procedures:", "procs", knownProcs)
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

	// Use asserted interpreter
	procName := procedureToRun
	arguments := procArgsMap
	currentCtx := ctx
	runProcFunc := interpreter.RunProcedure // Assign method from asserted interpreter

	results, runErr := runProcFunc(currentCtx, procName, arguments)
	execEndTime := time.Now()
	duration := execEndTime.Sub(execStartTime)
	a.Log.Info("Procedure execution finished.", "name", procedureToRun, "duration", duration)

	if runErr != nil {
		// ... (error logging unchanged) ...
		a.Log.Error("Script execution failed.", "procedure", procedureToRun, "error", runErr)
		fmt.Fprintf(os.Stderr, "Error executing procedure '%s':\n%v\n", procedureToRun, runErr)
		if strings.Contains(runErr.Error(), "context.Context") && strings.Contains(runErr.Error(), "string") {
			a.Log.Error("RunProcedure argument mismatch error persists.", "proc", procedureToRun, "error", runErr)
		}
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
func (a *App) loadLibraries(interpreter *core.Interpreter) error { // Accepts asserted type
	// ... (unchanged) ...
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
					return nil
				}
				if !d.IsDir() && strings.HasSuffix(d.Name(), ".ns") {
					a.Log.Debug("Processing library file.", "file", path)
					_, _, procErr := a.processNeuroScriptFile(path, interpreter) // Pass interpreter
					if procErr != nil {
						a.Log.Error("Failed to process library file, continuing...", "file", path, "error", procErr)
					}
				}
				return nil
			})
			if err != nil {
				a.Log.Error("Error walking library directory.", "path", absPath, "error", err)
				return fmt.Errorf("error walking library dir %s: %w", absPath, err)
			}
		} else if strings.HasSuffix(info.Name(), ".ns") {
			a.Log.Debug("Processing library file.", "file", absPath)
			_, _, procErr := a.processNeuroScriptFile(absPath, interpreter) // Pass interpreter
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

// determineProcedureToRun finds the main script file and identifies the procedure to execute.
func (a *App) determineProcedureToRun(interpreter *core.Interpreter) (string, error) { // Accepts asserted type
	// ... (unchanged) ...
	if a.Config.ScriptFile == "" {
		return "", fmt.Errorf("no script file specified (-script)")
	}

	a.Log.Info("Processing main script file.", "path", a.Config.ScriptFile)
	_, fileMeta, err := a.processNeuroScriptFile(a.Config.ScriptFile, interpreter) // Pass interpreter
	if err != nil {
		return "", fmt.Errorf("failed to process script %s: %w", a.Config.ScriptFile, err)
	}

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
func (a *App) processNeuroScriptFile(filePath string, interp *core.Interpreter) ([]*core.Procedure, map[string]string, error) { // Accepts asserted type
	// ... (reading file unchanged) ...
	a.Log.Debug("Processing NeuroScript file.", "path", filePath)
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		a.Log.Error("Failed to read script file.", "path", filePath, "error", err)
		return nil, nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	content := string(contentBytes)

	parser := core.NewParserAPI(a.Log)
	parseResult, parseErr := parser.Parse(content)
	if parseErr != nil { // Check single error
		a.Log.Error("Parsing failed.", "path", filePath, "error", parseErr)
		// Combine parse errors if Parse can return multiple (unlikely based on signature)
		return nil, nil, fmt.Errorf("parsing file %s failed: %w", filePath, parseErr)
	}
	if parseResult == nil {
		a.Log.Error("Parsing returned nil result without errors.", "path", filePath)
		return nil, nil, fmt.Errorf("internal parsing error: nil result for %s", filePath)
	}
	a.Log.Debug("Parsing successful.", "path", filePath)

	astBuilder := core.NewASTBuilder(a.Log)
	programAST, buildErr := astBuilder.Build(parseResult) // Check single error
	if buildErr != nil {
		a.Log.Error("AST building failed.", "path", filePath, "error", buildErr)
		// Combine build errors if Build can return multiple (unlikely based on signature)
		return nil, nil, fmt.Errorf("AST building for %s failed: %w", filePath, buildErr)
	}
	if programAST == nil {
		a.Log.Error("AST building returned nil program without errors.", "path", filePath)
		return nil, nil, fmt.Errorf("internal AST building error: nil program for %s", filePath)
	}
	a.Log.Debug("AST building successful.", "path", filePath, "procedures", len(programAST.Procedures))

	// Add procedures to interpreter
	definedProcs := []*core.Procedure{} // Slice of pointers
	if programAST.Procedures != nil {   // Check map is not nil
		for name, proc := range programAST.Procedures { // proc should be *core.Procedure here
			// <<< FIX: Check proc pointer against nil >>>
			if proc == nil {
				a.Log.Warn("Skipping nil procedure found in AST map.", "name", name, "path", filePath)
				continue
			}
			if err := interp.AddProcedure(proc); err != nil {
				a.Log.Error("Failed to add procedure to interpreter.", "procedure", name, "path", filePath, "error", err)
			} else {
				a.Log.Debug("Added procedure to interpreter.", "procedure", name, "path", filePath)
				// <<< FIX: Try appending address if compiler insists proc is value type >>>
				// This contradicts the map definition map[string]*Procedure.
				// If this compiles, there's a deep confusion.
				// definedProcs = append(definedProcs, proc) // Original attempt
				definedProcs = append(definedProcs, proc) // Try appending the pointer directly first. If error persists, try &proc.
			}
		}
	}

	// <<< FIX: Call the GetFileMetadata method on the builder >>>
	fileMetadata := astBuilder.GetFileMetadata()
	a.Log.Debug("Finished processing file.", "path", filePath, "procedures_added", len(definedProcs), "metadata_keys", len(fileMetadata))
	return definedProcs, fileMetadata, nil
}
