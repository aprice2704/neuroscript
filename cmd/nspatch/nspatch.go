package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt" // Needed for ReadAll
	"log"
	"os"
	"path/filepath"

	nspatch "github.com/aprice2704/neuroscript/pkg/nspatch"
)

var (
	dryRun = flag.Bool("dry", false, "Perform a dry run verification without modifying files")
)

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-dry] <patchfile.ndpatch.json>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Applies patches defined in a JSON file.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	patchFilePath := flag.Arg(0)

	// Load the patch definition file using the library function
	changes, err := nspatch.LoadPatchFile(patchFilePath)
	if err != nil {
		log.Fatalf("Error loading patch file %q: %v", patchFilePath, err)
	}

	// Group changes by target file (using the File field from PatchChange)
	changesByFile := make(map[string][]nspatch.PatchChange)
	for _, change := range changes {
		if change.File == "" {
			log.Printf("Warning: Change object %+v has empty 'file' field, grouping under empty key", change)
		}
		changesByFile[change.File] = append(changesByFile[change.File], change)
	}

	// Process changes for each file
	var encounteredError bool
	runMode := "Applying"
	if *dryRun {
		runMode = "Dry Run Verification"
	}
	log.Printf("--- Starting Patch (%s Mode) ---", runMode)

	for targetFilePath, fileChanges := range changesByFile {
		if targetFilePath == "" {
			log.Printf("Skipping changes grouped under empty file path")
			continue
		}
		log.Printf("Processing target: %s", targetFilePath)

		// --- File I/O is now handled here in main ---
		originalLines, err := readFileLines(targetFilePath) // Use local helper
		originalExists := !errors.Is(err, os.ErrNotExist)
		if err != nil && originalExists {
			log.Printf("ERROR reading target file %q: %v", targetFilePath, err)
			encounteredError = true
			continue // Skip to next file
		}
		if !originalExists {
			originalLines = []string{} // Treat non-existent file as empty
			log.Printf("  Info: Target file %q does not exist, starting with empty content.", targetFilePath)
		}
		// --------------------------------------------

		if *dryRun {
			results, firstErr := nspatch.VerifyChanges(originalLines, fileChanges)
			for _, res := range results {
				log.Printf("  - Dry Run: Line %d (Index %d): Op: %s, Result: %s",
					res.LineNumber, res.TargetIndex, res.Operation, res.Status)
			}
			if firstErr != nil {
				log.Printf("ERROR during dry run verification for %q: %v", targetFilePath, firstErr)
				// We don't set encounteredError for dry-run failures, just log them
			} else {
				log.Printf("Dry run verification finished successfully for %s", targetFilePath)
			}

		} else {
			// Apply patch for real using the library function
			modifiedLines, applyErr := nspatch.ApplyPatch(originalLines, fileChanges)
			if applyErr != nil {
				log.Printf("ERROR applying patch to %q: %v", targetFilePath, applyErr)
				encounteredError = true // Mark failure for real runs
			} else {
				// --- File I/O handled here ---
				writeErr := writeFileLines(targetFilePath, modifiedLines) // Use local helper
				if writeErr != nil {
					log.Printf("ERROR writing modified file %q: %v", targetFilePath, writeErr)
					encounteredError = true
				} else {
					log.Printf("Successfully applied %d changes and wrote %s", len(fileChanges), targetFilePath)
				}
				// --------------------------
			}
		}
	} // End loop through files

	if encounteredError {
		log.Fatal("--- Finished with errors. ---")
	}
	log.Println("--- Finished successfully. ---")
}

// --- Internal File I/O Helpers (moved from library) ---

func readFileLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err // Propagate os.ErrNotExist etc.
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	// Increase buffer size for potentially long lines
	const maxCapacity = 1024 * 1024 // 1MB buffer
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file %q: %w", filePath, err)
	}
	return lines, nil
}

func writeFileLines(filePath string, lines []string) error {
	// Ensure directory exists (optional, but good practice)
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %q for file %q: %w", dir, filePath, err)
		}
	}

	file, err := os.Create(filePath) // Overwrites existing file
	if err != nil {
		return fmt.Errorf("creating file %q: %w", filePath, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for i, line := range lines {
		_, err := writer.WriteString(line)
		if err != nil {
			return fmt.Errorf("writing line %d to %q: %w", i+1, filePath, err)
		}
		// Only add newline if it's not the very last line potentially
		// Or always add newline - standard text files usually end with one. Let's always add.
		_, err = writer.WriteString("\n")
		if err != nil {
			return fmt.Errorf("writing newline after line %d to %q: %w", i+1, filePath, err)
		}

	}
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("flushing writer for %q: %w", filePath, err)
	}
	return nil
}
