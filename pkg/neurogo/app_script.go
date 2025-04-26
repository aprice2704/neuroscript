// filename: pkg/neurogo/app_script.go
package neurogo

import (
	"context"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/neurodata/blocks"
	checklist "github.com/aprice2704/neuroscript/pkg/neurodata/checklist"
)

// runScriptMode handles the execution of NeuroScript procedures.
func (a *App) runScriptMode(ctx context.Context) error {
	a.Logger.Info("--- Starting NeuroGo in Script Execution Mode ---")
	a.Logger.Info("Optional library paths (-lib): %v", a.Config.LibPaths)
	a.Logger.Info("Target to run: %s", a.Config.TargetArg)
	a.Logger.Info("Procedure args: %v", a.Config.ProcArgs)

	// +++ MODIFIED: Pass a.llmClient to NewInterpreter +++
	interpreter := core.NewInterpreter(a.DebugLog, a.llmClient)
	// --- END MODIFIED ---

	// --- Tool Registration (unchanged) ---
	coreRegistry := interpreter.ToolRegistry()
	if coreRegistry == nil {
		return fmt.Errorf("internal error: Interpreter's ToolRegistry is nil after creation")
	}
	core.RegisterCoreTools(coreRegistry)
	if err := blocks.RegisterBlockTools(coreRegistry); err != nil {
		a.Logger.Error("CRITICAL: Failed to register blocks tools: %v", err)
	} else {
		a.Logger.Debug("Registered blocks tools.")
	}
	if err := checklist.RegisterChecklistTools(coreRegistry); err != nil {
		a.Logger.Error("CRITICAL: Failed to register checklist tools: %v", err)
	} else {
		a.Logger.Debug("Registered checklist tools.")
	}
	// --- End Tool Registration ---

	// --- Load Libraries (unchanged) ---
	if err := a.loadLibraries(interpreter); err != nil {
		a.Logger.Error("Warning: Error during library loading: %v (continuing execution)", err)
	}
	// --- End Load Libraries ---

	// --- Determine Procedure (unchanged) ---
	procToRun, err := a.determineProcedureToRun(interpreter)
	if err != nil {
		return err
	}
	// --- End Determine Procedure ---

	// --- Execute Procedure (unchanged) ---
	if _, exists := interpreter.KnownProcedures()[procToRun]; !exists {
		return fmt.Errorf("procedure '%s' not found after loading libraries", procToRun)
	}

	a.Logger.Info("Attempting to execute procedure: '%s' with args: %v", procToRun, a.Config.ProcArgs)
	fmt.Println("--- Procedure Output Start ---")
	result, runErr := interpreter.RunProcedure(procToRun, a.Config.ProcArgs...)
	fmt.Println("--- Procedure Output End ---")
	a.Logger.Info("Execution finished.")

	if runErr != nil {
		a.Logger.Error("Execution Error: %v", runErr)
		if strings.Contains(runErr.Error(), "not defined or not loaded") {
			a.Logger.Error("Hint: Check if procedure '%s' exists in target file or specified libraries (-lib %v) and that there were no loading errors.", procToRun, a.Config.LibPaths)
		}
		return runErr
	}

	a.Logger.Info("Final Result: %v (%T)", result, result)
	return nil
	// --- End Execute Procedure ---
}

// loadLibraries (unchanged)
func (a *App) loadLibraries(interpreter *core.Interpreter) error {
	// ... (implementation unchanged) ...
	return nil
}

// determineProcedureToRun (unchanged)
func (a *App) determineProcedureToRun(interpreter *core.Interpreter) (string, error) {
	// ... (implementation unchanged) ...
	return "", nil
}

// processNeuroScriptFile (unchanged)
func (a *App) processNeuroScriptFile(path string, interp *core.Interpreter) ([]core.Procedure, string, error) {
	// ... (implementation unchanged) ...
	return nil, "", nil
}
