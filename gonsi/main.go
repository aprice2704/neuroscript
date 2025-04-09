// gonsi/main.go
package main

import (
	// Import errors package
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	core "github.com/aprice2704/neuroscript/pkg/core"
	blocks "github.com/aprice2704/neuroscript/pkg/neurodata/blocks"
	// checklist "github.com/aprice2704/neuroscript/pkg/neurodata/checklist"
)

// --- Custom type for repeatable string flag ---
type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSlice) Set(value string) error {
	if strings.Contains(value, "..") {
		cleanPath := filepath.Clean(value)
		if strings.HasPrefix(cleanPath, "..") {
			// Proceed and let SecureFilePath handle it later.
		}
	}
	*s = append(*s, value)
	return nil
}

// --- Logger Setup ---
var (
	infoLog  *log.Logger
	debugLog *log.Logger
	errorLog *log.Logger
)

func initLoggers(enableDebug bool) {
	infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugOutput := io.Discard
	if enableDebug {
		debugOutput = os.Stderr
		fmt.Fprintf(os.Stderr, "--- Debug Logging Enabled ---\n")
	}
	debugLog = log.New(debugOutput, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	log.SetOutput(os.Stderr)
	log.SetPrefix("FATAL: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// --- Command-line flags ---
var (
	debugAST         = flag.Bool("debug-ast", false, "Enable verbose AST node logging")
	debugInterpreter = flag.Bool("debug-interpreter", false, "Enable verbose interpreter execution logging")
	libPaths         stringSlice // Use custom type for repeatable flag
)

func init() {
	flag.Var(&libPaths, "lib", "Specify a library path (file or directory) to load procedures from (repeatable)")
}

// --- Procedure Loading Helper ---
func processNeuroScriptFile(path string, interp *core.Interpreter, enableASTDebug bool) (loadedProcs []core.Procedure, fileVersion string, err error) {
	fileName := filepath.Base(path)
	debugLog.Printf("--- Processing File: %s ---", path)

	contentBytes, readErr := os.ReadFile(path)
	if readErr != nil {
		errorLog.Printf("Could not read file %s: %v", path, readErr)
		return nil, "", fmt.Errorf("read error for %s: %w", path, readErr)
	}

	parseOptions := core.ParseOptions{DebugAST: enableASTDebug, Logger: debugLog}
	stringReader := strings.NewReader(string(contentBytes))
	procedures, fileVersion, parseErr := core.ParseNeuroScript(stringReader, fileName, parseOptions)

	if parseErr != nil {
		errorMsg := fmt.Sprintf("Parse error processing %s:\n%s", path, parseErr.Error())
		errorLog.Print(errorMsg)
		return nil, fileVersion, fmt.Errorf("parse error for %s: %w", path, parseErr)
	}

	if fileVersion != "" {
		debugLog.Printf("Found FILE_VERSION %q in %s", fileVersion, path)
	}

	addedProcedures := make([]core.Procedure, 0, len(procedures))
	for _, proc := range procedures {
		loadErr := interp.AddProcedure(proc)
		if loadErr != nil {
			errorLog.Printf("Load error adding procedure '%s' from %s: %v", proc.Name, path, loadErr)
		} else {
			debugLog.Printf("  Successfully added procedure '%s' from %s.", proc.Name, path)
			addedProcedures = append(addedProcedures, proc)
		}
	}
	if len(addedProcedures) > 0 {
		debugLog.Printf("  Added %d procedures in total from %s.", len(addedProcedures), path)
	}

	return addedProcedures, fileVersion, nil
}

// --- Main ---
func main() {
	flag.Parse()
	enableMainDebug := *debugAST || *debugInterpreter
	initLoggers(enableMainDebug)

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("Usage: gonsi [flags] <ProcedureToRun | FileToRun.ns.txt> [proc_args...]")
		fmt.Println("  Error: Missing procedure name or filename to run.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	targetArg := args[0]
	procArgs := args[1:]

	infoLog.Printf("Optional library paths specified (-lib): %v", libPaths)
	infoLog.Printf("Target to run: %s", targetArg)
	infoLog.Printf("Procedure args: %v", procArgs)

	interpreter := core.NewInterpreter(debugLog)

	coreRegistry := interpreter.ToolRegistry()
	if coreRegistry == nil {
		log.Fatal("Interpreter's ToolRegistry is nil after creation.")
	}
	core.RegisterCoreTools(coreRegistry)
	blocks.RegisterBlockTools(coreRegistry)
	// checklist.RegisterChecklistTools(coreRegistry)

	var proceduresLoadedCount int = 0
	if len(libPaths) > 0 {
		infoLog.Printf("Loading procedures from specified library paths...")
		for _, pathArg := range libPaths {
			pathInfo, statErr := os.Stat(pathArg)
			if statErr != nil {
				errorLog.Printf("Error accessing library path '%s': %v. Skipping.", pathArg, statErr)
				continue
			}

			if pathInfo.IsDir() {
				infoLog.Printf("Scanning library directory: %s", pathArg)
				walkErr := filepath.Walk(pathArg, func(path string, info os.FileInfo, walkErrIn error) error {
					if walkErrIn != nil {
						errorLog.Printf("Error accessing path %q during library walk: %v", path, walkErrIn)
						return nil
					}
					if !info.IsDir() && (strings.HasSuffix(info.Name(), ".ns.txt") || strings.HasSuffix(info.Name(), ".ns") || strings.HasSuffix(info.Name(), ".neuro")) {
						_, _, _ = processNeuroScriptFile(path, interpreter, *debugAST)
					}
					return nil
				})
				if walkErr != nil {
					errorLog.Printf("Error walking library directory '%s': %v.", pathArg, walkErr)
				}
			} else {
				if strings.HasSuffix(pathInfo.Name(), ".ns.txt") || strings.HasSuffix(pathInfo.Name(), ".ns") || strings.HasSuffix(pathInfo.Name(), ".neuro") {
					infoLog.Printf("Loading library file: %s", pathArg)
					_, _, _ = processNeuroScriptFile(pathArg, interpreter, *debugAST)
				} else {
					infoLog.Printf("Skipping non-NeuroScript file specified via -lib: %s", pathArg)
				}
			}
		}
		infoLog.Printf("Finished processing library paths specified via -lib.")
	} else {
		infoLog.Printf("No library paths specified via -lib.")
	}

	var procToRun string
	isFilePathTarget := strings.HasSuffix(targetArg, ".ns.txt") || strings.HasSuffix(targetArg, ".ns") || strings.HasSuffix(targetArg, ".neuro")

	if isFilePathTarget {
		infoLog.Printf("Target '%s' looks like a file. Loading it directly.", targetArg)
		cwd, errWd := os.Getwd()
		if errWd != nil {
			log.Fatalf("Failed to get working directory: %v", errWd)
		}
		// Use exported core.SecureFilePath
		_, secErr := core.SecureFilePath(targetArg, cwd)
		if secErr != nil {
			log.Fatalf("Target file path error: %v", secErr)
		}

		loadedProcs, _, loadErr := processNeuroScriptFile(targetArg, interpreter, *debugAST)
		if loadErr != nil {
			log.Fatalf("Failed to load target file '%s': %v", targetArg, loadErr)
		}
		if len(loadedProcs) == 0 {
			log.Fatalf("Target file '%s' loaded successfully but contains no valid procedures.", targetArg)
		}
		procToRun = loadedProcs[0].Name
		infoLog.Printf("Running first procedure '%s' found in file '%s'.", procToRun, targetArg)
	} else {
		procToRun = targetArg
		infoLog.Printf("Target '%s' treated as procedure name.", procToRun)
		// Use exported core.Interpreter.KnownProcedures()
		if _, exists := interpreter.KnownProcedures()[procToRun]; !exists {
			errorLog.Printf("Warning: Procedure '%s' not found in loaded libraries. Execution will likely fail.", procToRun)
		}
	}

	// Use exported core.Interpreter.KnownProcedures()
	proceduresLoadedCount = len(interpreter.KnownProcedures())
	if proceduresLoadedCount == 0 {
		errorLog.Printf("Warning: No procedures were successfully loaded overall.")
	} else {
		infoLog.Printf("Total procedures loaded and ready: %d", proceduresLoadedCount)
	}

	infoLog.Printf("Attempting to execute procedure: '%s' with args: %v", procToRun, procArgs)
	fmt.Println("--------------------")
	result, runErr := interpreter.RunProcedure(procToRun, procArgs...)
	fmt.Println("--------------------")
	infoLog.Println("Execution finished.")

	if runErr != nil {
		if strings.Contains(runErr.Error(), "not defined or not loaded") {
			errorLog.Printf("Execution Error: %v", runErr)
			if isFilePathTarget {
				errorLog.Printf("Hint: Procedure '%s' (from file '%s') failed to execute.", procToRun, targetArg)
			} else {
				errorLog.Printf("Hint: Procedure '%s' was not found in the specified library paths (-lib) or there were loading errors.", procToRun)
			}
		} else {
			errorLog.Printf("Execution Error: %v", runErr)
		}
		os.Exit(1)
	}

	infoLog.Printf("Final Result: %v (%T)", result, result)
}

// --- REMOVED Helper Function - Moved to core.Interpreter ---
// func (i *Interpreter) KnownProcedures() map[string]core.Procedure { ... }
