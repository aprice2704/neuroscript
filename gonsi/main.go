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
	core "github.com/aprice2704/neuroscript/pkg/core"

	// Import new neurodata packages
	blocks "github.com/aprice2704/neuroscript/pkg/neurodata/blocks"
	checklist "github.com/aprice2704/neuroscript/pkg/neurodata/checklist"
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
	// Set default log output to stderr for consistency, even for FATAL
	log.SetOutput(os.Stderr)
	log.SetPrefix("FATAL: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// --- Command-line flags ---
var (
	debugAST         = flag.Bool("debug-ast", false, "Enable verbose AST node logging")
	debugInterpreter = flag.Bool("debug-interpreter", false, "Enable verbose interpreter execution logging")
	noPreloadSkills  = flag.Bool("no-preload-skills", false, "Skip initial loading of all skills from the directory (only load the requested skill)")
)

func main() {
	flag.Parse()
	enableMainDebug := *debugAST || *debugInterpreter
	initLoggers(enableMainDebug)

	// --- Arg Parsing ---
	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Usage: gonsi [flags] <skills_directory> <ProcedureToRun> [args...]")
		flag.PrintDefaults()
		os.Exit(1)
	}
	skillsDir := args[0]
	procToRun := args[1]
	procArgs := args[2:]
	targetFilename := procToRun + ".ns.txt" // Define target filename

	// Create Interpreter (includes ToolRegistry)
	interpreter := core.NewInterpreter(debugLog)

	// --- Tool Registration ---
	coreRegistry := interpreter.ToolRegistry() // Use the getter
	if coreRegistry == nil {
		log.Fatal("Interpreter's ToolRegistry is nil after creation.") // Add nil check
	}
	// 1. Register core tools
	core.RegisterCoreTools(coreRegistry)
	// 2. Register checklist tools
	checklist.RegisterChecklistTools(coreRegistry)
	// 3. Register block tools
	blocks.RegisterBlockTools(coreRegistry)

	// --- Load Procedures ---
	infoLog.Printf("Loading procedures from directory: %s (Preload: %t)", skillsDir, !*noPreloadSkills)
	var firstErrorEncountered error
	var procedureFound bool = false
	var proceduresLoadedCount int = 0 // Tracks successfully added procedures

	walkErr := filepath.Walk(skillsDir, func(path string, info os.FileInfo, walkErrIn error) error {
		if firstErrorEncountered != nil {
			return firstErrorEncountered // Stop if a critical error occurred
		}
		if walkErrIn != nil {
			errorLog.Printf("Error accessing path %q: %v", path, walkErrIn)
			return nil // Continue walking if possible
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ns.txt") {
			fileName := info.Name()
			shouldParse := !*noPreloadSkills || (fileName == targetFilename)

			if *noPreloadSkills && fileName == targetFilename {
				procedureFound = true
				debugLog.Printf("No-preload mode: Found target file %s, parsing.", fileName)
			} else if *noPreloadSkills {
				// Only log skipping if debug is enabled to avoid clutter
				if enableMainDebug {
					debugLog.Printf("No-preload mode: Skipping file %s", fileName)
				}
				return nil // Skip this file
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
					errorMsg := fmt.Sprintf("Parse error processing %s:\n%s", fileName, parseErr.Error())
					errorLog.Print(errorMsg)
					return nil // Continue walking, but don't load from this file
				}

				if fileVersion != "" {
					debugLog.Printf("Found FILE_VERSION %q in %s", fileVersion, fileName)
				}

				// Load procedures using the exported method
				loadedFromFile := 0
				for _, proc := range procedures {
					loadErr := interpreter.AddProcedure(proc) // Use AddProcedure
					if loadErr != nil {
						// Log duplicate/load errors but continue walking unless it's critical
						errorLog.Printf("Load error adding procedure '%s' from %s: %v", proc.Name, fileName, loadErr)
						// Decide if this error should stop the whole process
						// For now, let's allow duplicates from different files but log it.
						// If AddProcedure returned a specific critical error type, handle it here.
						// firstErrorEncountered = loadErr // Uncomment to make loading errors critical
						// return firstErrorEncountered
					} else {
						loadedFromFile++ // Count successfully added procedures
					}
				}
				if loadedFromFile > 0 {
					proceduresLoadedCount += loadedFromFile
					debugLog.Printf("  Successfully added %d procedures from %s.", loadedFromFile, fileName)
				}

				// If in no-preload mode and we just parsed the target, stop walking
				if *noPreloadSkills && fileName == targetFilename {
					debugLog.Printf("No-preload mode: Target procedure file parsed and procedures added, stopping walk.")
					return filepath.SkipDir // Stop walk successfully
				}
			}
		}
		return nil // Continue walking
	})

	// --- Post-Walk Handling ---
	if firstErrorEncountered != nil {
		errorLog.Printf("Processing stopped due to critical error during procedure loading: %v", firstErrorEncountered)
		os.Exit(1)
	}
	if walkErr != nil && walkErr != filepath.SkipDir {
		errorLog.Printf("Error walking skills directory '%s': %v", skillsDir, walkErr)
		os.Exit(1)
	}
	if *noPreloadSkills && !procedureFound {
		errorLog.Printf("Error: Target procedure file '%s' not found in directory '%s'", targetFilename, skillsDir)
		os.Exit(1)
	}

	infoLog.Printf("Finished processing directory. Procedures loaded: %d", proceduresLoadedCount) // Use accurate count

	// --- Execute Requested Procedure ---
	infoLog.Printf("Executing procedure: %s with args: %v", procToRun, procArgs)
	fmt.Println("--------------------") // Separator before execution output
	result, runErr := interpreter.RunProcedure(procToRun, procArgs...)
	fmt.Println("--------------------") // Separator after execution output
	infoLog.Println("Execution finished.")
	if runErr != nil {
		errorLog.Printf("Execution Error: %v", runErr)
		// Add hint if procedure not found after loading attempt
		if strings.Contains(runErr.Error(), "not defined or not loaded") {
			errorLog.Printf("Hint: Ensure procedure '%s' exists in '%s' and was loaded without errors (check previous logs).", procToRun, targetFilename)
		}
		os.Exit(1)
	}
	infoLog.Printf("Final Result: %v (%T)", result, result)
}
