// filename: cmd/syntax-check/main.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/logging"
)

func main() {
	fmt.Println("--- Running Syntax Smoke Test ---")
	logger := logging.NewNoOpLogger()
	parserAPI := NewParserAPI(logger)

	var unexpectedFailures []string
	var unexpectedSuccesses []string

	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}
	testDataDir := filepath.Join(rootDir, "pkg", "core", "testdata")
	fmt.Printf("Scanning for .ns files in: %s\n", testDataDir)

	err := filepath.Walk(testDataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".ns") || strings.HasSuffix(path, ".ns.txt")) {
			baseName := filepath.Base(path)
			isInvalidTest := strings.HasPrefix(baseName, "invalid_")

			fmt.Printf("Checking: %s...", path)
			contentBytes, readErr := os.ReadFile(path)
			if readErr != nil {
				fmt.Printf(" FAILED TO READ\n")
				unexpectedFailures = append(unexpectedFailures, fmt.Sprintf("%s (read error: %v)", path, readErr))
				return nil
			}

			_, parseErr := parserAPI.Parse(string(contentBytes))

			if isInvalidTest {
				if parseErr == nil {
					fmt.Printf(" PARSE SUCCEEDED (UNEXPECTED)\n")
					unexpectedSuccesses = append(unexpectedSuccesses, path)
				} else {
					fmt.Printf(" OK (failed as expected)\n")
				}
			} else {
				if parseErr != nil {
					fmt.Printf(" PARSE FAILED (UNEXPECTED)\n")
					unexpectedFailures = append(unexpectedFailures, fmt.Sprintf("%s (parse error: %v)", path, parseErr))
				} else {
					fmt.Printf(" OK\n")
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError walking testdata directory: %v\n", err)
		os.Exit(1)
	}

	// Report results and exit if there were any problems.
	hasErrors := false
	if len(unexpectedFailures) > 0 {
		hasErrors = true
		fmt.Fprintf(os.Stderr, "\n--- SYNTAX CHECK FAILED ---\n")
		fmt.Fprintf(os.Stderr, "The following valid files failed to parse:\n")
		for _, file := range unexpectedFailures {
			fmt.Fprintf(os.Stderr, "- %s\n", file)
		}
	}
	if len(unexpectedSuccesses) > 0 {
		hasErrors = true
		fmt.Fprintf(os.Stderr, "\n--- SYNTAX CHECK FAILED ---\n")
		fmt.Fprintf(os.Stderr, "The following invalid files parsed successfully (they may need to be updated):\n")
		for _, file := range unexpectedSuccesses {
			fmt.Fprintf(os.Stderr, "- %s\n", file)
		}
	}

	if hasErrors {
		os.Exit(1)
	}

	fmt.Println("\n--- SYNTAX CHECK PASSED ---")
}
