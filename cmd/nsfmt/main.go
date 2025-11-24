// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Adds support for -h/--help flag to display usage information.
// filename: cmd/nsfmt/main.go
// nlines: 83

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/aprice2704/neuroscript/pkg/nsfmt"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		input []byte
		err   error
	)

	// nsfmt operates like gofmt:
	// - If no arguments are given, read from stdin.
	// - If arguments (files) are given, process the first one.
	//   (A more complex version would process all, but this is a simple start).
	args := os.Args[1:]

	// Check for help flag
	if len(args) > 0 {
		if args[0] == "-h" || args[0] == "--help" {
			printHelp()
			return nil
		}
	}

	if len(args) == 0 {
		// Read from stdin
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
	} else {
		// Read from the first file specified
		filePath := args[0]
		input, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", filePath, err)
		}
	}

	// Format the content
	formatted, err := nsfmt.Format(input)
	if err != nil {
		// Formatting errors (like syntax errors) go to stderr
		return fmt.Errorf("failed to format: %w", err)
	}

	// Write the formatted output to stdout
	_, err = os.Stdout.Write(formatted)
	if err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	return nil
}

func printHelp() {
	helpText := `Usage: nsfmt [flags] [path]

nsfmt formats NeuroScript source code.

Flags:
  -h, --help    Show this help message

If no path is provided, nsfmt reads from standard input.
Output is written to standard output.
`
	fmt.Print(helpText)
}
