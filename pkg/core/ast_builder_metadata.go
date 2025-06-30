// filename: pkg/core/ast_builder_metadata.go
// NeuroScript Version: 0.5.2
// File version: 4
// Purpose: Made ExitMetadata_block context-aware to handle metadata for both procedures and commands.
package core

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// processMetadataLine is a reusable helper to parse a single metadata line token.
func (l *neuroScriptListenerImpl) processMetadataLine(targetMap map[string]string, token antlr.Token) {
	lineText := token.GetText()
	l.logDebugAST("   - Processing Metadata Line: %s", lineText)

	// This logic correctly parses ":: key: value"
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

// ExitMetadata_block handles metadata within a procedure or command block.
func (l *neuroScriptListenerImpl) ExitMetadata_block(ctx *gen.Metadata_blockContext) {
	l.logDebugAST("  << Exit Metadata_block")
	var targetMap map[string]string

	// MODIFIED: Determine the context (procedure or command) and select the correct map.
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
		l.addError(ctx, "metadata_block found outside of a procedure or command context")
		return
	}

	for _, metaLineNode := range ctx.AllMETADATA_LINE() {
		l.processMetadataLine(targetMap, metaLineNode.GetSymbol())
	}
}
