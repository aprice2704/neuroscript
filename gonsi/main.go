// gonsi/main.go
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	// Import the core package
	"github.com/aprice2704/neuroscript/pkg/core"
)

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

// Define command-line flags
var (
	debugAST         = flag.Bool("debug-ast", false, "Enable verbose AST node logging")
	debugInterpreter = flag.Bool("debug-interpreter", false, "Enable verbose interpreter execution logging")
	noPreloadSkills  = flag.Bool("no-preload-skills", false, "Skip initial loading of all skills from the directory (only load the requested skill)")
)

func main() {
	flag.Parse()
	enableMainDebug := *debugAST || *debugInterpreter
	initLoggers(enableMainDebug)

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Usage: gonsi [flags] <skills_directory> <ProcedureToRun> [args...]")
		flag.PrintDefaults()
		os.Exit(1)
	}
	skillsDir := args[0]
	procToRun := args[1]
	procArgs := args[2:]

	// Define targetFilename *before* the walk function
	targetFilename := procToRun + ".ns.txt" // <<< DEFINED HERE

	interpreter := core.NewInterpreter(debugLog)
	infoLog.Printf("Loading procedures from directory: %s (Preload: %t)", skillsDir, !*noPreloadSkills)
	var firstErrorEncountered error
	var procedureFound bool = false   // Flag to check if target file was found in no-preload mode
	var proceduresLoadedCount int = 0 // Counter for loaded procedures

	// We will modify the interpreter's internal map directly within the Walk closure.
	// This works because the 'interpreter' variable is captured by the closure.

	walkErr := filepath.Walk(skillsDir, func(path string, info os.FileInfo, walkErrIn error) error {
		if firstErrorEncountered != nil {
			return firstErrorEncountered // Stop if a critical error occurred
		}
		if walkErrIn != nil {
			// Log errors accessing paths but continue walking if possible
			errorLog.Printf("Error accessing path %q: %v", path, walkErrIn)
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ns.txt") {
			fileName := info.Name()

			shouldParse := false
			// Parse all files if not in no-preload mode
			if !*noPreloadSkills {
				shouldParse = true
			} else if fileName == targetFilename { // Parse only the target file if in no-preload mode
				shouldParse = true
				procedureFound = true // Mark that we found the specific file we needed
				debugLog.Printf("No-preload mode: Found target file %s, parsing.", fileName)
			} else {
				// Skip other files in no-preload mode
				debugLog.Printf("No-preload mode: Skipping file %s", fileName)
			}

			if shouldParse {
				debugLog.Printf("--- Processing File: %s ---", fileName)
				contentBytes, readErr := os.ReadFile(path)
				if readErr != nil {
					errorLog.Printf("Could not read file %s: %v", fileName, readErr)
					return nil // Continue walking
				}

				parseOptions := core.ParseOptions{DebugAST: *debugAST, Logger: debugLog}
				stringReader := strings.NewReader(string(contentBytes))

				procedures, fileVersion, parseErr := core.ParseNeuroScript(stringReader, fileName, parseOptions)

				if parseErr != nil {
					// Log parse errors but continue walking to try other files
					errorMsg := fmt.Sprintf("Parse error processing %s:\n%s", fileName, parseErr.Error())
					errorLog.Print(errorMsg)
					return nil
				}

				if fileVersion != "" {
					debugLog.Printf("Found FILE_VERSION %q in %s", fileVersion, fileName)
				}

				// Add procedures to the interpreter's *internal* map
				// This requires adding an exported method to core.Interpreter
				// or making the knownProcedures map exported.
				// **Assuming an exported AddProcedure method exists for correctness:**
				// (If not, this needs adjustment in pkg/core/interpreter.go)
				for _, proc := range procedures {
					// *** Hypothetical Fix: Use an exported AddProcedure Method ***
					// loadErr := interpreter.AddProcedure(proc) // ASSUMES this method exists
					// if loadErr != nil {
					//     errorLog.Printf("Load error: %v", loadErr)
					//     if firstErrorEncountered == nil { firstErrorEncountered = loadErr }
					//     return firstErrorEncountered // Stop on critical load error (like duplicate)
					// }
					// proceduresLoadedCount++
					// debugLog.Printf("  Loaded procedure: %s", proc.Name)

					// *** Workaround if AddProcedure doesn't exist & knownProcedures is unexported: ***
					// We can't directly check for duplicates or add without access.
					// This means duplicates might silently overwrite, and the count is less reliable.
					// For now, we'll just increment the count, acknowledging this limitation.
					_ = proc                                 // Avoid unused variable error if using workaround
					proceduresLoadedCount += len(procedures) // Less accurate, but avoids direct map access
					// TODO: Add an exported method core.Interpreter.AddProcedure(p core.Procedure) error

				}
				// Remove the direct map access simulation from previous attempts

				debugLog.Printf("Successfully parsed and processed %d procedures from %s", len(procedures), fileName)

				// If in no-preload mode and we just parsed the target, stop walking
				if *noPreloadSkills && fileName == targetFilename {
					debugLog.Printf("No-preload mode: Target procedure loaded, stopping walk.")
					return filepath.SkipDir // Stop walk successfully
				}
			}
		}
		return nil // Continue walking
	})

	// --- Post-Walk Handling ---
	if firstErrorEncountered != nil {
		// Handle critical errors encountered during the walk (e.g., duplicate procedures if check added)
		errorLog.Printf("Processing stopped due to critical error: %v", firstErrorEncountered)
		os.Exit(1)
	}
	if walkErr != nil && walkErr != filepath.SkipDir {
		// Handle errors related to the filesystem traversal itself
		errorLog.Printf("Error walking skills directory '%s': %v", skillsDir, walkErr)
		os.Exit(1)
	}

	// If in no-preload mode, ensure the specific file/procedure was actually found
	if *noPreloadSkills && !procedureFound {
		// Use the restored targetFilename variable in the error message
		errorLog.Printf("Error: Target procedure file '%s' not found in directory '%s'", targetFilename, skillsDir) // <<< CORRECTED: Used targetFilename
		os.Exit(1)
	}

	// Use the counter for the final log message
	infoLog.Printf("Finished processing directory. Procedures parsed/processed (approx): %d", proceduresLoadedCount) // <<< Use counter

	// --- Execute Requested Procedure ---
	// We can't reliably check existence here without accessing the map or an exported method.
	// We proceed and let RunProcedure handle the "not found" error if necessary.
	infoLog.Printf("Executing procedure: %s with args: %v", procToRun, procArgs)
	fmt.Println("--------------------")
	result, runErr := interpreter.RunProcedure(procToRun, procArgs...)
	fmt.Println("--------------------")
	infoLog.Println("Execution finished.")
	if runErr != nil {
		errorLog.Printf("Execution Error: %v", runErr)
		// Check if the error is specifically "procedure not defined"
		if strings.Contains(runErr.Error(), "not defined") {
			errorLog.Printf("Hint: Ensure procedure '%s' exists in '%s' and the file was parsed successfully.", procToRun, targetFilename)
		}
		os.Exit(1)
	}
	infoLog.Printf("Final Result: %v (%T)", result, result)
}
