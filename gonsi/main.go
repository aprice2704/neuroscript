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
	// debugTokens is no longer used directly, lexer errors handled differently
	// debugTokens      = flag.Bool("debug-tokens", false, "Enable verbose token logging")
	debugAST         = flag.Bool("debug-ast", false, "Enable verbose AST node logging")
	debugInterpreter = flag.Bool("debug-interpreter", false, "Enable verbose interpreter execution logging")
	noPreloadSkills  = flag.Bool("no-preload-skills", false, "Skip initial loading of all skills from the directory (only load the requested skill)")
)

func main() {
	flag.Parse()
	// Removed debugTokens from this check
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

	interpreter := core.NewInterpreter(debugLog)
	infoLog.Printf("Loading procedures from directory: %s (Preload: %t)", skillsDir, !*noPreloadSkills)
	var firstErrorEncountered error
	var procedureFound bool = false

	targetFilename := procToRun + ".ns.txt"

	walkErr := filepath.Walk(skillsDir, func(path string, info os.FileInfo, walkErrIn error) error {
		if firstErrorEncountered != nil {
			return firstErrorEncountered
		}
		if walkErrIn != nil {
			errorLog.Printf("Error accessing path %q: %v", path, walkErrIn)
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ns.txt") {
			fileName := info.Name()

			shouldParse := false
			if !*noPreloadSkills {
				shouldParse = true
			} else if fileName == targetFilename {
				shouldParse = true
				procedureFound = true
				debugLog.Printf("No-preload mode: Found target file %s, parsing.", fileName)
			} else {
				debugLog.Printf("No-preload mode: Skipping file %s", fileName)
			}

			if shouldParse {
				debugLog.Printf("--- Processing File: %s ---", fileName)
				contentBytes, readErr := os.ReadFile(path)
				if readErr != nil {
					errorLog.Printf("Could not read file %s: %v", fileName, readErr)
					return nil
				}

				parseOptions := core.ParseOptions{DebugAST: *debugAST, Logger: debugLog}
				stringReader := strings.NewReader(string(contentBytes))

				// *** UPDATED CALL to ParseNeuroScript ***
				procedures, fileVersion, parseErr := core.ParseNeuroScript(stringReader, fileName, parseOptions)

				if parseErr != nil {
					errorMsg := fmt.Sprintf("Parse error processing %s:\n%s", fileName, parseErr.Error())
					errorLog.Print(errorMsg)
					if firstErrorEncountered == nil {
						firstErrorEncountered = fmt.Errorf(errorMsg)
					}
					return firstErrorEncountered // Stop walk on parse error
				}

				// Log the file version if found
				if fileVersion != "" {
					debugLog.Printf("Found FILE_VERSION %q in %s", fileVersion, fileName)
					// TODO: Potentially store or use the fileVersion later
				}

				// Load procedures
				if loadErr := interpreter.LoadProcedures(procedures); loadErr != nil {
					wrappedErr := fmt.Errorf("loading from '%s': %w", fileName, loadErr)
					errorLog.Printf("Load error: %v", wrappedErr)
					if firstErrorEncountered == nil {
						firstErrorEncountered = wrappedErr
					}
					return firstErrorEncountered
				}
				debugLog.Printf("Successfully parsed and loaded %d procedures from %s", len(procedures), fileName)

				if *noPreloadSkills && fileName == targetFilename {
					debugLog.Printf("No-preload mode: Target procedure loaded, stopping walk.")
					return filepath.SkipDir
				}
			}
		}
		return nil
	})

	// --- Post-Walk Handling ---
	if firstErrorEncountered != nil {
		infoLog.Printf("Directory walk stopped due to error: %v", firstErrorEncountered)
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

	infoLog.Println("Finished loading procedures successfully.")

	// --- Execute Requested Procedure ---
	infoLog.Printf("Executing procedure: %s with args: %v", procToRun, procArgs)
	fmt.Println("--------------------")
	result, runErr := interpreter.RunProcedure(procToRun, procArgs...)
	fmt.Println("--------------------")
	infoLog.Println("Execution finished.")
	if runErr != nil {
		errorLog.Printf("Execution Error: %v", runErr)
		os.Exit(1)
	}
	infoLog.Printf("Final Result: %v", result)
}
