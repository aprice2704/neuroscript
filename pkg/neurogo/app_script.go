// filename: pkg/neurogo/app_script.go
package neurogo

import (
	"context"
	"fmt"
	"strings"
	"time" // Added for logging timestamp

	// Ensure the core package is imported correctly
	"github.com/aprice2704/neuroscript/pkg/core"
	// Import other necessary packages
	"github.com/aprice2704/neuroscript/pkg/neurodata/blocks"
	checklist "github.com/aprice2704/neuroscript/pkg/neurodata/checklist"
)

// runScriptMode handles the execution of NeuroScript procedures.
func (a *App) runScriptMode(ctx context.Context) error {
	startTime := time.Now()
	a.Logger.Info("--- Starting NeuroGo in Script Execution Mode [%s] ---", startTime.Format(time.RFC3339))
	a.Logger.Info("Optional library paths (-lib): %v", a.Config.LibPaths)
	a.Logger.Info("Script file (-script): %s", a.Config.ScriptFile)
	a.Logger.Info("Target procedure (-target): %s", a.Config.TargetArg)
	a.Logger.Info("Procedure args: %v", a.Config.ProcArgs)

	// Ensure interpreter is available on the App struct
	if a.interpreter == nil {
		a.Logger.Error("Interpreter not initialized before runScriptMode")
		return fmt.Errorf("interpreter not initialized")
		// Or initialize it here if appropriate:
		// a.interpreter = core.NewInterpreter(a.Logger, a.llmClient)
	}
	interpreter := a.interpreter // Use the App's interpreter

	// --- Tool Registration ---
	// Tool registration might happen earlier in App initialization, verify this flow.
	// If not, register tools here.
	a.Logger.Debug("Verifying tool registration...")
	coreRegistry := interpreter.ToolRegistry()
	if coreRegistry == nil {
		return fmt.Errorf("internal error: Interpreter's ToolRegistry is nil after creation")
	}
	// Assuming core tools are registered in NewInterpreter
	// Register domain-specific tools (idempotency might be handled by registry)
	if err := blocks.RegisterBlockTools(coreRegistry); err != nil {
		a.Logger.Error("Failed to register blocks tools: %v", err) // Consider if fatal
	} else {
		a.Logger.Debug("Blocks tools registered.")
	}
	if err := checklist.RegisterChecklistTools(coreRegistry); err != nil {
		a.Logger.Error("Failed to register checklist tools: %v", err) // Consider if fatal
	} else {
		a.Logger.Debug("Checklist tools registered.")
	}
	a.Logger.Debug("Tool registration check complete.")
	// --- End Tool Registration ---

	// --- Load Libraries ---
	// Libraries are loaded first so they are available to the main script/target procedure.
	if err := a.loadLibraries(interpreter); err != nil {
		// Treat library loading errors as non-fatal for now, just log them.
		a.Logger.Error("Warning: Encountered errors during library loading: %v (execution will continue)", err)
	}
	// --- End Load Libraries ---

	// --- Determine Procedure ---
	// This might load the main script file if -target is not specified.
	procToRun, err := a.determineProcedureToRun(interpreter)
	if err != nil {
		// If we can't even determine which procedure to run, it's fatal.
		return fmt.Errorf("failed to determine procedure to run: %w", err)
	}
	a.Logger.Info("Determined procedure to run: '%s'", procToRun)
	// --- End Determine Procedure ---

	// --- Execute Procedure ---
	// Final check to ensure the determined procedure is actually known to the interpreter.
	if _, exists := interpreter.KnownProcedures()[procToRun]; !exists {
		// If the procedure determined (e.g., "main" or from -target) isn't found after loading libs
		// and potentially the main script (in determineProcedureToRun), then it's missing.
		errMsg := fmt.Sprintf("procedure '%s' not found after loading all specified files", procToRun)
		a.Logger.Error(errMsg)
		a.Logger.Error("Hint: Check procedure name correctness, ensure it's defined in '%s' or libraries %v, and check for loading errors above.",
			a.Config.ScriptFile, a.Config.LibPaths)
		// Provide list of known procedures for debugging aid
		knownProcs := []string{}
		procMap := interpreter.KnownProcedures() // Get map once
		if procMap != nil {
			for name := range procMap {
				knownProcs = append(knownProcs, name)
			}
		}
		a.Logger.Info("Known procedures: %v", knownProcs)
		return fmt.Errorf("%s", errMsg) // Use %s for simple string error

	}

	// Convert []string to []interface{} for RunProcedure arguments.
	procArgsInterface := make([]interface{}, len(a.Config.ProcArgs))
	for i, v := range a.Config.ProcArgs {
		procArgsInterface[i] = v
	}

	a.Logger.Info("Executing procedure: '%s' with args: %v", procToRun, a.Config.ProcArgs)
	fmt.Println("--- Procedure Output Start ---")
	// Pass the converted slice with the spread operator '...'
	result, runErr := interpreter.RunProcedure(procToRun, procArgsInterface...)
	fmt.Println("--- Procedure Output End ---")
	execEndTime := time.Now()
	duration := execEndTime.Sub(startTime)
	a.Logger.Info("Execution finished at %s (Duration: %s).", execEndTime.Format(time.RFC3339), duration)

	if runErr != nil {
		a.Logger.Error("Execution Error", "procedure", procToRun, "error", runErr)
		// Provide specific hints based on error type/content if possible
		if strings.Contains(runErr.Error(), "undefined tool") {
			a.Logger.Error("Hint: Check if the tool mentioned is registered correctly or if there's a typo.")
		} else if strings.Contains(runErr.Error(), "arity mismatch") {
			a.Logger.Error("Hint: Check the number of arguments passed to the procedure or tool call.")
		}
		return fmt.Errorf("procedure '%s' execution failed: %w", procToRun, runErr)
	}

	a.Logger.Info("Final Result", "value", result, "type", fmt.Sprintf("%T", result))
	return nil
	// --- End Execute Procedure ---
}

// loadLibraries loads procedures from specified library paths into the interpreter.
func (a *App) loadLibraries(interpreter *core.Interpreter) error {
	if interpreter == nil {
		return fmt.Errorf("cannot load libraries with a nil interpreter")
	}
	if len(a.Config.LibPaths) == 0 {
		a.Logger.Debug("No library paths specified (-lib). Skipping library loading.")
		return nil
	}
	a.Logger.Info("Loading libraries from paths: %v", a.Config.LibPaths)
	var loadErrors []string
	for _, libPath := range a.Config.LibPaths {
		a.Logger.Debug("Processing library file: %s", libPath)
		// processNeuroScriptFile parses and adds procedures to the interpreter.
		// We only care about the error here, not the returned procedures or default name.
		_, _, err := a.processNeuroScriptFile(libPath, interpreter) // Pass interpreter instance
		if err != nil {
			errMsg := fmt.Sprintf("error processing library file '%s': %v", libPath, err)
			a.Logger.Error(errMsg) // Log individual error immediately
			loadErrors = append(loadErrors, errMsg)
		} else {
			a.Logger.Info("Successfully processed library file: %s", libPath)
		}
	}
	if len(loadErrors) > 0 {
		// Return a single error summarizing the failures
		return fmt.Errorf("encountered %d error(s) loading libraries: %s", len(loadErrors), strings.Join(loadErrors, "; "))
	}
	a.Logger.Debug("Library loading complete.")
	return nil
}

// determineProcedureToRun decides which procedure to execute based on configuration.
// It will load the main script file if -target is not specified and -script is.
func (a *App) determineProcedureToRun(interpreter *core.Interpreter) (string, error) {
	if interpreter == nil {
		return "", fmt.Errorf("cannot determine procedure with a nil interpreter")
	}
	// 1. If a target procedure is explicitly given (-target), use that.
	if a.Config.TargetArg != "" {
		a.Logger.Info("Using specified target procedure from -target: '%s'", a.Config.TargetArg)
		// We don't need to load the main script file here, assuming the target
		// is either defined in already-loaded libraries or will be checked later.
		return a.Config.TargetArg, nil
	}

	// 2. If no target is given, but a script file is (-script), process it.
	if a.Config.ScriptFile != "" {
		a.Logger.Info("No -target specified. Processing script file '%s' to find default procedure.", a.Config.ScriptFile)
		// processNeuroScriptFile loads procedures from this file into the interpreter.
		// It returns the procedures, a default name (always "" now), and error.
		_, defaultProcName, err := a.processNeuroScriptFile(a.Config.ScriptFile, interpreter) // Pass interpreter
		if err != nil {
			// If the main script file cannot be processed, we can't proceed.
			return "", fmt.Errorf("error processing main script file '%s': %w", a.Config.ScriptFile, err)
		}

		// Since processNeuroScriptFile now always returns "" for defaultProcName based on
		// the actual Program struct, this check `if defaultProcName != ""` will always be false.
		// We directly proceed to the fallback.
		if defaultProcName != "" {
			// This block is technically unreachable now but kept for conceptual clarity.
			a.Logger.Info("Using default procedure '%s' determined from script file metadata (unexpected).", defaultProcName)
			return defaultProcName, nil
		} else {
			// Fallback: If no default procedure name is derived (which is always the case now), use "main".
			a.Logger.Info("No default procedure derived from '%s', defaulting to procedure name 'main'.", a.Config.ScriptFile)
			return "main", nil
		}
	}

	// 3. If neither -target nor -script is specified, we don't know what to run.
	a.Logger.Error("Cannot determine procedure to run: Neither -target nor -script was specified.")
	return "", fmt.Errorf("no procedure specified: use -target <proc_name> or -script <file_path>")
}

// processNeuroScriptFile parses a NeuroScript file, adds its procedures to the interpreter,
// and returns the list of procedures found, an empty string (as default name cannot be determined
// from the core.Program struct), and any processing error.
func (a *App) processNeuroScriptFile(path string, interp *core.Interpreter) ([]core.Procedure, string, error) {
	if interp == nil {
		return nil, "", fmt.Errorf("cannot process script '%s' with a nil interpreter", path)
	}
	a.Logger.Debug("Starting processing of NeuroScript file: %s", path)
	defaultProcName := "" // Cannot be determined from core.Program Metadata, always return empty.

	// Read file content
	contentBytes, err := core.ReadFileContent(path) // Use exported function
	if err != nil {
		return nil, defaultProcName, fmt.Errorf("reading file %s: %w", path, err)
	}
	content := string(contentBytes)
	a.Logger.Debug("Read %d bytes from %s", len(content), path)

	// Parse content into a syntax tree
	parser := core.NewParserAPI(a.Logger)
	// --- MODIFIED ERROR HANDLING ---
	tree, parseErr := parser.Parse(content) // Get single error
	if parseErr != nil {                    // Check if error is non-nil
		errMsg := fmt.Sprintf("parsing file %s failed: %v", path, parseErr)
		// Log the detailed error here
		a.Logger.Error("Parser error encountered", "file", path, "error", parseErr)
		return nil, defaultProcName, fmt.Errorf("%s", errMsg) // Return the error
	}
	// --- END MODIFIED ERROR HANDLING ---
	if tree == nil {
		// Handle case where parsing might succeed technically but yield no tree
		return nil, defaultProcName, fmt.Errorf("parsing file %s produced nil tree", path)
	}
	a.Logger.Debug("Parsed %s successfully.", path)

	// Build AST (core.Program) from the syntax tree
	astBuilder := core.NewASTBuilder(a.Logger)
	program, err := astBuilder.Build(tree)
	if err != nil {
		return nil, defaultProcName, fmt.Errorf("building AST for %s: %w", path, err)
	}
	if program == nil {
		// Should not happen if err is nil, but check defensively.
		return nil, defaultProcName, fmt.Errorf("internal error: AST builder returned nil program without error for %s", path)
	}

	// --- MODIFIED: Access Metadata for file version ---
	fileVersion := "(not specified)" // Default value
	if version, ok := program.Metadata["file_version"]; ok {
		fileVersion = version
	}
	// Use structured logging for clarity
	a.Logger.Debug("Built AST for file.", "path", path, "file_version", fileVersion)
	// --- END MODIFICATION ---

	// Process procedures found in the AST
	if program.Procedures == nil {
		// Handle case where file is valid but contains no procedures.
		a.Logger.Debug("No procedures found in '%s'.", path)
		program.Procedures = []core.Procedure{} // Ensure it's an empty slice, not nil
	}

	a.Logger.Debug("Found procedures, adding to interpreter.", "count", len(program.Procedures), "path", path)
	for _, proc := range program.Procedures {
		procCopy := proc // Create copy to pass by value
		if addErr := interp.AddProcedure(procCopy); addErr != nil {
			// Log warnings for issues like redefinition, but don't stop processing the file.
			a.Logger.Warn("Error adding procedure to interpreter", "procedure", proc.Name, "path", path, "error", addErr)
			// If AddProcedure could return a truly fatal error, might need to return here.
		} else {
			a.Logger.Debug("Added procedure.", "procedure", proc.Name, "path", path)
		}
	}

	// Log completion and return results
	// Note: defaultProcName is always "", as it cannot be derived from core.Program Metadata.
	a.Logger.Debug("Finished processing file.", "path", path, "procedures_found", len(program.Procedures))
	return program.Procedures, defaultProcName, nil
}
