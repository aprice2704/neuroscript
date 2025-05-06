// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Fix line numbering for inserts following deletes (replaces).
// Implements the GeneratePatch tool using go-diff.
// filename: pkg/nspatch/tools_generate.go

package nspatch

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	diffmatchpatch "github.com/sergi/go-diff/diffmatchpatch"
)

// --- Tool Definition: GeneratePatch ---

var toolGeneratePatchImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name: "GeneratePatch",
		Description: "Compares original_content and modified_content line by line and generates a list of patch operations " +
			"(compatible with ApplyPatch) required to transform the original into the modified.",
		Args: []core.ArgSpec{
			{Name: "original_content", Type: core.ArgTypeString, Required: true, Description: "The original text content."},
			{Name: "modified_content", Type: core.ArgTypeString, Required: true, Description: "The modified text content."},
			{Name: "path", Type: core.ArgTypeString, Required: false, Description: "(Optional) The file path this patch pertains to. If provided, it will be added to each operation."},
		},
		// Returns a list of maps, each representing a PatchChange struct
		ReturnType: core.ArgTypeSliceMap,
	},
	Func: toolGeneratePatch,
}

// normalizeNewlines ensures consistent \n endings for diffing
func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	// Ensure ends with newline for consistent splitting, unless empty
	if s != "" && !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return s
}

// splitLines handles splitting text into lines, ensuring trailing newline behavior is correct.
func splitLines(text string) []string {
	// If empty string, return empty slice, not {""}
	if text == "" {
		return []string{}
	}
	lines := strings.Split(text, "\n")
	// If the original text ended with \n, Split creates an extra empty string at the end.
	// Keep it if the text was just "\n", otherwise remove it.
	if strings.HasSuffix(text, "\n") && len(lines) > 1 {
		lines = lines[:len(lines)-1]
	} else if text == "\n" { // Handle the case of exactly one newline
		lines = []string{""} // Represent a single empty line
	}

	return lines
}

// toolGeneratePatch implements the GeneratePatch tool.
func toolGeneratePatch(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	logger := interpreter.Logger()

	// --- Argument Parsing ---
	if len(args) != 3 {
		return nil, fmt.Errorf("%w: GeneratePatch requires 3 arguments (original_content, modified_content, path)", core.ErrInvalidArgument)
	}
	originalContent, okO := args[0].(string)
	modifiedContent, okM := args[1].(string)
	filePath := ""
	if args[2] != nil {
		pathArg, okP := args[2].(string)
		if !okP {
			return nil, fmt.Errorf("%w: GeneratePatch invalid type for path argument, expected string, got %T", core.ErrInvalidArgument, args[2])
		}
		filePath = pathArg
	}

	if !okO || !okM {
		return nil, fmt.Errorf("%w: GeneratePatch invalid argument types (expected string, string, string|nil)", core.ErrInvalidArgument)
	}

	logger.Debug("[TOOL-GENERATEPATCH] Request", "originalLen", len(originalContent), "modifiedLen", len(modifiedContent), "path", filePath)

	// Normalize content
	normOriginal := normalizeNewlines(originalContent)
	normModified := normalizeNewlines(modifiedContent)

	// --- Diff Calculation using LinesToChars ---
	dmp := diffmatchpatch.New()
	charsA, charsB, lineArray := dmp.DiffLinesToChars(normOriginal, normModified)
	diffs := dmp.DiffMain(charsA, charsB, false)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)

	// --- Convert Diffs to PatchChange structs ---
	patches := []PatchChange{}
	currentLineNumber := 1                                               // Tracks the line number in the *original* file context
	var lastDeleteStartLine int = -1                                     // Track the start line of the last delete block
	var prevDiffType diffmatchpatch.Operation = diffmatchpatch.DiffEqual // Initialize previous diff type

	for _, diff := range diffs {
		lines := splitLines(diff.Text) // Use helper for splitting

		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			// For equal lines, just advance the line counter based on original lines consumed
			currentLineNumber += len(lines)
			lastDeleteStartLine = -1 // Reset delete tracking
		case diffmatchpatch.DiffDelete:
			// Generate a 'delete' operation for each line deleted
			lastDeleteStartLine = currentLineNumber // Record where this delete block started
			for _, lineContent := range lines {
				deletedLine := lineContent // Allocate memory for pointer
				patch := PatchChange{
					Operation: "delete",
					File:      filePath,
					Line:      currentLineNumber, // Deletes happen *at* the current original line number
					OldLine:   &deletedLine,
					NewLine:   nil,
				}
				patches = append(patches, patch)
				currentLineNumber++ // Increment original line number *after* recording the delete for that line
			}
		case diffmatchpatch.DiffInsert:
			// Generate an 'insert' operation for each line inserted
			insertTargetLine := currentLineNumber // Default: insert happens before the current line context

			// *** FIXED LOGIC for REPLACES ***
			// If this insert immediately follows a delete, the insert should target
			// the line number where the delete *started*.
			if prevDiffType == diffmatchpatch.DiffDelete && lastDeleteStartLine != -1 {
				insertTargetLine = lastDeleteStartLine
			}

			for _, lineContent := range lines {
				insertedLine := lineContent // Allocate memory for pointer
				patch := PatchChange{
					Operation: "insert",
					File:      filePath,
					Line:      insertTargetLine, // Use the calculated (potentially adjusted) target line
					OldLine:   nil,
					NewLine:   &insertedLine,
				}
				patches = append(patches, patch)
				// Increment the target line number for the *next* inserted line within this block
				insertTargetLine++
			}
			lastDeleteStartLine = -1 // Reset delete tracking after an insert
			// *** DO NOT increment currentLineNumber here *** - Inserts don't consume original lines
		}
		// Update previous diff type for the next iteration
		prevDiffType = diff.Type
	}

	// Convert []PatchChange to []interface{} of maps for tool return
	result := make([]interface{}, len(patches))
	for i, p := range patches {
		patchMap := map[string]interface{}{
			"op":   p.Operation,
			"file": p.File,
			"line": int64(p.Line), // Use int64
		}
		if p.OldLine != nil {
			patchMap["old"] = *p.OldLine
		}
		if p.NewLine != nil {
			patchMap["new"] = *p.NewLine
		}
		result[i] = patchMap
	}

	logger.Info("[TOOL-GENERATEPATCH] Patch generation complete", "opsGenerated", len(result))
	return result, nil
}
