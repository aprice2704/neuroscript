// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 21:04:41 PDT // Added GoImports tool
// filename: pkg/core/tools_go_fmt.go

package core

import (
	"bytes"
	"fmt"
	"go/format"

	// Import the goimports library
	"golang.org/x/tools/imports"
)

// toolGoFmt implementation (uses format package directly, no shell execution)
// --- (toolGoFmt implementation remains unchanged) ---
func toolGoFmt(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures 1 string argument (handled by interpreter)
	content := args[0].(string)
	srcBytes := []byte(content)

	if interpreter.logger != nil {
		logSnippet := content
		if len(logSnippet) > 100 {
			logSnippet = logSnippet[:100] + "..."
		}
		interpreter.logger.Debug("[TOOL-GOFMT] Formatting content (snippet): %q", logSnippet)
	}

	formattedBytes, fmtErr := format.Source(srcBytes)

	if fmtErr == nil {
		// Formatting succeeded
		formattedContent := string(formattedBytes)
		if interpreter.logger != nil {
			logMsg := "[TOOL-GOFMT] Successful (no changes needed)."
			if !bytes.Equal(srcBytes, formattedBytes) {
				logMsg = "[TOOL-GOFMT] Successful (content changed)."
			}
			interpreter.logger.Debug(logMsg)
		}
		// Return formatted string directly on success, nil Go error
		return formattedContent, nil
	}

	// Formatting failed (e.g., syntax error)
	errorString := fmtErr.Error()
	if interpreter.logger != nil {
		interpreter.logger.Error("[TOOL-GOFMT] Failed.", "error", errorString)
	}
	resultMap := map[string]interface{}{
		"formatted_content": content, // Return original content on error
		"error":             errorString,
		"success":           false,
	}
	// Return map AND wrap the original Go error for the interpreter
	return resultMap, fmt.Errorf("%w: formatting failed: %w", ErrInternalTool, fmtErr)
}

// --- New Tool: GoImports ---

// toolGoImports formats Go source code and adjusts imports using golang.org/x/tools/imports.
func toolGoImports(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation ensures 1 string argument (handled by interpreter)
	content := args[0].(string)
	srcBytes := []byte(content)
	filename := "" // Process assumes content is a standalone fragment unless filename context is needed/provided

	if interpreter.logger != nil {
		logSnippet := content
		if len(logSnippet) > 100 {
			logSnippet = logSnippet[:100] + "..."
		}
		interpreter.logger.Debug("[TOOL-GOIMPORTS] Processing content (snippet): %q", logSnippet)
	}

	// Default options for imports.Process
	// We might need to expose options later if needed (e.g., Fragment, Comments, TabIndent, TabWidth)
	options := &imports.Options{
		Fragment:  false, // Assume it's a full file
		AllErrors: true,  // Report all errors (not just first)
		Comments:  true,  // Keep comments
		TabIndent: true,  // Use tabs for indentation
		TabWidth:  8,     // Standard tab width
	}

	// Process the source code
	formattedBytes, importErr := imports.Process(filename, srcBytes, options)

	if importErr == nil {
		// Processing succeeded
		formattedContent := string(formattedBytes)
		if interpreter.logger != nil {
			logMsg := "[TOOL-GOIMPORTS] Successful (no changes needed)."
			if !bytes.Equal(srcBytes, formattedBytes) {
				logMsg = "[TOOL-GOIMPORTS] Successful (content changed)."
			}
			interpreter.logger.Debug(logMsg)
		}
		// Return formatted string directly on success, nil Go error
		return formattedContent, nil
	}

	// Processing failed (e.g., syntax error, import resolution issues)
	errorString := importErr.Error()
	if interpreter.logger != nil {
		interpreter.logger.Error("[TOOL-GOIMPORTS] Failed.", "error", errorString)
	}
	resultMap := map[string]interface{}{
		"formatted_content": content, // Return original content on error
		"error":             errorString,
		"success":           false,
	}
	// Return map AND wrap the original Go error for the interpreter
	return resultMap, fmt.Errorf("%w: goimports processing failed: %w", ErrInternalTool, importErr)
}
