// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Redesign for line-based diff and PatchChange struct.
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

	// Normalize content and split into lines for line-based diff
	// IMPORTANT: DiffLinesToChars works best if lines end consistently (e.g., with \n)
	normOriginal := normalizeNewlines(originalContent)
	normModified := normalizeNewlines(modifiedContent)

	// --- Diff Calculation using LinesToChars ---
	dmp := diffmatchpatch.New()
	// Encode lines to chars for efficient diffing
	charsA, charsB, lineArray := dmp.DiffLinesToChars(normOriginal, normModified)
	// Perform diff on encoded strings
	diffs := dmp.DiffMain(charsA, charsB, false) // Use word mode? false=char mode on encoded lines
	// Convert back to line-based diffs
	diffs = dmp.DiffCharsToLines(diffs, lineArray)

	// --- Convert Diffs to PatchChange structs ---
	patches := []PatchChange{}
	originalLineNo := 1 // 1-based line number for PatchChange

	for _, diff := range diffs {
		// Split the text block into individual lines for processing
		// Note: Split will produce an empty string after the last \n if the text ends with \n
		lines := strings.Split(diff.Text, "\n")
		// If the original diff text ended with \n, Split creates an extra empty string at the end. Remove it.
		if strings.HasSuffix(diff.Text, "\n") && len(lines) > 0 {
			lines = lines[:len(lines)-1]
		}

		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			// For equal lines, just advance the line counter
			originalLineNo += len(lines)
		case diffmatchpatch.DiffDelete:
			// Generate a 'delete' operation for each line deleted
			for _, line := range lines {
				// Need to allocate memory for the string pointer
				deletedLine := line
				patch := PatchChange{
					Operation: "delete",
					File:      filePath,
					Line:      originalLineNo, // Deletes happen *at* the original line number
					OldLine:   &deletedLine,   // Provide original line for verification
					NewLine:   nil,            // Not applicable for delete
				}
				patches = append(patches, patch)
				originalLineNo++ // Increment the line number *in the original file* for each deleted line
			}
		case diffmatchpatch.DiffInsert:
			// Generate an 'insert' operation for each line inserted
			for _, line := range lines {
				// Need to allocate memory for the string pointer
				insertedLine := line
				patch := PatchChange{
					Operation: "insert",
					File:      filePath,
					Line:      originalLineNo, // Inserts happen *before* the original line number
					OldLine:   nil,            // Not applicable for insert
					NewLine:   &insertedLine,
				}
				patches = append(patches, patch)
				// Do *not* increment originalLineNo for inserts, as they don't consume original lines.
				// The ApplyPatch logic needs to handle inserting before the specified original line number.
			}
		}
	}

	// Convert []PatchChange to []interface{} of maps for tool return
	result := make([]interface{}, len(patches))
	for i, p := range patches {
		patchMap := map[string]interface{}{
			"op":   p.Operation, // Renamed from "operation"
			"file": p.File,
			"line": int64(p.Line), // Use int64 for consistency
			// Only include optional fields if they have non-nil pointers
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
