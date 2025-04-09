// filename: pkg/neurogo/app_script.go
package neurogo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Assuming core imports remain the same relative path for now
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/neurodata/blocks"
	// checklist "github.com/aprice2704/neuroscript/pkg/neurodata/checklist"
)

// runScriptMode handles the execution of NeuroScript procedures.
func (a *App) runScriptMode(ctx context.Context) error {
	a.InfoLog.Printf("--- Starting NeuroGo in Script Execution Mode ---")
	a.InfoLog.Printf("Optional library paths (-lib): %v", a.Config.LibPaths)
	a.InfoLog.Printf("Target to run: %s", a.Config.TargetArg)
	a.InfoLog.Printf("Procedure args: %v", a.Config.ProcArgs)

	interpreter := core.NewInterpreter(a.DebugLog)

	// Register tools
	coreRegistry := interpreter.ToolRegistry()
	if coreRegistry == nil {
		return fmt.Errorf("internal error: Interpreter's ToolRegistry is nil after creation")
	}
	core.RegisterCoreTools(coreRegistry)
	blocks.RegisterBlockTools(coreRegistry)
	// checklist.RegisterChecklistTools(coreRegistry) // Keep commented

	// Load Libraries
	if err := a.loadLibraries(interpreter); err != nil {
		a.ErrorLog.Printf("Warning: Error during library loading: %v (continuing execution)", err)
		// Decide if loading errors should be fatal or just warnings
	}

	// Determine Procedure to Run
	procToRun, err := a.determineProcedureToRun(interpreter)
	if err != nil {
		return err // Fatal if target procedure cannot be determined/loaded
	}

	// Execute Procedure
	if _, exists := interpreter.KnownProcedures()[procToRun]; !exists {
		// This check might be redundant if determineProcedureToRun errors, but belt-and-suspenders
		return fmt.Errorf("procedure '%s' not found after loading libraries", procToRun)
	}

	a.InfoLog.Printf("Attempting to execute procedure: '%s' with args: %v", procToRun, a.Config.ProcArgs)
	fmt.Println("--- Procedure Output Start ---") // Keep direct output markers
	result, runErr := interpreter.RunProcedure(procToRun, a.Config.ProcArgs...)
	fmt.Println("--- Procedure Output End ---")
	a.InfoLog.Println("Execution finished.")

	if runErr != nil {
		a.ErrorLog.Printf("Execution Error: %v", runErr)
		if strings.Contains(runErr.Error(), "not defined or not loaded") {
			a.ErrorLog.Printf("Hint: Check if procedure '%s' exists in target file or specified libraries (-lib %v) and that there were no loading errors.", procToRun, a.Config.LibPaths)
		}
		return runErr // Propagate the original error
	}

	a.InfoLog.Printf("Final Result: %v (%T)", result, result)
	return nil // Success
}

// loadLibraries loads procedures from paths specified in Config.LibPaths.
func (a *App) loadLibraries(interpreter *core.Interpreter) error {
	if len(a.Config.LibPaths) == 0 {
		a.InfoLog.Printf("No library paths specified via -lib.")
		return nil
	}

	a.InfoLog.Printf("Loading procedures from specified library paths...")
	var firstError error

	for _, pathArg := range a.Config.LibPaths {
		pathInfo, statErr := os.Stat(pathArg)
		if statErr != nil {
			a.ErrorLog.Printf("Error accessing library path '%s': %v. Skipping.", pathArg, statErr)
			if firstError == nil {
				firstError = statErr
			}
			continue
		}

		if pathInfo.IsDir() {
			a.InfoLog.Printf("Scanning library directory: %s", pathArg)
			walkErr := filepath.WalkDir(pathArg, func(path string, d os.DirEntry, errIn error) error {
				if errIn != nil {
					a.ErrorLog.Printf("Error accessing path '%s' during library walk: %v", path, errIn)
					if firstError == nil {
						firstError = errIn
					}
					return filepath.SkipDir
				}
				if !d.IsDir() && (strings.HasSuffix(d.Name(), ".ns.txt") || strings.HasSuffix(d.Name(), ".ns") || strings.HasSuffix(d.Name(), ".neuro")) {
					_, _, fileErr := a.processNeuroScriptFile(path, interpreter)
					if fileErr != nil {
						// Log error but continue processing other files
						a.ErrorLog.Printf("Failed to process library file %s: %v", path, fileErr)
						if firstError == nil {
							firstError = fileErr
						}
					}
				}
				return nil // Continue walking
			})
			if walkErr != nil {
				a.ErrorLog.Printf("Error walking library directory '%s': %v.", pathArg, walkErr)
				if firstError == nil {
					firstError = walkErr
				}
			}
		} else { // It's a file
			if strings.HasSuffix(pathInfo.Name(), ".ns.txt") || strings.HasSuffix(pathInfo.Name(), ".ns") || strings.HasSuffix(pathInfo.Name(), ".neuro") {
				a.InfoLog.Printf("Loading library file: %s", pathArg)
				_, _, fileErr := a.processNeuroScriptFile(pathArg, interpreter)
				if fileErr != nil {
					a.ErrorLog.Printf("Failed to process library file %s: %v", pathArg, fileErr)
					if firstError == nil {
						firstError = fileErr
					}
				}
			} else {
				a.InfoLog.Printf("Skipping non-NeuroScript file specified via -lib: %s", pathArg)
			}
		}
	}
	a.InfoLog.Printf("Finished processing library paths.")
	// Return only the first error encountered during loading, allowing partial loads
	return firstError
}

// determineProcedureToRun figures out which procedure to run based on TargetArg.
func (a *App) determineProcedureToRun(interpreter *core.Interpreter) (string, error) {
	targetArg := a.Config.TargetArg
	isFilePathTarget := strings.HasSuffix(targetArg, ".ns.txt") || strings.HasSuffix(targetArg, ".ns") || strings.HasSuffix(targetArg, ".neuro")

	if isFilePathTarget {
		a.InfoLog.Printf("Target '%s' looks like a file. Loading it directly.", targetArg)
		cwd, errWd := os.Getwd()
		if errWd != nil {
			return "", fmt.Errorf("failed to get working directory: %w", errWd)
		}
		// Validate target path safety
		_, secErr := core.SecureFilePath(targetArg, cwd)
		if secErr != nil {
			return "", fmt.Errorf("target file path error: %w", secErr)
		}

		loadedProcs, _, loadErr := a.processNeuroScriptFile(targetArg, interpreter)
		if loadErr != nil {
			return "", fmt.Errorf("failed to load target file '%s': %w", targetArg, loadErr)
		}
		if len(loadedProcs) == 0 {
			return "", fmt.Errorf("target file '%s' loaded successfully but contains no valid procedures", targetArg)
		}
		procToRun := loadedProcs[0].Name
		a.InfoLog.Printf("Running first procedure '%s' found in file '%s'.", procToRun, targetArg)
		return procToRun, nil
	}

	// Target is treated as procedure name
	procToRun := targetArg
	a.InfoLog.Printf("Target '%s' treated as procedure name.", procToRun)
	if _, exists := interpreter.KnownProcedures()[procToRun]; !exists {
		// Error if the explicitly named procedure isn't found after loading libs
		return "", fmt.Errorf("procedure '%s' not found in loaded libraries (-lib %v)", procToRun, a.Config.LibPaths)
	}
	return procToRun, nil
}

// processNeuroScriptFile loads procedures from a single file.
func (a *App) processNeuroScriptFile(path string, interp *core.Interpreter) ([]core.Procedure, string, error) {
	fileName := filepath.Base(path)
	a.DebugLog.Printf("--- Processing File: %s ---", path)

	contentBytes, readErr := os.ReadFile(path)
	if readErr != nil {
		return nil, "", fmt.Errorf("could not read file %s: %w", path, readErr)
	}

	parseOptions := core.ParseOptions{DebugAST: a.Config.DebugAST, Logger: a.DebugLog}
	stringReader := strings.NewReader(string(contentBytes))
	procedures, fileVersion, parseErr := core.ParseNeuroScript(stringReader, fileName, parseOptions)

	if parseErr != nil {
		errorMsg := fmt.Sprintf("Parse error processing %s: %s", path, parseErr.Error())
		a.ErrorLog.Print(errorMsg) // Log parse error
		return nil, fileVersion, fmt.Errorf(errorMsg)
	}

	if fileVersion != "" {
		a.DebugLog.Printf("Found FILE_VERSION %q in %s", fileVersion, path)
	}

	addedProcedures := make([]core.Procedure, 0, len(procedures))
	var firstLoadError error
	for _, proc := range procedures {
		loadErr := interp.AddProcedure(proc)
		if loadErr != nil {
			a.ErrorLog.Printf("Load error adding procedure '%s' from %s: %v", proc.Name, path, loadErr)
			if firstLoadError == nil {
				firstLoadError = loadErr
			}
		} else {
			a.DebugLog.Printf("  Successfully added procedure '%s' from %s.", proc.Name, path)
			addedProcedures = append(addedProcedures, proc)
		}
	}
	if len(addedProcedures) > 0 {
		a.DebugLog.Printf("  Added %d procedures in total from %s.", len(addedProcedures), path)
	}
	// Return procedures found and the first loading error encountered
	return addedProcedures, fileVersion, firstLoadError
}
