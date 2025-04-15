package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt" // Needed for ReadAll
	"log"
	"os"
	"path/filepath"
	"strings" // Needed for path splitting

	nspatch "github.com/aprice2704/neuroscript/pkg/nspatch"
)

var (
	dryRun = flag.Bool("dry", false, "Perform a dry run verification without modifying files")
	// Define the -p flag as an integer
	pLevel = flag.Int("p", 0, "Strip <p> leading path components (e.g., -p1)")
)

// stripPrefixComponents removes the first 'level' path components from a path string.
// It handles both '/' and '\' separators.
func stripPrefixComponents(path string, level int) string {
	if level <= 0 {
		return path
	}
	// Normalize separators to '/' for consistent splitting
	normalizedPath := filepath.ToSlash(path)
	components := strings.Split(normalizedPath, "/")

	// If the path starts with '/', the first component might be empty
	if len(components) > 0 && components[0] == "" {
		// Keep the leading '/' if present, adjust level and components
		if len(components) > level+1 {
			return "/" + strings.Join(components[level+1:], "/")
		}
		// Not enough components to strip after removing the empty first one
		return "/"
	}

	if len(components) > level {
		// Join the remaining components
		return strings.Join(components[level:], "/")
	}

	// If level is too high, return the last component (filename) or "." if empty
	if len(components) > 0 {
		return components[len(components)-1]
	}
	return "." // Or consider returning empty string "" depending on desired behavior
}

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		// Update Usage message
		fmt.Fprintf(os.Stderr, "Usage: %s [-dry] [-p <level>] <patchfile.ndpatch.json>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Applies patches defined in a JSON file.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults() // This will now include the -p flag
	}
	flag.Parse() // Parse flags after defining them

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
		// Use the original file path from the patch as the key for grouping
		changesByFile[change.File] = append(changesByFile[change.File], change)
	}

	// Process changes for each file
	var encounteredError bool
	runMode := "Applying"
	if *dryRun {
		runMode = "Dry Run Verification"
	}
	log.Printf("--- Starting Patch (%s Mode) ---", runMode)

	// Iterate through the groups using the original path from the patch file
	for targetFilePathFromPatch, fileChanges := range changesByFile {
		if targetFilePathFromPatch == "" {
			log.Printf("Skipping changes grouped under empty file path")
			continue
		}

		// Determine the effective file path to use for file system operations
		effectiveTargetFilePath := targetFilePathFromPatch
		if *pLevel > 0 {
			effectiveTargetFilePath = stripPrefixComponents(targetFilePathFromPatch, *pLevel)
			// Log the original and effective paths if stripping occurred
			if effectiveTargetFilePath != targetFilePathFromPatch {
				log.Printf("Info: Applying -p%d for target %q, using effective path: %q", *pLevel, targetFilePathFromPatch, effectiveTargetFilePath)
			} else {
				log.Printf("Info: Applying -p%d for target %q, path unchanged: %q", *pLevel, targetFilePathFromPatch, effectiveTargetFilePath)
			}
		}

		// Use targetFilePathFromPatch for logging related to the patch instruction itself
		log.Printf("Processing target instruction for: %s", targetFilePathFromPatch)

		// --- File I/O is now handled here in main, using effectiveTargetFilePath ---
		originalLines, err := readFileLines(effectiveTargetFilePath) // Use potentially stripped path
		originalExists := !errors.Is(err, os.ErrNotExist)
		if err != nil && originalExists {
			// Log error with the path we attempted to read
			log.Printf("ERROR reading target file %q: %v", effectiveTargetFilePath, err)
			encounteredError = true
			continue // Skip to next file
		}
		if !originalExists {
			originalLines = []string{} // Treat non-existent file as empty
			// Log info with the path we checked
			log.Printf("  Info: Target file %q does not exist, starting with empty content.", effectiveTargetFilePath)
		}
		// -----------------------------------------------------------------------------

		if *dryRun {
			// Verification uses originalLines read from effectiveTargetFilePath
			results, firstErr := nspatch.VerifyChanges(originalLines, fileChanges)
			for _, res := range results {
				log.Printf("  - Dry Run: Line %d (Index %d): Op: %s, Result: %s",
					res.LineNumber, res.TargetIndex, res.Operation, res.Status)
			}
			if firstErr != nil {
				// Log error referring to the original patch target path for clarity
				log.Printf("ERROR during dry run verification for %q: %v", targetFilePathFromPatch, firstErr)
				// We don't set encounteredError for dry-run failures, just log them
			} else {
				log.Printf("Dry run verification finished successfully for %s", targetFilePathFromPatch)
			}

		} else {
			// Apply patch for real using the library function
			// Application uses originalLines read from effectiveTargetFilePath
			modifiedLines, applyErr := nspatch.ApplyPatch(originalLines, fileChanges)
			if applyErr != nil {
				// Log error referring to the original patch target path
				log.Printf("ERROR applying patch for %q: %v", targetFilePathFromPatch, applyErr)
				encounteredError = true // Mark failure for real runs
			} else {
				// --- File I/O handled here, using effectiveTargetFilePath ---
				writeErr := writeFileLines(effectiveTargetFilePath, modifiedLines) // Use potentially stripped path
				if writeErr != nil {
					// Log error with the path we attempted to write
					log.Printf("ERROR writing modified file %q: %v", effectiveTargetFilePath, writeErr)
					encounteredError = true
				} else {
					// Log success mentioning the path that was written
					log.Printf("Successfully applied %d changes and wrote %s", len(fileChanges), effectiveTargetFilePath)
				}
				// ----------------------------------------------------------
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
	// Ensure directory exists based on the filePath we intend to write
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %q for file %q: %w", dir, filePath, err)
		}
	}

	file, err := os.Create(filePath) // Overwrites existing file at the target path
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
		// Always add newline
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
