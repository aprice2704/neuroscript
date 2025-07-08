// filename: pkg/parser/ast_builder_metadata.go
// NeuroScript Version: 0.5.2
// File version: 6
// Purpose: Adds a dedicated listener for file_header and cleans up metadata_block logic.
package parser

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
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
	parts := strings.SplitN(trimmedContent, ":", 2)
	key := strings.TrimSpace(parts[0])
	var value string
	if len(parts) == 2 {
		value = strings.TrimSpace(parts[1])
	}

	if key != "" {
		targetMap[key] = value
		l.logDebugAST("     Stored Metadata: '%s' = '%s'", key, value)
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
}

// ExitMetadata_block handles metadata ONLY within a procedure or command block.
func (l *neuroScriptListenerImpl) ExitMetadata_block(ctx *gen.Metadata_blockContext) {
	l.logDebugAST("  << Exit Metadata_block")
	var targetMap map[string]string

	// MODIFIED: Logic simplified to only handle block-level contexts.
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
		// File-level metadata is now handled by ExitFile_header. This is an error.
		l.addError(ctx, "metadata_block found outside of a known context (procedure or command)")
		return
	}

	for _, metaLineNode := range ctx.AllMETADATA_LINE() {
		l.processMetadataLine(targetMap, metaLineNode.GetSymbol())
	}
}
