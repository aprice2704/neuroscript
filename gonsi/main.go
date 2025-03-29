package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core" // Adjust if needed
)

func main() {
	// Updated arguments: main.go <skills_directory> <ProcedureToRun> [args...]
	if len(os.Args) < 3 {
		fmt.Println("Usage: gonsi <skills_directory> <ProcedureToRun> [args...]")
		os.Exit(1)
	}
	skillsDir := os.Args[1]
	procToRun := os.Args[2]
	// ** Capture procedure arguments from command line **
	procArgs := []string{} // Initialize as empty string slice
	if len(os.Args) > 3 {
		procArgs = os.Args[3:]
	}

	// --- Load Procedures ---
	interpreter := core.NewInterpreter()
	fmt.Printf("Loading procedures from directory: %s\n", skillsDir)
	err := filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
		// ... (loading logic unchanged) ...
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ns") {
			// fmt.Printf("  Loading file: %s\n", path) // Reduce noise maybe
			file, openErr := os.Open(path)
			if openErr != nil {
				fmt.Printf("    Error opening %s: %v\n", path, openErr)
				return nil
			}
			defer file.Close()
			procedures, parseErr := core.ParseNeuroScript(file)
			if parseErr != nil {
				fmt.Printf("    Error parsing %s: %v\n", path, parseErr)
				return nil
			}
			if loadErr := interpreter.LoadProcedures(procedures); loadErr != nil {
				fmt.Printf("    Error loading procedures from %s: %v\n", path, loadErr)
				return nil
			}
			// fmt.Printf("    Loaded %d procedures from %s\n", len(procedures), path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking skills directory '%s': %v\n", skillsDir, err)
		os.Exit(1)
	}
	fmt.Println("Finished loading procedures.")

	// --- Execute Requested Procedure ---
	fmt.Printf("\nExecuting procedure: %s with args: %v\n", procToRun, procArgs) // Log args
	fmt.Println("--------------------")

	// ** Pass captured procArgs to RunProcedure **
	result, runErr := interpreter.RunProcedure(procToRun, procArgs...) // Use variadic slice expansion

	fmt.Println("--------------------")
	fmt.Println("Execution finished.")

	if runErr != nil {
		fmt.Printf("Execution Error: %v\n", runErr)
		os.Exit(1)
	}
	fmt.Printf("Final Result: %v\n", result)
}
