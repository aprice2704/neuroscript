// NeuroScript Version: 0.3.1
// File version: 0.1.40 // Extracts full tool path (e.g., FS.List) and validates against registry.
// Purpose: Extracts a potential tool name from NeuroScript content at a given LSP position using AST analysis.
// filename: pkg/nslsp/server_extracttool_b.go
// nlines: 220
// risk_rating: MEDIUM

package nslsp

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	lsp "github.com/sourcegraph/go-lsp"
)

const serverExtractToolFileVersion = "0.1.40" // Updated version

// dlogf is a local debug logger helper.
var dlogf = func(logger *log.Logger, debugHover bool, format string, args ...interface{}) {
	if debugHover && logger != nil {
		logger.Printf(format, args...)
	}
}

func truncateStringForLog(s string, maxLen int) string {
	if len(s) == 0 || maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) > maxLen {
		if maxLen <= 3 {
			return string(runes[:maxLen])
		}
		return string(runes[:maxLen-3]) + "..."
	}
	return s
}

func findInitialNodeManually(node antlr.ParseTree, targetLine, targetChar int, logger *log.Logger, isDebug bool) antlr.TerminalNode {
	if node == nil {
		return nil
	}

	if tn, ok := node.(antlr.TerminalNode); ok {
		token := tn.GetSymbol()
		if token == nil {
			return nil
		}
		tokenLine0Based := token.GetLine() - 1
		tokenStartCol0Based := token.GetColumn()
		tokenText := token.GetText()
		tokenEndCol0Based := tokenStartCol0Based + len(tokenText)

		lineMatch := (tokenLine0Based == targetLine)
		charInSpan := (targetChar >= tokenStartCol0Based && targetChar < tokenEndCol0Based)

		if lineMatch && charInSpan {
			dlogf(logger, isDebug, "RAW_LOG_MANUAL: MATCH FOUND! Terminal: '%s' at L%dC%d (target L%dC%d)", truncateStringForLog(tokenText, 20), tokenLine0Based, tokenStartCol0Based, targetLine, targetChar)
			return tn
		}
		return nil
	}

	for i := 0; i < node.GetChildCount(); i++ {
		child := node.GetChild(i)
		childParseTree, ok := child.(antlr.ParseTree)
		if !ok {
			continue
		}
		found := findInitialNodeManually(childParseTree, targetLine, targetChar, logger, isDebug)
		if found != nil {
			return found
		}
	}
	return nil
}

func getRuleNameSafe(node antlr.Tree, ruleNames []string) string {
	if rCtx, ok := node.(antlr.RuleContext); ok {
		idx := rCtx.GetRuleIndex()
		if idx >= 0 && idx < len(ruleNames) {
			return ruleNames[idx]
		}
		return fmt.Sprintf("InvalidRuleIndex_%d", idx)
	}
	if _, ok := node.(antlr.TerminalNode); ok {
		return "TerminalNode"
	}
	return "UnknownNodeType"
}

func getIdentifiersTextsFromQIGeneric(qiRuleCtx antlr.RuleContext, logger *log.Logger, isDebug bool) []string {
	var identTexts []string
	if qiRuleCtx == nil || qiRuleCtx.GetRuleIndex() != gen.NeuroScriptParserRULE_qualified_identifier {
		dlogf(logger, isDebug, "DEBUG_QI_HELPER: Provided node is not a qualified_identifier. RuleIndex: %v", qiRuleCtx)
		return identTexts
	}
	for i := 0; i < qiRuleCtx.GetChildCount(); i++ {
		child := qiRuleCtx.GetChild(i)
		termNode, okTerm := child.(antlr.TerminalNode)
		if !okTerm {
			continue
		}
		if termNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerIDENTIFIER {
			identTexts = append(identTexts, termNode.GetSymbol().GetText())
		}
	}
	dlogf(logger, isDebug, "DEBUG_QI_HELPER: Identifiers from QI: %v", identTexts)
	return identTexts
}

// extractAndValidateFullToolName extracts the full tool name (e.g., "FS.List") from a qualified_identifier
// and validates it against the tool registry.
func (s *Server) extractAndValidateFullToolName(qiRuleCtx antlr.RuleContext, logger *log.Logger, debugHover bool) string {
	if qiRuleCtx == nil {
		return ""
	}
	ids := getIdentifiersTextsFromQIGeneric(qiRuleCtx, logger, debugHover)
	if len(ids) == 0 {
		dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: qualified_identifier has no IDENTIFIER children.")
		return ""
	}
	candidateToolName := strings.Join(ids, ".")
	dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: Candidate full tool name from QI: '%s'", candidateToolName)

	if s.toolRegistry == nil {
		logger.Printf("ERROR: Hover: ToolRegistry not available in LSP server for validation.")
		return "" // Cannot validate
	}
	if _, found := s.toolRegistry.GetTool(candidateToolName); found {
		dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: Tool '%s' IS registered.", candidateToolName)
		return candidateToolName
	}
	dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: Tool '%s' IS NOT registered.", candidateToolName)
	return "" // Not a registered tool
}

// extractToolNameFromKWRule handles the case where the cursor is on KW_TOOL.
func (s *Server) extractToolNameFromKWRule(foundTokenNode antlr.TerminalNode, logger *log.Logger, debugHover bool) string {
	dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: Cursor on KW_TOOL.")
	parent := foundTokenNode.GetParent()
	parentRuleCtx, okParentRule := parent.(antlr.RuleContext)
	if !okParentRule || parentRuleCtx.GetRuleIndex() != gen.NeuroScriptParserRULE_call_target {
		dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: KW_TOOL's parent is not call_target. Parent: %s", getRuleNameSafe(parent, gen.NeuroScriptParserStaticData.RuleNames))
		return ""
	}

	dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: Parent of KW_TOOL is call_target. Searching for qualified_identifier child.")
	for i := 0; i < parentRuleCtx.GetChildCount(); i++ {
		child := parentRuleCtx.GetChild(i)
		qiCandRuleCtx, okRule := child.(antlr.RuleContext)
		if !okRule || qiCandRuleCtx.GetRuleIndex() != gen.NeuroScriptParserRULE_qualified_identifier {
			continue
		}
		dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: Found qualified_identifier child of call_target.")
		return s.extractAndValidateFullToolName(qiCandRuleCtx, logger, debugHover)
	}
	return ""
}

// extractToolNameFromIdentifierRule handles the case where the cursor is on an IDENTIFIER.
func (s *Server) extractToolNameFromIdentifierRule(foundTokenNode antlr.TerminalNode, logger *log.Logger, debugHover bool) string {
	tokenText := foundTokenNode.GetSymbol().GetText()
	dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: Cursor on IDENTIFIER '%s'. Checking context.", tokenText)

	// Navigate up to the qualified_identifier node that this IDENTIFIER is part of.
	var qiNode antlr.RuleContext
	currentNode := foundTokenNode.GetParent()
	for currentNode != nil {
		if rCtx, ok := currentNode.(antlr.RuleContext); ok && rCtx.GetRuleIndex() == gen.NeuroScriptParserRULE_qualified_identifier {
			qiNode = rCtx
			break
		}
		if rCtx, ok := currentNode.(antlr.RuleContext); ok && rCtx.GetRuleIndex() == gen.NeuroScriptParserRULE_call_target {
			// We've gone too far up without finding a QI, or the ID is the func name itself
			break
		}
		currentNode = currentNode.GetParent()
	}

	if qiNode == nil {
		dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: IDENTIFIER '%s' is not part of a qualified_identifier relevant to tool calls.", tokenText)
		return ""
	}
	dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: IDENTIFIER '%s' is part of QI: %s", tokenText, qiNode.GetText())

	// Now check if this QI is part of a tool.X call
	p2Node := qiNode.GetParent()
	p2RuleCtx, okP2Rule := p2Node.(antlr.RuleContext)
	if !okP2Rule || p2RuleCtx.GetRuleIndex() != gen.NeuroScriptParserRULE_call_target {
		dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: QI's parent is not call_target. Grandparent: %s", getRuleNameSafe(p2Node, gen.NeuroScriptParserStaticData.RuleNames))
		return ""
	}

	dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: QI parent is call_target. Checking if call_target starts with KW_TOOL.")
	if p2RuleCtx.GetChildCount() == 0 {
		return ""
	}

	firstChildOfCallTarget := p2RuleCtx.GetChild(0)
	ftTerm, okFt := firstChildOfCallTarget.(antlr.TerminalNode)
	if !okFt || ftTerm.GetSymbol().GetTokenType() != gen.NeuroScriptLexerKW_TOOL {
		dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: call_target does not start with KW_TOOL. First child: %s", getRuleNameSafe(firstChildOfCallTarget, gen.NeuroScriptParserStaticData.RuleNames))
		return ""
	}

	dlogf(logger, debugHover, "DEBUG_LSP_TOOL_PATH: call_target starts with KW_TOOL. Validating full tool name from QI.")
	return s.extractAndValidateFullToolName(qiNode, logger, debugHover)
}

func (s *Server) extractToolNameAtPosition(content string, position lsp.Position, sourceName string) string {
	s.logger.Printf("INFO: extractToolNameAtPosition called. FileVersion: %s. Source: %s L%dC%d",
		serverExtractToolFileVersion, sourceName, position.Line+1, position.Character+1)
	debugHover := os.Getenv("NSLSP_DEBUG_HOVER") != ""

	treeFromParser, parseErrors := s.coreParserAPI.ParseForLSP(sourceName, content)
	if treeFromParser == nil {
		s.logger.Printf("WARN: extractToolName: AST is nil after parsing.")
		return ""
	}
	if len(parseErrors) > 0 {
		s.logger.Printf("INFO: extractToolName: Parsed with %d errors, but AST tree was returned.", len(parseErrors))
	}

	parseTreeRoot, ok := treeFromParser.(antlr.ParseTree)
	if !ok {
		s.logger.Printf("ERROR_LSP_HOVER: Parsed tree is not an antlr.ParseTree. Type: %T", treeFromParser)
		return ""
	}

	if debugHover {
		enhancedTreeString := formatASTNodeWithPosition(parseTreeRoot, gen.NeuroScriptParserStaticData.RuleNames, "", s.logger)
		s.logger.Printf("DEBUG_LSP_HOVER: Enhanced AST Tree for '%s':\n%s\n--------------------", sourceName, enhancedTreeString)
	}

	dlogf(s.logger, debugHover, "DEBUG_LSP_DIRECT_FIND: Calling findInitialNodeManually. Target: L%dC%d.", position.Line, position.Character)
	foundTokenNode := findInitialNodeManually(parseTreeRoot, position.Line, position.Character, s.logger, debugHover)
	determinedToolName := ""

	if foundTokenNode == nil {
		dlogf(s.logger, debugHover, "DEBUG_LSP_DIRECT_FIND: findInitialNodeManually returned nil. No token at cursor L%dC%d.", position.Line, position.Character)
		s.logger.Printf("INFO: extractToolName: No token at cursor L%dC%d. Token: <foundNode_is_nil>", position.Line+1, position.Character+1)
		return ""
	}

	tokenSymbol := foundTokenNode.GetSymbol()
	tokenType := tokenSymbol.GetTokenType()
	tokenText := tokenSymbol.GetText()
	dlogf(s.logger, debugHover, "DEBUG_LSP_DIRECT_FIND: Found token: '%s' (Type: %s)", tokenText, gen.NeuroScriptLexerLexerStaticData.SymbolicNames[tokenType])

	switch tokenType {
	case gen.NeuroScriptLexerKW_TOOL:
		determinedToolName = s.extractToolNameFromKWRule(foundTokenNode, s.logger, debugHover)
	case gen.NeuroScriptLexerIDENTIFIER, gen.NeuroScriptLexerDOT: // DOT is included if cursor is on it.
		// If on DOT, we try to find the associated IDENTIFIER logic by looking at parent constructs.
		// The extractToolNameFromIdentifierRule is designed to walk up and find the relevant QI.
		determinedToolName = s.extractToolNameFromIdentifierRule(foundTokenNode, s.logger, debugHover)
	}

	if determinedToolName != "" {
		initialTokenTypeAtCursor := tokenSymbol.GetTokenType()
		isCursorOnStringLitToken := initialTokenTypeAtCursor == gen.NeuroScriptLexerSTRING_LIT || initialTokenTypeAtCursor == gen.NeuroScriptLexerTRIPLE_BACKTICK_STRING
		isCursorInsideStringContent := false

		if isCursorOnStringLitToken {
			tokenStartCol := tokenSymbol.GetColumn()
			tokenTextLen := len(tokenSymbol.GetText())
			quoteLen := 1
			if initialTokenTypeAtCursor == gen.NeuroScriptLexerTRIPLE_BACKTICK_STRING {
				quoteLen = 3
			}
			if tokenTextLen > (2*quoteLen-1) && position.Character > tokenStartCol+(quoteLen-1) && position.Character < (tokenStartCol+tokenTextLen-quoteLen) {
				isCursorInsideStringContent = true
			}
		}
		if isCursorInsideStringContent {
			dlogf(s.logger, debugHover, "DEBUG_LSP_DIRECT_FIND: Cursor was inside string literal ('%s'). Clearing determined tool name '%s'.", tokenSymbol.GetText(), determinedToolName)
			determinedToolName = ""
		}
	}

	finalFoundNodeText := "'" + foundTokenNode.GetText() + "'"
	if determinedToolName != "" {
		s.logger.Printf("INFO: extractToolName: Tool name '%s' extracted and validated for L%dC%d. Token: %s",
			determinedToolName, position.Line+1, position.Character+1, finalFoundNodeText)
	} else {
		s.logger.Printf("INFO: extractToolName: No registered 'tool.X' style tool name identified at L%dC%d. Token: %s",
			position.Line+1, position.Character+1, finalFoundNodeText)
	}
	return determinedToolName
}

func formatASTNodeWithPosition(node antlr.Tree, ruleNames []string, indent string, logger *log.Logger) string {
	// ... (rest of the function remains the same as in previous version 0.1.39)
	if node == nil {
		return ""
	}
	var b strings.Builder
	isTerminal := false
	var nodeText string
	var line, col int = -1, -1
	var endLine, endCol int = -1, -1
	var tokenTextForDisplay string

	switch n := node.(type) {
	case antlr.TerminalNode:
		isTerminal = true
		token := n.GetSymbol()
		if token.GetTokenType() == antlr.TokenEOF {
			nodeText, tokenTextForDisplay = "EOF", "<EOF>"
			line, col, endLine, endCol = token.GetLine(), token.GetColumn(), line, col
		} else {
			nodeText = fmt.Sprintf("TOKEN[%s]", gen.NeuroScriptLexerLexerStaticData.SymbolicNames[token.GetTokenType()])
			tokenTextForDisplay = fmt.Sprintf("'%s'", token.GetText())
			line, col, endLine = token.GetLine(), token.GetColumn(), token.GetLine()
			endCol = col + len(token.GetText())
		}
	case antlr.RuleContext:
		if prc, ok := n.(antlr.ParserRuleContext); ok {
			ruleIndex := prc.GetRuleIndex()
			if ruleIndex >= 0 && ruleIndex < len(ruleNames) {
				nodeText = ruleNames[ruleIndex]
			} else {
				nodeText = fmt.Sprintf("INVALID_RULE_INDEX_%d", ruleIndex)
			}
			if prc.GetStart() != nil {
				line, col = prc.GetStart().GetLine(), prc.GetStart().GetColumn()
			}
			if prc.GetStop() != nil {
				endLine, endCol = prc.GetStop().GetLine(), prc.GetStop().GetColumn()+len(prc.GetStop().GetText())
			} else {
				endLine, endCol = line, col // Default if stop is nil
			}
			tokenTextForDisplay = truncateStringForLog(prc.GetText(), 40)
		} else {
			nodeText = fmt.Sprintf("RuleContext (Type: %T, Text: %s)", n, truncateStringForLog(n.GetText(), 40))
		}
	default:
		nodeText, tokenTextForDisplay = fmt.Sprintf("UnknownNodeType:%T", node), ""
	}

	b.WriteString(indent)
	b.WriteString(nodeText)
	if line != -1 { // Only print position if valid
		b.WriteString(fmt.Sprintf(" (L%d:C%d - L%d:C%d)", line, col, endLine, endCol))
	}
	if tokenTextForDisplay != "" {
		b.WriteString(fmt.Sprintf(" Text: %s", tokenTextForDisplay))
	}
	b.WriteString("\n")

	if !isTerminal {
		if rc, ok := node.(antlr.RuleContext); ok {
			for i := 0; i < rc.GetChildCount(); i++ {
				child := rc.GetChild(i)
				if childTree, okTree := child.(antlr.Tree); okTree {
					b.WriteString(formatASTNodeWithPosition(childTree, ruleNames, indent+"  ", logger))
				} else if logger != nil {
					logger.Printf("%s  Nil or non-Tree child at index %d for node %s. Child type: %T", indent, i, nodeText, child)
				}
			}
		}
	}
	return b.String()
}
