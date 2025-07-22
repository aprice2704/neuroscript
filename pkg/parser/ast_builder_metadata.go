// filename: pkg/parser/ast_builder_metadata.go
// NeuroScript Version: 0.5.2
// File version: 8
// Purpose: Corrected metadata parsing to properly strip inline comments.
package parser

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
)

// processMetadataLine is a reusable helper to parse a single metadata line token.
func (l *neuroScriptListenerImpl) processMetadataLine(targetMap map[string]string, token antlr.Token) {
	lineText := token.GetText()
	l.logDebugAST("   - Processing Metadata Line: %s", lineText)

	idx := strings.Index(lineText, "::")
	if idx == -1 {
		l.addErrorf(token, "METADATA_LINE token did not contain '::' separator: '%s'", lineText)
		return
	}
	contentAfterDoubleColon := lineText[idx+2:]
	trimmedContent := strings.TrimSpace(contentAfterDoubleColon)

	// FIX: Strip comments before splitting key-value pairs.
	commentIdx := strings.Index(trimmedContent, "#")
	if commentIdx != -1 {
		trimmedContent = trimmedContent[:commentIdx]
	}
	commentIdx = strings.Index(trimmedContent, "--")
	if commentIdx != -1 {
		trimmedContent = trimmedContent[:commentIdx]
	}

	parts := strings.SplitN(trimmedContent, ":", 2)
	key := strings.TrimSpace(parts[0])
	var value string
	if len(parts) == 2 {
		value = strings.TrimSpace(parts[1])
	}

	if key != "" {
		if targetMap != nil {
			targetMap[key] = value
			l.logDebugAST("     Stored Metadata: '%s' = '%s'", key, value)
		} else {
			l.addErrorf(token, "cannot store metadata for key '%s' because target map is nil", key)
		}
	} else {
		l.addErrorf(token, "Ignoring metadata line with empty key: '%s'", lineText)
	}
}

// ExitFile_header is the dedicated listener for parsing file-level metadata.
func (l *neuroScriptListenerImpl) ExitFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("<< Exit File_header")
	// The target is always the file-level metadata map here.
	for _, metaLineNode := range ctx.AllMETADATA_LINE() {
		l.processMetadataLine(l.fileMetadata, metaLineNode.GetSymbol())
	}
	// After processing, copy to the program's metadata map.
	if l.program != nil {
		if l.program.Metadata == nil {
			l.program.Metadata = make(map[string]string)
		}
		for k, v := range l.fileMetadata {
			l.program.Metadata[k] = v
		}
	}
}

// ExitMetadata_block handles metadata ONLY within a procedure or command block.
func (l *neuroScriptListenerImpl) ExitMetadata_block(ctx *gen.Metadata_blockContext) {
	l.logDebugAST("  << Exit Metadata_block")
	var targetMap map[string]string

	if l.currentProc != nil {
		if l.currentProc.Metadata == nil {
			l.currentProc.Metadata = make(map[string]string)
		}
		targetMap = l.currentProc.Metadata
	} else if l.currentCommand != nil {
		if l.currentCommand.Metadata == nil {
			l.currentCommand.Metadata = make(map[string]string)
		}
		targetMap = l.currentCommand.Metadata
	} else {
		// This case should ideally not be hit if the grammar is followed,
		// as file-level metadata is handled by ExitFile_header.
		l.addError(ctx, "metadata_block found outside of a known context (procedure or command)")
		return
	}

	for _, metaLineNode := range ctx.AllMETADATA_LINE() {
		l.processMetadataLine(targetMap, metaLineNode.GetSymbol())
	}
}
