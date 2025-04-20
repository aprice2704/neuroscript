// filename: pkg/neurogo/patch_handler.go
package neurogo

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"    // For core types like Interpreter, SecurityLayer & SecureFilePath func
	"github.com/aprice2704/neuroscript/pkg/nspatch" // For patch types and application logic
)

// ErrPatchAbortedByUser indicates the user cancelled the patch operation.
var ErrPatchAbortedByUser = errors.New("patch application aborted by user confirmation")

// --- Helper function to read file lines ---
// Reads file content, returns []string. Returns empty slice if file doesn't exist.
func readLines(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		// If file doesn't exist, return empty slice and nil error
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("reading file %q: %w", filePath, err)
	}
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, nil
}

// --- Helper function to write file lines ---
// Writes []string to file, ensuring parent directory exists. Uses '\n' line endings.
func writeLines(filePath string, lines []string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %q: %w", dir, err)
	}
	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("writing file %q: %w", filePath, err)
	}
	return nil
}

// handleReceivedPatch processes a JSON patch string, prompts for confirmation, and applies it.
// It requires the interpreter, security layer, and sandbox root for context and validation.
func handleReceivedPatch(patchJSON string, interp *core.Interpreter, securityLayer *core.SecurityLayer, sandboxRoot string, infoLog, errorLog *log.Logger) error {
	infoLog.Printf("[PATCH] Received patch request.")

	// 1. Unmarshal JSON
	var changes []nspatch.PatchChange
	err := json.Unmarshal([]byte(patchJSON), &changes)
	if err != nil {
		errorLog.Printf("[PATCH] Error unmarshaling patch JSON: %v", err)
		return fmt.Errorf("invalid patch format: %w", err)
	}

	if len(changes) == 0 {
		infoLog.Println("[PATCH] Received empty patch list. No action taken.")
		return nil
	}

	// 2. Group changes by file
	changesByFile := make(map[string][]nspatch.PatchChange)
	fileOrder := []string{}
	for _, change := range changes {
		if filepath.IsAbs(change.File) || strings.Contains(change.File, "..") {
			errorLog.Printf("[PATCH] Security violation: Patch contains non-relative or invalid path element '..': %q", change.File)
			return fmt.Errorf("invalid path in patch for file %q: %w", change.File, core.ErrPathViolation)
		}
		if _, exists := changesByFile[change.File]; !exists {
			changesByFile[change.File] = []nspatch.PatchChange{}
			fileOrder = append(fileOrder, change.File)
		}
		changesByFile[change.File] = append(changesByFile[change.File], change)
	}

	// 3. Confirmation Prompt
	fmt.Printf("\n--- Proposed Patch ---\n")
	fmt.Printf("This patch will modify %d file(s):\n", len(fileOrder))
	totalChanges := 0
	for _, fileRelPath := range fileOrder {
		numFileChanges := len(changesByFile[fileRelPath])
		fmt.Printf("  - %s (%d changes)\n", fileRelPath, numFileChanges)
		totalChanges += numFileChanges
	}
	fmt.Printf("Total changes: %d\n", totalChanges)
	fmt.Print("Apply this patch? (y/n): ")

	reader := bufio.NewReader(os.Stdin)
	confirmation, err := reader.ReadString('\n')
	if err != nil {
		errorLog.Printf("[PATCH] Error reading confirmation: %v", err)
		return fmt.Errorf("failed to read confirmation: %w", err)
	}
	confirmation = strings.ToLower(strings.TrimSpace(confirmation))

	if confirmation != "y" && confirmation != "yes" {
		infoLog.Println("[PATCH] User aborted patch application.")
		return ErrPatchAbortedByUser
	}
	infoLog.Println("[PATCH] User confirmed patch application.")

	// 4. Process each file
	for _, fileRelPath := range fileOrder {
		fileSpecificChanges := changesByFile[fileRelPath]
		infoLog.Printf("[PATCH] Processing %d changes for file: %s", len(fileSpecificChanges), fileRelPath)

		// a. Security Check (Get validated absolute path using the correct function call)
		// *** CORRECTED CALL: Use core.SecureFilePath ***
		absPath, err := core.SecureFilePath(fileRelPath, sandboxRoot)
		if err != nil {
			errorLog.Printf("[PATCH] Security violation for file %q: %v", fileRelPath, err)
			// Ensure the error is wrapped or identifiable as a security path error
			return fmt.Errorf("patch security error for %q: %w", fileRelPath, err)
		}
		infoLog.Printf("[PATCH] Secured absolute path: %s", absPath)

		// b. Read Original File Content
		originalLines, err := readLines(absPath)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				errorLog.Printf("[PATCH] Error reading original file %q (%q): %v", fileRelPath, absPath, err)
				return fmt.Errorf("failed to read original file %q: %w", fileRelPath, err)
			}
			infoLog.Printf("[PATCH] Original file %q does not exist, starting with empty content.", fileRelPath)
			originalLines = []string{}
		} else {
			infoLog.Printf("[PATCH] Read %d lines from %q", len(originalLines), fileRelPath)
		}

		// c. Apply Patch using nspatch library
		modifiedLines, err := nspatch.ApplyPatch(originalLines, fileSpecificChanges)
		if err != nil {
			errorLog.Printf("[PATCH] Error applying patch logic to %q: %v", fileRelPath, err)
			return fmt.Errorf("patch application failed for file %q: %w", fileRelPath, err)
		}
		infoLog.Printf("[PATCH] Patch logic applied successfully for %q. Resulting lines: %d", fileRelPath, len(modifiedLines))

		// d. Write Modified File Content
		err = writeLines(absPath, modifiedLines)
		if err != nil {
			errorLog.Printf("[PATCH] Error writing modified file %q (%q): %v", fileRelPath, absPath, err)
			return fmt.Errorf("failed to write modified file %q: %w", fileRelPath, err)
		}
		infoLog.Printf("[PATCH] Successfully wrote modified content to %q", fileRelPath)
	} // End loop through files

	infoLog.Println("[PATCH] All files processed successfully.")
	return nil // Success
}
