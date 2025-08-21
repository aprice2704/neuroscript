// NeuroScript Version: 0.6.0
// File version: 30
// Purpose: Re-instated correct line-number-based logic for blank line detection WHILE RETAINING VERBOSE DEBUGGING to validate the final fix for the metadata bug.
// filename: pkg/parser/ast_builder_metadata.go
// nlines: 200
// risk_rating: HIGH

package parser

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
)

// ExitFile_header handles top-level metadata, adding it to the pending map.
func (l *neuroScriptListenerImpl) ExitFile_header(ctx *gen.File_headerContext) {
	// This function populates the temporary holding map for metadata.
	for _, metaLineNode := range ctx.AllMETADATA_LINE() {
		token := metaLineNode.GetSymbol()
		if key, value, ok := l.parseSingleMetadataLine(token); ok {
			l.pendingMetadata[key] = value
			l.lastPendingMetadataToken = token
		}
	}
}

// ExitMetadata_block handles metadata ONLY within a procedure, command, or event block.
func (l *neuroScriptListenerImpl) ExitMetadata_block(ctx *gen.Metadata_blockContext) {
	var targetMap map[string]string
	// Determine which AST node (procedure, command, etc.) is currently being built.
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
	} else if l.currentEvent != nil {
		if l.currentEvent.Metadata == nil {
			l.currentEvent.Metadata = make(map[string]string)
		}
		targetMap = l.currentEvent.Metadata
	} else {
		l.addError(ctx, "metadata_block found outside of a known context (procedure, command, or event)")
		return
	}
	// Directly assign metadata to the current block's map.
	for _, metaLineNode := range ctx.AllMETADATA_LINE() {
		token := metaLineNode.GetSymbol()
		if key, value, ok := l.parseSingleMetadataLine(token); ok {
			targetMap[key] = value
		}
	}
}

// parseSingleMetadataLine extracts a key-value pair from a metadata token.
func (l *neuroScriptListenerImpl) parseSingleMetadataLine(token antlr.Token) (key, value string, isValid bool) {
	lineText := token.GetText()
	idx := strings.Index(lineText, "::")
	if idx == -1 {
		l.addErrorf(token, "METADATA_LINE token did not contain '::' separator: '%s'", lineText)
		return "", "", false
	}
	contentAfterDoubleColon := lineText[idx+2:]
	trimmedContent := strings.TrimSpace(contentAfterDoubleColon)
	commentIdx := strings.Index(trimmedContent, "#")
	if commentIdx != -1 {
		trimmedContent = trimmedContent[:commentIdx]
	}
	commentIdx = strings.Index(trimmedContent, "--")
	if commentIdx != -1 {
		trimmedContent = trimmedContent[:commentIdx]
	}
	parts := strings.SplitN(trimmedContent, ":", 2)
	key = strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		value = strings.TrimSpace(parts[1])
	}
	if key == "" {
		l.addErrorf(token, "Ignoring metadata line with empty key: '%s'", lineText)
		return "", "", false
	}
	return key, value, true
}

// assignPendingMetadata is the centralized logic for assigning top-level metadata.
func (l *neuroScriptListenerImpl) assignPendingMetadata(declarationToken antlr.Token, targetMap map[string]string) {
	//	fmt.Fprintf(os.Stderr, "\n[DEBUG] ========= assignPendingMetadata (v3) =========\n")
	//	fmt.Fprintf(os.Stderr, "[DEBUG] Pending Metadata Count: %d\n", len(l.pendingMetadata))
	// if len(l.pendingMetadata) == 0 {
	// 	fmt.Fprintf(os.Stderr, "[DEBUG] No pending metadata. Exiting.\n")
	// 	fmt.Fprintf(os.Stderr, "[DEBUG] =============================================\n")
	// 	return
	// }

	if l.lastPendingMetadataToken != nil {
		//		fmt.Fprintf(os.Stderr, "[DEBUG] Last Metadata Token: '%s' (Line: %d)\n", strings.TrimSpace(l.lastPendingMetadataToken.GetText()), l.lastPendingMetadataToken.GetLine())
	}
	if declarationToken != nil {
		//		fmt.Fprintf(os.Stderr, "[DEBUG] Declaration Token:   '%s' (Line: %d)\n", strings.TrimSpace(declarationToken.GetText()), declarationToken.GetLine())
	} else {
		//		fmt.Fprintf(os.Stderr, "[DEBUG] Declaration Token:   nil (end of file assignment)\n")
	}

	isSeparatedByBlankLine := declarationToken == nil || l.hasBlankLineBetweenTokens(l.lastPendingMetadataToken, declarationToken)
	//	fmt.Fprintf(os.Stderr, "[DEBUG] Call to hasBlankLineBetweenTokens returned: %v\n", isSeparatedByBlankLine)

	if isSeparatedByBlankLine {
		if l.program.Metadata == nil {
			l.program.Metadata = make(map[string]string)
		}
		for k, v := range l.pendingMetadata {
			l.program.Metadata[k] = v
		}
		//		fmt.Fprintf(os.Stderr, "[DEBUG] DECISION -> Assigned %d items to FILE scope.\n", len(l.pendingMetadata))
	} else if targetMap != nil {
		for k, v := range l.pendingMetadata {
			targetMap[k] = v
		}
		//		fmt.Fprintf(os.Stderr, "[DEBUG] DECISION -> Assigned %d items to BLOCK scope.\n", len(l.pendingMetadata))
	} else {
		//		fmt.Fprintf(os.Stderr, "[DEBUG] DECISION -> Metadata DROPPED (not file-level and no target map).\n")
	}

	l.pendingMetadata = make(map[string]string)
	l.lastPendingMetadataToken = nil
	// fmt.Fprintf(os.Stderr, "[DEBUG] Cleared pending metadata.\n")
	// fmt.Fprintf(os.Stderr, "[DEBUG] =============================================\n\n")
}

// hasBlankLineBetweenTokens checks for a semantic blank line between two tokens.
func (l *neuroScriptListenerImpl) hasBlankLineBetweenTokens(start, end antlr.Token) bool {
	if start == nil || end == nil {
		//		fmt.Fprintf(os.Stderr, "  [DEBUG/hasBlankLine] called with nil token, returning false.\n")
		return false
	}

	startLine := start.GetLine()
	endLine := end.GetLine()

	// THIS IS THE CORRECTED LOGIC.
	// If end is on line 5 and start is on line 3, (5-3) is 2. This means line 4 is blank.
	// If end is on line 4 and start is on line 3, (4-3) is 1. No blank line.
	isSeparated := (endLine - startLine) >= 2

	//	fmt.Fprintf(os.Stderr, "  [DEBUG/hasBlankLine] Start Line: %d | End Line: %d | Calculation: (end - start >= 2) -> (%d >= 2) -> %v\n",
	//		startLine, endLine, (endLine - startLine), isSeparated)

	return isSeparated
}
