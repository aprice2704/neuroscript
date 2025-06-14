// NeuroScript Version: 0.3.1
// File version: 2
// Purpose: Corrects metadata handling by removing faulty nil check and simplifying logic.
// filename: pkg/core/ast_builder_metadata.go
// nlines: 41
// risk_rating: MEDIUM

package core

import (
	"strings"

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

func (l *neuroScriptListenerImpl) EnterFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("   >> Enter File Header")
	// The listener's fileMetadata map is guaranteed to be initialized by newNeuroScriptListener.
	// We directly populate it without nil checks.

	for _, metaLineNode := range ctx.AllMETADATA_LINE() {
		lineText := metaLineNode.GetText()
		token := metaLineNode.GetSymbol()
		l.logDebugAST("   - Processing File Metadata Line: %s", lineText)

		idx := strings.Index(lineText, "::")
		if idx == -1 {
			l.addErrorf(token, "METADATA_LINE token did not contain '::' separator as expected: '%s'", lineText)
			continue
		}

		contentAfterDoubleColon := lineText[idx+2:]
		trimmedContent := strings.TrimSpace(contentAfterDoubleColon)

		parts := strings.SplitN(trimmedContent, ":", 2)
		var key, value string
		key = strings.TrimSpace(parts[0])

		if len(parts) == 2 {
			value = strings.TrimSpace(parts[1])
		}

		if key != "" {
			if _, exists := l.fileMetadata[key]; exists {
				l.logDebugAST("     Overwriting File Metadata: '%s'", key)
			}
			l.fileMetadata[key] = value
			l.logDebugAST("     Stored File Metadata: '%s' = '%s'", key, value)
		} else {
			l.addErrorf(token, "Ignoring file metadata line with empty key: '%s'", lineText)
		}
	}
}

func (l *neuroScriptListenerImpl) ExitFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("   << Exit File Header")
}
