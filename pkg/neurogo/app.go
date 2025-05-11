// NeuroScript Version: 0.3.0
// File version: 0.1.1 // Changed INFO logs to DEBUG
// Refactored App struct and methods for AI Worker Manager integration
// filename: pkg/neurogo/app.go
package neurogo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync" // Added for Mutex
	"time" // Added for ExecuteScriptFile timing

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/neurodata/models" // Keep the import for type signature
	// Import nspatch package if needed by PatchHandler interface/impl
	// "github.com/aprice2704/neuroscript/pkg/nspatch"
	// Keep for registration logic if decided later
)

// App orchestrates the NeuroScript application components.
type App struct {
	Config *Config
	Log    logging.Logger // Use the Logger interface

	// LLM Interaction
	llmClient core.LLMClient

	// Interpreter and State
	interpreter     *core.Interpreter
	aiWorkerManager *core.AIWorkerManager // Added AI Worker Manager
	mu              sync.RWMutex          // Use RWMutex for potentially more reads than writes

	// TUI components (if TUI is enabled)
	// tui *tea.Program // Assuming bubbletea is used

	// Patch Handling
	patchHandler PatchHandler // Use the interface defined in app_interface.go

	// Loaded Schema (optional, depends on application needs)
	loadedSchema *models.Schema // Keep the field for future use
}

// NewApp creates a new NeuroGo application instance.
func NewApp(logger logging.Logger) *App {
	if logger == nil {
		// This case should be prevented by main.go, but handle defensively.
		logger = adapters.NewNoOpLogger() // Use NoOpLogger as a safe fallback
		fmt.Fprintf(os.Stderr, "Critical Warning: NewApp called with nil logger. Using NoOpLogger.\n")
	} else {
		logger.Debug("Creating new App instance.")
	}
	return &App{
		Log:    logger,
		Config: NewConfig(), // Initialize config immediately
	}
}

// SetInterpreter sets the interpreter instance for the app.
func (app *App) SetInterpreter(interpreter *core.Interpreter) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.interpreter = interpreter
}

// SetAIWorkerManager sets the AI Worker Manager instance for the app.
func (app *App) SetAIWorkerManager(manager *core.AIWorkerManager) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.aiWorkerManager = manager
	// Optionally link it to the interpreter if needed
	if app.interpreter != nil {
		app.interpreter.SetAIWorkerManager(manager) // Assuming this method exists in core.Interpreter
	} else {
		app.Log.Warn("SetAIWorkerManager called before interpreter was set.")
	}
}

// GetInterpreter returns the application's core interpreter instance safely.
func (app *App) GetInterpreter() *core.Interpreter {
	app.mu.RLock()
	defer app.mu.RUnlock()
	if app.interpreter == nil {
		app.Log.Error("GetInterpreter called before interpreter was initialized!")
	}
	return app.interpreter
}

// GetAIWorkerManager returns the AI Worker Manager instance safely.
func (app *App) GetAIWorkerManager() *core.AIWorkerManager {
	app.mu.RLock()
	defer app.mu.RUnlock()
	if app.aiWorkerManager == nil {
		app.Log.Warn("GetAIWorkerManager called before AI Worker Manager was initialized.")
	}
	return app.aiWorkerManager
}

// CreateLLMClient creates the LLM client based on config.
// Renamed from createLLMClient for clarity as it's called externally now.
func (app *App) CreateLLMClient() (core.LLMClient, error) {
	if app.Config == nil {
		return nil, fmt.Errorf("cannot create LLM client: app config is nil")
	}

	// LLM is now implicitly enabled if API key is provided.
	// Create NoOp client only if API key is truly missing.
	apiKey := app.Config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("NEUROSCRIPT_API_KEY") // Standardized env var name
		if apiKey == "" {
			app.Log.Debug("API key is missing in config and environment variable (NEUROSCRIPT_API_KEY). Creating NoOpLLMClient.") // Changed from Info
			return adapters.NewNoOpLLMClient(), nil
		}
		app.Log.Debug("Using LLM API key from environment variable NEUROSCRIPT_API_KEY.") // Changed from Info
	} else {
		app.Log.Debug("Using LLM API key from configuration.")
	}

	app.Log.Debug("Creating real LLMClient.") // Changed from Info
	apiHost := app.Config.APIHost
	modelName := app.Config.ModelName

	// Ensure logger is not nil before passing
	logger := app.Log
	if logger == nil {
		logger = adapters.NewNoOpLogger() // Safety fallback
	}

	// Use NewLLMClient from core package
	llmClient := core.NewLLMClient(apiKey, apiHost, modelName, logger, !app.Config.Insecure) // Pass insecure flag inversely
	if llmClient == nil {
		app.Log.Error("core.NewLLMClient returned nil unexpectedly.")
		return nil, fmt.Errorf("failed to create LLM client instance (core.NewLLMClient returned nil)")
	}

	app.Log.Debug("Real LLMClient created.", "host", apiHost, "model", modelName) // Changed from Info
	return llmClient, nil
}

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
			if err := interp.AddProcedure(*proc); err != nil { // Assumes AddProcedure takes value
				a.Log.Error("Failed to add procedure to interpreter.", "procedure", name, "path", filePath, "error", err)
				return definedProcs, fileMetadata, fmt.Errorf("failed to add procedure '%s' from '%s': %w", name, filePath, err)
			} else {
				a.Log.Debug("Added procedure to interpreter.", "procedure", name, "path", filePath)
				definedProcs = append(definedProcs, proc)
			}
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
	// Still requires RunProcedure to accept a map or handle args differently.
	// Sticking with the temporary fix: Run without args from map for now.
	var runErr error
	var results interface{}
	if len(procArgsMap) > 0 {
		app.Log.Warn("Procedure arguments provided via flags/map, but current RunProcedure call doesn't support passing them easily. Executing procedure without these arguments.", "procedure", procedureToRun)
		results, runErr = interpreter.RunProcedure(procedureToRun) // No args passed
	} else {
		results, runErr = interpreter.RunProcedure(procedureToRun) // No args passed
	}
	// --- End RunProcedure Call ---

	execEndTime := time.Now()
	duration := execEndTime.Sub(execStartTime)
	app.Log.Debug("Procedure execution finished.", "name", procedureToRun, "duration", duration) // Changed from Info

	if runErr != nil {
		app.Log.Error("Script execution failed.", "procedure", procedureToRun, "error", runErr)
		// Maybe don't print directly to Stderr here, let the caller handle it
		return fmt.Errorf("error executing procedure '%s': %w", procedureToRun, runErr)
	}

	app.Log.Debug("Script executed successfully.", "procedure", procedureToRun) // Changed from Info
	if results != nil {
		// Maybe return results or log them, avoid printing directly here
		app.Log.Debug("Script Result Value", "result", fmt.Sprintf("%+v", results))
	} else {
		app.Log.Debug("Script Result Value: nil")
	}

	totalDuration := time.Since(startTime)
	app.Log.Debug("--- Script File Execution Finished ---", "path", scriptPath, "total_duration", totalDuration) // Changed from Info
	return nil
}

// Placeholder for Run method - contents moved to main.go or other methods
func (app *App) Run(ctx context.Context) error {
	app.Log.Error("App.Run called directly, this logic should now be in main.go or specific handlers.")
	return fmt.Errorf("App.Run is deprecated; execution flow managed by main.go")
}

// --- Methods implementing AppAccess interface for TUI ---
// These remain largely the same, using the Config fields

func (a *App) GetModelName() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.Config == nil {
		return ""
	}
	return a.Config.ModelName // Use ModelName
}

func (a *App) GetSyncDir() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.Config == nil {
		return ""
	}
	return a.Config.SyncDir
}

func (a *App) GetSandboxDir() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.Config == nil {
		return ""
	}
	return a.Config.SandboxDir
}

func (a *App) GetSyncFilter() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.Config == nil {
		return ""
	}
	return a.Config.SyncFilter
}

func (a *App) GetSyncIgnoreGitignore() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.Config == nil {
		return false
	}
	return a.Config.SyncIgnoreGitignore
}

func (a *App) GetLogger() logging.Logger {
	// No lock needed typically as Log is set at creation
	if a.Log == nil {
		fmt.Fprintf(os.Stderr, "Warning: GetLogger called when App.Log is nil. Returning NoOpLogger.\n")
		return adapters.NewNoOpLogger() // Return NoOp instead of nil
	}
	return a.Log
}

func (a *App) GetLLMClient() core.LLMClient {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.llmClient == nil {
		// Logger might be nil if called very early or in error state
		logger := a.GetLogger()
		logger.Warn("GetLLMClient called when App.llmClient is nil.")
		return nil // Return nil as getting a NoOp client might hide issues
	}
	return a.llmClient
}
