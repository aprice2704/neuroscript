package nspatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// --- Exported Errors (remain the same) ---
var (
	ErrVerificationFailed = errors.New("verification failed: original line content mismatch")
	ErrOutOfBounds        = errors.New("target index out of bounds")
	ErrInvalidOperation   = errors.New("invalid patch operation")
	ErrMissingField       = errors.New("missing required field in patch object")
	ErrInvalidPatchFile   = errors.New("invalid patch file")
	ErrApplyFailed        = errors.New("patch application failed")
	ErrInternal           = errors.New("internal patch error")
)

// --- Structs ---
// Using the base version provided by user with original tags
type PatchChange struct {
	File      string  `json:"file"`
	Line      int     `json:"line"`
	Operation string  `json:"op"`
	OldLine   *string `json:"old"`
	NewLine   *string `json:"new"`
}

type VerificationResult struct {
	ChangeIndex int
	LineNumber  int
	TargetIndex int // 0-based index calculated for slice access, reflects state *before* current op
	Operation   string
	Status      string
	IsError     bool
	Err         error
}

// --- Core Logic Functions ---

// VerifyChanges performs the verification pass against the provided lines.
// *** INCORPORATING PREVIOUS LOGIC FIX + DEBUGGING ***
func VerifyChanges(originalLines []string, changes []PatchChange) ([]VerificationResult, error) {
	fmt.Println("--- Starting VerifyChanges ---") // DEBUG
	results := make([]VerificationResult, 0, len(changes))
	var firstError error = nil // Initialize error to nil
	originalContentLen := len(originalLines)
	currentContentLen := originalContentLen                                                                                            // Track conceptual length for bounds checking
	verificationOffset := 0                                                                                                            // Tracks index shifts due to inserts/deletes
	fmt.Printf("[DEBUG VerifyChanges] Initial originalContentLen: %d, currentContentLen: %d\n", originalContentLen, currentContentLen) // DEBUG

	for i, change := range changes {
		fmt.Printf("\n[DEBUG VerifyChanges] Processing Change #%d: Op=%s, Line=%d, Offset=%d\n", i, change.Operation, change.Line, verificationOffset) // DEBUG
		targetIndex := change.Line - 1 + verificationOffset
		status := "Not Checked"
		isOutOfBounds := false
		isOperationError := false
		isVerificationError := false
		var currentError error = nil // Error for THIS change

		fmt.Printf("[DEBUG VerifyChanges] Calculated TargetIndex: %d (currentContentLen: %d)\n", targetIndex, currentContentLen) // DEBUG

		res := VerificationResult{
			ChangeIndex: i,
			LineNumber:  change.Line,
			TargetIndex: targetIndex,
			Operation:   change.Operation,
		}

		// 1. Validate Operation type first
		switch change.Operation {
		case "replace", "delete", "insert":
			fmt.Printf("[DEBUG VerifyChanges] Operation '%s' is valid.\n", change.Operation) // DEBUG
		default:
			isOperationError = true
			currentError = fmt.Errorf("%w: unknown operation '%s'", ErrInvalidOperation, change.Operation)
			status = fmt.Sprintf("Error: %v", currentError)
			fmt.Printf("[DEBUG VerifyChanges] Invalid Operation Error: %v\n", currentError) // DEBUG
		}

		// 2. Bounds checks (if operation is valid)
		if !isOperationError {
			if targetIndex < 0 {
				isOutOfBounds = true
				fmt.Printf("[DEBUG VerifyChanges] Bounds Check: FAILED (targetIndex %d < 0)\n", targetIndex) // DEBUG
			} else {
				switch change.Operation {
				case "replace", "delete":
					if targetIndex >= currentContentLen {
						isOutOfBounds = true
						fmt.Printf("[DEBUG VerifyChanges] Bounds Check (%s): FAILED (targetIndex %d >= currentContentLen %d)\n", change.Operation, targetIndex, currentContentLen) // DEBUG
					} else {
						fmt.Printf("[DEBUG VerifyChanges] Bounds Check (%s): OK (targetIndex %d < currentContentLen %d)\n", change.Operation, targetIndex, currentContentLen) // DEBUG
					}
				case "insert":
					if targetIndex > currentContentLen {
						isOutOfBounds = true
						fmt.Printf("[DEBUG VerifyChanges] Bounds Check (%s): FAILED (targetIndex %d > currentContentLen %d)\n", change.Operation, targetIndex, currentContentLen) // DEBUG
					} else {
						fmt.Printf("[DEBUG VerifyChanges] Bounds Check (%s): OK (targetIndex %d <= currentContentLen %d)\n", change.Operation, targetIndex, currentContentLen) // DEBUG
					}
				}
			}

			if isOutOfBounds {
				// isOperationError = true // Don't set this here, keep bounds error separate
				currentError = fmt.Errorf("%w: target index %d for %s out of bounds (conceptual lines: %d, line: %d, offset: %d)",
					ErrOutOfBounds, targetIndex, change.Operation, currentContentLen, change.Line, verificationOffset)
				status = fmt.Sprintf("Error: %v", currentError)
				fmt.Printf("[DEBUG VerifyChanges] Bounds Error Set: %v\n", currentError) // DEBUG
			}
		}

		// 3. Verification Check (only if operation and bounds are okay)
		if !isOperationError && !isOutOfBounds {
			fmt.Printf("[DEBUG VerifyChanges] Proceeding to Verification Check (change.OldLine provided: %t)\n", change.OldLine != nil) // DEBUG
			if change.OldLine != nil {                                                                                                  // Verification requested
				originalLineIndex := change.Line - 1 // Use original line number for indexing originalLines

				fmt.Printf("[DEBUG VerifyChanges] Verification: originalLineIndex = %d (originalContentLen: %d)\n", originalLineIndex, originalContentLen) // DEBUG
				// Check if original index is valid for the *original* slice
				if originalLineIndex >= 0 && originalLineIndex < originalContentLen {
					originalFromFileRaw := originalLines[originalLineIndex]
					originalFromFileTrimmed := strings.TrimSpace(strings.TrimSuffix(originalFromFileRaw, "\n"))
					oldLineFromPatch := strings.TrimSpace(*change.OldLine)

					fmt.Printf("[DEBUG VerifyChanges] Comparing:\n  Original Line %d (Trimmed): %q\n  Patch 'old'     (Trimmed): %q\n", originalLineIndex+1, originalFromFileTrimmed, oldLineFromPatch) // DEBUG

					if originalFromFileTrimmed == oldLineFromPatch {
						status = "Matched"
						fmt.Println("[DEBUG VerifyChanges] Comparison Result: Matched") // DEBUG
					} else {
						status = fmt.Sprintf("MISMATCHED (Expected: %q, Found: %q)", *change.OldLine, originalFromFileRaw)
						isVerificationError = true // Mark verification specifically failed
						currentError = fmt.Errorf("%w: line %d: expected %q, found %q", ErrVerificationFailed, change.Line, *change.OldLine, originalFromFileRaw)
						fmt.Printf("[DEBUG VerifyChanges] Comparison Result: MISMATCHED! Error: %v\n", currentError) // DEBUG
					}
				} else {
					// OldLine provided, but original index itself is invalid
					status = fmt.Sprintf("Error: Verification Failed (Original line number %d outside original bounds [1-%d])", change.Line, originalContentLen)
					isVerificationError = true // Mark verification specifically failed
					currentError = fmt.Errorf("%w: line %d verification requested, but original file only has %d lines", ErrVerificationFailed, change.Line, originalContentLen)
					fmt.Printf("[DEBUG VerifyChanges] Verification Error (Original Index Out of Bounds): %v\n", currentError) // DEBUG
				}
			} else if change.Operation == "replace" || change.Operation == "delete" {
				status = "Not Verified (No Ref)"
				fmt.Println("[DEBUG VerifyChanges] Status: Not Verified (No Ref)") // DEBUG
			} else { // insert
				status = "OK (No Verification Needed)"
				fmt.Println("[DEBUG VerifyChanges] Status: OK (No Verification Needed for Insert)") // DEBUG
			}
		} else {
			fmt.Printf("[DEBUG VerifyChanges] Skipping Verification Check because isOperationError=%t or isOutOfBounds=%t\n", isOperationError, isOutOfBounds) // DEBUG
		}

		// --- Finalize Result ---
		res.IsError = isOperationError || isOutOfBounds || isVerificationError // Combine all possible error flags
		res.Err = currentError
		// Ensure status reflects the final state if it wasn't set by a specific condition
		if status == "Not Checked" {
			if res.IsError {
				status = fmt.Sprintf("Error: %v", currentError)
			} else {
				status = "OK (No Verification Needed)" // Should not happen if logic above is sound
			}
		}
		res.Status = status
		results = append(results, res)
		fmt.Printf("[DEBUG VerifyChanges] Final Result for Change #%d: Status=%s, IsError=%t, Err=%v\n", i, res.Status, res.IsError, res.Err) // DEBUG

		// Store the first *actual* error encountered for the function's return value
		// ** CRITICAL: Check res.IsError which combines all error sources **
		if res.IsError && firstError == nil {
			firstError = fmt.Errorf("change #%d (%s line %d): %w", i+1, change.Operation, change.Line, currentError)
			fmt.Printf("[DEBUG VerifyChanges] Storing First Error: %v\n", firstError) // DEBUG
		}

		// --- Update conceptual length and offset for the *next* iteration ---
		// ** CRITICAL: Only adjust if the CURRENT operation had no error of any kind **
		if !res.IsError { // Only adjust if the current operation was valid and verification passed (if applicable)
			switch change.Operation {
			case "insert":
				verificationOffset++
				currentContentLen++
			case "delete":
				verificationOffset--
				currentContentLen--
				// case "replace": offset and length remain the same
			}
		} else {
			fmt.Printf("[DEBUG VerifyChanges] Offset/Length NOT adjusted due to error in current change.\n") // DEBUG
		}
		fmt.Printf("[DEBUG VerifyChanges] End of Loop #%d: Next verificationOffset=%d, Next currentContentLen=%d\n", i, verificationOffset, currentContentLen) // DEBUG

	} // End verification loop

	fmt.Printf("--- Finished VerifyChanges --- Returning Error: %v\n", firstError) // DEBUG: Check error just before return
	return results, firstError                                                     // Return the first error encountered
}

// ApplyPatch performs the two-pass (verify, then apply) patch operation.
// (ApplyPatch function remains unchanged from the user-provided base version)
func ApplyPatch(originalLines []string, changes []PatchChange) ([]string, error) {
	fmt.Println("--- Starting ApplyPatch ---") // DEBUG
	// --- Pass 1: Verify ---
	_, firstVerificationError := VerifyChanges(originalLines, changes)
	if firstVerificationError != nil {
		fmt.Printf("[DEBUG ApplyPatch] Verification failed: %v. Returning error.\n", firstVerificationError) // DEBUG
		return nil, firstVerificationError                                                                   // Return the specific error from VerifyChanges
	}
	fmt.Println("[DEBUG ApplyPatch] Verification successful.") // DEBUG

	// --- Pass 2: Application ---
	fmt.Println("[DEBUG ApplyPatch] Starting application phase.") // DEBUG
	modifiedLines := make([]string, len(originalLines))
	copy(modifiedLines, originalLines)
	applyOffset := 0

	for i, change := range changes {
		targetIndex := change.Line - 1 + applyOffset
		fmt.Printf("[DEBUG ApplyPatch] Applying Change #%d: Op=%s, Line=%d, TargetIndex=%d\n", i, change.Operation, change.Line, targetIndex) //DEBUG

		// Re-check bounds against the *current* state of modifiedLines before applying
		currentLen := len(modifiedLines)
		isValid := true
		switch change.Operation {
		case "replace", "delete":
			if targetIndex < 0 || targetIndex >= currentLen {
				isValid = false
			}
		case "insert":
			if targetIndex < 0 || targetIndex > currentLen { // Allow insert at end
				isValid = false
			}
		default:
			fmt.Printf("[DEBUG ApplyPatch] ERROR: Unknown operation '%s' during apply.\n", change.Operation) // DEBUG
			return nil, fmt.Errorf("%w: change #%d: unknown operation '%s' encountered during apply phase", ErrInternal, i+1, change.Operation)

		}

		if !isValid {
			fmt.Printf("[DEBUG ApplyPatch] ERROR: Index %d became invalid during apply (Op=%s, Line=%d, Offset=%d, currentLen=%d).\n", targetIndex, change.Operation, change.Line, applyOffset, currentLen) //DEBUG
			return nil, fmt.Errorf("%w: change #%d (%s line %d): index %d became invalid during apply (lines: %d, offset: %d)", ErrInternal, i+1, change.Operation, change.Line, targetIndex, currentLen, applyOffset)
		}

		switch change.Operation {
		case "replace":
			if change.NewLine == nil {
				return nil, fmt.Errorf("%w: change #%d (%s line %d): missing 'new'", ErrInternal, i+1, change.Operation, change.Line)
			}
			fmt.Printf("[DEBUG ApplyPatch] Replacing line %d with %q\n", targetIndex, *change.NewLine) // DEBUG
			modifiedLines[targetIndex] = *change.NewLine
		case "insert":
			if change.NewLine == nil {
				return nil, fmt.Errorf("%w: change #%d (%s line %d): missing 'new'", ErrInternal, i+1, change.Operation, change.Line)
			}
			newLine := *change.NewLine
			fmt.Printf("[DEBUG ApplyPatch] Inserting %q at index %d\n", newLine, targetIndex) // DEBUG
			// Efficient insertion
			modifiedLines = append(modifiedLines, "")                        // Grow slice cap/len if needed
			copy(modifiedLines[targetIndex+1:], modifiedLines[targetIndex:]) // Shift elements right
			modifiedLines[targetIndex] = newLine                             // Insert new element

			applyOffset++
		case "delete":
			fmt.Printf("[DEBUG ApplyPatch] Deleting line %d (%q)\n", targetIndex, modifiedLines[targetIndex]) // DEBUG
			modifiedLines = append(modifiedLines[:targetIndex], modifiedLines[targetIndex+1:]...)
			applyOffset--
		}
		fmt.Printf("[DEBUG ApplyPatch] After Change #%d: Offset=%d, Lines=%d\n", i, applyOffset, len(modifiedLines)) //DEBUG
	}

	fmt.Println("--- Finished ApplyPatch Successfully ---") // DEBUG
	return modifiedLines, nil                               // Success
}

// --- File Loading Helper ---
// (LoadPatchFile function remains unchanged from the user-provided base version)
func LoadPatchFile(filePath string) ([]PatchChange, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: opening file %q: %w", ErrInvalidPatchFile, filePath, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("%w: reading file %q: %w", ErrInvalidPatchFile, filePath, err)
	}

	var changes []PatchChange
	err = json.Unmarshal(data, &changes)
	if err != nil {
		var syntaxError *json.SyntaxError
		if errors.As(err, &syntaxError) {
			line, char := calculateErrorPosition(data, syntaxError.Offset)
			return nil, fmt.Errorf("%w: unmarshaling json in %q at line %d, char %d: %w", ErrInvalidPatchFile, filePath, line, char, err)

		}
		return nil, fmt.Errorf("%w: unmarshaling json from %q: %w", ErrInvalidPatchFile, filePath, err)
	}

	for i, change := range changes {
		if change.File == "" { /* Allow empty file field */
		}
		if change.Line <= 0 {
			// Add the file path context to the error message
			return nil, fmt.Errorf("%w: change #%d in file %q (from patch %q) has invalid 'line': %d (must be >= 1)", ErrOutOfBounds, i+1, change.File, filePath, change.Line)
		}
		switch change.Operation {
		case "replace", "insert":
			if change.NewLine == nil {
				// Add the file path context to the error message
				return nil, fmt.Errorf("%w: change #%d (%s line %d) in file %q (from patch %q) missing 'new'", ErrMissingField, i+1, change.Operation, change.Line, change.File, filePath)
			}
		case "delete": /* Ok */
		default:
			// Add the file path context to the error message
			return nil, fmt.Errorf("%w: change #%d in file %q (from patch %q) has unknown operation: %q", ErrInvalidOperation, i+1, change.File, filePath, change.Operation)
		}
	}
	return changes, nil
}

// calculateErrorPosition function (remains unchanged from the user-provided base version)
func calculateErrorPosition(data []byte, offset int64) (line, char int) {
	line = 1
	char = 1
	currentOffset := int64(0)
	for _, b := range data {
		if currentOffset >= offset {
			break
		}
		if b == '\n' {
			line++
			char = 1
		} else {
			// Assuming UTF-8, multi-byte characters advance char position by 1 visually
			char++
		}
		currentOffset++
	}
	if offset > 0 && offset == currentOffset && len(data) > 0 && data[offset-1] == '\n' {
		line++
		char = 1
	} else if offset > 0 && offset == currentOffset {
	}

	return line, char
}
