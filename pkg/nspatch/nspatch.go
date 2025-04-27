package nspatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog" // Correct import for slog
	"os"
	"strings"
	// Assuming adapters package is correctly located relative to this file
	// Adjust if your adapters package is elsewhere (e.g., internal/adapters)
	// Example path
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
	OldLine   *string `json:"old"` // Pointer to distinguish between unset and empty string
	NewLine   *string `json:"new"` // Pointer to distinguish between unset and empty string
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

// Use a package-level logger, initialized simply here.
// Consider proper initialization if more complex setup is needed.
var logger = slog.New(slog.NewTextHandler(io.Discard, nil)) // Default to discard, can be set externally

// SetLogger allows setting the logger for the package. Useful for testing or central config.
func SetLogger(l *slog.Logger) {
	if l != nil {
		logger = l
	}
}

// --- Core Logic Functions ---

// VerifyChanges performs the verification pass against the provided lines.
func VerifyChanges(originalLines []string, changes []PatchChange) ([]VerificationResult, error) {
	// Using logger.Debug instead of fmt.Println
	logger.Debug("--- Starting VerifyChanges ---")
	results := make([]VerificationResult, 0, len(changes))
	var firstError error = nil // Initialize error to nil
	originalContentLen := len(originalLines)
	currentContentLen := originalContentLen // Track conceptual length for bounds checking
	verificationOffset := 0                 // Tracks index shifts due to inserts/deletes
	logger.Debug("VerifyChanges initial state", "originalLen", originalContentLen, "currentLen", currentContentLen, "offset", verificationOffset)

	for i, change := range changes {
		logger.Debug("Processing change", "index", i, "op", change.Operation, "line", change.Line, "offset", verificationOffset)
		targetIndex := change.Line - 1 + verificationOffset
		status := "Not Checked"
		isOutOfBounds := false
		isOperationError := false
		isVerificationError := false
		var currentError error = nil // Error for THIS change

		logger.Debug("Calculated target index", "targetIndex", targetIndex, "currentLen", currentContentLen)

		res := VerificationResult{
			ChangeIndex: i,
			LineNumber:  change.Line,
			TargetIndex: targetIndex,
			Operation:   change.Operation,
		}

		// 1. Validate Operation type first
		switch change.Operation {
		case "replace", "delete", "insert":
			logger.Debug("Operation valid", "op", change.Operation)
		default:
			isOperationError = true
			currentError = fmt.Errorf("%w: unknown operation '%s'", ErrInvalidOperation, change.Operation)
			status = fmt.Sprintf("Error: %v", currentError)
			logger.Debug("Invalid operation error", "error", currentError)
		}

		// 2. Bounds checks (if operation is valid)
		if !isOperationError {
			if targetIndex < 0 {
				isOutOfBounds = true
				logger.Debug("Bounds check failed: target index < 0", "targetIndex", targetIndex)
			} else {
				switch change.Operation {
				case "replace", "delete":
					if targetIndex >= currentContentLen {
						isOutOfBounds = true
						logger.Debug("Bounds check failed (replace/delete)", "targetIndex", targetIndex, "currentLen", currentContentLen)
					} else {
						logger.Debug("Bounds check OK (replace/delete)", "targetIndex", targetIndex, "currentLen", currentContentLen)
					}
				case "insert":
					// Allow inserting at the very end (index == currentContentLen)
					if targetIndex > currentContentLen {
						isOutOfBounds = true
						logger.Debug("Bounds check failed (insert)", "targetIndex", targetIndex, "currentLen", currentContentLen)
					} else {
						logger.Debug("Bounds check OK (insert)", "targetIndex", targetIndex, "currentLen", currentContentLen)
					}
				}
			}

			if isOutOfBounds {
				currentError = fmt.Errorf("%w: target index %d for %s out of bounds (conceptual lines: %d, line: %d, offset: %d)",
					ErrOutOfBounds, targetIndex, change.Operation, currentContentLen, change.Line, verificationOffset)
				status = fmt.Sprintf("Error: %v", currentError)
				logger.Debug("Bounds error set", "error", currentError)
			}
		}

		// 3. Verification Check (only if operation and bounds are okay)
		if !isOperationError && !isOutOfBounds {
			oldLineProvided := change.OldLine != nil
			logger.Debug("Proceeding to verification check", "oldLineProvided", oldLineProvided)
			if oldLineProvided { // Verification requested
				originalLineIndex := change.Line - 1 // Use original line number for indexing originalLines

				logger.Debug("Verification details", "originalLineIndex", originalLineIndex, "originalLen", originalContentLen)
				// Check if original index is valid for the *original* slice
				if originalLineIndex >= 0 && originalLineIndex < originalContentLen {
					originalFromFileRaw := originalLines[originalLineIndex]
					// Trim space only for comparison - error message shows raw value
					originalFromFileTrimmed := strings.TrimSpace(originalFromFileRaw)
					oldLineFromPatch := strings.TrimSpace(*change.OldLine)

					logger.Debug("Comparing lines", "originalTrimmed", originalFromFileTrimmed, "patchOldTrimmed", oldLineFromPatch)

					if originalFromFileTrimmed == oldLineFromPatch {
						status = "Matched"
						logger.Debug("Comparison result: Matched")
					} else {
						// Show non-trimmed values in error for clarity
						status = fmt.Sprintf("MISMATCHED (Expected: %q, Found: %q)", *change.OldLine, originalFromFileRaw)
						isVerificationError = true // Mark verification specifically failed
						currentError = fmt.Errorf("%w: line %d: expected %q, found %q", ErrVerificationFailed, change.Line, *change.OldLine, originalFromFileRaw)
						logger.Debug("Comparison result: Mismatched", "error", currentError)
					}
				} else {
					// OldLine provided, but original index itself is invalid
					status = fmt.Sprintf("Error: Verification Failed (Original line number %d outside original bounds [1-%d])", change.Line, originalContentLen)
					isVerificationError = true // Mark verification specifically failed
					currentError = fmt.Errorf("%w: line %d verification requested, but original file only has %d lines", ErrVerificationFailed, change.Line, originalContentLen)
					logger.Debug("Verification error: Original index out of bounds", "error", currentError)
				}
			} else if change.Operation == "replace" || change.Operation == "delete" {
				status = "Not Verified (No Ref)"
				logger.Debug("Status: Not Verified (No Ref)")
			} else { // insert
				status = "OK (No Verification Needed)"
				logger.Debug("Status: OK (No Verification Needed for Insert)")
			}
		} else {
			logger.Debug("Skipping verification check", "isOperationError", isOperationError, "isOutOfBounds", isOutOfBounds)
		}

		// --- Finalize Result ---
		res.IsError = isOperationError || isOutOfBounds || isVerificationError // Combine all possible error flags
		res.Err = currentError
		// Ensure status reflects the final state if it wasn't set by a specific condition
		if status == "Not Checked" {
			if res.IsError {
				status = fmt.Sprintf("Error: %v", currentError)
			} else {
				// This path should ideally not be hit if logic above is exhaustive
				status = "OK (Internal Status Error)"
				logger.Warn("Internal status error: Reached 'Not Checked' state without error flag set.")
			}
		}
		res.Status = status
		results = append(results, res)
		logger.Debug("Final result for change", "index", i, "status", res.Status, "isError", res.IsError, "error", res.Err)

		// Store the first *actual* error encountered for the function's return value
		if res.IsError && firstError == nil {
			firstError = fmt.Errorf("change #%d (%s line %d): %w", i+1, change.Operation, change.Line, currentError)
			logger.Debug("Storing first error", "error", firstError)
		}

		// --- Update conceptual length and offset for the *next* iteration ---
		if !res.IsError { // Only adjust if the current operation was valid and verification passed
			switch change.Operation {
			case "insert":
				verificationOffset++
				currentContentLen++
			case "delete":
				verificationOffset--
				currentContentLen--
				// case "replace": offset and length remain the same
			}
			logger.Debug("Offset/Length updated", "nextOffset", verificationOffset, "nextCurrentLen", currentContentLen)
		} else {
			logger.Debug("Offset/Length NOT updated due to error in current change.")
		}

	} // End verification loop

	logger.Debug("--- Finished VerifyChanges ---", "returningError", firstError)
	return results, firstError // Return the first error encountered
}

// ApplyPatch performs the two-pass (verify, then apply) patch operation.
func ApplyPatch(originalLines []string, changes []PatchChange) ([]string, error) {
	logger.Debug("--- Starting ApplyPatch ---")
	// --- Pass 1: Verify ---
	_, firstVerificationError := VerifyChanges(originalLines, changes)
	if firstVerificationError != nil {
		logger.Debug("Verification failed, returning error.", "error", firstVerificationError)
		return nil, firstVerificationError // Return the specific error from VerifyChanges
	}
	logger.Debug("Verification successful.")

	// --- Pass 2: Application ---
	logger.Debug("Starting application phase.")
	modifiedLines := make([]string, len(originalLines))
	copy(modifiedLines, originalLines)
	applyOffset := 0

	for i, change := range changes {
		targetIndex := change.Line - 1 + applyOffset
		logger.Debug("Applying change", "index", i, "op", change.Operation, "line", change.Line, "targetIndex", targetIndex)

		// Re-check bounds against the *current* state of modifiedLines before applying
		currentLen := len(modifiedLines)
		isValid := true
		switch change.Operation {
		case "replace", "delete":
			if targetIndex < 0 || targetIndex >= currentLen {
				isValid = false
			}
		case "insert":
			if targetIndex < 0 || targetIndex > currentLen { // Allow insert at end (index == currentLen)
				isValid = false
			}
		default:
			logger.Error("Unknown operation during apply phase", "index", i, "op", change.Operation)
			return nil, fmt.Errorf("%w: change #%d: unknown operation '%s' encountered during apply phase", ErrInternal, i+1, change.Operation)
		}

		if !isValid {
			logger.Error("Index became invalid during apply phase", "index", i, "op", change.Operation, "line", change.Line, "targetIndex", targetIndex, "currentLen", currentLen, "offset", applyOffset)
			return nil, fmt.Errorf("%w: change #%d (%s line %d): index %d became invalid during apply (lines: %d, offset: %d)", ErrInternal, i+1, change.Operation, change.Line, targetIndex, currentLen, applyOffset)
		}

		// Apply the change
		switch change.Operation {
		case "replace":
			if change.NewLine == nil {
				logger.Error("Missing 'new' field for replace operation", "index", i, "line", change.Line)
				return nil, fmt.Errorf("%w: change #%d (replace line %d): missing 'new' field during apply phase", ErrInternal, i+1, change.Line)
			}
			logger.Debug("Replacing line", "targetIndex", targetIndex, "oldValue", modifiedLines[targetIndex], "newValue", *change.NewLine)
			modifiedLines[targetIndex] = *change.NewLine
		case "insert":
			if change.NewLine == nil {
				logger.Error("Missing 'new' field for insert operation", "index", i, "line", change.Line)
				return nil, fmt.Errorf("%w: change #%d (insert line %d): missing 'new' field during apply phase", ErrInternal, i+1, change.Line)
			}
			newLine := *change.NewLine
			logger.Debug("Inserting line", "targetIndex", targetIndex, "value", newLine)

			// --- Replace the existing insert logic lines with this line: ---
			modifiedLines = append(modifiedLines[:targetIndex], append([]string{newLine}, modifiedLines[targetIndex:]...)...)
			// --- End of replacement ---

			applyOffset++ // Keep this line to increment the offset
			applyOffset++ // Increment offset *after* successful insertion
		case "delete":
			logger.Debug("Deleting line", "targetIndex", targetIndex, "value", modifiedLines[targetIndex])
			modifiedLines = append(modifiedLines[:targetIndex], modifiedLines[targetIndex+1:]...)
			applyOffset-- // Decrement offset *after* successful deletion
		}
		logger.Debug("State after change", "index", i, "nextOffset", applyOffset, "currentLineCount", len(modifiedLines))
	}

	// --- Replace previous debug block with this UNCONDITIONAL print ---
	fmt.Println(">>> APPLY PATCH FUNCTION WAS DEFINITELY CALLED <<<")
	// --- End of simplified print ---

	// --- Add this block just before the final return ---
	// Heuristic to target the failing test case (inserting 2 lines into empty)
	if len(originalLines) == 0 && len(changes) == 2 {
		fmt.Printf("[DEBUG ApplyPatch End - InsertEmpty Case] Final modifiedLines (len=%d):\n", len(modifiedLines))
		for idx, line := range modifiedLines {
			// Use %q to clearly show quotes and escaped characters if any
			fmt.Printf("  [%d]: %q\n", idx, line)
		}
	}
	// --- End of added block ---

	logger.Debug("--- Finished ApplyPatch Successfully ---")
	return modifiedLines, nil // Success
}

// --- File Loading Helper ---
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
			// Include file path in syntax error message
			return nil, fmt.Errorf("%w: unmarshaling json in %q at line %d, char %d: %w", ErrInvalidPatchFile, filePath, line, char, err)
		}
		// Include file path in general unmarshal error message
		return nil, fmt.Errorf("%w: unmarshaling json from %q: %w", ErrInvalidPatchFile, filePath, err)
	}

	// Validate individual changes after successful unmarshal
	for i, change := range changes {
		// File field is optional, no check needed.
		if change.Line <= 0 {
			return nil, fmt.Errorf("%w: change #%d in patch file %q has invalid 'line': %d (must be >= 1)", ErrOutOfBounds, i+1, filePath, change.Line)
		}
		switch change.Operation {
		case "replace", "insert":
			// Check if 'new' field is missing (nil pointer)
			if change.NewLine == nil {
				return nil, fmt.Errorf("%w: change #%d (%s line %d) in patch file %q missing required 'new' field", ErrMissingField, i+1, change.Operation, change.Line, filePath)
			}
		case "delete":
			// 'old' field is optional for verification, not required for loading
			// 'new' field is not applicable
		default:
			return nil, fmt.Errorf("%w: change #%d in patch file %q has unknown operation: %q", ErrInvalidOperation, i+1, filePath, change.Operation)
		}
		// Add check for OldLine being nil if needed by application logic validation beyond basic parsing
	}
	return changes, nil
}

// calculateErrorPosition calculates the line and character number from a byte offset.
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
			// Assuming UTF-8, multi-byte characters still advance char position by 1 visually
			// This might not be perfect for complex scripts but works for basic JSON/text
			char++
		}
		currentOffset++
	}
	// Handle case where error is exactly at the end of file or line
	if offset > 0 && offset == currentOffset {
		// If the character *at* the offset (which caused the error) is a newline,
		// the error is arguably on the *next* line, position 1.
		// However, syntax errors often point *after* the problematic token.
		// Sticking to the position *before* the newline seems more common for JSON errors.
	}

	return line, char
}
