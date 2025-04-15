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

// --- Structs (PatchChange, VerificationResult remain the same) ---
type PatchChange struct {
	File                     string  `json:"file"`
	LineNumber               int     `json:"line_number"`
	Operation                string  `json:"operation"`
	OriginalLineForReference *string `json:"original_line_for_reference"`
	NewLineContent           *string `json:"new_line_content"`
}

type VerificationResult struct {
	ChangeIndex int
	LineNumber  int
	TargetIndex int
	Operation   string
	Status      string
	IsError     bool
	Err         error
}

// --- Core Logic Functions ---

// VerifyChanges performs the verification pass against the provided lines.
func VerifyChanges(originalLines []string, changes []PatchChange) ([]VerificationResult, error) {
	results := make([]VerificationResult, 0, len(changes))
	var firstError error = nil
	currentContentLen := len(originalLines) // Track conceptual length
	verificationOffset := 0

	for i, change := range changes {
		targetIndex := change.LineNumber - 1 + verificationOffset
		status := "Not Checked"
		isOutOfBounds := false
		isError := false
		var currentError error = nil

		res := VerificationResult{
			ChangeIndex: i,
			LineNumber:  change.LineNumber,
			TargetIndex: targetIndex,
			Operation:   change.Operation,
		}

		// Bounds checks - Use currentContentLen which tracks conceptual length
		if targetIndex < 0 {
			isOutOfBounds = true
			isError = true
			currentError = fmt.Errorf("%w: invalid target index %d (from line %d, offset %d)", ErrOutOfBounds, targetIndex, change.LineNumber, verificationOffset)
			status = fmt.Sprintf("Error: %v", currentError)
		} else if (change.Operation == "replace" || change.Operation == "delete" || change.OriginalLineForReference != nil) && targetIndex >= currentContentLen { // Check against conceptual length
			isOutOfBounds = true
			if change.Operation == "replace" || change.Operation == "delete" {
				isError = true
				currentError = fmt.Errorf("%w: target index %d for %s (conceptual lines: %d, offset: %d)", ErrOutOfBounds, targetIndex, change.Operation, currentContentLen, verificationOffset)
				status = fmt.Sprintf("Error: %v", currentError)
			} else {
				status = "Out Of Bounds"
			}
		} else if change.Operation == "insert" && targetIndex > currentContentLen { // Check against conceptual length
			isOutOfBounds = true
			isError = true
			currentError = fmt.Errorf("%w: target index %d for insert (conceptual lines: %d, offset: %d)", ErrOutOfBounds, targetIndex, currentContentLen, verificationOffset)
			status = fmt.Sprintf("Error: %v", currentError)
		}

		// Verification check - Use originalLines with calculated targetIndex
		if !isError && !isOutOfBounds && change.OriginalLineForReference != nil {
			originalFromFile := strings.TrimSuffix(originalLines[targetIndex], "\n")
			if originalFromFile == *change.OriginalLineForReference {
				status = "Matched"
			} else {
				status = fmt.Sprintf("MISMATCHED (Expected: %q, Found: %q)", *change.OriginalLineForReference, originalFromFile)
				isError = true
				currentError = fmt.Errorf("%w: expected %q, found %q", ErrVerificationFailed, *change.OriginalLineForReference, originalFromFile)
			}
		} else if change.OriginalLineForReference != nil && isOutOfBounds && status == "Out Of Bounds" {
			// Status already set
		} else if !isError && !isOutOfBounds && change.OriginalLineForReference == nil && (change.Operation == "replace" || change.Operation == "delete") {
			status = "Not Verified (No Ref)"
		}

		res.Status = status
		res.IsError = isError
		res.Err = currentError
		results = append(results, res)

		if isError && firstError == nil {
			firstError = fmt.Errorf("change #%d (%s line %d): %w", i+1, change.Operation, change.LineNumber, currentError)
		}

		// Update conceptual offset AND length for the next check
		switch change.Operation {
		case "insert":
			verificationOffset++
			currentContentLen++ // Content grew by 1
		case "delete":
			if !isOutOfBounds { // Only adjust if delete target was valid
				verificationOffset--
				currentContentLen-- // Content shrank by 1
			}
		}
	} // End verification loop

	return results, firstError
}

// ApplyPatch performs the two-pass (verify, then apply) patch operation.
func ApplyPatch(originalLines []string, changes []PatchChange) ([]string, error) {
	// --- Pass 1: Verify ---
	// Use the updated VerifyChanges logic
	_, firstVerificationError := VerifyChanges(originalLines, changes)
	if firstVerificationError != nil {
		return nil, firstVerificationError
	}

	// --- Pass 2: Application ---
	modifiedLines := make([]string, len(originalLines))
	copy(modifiedLines, originalLines)
	applyOffset := 0

	for i, change := range changes {
		targetIndex := change.LineNumber - 1 + applyOffset

		switch change.Operation {
		case "replace":
			if targetIndex < 0 || targetIndex >= len(modifiedLines) {
				return nil, fmt.Errorf("%w: change #%d (%s line %d): index %d became invalid during apply (lines: %d)", ErrInternal, i+1, change.Operation, change.LineNumber, targetIndex, len(modifiedLines))
			}
			modifiedLines[targetIndex] = *change.NewLineContent
		case "insert":
			if targetIndex < 0 || targetIndex > len(modifiedLines) {
				return nil, fmt.Errorf("%w: change #%d (%s line %d): index %d became invalid during apply (lines: %d)", ErrInternal, i+1, change.Operation, change.LineNumber, targetIndex, len(modifiedLines))
			}
			// Correct insertion: Insert element *before* targetIndex
			// Need temporary slice to hold the new element
			newLine := *change.NewLineContent
			// Append the part after targetIndex first (if any)
			tail := []string{}
			if targetIndex < len(modifiedLines) {
				tail = modifiedLines[targetIndex:]
			}
			// Append the new line
			modifiedLines = append(modifiedLines[:targetIndex], newLine)
			// Append the tail
			modifiedLines = append(modifiedLines, tail...)

			applyOffset++
		case "delete":
			if targetIndex < 0 || targetIndex >= len(modifiedLines) {
				return nil, fmt.Errorf("%w: change #%d (%s line %d): index %d became invalid during apply (lines: %d)", ErrInternal, i+1, change.Operation, change.LineNumber, targetIndex, len(modifiedLines))
			}
			modifiedLines = append(modifiedLines[:targetIndex], modifiedLines[targetIndex+1:]...)
			applyOffset--
		}
	}

	return modifiedLines, nil // Success
}

// --- File Loading Helper ---
// LoadPatchFile function (remains the same as previous version)
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

	// Basic validation of loaded changes
	for i, change := range changes {
		if change.File == "" { /* Allow empty file field */
		}
		if change.LineNumber <= 0 {
			return nil, fmt.Errorf("%w: change #%d in file %q has invalid 'line_number': %d", ErrOutOfBounds, i+1, change.File, change.LineNumber)
		}
		switch change.Operation {
		case "replace", "insert":
			if change.NewLineContent == nil {
				return nil, fmt.Errorf("%w: change #%d (%s line %d) in file %q missing 'new_line_content'", ErrMissingField, i+1, change.Operation, change.LineNumber, change.File)
			}
		case "delete": /* Ok */
		default:
			return nil, fmt.Errorf("%w: change #%d in file %q has unknown operation: %q", ErrInvalidOperation, i+1, change.File, change.Operation)
		}
	}
	return changes, nil
}

// calculateErrorPosition function (remains the same)
func calculateErrorPosition(data []byte, offset int64) (line, char int) {
	line = 1
	char = 1
	for i, r := range string(data) {
		if int64(i) >= offset {
			break
		}
		if r == '\n' {
			line++
			char = 1
		} else {
			char++
		}
	}
	return line, char
}
