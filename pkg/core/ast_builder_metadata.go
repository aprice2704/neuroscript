// NeuroScript Version: 0.3.1
// File version: 3
// Purpose: Repurposed to handle procedure-level metadata via the 'metadata_block' rule.
// filename: pkg/core/ast_builder_metadata.go
// nlines: 48
// risk_rating: LOW

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
		if _, exists := targetMap[key]; exists {
			l.logDebugAST("     Overwriting Metadata: '%s'", key)
		}
		targetMap[key] = value
		l.logDebugAST("     Stored Metadata: '%s' = '%s'", key, value)
	} else {
		l.addErrorf(token, "Ignoring metadata line with empty key: '%s'", lineText)
	}
}

// ExitMetadata_block handles the metadata block within a procedure.
func (l *neuroScriptListenerImpl) ExitMetadata_block(ctx *gen.Metadata_blockContext) {
	l.logDebugAST("  << Exit Metadata_block")
	if l.currentProc == nil {
		l.addError(ctx, "metadata_block found outside of a procedure context")
		return
	}
	// Ensure the procedure's metadata map is initialized
	if l.currentProc.Metadata == nil {
		l.currentProc.Metadata = make(map[string]string)
	}

	for _, metaLineNode := range ctx.AllMETADATA_LINE() {
		l.processMetadataLine(l.currentProc.Metadata, metaLineNode.GetSymbol())
	}
}
