// NeuroScript Version: 0.3.1
// File version: 1
// Purpose: Contains listener methods for handling file header metadata.
// filename: pkg/core/ast_builder_metadata.go
// nlines: 35
// risk_rating: LOW

package core

import (
	"errors"
	"strings"

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

func (l *neuroScriptListenerImpl) EnterFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("   >> Enter File Header")
	if l.program == nil || l.program.Metadata == nil {
		l.logger.Error("EnterFile_header called with nil program or metadata map! This should have been initialized.")
		// Attempt to recover if program is nil, which is highly problematic
		if l.program == nil {
			l.program = &Program{Metadata: make(map[string]string), Procedures: make(map[string]*Procedure)}
		} else if l.program.Metadata == nil { // If only metadata is nil
			l.program.Metadata = make(map[string]string)
		}
		l.fileMetadata = l.program.Metadata // Re-assign
		l.errors = append(l.errors, errors.New("internal AST builder error: program/metadata nil in EnterFile_header"))
		// No return here, try to process metadata anyway
	}
	for _, metaLineNode := range ctx.AllMETADATA_LINE() {
		lineText := metaLineNode.GetText()
		token := metaLineNode.GetSymbol()
		l.logDebugAST("   - Processing File Metadata Line: %s", lineText)
		// The lexer rule for METADATA_LINE: [\t ]* '::' [ \t]+ ~[\r\n]* ;
		idx := strings.Index(lineText, "::")
		if idx == -1 { // Should not happen if lexer rule is correct
			l.addErrorf(token, "METADATA_LINE token did not contain '::' separator as expected: '%s'", lineText)
			continue
		}

		contentAfterDoubleColon := lineText[idx+2:] // Content after "::"
		trimmedContent := strings.TrimSpace(contentAfterDoubleColon)

		parts := strings.SplitN(trimmedContent, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				if _, exists := l.program.Metadata[key]; exists {
					l.logDebugAST("     Overwriting File Metadata: '%s'", key)
				}
				l.program.Metadata[key] = value
				l.logDebugAST("     Stored File Metadata: '%s' = '%s'", key, value)
			} else {
				l.addErrorf(token, "Ignoring file metadata line with empty key: '%s'", lineText)
			}
		} else { // Only one part, means no ':' or key is empty if content was just ':'
			keyOnly := strings.TrimSpace(parts[0])
			if keyOnly != "" {
				if _, exists := l.program.Metadata[keyOnly]; exists {
					l.logDebugAST("     Overwriting File Metadata (key only): '%s' to empty value", keyOnly)
				}
				l.program.Metadata[keyOnly] = "" // Store key with empty value
				l.logDebugAST("     Stored File Metadata (key only): '%s' = ''", keyOnly)

			} else {
				l.addErrorf(token, "Ignoring malformed file metadata line (empty key or content after '::'): '%s'", lineText)
			}
		}
	}
}

func (l *neuroScriptListenerImpl) ExitFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("   << Exit File Header")
}
