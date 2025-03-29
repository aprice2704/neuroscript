// gonsi/main.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Use your actual module path here if different
	"github.com/aprice2704/neuroscript/pkg/core"
)

func main() {
	// Updated arguments: main.go <skills_directory> <ProcedureToRun> [args...]
	if len(os.Args) < 3 {
		fmt.Println("Usage: gonsi <skills_directory> <ProcedureToRun> [args...]")
		os.Exit(1)
	}
	skillsDir := os.Args[1]
	procToRun := os.Args[2]
	procArgs := []string{} // Initialize as empty string slice
	if len(os.Args) > 3 {
		procArgs = os.Args[3:]
	}

	// --- Load Procedures ---
	interpreter := core.NewInterpreter()
	fmt.Printf("Loading procedures from directory: %s\n", skillsDir)

	// Walk the directory to find .ns files
	walkErr := filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Error accessing path, report and potentially stop walking?
			// For now, let's report and continue.
			fmt.Printf("    Warning: Error accessing path %q: %v\n", path, err)
			return nil // Continue walking if possible
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ns") {
			// Attempt to open the file
			file, openErr := os.Open(path)
			if openErr != nil {
				fmt.Printf("    [Error] Could not open file %s: %v\n", filepath.Base(path), openErr)
				return nil // Continue walking
			}
			defer file.Close() // Ensure file is closed

			// Parse the file using the new parser entry point
			procedures, parseErr := core.ParseNeuroScript(file) // Use new parser API
			if parseErr != nil {
				// ** Log error with filename and continue walking **
				wrappedErr := fmt.Errorf("parsing '%s': %w", filepath.Base(path), parseErr)
				fmt.Printf("    [Warning] Skipping file due to error: %v\n", wrappedErr)
				return nil // <<< Return nil here to continue walking
			}

			// Load successfully parsed procedures
			if loadErr := interpreter.LoadProcedures(procedures); loadErr != nil {
				// Log error and continue walking
				wrappedErr := fmt.Errorf("loading from '%s': %w", filepath.Base(path), loadErr)
				fmt.Printf("    [Warning] Skipping file due to error: %v\n", wrappedErr)
				return nil // <<< Return nil here to continue walking
			}
			// Optional: Log successful load
			// fmt.Printf("    Loaded %d procedures from %s\n", len(procedures), filepath.Base(path))
		}
		return nil // Continue walking normally
	})

	// Check for fatal error during walk itself (not file parsing errors)
	if walkErr != nil {
		fmt.Printf("Error walking skills directory '%s': %v\n", skillsDir, walkErr)
		os.Exit(1)
	}

	fmt.Println("Finished loading procedures.")

	// --- Execute Requested Procedure ---
	fmt.Printf("\nExecuting procedure: %s with args: %v\n", procToRun, procArgs)
	fmt.Println("--------------------")

	result, runErr := interpreter.RunProcedure(procToRun, procArgs...) // Use variadic slice expansion

	fmt.Println("--------------------")
	fmt.Println("Execution finished.")

	if runErr != nil {
		fmt.Printf("Execution Error: %v\n", runErr)
		os.Exit(1) // Exit with error code if execution failed
	}

	// Only print result if execution was successful
	fmt.Printf("Final Result: %v\n", result)
	// Exit normally
}
