// cmd/block_printer/main.go
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	// Adjust the import path based on your Go module setup
	"github.com/aprice2704/neuroscript/pkg/neurodata/blocks"
)

func main() {
	// --- Argument Handling ---
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <filename>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Prints extracted fenced blocks from the given file.\n")
		os.Exit(1)
	}
	filename := os.Args[1]

	// --- File Reading ---
	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file '%s': %v\n", filename, err)
		os.Exit(1)
	}
	content := string(contentBytes)

	// --- Block Extraction ---
	// Use a logger that discards output for this simple tool,
	// unless debugging is needed. Change io.Discard to os.Stderr to see logs.
	logger := log.New(io.Discard, "[BlockPrinter] ", log.Lshortfile)

	extractedBlocks, err := blocks.ExtractAll(content, logger)
	if err != nil {
		// ExtractAll now logs errors but might return partial results on EOF error
		fmt.Fprintf(os.Stderr, "Error during block extraction from '%s': %v\n", filename, err)
		// Decide if you want to print partial results even if there was an error
		if len(extractedBlocks) == 0 {
			os.Exit(1) // Exit if error and no blocks were extracted
		}
		fmt.Fprintf(os.Stderr, "\n--- WARNING: Proceeding with partially extracted blocks ---\n\n")
	}

	// --- Formatting and Printing ---
	formattedOutput := blocks.FormatBlocks(extractedBlocks)
	fmt.Println(formattedOutput)
}
